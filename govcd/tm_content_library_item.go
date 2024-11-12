package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
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

// CreateContentLibraryItem creates a Content Library Item
func (cl *ContentLibrary) CreateContentLibraryItem(config *types.ContentLibraryItem, filePath string) (*ContentLibraryItem, error) {
	cli, err := createContentLibraryItem(cl, config, filePath)
	if err != nil {
		return nil, err
	}
	files, err := cli.GetFiles(nil)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		// TODO: TM: Maybe include retries here?
		return nil, fmt.Errorf("could not retrieve upload links for Content Library Item %s", cli.ContentLibraryItem.Name)
	}

	// Refresh
	cli, err = cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
	if err != nil {
		return nil, err
	}

	uploadFile := func(name string, filesToUpload []*types.ContentLibraryItemFile, localFilePath string) error {
		var fileToUpload *types.ContentLibraryItemFile
		// Search for OVF descriptor
		for _, f := range files {
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
		asgjdf := ""
		if name == "descriptor.ovf" {
			asgjdf, err = getOvfPath(filesAbsPaths)
			if err != nil {
				return fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
			}
		}
		fileContents, err := os.ReadFile(filepath.Clean(asgjdf))
		if err != nil {
			return err
		}
		// TODO: TM: Workaround, the link is missing /tm in path, so it gives 404 as-is
		request, err := newFileUploadRequest(cli.client, strings.ReplaceAll(fileToUpload.TransferUrl, "/transfer", "/tm/transfer"), fileContents, 0, int64(len(fileContents)), int64(len(fileContents)))
		if err != nil {
			return err
		}
		response, err := cli.client.Http.Do(request)
		if err != nil {
			return err
		}
		defer func() {
			err = response.Body.Close()
			if err != nil {
				util.Logger.Printf("Error closing response: %s\n", err)
			}
		}()
		response, err = checkRespWithErrType(types.BodyTypeJSON, response, err, &types.OpenApiError{})
		if err != nil {
			return err
		}
		// Poll for the files list until the ovf file is uploaded(expectedSizeBytes === bytesTransferred)
		// and then fetch the files list again to get the urls for the vmdk files
		for {
			files, err = cli.GetFiles(nil)
			if err != nil {
				return err
			}
			for _, f := range files {
				if f.Name == fileToUpload.Name {
					fileToUpload = f
				}
			}
			if fileToUpload.BytesTransferred != fileToUpload.ExpectedSizeBytes {
				time.Sleep(30 * time.Second)
				continue
			}
			break
		}
		return nil
	}

	if cli.ContentLibraryItem.ItemType == "TEMPLATE" {
		// The descriptor must be uploaded first
		err = uploadFile("descriptor.ovf", files, filePath)
		if err != nil {
			return nil, err
		}
		// When descriptor is uploaded, the links for the remaining files will be present in the file list.
		// Refresh the file list and upload each one of them.
		files, err = cli.GetFiles(nil)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if f.Name == "descriptor.ovf" {
				// Already uploaded
				continue
			}
			err = uploadFile(f.Name, files, filePath)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// TODO: TM: ISO upload
		return nil, fmt.Errorf("not supported")
	}

	return cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
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
		entityLabel:     labelContentLibrary,
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
		entityLabel:    labelContentLibrary,
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
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		endpointParams: []string{cli.ContentLibraryItem.ID},
	}
	return deleteEntityById(cli.client, c)
}
