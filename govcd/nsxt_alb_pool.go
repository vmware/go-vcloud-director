/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type NsxtAlbPool struct {
	NsxtAlbPool *types.NsxtAlbPool
	vcdClient   *VCDClient
}

func (vcdClient *VCDClient) GetAllAlbPools(queryParameters url.Values) ([]*NsxtAlbPool, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPools
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtAlbPool{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtAlbPool types with client
	wrappedResponses := make([]*NsxtAlbPool, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbPool{
			NsxtAlbPool: typeResponses[sliceIndex],
			vcdClient:   vcdClient,
		}
	}

	return wrappedResponses, nil
}

func (vcdClient *VCDClient) GetAlbPoolByName(name string) (*NsxtAlbPool, error) {
	queryParameters := copyOrNewUrlValues(nil)
	queryParameters.Add("filter", "name=="+name)

	allAlbPools, err := vcdClient.GetAllAlbPools(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error reading ALB Pool with Name '%s': %s", name, err)
	}

	if len(allAlbPools) == 0 {
		return nil, fmt.Errorf("%s: could not find ALB Pool with Name '%s'", ErrorEntityNotFound, name)
	}

	if len(allAlbPools) > 1 {
		return nil, fmt.Errorf("found more than 1 ALB Pool with Name '%s'", name)
	}

	return allAlbPools[0], nil
}

func (vcdClient *VCDClient) GetAlbPoolById(id string) (*NsxtAlbPool, error) {
	client := vcdClient.Client

	if id == "" {
		return nil, fmt.Errorf("ID is required to lookup NSX-T ALB Pool by ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPools
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	typeResponse := &types.NsxtAlbPool{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponse := &NsxtAlbPool{
		NsxtAlbPool: typeResponse,
		vcdClient:   vcdClient,
	}

	return wrappedResponse, nil
}

func (vcdClient *VCDClient) CreateNsxtAlbPool(albPoolConfig *types.NsxtAlbPool) (*NsxtAlbPool, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPools
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtAlbPool{
		NsxtAlbPool: &types.NsxtAlbPool{},
		vcdClient:   vcdClient,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, albPoolConfig, returnObject.NsxtAlbPool, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T ALB Pool: %s", err)
	}

	return returnObject, nil
}

func (nsxtAlbPool *NsxtAlbPool) Update(albPoolConfig *types.NsxtAlbPool) (*NsxtAlbPool, error) {
	client := nsxtAlbPool.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPools
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if albPoolConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T ALB Pool without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, albPoolConfig.ID)
	if err != nil {
		return nil, err
	}

	responseAlbController := &NsxtAlbPool{
		NsxtAlbPool: &types.NsxtAlbPool{},
		vcdClient:   nsxtAlbPool.vcdClient,
	}

	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, albPoolConfig, responseAlbController.NsxtAlbPool, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T ALB Pool: %s", err)
	}

	return responseAlbController, nil
}

func (nsxtAlbPool *NsxtAlbPool) Delete() error {
	client := nsxtAlbPool.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPools
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if nsxtAlbPool.NsxtAlbPool.ID == "" {
		return fmt.Errorf("cannot delete NSX-T ALB Pool without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, nsxtAlbPool.NsxtAlbPool.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T ALB Pool: %s", err)
	}

	return nil
}
