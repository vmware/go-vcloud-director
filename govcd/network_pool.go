/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

type NetworkPool struct {
	NetworkPool *types.NetworkPool
	vcdClient   *VCDClient
}

// GetOpenApiUrl retrieves the full URL of a network pool
func (np *NetworkPool) GetOpenApiUrl() (string, error) {
	response, err := url.JoinPath(np.vcdClient.sessionHREF.String(), "admin", "extension", "networkPool", np.NetworkPool.Id)
	if err != nil {
		return "", err
	}
	return response, nil
}

// GetNetworkPoolSummaries retrieves the list of all available network pools
func (vcdClient *VCDClient) GetNetworkPoolSummaries(queryParameters url.Values) ([]*types.NetworkPool, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPoolSummaries
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	typeResponse := []*types.NetworkPool{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	return typeResponse, nil
}

// GetNetworkPoolById retrieves Network Pool with a given ID
func (vcdClient *VCDClient) GetNetworkPoolById(id string) (*NetworkPool, error) {
	if id == "" {
		return nil, fmt.Errorf("network pool lookup requires ID")
	}

	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	response := &NetworkPool{
		vcdClient:   vcdClient,
		NetworkPool: &types.NetworkPool{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, response.NetworkPool, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetNetworkPoolByName retrieves a network pool with a given name
// Note. It will return an error if multiple network pools exist with the same name
func (vcdClient *VCDClient) GetNetworkPoolByName(name string) (*NetworkPool, error) {
	if name == "" {
		return nil, fmt.Errorf("network pool lookup requires name")
	}

	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	filteredNetworkPools, err := vcdClient.GetNetworkPoolSummaries(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error getting network pools: %s", err)
	}

	if len(filteredNetworkPools) == 0 {
		return nil, fmt.Errorf("no network pool found with name '%s' - %s", name, ErrorEntityNotFound)
	}

	if len(filteredNetworkPools) > 1 {
		return nil, fmt.Errorf("more than one network pool found with name '%s'", name)
	}

	return vcdClient.GetNetworkPoolById(filteredNetworkPools[0].Id)
}

// CreateNetworkPool creates a network pool using the given configuration
// It can create any type of network pool
func (vcdClient *VCDClient) CreateNetworkPool(config *types.NetworkPool) (*NetworkPool, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &NetworkPool{
		NetworkPool: &types.NetworkPool{},
		vcdClient:   vcdClient,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, config, result.NetworkPool, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update will change all changeable network pool items
func (np *NetworkPool) Update() error {
	if np == nil || np.NetworkPool == nil || np.NetworkPool.Id == "" {
		return fmt.Errorf("network pool must have ID")
	}
	if np.vcdClient == nil || np.vcdClient.Client.APIVersion == "" {
		return fmt.Errorf("network pool '%s': no client found", np.NetworkPool.Name)
	}

	client := np.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, np.NetworkPool.Id)
	if err != nil {
		return err
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, np.NetworkPool, np.NetworkPool, nil)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("error updating network pool '%s': %s", np.NetworkPool.Name, err)
	}

	return nil
}

// Delete removes a network pool
func (np *NetworkPool) Delete() error {
	if np == nil || np.NetworkPool == nil || np.NetworkPool.Id == "" {
		return fmt.Errorf("network pool must have ID")
	}
	if np.vcdClient == nil || np.vcdClient.Client.APIVersion == "" {
		return fmt.Errorf("network pool '%s': no client found", np.NetworkPool.Name)
	}

	client := np.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, np.NetworkPool.Id)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("error deleting network pool '%s': %s", np.NetworkPool.Name, err)
	}

	return nil
}

// CreateNetworkPoolGeneve creates a network pool of GENEVE type
// The function retrieves the given NSX-T manager and corresponding transport zone names
// If the transport zone name is empty, the first available will be used
func (vcdClient *VCDClient) CreateNetworkPoolGeneve(name, description, managerName, transportZoneName string) (*NetworkPool, error) {
	managers, err := vcdClient.QueryNsxtManagerByName(managerName)
	if err != nil {
		return nil, err
	}

	if len(managers) == 0 {
		return nil, fmt.Errorf("no manager '%s' found", managerName)
	}
	if len(managers) > 1 {
		return nil, fmt.Errorf("more than one manager '%s' found", managerName)
	}
	manager := managers[0]

	managerId := "urn:vcloud:nsxtmanager:" + extractUuid(managers[0].HREF)
	transportZones, err := vcdClient.GetAllNsxtTransportZones(managerId, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving transport zones for manager '%s': %s", manager.Name, err)
	}
	var transportZone *types.TransportZone
	for _, tz := range transportZones {
		// if the transport zone name was empty, we take the first available
		// otherwise, we take the wanted transport zone
		if (transportZoneName == "" && !tz.AlreadyImported) || tz.Name == transportZoneName {
			transportZone = tz
			break
		}
	}

	if transportZone == nil {
		if transportZoneName == "" {
			return nil, fmt.Errorf("no unimported transport zone was found")
		}
		return nil, fmt.Errorf("transport zone '%s' not found", transportZoneName)
	}

	if transportZone.AlreadyImported {
		return nil, fmt.Errorf("transport zone '%s' is already imported", transportZone.Name)
	}

	// Note: in this type of network pool, the managing owner is the NSX-T manager
	managingOwner := types.OpenApiReference{
		Name: manager.Name,
		ID:   managerId,
	}
	var config = &types.NetworkPool{
		Name:             name,
		Description:      description,
		PoolType:         types.NetworkPoolGeneveType,
		ManagingOwnerRef: managingOwner,
		Backing: types.NetworkPoolBacking{
			TransportZoneRef: types.OpenApiReference{
				ID:   transportZone.Id,
				Name: transportZone.Name,
			},
			ProviderRef: managingOwner,
		},
	}
	return vcdClient.CreateNetworkPool(config)
}

// CreateNetworkPoolPortGroup creates a network pool of PORTGROUP_BACKED type
// The function retrieves the given vCenter and corresponding port group names
// If the port group name is empty, the first available will be used
func (vcdClient *VCDClient) CreateNetworkPoolPortGroup(name, description, vCenterName, portgroupName string) (*NetworkPool, error) {
	vCenter, err := vcdClient.GetVCenterByName(vCenterName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vCenter '%s': %s", vCenterName, err)
	}
	var params = make(url.Values)
	params.Set("virtualCenter.id", vCenter.VSphereVCenter.VcId)
	portgroups, err := vcdClient.GetAllVcenterImportableDvpgs(params)
	if err != nil {
		return nil, fmt.Errorf("error retrieving portgroups for vCenter '%s': %s", vCenterName, err)
	}
	var portgroup *VcenterImportableDvpg
	for _, pg := range portgroups {
		// If the port group name was empty, we take the first available
		// otherwise, we take the wanted one
		if portgroupName == "" || portgroupName == pg.VcenterImportableDvpg.BackingRef.Name {
			portgroup = pg
			break
		}
	}
	if portgroup == nil {
		if portgroupName == "" {
			return nil, fmt.Errorf("no available port groups found in vCenter '%s'", vCenterName)
		}
		return nil, fmt.Errorf("port group '%s' not found in vCenter '%s", portgroupName, vCenterName)
	}

	// Note: in this type of network pool, the managing owner is the vCenter
	managingOwner := types.OpenApiReference{
		Name: vCenter.VSphereVCenter.Name,
		ID:   vCenter.VSphereVCenter.VcId,
	}
	config := types.NetworkPool{
		Name:             name,
		Description:      description,
		PoolType:         types.NetworkPoolPortGroupType,
		ManagingOwnerRef: managingOwner,
		Backing: types.NetworkPoolBacking{
			PortGroupRefs: []types.OpenApiReference{
				{
					ID:   portgroup.VcenterImportableDvpg.BackingRef.ID,
					Name: portgroup.VcenterImportableDvpg.BackingRef.Name,
				},
			},
			ProviderRef: managingOwner,
		},
	}
	return vcdClient.CreateNetworkPool(&config)
}

// CreateNetworkPoolVlan creates a network pool of VLAN type
// The function retrieves the given vCenter and corresponding distributed switch names
// If the distributed switch name is empty, the first available will be used
func (vcdClient *VCDClient) CreateNetworkPoolVlan(name, description, vCenterName, dsName string, ranges []types.VlanIdRange) (*NetworkPool, error) {
	vCenter, err := vcdClient.GetVCenterByName(vCenterName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vCenter '%s': %s", vCenterName, err)
	}
	var params = make(url.Values)
	params.Set("virtualCenter.id", vCenter.VSphereVCenter.VcId)

	dswitches, err := vcdClient.GetAllVcenterDistributedSwitches(vCenter.VSphereVCenter.VcId, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving distributed switches for vCenter '%s': %s", vCenterName, err)
	}
	var dswitch *types.VcenterDistributedSwitch
	for _, dsw := range dswitches {
		// If the distributed switch name was empty, we take the first available
		// otherwise, we take the wanted one
		if dsName == "" || dsName == dsw.BackingRef.Name {
			dswitch = dsw
			break
		}
	}
	if dswitch == nil {
		if dsName == "" {
			return nil, fmt.Errorf("no available distributed switches found in vCenter '%s'", vCenterName)
		}
		return nil, fmt.Errorf("distributed switch '%s' not found in vCenter '%s", dsName, vCenterName)
	}

	// Note: in this type of network pool, the managing owner is the vCenter
	managingOwner := types.OpenApiReference{
		Name: vCenter.VSphereVCenter.Name,
		ID:   vCenter.VSphereVCenter.VcId,
	}
	config := types.NetworkPool{
		Name:             name,
		Description:      description,
		PoolType:         types.NetworkPoolVlanType,
		ManagingOwnerRef: managingOwner,
		Backing: types.NetworkPoolBacking{
			VlanIdRanges: types.VlanIdRanges{
				Values: ranges,
			},
			VdsRefs: []types.OpenApiReference{
				{
					Name: dswitch.BackingRef.Name,
					ID:   dswitch.BackingRef.ID,
				},
			},
			ProviderRef: managingOwner,
		},
	}
	return vcdClient.CreateNetworkPool(&config)
}
