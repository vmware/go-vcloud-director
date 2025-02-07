package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/url"
	"strings"
)

const labelTmOrgVdcStoragePolicies = "TM Org Vdc Storage Policies"

// TmVdcStoragePolicy defines Tenant Manager Virtual Datacenter Storage Policy structure
type TmVdcStoragePolicy struct {
	VirtualDatacenterStoragePolicy *types.VirtualDatacenterStoragePolicy
	vcdClient                      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmVdcStoragePolicy) wrap(inner *types.VirtualDatacenterStoragePolicy) *TmVdcStoragePolicy {
	g.VirtualDatacenterStoragePolicy = inner
	return &g
}

// CreateStoragePolicies creates new VDC Storage Policies in a VDC.
// The request will fail if the list of VDC Storage Policies is empty.
// It returns the list of all VCD Storage Policies that are available in the VDC after creation.
func (o *OrgRegionQuota) CreateStoragePolicies(regionStoragePolicies *types.VirtualDatacenterStoragePolicies) ([]*TmVdcStoragePolicy, error) {
	c := crudConfig{
		entityLabel: labelTmOrgVdcStoragePolicies,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		requiresTm:  true,
	}

	_, err := createInnerEntity[types.VirtualDatacenterStoragePolicies](&o.vcdClient.Client, c, regionStoragePolicies)
	if err != nil {
		// TODO: TM: The returned task contains a wrong URN in the "Owner" field, so the VDC can't be retrieved.
		//           We don't really need it either, so we ignore this error.
		if !strings.Contains(err.Error(), "error retrieving item after creation") {
			return nil, err
		}
	}

	// Get all the storage policies from the VDC and return them
	allPolicies, err := o.GetAllStoragePolicies(nil)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve all the VDC Storage Policies after creation: %s", err)
	}
	return allPolicies, nil
}

// GetAllTmVdcStoragePolicies retrieves all Tenant Manager VDC Storage Policies
func (vcdClient *VCDClient) GetAllTmVdcStoragePolicies(queryParameters url.Values) ([]*TmVdcStoragePolicy, error) {
	c := crudConfig{
		entityLabel:     labelTmOrgVdcStoragePolicies,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmVdcStoragePolicy{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetAllStoragePolicies retrieves all VDC Storage Policies from the given VDC
func (vdc *OrgRegionQuota) GetAllStoragePolicies(queryParameters url.Values) ([]*TmVdcStoragePolicy, error) {
	params := queryParameterFilterAnd("virtualDatacenter.id=="+vdc.OrgRegionQuota.ID, queryParameters)
	return vdc.vcdClient.GetAllTmVdcStoragePolicies(params)
}

// GetTmVdcStoragePolicyByName retrieves a VDC Storage Policy by the given name
func (vdc *OrgRegionQuota) GetTmVdcStoragePolicyByName(name string) (*TmVdcStoragePolicy, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name to be present", labelTmOrgVdcStoragePolicies)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("virtualDatacenter.id==%s;name==%s", vdc.OrgRegionQuota.ID, name))

	filteredEntities, err := vdc.vcdClient.GetAllTmVdcStoragePolicies(queryParams)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetTmVdcStoragePolicyById retrieves a VDC Storage Policy by a given ID
func (vcdClient *VCDClient) GetTmVdcStoragePolicyById(id string) (*TmVdcStoragePolicy, error) {
	c := crudConfig{
		entityLabel:    labelTmOrgVdcStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmVdcStoragePolicy{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetStoragePolicyById retrieves a Tenant Manager VDC Storage Policy by a given ID that must belong
// to the receiver VDC
func (vdc *OrgRegionQuota) GetStoragePolicyById(id string) (*TmVdcStoragePolicy, error) {
	policy, err := vdc.vcdClient.GetTmVdcStoragePolicyById(id)
	if err != nil {
		return nil, err
	}
	if policy.VirtualDatacenterStoragePolicy.VirtualDatacenter.ID != vdc.OrgRegionQuota.ID {
		return nil, fmt.Errorf("no VDC Storage Policy with ID '%s' found in VDC '%s': %s", id, vdc.OrgRegionQuota.ID, ErrorEntityNotFound)
	}
	return policy, err
}

// Update Tenant Manager VDC
func (vsp *TmVdcStoragePolicy) Update(tmVdcConfig *types.VirtualDatacenterStoragePolicy) (*TmVdcStoragePolicy, error) {
	c := crudConfig{
		entityLabel:    labelTmOrgVdcStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		endpointParams: []string{vsp.VirtualDatacenterStoragePolicy.ID},
		requiresTm:     true,
	}
	outerType := TmVdcStoragePolicy{vcdClient: vsp.vcdClient}
	return updateOuterEntity(&vsp.vcdClient.Client, outerType, c, tmVdcConfig)
}

// Delete deletes a VDC Storage Policy
func (vsp *TmVdcStoragePolicy) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmOrgVdcStoragePolicies,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies,
		endpointParams: []string{vsp.VirtualDatacenterStoragePolicy.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&vsp.vcdClient.Client, c)
}
