/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type IpSpace struct {
	IpSpace   *types.IpSpace
	vcdClient *VCDClient
}

// CreateIpSpace creates IP Space with desired configuration
func (vcdClient *VCDClient) CreateIpSpace(ipSpaceConfig *types.IpSpace) (*IpSpace, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &IpSpace{
		IpSpace:   &types.IpSpace{},
		vcdClient: vcdClient,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, ipSpaceConfig, result.IpSpace, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetAllIpSpaceSummaries retrieve summaries of all IP Spaces with an optional filter
// Note. There is no API endpoint to get multiple IP Spaces with their full definitions. Only
// "summaries" endpoint exists, but it does not include all fields. To retrieve complete structure
// one can use `GetIpSpaceById` or `GetIpSpaceByName`
func (vcdClient *VCDClient) GetAllIpSpaceSummaries(queryParameters url.Values) ([]*IpSpace, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceSummaries
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.IpSpace{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into DefinedEntityType types with client
	results := make([]*IpSpace, len(typeResponses))
	for sliceIndex := range typeResponses {
		results[sliceIndex] = &IpSpace{
			IpSpace:   typeResponses[sliceIndex],
			vcdClient: vcdClient,
		}
	}

	return results, nil
}

// GetIpSpaceById retrieves IP Space with a given ID
func (vcdClient *VCDClient) GetIpSpaceById(id string) (*IpSpace, error) {
	if id == "" {
		return nil, fmt.Errorf("IP Space lookup requires ID")
	}

	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	response := &IpSpace{
		vcdClient: vcdClient,
		IpSpace:   &types.IpSpace{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, response.IpSpace, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetIpSpaceByName retrieves IP Space with a given name
// Note. It will return an error if multiple IP Spaces exist with the same name
func (vcdClient *VCDClient) GetIpSpaceByName(name string) (*IpSpace, error) {
	if name == "" {
		return nil, fmt.Errorf("IP Space lookup requires name")
	}

	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	filteredIpSpaces, err := vcdClient.GetAllIpSpaceSummaries(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error getting IP spaces: %s", err)
	}

	singleIpSpace, err := oneOrError("name", name, filteredIpSpaces)
	if err != nil {
		return nil, fmt.Errorf("error ")
	}

	return vcdClient.GetIpSpaceById(singleIpSpace.IpSpace.ID)

}

// Update updates IP Space with new config
func (ipSpace *IpSpace) Update(ipSpaceConfig *types.IpSpace) (*IpSpace, error) {
	client := ipSpace.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	ipSpaceConfig.ID = ipSpace.IpSpace.ID
	urlRef, err := client.OpenApiBuildEndpoint(endpoint, ipSpaceConfig.ID)
	if err != nil {
		return nil, err
	}

	returnIpSpace := &IpSpace{
		IpSpace:   &types.IpSpace{},
		vcdClient: ipSpace.vcdClient,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, ipSpaceConfig, returnIpSpace.IpSpace, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating IP Space: %s", err)
	}

	return returnIpSpace, nil

}

// Delete deletes IP Space
func (ipSpace *IpSpace) Delete() error {
	if ipSpace == nil || ipSpace.IpSpace == nil || ipSpace.IpSpace.ID == "" {
		return fmt.Errorf("IP Space must have ID")
	}

	client := ipSpace.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, ipSpace.IpSpace.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("error deleting IP space: %s", err)
	}

	return nil
}
