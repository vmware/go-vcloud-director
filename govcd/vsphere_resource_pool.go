/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net/url"
)

type ResourcePool struct {
	ResourcePool *types.ResourcePool
	vcenter      *VCenter
	client       *VCDClient
}

// GetAllResourcePools retrieves all resource pools for a given vCenter
func (vcenter VCenter) GetAllResourcePools(queryParams url.Values) ([]*ResourcePool, error) {
	client := vcenter.client.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointResourcePoolsBrowseAll
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vcenter.VSphereVCenter.VcId))
	if err != nil {
		return nil, err
	}

	retrieved := []*types.ResourcePool{{}}

	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParams, &retrieved, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting resource pool list: %s", err)
	}

	if len(retrieved) == 0 {
		return nil, nil
	}
	var returnList []*ResourcePool

	for _, r := range retrieved {
		newRp := r
		returnList = append(returnList, &ResourcePool{
			ResourcePool: newRp,
			vcenter:      &vcenter,
			client:       vcenter.client,
		})
	}
	return returnList, nil
}

// GetAvailableHardwareVersions finds the hardware versions of a given resource pool
// In addition to proper resource pools, this method also works for any entity that is retrieved as a resource pool,
// such as provider VDCs and Org VDCs
func (rp ResourcePool) GetAvailableHardwareVersions() (*types.OpenApiSupportedHardwareVersions, error) {

	client := rp.client.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointResourcePoolHardware
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, rp.vcenter.VSphereVCenter.VcId, rp.ResourcePool.Moref))
	if err != nil {
		return nil, err
	}

	retrieved := types.OpenApiSupportedHardwareVersions{}
	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, &retrieved, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting resource pool hardware versions: %s", err)
	}

	return &retrieved, nil
}

// GetDefaultHardwareVersion retrieves the default hardware version for a given resource pool.
// The default version is usually the highest available, but it's not guaranteed
func (rp ResourcePool) GetDefaultHardwareVersion() (string, error) {

	versions, err := rp.GetAvailableHardwareVersions()
	if err != nil {
		return "", err
	}

	for _, v := range versions.SupportedVersions {
		if v.IsDefault {
			return v.Name, nil
		}
	}
	return "", fmt.Errorf("no default hardware version found for resource pool %s", rp.ResourcePool.Name)
}

// GetResourcePoolById retrieves a resource pool by its ID (Moref)
func (vcenter VCenter) GetResourcePoolById(id string) (*ResourcePool, error) {
	resourcePools, err := vcenter.GetAllResourcePools(nil)
	if err != nil {
		return nil, err
	}
	for _, rp := range resourcePools {
		if rp.ResourcePool.Moref == id {
			return rp, nil
		}
	}
	return nil, fmt.Errorf("no resource pool found with ID '%s' :%s", id, ErrorEntityNotFound)
}

// GetResourcePoolByName retrieves a resource pool by name.
// It may fail if there are several resource pools with the same name
func (vcenter VCenter) GetResourcePoolByName(name string) (*ResourcePool, error) {
	resourcePools, err := vcenter.GetAllResourcePools(nil)
	if err != nil {
		return nil, err
	}
	var found []*ResourcePool
	for _, rp := range resourcePools {
		if rp.ResourcePool.Name == name {
			found = append(found, rp)
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("no resource pool found with name '%s' :%s", name, ErrorEntityNotFound)
	}
	if len(found) > 1 {
		var idList []string
		for _, f := range found {
			idList = append(idList, f.ResourcePool.Moref)
		}
		return nil, fmt.Errorf("more than one resource pool was found with name %s - use resource pool ID instead - %v", name, idList)
	}
	return found[0], nil
}

// GetAllResourcePools retrieves all available resource pool, across all vCenters
func (vcdClient *VCDClient) GetAllResourcePools(queryParams url.Values) ([]*ResourcePool, error) {

	vcenters, err := vcdClient.GetAllVCenters(queryParams)
	if err != nil {
		return nil, err
	}
	var result []*ResourcePool
	for _, vc := range vcenters {
		resourcePools, err := vc.GetAllResourcePools(queryParams)
		if err != nil {
			return nil, err
		}
		result = append(result, resourcePools...)
	}
	return result, nil
}

// ResourcePoolsFromIds returns a slice of resource pools from a slice of resource pool IDs
func (vcdClient *VCDClient) ResourcePoolsFromIds(resourcePoolIds []string) ([]*ResourcePool, error) {
	if len(resourcePoolIds) == 0 {
		return nil, nil
	}

	var result []*ResourcePool

	// 1. make sure there are no duplicates in the input IDs
	uniqueIds := make(map[string]bool)
	var duplicates []string
	for _, id := range resourcePoolIds {
		_, seen := uniqueIds[id]
		if seen {
			duplicates = append(duplicates, id)
		}
		uniqueIds[id] = true
	}

	if len(duplicates) > 0 {
		return nil, fmt.Errorf("duplicate IDs found in input: %v", duplicates)
	}

	// 2. get all resource pools
	resourcePools, err := vcdClient.GetAllResourcePools(nil)
	if err != nil {
		return nil, err
	}

	util.Logger.Printf("wantedRecords: %v\n", resourcePoolIds)
	// 3. build a map of resource pools, indexed by ID, for easy search
	var foundRecords = make(map[string]*ResourcePool)

	for _, rpr := range resourcePools {
		foundRecords[rpr.ResourcePool.Moref] = rpr
	}

	// 4. loop through the requested IDs
	for wanted := range uniqueIds {
		// 4.1 if the wanted ID is not found, exit with an error
		foundResourcePool, ok := foundRecords[wanted]
		if !ok {
			return nil, fmt.Errorf("resource pool ID '%s' not found in VCD", wanted)
		}
		result = append(result, foundResourcePool)
	}

	// 5. Check that we got as many resource pools as the requested IDs
	if len(result) != len(uniqueIds) {
		return result, fmt.Errorf("%d IDs were requested, but only %d found", len(uniqueIds), len(result))
	}

	return result, nil
}
