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
	links, err := cli.GetFiles(nil)
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		// TODO: TM: Maybe include retries here?
		return nil, fmt.Errorf("could not retrieve upload links for Content Library Item %s", cli.ContentLibraryItem.Name)
	}

	// Refresh
	cli, err = cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
	if err != nil {
		return nil, err
	}

	var firstFile *types.ContentLibraryItemFile
	var fileContents []byte
	if cli.ContentLibraryItem.ItemType == "TEMPLATE" {
		// Search for OVF descriptor
		for _, link := range links {
			if link.Name == "descriptor.ovf" {
				firstFile = link
			}
		}
		if firstFile == nil {
			return nil, fmt.Errorf("descriptor.ovf not found for Content Library Item '%s'", cli.ContentLibraryItem.Name)
		}
		filesAbsPaths, tmpDir, err := util.Unpack(filePath)
		if err != nil {
			return nil, fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
		}
		ovfFilePath, err := getOvfPath(filesAbsPaths)
		if err != nil {
			return nil, fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
		}
		fileContents, err = os.ReadFile(filepath.Clean(ovfFilePath))
		if err != nil {
			return nil, err
		}

	} else {
		// ISO file
		// TODO: TM: ISO upload?
		firstFile = links[0]
		fileContents, err = os.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return nil, err
		}
	}

	request, err := newFileUploadRequest(cli.client, firstFile.TransferUrl, fileContents, 0, int64(len(fileContents)), int64(len(fileContents)))
	if err != nil {
		return nil, err
	}

	response, err := cli.client.Http.Do(request)
	if err != nil {
		return nil, err
	}
	err = response.Body.Close()
	if err != nil {
		panic(err)
	}

	// Poll for the files list until the ovf file is uploaded(expectedSizeBytes === bytesTransferred)
	// and then fetch the files list again to get the urls for the vmdk files
	for firstFile.BytesTransferred != firstFile.ExpectedSizeBytes {
		links, err = cli.GetFiles(nil)
		if err != nil {
			return nil, err
		}
		for _, link := range links {
			if link.Name == firstFile.Name {
				firstFile = link
			}
		}
		time.Sleep(30 * time.Second)
	}
	// Upload the vmdk files
	links, err = cli.GetFiles(nil)
	if err != nil {
		return nil, err
	}
	for _, link := range links {
		if link.Name == "descriptor.ovf" {
			continue
		}
		filesAbsPaths, tmpDir, err := util.Unpack(filePath)
		if err != nil {
			return nil, fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
		}
		for _, path := range filesAbsPaths {
			if strings.Contains(path, link.Name) {
				fileContents, err = os.ReadFile(filepath.Clean(path))
				if err != nil {
					return nil, err
				}
				request, err := newFileUploadRequest(cli.client, link.TransferUrl, fileContents, 0, int64(len(fileContents)), int64(len(fileContents)))
				if err != nil {
					return nil, err
				}

				response, err := cli.client.Http.Do(request)
				if err != nil {
					return nil, err
				}
				err = response.Body.Close()
				if err != nil {
					panic(err)
				}
				break
			}
		}
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
	cli.ContentLibraryItem = nil
	return nil
}
