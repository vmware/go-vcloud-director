// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TmVCenterComputePolicy manages a Tenant Manager VCenterComputePolicy.
type TmVCenterComputePolicy struct {
	TmVCenterComputePolicy *types.TmVCenterComputePolicy
	vcdClient              *VCDClient
	options                []TmOperationOptions
}

// GetAllTmVCenterComputePolicies retrieves a paginated list of vCenter compute policies. Requires TM API
// 40.1+.
func (vcdClient *VCDClient) GetAllTmVCenterComputePolicies(queryParameters url.Values, options ...TmOperationOptions) ([]*TmVCenterComputePolicy, error) {
	c := crudConfig{
		entityLabel:     "vCenterComputePolicies",
		endpoint:        types.OpenApiPathVcf + "vCenterComputePolicies",
		requiresTm:      true,
		minimumVersion:  "40.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmVCenterComputePolicy](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	wrapped := make([]*TmVCenterComputePolicy, len(result))
	for i := range result {
		wrapped[i] = &TmVCenterComputePolicy{TmVCenterComputePolicy: result[i], vcdClient: vcdClient, options: retainTmOptions(options)}
	}
	return wrapped, nil
}

// GetTmVCenterComputePolicyById retrieves a vCenter compute policy. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmVCenterComputePolicyById(urn string, options ...TmOperationOptions) (*TmVCenterComputePolicy, error) {
	c := crudConfig{
		entityLabel:    "vCenterComputePolicies",
		endpoint:       types.OpenApiPathVcf + "vCenterComputePolicies/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{urn},
	}
	result, err := getInnerEntity[types.TmVCenterComputePolicy](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmVCenterComputePolicy{TmVCenterComputePolicy: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// GetAllTags retrieves tags for a vCenter compute policy. Requires TM API 40.1+.
func (o *TmVCenterComputePolicy) GetAllTags(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmVCenterComputePolicyTag, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "vCenterComputePolicies/tags",
		endpoint:        types.OpenApiPathVcf + "vCenterComputePolicies/%s/tags",
		requiresTm:      true,
		minimumVersion:  "40.1",
		endpointParams:  []string{o.TmVCenterComputePolicy.ID},
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmVCenterComputePolicyTag](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAllTmVCenterComputePolicyNames retrieves a list of unique vCenter compute policy names. Requires TM
// API 40.1+.
func (vcdClient *VCDClient) GetAllTmVCenterComputePolicyNames(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmVCenterComputePolicyName, error) {
	c := crudConfig{
		entityLabel:     "vcenterComputePolicyNames",
		endpoint:        types.OpenApiPathVcf + "vcenterComputePolicyNames",
		requiresTm:      true,
		minimumVersion:  "40.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmVCenterComputePolicyName](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return result, nil
}
