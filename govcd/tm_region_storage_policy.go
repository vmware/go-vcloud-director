package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelRegionStoragePolicy = "Region Storage Policy"

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

func (vcdClient *VCDClient) CreateRegionStoragePolicy(config *types.RegionStoragePolicy) (*RegionStoragePolicy, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("err")
	}
	c := crudConfig{
		entityLabel: labelRegionStoragePolicy,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
	}
	outerType := RegionStoragePolicy{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

func (vcdClient *VCDClient) GetAllRegionStoragePolicies(queryParameters url.Values) ([]*RegionStoragePolicy, error) {
	c := crudConfig{
		entityLabel:     labelRegionStoragePolicy,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		queryParameters: queryParameters,
	}

	outerType := RegionStoragePolicy{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetRegionStoragePolicyByName(name string) (*RegionStoragePolicy, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelRegionStoragePolicy)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllRegionStoragePolicies(queryParams)
	if err != nil {
		return nil, err
	}

	// TODO: API returns same result twice for some reason
	if len(filteredEntities) == 0 {
		return nil, fmt.Errorf("TODO: found 0 storage policies")
	}
	singleEntity := filteredEntities[0]
	//singleEntity, err := oneOrError("name", name, filteredEntities)
	//if err != nil {
	//	return nil, err
	//}

	return vcdClient.GetRegionStoragePolicyById(singleEntity.RegionStoragePolicy.ID)
}

func (vcdClient *VCDClient) GetRegionStoragePolicyById(id string) (*RegionStoragePolicy, error) {
	c := crudConfig{
		entityLabel:    labelRegionStoragePolicy,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		endpointParams: []string{id},
	}

	outerType := RegionStoragePolicy{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (o *RegionStoragePolicy) Update(RegionStoragePolicyConfig *types.RegionStoragePolicy) (*RegionStoragePolicy, error) {
	c := crudConfig{
		entityLabel:    labelRegionStoragePolicy,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		endpointParams: []string{o.RegionStoragePolicy.ID},
	}
	outerType := RegionStoragePolicy{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, RegionStoragePolicyConfig)
}

func (o *RegionStoragePolicy) Delete() error {
	c := crudConfig{
		entityLabel:    labelRegionStoragePolicy,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
		endpointParams: []string{o.RegionStoragePolicy.ID},
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
