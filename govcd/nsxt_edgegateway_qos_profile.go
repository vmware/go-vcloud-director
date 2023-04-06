/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtEdgeGatewayQosProfiles uses OpenAPI endpoint to fetch NSX-T Edge Gateway QoS Profiles defined
// in NSX-T Manager. They can be used to configure QoS on NSX-T Edge Gateway.
type NsxtEdgeGatewayQosProfile struct {
	NsxtEdgeGatewayQosProfile *types.NsxtEdgeGatewayQosProfile
	client                    *Client
}

// GetAllNsxtEdgeGatewayQosProfiles retrieves all NSX-T Edge Gateway QoS Profiles defined in NSX-T Manager
func (vcdClient *VCDClient) GetAllNsxtEdgeGatewayQosProfiles(nsxtManagerId string, queryParameters url.Values) ([]*NsxtEdgeGatewayQosProfile, error) {
	if nsxtManagerId == "" {
		return nil, fmt.Errorf("empty NSX-T manager ID")
	}

	if !isUrn(nsxtManagerId) {
		return nil, fmt.Errorf("NSX-T manager ID is not URN (e.g. 'urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc)', got: %s", nsxtManagerId)
	}

	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointQosProfiles
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("nsxTManagerRef.id=="+nsxtManagerId, queryParams)

	typeResponses := []*types.NsxtEdgeGatewayQosProfile{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	returnObjects := make([]*NsxtEdgeGatewayQosProfile, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnObjects[sliceIndex] = &NsxtEdgeGatewayQosProfile{
			NsxtEdgeGatewayQosProfile: typeResponses[sliceIndex],
			client:                    &client,
		}
	}

	return returnObjects, nil
}

// GetNsxtEdgeGatewayQosProfileById retrieves NSX-T Edge Gateway QoS Profile by Display Name
func (vcdClient *VCDClient) GetNsxtEdgeGatewayQosProfileByDisplayName(nsxtManagerId, displayName string) (*NsxtEdgeGatewayQosProfile, error) {
	if displayName == "" {
		return nil, fmt.Errorf("empty QoS profile Display Name")
	}

	// Ideally FIQL filter could be used to filter on server side and get only desired result, but filtering on
	// 'displayName' is not yet supported.
	/*
		queryParameters := copyOrNewUrlValues(nil)
		queryParameters.Add("filter", "displayName=="+displayName)
	*/
	nsxtEdgeClusters, err := vcdClient.GetAllNsxtEdgeGatewayQosProfiles(nsxtManagerId, nil)
	if err != nil {
		return nil, fmt.Errorf("could not find QoS profile with DisplayName '%s' for NSX-T Manager with ID '%s': %s",
			displayName, nsxtManagerId, err)
	}

	// TODO remove this when FIQL supports filtering on 'displayName'
	nsxtEdgeClusters = filterQosProfiles(displayName, nsxtEdgeClusters)
	// EOF TODO remove this when FIQL supports filtering on 'displayName'

	if len(nsxtEdgeClusters) == 0 {
		// ErrorEntityNotFound is injected here for the ability to validate problem using ContainsNotFound()
		return nil, fmt.Errorf("%s: no NSX-T QoS profiles with DisplayName '%s' for NSX-T Manager with ID '%s' found",
			ErrorEntityNotFound, displayName, nsxtManagerId)
	}

	if len(nsxtEdgeClusters) > 1 {
		return nil, fmt.Errorf("more than one (%d) QoS profile with DisplayName '%s' for NSX-T Manager with ID '%s' found",
			len(nsxtEdgeClusters), displayName, nsxtManagerId)
	}

	return nsxtEdgeClusters[0], nil
}

func filterQosProfiles(displayName string, allQosProfiles []*NsxtEdgeGatewayQosProfile) []*NsxtEdgeGatewayQosProfile {
	filteredQosProfiles := make([]*NsxtEdgeGatewayQosProfile, 0)
	for index, nsxtEdgeCluster := range allQosProfiles {
		if allQosProfiles[index].NsxtEdgeGatewayQosProfile.DisplayName == displayName {
			filteredQosProfiles = append(filteredQosProfiles, nsxtEdgeCluster)
		}
	}

	return filteredQosProfiles
}
