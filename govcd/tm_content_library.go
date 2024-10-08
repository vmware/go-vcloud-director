package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelContentLibrary = "Content Library"

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

func (vcdClient *VCDClient) CreateContentLibrary(config *types.ContentLibrary) (*ContentLibrary, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("err")
	}
	c := crudConfig{
		entityLabel: labelContentLibrary,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
	}
	outerType := ContentLibrary{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

func (vcdClient *VCDClient) GetAllContentLibrarys(queryParameters url.Values) ([]*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:     labelContentLibrary,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		queryParameters: queryParameters,
	}

	outerType := ContentLibrary{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetContentLibraryByName(name string) (*ContentLibrary, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelContentLibrary)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllContentLibrarys(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetContentLibraryById(singleEntity.ContentLibrary.ID)
}

func (vcdClient *VCDClient) GetContentLibraryById(id string) (*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams: []string{id},
	}

	outerType := ContentLibrary{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (o *ContentLibrary) Update(ContentLibraryConfig *types.ContentLibrary) (*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams: []string{o.ContentLibrary.ID},
	}
	outerType := ContentLibrary{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, ContentLibraryConfig)
}

func (o *ContentLibrary) Delete() error {
	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams: []string{o.ContentLibrary.ID},
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
