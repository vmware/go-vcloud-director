// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/url"
)

const labelTmVCenterComputePolicy = "TM vCenter Compute Policy"

// TmVCenterComputePolicy manages a vCenter compute policy.
type TmVCenterComputePolicy struct {
	TmVCenterComputePolicy *types.TmVCenterComputePolicy
	vcdClient              *VCDClient
	options                []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmVCenterComputePolicy) wrap(inner *types.TmVCenterComputePolicy) *TmVCenterComputePolicy {
	g.TmVCenterComputePolicy = inner
	return &g
}

// tmVCenterComputePolicyConfig returns the CRUD configuration for vCenter compute policy operations. Optional params are appended to the endpoint.
func tmVCenterComputePolicyConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmVCenterComputePolicy,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVCenterComputePolicies,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "40.1",
	}.withTmOptions(options)
}

// GetAllTmVCenterComputePolicies retrieves all vCenter compute policys. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmVCenterComputePolicies(queryParameters url.Values, options ...TmOperationOptions) ([]*TmVCenterComputePolicy, error) {
	c := tmVCenterComputePolicyConfig(options)
	c.queryParameters = queryParameters
	outerType := TmVCenterComputePolicy{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmVCenterComputePolicyByName retrieves a vCenter compute policy by name. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmVCenterComputePolicyByName(name string, options ...TmOperationOptions) (*TmVCenterComputePolicy, error) {
	entities, err := vcdClient.GetAllTmVCenterComputePolicies(nil, options...)
	if err != nil {
		return nil, err
	}
	return tmEntityByName(labelTmVCenterComputePolicy, name, entities, func(e *TmVCenterComputePolicy) string { return e.TmVCenterComputePolicy.Name })
}

// GetTmVCenterComputePolicyById retrieves a vCenter compute policy by ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmVCenterComputePolicyById(id string, options ...TmOperationOptions) (*TmVCenterComputePolicy, error) {
	outerType := TmVCenterComputePolicy{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmVCenterComputePolicyConfig(options, id))
}

// GetTmVCenterComputePolicyByNameOrId retrieves a vCenter compute policy by name or ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmVCenterComputePolicyByNameOrId(identifier string, options ...TmOperationOptions) (*TmVCenterComputePolicy, error) {
	return tmEntityByNameOrId(identifier,
		func(id string) (*TmVCenterComputePolicy, error) {
			return vcdClient.GetTmVCenterComputePolicyById(id, options...)
		},
		func(name string) (*TmVCenterComputePolicy, error) {
			return vcdClient.GetTmVCenterComputePolicyByName(name, options...)
		})
}

// GetAllTags retrieves the tags of a vCenter compute policy. Requires TM API 40.1+.
func (o *TmVCenterComputePolicy) GetAllTags(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmVCenterComputePolicyTag, error) {
	options = inheritTmOptions(o.options, options)
	c := tmCrudConfig("TM vCenter Compute Policy tag", types.OpenApiEndpointTmVCenterComputePolicyTags, "40.1", options, o.TmVCenterComputePolicy.ID)
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmVCenterComputePolicyTag](&o.vcdClient.Client, c)
}

// GetAllTmVCenterComputePolicyNames retrieves a list of unique vCenter compute policy names. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmVCenterComputePolicyNames(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmVCenterComputePolicyName, error) {
	c := tmCrudConfig("TM vCenter Compute Policy name", types.OpenApiEndpointTmVCenterComputePolicyNames, "40.1", options)
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmVCenterComputePolicyName](&vcdClient.Client, c)
}
