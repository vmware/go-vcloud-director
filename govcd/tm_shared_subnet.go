// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmSharedSubnet = "TM Shared Subnet"

// TmSharedSubnet provides configuration of subnets that are extended to a VLAN on a physical
// networking infrastructure
type TmSharedSubnet struct {
	TmSharedSubnet *types.TmSharedSubnet
	vcdClient      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmSharedSubnet) wrap(inner *types.TmSharedSubnet) *TmSharedSubnet {
	g.TmSharedSubnet = inner
	return &g
}

// CreateTmSharedSubnet creates a TM Shared Subnet with a given configuration
func (vcdClient *VCDClient) CreateTmSharedSubnet(config *types.TmSharedSubnet) (*TmSharedSubnet, error) {
	c := crudConfig{
		entityLabel: labelTmSharedSubnet,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmSharedSubnets,
		requiresTm:  true,
	}
	outerType := TmSharedSubnet{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// CreateTmSharedSubnetAsync creates a new TM Shared Subnet and returns its tracking task
func (vcdClient *VCDClient) CreateTmSharedSubnetAsync(config *types.TmSharedSubnet) (*Task, error) {
	c := crudConfig{
		entityLabel: labelTmSharedSubnet,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmSharedSubnets,
		requiresTm:  true,
	}
	return createInnerEntityAsync(&vcdClient.Client, c, config)
}

// GetAllTmSharedSubnets fetches all TM Shared Subnets with an optional query filter
func (vcdClient *VCDClient) GetAllTmSharedSubnets(queryParameters url.Values) ([]*TmSharedSubnet, error) {
	c := crudConfig{
		entityLabel:     labelTmSharedSubnet,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmSharedSubnets,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmSharedSubnet{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmSharedSubnetByName retrieves TM Shared Subnets with a given name
func (vcdClient *VCDClient) GetTmSharedSubnetByName(name string) (*TmSharedSubnet, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmSharedSubnet)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmSharedSubnets(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmSharedSubnetById(singleEntity.TmSharedSubnet.ID)
}

// GetTmSharedSubnetById retrieves an exact Shared Subnet with a given ID
func (vcdClient *VCDClient) GetTmSharedSubnetById(id string) (*TmSharedSubnet, error) {
	c := crudConfig{
		entityLabel:    labelTmSharedSubnet,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmSharedSubnets,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmSharedSubnet{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetTmSharedSubnetByNameAndRegionId retrieves TM Shared Subnets with a given name in a provided Region
func (vcdClient *VCDClient) GetTmSharedSubnetByNameAndRegionId(name, regionId string) (*TmSharedSubnet, error) {
	if name == "" || regionId == "" {
		return nil, fmt.Errorf("%s lookup requires name and Region ID", labelTmSharedSubnet)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("regionRef.id=="+regionId, queryParams)

	filteredEntities, err := vcdClient.GetAllTmSharedSubnets(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmSharedSubnetById(singleEntity.TmSharedSubnet.ID)
}

// Update TM Shared Subnet
func (o *TmSharedSubnet) Update(TmSharedSubnetConfig *types.TmSharedSubnet) (*TmSharedSubnet, error) {
	c := crudConfig{
		entityLabel:    labelTmSharedSubnet,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmSharedSubnets,
		endpointParams: []string{o.TmSharedSubnet.ID},
		requiresTm:     true,
	}
	outerType := TmSharedSubnet{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmSharedSubnetConfig)
}

// Delete TM Shared Subnet
func (o *TmSharedSubnet) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmSharedSubnet,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmSharedSubnets,
		endpointParams: []string{o.TmSharedSubnet.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
