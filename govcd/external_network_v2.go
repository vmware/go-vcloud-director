// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelNsxtTier0RouterInterface = "NSX-T Tier-0 Router Interface"

// ExternalNetworkV2 is a type for version 2 of external network which uses OpenAPI endpoint to
// manage external networks of both types (NSX-V and NSX-T)
type ExternalNetworkV2 struct {
	ExternalNetwork *types.ExternalNetworkV2
	client          *Client
}

// CreateExternalNetworkV2 creates a new external network using OpenAPI endpoint. It can create
// NSX-V and NSX-T backed networks based on what ExternalNetworkV2.NetworkBackings is
// provided. types.ExternalNetworkV2 has documented fields.
func CreateExternalNetworkV2(vcdClient *VCDClient, newExtNet *types.ExternalNetworkV2) (*ExternalNetworkV2, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnExtNet := &ExternalNetworkV2{
		ExternalNetwork: &types.ExternalNetworkV2{},
		client:          &vcdClient.Client,
	}

	err = vcdClient.Client.OpenApiPostItem(apiVersion, urlRef, nil, newExtNet, returnExtNet.ExternalNetwork, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating external network: %s", err)
	}

	return returnExtNet, nil
}

// GetExternalNetworkV2ById retrieves external network by given ID using OpenAPI endpoint
func GetExternalNetworkV2ById(vcdClient *VCDClient, id string) (*ExternalNetworkV2, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty external network id")
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	extNet := &ExternalNetworkV2{
		ExternalNetwork: &types.ExternalNetworkV2{},
		client:          &vcdClient.Client,
	}

	err = vcdClient.Client.OpenApiGetItem(apiVersion, urlRef, nil, extNet.ExternalNetwork, nil)
	if err != nil {
		return nil, err
	}

	return extNet, nil
}

// GetExternalNetworkV2ByName retrieves external network by given name using OpenAPI endpoint.
// Returns an error if not exactly one network is found.
func GetExternalNetworkV2ByName(vcdClient *VCDClient, name string) (*ExternalNetworkV2, error) {

	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	res, err := GetAllExternalNetworksV2(vcdClient, queryParams)
	if err != nil {
		return nil, fmt.Errorf("could not find external network by name: %s", err)
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("%s: expected exactly one external network with name '%s'. Got %d", ErrorEntityNotFound, name, len(res))
	}

	if len(res) > 1 {
		return nil, fmt.Errorf("expected exactly one external network with name '%s'. Got %d", name, len(res))
	}

	return res[0], nil
}

// GetAllExternalNetworksV2 retrieves all external networks using OpenAPI endpoint. Query parameters can be supplied to
// perform additional filtering
func GetAllExternalNetworksV2(vcdClient *VCDClient, queryParameters url.Values) ([]*ExternalNetworkV2, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.ExternalNetworkV2{{}}
	err = vcdClient.Client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into external network types with client
	returnExtNetworks := make([]*ExternalNetworkV2, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnExtNetworks[sliceIndex] = &ExternalNetworkV2{
			ExternalNetwork: typeResponses[sliceIndex],
			client:          &vcdClient.Client,
		}
	}

	return returnExtNetworks, nil
}

// Update updates existing external network using OpenAPI endpoint
func (extNet *ExternalNetworkV2) Update() (*ExternalNetworkV2, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks
	apiVersion, err := extNet.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if extNet.ExternalNetwork.ID == "" {
		return nil, fmt.Errorf("cannot update external network without id")
	}

	urlRef, err := extNet.client.OpenApiBuildEndpoint(endpoint, extNet.ExternalNetwork.ID)
	if err != nil {
		return nil, err
	}

	returnExtNet := &ExternalNetworkV2{
		ExternalNetwork: &types.ExternalNetworkV2{},
		client:          extNet.client,
	}

	err = extNet.client.OpenApiPutItem(apiVersion, urlRef, nil, extNet.ExternalNetwork, returnExtNet.ExternalNetwork, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating external network: %s", err)
	}

	return returnExtNet, nil
}

// Delete deletes external network using OpenAPI endpoint
func (extNet *ExternalNetworkV2) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks
	apiVersion, err := extNet.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if extNet.ExternalNetwork.ID == "" {
		return fmt.Errorf("cannot delete external network without id")
	}

	urlRef, err := extNet.client.OpenApiBuildEndpoint(endpoint, extNet.ExternalNetwork.ID)
	if err != nil {
		return err
	}

	err = extNet.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting extNet: %s", err)
	}

	return nil
}

// GetAllTier0RouterInterfaces returns all Provider Gateway (aka Tier0 Router) associated interfaces
func (vcdClient *VCDClient) GetAllTier0RouterInterfaces(externalNetworkId string, queryParameters url.Values) ([]*types.NsxtTier0RouterInterface, error) {
	if externalNetworkId == "" {
		return nil, fmt.Errorf("mandatory External Network ID is empty")
	}

	queryparams := queryParameterFilterAnd(fmt.Sprintf("externalNetworkId==%s", externalNetworkId), queryParameters)
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtTier0RouterInterfaces,
		entityLabel:     labelNsxtTier0RouterInterface,
		queryParameters: queryparams,
	}

	return getAllInnerEntities[types.NsxtTier0RouterInterface](&vcdClient.Client, c)
}

// GetTier0RouterInterfaceByName retrieves a Provider Gateway (aka Tier0 Router) associated interface by Name in a given External Network
func (vcdClient *VCDClient) GetTier0RouterInterfaceByName(externalNetworkId, displayName string) (*types.NsxtTier0RouterInterface, error) {
	allT0Interfaces, err := vcdClient.GetAllTier0RouterInterfaces(externalNetworkId, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting %s by DisplayName '%s':%s", labelNsxtTier0RouterInterface, displayName, err)
	}
	if len(allT0Interfaces) == 0 {
		return nil, fmt.Errorf("%s: error getting %s by DisplayName '%s'", ErrorEntityNotFound, labelNsxtTier0RouterInterface, displayName)
	}

	return localFilterOneOrError(labelNsxtTier0RouterInterface, allT0Interfaces, "DisplayName", displayName)
}
