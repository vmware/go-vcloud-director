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
	FilePath        string // Path to the file to upload
	UploadPieceSize int64  // When uploading big files, the payloads are divided into chunks of this size in bytes. Defaults to 'defaultPieceSize'
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g ContentLibraryItem) wrap(inner *types.ContentLibraryItem) *ContentLibraryItem {
	g.ContentLibraryItem = inner
	return &g
}

// CreateContentLibraryItem creates a Content Library Item with the given file located in 'filePath' parameter, which must
// be an OVA or ISO file.
func (cl *ContentLibrary) CreateContentLibraryItem(config *types.ContentLibraryItem, args ContentLibraryItemUploadArguments) (*ContentLibraryItem, error) {
	if _, err := os.Stat(args.FilePath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	cli, err := createContentLibraryItem(cl, config, args.FilePath)
	if err != nil {
		if cli == nil || cli.ContentLibraryItem == nil {
			return nil, err
		}
		// We use Name for cleanup because ID may or may not be available
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.Name, err)
	}
	files, err := getContentLibraryItemPendingFilesToUpload(cli, 1, retriesForPollingContentLibraryItemFilesToUpload)
	if err != nil {
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
	}

	// TODO: TM: We do a refresh as POST response does not populate ItemType. This should be fixed in TM
	//       at some point
	cli, err = cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
	if err != nil {
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
	}

	if cli.ContentLibraryItem.ItemType == "TEMPLATE" {
		// The descriptor must be uploaded first
		err = uploadContentLibraryItemFile("descriptor.ovf", cli, files, args)
		if err != nil {
			return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
		}
		// When descriptor.ovf is uploaded, the links for the remaining files will be present in the file list.
		// Refresh the file list and upload each one of them.
		files, err = getContentLibraryItemPendingFilesToUpload(cli, 2, retriesForPollingContentLibraryItemFilesToUpload)
		if err != nil {
			return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
		}

		for _, f := range files {
			if f.Name == "descriptor.ovf" {
				// Already uploaded
				continue
			}
			err = uploadContentLibraryItemFile(f.Name, cli, files, args)
			if err != nil {
				return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
			}
		}
	} else {
		// TODO: TM: ISO upload
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, fmt.Errorf("ISO upload not supported"))
	}

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

	cli, err = cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
	if err != nil {
		return nil, cleanupContentLibraryItemOnUploadError(cl, cli.ContentLibraryItem.ID, err)
	}
	return cli, nil
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
	}
	outerType := ContentLibraryItem{vcdClient: cl.vcdClient}
	if config != nil && config.ContentLibrary.Name == "" {
		config.ContentLibrary.Name = cl.ContentLibrary.Name
	}
	if config != nil && config.ContentLibrary.ID == "" {
		config.ContentLibrary.ID = cl.ContentLibrary.ID
	}
	if filepath.Ext(filePath) == ".iso" {
		// TODO: TM: Early exit for ISO uploads as they are not supported yet in TM
		return nil, fmt.Errorf("ISO uploads not supported")
		// config.ItemType = "ISO"
	} else {
		config.ItemType = "TEMPLATE"
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

// uploadContentLibraryItemFile uploads a Content Library Item file from the given slice with the given file present on disk
func uploadContentLibraryItemFile(name string, cli *ContentLibraryItem, filesToUpload []*types.ContentLibraryItemFile, args ContentLibraryItemUploadArguments) error {
	if cli == nil || len(filesToUpload) == 0 {
		return fmt.Errorf("the Content Library Item or its files cannot be nil / empty")
	}

	// We just want to upload the selected file (named after the 'name' input parameter)
	var fileToUpload *types.ContentLibraryItemFile
	for _, f := range filesToUpload {
		if f.Name == name {
			fileToUpload = f
			break
		}
	}
	if fileToUpload == nil {
		return fmt.Errorf("'%s' not found among the Content Library Item '%s' files", name, cli.ContentLibraryItem.Name)
	}
	filesAbsPaths, tmpDir, err := util.Unpack(args.FilePath)
	if err != nil {
		return fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
	}
	defer func() {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			util.Logger.Printf("[DEBUG] could not clean up tmp directory %s", tmpDir)
		}
	}()

	ud := uploadDetails{
		uploadLink:               strings.ReplaceAll(fileToUpload.TransferUrl, "/transfer", "/tm/transfer"), // TODO: TM: Workaround, the link is missing /tm in path, so it gives 404 as-is
		uploadedBytes:            0,
		fileSizeToUpload:         fileToUpload.ExpectedSizeBytes,
		uploadPieceSize:          args.UploadPieceSize,
		uploadedBytesForCallback: 0,
		allFilesSize:             fileToUpload.ExpectedSizeBytes,
		callBack: func(bytesUpload, totalSize int64) {
			util.Logger.Printf("[DEBUG] Uploaded Content Library Item file '%s': %d/%d", name, bytesUpload, totalSize)
		},
		uploadError: addrOf(fmt.Errorf("error uploading Content Library Item file '%s'", name)),
	}

	// When TM asks for a file called 'descriptor.ovf', it can be that inside the OVA
	// it is not named like that. We search for an .ovf file in this case and we use it
	foundName := name
	if name == "descriptor.ovf" {
		for _, f := range filesAbsPaths {
			if filepath.Ext(f) == ".ovf" {
				_, foundName = filepath.Split(f)
				break
			}
		}
	}

	_, err = uploadFile(&cli.vcdClient.Client, findFilePath(filesAbsPaths, foundName), ud)
	if err != nil {
		return fmt.Errorf("could not upload the file: %s", err)
	}
	return nil
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
	}
	return getAllInnerEntities[types.ContentLibraryItemFile](&cli.vcdClient.Client, c)
}

// GetAllContentLibraryItems retrieves all Content Library Items with the given query parameters, which allow setting filters
// and other constraints
func (cl *ContentLibrary) GetAllContentLibraryItems(queryParameters url.Values) ([]*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel:     labelContentLibraryItem,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		queryParameters: queryParameters,
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
	}

	outerType := ContentLibraryItem{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update updates an existing Content Library Item with the given configuration
// TODO: TM: Not supported in UI yet
func (o *ContentLibraryItem) Update(contentLibraryItemConfig *types.ContentLibraryItem) (*ContentLibraryItem, error) {
	return nil, fmt.Errorf("not supported")
}

// Delete deletes the receiver Content Library Item
func (cli *ContentLibraryItem) Delete() error {
	c := crudConfig{
		entityLabel:    labelContentLibraryItem,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		endpointParams: []string{cli.ContentLibraryItem.ID},
	}
	return deleteEntityById(&cli.vcdClient.Client, c)
}
