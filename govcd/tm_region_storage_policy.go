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

const labelRegionStoragePolicy = "Region Storage Policy"

// RegionStoragePolicy defines the Region Storage Policy data structure
type RegionStoragePolicy struct {
	RegionStoragePolicy *types.RegionStoragePolicy
	vcdClient           *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g RegionStoragePolicy) wrap(inner *types.RegionStoragePolicy) *RegionStoragePolicy {
	g.RegionStoragePolicy = inner
	return &g
}

// GetAllRegionStoragePolicies retrieves all Region Storage Policies with the given query parameters, which allow setting filters
// and other constraints
func (vcdClient *VCDClient) GetAllRegionStoragePolicies(queryParameters url.Values) ([]*RegionStoragePolicy, error) {
	c := crudConfig{
		entityLabel:     labelRegionStoragePolicy,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		queryParameters: queryParameters,
	}

	outerType := RegionStoragePolicy{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetStoragePolicyByName retrieves a Region Storage Policy by name, that belongs to the given Region
func (r *Region) GetStoragePolicyByName(name string) (*RegionStoragePolicy, error) {
	return getStoragePolicyByName(r.vcdClient, name, r.Region.ID)
}

// GetRegionStoragePolicyByName retrieves a Region Storage Policy by name
func (vcdClient *VCDClient) GetRegionStoragePolicyByName(name string) (*RegionStoragePolicy, error) {
	return getStoragePolicyByName(vcdClient, name, "")
}

func getStoragePolicyByName(vcdClient *VCDClient, name, regionId string) (*RegionStoragePolicy, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelRegionStoragePolicy)
	}
	filter := fmt.Sprintf("(name==%s)", name)
	if regionId != "" {
		filter = fmt.Sprintf("(name==%s;region.id==%s)", name, regionId)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", filter)

	filteredEntities, err := vcdClient.GetAllRegionStoragePolicies(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetRegionStoragePolicyById(singleEntity.RegionStoragePolicy.ID)
}

// GetRegionStoragePolicyById retrieves a Region Storage Policy by ID
func (vcdClient *VCDClient) GetRegionStoragePolicyById(id string) (*RegionStoragePolicy, error) {
	c := crudConfig{
		entityLabel:    labelRegionStoragePolicy,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		endpointParams: []string{id},
	}

	outerType := RegionStoragePolicy{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// CreateRegionStoragePolicy creates a policy, waits for completion and retrieves it by name and region.
// The creation task belongs to the region, so its owner ID cannot be used as a policy ID.
// Requires TM API 40.0 or later.
func (vcdClient *VCDClient) CreateRegionStoragePolicy(config *types.RegionStoragePolicy, options ...TmOperationOptions) (*RegionStoragePolicy, error) {
	if config == nil || config.Name == "" || config.Region == nil || config.Region.ID == "" {
		return nil, fmt.Errorf("creating a region storage policy requires a name and region ID")
	}
	task, err := vcdClient.CreateRegionStoragePolicyAsync(config, options...)
	if err != nil {
		return nil, err
	}
	if err := task.WaitTaskCompletion(); err != nil {
		return nil, err
	}
	c := crudConfig{
		entityLabel:     types.OpenApiEndpointRegionStoragePolicies,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		queryParameters: url.Values{"filter": {"region.id==" + config.Region.ID}},
		requiresTm:      true,
		minimumVersion:  "40.0",
	}
	policies, err := getAllInnerEntities[types.RegionStoragePolicy](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	// Match names literally: FIQL operators, escape characters and encoded literals
	// in a valid vCenter storage-policy name must not alter the lookup expression.
	var matchingPolicies []*types.RegionStoragePolicy
	for _, policy := range policies {
		if policy.Name == config.Name {
			matchingPolicies = append(matchingPolicies, policy)
		}
	}
	policy, err := oneOrError("name and region", config.Name, matchingPolicies)
	if err != nil {
		return nil, err
	}
	c.queryParameters = nil
	result, err := getInnerEntity[types.RegionStoragePolicy](&vcdClient.Client, c.item(policy.ID).withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &RegionStoragePolicy{RegionStoragePolicy: result, vcdClient: vcdClient}, nil
}

// CreateRegionStoragePolicyAsync creates a policy and returns its tracking task (API 40.0+).
func (vcdClient *VCDClient) CreateRegionStoragePolicyAsync(config *types.RegionStoragePolicy, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("region storage policy config must not be nil")
	}
	c := crudConfig{
		entityLabel:    types.OpenApiEndpointRegionStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// Update updates a region storage policy's isRestricted flag and returns the result (API 9.1.2+).
func (o *RegionStoragePolicy) Update(config *types.RegionStoragePolicy, options ...TmOperationOptions) (*RegionStoragePolicy, error) {
	if config == nil {
		return nil, fmt.Errorf("region storage policy config must not be nil")
	}
	// The resource is identified by the URL. Do not send the read-only body ID:
	// some servers validate it against a differently cased URN type than they return.
	payload := *config
	payload.ID = ""
	c := crudConfig{
		entityLabel:    types.OpenApiEndpointRegionStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		requiresTm:     true,
		minimumVersion: "9.1.2",
	}
	result, err := updateInnerEntity(&o.vcdClient.Client, c.item(o.RegionStoragePolicy.ID).withTmOptions(options), &payload)
	if err != nil {
		return nil, err
	}
	return &RegionStoragePolicy{RegionStoragePolicy: result, vcdClient: o.vcdClient}, nil
}

// Delete deletes a region storage policy and waits for completion (API 40.0+).
func (o *RegionStoragePolicy) Delete(options ...TmOperationOptions) error {
	c := crudConfig{
		entityLabel:    types.OpenApiEndpointRegionStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	return runOpenApiOperation(&o.vcdClient.Client, c.item(o.RegionStoragePolicy.ID).withTmOptions(options), http.MethodDelete, nil)
}
