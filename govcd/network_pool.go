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

func NewNetworkPool(client *VCDClient) *NetworkPool {
	return &NetworkPool{
		NetworkPool: &types.NetworkPool{},
		vcdClient:   client,
	}
}

func (np NetworkPool) GetOpenApiUrl() (string, error) {
	response, err := url.JoinPath(np.vcdClient.sessionHREF.String(), "admin", "extension", "networkPool", np.NetworkPool.Id)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (vcdClient VCDClient) GetNetworkPoolSummaries(queryParameters url.Values) (*types.NetworkPoolSummary, error) {
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

	typeResponse := types.NetworkPoolSummary{}
	err = client.OpenApiGetItem(apiVersion, urlRef, queryParameters, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	return &typeResponse, nil
}

// GetNetworkPoolById retrieves IP Space with a given ID
func (vcdClient VCDClient) GetNetworkPoolById(id string) (*NetworkPool, error) {
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
		vcdClient:   &vcdClient,
		NetworkPool: &types.NetworkPool{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, response.NetworkPool, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetNetworkPoolByName retrieves IP Space with a given name
// Note. It will return an error if multiple network pools exist with the same name
func (vcdClient VCDClient) GetNetworkPoolByName(name string) (*NetworkPool, error) {
	if name == "" {
		return nil, fmt.Errorf("network pool lookup requires name")
	}

	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	filteredNetworkPools, err := vcdClient.GetNetworkPoolSummaries(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error getting network pools: %s", err)
	}

	if len(filteredNetworkPools.Values) > 1 {
		return nil, fmt.Errorf("more than one network pool found with name '%s'", name)
	}

	return vcdClient.GetNetworkPoolById(filteredNetworkPools.Values[0].Id)
}
