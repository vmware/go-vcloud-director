package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/url"
	"strings"
)

const labelContentLibrary = "Content Library"

// ContentLibrary defines the Content Library data structure
type ContentLibrary struct {
	ContentLibrary *types.ContentLibrary
	vcdClient      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g ContentLibrary) wrap(inner *types.ContentLibrary) *ContentLibrary {
	g.ContentLibrary = inner
	return &g
}

// CreateContentLibrary creates a Content Library
// TODO: TM: This one probably needs TenantContext, as can be created as Tenants
func (vcdClient *VCDClient) CreateContentLibrary(config *types.ContentLibrary) (*ContentLibrary, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("creating Content Libraries is only supported in TM")
	}
	c := crudConfig{
		entityLabel: labelContentLibrary,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
	}
	outerType := ContentLibrary{vcdClient: vcdClient}
	// FIXME: TM: Workaround, this should be eventually refactored to match other OpenAPI endpoints.
	//        - Problem: When creating a Content Library, it always throws an error 500: "Failed to validate Content Library UUID..."
	//        - Solution: Retry fetching the entity again with the name provided
	result, err := createOuterEntity(&vcdClient.Client, outerType, c, config)
	if err != nil {
		// The error we want is like:
		// Failed to validate Content Library UUID f215ce12-08ac-488e-bbfb-e13c5bad461b, error: not found
		if !strings.Contains(err.Error(), "Failed to validate Content Library UUID") {
			return nil, err
		}
		result, err = vcdClient.GetContentLibraryByName(config.Name)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// GetAllContentLibraries retrieves all Content Libraries with the given query parameters, which allow setting filters
// and other constraints
func (vcdClient *VCDClient) GetAllContentLibraries(queryParameters url.Values) ([]*ContentLibrary, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("retrieving Content Libraries is only supported in TM")
	}
	c := crudConfig{
		entityLabel:     labelContentLibrary,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		queryParameters: queryParameters,
	}

	outerType := ContentLibrary{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetContentLibraryByName retrieves a Content Library with the given name
func (vcdClient *VCDClient) GetContentLibraryByName(name string) (*ContentLibrary, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("retrieving Content Libraries is only supported in TM")
	}

	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelContentLibrary)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllContentLibraries(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetContentLibraryById(singleEntity.ContentLibrary.ID)
}

// GetContentLibraryById retrieves a Content Library with the given ID
func (vcdClient *VCDClient) GetContentLibraryById(id string) (*ContentLibrary, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("retrieving Content Libraries is only supported in TM")
	}

	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams: []string{id},
	}

	outerType := ContentLibrary{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update updates an existing Content Library with the given configuration
// TODO: TM: Not supported in UI yet
func (o *ContentLibrary) Update(contentLibraryConfig *types.ContentLibrary) (*ContentLibrary, error) {
	return nil, fmt.Errorf("not supported")
}

// Delete deletes the receiver Content Library
func (o *ContentLibrary) Delete() error {
	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams: []string{o.ContentLibrary.ID},
	}
	err := deleteEntityById(&o.vcdClient.Client, c)
	// FIXME: TM: Workaround, this should be eventually refactored to match other OpenAPI endpoints.
	//        - Problem: When deleting a Content Library, it always throws an error 500: "Failed to validate Content Library UUID..."
	//        - Solution: Ignore this error, the Content Library is deleted correctly
	if err != nil {
		if strings.Contains(err.Error(), "Failed to validate Content Library UUID") {
			return nil
		}
		return err
	}
	return nil
}
