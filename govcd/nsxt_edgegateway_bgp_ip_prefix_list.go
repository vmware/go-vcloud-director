/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// EdgeBgpIpPrefixList helps to configure BGP IP Prefix Lists in NSX-T Edge Gateways
type EdgeBgpIpPrefixList struct {
	EdgeBgpIpPrefixList *types.EdgeBgpIpPrefixList
	client              *Client
	// edgeGatewayId is stored for usage in EdgeBgpIpPrefixList receiver functions
	edgeGatewayId string
}

// CreateBgpIpPrefixList creates a BGP IP Prefix List with supplied configuration
//
// Note. VCD 10.2 versions do not automatically return ID for created BGP IP Prefix List. To work around it this code
// automatically retrieves the entity by Name after the task is finished. This function may fail on VCD 10.2 if
// duplicate BGP IP Prefix Lists exist.
func (egw *NsxtEdgeGateway) CreateBgpIpPrefixList(bgpIpPrefixList *types.EdgeBgpIpPrefixList) (*EdgeBgpIpPrefixList, error) {
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

	returnObject := &EdgeBgpIpPrefixList{
		client:              egw.client,
		edgeGatewayId:       egw.EdgeGateway.ID,
		EdgeBgpIpPrefixList: &types.EdgeBgpIpPrefixList{},
	}

	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, bgpIpPrefixList)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	// API is not consistent across different versions therefore explicit manual handling is
	// required to lookup newly created object
	//
	// VCD 10.2 -> no ID for newly created object is returned at all
	// VCD 10.3 -> `Details` field in task contains ID of newly created object
	// To cover all cases this code will at first look for ID in `Details` field and fall back to
	// lookup by name if `Details` field is empty.
	//
	// The drawback of this is that it is possible to create duplicate records with the same name on
	// VCD versions that don't return IDs, but there is no better way for VCD versions that don't
	// return IDs for created objects

	bgpIpPrefixListId := task.Task.Details
	if bgpIpPrefixListId != "" {
		getUrlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), bgpIpPrefixListId)
		if err != nil {
			return nil, err
		}
		err = client.OpenApiGetItem(apiVersion, getUrlRef, nil, returnObject.EdgeBgpIpPrefixList, nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP IP Prefix List after creation: %s", err)
		}
	} else {
		// ID after object creation was not returned therefore retrieving the entity by Name to lookup ID
		// This has a risk of duplicate items, but is the only way to find the object when ID is not returned
		bgpIpPrefixList, err := egw.GetBgpIpPrefixListByName(bgpIpPrefixList.Name)
		if err != nil {
			return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP IP Prefix List after creation: %s", err)
		}
		returnObject = bgpIpPrefixList
	}

	return returnObject, nil
}

// GetAllBgpIpPrefixLists retrieves all BGP IP Prefix Lists in a given NSX-T Edge Gateway with optional queryParameters
func (egw *NsxtEdgeGateway) GetAllBgpIpPrefixLists(queryParameters url.Values) ([]*EdgeBgpIpPrefixList, error) {
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

	typeResponses := []*types.EdgeBgpIpPrefixList{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtNatRule types with client
	wrappedResponses := make([]*EdgeBgpIpPrefixList, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &EdgeBgpIpPrefixList{
			EdgeBgpIpPrefixList: typeResponses[sliceIndex],
			client:              client,
			edgeGatewayId:       egw.EdgeGateway.ID,
		}
	}

	return wrappedResponses, nil
}

// GetBgpIpPrefixListByName retrieves BGP IP Prefix List By Name
// It is meant to retrieve exactly one entry:
// * Will fail if more than one entry with the same name found
// * Will return an error containing `ErrorEntityNotFound` if no entries are found
//
// Note. API does not support filtering by 'name' field therefore filtering is performed on client
// side
func (egw *NsxtEdgeGateway) GetBgpIpPrefixListByName(name string) (*EdgeBgpIpPrefixList, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	allBgpIpPrefixLists, err := egw.GetAllBgpIpPrefixLists(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	var filteredBgpIpPrefixLists []*EdgeBgpIpPrefixList
	for _, bgpIpPrefixList := range allBgpIpPrefixLists {
		if bgpIpPrefixList.EdgeBgpIpPrefixList.Name == name {
			filteredBgpIpPrefixLists = append(filteredBgpIpPrefixLists, bgpIpPrefixList)
		}
	}

	if len(filteredBgpIpPrefixLists) > 1 {
		return nil, fmt.Errorf("more than one NSX-T Edge Gateway BGP IP Prefix List found with Name '%s'", name)
	}

	if len(filteredBgpIpPrefixLists) == 0 {
		return nil, fmt.Errorf("%s: no NSX-T Edge Gateway BGP IP Prefix List found with name '%s'", ErrorEntityNotFound, name)
	}

	return filteredBgpIpPrefixLists[0], nil
}

// GetBgpIpPrefixListById retrieves BGP IP Prefix List By ID
func (egw *NsxtEdgeGateway) GetBgpIpPrefixListById(id string) (*EdgeBgpIpPrefixList, error) {
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

	returnObject := &EdgeBgpIpPrefixList{
		client:              egw.client,
		edgeGatewayId:       egw.EdgeGateway.ID,
		EdgeBgpIpPrefixList: &types.EdgeBgpIpPrefixList{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject.EdgeBgpIpPrefixList, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	return returnObject, nil
}

// Update updates existing BGP IP Prefix List with new configuration and returns it
func (bgpIpPrefixListCfg *EdgeBgpIpPrefixList) Update(bgpIpPrefixList *types.EdgeBgpIpPrefixList) (*EdgeBgpIpPrefixList, error) {
	client := bgpIpPrefixListCfg.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfigPrefixLists
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, bgpIpPrefixListCfg.edgeGatewayId), bgpIpPrefixList.ID)
	if err != nil {
		return nil, err
	}

	returnObject := &EdgeBgpIpPrefixList{
		client:              bgpIpPrefixListCfg.client,
		edgeGatewayId:       bgpIpPrefixListCfg.edgeGatewayId,
		EdgeBgpIpPrefixList: &types.EdgeBgpIpPrefixList{},
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, bgpIpPrefixList, returnObject.EdgeBgpIpPrefixList, nil)
	if err != nil {
		return nil, fmt.Errorf("error setting NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	return returnObject, nil
}

// Delete deletes existing BGP IP Prefix List
func (bgpIpPrefixListCfg *EdgeBgpIpPrefixList) Delete() error {
	client := bgpIpPrefixListCfg.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfigPrefixLists
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, bgpIpPrefixListCfg.edgeGatewayId), bgpIpPrefixListCfg.EdgeBgpIpPrefixList.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	return nil
}
