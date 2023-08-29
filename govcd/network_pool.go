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

// CreateNetworkPool creates a new network pool using the given configuration
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
// The function retrieves the given NSX-T manager and corresponding transport zone
// If the trasport zone name is empty, the first available will be used
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
		return nil, fmt.Errorf("error retrieving transport xones for manager '%s': %s", manager.Name, err)
	}
	var transportZone *types.TransportZone
	for _, tz := range transportZones {
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

	var config = &types.NetworkPool{
		Name:        name,
		Description: description,
		PoolType:    "GENEVE",
		ManagingOwnerRef: types.OpenApiReference{
			Name: managers[0].Name,
			ID:   managerId,
		},
		Backing: types.NetworkPoolBacking{
			TransportZoneRef: types.OpenApiReference{
				ID:   transportZone.Id,
				Name: transportZone.Name,
			},
			ProviderRef: types.OpenApiReference{
				Name: manager.Name,
				ID:   managerId,
			},
		},
	}
	return vcdClient.CreateNetworkPool(config)
}
