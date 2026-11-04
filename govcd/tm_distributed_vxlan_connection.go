// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/http"
	"net/url"
)

const labelTmDistributedVxlanConnection = "TM Distributed VXLAN Connection"

// TmDistributedVxlanConnection manages a Distributed VXLAN Connection.
type TmDistributedVxlanConnection struct {
	TmDistributedVxlanConnection *types.TmDistributedVxlanConnection
	vcdClient                    *VCDClient
	options                      []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmDistributedVxlanConnection) wrap(inner *types.TmDistributedVxlanConnection) *TmDistributedVxlanConnection {
	g.TmDistributedVxlanConnection = inner
	return &g
}

// tmDistributedVxlanConnectionConfig returns the CRUD configuration for Distributed VXLAN Connection operations. Optional params are appended to the endpoint.
func tmDistributedVxlanConnectionConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmDistributedVxlanConnection,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmDistributedVxlanConnections,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "9.1.1",
	}.withTmOptions(options)
}

// CreateTmDistributedVxlanConnection creates a Distributed VXLAN Connection. Waits for task completion. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) CreateTmDistributedVxlanConnection(config *types.TmDistributedVxlanConnection, options ...TmOperationOptions) (*TmDistributedVxlanConnection, error) {
	outerType := TmDistributedVxlanConnection{vcdClient: vcdClient, options: retainTmOptions(options)}
	return createOuterEntity(&vcdClient.Client, outerType, tmDistributedVxlanConnectionConfig(options), config)
}

// CreateTmDistributedVxlanConnectionAsync creates a Distributed VXLAN Connection. Returns the tracking task without waiting. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) CreateTmDistributedVxlanConnectionAsync(config *types.TmDistributedVxlanConnection, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmDistributedVxlanConnection config must not be nil")
	}
	return createInnerEntityAsync(&vcdClient.Client, tmDistributedVxlanConnectionConfig(options), config)
}

// GetAllTmDistributedVxlanConnections retrieves all Distributed VXLAN Connections. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) GetAllTmDistributedVxlanConnections(queryParameters url.Values, options ...TmOperationOptions) ([]*TmDistributedVxlanConnection, error) {
	c := tmDistributedVxlanConnectionConfig(options)
	c.queryParameters = queryParameters
	outerType := TmDistributedVxlanConnection{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmDistributedVxlanConnectionByName retrieves a Distributed VXLAN Connection by name. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) GetTmDistributedVxlanConnectionByName(name string, options ...TmOperationOptions) (*TmDistributedVxlanConnection, error) {
	entities, err := vcdClient.GetAllTmDistributedVxlanConnections(nil, options...)
	if err != nil {
		return nil, err
	}
	return tmEntityByName(labelTmDistributedVxlanConnection, name, entities, func(e *TmDistributedVxlanConnection) string { return e.TmDistributedVxlanConnection.Name })
}

// GetTmDistributedVxlanConnectionById retrieves a Distributed VXLAN Connection by ID. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) GetTmDistributedVxlanConnectionById(id string, options ...TmOperationOptions) (*TmDistributedVxlanConnection, error) {
	outerType := TmDistributedVxlanConnection{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmDistributedVxlanConnectionConfig(options, id))
}

// GetTmDistributedVxlanConnectionByNameOrId retrieves a Distributed VXLAN Connection by name or ID. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) GetTmDistributedVxlanConnectionByNameOrId(identifier string, options ...TmOperationOptions) (*TmDistributedVxlanConnection, error) {
	return tmEntityByNameOrId(identifier,
		func(id string) (*TmDistributedVxlanConnection, error) {
			return vcdClient.GetTmDistributedVxlanConnectionById(id, options...)
		},
		func(name string) (*TmDistributedVxlanConnection, error) {
			return vcdClient.GetTmDistributedVxlanConnectionByName(name, options...)
		})
}

// Update updates a Distributed VXLAN Connection. Waits for task completion. Requires TM API 9.1.1+.
func (o *TmDistributedVxlanConnection) Update(config *types.TmDistributedVxlanConnection, options ...TmOperationOptions) (*TmDistributedVxlanConnection, error) {
	options = inheritTmOptions(o.options, options)
	outerType := TmDistributedVxlanConnection{vcdClient: o.vcdClient, options: retainTmOptions(options)}
	return updateOuterEntity(&o.vcdClient.Client, outerType, tmDistributedVxlanConnectionConfig(options, o.TmDistributedVxlanConnection.ID), config)
}

// UpdateAsync updates a Distributed VXLAN Connection. Returns the tracking task without waiting. Requires TM API 9.1.1+.
func (o *TmDistributedVxlanConnection) UpdateAsync(config *types.TmDistributedVxlanConnection, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	options = inheritTmOptions(o.options, options)
	return runOpenApiOperationAsync(&o.vcdClient.Client, tmDistributedVxlanConnectionConfig(options, o.TmDistributedVxlanConnection.ID), http.MethodPut, config)
}

// Delete deletes a Distributed VXLAN Connection. Waits for task completion. Requires TM API 9.1.1+.
func (o *TmDistributedVxlanConnection) Delete(options ...TmOperationOptions) error {
	options = inheritTmOptions(o.options, options)
	c := tmDistributedVxlanConnectionConfig(options, o.TmDistributedVxlanConnection.ID)
	return deleteEntityById(&o.vcdClient.Client, c)
}

// DeleteAsync deletes a Distributed VXLAN Connection. Returns the tracking task without waiting. Requires TM API 9.1.1+.
func (o *TmDistributedVxlanConnection) DeleteAsync(options ...TmOperationOptions) (*Task, error) {
	options = inheritTmOptions(o.options, options)
	c := tmDistributedVxlanConnectionConfig(options, o.TmDistributedVxlanConnection.ID)
	return runOpenApiOperationAsync(&o.vcdClient.Client, c, http.MethodDelete, nil)
}
