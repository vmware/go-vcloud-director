/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// IpSpaceUplink provides the capability to assign one or more IP Spaces as Uplinks to External
// Networks
type IpSpaceUplink struct {
	IpSpaceUplink *types.IpSpaceUplink
	vcdClient     *VCDClient
}

// CreateIpSpaceUplink creates an IP Space Uplink with a given configuration
func (vcdClient *VCDClient) CreateIpSpaceUplink(ipSpaceUplinkConfig *types.IpSpaceUplink) (*IpSpaceUplink, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &IpSpaceUplink{
		IpSpaceUplink: &types.IpSpaceUplink{},
		vcdClient:     vcdClient,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, ipSpaceUplinkConfig, result.IpSpaceUplink, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetAllIpSpaceUplinks retrieves all IP Space Uplinks for a given External Network ID
//
// externalNetworkId is mandatory
func (vcdClient *VCDClient) GetAllIpSpaceUplinks(externalNetworkId string, queryParameters url.Values) ([]*IpSpaceUplink, error) {
	if externalNetworkId == "" {
		return nil, fmt.Errorf("mandatory External Network ID is empty")
	}

	queryparams := queryParameterFilterAnd(fmt.Sprintf("externalNetworkRef.id==%s", externalNetworkId), queryParameters)

	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.IpSpaceUplink{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryparams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into IpSpaceUplink types with client
	results := make([]*IpSpaceUplink, len(typeResponses))
	for sliceIndex := range typeResponses {
		results[sliceIndex] = &IpSpaceUplink{
			IpSpaceUplink: typeResponses[sliceIndex],
			vcdClient:     vcdClient,
		}
	}

	return results, nil
}

// GetIpSpaceUplinkByName retrieves a single IP Space Uplink by Name in a given External Network
func (vcdClient *VCDClient) GetIpSpaceUplinkByName(externalNetworkId, name string) (*IpSpaceUplink, error) {
	queryParams := queryParameterFilterAnd(fmt.Sprintf("name==%s", name), nil)
	allIpSpaceUplinks, err := vcdClient.GetAllIpSpaceUplinks(externalNetworkId, queryParams)
	if err != nil {
		return nil, fmt.Errorf("error getting IP Space Uplink by Name '%s':%s", name, err)
	}

	return oneOrError("name", name, allIpSpaceUplinks)
}

// GetIpSpaceUplinkById retrieves IP Space Uplink with a given ID
func (vcdClient *VCDClient) GetIpSpaceUplinkById(id string) (*IpSpaceUplink, error) {
	if id == "" {
		return nil, fmt.Errorf("IP Space Uplink lookup requires ID")
	}

	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	response := &IpSpaceUplink{
		vcdClient:     vcdClient,
		IpSpaceUplink: &types.IpSpaceUplink{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, response.IpSpaceUplink, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Update IP Space Uplink
func (ipSpaceUplink *IpSpaceUplink) Update(ipSpaceUplinkConfig *types.IpSpaceUplink) (*IpSpaceUplink, error) {
	client := ipSpaceUplink.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	ipSpaceUplinkConfig.ID = ipSpaceUplink.IpSpaceUplink.ID
	urlRef, err := client.OpenApiBuildEndpoint(endpoint, ipSpaceUplinkConfig.ID)
	if err != nil {
		return nil, err
	}

	result := &IpSpaceUplink{
		IpSpaceUplink: &types.IpSpaceUplink{},
		vcdClient:     ipSpaceUplink.vcdClient,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, ipSpaceUplinkConfig, result.IpSpaceUplink, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating IP Space: %s", err)
	}

	return result, nil
}

// Delete IP Space Uplink
func (ipSpaceUplink *IpSpaceUplink) Delete() error {
	if ipSpaceUplink == nil || ipSpaceUplink.IpSpaceUplink == nil || ipSpaceUplink.IpSpaceUplink.ID == "" {
		return fmt.Errorf("IP Space Uplink must have ID")
	}

	client := ipSpaceUplink.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, ipSpaceUplink.IpSpaceUplink.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("error deleting IP Space Uplink: %s", err)
	}

	return nil
}
