/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

type VCenter struct {
	VSphereVcenter *types.VSphereVirtualCenter
	client         *VCDClient
}

func NewVcenter(client *VCDClient) *VCenter {
	return &VCenter{
		VSphereVcenter: &types.VSphereVirtualCenter{},
		client:         client,
	}
}

func (vcdClient VCDClient) GetAllVcenters(queryParams url.Values) ([]*VCenter, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var retrieved []*types.VSphereVirtualCenter

	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParams, &retrieved, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting vCenters list: %s", err)
	}

	if len(retrieved) == 0 {
		return nil, nil
	}
	var returnList []*VCenter

	for _, r := range retrieved {
		returnList = append(returnList, &VCenter{
			VSphereVcenter: r,
			client:         &vcdClient,
		})
	}
	return returnList, nil
}

func (vcdClient VCDClient) GetVcenterByName(name string) (*VCenter, error) {
	vcenters, err := vcdClient.GetAllVcenters(nil)
	if err != nil {
		return nil, err
	}
	for _, vc := range vcenters {
		if vc.VSphereVcenter.Name == name {
			return vc, nil
		}
	}
	return nil, fmt.Errorf("vcenter %s not found: %s", name, ErrorEntityNotFound)
}

func (vcdClient VCDClient) GetVcenterById(id string) (*VCenter, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint + "/" + id)
	if err != nil {
		return nil, err
	}

	returnObject := &VCenter{
		VSphereVcenter: &types.VSphereVirtualCenter{},
		client:         &vcdClient,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, returnObject.VSphereVcenter, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting vCenter: %s", err)
	}

	return returnObject, nil
}

func (vcenter VCenter) GetVimServerUrl() (string, error) {
	return url.JoinPath(vcenter.client.Client.VCDHREF.String(), "admin", "extension", "vimServer", extractUuid(vcenter.VSphereVcenter.VcId))
}
