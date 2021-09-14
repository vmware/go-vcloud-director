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

// NsxtAlbImportableCloud allows user to list importable NSX-T ALB Clouds. Each importable cloud can only be imported
// once by using NsxtAlbCloud construct. It has a flag AlreadyImported which hints if it is already consumed or not.
type NsxtAlbImportableCloud struct {
	NsxtAlbImportableCloud *types.NsxtAlbImportableCloud
	vcdClient              *VCDClient
}

// GetAllAlbImportableClouds returns importable NSX-T ALB Clouds.
// parentAlbControllerUrn (ID in URN format of a parent ALB Controller) is mandatory
func (vcdClient *VCDClient) GetAllAlbImportableClouds(parentAlbControllerUrn string, queryParameters url.Values) ([]*NsxtAlbImportableCloud, error) {
	client := vcdClient.Client
	if parentAlbControllerUrn == "" {
		return nil, fmt.Errorf("parent ALB Controller ID is required")
	}
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Importable Clouds require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbImportableClouds
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", parentAlbControllerUrn), queryParams)
	typeResponses := []*types.NsxtAlbImportableCloud{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*NsxtAlbImportableCloud, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbImportableCloud{
			NsxtAlbImportableCloud: typeResponses[sliceIndex],
			vcdClient:              vcdClient,
		}
	}

	return wrappedResponses, nil
}

// GetAlbImportableCloudByName returns importable NSX-T ALB Clouds.
func (vcdClient *VCDClient) GetAlbImportableCloudByName(parentAlbControllerUrn, name string) (*NsxtAlbImportableCloud, error) {
	albImportableClouds, err := vcdClient.GetAllAlbImportableClouds(parentAlbControllerUrn, nil)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Importable Cloud by Name '%s': %s", name, err)
	}

	// Filtering by Name is not supported by API therefore it must be filtered on client side
	var foundResult bool
	var foundAlbImportableCloud *NsxtAlbImportableCloud
	for i, value := range albImportableClouds {
		if albImportableClouds[i].NsxtAlbImportableCloud.DisplayName == name {
			foundResult = true
			foundAlbImportableCloud = value
			break
		}
	}

	if !foundResult {
		return nil, fmt.Errorf("%s: could not find NSX-T ALB Importable Cloud by Name %s", ErrorEntityNotFound, name)
	}

	return foundAlbImportableCloud, nil
}

// GetAlbImportableCloudById returns importable NSX-T ALB Clouds.
// Note. ID filtering is performed on client side
func (vcdClient *VCDClient) GetAlbImportableCloudById(parentAlbControllerUrn, id string) (*NsxtAlbImportableCloud, error) {
	albImportableClouds, err := vcdClient.GetAllAlbImportableClouds(parentAlbControllerUrn, nil)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Importable Cloud by ID '%s': %s", id, err)
	}

	// Filtering by ID is not supported by API therefore it must be filtered on client side
	var foundResult bool
	var foundAlbImportableCloud *NsxtAlbImportableCloud
	for i, value := range albImportableClouds {
		if albImportableClouds[i].NsxtAlbImportableCloud.ID == id {
			foundResult = true
			foundAlbImportableCloud = value
		}
	}

	if !foundResult {
		return nil, fmt.Errorf("%s: could not find NSX-T ALB Importable Cloud by ID %s", ErrorEntityNotFound, id)
	}

	return foundAlbImportableCloud, nil
}

// GetAllAlbImportableClouds is attached to NsxtAlbController type for a convenient parent/child relationship
func (nsxtAlbController *NsxtAlbController) GetAllAlbImportableClouds(queryParameters url.Values) ([]*NsxtAlbImportableCloud, error) {
	return nsxtAlbController.vcdClient.GetAllAlbImportableClouds(nsxtAlbController.NsxtAlbController.ID, queryParameters)
}

// GetAlbImportableCloudByName is attached to NsxtAlbController type for a convenient parent/child relationship
func (nsxtAlbController *NsxtAlbController) GetAlbImportableCloudByName(name string) (*NsxtAlbImportableCloud, error) {
	return nsxtAlbController.vcdClient.GetAlbImportableCloudByName(nsxtAlbController.NsxtAlbController.ID, name)
}
