/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtAlbVirtualService combines Load Balancer Pools with Service Engine Groups and exposes a virtual service on
// defined VIP (virtual IP address) while optionally allowing to use encrypted traffic
type NsxtAlbVirtualService struct {
	NsxtAlbVirtualService *types.NsxtAlbVirtualService
	vcdClient             *VCDClient
}

// GetAllAlbVirtualServiceSummaries returns a limited subset of NsxtAlbVirtualService values, but does it in single
// query. To fetch complete information for ALB Virtual Services one can use GetAllAlbVirtualServices(), but it is slower
// as it has to retrieve Virtual Services one by one.
func (vcdClient *VCDClient) GetAllAlbVirtualServiceSummaries(edgeGatewayId string, queryParameters url.Values) ([]*NsxtAlbVirtualService, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceSummaries
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, edgeGatewayId))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtAlbVirtualService{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtAlbPool types with client
	wrappedResponses := make([]*NsxtAlbVirtualService, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbVirtualService{
			NsxtAlbVirtualService: typeResponses[sliceIndex],
			vcdClient:             vcdClient,
		}
	}

	return wrappedResponses, nil
}

// GetAllAlbVirtualServices fetches ALB Virtual Services by at first listing all Virtual Services summaries and then
// fetching complete structure one by one
func (vcdClient *VCDClient) GetAllAlbVirtualServices(edgeGatewayId string, queryParameters url.Values) ([]*NsxtAlbVirtualService, error) {
	allAlbVirtualServiceSummaries, err := vcdClient.GetAllAlbVirtualServiceSummaries(edgeGatewayId, queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all ALB Virtual Service summaries: %s", err)
	}

	// Loop over all Summaries and retrieve complete information
	allAlbVirtualServices := make([]*NsxtAlbVirtualService, len(allAlbVirtualServiceSummaries))
	for index := range allAlbVirtualServiceSummaries {
		allAlbVirtualServices[index], err = vcdClient.GetAlbVirtualServiceById(allAlbVirtualServiceSummaries[index].NsxtAlbVirtualService.ID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving complete ALB Virtual Service: %s", err)
		}

	}

	return allAlbVirtualServices, nil
}

// GetAlbVirtualServiceByName fetches ALB Virtual Service By Name
func (vcdClient *VCDClient) GetAlbVirtualServiceByName(edgeGatewayId string, name string) (*NsxtAlbVirtualService, error) {
	queryParameters := copyOrNewUrlValues(nil)
	queryParameters.Add("filter", "name=="+name)

	allAlbVirtualServices, err := vcdClient.GetAllAlbVirtualServices(edgeGatewayId, queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error reading ALB Virtual Service with Name '%s': %s", name, err)
	}

	if len(allAlbVirtualServices) == 0 {
		return nil, fmt.Errorf("%s: could not find ALB Virtual Service with Name '%s'", ErrorEntityNotFound, name)
	}

	if len(allAlbVirtualServices) > 1 {
		return nil, fmt.Errorf("found more than 1 ALB Virtual Service with Name '%s'", name)
	}

	return allAlbVirtualServices[0], nil
}

// GetAlbVirtualServiceById fetches ALB Virtual Service By ID
func (vcdClient *VCDClient) GetAlbVirtualServiceById(id string) (*NsxtAlbVirtualService, error) {
	client := vcdClient.Client

	if id == "" {
		return nil, fmt.Errorf("ID is required to lookup NSX-T ALB Virtual Service by ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServices
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	typeResponse := &types.NsxtAlbVirtualService{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponse := &NsxtAlbVirtualService{
		NsxtAlbVirtualService: typeResponse,
		vcdClient:             vcdClient,
	}

	return wrappedResponse, nil
}

// CreateNsxtAlbVirtualService creates NSX-T ALB Virtual Service based on supplied configuration
func (vcdClient *VCDClient) CreateNsxtAlbVirtualService(albVirtualServiceConfig *types.NsxtAlbVirtualService) (*NsxtAlbVirtualService, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServices
	minimumApiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtAlbVirtualService{
		NsxtAlbVirtualService: &types.NsxtAlbVirtualService{},
		vcdClient:             vcdClient,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, albVirtualServiceConfig, returnObject.NsxtAlbVirtualService, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T ALB Virtual Service: %s", err)
	}

	return returnObject, nil
}

// Update updates NSX-T ALB Virtual Service based on supplied configuration
func (nsxtAlbVirtualService *NsxtAlbVirtualService) Update(albVirtualServiceConfig *types.NsxtAlbVirtualService) (*NsxtAlbVirtualService, error) {
	client := nsxtAlbVirtualService.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServices
	minimumApiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if albVirtualServiceConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T ALB Virtual Service without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, albVirtualServiceConfig.ID)
	if err != nil {
		return nil, err
	}

	responseAlbController := &NsxtAlbVirtualService{
		NsxtAlbVirtualService: &types.NsxtAlbVirtualService{},
		vcdClient:             nsxtAlbVirtualService.vcdClient,
	}

	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, albVirtualServiceConfig, responseAlbController.NsxtAlbVirtualService, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T ALB Virtual Service: %s", err)
	}

	return responseAlbController, nil
}

// Delete deletes NSX-T ALB Virtual Service
func (nsxtAlbVirtualService *NsxtAlbVirtualService) Delete() error {
	client := nsxtAlbVirtualService.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServices
	minimumApiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if nsxtAlbVirtualService.NsxtAlbVirtualService.ID == "" {
		return fmt.Errorf("cannot delete NSX-T ALB Virtual Service without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, nsxtAlbVirtualService.NsxtAlbVirtualService.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T ALB Virtual Service: %s", err)
	}

	return nil
}
