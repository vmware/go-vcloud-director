// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/url"
	"strings"
)

const labelRegionQuotaStoragePolicies = "Region Quota Storage Policies"

// RegionQuotaStoragePolicy defines Tenant Manager Virtual Datacenter Storage Policy structure
type RegionQuotaStoragePolicy struct {
	VirtualDatacenterStoragePolicy *types.VirtualDatacenterStoragePolicy
	vcdClient                      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g RegionQuotaStoragePolicy) wrap(inner *types.VirtualDatacenterStoragePolicy) *RegionQuotaStoragePolicy {
	g.VirtualDatacenterStoragePolicy = inner
	return &g
}

// CreateRegionQuotaStoragePolicies creates new Region Quota Storage Policies in a Region Quota.
// The request will fail if the list of Storage Policies is empty.
// It returns the list of all Storage Policies that are available in the Region Quota after creation.
func (vcdClient *VCDClient) CreateRegionQuotaStoragePolicies(regionQuotaId string, regionStoragePolicies *types.VirtualDatacenterStoragePolicies) ([]*RegionQuotaStoragePolicy, error) {
	c := crudConfig{
		entityLabel: labelRegionQuotaStoragePolicies,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		requiresTm:  true,
	}

	_, err := createInnerEntity[types.VirtualDatacenterStoragePolicies](&vcdClient.Client, c, regionStoragePolicies)
	if err != nil {
		// TODO: TM: The returned task contains a wrong URN in the "Owner" field, so the VDC can't be retrieved.
		//           We don't really need it either, so we ignore this error.
		if !strings.Contains(err.Error(), "error retrieving item after creation") {
			return nil, err
		}
	}

	// Get all the storage policies from the Region Quota and return them
	allPolicies, err := vcdClient.GetAllRegionQuotaStoragePolicies(nil)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve all the %s after creation: %s", labelRegionQuotaStoragePolicies, err)
	}
	return allPolicies, nil
}

// CreateStoragePolicies creates new Region Quota Storage Policies in a Region Quota.
// The request will fail if the list of Storage Policies is empty.
// It returns the list of all Storage Policies that are available in the Region Quota after creation.
func (o *RegionQuota) CreateStoragePolicies(regionStoragePolicies *types.VirtualDatacenterStoragePolicies) ([]*RegionQuotaStoragePolicy, error) {
	return o.vcdClient.CreateRegionQuotaStoragePolicies(o.TmVdc.ID, regionStoragePolicies)
}

// GetAllRegionQuotaStoragePolicies retrieves all Region Quota Storage Policies
func (vcdClient *VCDClient) GetAllRegionQuotaStoragePolicies(queryParameters url.Values) ([]*RegionQuotaStoragePolicy, error) {
	c := crudConfig{
		entityLabel:     labelRegionQuotaStoragePolicies,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := RegionQuotaStoragePolicy{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetAllStoragePolicies retrieves all Region Quota Storage Policies from the given Region Quota
func (regionQuota *RegionQuota) GetAllStoragePolicies(queryParameters url.Values) ([]*RegionQuotaStoragePolicy, error) {
	params := queryParameterFilterAnd("virtualDatacenter.id=="+regionQuota.TmVdc.ID, queryParameters)
	return regionQuota.vcdClient.GetAllRegionQuotaStoragePolicies(params)
}

// GetStoragePolicyByName retrieves a Region Quota Storage Policy by the given name. This method runs in O(n)
// where n is the number of Storage Policies in the receiver Region Quota, as the endpoint does not support filtering.
func (regionQuota *RegionQuota) GetStoragePolicyByName(name string) (*RegionQuotaStoragePolicy, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name to be present", labelRegionQuotaStoragePolicies)
	}

	allSps, err := regionQuota.GetAllStoragePolicies(nil)
	if err != nil {
		return nil, err
	}
	for _, sp := range allSps {
		if sp.VirtualDatacenterStoragePolicy.Name == name {
			return sp, nil
		}
	}
	return nil, fmt.Errorf("unable to find storage policy with name '%s' after lookup: %s", name, ErrorEntityNotFound)
}

// GetRegionQuotaStoragePolicyById retrieves a Region Quota Storage Policy by a given ID
func (vcdClient *VCDClient) GetRegionQuotaStoragePolicyById(id string) (*RegionQuotaStoragePolicy, error) {
	c := crudConfig{
		entityLabel:    labelRegionQuotaStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := RegionQuotaStoragePolicy{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetStoragePolicyById retrieves a Region Quota Storage Policy by a given ID that must belong
// to the receiver Region Quota
func (rq *RegionQuota) GetStoragePolicyById(id string) (*RegionQuotaStoragePolicy, error) {
	policy, err := rq.vcdClient.GetRegionQuotaStoragePolicyById(id)
	if err != nil {
		return nil, err
	}
	if policy.VirtualDatacenterStoragePolicy.VirtualDatacenter.ID != rq.TmVdc.ID {
		return nil, fmt.Errorf("no %s with ID '%s' found in %s '%s': %s", labelRegionQuotaStoragePolicies, id, labelRegionQuota, rq.TmVdc.ID, ErrorEntityNotFound)
	}
	return policy, err
}

// UpdateRegionQuotaStoragePolicy updates the Region Quota Storage Policy with given ID
func (vcdClient *VCDClient) UpdateRegionQuotaStoragePolicy(regionQuotaStoragePolicyId string, tmVdcConfig *types.VirtualDatacenterStoragePolicy) (*RegionQuotaStoragePolicy, error) {
	c := crudConfig{
		entityLabel:    labelRegionQuotaStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		endpointParams: []string{regionQuotaStoragePolicyId},
		requiresTm:     true,
	}
	outerType := RegionQuotaStoragePolicy{vcdClient: vcdClient}
	return updateOuterEntity(&vcdClient.Client, outerType, c, tmVdcConfig)
}

// Update updates the receiver Region Quota Storage Policy
func (sp *RegionQuotaStoragePolicy) Update(spConfig *types.VirtualDatacenterStoragePolicy) (*RegionQuotaStoragePolicy, error) {
	return sp.vcdClient.UpdateRegionQuotaStoragePolicy(sp.VirtualDatacenterStoragePolicy.ID, spConfig)
}

// DeleteRegionQuotaStoragePolicy deletes a Region Quota Storage Policy with given ID
func (vcdClient *VCDClient) DeleteRegionQuotaStoragePolicy(regionQuotaStoragePolicyId string) error {
	c := crudConfig{
		entityLabel:    labelRegionQuotaStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		endpointParams: []string{regionQuotaStoragePolicyId},
		requiresTm:     true,
	}
	return deleteEntityById(&vcdClient.Client, c)
}

// Delete deletes a Region Quota Storage Policy
func (sp *RegionQuotaStoragePolicy) Delete() error {
	return sp.vcdClient.DeleteRegionQuotaStoragePolicy(sp.VirtualDatacenterStoragePolicy.ID)
}
