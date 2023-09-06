/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

func (vcdClient *VCDClient) GetAllNsxtTransportZones(nsxtManagerId string, queryParameters url.Values) ([]*types.TransportZone, error) {
	if nsxtManagerId == "" {
		return nil, fmt.Errorf("empty NSX-T manager ID")
	}

	if !isUrn(nsxtManagerId) {
		return nil, fmt.Errorf("NSX-T manager ID is not URN (e.g. 'urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc)', got: %s", nsxtManagerId)
	}

	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableTransportZones
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	queryParams := copyOrNewUrlValues(queryParameters)
	filterField := "_context"
	if client.APIClientVersionIs(">=38.0") {
		// field "networkProviderId" does not exist prior to API 38.0, where field "_context" is deprecated
		filterField = "networkProviderId"
	}
	queryParams = queryParameterFilterAnd(fmt.Sprintf("%s==%s", filterField, nsxtManagerId), queryParams)
	var typeResponses []*types.TransportZone
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}
