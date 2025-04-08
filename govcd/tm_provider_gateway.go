// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmProviderGateway = "TM Provider Gateway"

// TmProviderGateway manages Provider Gateway creation and configuration.
//
// NOTE. While creation of Provider Gateway requires at least one IP Space (`TmIpSpace`) reference,
// they are not being returned by API after creation. One must use `TmIpSpaceAssociation` for
// managing IP Space associations with Provider gateways after their initial creation is done.
type TmProviderGateway struct {
	TmProviderGateway *types.TmProviderGateway
	vcdClient         *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmProviderGateway) wrap(inner *types.TmProviderGateway) *TmProviderGateway {
	g.TmProviderGateway = inner
	return &g
}

// Creates TM Provider Gateway with provided configuration
func (vcdClient *VCDClient) CreateTmProviderGateway(config *types.TmProviderGateway) (*TmProviderGateway, error) {
	c := crudConfig{
		entityLabel: labelTmProviderGateway,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		requiresTm:  true,
	}
	outerType := TmProviderGateway{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// CreateTmProviderGatewayAsync adds new Provider gateway and returns its task for tracking
func (vcdClient *VCDClient) CreateTmProviderGatewayAsync(config *types.TmProviderGateway) (*Task, error) {
	c := crudConfig{
		entityLabel: labelTmProviderGateway,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		requiresTm:  true,
	}
	return createInnerEntityAsync(&vcdClient.Client, c, config)
}

// GetAllTmProviderGateways retrieves all Provider Gateways with optional filter
func (vcdClient *VCDClient) GetAllTmProviderGateways(queryParameters url.Values) ([]*TmProviderGateway, error) {
	c := crudConfig{
		entityLabel:     labelTmProviderGateway,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmProviderGateway{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmProviderGatewayByName retrieves Provider Gateway by Name
func (vcdClient *VCDClient) GetTmProviderGatewayByName(name string) (*TmProviderGateway, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmProviderGateway)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmProviderGateways(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmProviderGatewayById(singleEntity.TmProviderGateway.ID)
}

// GetTmProviderGatewayById retrieves a given Provider Gateway by ID
func (vcdClient *VCDClient) GetTmProviderGatewayById(id string) (*TmProviderGateway, error) {
	c := crudConfig{
		entityLabel:    labelTmProviderGateway,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmProviderGateway{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetTmProviderGatewayByNameAndRegionId retrieves Provider Gateway by name in a given Region
func (vcdClient *VCDClient) GetTmProviderGatewayByNameAndRegionId(name, regionId string) (*TmProviderGateway, error) {
	if name == "" || regionId == "" {
		return nil, fmt.Errorf("%s lookup requires name and Org ID", labelTmProviderGateway)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("regionRef.id=="+regionId, queryParams)

	filteredEntities, err := vcdClient.GetAllTmProviderGateways(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmProviderGatewayById(singleEntity.TmProviderGateway.ID)
}

// Update existing Provider Gateway
func (o *TmProviderGateway) Update(TmProviderGatewayConfig *types.TmProviderGateway) (*TmProviderGateway, error) {
	c := crudConfig{
		entityLabel:    labelTmProviderGateway,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		endpointParams: []string{o.TmProviderGateway.ID},
		requiresTm:     true,
	}
	outerType := TmProviderGateway{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmProviderGatewayConfig)
}

// Delete Provider Gateway
func (o *TmProviderGateway) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmProviderGateway,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		endpointParams: []string{o.TmProviderGateway.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
