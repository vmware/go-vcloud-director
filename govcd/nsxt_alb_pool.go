/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtAlbPool defines configuration of a single NSX-T ALB Pool. Pools maintain the list of servers assigned to them and
// perform health monitoring, load balancing, persistence. A pool may only be used or referenced by only one virtual
// service at a time.
type NsxtAlbPool struct {
	NsxtAlbPool *types.NsxtAlbPool
	vcdClient   *VCDClient
}

// GetAllAlbPoolSummaries retrieves partial information for type `NsxtAlbPool`, but it is the only way to retrieve all ALB
// pools for Edge Gateway
func (vcdClient *VCDClient) GetAllAlbPoolSummaries(edgeGatewayId string, queryParameters url.Values) ([]*NsxtAlbPool, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPoolSummaries
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, edgeGatewayId))
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

// GetAllAlbPools uses GetAllAlbPoolSummaries behind the scenes and the fetches complete data for all ALB Pools. This
// has performance penalty because each ALB Pool is fetched individually.
func (vcdClient *VCDClient) GetAllAlbPools(edgeGatewayId string, queryParameters url.Values) ([]*NsxtAlbPool, error) {
	allAlbPoolSummaries, err := vcdClient.GetAllAlbPoolSummaries(edgeGatewayId, queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all ALB Pool summaries: %s", err)
	}

	// Loop over all Summaries and retrieve complete information
	allAlbPools := make([]*NsxtAlbPool, len(allAlbPoolSummaries))
	for index := range allAlbPoolSummaries {

		allAlbPools[index], err = vcdClient.GetAlbPoolById(allAlbPoolSummaries[index].NsxtAlbPool.ID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving complete ALB Pool: %s", err)
		}

	}

	return allAlbPools, nil
}

// GetAlbPoolByName fetches ALB Pool By Name
func (vcdClient *VCDClient) GetAlbPoolByName(edgeGatewayId string, name string) (*NsxtAlbPool, error) {
	queryParameters := copyOrNewUrlValues(nil)
	queryParameters.Add("filter", "name=="+name)

	allAlbPools, err := vcdClient.GetAllAlbPools(edgeGatewayId, queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving ALB Pool with Name '%s': %s", name, err)
	}

	if len(allAlbPools) == 0 {
		return nil, fmt.Errorf("%s: could not find ALB Pool with Name '%s'", ErrorEntityNotFound, name)
	}

	if len(allAlbPools) > 1 {
		return nil, fmt.Errorf("found more than 1 ALB Pool with Name '%s'", name)
	}

	return allAlbPools[0], nil
}

// GetAlbPoolById fetches ALB Pool By Id
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

// CreateNsxtAlbPool creates NSX-T ALB Pool based on supplied configuration
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

// Update updates NSX-T ALB Pool based on supplied configuration
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

// Delete deletes NSX-T ALB Pool
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
