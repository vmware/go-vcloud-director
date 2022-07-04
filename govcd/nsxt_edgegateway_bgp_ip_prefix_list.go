/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type EdgeBgpConfigPrefixList struct {
	EdgeBgpConfigPrefixList *types.EdgeBgpConfigPrefixList
	client                  *Client
	// edgeGatewayId is stored for usage in EdgeBgpConfigPrefixList receiver functions
	edgeGatewayId string
}

func (egw *NsxtEdgeGateway) GetAllBgpIpPrefixLists(queryParameters url.Values) ([]*EdgeBgpConfigPrefixList, error) {
	queryParams := copyOrNewUrlValues(queryParameters)

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfigPrefixLists
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.EdgeBgpConfigPrefixList{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtNatRule types with client
	wrappedResponses := make([]*EdgeBgpConfigPrefixList, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &EdgeBgpConfigPrefixList{
			EdgeBgpConfigPrefixList: typeResponses[sliceIndex],
			client:                  client,
			edgeGatewayId:           egw.EdgeGateway.ID,
		}
	}

	return wrappedResponses, nil
}

// GetBgpIpPrefixListByName retrieves BGP IP Prefix List By Name
//
// Note. API does not support filtering by 'name' field therefore filtering is performed on client
// side
func (egw *NsxtEdgeGateway) GetBgpIpPrefixListByName(name string) (*EdgeBgpConfigPrefixList, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	allBgpIpPrefixLists, err := egw.GetAllBgpIpPrefixLists(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	var filteredBgpIpPrefixLists []*EdgeBgpConfigPrefixList
	for _, bgpIpPrefixList := range allBgpIpPrefixLists {
		if bgpIpPrefixList.EdgeBgpConfigPrefixList.Name == name {
			filteredBgpIpPrefixLists = append(filteredBgpIpPrefixLists, bgpIpPrefixList)
		}
	}

	if len(filteredBgpIpPrefixLists) > 1 {
		return nil, fmt.Errorf("more than one NSX-T Edge Gateway BGP IP Prefix List found with Name '%s'", name)
	}

	if len(filteredBgpIpPrefixLists) == 0 {
		return nil, fmt.Errorf("no NSX-T Edge Gateway BGP IP Prefix List found with name '%s'", name)
	}

	return filteredBgpIpPrefixLists[0], nil
}

func (egw *NsxtEdgeGateway) GetBgpIpPrefixListById(id string) (*EdgeBgpConfigPrefixList, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfigPrefixLists
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), id)
	if err != nil {
		return nil, err
	}

	returnObject := &EdgeBgpConfigPrefixList{
		client:                  egw.client,
		edgeGatewayId:           egw.EdgeGateway.ID,
		EdgeBgpConfigPrefixList: &types.EdgeBgpConfigPrefixList{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject.EdgeBgpConfigPrefixList, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	return returnObject, nil
}

func (egw *NsxtEdgeGateway) CreateBgpIpPrefixList(bgpConfig *types.EdgeBgpConfigPrefixList) (*EdgeBgpConfigPrefixList, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfigPrefixLists
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	returnObject := &EdgeBgpConfigPrefixList{
		client:                  egw.client,
		edgeGatewayId:           egw.EdgeGateway.ID,
		EdgeBgpConfigPrefixList: &types.EdgeBgpConfigPrefixList{},
	}

	// API returns newly created object ID in Details field therefore regular "automatic"
	// OpenApiPostItem function cannot handle it and a more verbose way is used by tracking task
	// and looking up inside Details field for object ID lookup
	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, bgpConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	bgpIpPrefixListId := task.Task.Details
	if bgpIpPrefixListId == "" {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway BGP IP Prefix List: empty ID returned")
	}

	getUrlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), bgpIpPrefixListId)
	if err != nil {
		return nil, err
	}
	err = client.OpenApiGetItem(apiVersion, getUrlRef, nil, returnObject.EdgeBgpConfigPrefixList, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP IP Prefix List after creation: %s", err)
	}

	return returnObject, nil
}

//
// Note. Update of BGP configuration requires version to be specified in 'Version' field. This
// function automatically handles it.
func (egwBgpConfig *EdgeBgpConfigPrefixList) UpdateBgpIpPrefixList(bgpConfig *types.EdgeBgpConfigPrefixList) (*EdgeBgpConfigPrefixList, error) {
	client := egwBgpConfig.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfigPrefixLists
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egwBgpConfig.edgeGatewayId), bgpConfig.ID)
	if err != nil {
		return nil, err
	}

	returnObject := &EdgeBgpConfigPrefixList{
		client:                  egwBgpConfig.client,
		edgeGatewayId:           egwBgpConfig.edgeGatewayId,
		EdgeBgpConfigPrefixList: &types.EdgeBgpConfigPrefixList{},
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, bgpConfig, returnObject.EdgeBgpConfigPrefixList, nil)
	if err != nil {
		return nil, fmt.Errorf("error setting NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	return returnObject, nil
}

func (egwBgpConfig *EdgeBgpConfigPrefixList) Delete() error {
	client := egwBgpConfig.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfigPrefixLists
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egwBgpConfig.edgeGatewayId), egwBgpConfig.EdgeBgpConfigPrefixList.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	return nil
}
