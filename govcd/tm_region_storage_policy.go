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
	client              *Client
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g RegionStoragePolicy) wrap(inner *types.RegionStoragePolicy) *RegionStoragePolicy {
	g.RegionStoragePolicy = inner
	return &g
}

// GetAllStoragePolicies retrieves all Region Storage Policies with the given query parameters, which allow setting filters
// and other constraints
func (r *Region) GetAllStoragePolicies(queryParameters url.Values) ([]*RegionStoragePolicy, error) {
	return getAllStoragePolicies(r, queryParameters)
}
func getAllStoragePolicies(r *Region, queryParameters url.Values) ([]*RegionStoragePolicy, error) {
	c := crudConfig{
		entityLabel:     labelRegionStoragePolicy,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		queryParameters: queryParameters,
	}

	outerType := RegionStoragePolicy{client: &r.vcdClient.Client}
	return getAllOuterEntities(&r.vcdClient.Client, outerType, c)
}

// GetStoragePolicyByName retrieves a Region Storage Policy by name
func (r *Region) GetStoragePolicyByName(name string) (*RegionStoragePolicy, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelRegionStoragePolicy)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("(name==%s;region.id==%s)", name, r.Region.ID))

	filteredEntities, err := getAllStoragePolicies(r, queryParams)
	if err != nil {
		return nil, err
	}

	// TODO: TM: API returns same result twice for some reason
	if len(filteredEntities) == 0 {
		return nil, fmt.Errorf("TODO: TM: found 0 storage policies: %s", ErrorEntityNotFound)
	}
	singleEntity := filteredEntities[0]
	//singleEntity, err := oneOrError("name", name, filteredEntities)
	//if err != nil {
	//	return nil, err
	//}

	return getStoragePolicyById(&r.vcdClient.Client, singleEntity.RegionStoragePolicy.ID)
}

// GetStoragePolicyById retrieves a Region Storage Policy by ID
func (r *Region) GetStoragePolicyById(id string) (*RegionStoragePolicy, error) {
	return getStoragePolicyById(&r.vcdClient.Client, id)
}

// GetRegionStoragePolicyById retrieves a Region Storage Policy by ID
func (vcdClient *VCDClient) GetRegionStoragePolicyById(id string) (*RegionStoragePolicy, error) {
	return getStoragePolicyById(&vcdClient.Client, id)
}

func getStoragePolicyById(client *Client, id string) (*RegionStoragePolicy, error) {
	c := crudConfig{
		entityLabel:    labelRegionStoragePolicy,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		endpointParams: []string{id},
	}

	outerType := RegionStoragePolicy{client: client}
	return getOuterEntity(client, outerType, c)
}
