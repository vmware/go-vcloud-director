package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
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
