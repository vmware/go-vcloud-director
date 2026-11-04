// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TmInfraPolicy manages a Tenant Manager InfraPolicy.
type TmInfraPolicy struct {
	TmInfraPolicy *types.TmInfraPolicy
	vcdClient     *VCDClient
	options       []TmOperationOptions
}

// GetAllTmInfraPolicies retrieves a paginated list of infra policies. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmInfraPolicies(queryParameters url.Values, options ...TmOperationOptions) ([]*TmInfraPolicy, error) {
	c := crudConfig{
		entityLabel:     "infraPolicies",
		endpoint:        types.OpenApiPathVcf + "infraPolicies",
		requiresTm:      true,
		minimumVersion:  "40.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmInfraPolicy](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	wrapped := make([]*TmInfraPolicy, len(result))
	for i := range result {
		wrapped[i] = &TmInfraPolicy{TmInfraPolicy: result[i], vcdClient: vcdClient, options: retainTmOptions(options)}
	}
	return wrapped, nil
}

// CreateTmInfraPolicy creates an infra policy. Waits for task completion. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmInfraPolicy(config *types.TmInfraPolicy, options ...TmOperationOptions) (*TmInfraPolicy, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmInfraPolicy config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "infraPolicies",
		endpoint:       types.OpenApiPathVcf + "infraPolicies",
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	result, err := createInnerEntity(&vcdClient.Client, c.withTmOptions(options), config)
	if err != nil {
		return nil, err
	}
	return &TmInfraPolicy{TmInfraPolicy: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// CreateTmInfraPolicyAsync creates an infra policy. Returns the tracking task without waiting. Requires TM
// API 40.1+.
func (vcdClient *VCDClient) CreateTmInfraPolicyAsync(config *types.TmInfraPolicy, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmInfraPolicy config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "infraPolicies",
		endpoint:       types.OpenApiPathVcf + "infraPolicies",
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// GetTmInfraPolicyById retrieves an infra policy. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmInfraPolicyById(urn string, options ...TmOperationOptions) (*TmInfraPolicy, error) {
	c := crudConfig{
		entityLabel:    "infraPolicies",
		endpoint:       types.OpenApiPathVcf + "infraPolicies/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{urn},
	}
	result, err := getInnerEntity[types.TmInfraPolicy](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmInfraPolicy{TmInfraPolicy: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// Update updates an infra policy. Requires TM API 40.1+.
func (o *TmInfraPolicy) Update(config *types.TmInfraPolicy, options ...TmOperationOptions) (*TmInfraPolicy, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "infraPolicies",
		endpoint:       types.OpenApiPathVcf + "infraPolicies/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmInfraPolicy.ID},
	}
	result, err := writeInnerEntity(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config)
	if err != nil {
		return nil, err
	}
	return &TmInfraPolicy{TmInfraPolicy: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// Delete deletes an infra policy. Waits for task completion. Requires TM API 40.1+.
func (o *TmInfraPolicy) Delete(force bool, options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "infraPolicies",
		endpoint:        types.OpenApiPathVcf + "infraPolicies/%s",
		requiresTm:      true,
		minimumVersion:  "40.1",
		endpointParams:  []string{o.TmInfraPolicy.ID},
		queryParameters: url.Values{"force": {strconv.FormatBool(force)}},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// DeleteAsync deletes an infra policy. Returns the tracking task without waiting. Requires TM API 40.1+.
func (o *TmInfraPolicy) DeleteAsync(force bool, options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "infraPolicies",
		endpoint:        types.OpenApiPathVcf + "infraPolicies/%s",
		requiresTm:      true,
		minimumVersion:  "40.1",
		endpointParams:  []string{o.TmInfraPolicy.ID},
		queryParameters: url.Values{"force": {strconv.FormatBool(force)}},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// GetAllVCenterPolicies retrieves vCenter infra policies for an infra policy. Requires TM API 40.1+.
func (o *TmInfraPolicy) GetAllVCenterPolicies(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmVCenterInfraPolicy, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "infraPolicies/vCenterInfraPolicies",
		endpoint:        types.OpenApiPathVcf + "infraPolicies/%s/vCenterInfraPolicies",
		requiresTm:      true,
		minimumVersion:  "40.1",
		endpointParams:  []string{o.TmInfraPolicy.ID},
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmVCenterInfraPolicy](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SyncToVCenters synchronizes infra policy to vCenters. Waits for task completion. Requires TM API 40.1+.
func (o *TmInfraPolicy) SyncToVCenters(options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "infraPolicies/syncToVCenters",
		endpoint:       types.OpenApiPathVcf + "infraPolicies/%s/syncToVCenters",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmInfraPolicy.ID},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, nil)
}

// SyncToVCentersAsync synchronizes infra policy to vCenters. Returns the tracking task without waiting.
// Requires TM API 40.1+.
func (o *TmInfraPolicy) SyncToVCentersAsync(options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "infraPolicies/syncToVCenters",
		endpoint:       types.OpenApiPathVcf + "infraPolicies/%s/syncToVCenters",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmInfraPolicy.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, nil)
}
