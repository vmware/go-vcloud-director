package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"errors"
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const labelContentLibraryItem = "Content Library Item"

// retriesForPollingContentLibraryItemFilesToUpload specifies the amount of retries used when retrieving the
// Content Library Items files to upload
const retriesForPollingContentLibraryItemFilesToUpload = 10

// ContentLibraryItem defines the Content Library Item data structure
type ContentLibraryItem struct {
	ContentLibraryItem *types.ContentLibraryItem
	vcdClient          *VCDClient
}

// ContentLibraryItemUploadArguments defines the required arguments for Content Library Item uploads
type ContentLibraryItemUploadArguments struct {
	FilePath        string   // Path to the main file to upload
	OvfFilesPaths   []string // OVF only: Path to the files referenced by the OVF
	UploadPieceSize int64    // When uploading big files, the payloads are divided into chunks of this size in bytes. Defaults to 'defaultPieceSize'
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g ContentLibraryItem) wrap(inner *types.ContentLibraryItem) *ContentLibraryItem {
	g.ContentLibraryItem = inner
	return &g
}

// CreateContentLibraryItem creates a Content Library Item with the given file located in 'FilePath' parameter, which must
// be an OVA, OVF or ISO file. If OVF is uploaded, 'OvfFilesPaths' must be also set with the path of the referenced files.
func (cl *ContentLibrary) CreateContentLibraryItem(config *types.ContentLibraryItem, args ContentLibraryItemUploadArguments) (*ContentLibraryItem, error) {
	// Clean up given paths
	cleanFilePath := filepath.Clean(args.FilePath)
	if _, err := os.Stat(cleanFilePath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	cleanOvfFilesPaths := make([]string, len(args.OvfFilesPaths))
	for i := range args.OvfFilesPaths {
		cleanOvfFilesPaths[i] = filepath.Clean(args.FilePath)
		if _, err := os.Stat(cleanOvfFilesPaths[i]); errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	// Only OVA files have all the required files packed inside, so we need to extract and process them
	if filepath.Ext(cleanFilePath) == ".ova" {
		ovaInnerFilesPaths, tmpDir, err := util.Unpack(args.FilePath)
		if err != nil {
			return nil, fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
		}
		defer func() {
			err = os.RemoveAll(tmpDir)
			if err != nil {
				util.Logger.Printf("[DEBUG] could not clean up tmp directory %s", tmpDir)
			}
		}()
		// For convenience, use the args structure to hold the result. This way, the processing is the same as
		// for OVF uploads
		for _, p := range ovaInnerFilesPaths {
			if filepath.Ext(p) == ".ovf" {
				args.FilePath = p
			} else {
				args.OvfFilesPaths = append(args.OvfFilesPaths, p)
			}
		}
	}

	fileExt := filepath.Ext(args.FilePath)
	if fileExt != ".iso" && fileExt != ".ovf" {
		return nil, fmt.Errorf("%s is not a valid ISO/OVF file", args.FilePath)
	}

	// Create the "skeleton" of the Content Library Item. This is empty, we need to send the files afterward
	cli, err := createContentLibraryItem(cl, config, args.FilePath)
	if err != nil {
		if cli == nil || cli.ContentLibraryItem == nil {
			return nil, err
		}
		// We use Name for cleanup because ID may or may not be available
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.Name, err)
	}

	// Get the files that need to be uploaded after creation, this should always return 1: Either the ISO file or the "descriptor.ovf"
	filesToUpload, err := getContentLibraryItemPendingFilesToUpload(cli, 1, retriesForPollingContentLibraryItemFilesToUpload)
	if err != nil {
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
	}
	if len(filesToUpload) != 1 {
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, fmt.Errorf("expected 1 %s File to upload, got %d", labelContentLibraryItem, len(filesToUpload)))
	}

	// Upload either the requested ISO file or the "descriptor.ovf"
	err = uploadContentLibraryItemFile(&cl.vcdClient.Client, filesToUpload[0], []string{args.FilePath}, args.UploadPieceSize)
	if err != nil {
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
	}

	if cli.ContentLibraryItem.ItemType == "TEMPLATE" {
		// When descriptor.ovf is uploaded, the links for the remaining files will be present in the file list.
		// Refresh the file list and upload each one of them.
		filesToUpload, err = getContentLibraryItemPendingFilesToUpload(cli, 2, retriesForPollingContentLibraryItemFilesToUpload)
		if err != nil {
			return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
		}
		if len(filesToUpload) < 1 {
			return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, fmt.Errorf("expected at least 1 file to upload during OVA processing, got %d", len(filesToUpload)))
		}

		for _, fileToUpload := range filesToUpload {
			if fileToUpload.BytesTransferred != 0 && fileToUpload.ExpectedSizeBytes == fileToUpload.BytesTransferred {
				continue
			}
			err = uploadContentLibraryItemFile(&cl.vcdClient.Client, fileToUpload, args.OvfFilesPaths, args.UploadPieceSize)
			if err != nil {
				return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
			}
		}
	}

	// Track the upload task to see whether it is finished
	err = getContentLibraryItemUploadTask(cli, func(task *Task) error {
		if task == nil {
			// The task does not exist, so upload has finished already
			return nil
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
	}

	// Return the created Content Library Item
	cli, err = cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
	if err != nil {
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
	}
	return cli, nil
}

// uploadContentLibraryItemFile uploads a single Content Library Item File that must be present in one of the given paths and is
// requested by VCFA.
func uploadContentLibraryItemFile(client *Client, fileToUpload *types.ContentLibraryItemFile, filePaths []string, uploadPieceSize int64) error {
	if fileToUpload == nil || len(filePaths) == 0 {
		return fmt.Errorf("the Content Library Item or its files cannot be nil / empty")
	}
	if fileToUpload.BytesTransferred != 0 && fileToUpload.ExpectedSizeBytes == fileToUpload.BytesTransferred {
		return fmt.Errorf("the file %s is already uploaded", fileToUpload.Name)
	}

	for _, filePath := range filePaths {
		isIso := strings.HasSuffix(fileToUpload.Name, ".iso")
		isOvf := strings.HasSuffix(fileToUpload.Name, ".ovf")
		if !isIso && !isOvf && filepath.Base(filePath) != fileToUpload.Name {
			continue
		}
		if isIso && filepath.Ext(filePath) != ".iso" {
			continue
		}
		if isOvf && filepath.Ext(filePath) != ".ovf" {
			continue
		}
		_, err := uploadFile(client, filePath, uploadDetails{
			uploadLink:               fileToUpload.TransferUrl,
			uploadedBytes:            0,
			fileSizeToUpload:         fileToUpload.ExpectedSizeBytes,
			uploadPieceSize:          uploadPieceSize,
			uploadedBytesForCallback: 0,
			allFilesSize:             fileToUpload.ExpectedSizeBytes,
			callBack: func(bytesUpload, totalSize int64) {
				util.Logger.Printf("[DEBUG] Uploaded Content Library Item file '%s': %d/%d", fileToUpload.Name, bytesUpload, totalSize)
			},
			uploadError: addrOf(fmt.Errorf("error uploading Content Library Item file '%s'", fileToUpload.Name)),
		})
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("'%s' not found among the local file paths: %v", fileToUpload.Name, filePaths)
}

// getContentLibraryItemPendingFilesToUpload polls the Content Library Item until a minimum amount of expected files is obtained for
// the given amount of retries. If retries are reached and the expected files are not retrieved, returns an error.
// This function should be used to poll TM until it returns the files that the client SDK should upload next during the upload process.
// For example, at the beginning of the upload, only one file is expected (descriptor.ovf), but once it is uploaded, more files will be
// *eventually* returned by TM to upload next. This function can poll TM until these are returned correctly.
func getContentLibraryItemPendingFilesToUpload(cli *ContentLibraryItem, expectedAtLeast, retries int) ([]*types.ContentLibraryItemFile, error) {
	i := 0
	var files []*types.ContentLibraryItemFile
	var err error
	for i < retries {
		files, err = getContentLibraryItemFiles(cli)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 || len(files) < expectedAtLeast {
			time.Sleep(10 * time.Second)
			i++
			continue
		}
		return files, nil
	}
	return nil, fmt.Errorf("was expecting at least %d files to upload for Content Library Item '%s' in %d retries, but failed", expectedAtLeast, cli.ContentLibraryItem.Name, retries)
}

// createContentLibraryItem creates a hollow Content Library Item with the provided configuration and returns
// the generated result, that should be used to upload the files next.
func createContentLibraryItem(cl *ContentLibrary, config *types.ContentLibraryItem, filePath string) (*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel: labelContentLibraryItem,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		additionalHeader: getTenantContextHeader(&TenantContext{
			OrgId:   cl.ContentLibrary.Org.ID,
			OrgName: cl.ContentLibrary.Org.Name,
		}),
	}
	outerType := ContentLibraryItem{vcdClient: cl.vcdClient}

	// If some config data is missing, we re-use the info from the target Content Library
	if config != nil && config.ContentLibrary.Name == "" {
		config.ContentLibrary.Name = cl.ContentLibrary.Name
	}
	if config != nil && config.ContentLibrary.ID == "" {
		config.ContentLibrary.ID = cl.ContentLibrary.ID
	}
	config.ItemType = "TEMPLATE"

	// If the file is an ISO file, it is required to send its size during creation
	cleanPath := filepath.Clean(filePath)
	if filepath.Ext(cleanPath) == ".iso" {
		config.ItemType = "ISO"
		fileInfo, err := os.Stat(cleanPath)
		if err != nil {
			return nil, err
		}
		config.FileUploadSizeBytes = fileInfo.Size()
	}
	return createOuterEntity(&cl.vcdClient.Client, outerType, c, config)
}

// cleanupContentLibraryItemOnUploadError prevents leaving stranded Content Library Items when any step of the creation (upload)
// process fails.
func cleanupContentLibraryItemOnUploadError(cl *ContentLibrary, identifier string, originalError error) error {
	var err error
	var cli *ContentLibraryItem
	if strings.Contains(identifier, "urn:vcloud:contentLibraryItem") {
		cli, err = cl.GetContentLibraryItemById(identifier)
	} else {
		cli, err = cl.GetContentLibraryItemByName(identifier)
	}
	if ContainsNotFound(err) {
		// Nothing to do
		return originalError
	}
	if err != nil {
		return fmt.Errorf("the Content Library Item creation failed with error: %s\nCleanup of stranded Content Library Item also failed: %s", originalError, err)
	}
	err = getContentLibraryItemUploadTask(cli, func(task *Task) error {
		var innerErr error
		if task == nil {
			// The task does not exist, so we try to delete the Content Library Item directly
			innerErr = cli.Delete()
			if innerErr != nil {
				return innerErr
			}
			return nil
		}
		// Task exists, so we cancel it and let TM do the cleanup
		innerErr = task.CancelTask()
		if innerErr != nil {
			return innerErr
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("the Content Library Item creation failed with error: %s\nCleanup of stranded Content Library Item also failed: %s", originalError, err)
	}
	return originalError
}

// getContentLibraryItemUploadTask searches for the task associated to the given Content Library Item upload and runs the input
// function on it
func getContentLibraryItemUploadTask(cli *ContentLibraryItem, operation func(task *Task) error) error {
	taskRecords, err := cli.vcdClient.Client.QueryTaskList(map[string]string{
		"name":       "contentLibraryItemUpload",
		"status":     "running,preRunning,queued,error",
		"objectType": "contentLibraryItem",
		"objectName": cli.ContentLibraryItem.Name,
	})
	if err != nil {
		return err
	}
	var task *Task
	for _, tr := range taskRecords {
		if equalIds(cli.ContentLibraryItem.ID, tr.Object, tr.HREF) {
			task, err = cli.vcdClient.Client.GetTaskByHREF(tr.HREF)
			if err != nil && !ContainsNotFound(err) {
				return err
			}
			break
		}
	}
	err = operation(task)
	if err != nil {
		return err
	}
	return nil
}

// getContentLibraryItemFiles retrieves the Content Library Item files that need to be uploaded
func getContentLibraryItemFiles(cli *ContentLibraryItem) ([]*types.ContentLibraryItemFile, error) {
	c := crudConfig{
		entityLabel:    labelContentLibraryItem,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItemFiles,
		endpointParams: []string{cli.ContentLibraryItem.ID},
		requiresTm:     true,
	}
	return getAllInnerEntities[types.ContentLibraryItemFile](&cli.vcdClient.Client, c)
}

// GetAllContentLibraryItems retrieves all Content Library Items with the given query parameters, which allow setting filters
// and other constraints
func (cl *ContentLibrary) GetAllContentLibraryItems(queryParameters url.Values) ([]*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel:     labelContentLibraryItem,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		queryParameters: defaultPageSize(queryParameters, "64"), // Content Library Item endpoint forces to use maximum 64 items per page
		additionalHeader: getTenantContextHeader(&TenantContext{
			OrgId:   cl.ContentLibrary.Org.ID,
			OrgName: cl.ContentLibrary.Org.Name,
		}),
		requiresTm: true,
	}

	outerType := ContentLibraryItem{vcdClient: cl.vcdClient}
	return getAllOuterEntities(&cl.vcdClient.Client, outerType, c)
}

// GetContentLibraryItemByName retrieves a Content Library Item with the given name
func (cl *ContentLibrary) GetContentLibraryItemByName(name string) (*ContentLibraryItem, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelContentLibraryItem)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := cl.GetAllContentLibraryItems(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return cl.GetContentLibraryItemById(singleEntity.ContentLibraryItem.ID)
}

// GetContentLibraryItemById retrieves a Content Library Item with the given ID
func (cl *ContentLibrary) GetContentLibraryItemById(id string) (*ContentLibraryItem, error) {
	return cl.vcdClient.GetContentLibraryItemById(id)
}

// GetContentLibraryItemById retrieves a Content Library Item with the given ID
func (vcdClient *VCDClient) GetContentLibraryItemById(id string) (*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel:    labelContentLibraryItem,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := ContentLibraryItem{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update updates an existing Content Library Item with the given configuration
func (o *ContentLibraryItem) Update(contentLibraryItemConfig *types.ContentLibraryItem) (*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel:    labelContentLibraryItem,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		endpointParams: []string{o.ContentLibraryItem.ID},
		requiresTm:     true,
	}
	outerType := ContentLibraryItem{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, contentLibraryItemConfig)
}

// Delete deletes the receiver Content Library Item
func (cli *ContentLibraryItem) Delete() error {
	c := crudConfig{
		entityLabel:    labelContentLibraryItem,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		endpointParams: []string{cli.ContentLibraryItem.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&cli.vcdClient.Client, c)
}
