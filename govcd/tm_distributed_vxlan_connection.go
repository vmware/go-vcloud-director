// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TmDistributedVxlanConnection manages a Tenant Manager DistributedVxlanConnection.
type TmDistributedVxlanConnection struct {
	TmDistributedVxlanConnection *types.TmDistributedVxlanConnection
	vcdClient                    *VCDClient
	options                      []TmOperationOptions
}

// GetAllTmDistributedVxlanConnections retrieves all Distributed VXLAN Connections. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) GetAllTmDistributedVxlanConnections(queryParameters url.Values, options ...TmOperationOptions) ([]*TmDistributedVxlanConnection, error) {
	c := crudConfig{
		entityLabel:     "distributedVxlanConnections",
		endpoint:        types.OpenApiPathVcf + "distributedVxlanConnections",
		requiresTm:      true,
		minimumVersion:  "9.1.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmDistributedVxlanConnection](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	wrapped := make([]*TmDistributedVxlanConnection, len(result))
	for i := range result {
		wrapped[i] = &TmDistributedVxlanConnection{TmDistributedVxlanConnection: result[i], vcdClient: vcdClient, options: retainTmOptions(options)}
	}
	return wrapped, nil
}

// CreateTmDistributedVxlanConnection creates a new Distributed VXLAN Connection. Waits for task completion.
// Requires TM API 9.1.1+.
func (vcdClient *VCDClient) CreateTmDistributedVxlanConnection(config *types.TmDistributedVxlanConnection, options ...TmOperationOptions) (*TmDistributedVxlanConnection, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmDistributedVxlanConnection config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "distributedVxlanConnections",
		endpoint:       types.OpenApiPathVcf + "distributedVxlanConnections",
		requiresTm:     true,
		minimumVersion: "9.1.1",
	}
	result, err := createInnerEntity(&vcdClient.Client, c.withTmOptions(options), config)
	if err != nil {
		return nil, err
	}
	return &TmDistributedVxlanConnection{TmDistributedVxlanConnection: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// CreateTmDistributedVxlanConnectionAsync creates a new Distributed VXLAN Connection. Returns the tracking
// task without waiting. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) CreateTmDistributedVxlanConnectionAsync(config *types.TmDistributedVxlanConnection, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmDistributedVxlanConnection config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "distributedVxlanConnections",
		endpoint:       types.OpenApiPathVcf + "distributedVxlanConnections",
		requiresTm:     true,
		minimumVersion: "9.1.1",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// GetTmDistributedVxlanConnectionById retrieves a Distributed VXLAN Connection. Requires TM API 9.1.1+.
func (vcdClient *VCDClient) GetTmDistributedVxlanConnectionById(distributedVxlanConnectionId string, options ...TmOperationOptions) (*TmDistributedVxlanConnection, error) {
	c := crudConfig{
		entityLabel:    "distributedVxlanConnections",
		endpoint:       types.OpenApiPathVcf + "distributedVxlanConnections/%s",
		requiresTm:     true,
		minimumVersion: "9.1.1",
		endpointParams: []string{distributedVxlanConnectionId},
	}
	result, err := getInnerEntity[types.TmDistributedVxlanConnection](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmDistributedVxlanConnection{TmDistributedVxlanConnection: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// Update updates a Distributed VXLAN Connection. Waits for task completion. Requires TM API 9.1.1+.
func (o *TmDistributedVxlanConnection) Update(config *types.TmDistributedVxlanConnection, options ...TmOperationOptions) (*TmDistributedVxlanConnection, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "distributedVxlanConnections",
		endpoint:       types.OpenApiPathVcf + "distributedVxlanConnections/%s",
		requiresTm:     true,
		minimumVersion: "9.1.1",
		endpointParams: []string{o.TmDistributedVxlanConnection.ID},
	}
	if err := runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config); err != nil {
		return nil, err
	}
	result, err := getInnerEntity[types.TmDistributedVxlanConnection](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmDistributedVxlanConnection{TmDistributedVxlanConnection: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// UpdateAsync updates a Distributed VXLAN Connection. Returns the tracking task without waiting. Requires
// TM API 9.1.1+.
func (o *TmDistributedVxlanConnection) UpdateAsync(config *types.TmDistributedVxlanConnection, options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "distributedVxlanConnections",
		endpoint:       types.OpenApiPathVcf + "distributedVxlanConnections/%s",
		requiresTm:     true,
		minimumVersion: "9.1.1",
		endpointParams: []string{o.TmDistributedVxlanConnection.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config)
}

// Delete deletes a specific Distributed VXLAN Connection. Waits for task completion. Requires TM API
// 9.1.1+.
func (o *TmDistributedVxlanConnection) Delete(options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "distributedVxlanConnections",
		endpoint:       types.OpenApiPathVcf + "distributedVxlanConnections/%s",
		requiresTm:     true,
		minimumVersion: "9.1.1",
		endpointParams: []string{o.TmDistributedVxlanConnection.ID},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// DeleteAsync deletes a specific Distributed VXLAN Connection. Returns the tracking task without waiting.
// Requires TM API 9.1.1+.
func (o *TmDistributedVxlanConnection) DeleteAsync(options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "distributedVxlanConnections",
		endpoint:       types.OpenApiPathVcf + "distributedVxlanConnections/%s",
		requiresTm:     true,
		minimumVersion: "9.1.1",
		endpointParams: []string{o.TmDistributedVxlanConnection.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}
