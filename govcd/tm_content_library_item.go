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
	"path/filepath"
	"strings"
	"time"
)

const labelContentLibraryItem = "Content Library Item"

// ContentLibraryItem defines the Content Library Item data structure
type ContentLibraryItem struct {
	ContentLibraryItem *types.ContentLibraryItem
	client             *Client
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g ContentLibraryItem) wrap(inner *types.ContentLibraryItem) *ContentLibraryItem {
	g.ContentLibraryItem = inner
	return &g
}

func getContentLibraryItemFiles(cli *ContentLibraryItem, expectedAtLeast, retries int) ([]*types.ContentLibraryItemFile, error) {
	i := 0
	var files []*types.ContentLibraryItemFile
	var err error
	for i < retries {
		files, err = cli.GetFiles(nil)
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
	return nil, fmt.Errorf("was expecting at least %d files to upload for Content Library Item '%s' but got none in %d retries", expectedAtLeast, cli.ContentLibraryItem.Name, retries)
}

// CreateContentLibraryItem creates a Content Library Item
func (cl *ContentLibrary) CreateContentLibraryItem(config *types.ContentLibraryItem, filePath string) (*ContentLibraryItem, error) {

	cli, err := createContentLibraryItem(cl, config, filePath)
	if err != nil {
		return nil, err
	}
	files, err := getContentLibraryItemFiles(cli, 1, 10)
	if err != nil {
		return nil, err
	}

	// Refresh
	cli, err = cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
	if err != nil {
		return nil, err
	}

	if cli.ContentLibraryItem.ItemType == "TEMPLATE" {
		// The descriptor must be uploaded first

		err = uploadContentLibraryItemFile("descriptor.ovf", cli, files, filePath)
		if err != nil {
			return nil, err
		}
		// When descriptor.ovf is uploaded, the links for the remaining files will be present in the file list.
		// Refresh the file list and upload each one of them.
		files, err = getContentLibraryItemFiles(cli, 2, 10)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if f.Name == "descriptor.ovf" {
				// Already uploaded
				continue
			}
			err = uploadContentLibraryItemFile(f.Name, cli, files, filePath)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// TODO: TM: ISO upload
		return nil, fmt.Errorf("ISO upload not supported")
	}

	return cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
}

func uploadContentLibraryItemFile(name string, cli *ContentLibraryItem, filesToUpload []*types.ContentLibraryItemFile, localFilePath string) error {
	if cli == nil || len(filesToUpload) == 0 {
		return fmt.Errorf("the Content Library Item or its files cannot be nil / empty")
	}

	// We just want to upload the selected file (named after the 'name' input parameter)
	var fileToUpload *types.ContentLibraryItemFile
	for _, f := range filesToUpload {
		if f.Name == name {
			fileToUpload = f
		}
	}
	if fileToUpload == nil {
		return fmt.Errorf("'%s' not found among the Content Library Item '%s' files", name, cli.ContentLibraryItem.Name)
	}
	filesAbsPaths, tmpDir, err := util.Unpack(localFilePath)
	if err != nil {
		return fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
	}

	ud := uploadDetails{
		uploadLink:               strings.ReplaceAll(fileToUpload.TransferUrl, "/transfer", "/tm/transfer"), // TODO: TM: Workaround, the link is missing /tm in path, so it gives 404 as-is
		uploadedBytes:            0,
		fileSizeToUpload:         fileToUpload.ExpectedSizeBytes,
		uploadPieceSize:          1024,
		uploadedBytesForCallback: 0,
		allFilesSize:             fileToUpload.ExpectedSizeBytes,
		callBack: func(bytesUpload, totalSize int64) {
			util.Logger.Printf("[DEBUG] Uploaded Content Library Item file '%s': %d/%d", name, bytesUpload, totalSize)
		},
		uploadError: addrOf(errors.New("")),
	}

	_, err = uploadFile(cli.client, findFilePath(filesAbsPaths, name), ud)
	if err != nil {
		return fmt.Errorf("could not upload the file: %s", err)
	}
	return nil
}

// createContentLibraryItem creates a hollow Content Library Item with the provided configuration and returns
// the generated result, that should be used to upload the files next.
func createContentLibraryItem(cl *ContentLibrary, config *types.ContentLibraryItem, filePath string) (*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel: labelContentLibraryItem,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
	}
	outerType := ContentLibraryItem{client: cl.client}
	if config != nil && config.ContentLibrary.Name == "" {
		config.ContentLibrary.Name = cl.ContentLibrary.Name
	}
	if config != nil && config.ContentLibrary.ID == "" {
		config.ContentLibrary.ID = cl.ContentLibrary.ID
	}
	if filepath.Ext(filePath) == ".iso" {
		config.ItemType = "ISO"
	} else {
		config.ItemType = "TEMPLATE"
	}
	return createOuterEntity(cl.client, outerType, c, config)
}

// GetFiles retrieves the files information for a given Content Library Item
func (cli *ContentLibraryItem) GetFiles(queryParameters url.Values) ([]*types.ContentLibraryItemFile, error) {
	c := crudConfig{
		entityLabel:     labelContentLibraryItem,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItemFiles,
		endpointParams:  []string{cli.ContentLibraryItem.ID},
		queryParameters: queryParameters,
	}
	return getAllInnerEntities[types.ContentLibraryItemFile](cli.client, c)
}

// GetAllContentLibraryItems retrieves all Content Library Items with the given query parameters, which allow setting filters
// and other constraints
func (cl *ContentLibrary) GetAllContentLibraryItems(queryParameters url.Values) ([]*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel:     labelContentLibraryItem,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		queryParameters: queryParameters,
	}

	outerType := ContentLibraryItem{client: cl.client}
	return getAllOuterEntities(cl.client, outerType, c)
}

// GetContentLibraryItemByName retrieves a Content Library Item with the given name
func (cl *ContentLibrary) GetContentLibraryItemByName(name string) (*ContentLibraryItem, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelContentLibraryItem)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := getAllContentLibraries(cl.client, queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return cl.GetContentLibraryItemById(singleEntity.ContentLibrary.ID)
}

// GetContentLibraryItemById retrieves a Content Library Item with the given ID
func (cl *ContentLibrary) GetContentLibraryItemById(id string) (*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel:    labelContentLibraryItem,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		endpointParams: []string{id},
	}

	outerType := ContentLibraryItem{client: cl.client}
	return getOuterEntity(cl.client, outerType, c)
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
	return deleteEntityById(cli.client, c)
}
