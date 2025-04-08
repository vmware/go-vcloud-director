// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
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

// CreateNsxtManagerOpenApi attaches NSX-T Manager
func (vcdClient *VCDClient) CreateNsxtManagerOpenApi(config *types.NsxtManagerOpenApi) (*NsxtManagerOpenApi, error) {
	c := crudConfig{
		entityLabel: labelNsxtManagerOpenApi,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointNsxManagers,
		requiresTm:  true,
	}
	outerType := NsxtManagerOpenApi{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllNsxtManagersOpenApi retrieves all NSX-T Managers with an optional filter
func (vcdClient *VCDClient) GetAllNsxtManagersOpenApi(queryParameters url.Values) ([]*NsxtManagerOpenApi, error) {
	c := crudConfig{
		entityLabel:     labelNsxtManagerOpenApi,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointNsxManagers,
		queryParameters: defaultPageSize(queryParameters, "15"),
		requiresTm:      true,
	}

	outerType := NsxtManagerOpenApi{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetNsxtManagerOpenApiById retrieves NSX-T Manager by ID
func (vcdClient *VCDClient) GetNsxtManagerOpenApiById(id string) (*NsxtManagerOpenApi, error) {
	c := crudConfig{
		entityLabel:    labelNsxtManagerOpenApi,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointNsxManagers,
		endpointParams: []string{id},
		requiresTm:     true,
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

// GetNsxtManagerOpenApiByName retrieves NSX-T Manager by name
func (vcdClient *VCDClient) GetNsxtManagerOpenApiByUrl(nsxtManagerUrl string) (*NsxtManagerOpenApi, error) {
	if nsxtManagerUrl == "" {
		return nil, fmt.Errorf("%s lookup requires URL", labelNsxtManagerOpenApi)
	}

	// API filtering by URL is not supported so relying on local filtering
	nsxtManagers, err := vcdClient.GetAllNsxtManagersOpenApi(nil)
	if err != nil {
		return nil, err
	}

	filteredEntities := make([]*NsxtManagerOpenApi, 0)
	for _, nsxtManager := range nsxtManagers {
		if nsxtManager.NsxtManagerOpenApi.Url == nsxtManagerUrl {
			filteredEntities = append(filteredEntities, nsxtManager)
		}

	}

	singleEntity, err := oneOrError("Url", nsxtManagerUrl, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleEntity, nil
}

// Update NSX-T Manager configuration
func (t *NsxtManagerOpenApi) Update(TmNsxtManagerConfig *types.NsxtManagerOpenApi) (*NsxtManagerOpenApi, error) {
	c := crudConfig{
		entityLabel:    labelNsxtManagerOpenApi,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointNsxManagers,
		endpointParams: []string{t.NsxtManagerOpenApi.ID},
		requiresTm:     true,
	}
	outerType := NsxtManagerOpenApi{vcdClient: t.vcdClient}
	return updateOuterEntity(&t.vcdClient.Client, outerType, c, TmNsxtManagerConfig)
}

// Delete NSX-T Manager configuration
func (t *NsxtManagerOpenApi) Delete() error {
	c := crudConfig{
		entityLabel:    labelNsxtManagerOpenApi,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointNsxManagers,
		endpointParams: []string{t.NsxtManagerOpenApi.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&t.vcdClient.Client, c)
}

// BuildHref returns an HREF for an NSX-T Manager
// Sample path https://{PATH}/api/admin/extension/nsxtManagers/48a6dbe1-9b4e-4a75-9947-9d87a3006496
func (t *NsxtManagerOpenApi) BuildHref() string {
	uuid := extractUuid(t.NsxtManagerOpenApi.ID)
	return fmt.Sprintf("%s/api/admin/extension/nsxtManagers/%s", t.vcdClient.Client.rootVcdHref(), uuid)
}
