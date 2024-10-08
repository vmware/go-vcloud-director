/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelTmNsxtManager = "NSX-T Manager"

type TmNsxtManager struct {
	TmNsxtManager *types.TmNsxtManager
	vcdClient     *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (t TmNsxtManager) wrap(inner *types.TmNsxtManager) *TmNsxtManager {
	t.TmNsxtManager = inner
	return &t
}

func (vcdClient *VCDClient) CreateTmNsxtManager(config *types.TmNsxtManager) (*TmNsxtManager, error) {
	c := crudConfig{
		entityLabel: labelTmNsxtManager,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
	}
	outerType := TmNsxtManager{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

func (vcdClient *VCDClient) GetAllTmNsxtManagers(queryParameters url.Values) ([]*TmNsxtManager, error) {
	c := crudConfig{
		entityLabel:     labelTmNsxtManager,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
		queryParameters: queryParameters,
	}

	outerType := TmNsxtManager{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetTmNsxtManagerById(id string) (*TmNsxtManager, error) {
	c := crudConfig{
		entityLabel:    labelTmNsxtManager,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
		endpointParams: []string{id},
	}

	outerType := TmNsxtManager{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetTmNsxtManagerByName(name string) (*TmNsxtManager, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmNsxtManager)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmNsxtManagers(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleEntity, nil
}

func (t *TmNsxtManager) Update(TmNsxtManagerConfig *types.TmNsxtManager) (*TmNsxtManager, error) {
	c := crudConfig{
		entityLabel:    labelTmNsxtManager,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
		endpointParams: []string{t.TmNsxtManager.ID},
	}
	outerType := TmNsxtManager{vcdClient: t.vcdClient}
	return updateOuterEntity(&t.vcdClient.Client, outerType, c, TmNsxtManagerConfig)
}

func (t *TmNsxtManager) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmNsxtManager,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmNsxManagers,
		endpointParams: []string{t.TmNsxtManager.ID},
	}
	return deleteEntityById(&t.vcdClient.Client, c)
}
