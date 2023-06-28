/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

type ResourcePool struct {
	ResourcePool *types.ResourcePool
	vcenter      *VCenter
	client       *VCDClient
}

func (vcenter VCenter) GetAllResourcePools(queryParams url.Values) ([]*ResourcePool, error) {
	client := vcenter.client.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointResourcePoolsBrowseAll
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vcenter.VSphereVcenter.VcId))
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

func (rp ResourcePool) GetAvailableHardwareVersions() (*types.OpenApiSupportedHardwareVersions, error) {

	client := rp.client.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointResourcePoolHardware
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, rp.vcenter.VSphereVcenter.VcId, rp.ResourcePool.Moref))
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

func getClusterNameFromId(id string, rpList []*ResourcePool) (string, error) {
	for _, rp := range rpList {
		if rp.ResourcePool.Moref == id {
			return rp.ResourcePool.Name, nil
		}
	}
	return "", fmt.Errorf("no Cluster found with ID '%s'", id)
}

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
