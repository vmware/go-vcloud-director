// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmIpSpaceAssociation = "TM IP Space Association"

// TmIpSpaceAssociation manages associations between Provider Gateways and IP Spaces. Each
// association results in a separate entity. There is no update option. The first association is
// created automatically when a Provider Gateway (`TmProviderGateway`) is created.
type TmIpSpaceAssociation struct {
	TmIpSpaceAssociation *types.TmIpSpaceAssociation
	vcdClient            *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmIpSpaceAssociation) wrap(inner *types.TmIpSpaceAssociation) *TmIpSpaceAssociation {
	g.TmIpSpaceAssociation = inner
	return &g
}

// Creates TM IP Space Association with Provider Gateway
func (vcdClient *VCDClient) CreateTmIpSpaceAssociation(config *types.TmIpSpaceAssociation) (*TmIpSpaceAssociation, error) {
	c := crudConfig{
		entityLabel: labelTmIpSpaceAssociation,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations,
		requiresTm:  true,
	}
	outerType := TmIpSpaceAssociation{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllTmIpSpaceAssociations retrieves all TM IP Space and Provider Gateway associations
func (vcdClient *VCDClient) GetAllTmIpSpaceAssociations(queryParameters url.Values) ([]*TmIpSpaceAssociation, error) {
	c := crudConfig{
		entityLabel:     labelTmIpSpaceAssociation,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmIpSpaceAssociation{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmIpSpaceAssociationById retrieves a single IP Spaces and Provider Gateway association by ID
func (vcdClient *VCDClient) GetTmIpSpaceAssociationById(id string) (*TmIpSpaceAssociation, error) {
	c := crudConfig{
		entityLabel:    labelTmIpSpaceAssociation,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmIpSpaceAssociation{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetAllTmIpSpaceAssociationsByProviderGatewayId retrieves all IP Space associations to a
// particulat Provider Gateway
func (vcdClient *VCDClient) GetAllTmIpSpaceAssociationsByProviderGatewayId(providerGatewayId string) ([]*TmIpSpaceAssociation, error) {
	if providerGatewayId == "" {
		return nil, fmt.Errorf("%s lookup requires %s ID", labelTmIpSpaceAssociation, labelTmProviderGateway)
	}

	queryParams := url.Values{}
	queryParams = queryParameterFilterAnd("providerGatewayRef.id=="+providerGatewayId, queryParams)

	return vcdClient.GetAllTmIpSpaceAssociations(queryParams)
}

// GetAllTmIpSpaceAssociationsByIpSpaceId retrieves all associations for a given IP Space ID
func (vcdClient *VCDClient) GetAllTmIpSpaceAssociationsByIpSpaceId(ipSpaceId string) ([]*TmIpSpaceAssociation, error) {
	if ipSpaceId == "" {
		return nil, fmt.Errorf("%s lookup requires %s ID", labelTmIpSpaceAssociation, labelTmProviderGateway)
	}

	queryParams := url.Values{}
	queryParams = queryParameterFilterAnd("ipSpaceRef.id=="+ipSpaceId, queryParams)
	return vcdClient.GetAllTmIpSpaceAssociations(queryParams)
}

// Delete the association
func (o *TmIpSpaceAssociation) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmIpSpaceAssociation,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations,
		endpointParams: []string{o.TmIpSpaceAssociation.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
