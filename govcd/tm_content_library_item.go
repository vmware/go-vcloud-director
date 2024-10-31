package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/url"
)

const labelContentLibraryItem = "Content Library Item"

// ContentLibraryItem defines the Content Library Item data structure
type ContentLibraryItem struct {
	ContentLibraryItem *types.ContentLibraryItem
	vcdClient          *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g ContentLibraryItem) wrap(inner *types.ContentLibraryItem) *ContentLibraryItem {
	g.ContentLibraryItem = inner
	return &g
}

// CreateContentLibraryItem creates a Content Library Item
func (cl *ContentLibrary) CreateContentLibraryItem(config *types.ContentLibraryItem) (*ContentLibraryItem, error) {
	// POST https:///tm/cloudapi/vcf/contentLibraryItems
	//application/json;version=41.0.0-alpha
	//{"name":"test-iso","description":"","contentLibrary":{"id":"urn:vcloud:contentLibrary:2e3ea73c-af69-4ae2-9a7a-bdeb3f213e61"}}
	//
	//GET https:///tm/cloudapi/vcf/contentLibraryItems/urn:vcloud:contentLibraryItem:28e4fdc2-bb0d-48b9-95fa-07741b6c95c3/files?page=1&pageSize=128&links=true
	//{
	//            "expectedSizeBytes": -1,
	//            "bytesTransferred": 0,
	//            "name": "descriptor.ovf",
	//            "transferUrl": "https:///transfer/3b71f71d-c819-42a5-8e85-c520c0892323/descriptor.ovf"
	//        }
	//
	//PUT https://tm/transfer/3b71f71d-c819-42a5-8e85-c520c0892323/descriptor.ovf
	//Payload is OVF XML
	//
	//PUT https:/tm/transfer/3b71f71d-c819-42a5-8e85-c520c0892323/photon-ova-disk1.vmdk
	//Payload is OVA
	return nil, nil
}

// GetAllContentLibraryItems retrieves all Content Library Items with the given query parameters, which allow setting filters
// and other constraints
func (cl *ContentLibrary) GetAllContentLibraryItems(queryParameters url.Values) ([]*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel:     labelContentLibrary,
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

	filteredEntities, err := cl.vcdClient.GetAllContentLibraries(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return cl.GetContentLibraryItemById(singleEntity.ContentLibrary.Id)
}

// GetContentLibraryItemById retrieves a Content Library Item with the given ID
func (cl *ContentLibrary) GetContentLibraryItemById(id string) (*ContentLibraryItem, error) {
	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems,
		endpointParams: []string{id},
	}

	outerType := ContentLibraryItem{vcdClient: cl.vcdClient}
	return getOuterEntity(&cl.vcdClient.Client, outerType, c)
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
