// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/http"
	"net/url"
	"strconv"
)

const labelTmInfraPolicy = "TM Infra Policy"

// TmInfraPolicy manages an infra policy.
type TmInfraPolicy struct {
	TmInfraPolicy *types.TmInfraPolicy
	vcdClient     *VCDClient
	options       []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmInfraPolicy) wrap(inner *types.TmInfraPolicy) *TmInfraPolicy {
	g.TmInfraPolicy = inner
	return &g
}

// tmInfraPolicyConfig returns the CRUD configuration for infra policy operations. Optional params are appended to the endpoint.
func tmInfraPolicyConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmInfraPolicy,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmInfraPolicies,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "40.1",
	}.withTmOptions(options)
}

// CreateTmInfraPolicy creates an infra policy. Waits for task completion. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmInfraPolicy(config *types.TmInfraPolicy, options ...TmOperationOptions) (*TmInfraPolicy, error) {
	outerType := TmInfraPolicy{vcdClient: vcdClient, options: retainTmOptions(options)}
	return createOuterEntity(&vcdClient.Client, outerType, tmInfraPolicyConfig(options), config)
}

// CreateTmInfraPolicyAsync creates an infra policy. Returns the tracking task without waiting. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmInfraPolicyAsync(config *types.TmInfraPolicy, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmInfraPolicy config must not be nil")
	}
	return createInnerEntityAsync(&vcdClient.Client, tmInfraPolicyConfig(options), config)
}

// GetAllTmInfraPolicies retrieves all infra policys. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmInfraPolicies(queryParameters url.Values, options ...TmOperationOptions) ([]*TmInfraPolicy, error) {
	c := tmInfraPolicyConfig(options)
	c.queryParameters = queryParameters
	outerType := TmInfraPolicy{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmInfraPolicyByName retrieves an infra policy by name. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmInfraPolicyByName(name string, options ...TmOperationOptions) (*TmInfraPolicy, error) {
	entities, err := vcdClient.GetAllTmInfraPolicies(nil, options...)
	if err != nil {
		return nil, err
	}
	return tmEntityByName(labelTmInfraPolicy, name, entities, func(e *TmInfraPolicy) string { return e.TmInfraPolicy.Name })
}

// GetTmInfraPolicyById retrieves an infra policy by ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmInfraPolicyById(id string, options ...TmOperationOptions) (*TmInfraPolicy, error) {
	outerType := TmInfraPolicy{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmInfraPolicyConfig(options, id))
}

// GetTmInfraPolicyByNameOrId retrieves an infra policy by name or ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmInfraPolicyByNameOrId(identifier string, options ...TmOperationOptions) (*TmInfraPolicy, error) {
	return tmEntityByNameOrId(identifier,
		func(id string) (*TmInfraPolicy, error) { return vcdClient.GetTmInfraPolicyById(id, options...) },
		func(name string) (*TmInfraPolicy, error) { return vcdClient.GetTmInfraPolicyByName(name, options...) })
}

// Update updates an infra policy. Waits for task completion. Requires TM API 40.1+.
func (o *TmInfraPolicy) Update(config *types.TmInfraPolicy, options ...TmOperationOptions) (*TmInfraPolicy, error) {
	options = inheritTmOptions(o.options, options)
	outerType := TmInfraPolicy{vcdClient: o.vcdClient, options: retainTmOptions(options)}
	return updateOuterEntity(&o.vcdClient.Client, outerType, tmInfraPolicyConfig(options, o.TmInfraPolicy.ID), config)
}

// Delete deletes an infra policy. Waits for task completion. Requires TM API 40.1+.
func (o *TmInfraPolicy) Delete(force bool, options ...TmOperationOptions) error {
	options = inheritTmOptions(o.options, options)
	c := tmInfraPolicyConfig(options, o.TmInfraPolicy.ID)
	c.queryParameters = url.Values{"force": {strconv.FormatBool(force)}}
	return deleteEntityById(&o.vcdClient.Client, c)
}

// DeleteAsync deletes an infra policy. Returns the tracking task without waiting. Requires TM API 40.1+.
func (o *TmInfraPolicy) DeleteAsync(force bool, options ...TmOperationOptions) (*Task, error) {
	options = inheritTmOptions(o.options, options)
	c := tmInfraPolicyConfig(options, o.TmInfraPolicy.ID)
	c.queryParameters = url.Values{"force": {strconv.FormatBool(force)}}
	return runOpenApiOperationAsync(&o.vcdClient.Client, c, http.MethodDelete, nil)
}

// GetAllVCenterPolicies retrieves the vCenter infra policies of an infra policy. Requires TM API 40.1+.
func (o *TmInfraPolicy) GetAllVCenterPolicies(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmVCenterInfraPolicy, error) {
	options = inheritTmOptions(o.options, options)
	c := tmCrudConfig("TM vCenter Infra Policy", types.OpenApiEndpointTmInfraPolicyVCenterPolicies, "40.1", options, o.TmInfraPolicy.ID)
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmVCenterInfraPolicy](&o.vcdClient.Client, c)
}

// SyncToVCenters synchronizes the infra policy to vCenters. Waits for task completion. Requires TM API 40.1+.
func (o *TmInfraPolicy) SyncToVCenters(options ...TmOperationOptions) error {
	options = inheritTmOptions(o.options, options)
	return runOpenApiOperation(&o.vcdClient.Client, tmCrudConfig("TM Infra Policy sync", types.OpenApiEndpointTmInfraPolicySyncToVCenters, "40.1", options, o.TmInfraPolicy.ID), http.MethodPost, nil)
}

// SyncToVCentersAsync synchronizes the infra policy to vCenters. Returns the tracking task without waiting. Requires TM API 40.1+.
func (o *TmInfraPolicy) SyncToVCentersAsync(options ...TmOperationOptions) (*Task, error) {
	options = inheritTmOptions(o.options, options)
	return runOpenApiOperationAsync(&o.vcdClient.Client, tmCrudConfig("TM Infra Policy sync", types.OpenApiEndpointTmInfraPolicySyncToVCenters, "40.1", options, o.TmInfraPolicy.ID), http.MethodPost, nil)
}
