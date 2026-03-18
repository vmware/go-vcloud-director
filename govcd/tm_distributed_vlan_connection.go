// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmDistributedVlanConnection = "TM Distributed VLAN Connection"

// TmDistributedVlanConnection manages Distributed Vlan Connection creation and configuration.
type TmDistributedVlanConnection struct {
	TmDistributedVlanConnection *types.TmDistributedVlanConnection
	vcdClient                   *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmDistributedVlanConnection) wrap(inner *types.TmDistributedVlanConnection) *TmDistributedVlanConnection {
	g.TmDistributedVlanConnection = inner
	return &g
}

// Creates TM Distributed Vlan Connection with provided configuration
func (vcdClient *VCDClient) CreateTmDistributedVlanConnection(config *types.TmDistributedVlanConnection) (*TmDistributedVlanConnection, error) {
	c := crudConfig{
		entityLabel: labelTmDistributedVlanConnection,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmDistributedVlanConnections,
		requiresTm:  true,
	}
	outerType := TmDistributedVlanConnection{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// CreateTmDistributedVlanConnectionAsync adds new Distributed Vlan Connection and returns its task for tracking
func (vcdClient *VCDClient) CreateTmDistributedVlanConnectionAsync(config *types.TmDistributedVlanConnection) (*Task, error) {
	c := crudConfig{
		entityLabel: labelTmDistributedVlanConnection,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmDistributedVlanConnections,
		requiresTm:  true,
	}
	return createInnerEntityAsync(&vcdClient.Client, c, config)
}

// GetAllTmDistributedVlanConnections retrieves all Distributed Vlan Connections with optional filter
func (vcdClient *VCDClient) GetAllTmDistributedVlanConnections(queryParameters url.Values) ([]*TmDistributedVlanConnection, error) {
	c := crudConfig{
		entityLabel:     labelTmDistributedVlanConnection,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmDistributedVlanConnections,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmDistributedVlanConnection{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmDistributedVlanConnectionByName retrieves a Distributed Vlan Connection by Name
func (vcdClient *VCDClient) GetTmDistributedVlanConnectionByName(name string) (*TmDistributedVlanConnection, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmDistributedVlanConnection)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmDistributedVlanConnections(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmDistributedVlanConnectionById(singleEntity.TmDistributedVlanConnection.ID)
}

// GetTmDistributedVlanConnectionById retrieves a given Distributed Vlan Connection by ID
func (vcdClient *VCDClient) GetTmDistributedVlanConnectionById(id string) (*TmDistributedVlanConnection, error) {
	c := crudConfig{
		entityLabel:    labelTmDistributedVlanConnection,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmDistributedVlanConnections,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmDistributedVlanConnection{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetTmDistributedVlanConnectionByNameAndRegionId retrieves a given Distributed Vlan Connection by name in a given Region
func (vcdClient *VCDClient) GetTmDistributedVlanConnectionByNameAndRegionId(name, regionId string) (*TmDistributedVlanConnection, error) {
	if name == "" || regionId == "" {
		return nil, fmt.Errorf("%s lookup requires name and Region ID", labelTmDistributedVlanConnection)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("regionRef.id=="+regionId, queryParams)

	filteredEntities, err := vcdClient.GetAllTmDistributedVlanConnections(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmDistributedVlanConnectionById(singleEntity.TmDistributedVlanConnection.ID)
}

// Update existing Distributed Vlan Connection
func (o *TmDistributedVlanConnection) Update(TmDistributedVlanConnectionConfig *types.TmDistributedVlanConnection) (*TmDistributedVlanConnection, error) {
	c := crudConfig{
		entityLabel:    labelTmDistributedVlanConnection,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmDistributedVlanConnections,
		endpointParams: []string{o.TmDistributedVlanConnection.ID},
		requiresTm:     true,
	}
	outerType := TmDistributedVlanConnection{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmDistributedVlanConnectionConfig)
}

// Delete a Distributed Vlan Connection
func (o *TmDistributedVlanConnection) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmDistributedVlanConnection,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmDistributedVlanConnections,
		endpointParams: []string{o.TmDistributedVlanConnection.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
