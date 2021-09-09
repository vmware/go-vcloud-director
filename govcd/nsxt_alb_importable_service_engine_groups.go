/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtAlbImportableServiceEngineGroups provides capability to list all Importable Service Engine Groups available in
// ALB Controller so that they can be consumed by NsxtAlbServiceEngineGroup
//
// Note. The API does not return Importable Service Engine Group once it is consumed.
type NsxtAlbImportableServiceEngineGroups struct {
	NsxtAlbImportableServiceEngineGroups *types.NsxtAlbImportableServiceEngineGroups
	vcdClient                            *VCDClient
}

// GetAllAlbImportableServiceEngineGroups lists all Importable Service Engine Groups available in ALB Controller
func (vcdClient *VCDClient) GetAllAlbImportableServiceEngineGroups(parentAlbCloudUrn string, queryParameters url.Values) ([]*NsxtAlbImportableServiceEngineGroups, error) {
	client := vcdClient.Client
	if parentAlbCloudUrn == "" {
		return nil, fmt.Errorf("parentAlbCloudUrn is required")
	}
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Importable Service Engine Groups requires System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbImportableServiceEngineGroups
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", parentAlbCloudUrn), queryParams)
	typeResponses := []*types.NsxtAlbImportableServiceEngineGroups{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*NsxtAlbImportableServiceEngineGroups, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbImportableServiceEngineGroups{
			NsxtAlbImportableServiceEngineGroups: typeResponses[sliceIndex],
			vcdClient:                            vcdClient,
		}
	}

	return wrappedResponses, nil
}

// GetAlbImportableServiceEngineGroupByName returns importable NSX-T ALB Clouds.
func (vcdClient *VCDClient) GetAlbImportableServiceEngineGroupByName(parentAlbCloudUrn, name string) (*NsxtAlbImportableServiceEngineGroups, error) {
	albClouds, err := vcdClient.GetAllAlbImportableServiceEngineGroups(parentAlbCloudUrn, nil)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Importable Service Engine Group by Name '%s': %s", name, err)
	}

	// Filtering by Name is not supported by API therefore it must be filtered on client side
	var foundResult bool
	var foundAlbCloud *NsxtAlbImportableServiceEngineGroups
	for i, value := range albClouds {
		if albClouds[i].NsxtAlbImportableServiceEngineGroups.DisplayName == name {
			foundResult = true
			foundAlbCloud = value
			break
		}
	}

	if !foundResult {
		return nil, fmt.Errorf("%s: could not find NSX-T ALB Importable Service Engine Group by Name %s", ErrorEntityNotFound, name)
	}

	return foundAlbCloud, nil
}

// GetAlbImportableServiceEngineGroupById
// Note. ID filtering is performed on client side
func (vcdClient *VCDClient) GetAlbImportableServiceEngineGroupById(parentAlbCloudUrn, id string) (*NsxtAlbImportableServiceEngineGroups, error) {
	albClouds, err := vcdClient.GetAllAlbImportableServiceEngineGroups(parentAlbCloudUrn, nil)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Importable Service Engine Group by ID '%s': %s", id, err)
	}

	// Filtering by ID is not supported by API therefore it must be filtered on client side
	var foundResult bool
	var foundImportableSEGroups *NsxtAlbImportableServiceEngineGroups
	for i, value := range albClouds {
		if albClouds[i].NsxtAlbImportableServiceEngineGroups.ID == id {
			foundResult = true
			foundImportableSEGroups = value
		}
	}

	if !foundResult {
		return nil, fmt.Errorf("%s: could not find NSX-T ALB Importable Service Engine Group by ID %s", ErrorEntityNotFound, id)
	}

	return foundImportableSEGroups, nil
}

// GetAllAlbImportableServiceEngineGroups lists all Importable Service Engine Groups available in ALB Controller
func (nsxtAlbCloud *NsxtAlbCloud) GetAllAlbImportableServiceEngineGroups(parentAlbCloudUrn string, queryParameters url.Values) ([]*NsxtAlbImportableServiceEngineGroups, error) {
	client := nsxtAlbCloud.vcdClient.Client
	if parentAlbCloudUrn == "" {
		return nil, fmt.Errorf("parentAlbCloudUrn is required")
	}
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Importable Service Engine Groups requires System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbImportableServiceEngineGroups
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", parentAlbCloudUrn), queryParams)
	typeResponses := []*types.NsxtAlbImportableServiceEngineGroups{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*NsxtAlbImportableServiceEngineGroups, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbImportableServiceEngineGroups{
			NsxtAlbImportableServiceEngineGroups: typeResponses[sliceIndex],
			vcdClient:                            nsxtAlbCloud.vcdClient,
		}
	}

	return wrappedResponses, nil
}

// GetAlbImportableServiceEngineGroupByName returns importable NSX-T ALB Clouds.
func (nsxtAlbCloud *NsxtAlbCloud) GetAlbImportableServiceEngineGroupByName(parentAlbCloudUrn, name string) (*NsxtAlbImportableServiceEngineGroups, error) {
	albClouds, err := nsxtAlbCloud.vcdClient.GetAllAlbImportableServiceEngineGroups(parentAlbCloudUrn, nil)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Importable Service Engine Group by Name '%s': %s", name, err)
	}

	// Filtering by ID is not supported by API therefore it must be filtered on client side
	var foundResult bool
	var foundAlbCloud *NsxtAlbImportableServiceEngineGroups
	for i, value := range albClouds {
		if albClouds[i].NsxtAlbImportableServiceEngineGroups.DisplayName == name {
			foundResult = true
			foundAlbCloud = value
			break
		}
	}

	if !foundResult {
		return nil, fmt.Errorf("%s: could not find NSX-T ALB Importable Service Engine Group by Name %s", ErrorEntityNotFound, name)
	}

	return foundAlbCloud, nil
}

// GetAlbImportableServiceEngineGroupById
// Note. ID filtering is performed on client side
func (nsxtAlbCloud *NsxtAlbCloud) GetAlbImportableServiceEngineGroupById(parentAlbCloudUrn, id string) (*NsxtAlbImportableServiceEngineGroups, error) {
	albClouds, err := nsxtAlbCloud.vcdClient.GetAllAlbImportableServiceEngineGroups(parentAlbCloudUrn, nil)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Importable Service Engine Group by ID '%s': %s", id, err)
	}

	// Filtering by ID is not supported by API therefore it must be filtered on client side
	var foundResult bool
	var foundImportableSEGroups *NsxtAlbImportableServiceEngineGroups
	for i, value := range albClouds {
		if albClouds[i].NsxtAlbImportableServiceEngineGroups.ID == id {
			foundResult = true
			foundImportableSEGroups = value
		}
	}

	if !foundResult {
		return nil, fmt.Errorf("%s: could not find NSX-T ALB Importable Service Engine Group by ID %s", ErrorEntityNotFound, id)
	}

	return foundImportableSEGroups, nil
}
