/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelNsxtManagerOpenApi = "NSX-T Manager"

type NsxtManagerOpenApi struct {
	NsxtManagerOpenApi *types.NsxtManagerOpenApi
	vcdClient          *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (t NsxtManagerOpenApi) wrap(inner *types.NsxtManagerOpenApi) *NsxtManagerOpenApi {
	t.NsxtManagerOpenApi = inner
	return &t
}

// CreateNsxtManagerOpenApi creates NSX-T
func (vcdClient *VCDClient) CreateNsxtManagerOpenApi(config *types.NsxtManagerOpenApi) (*NsxtManagerOpenApi, error) {
	c := crudConfig{
		entityLabel: labelNsxtManagerOpenApi,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
	}
	outerType := NsxtManagerOpenApi{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllNsxtManagersOpenApi retrieves all NSX-T Managers with an optional filter
func (vcdClient *VCDClient) GetAllNsxtManagersOpenApi(queryParameters url.Values) ([]*NsxtManagerOpenApi, error) {
	c := crudConfig{
		entityLabel:     labelNsxtManagerOpenApi,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
		queryParameters: queryParameters,
	}

	outerType := NsxtManagerOpenApi{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetNsxtManagerOpenApiById retrieves NSX-T Manager by ID
func (vcdClient *VCDClient) GetNsxtManagerOpenApiById(id string) (*NsxtManagerOpenApi, error) {
	c := crudConfig{
		entityLabel:    labelNsxtManagerOpenApi,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
		endpointParams: []string{id},
	}

	outerType := NsxtManagerOpenApi{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetNsxtManagerOpenApiByName retrieves NSX-T Manager by name
func (vcdClient *VCDClient) GetNsxtManagerOpenApiByName(name string) (*NsxtManagerOpenApi, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelNsxtManagerOpenApi)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllNsxtManagersOpenApi(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleEntity, nil
}

// Update NSX-T Manager configuration
func (t *NsxtManagerOpenApi) Update(TmNsxtManagerConfig *types.NsxtManagerOpenApi) (*NsxtManagerOpenApi, error) {
	c := crudConfig{
		entityLabel:    labelNsxtManagerOpenApi,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
		endpointParams: []string{t.NsxtManagerOpenApi.ID},
	}
	outerType := NsxtManagerOpenApi{vcdClient: t.vcdClient}
	return updateOuterEntity(&t.vcdClient.Client, outerType, c, TmNsxtManagerConfig)
}

// Delete NSX-T Manager configuration
func (t *NsxtManagerOpenApi) Delete() error {
	c := crudConfig{
		entityLabel:    labelNsxtManagerOpenApi,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
		endpointParams: []string{t.NsxtManagerOpenApi.ID},
	}
	return deleteEntityById(&t.vcdClient.Client, c)
}
