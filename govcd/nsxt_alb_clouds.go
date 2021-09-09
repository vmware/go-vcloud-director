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

// NsxtAlbCloud helps to use the virtual infrastructure provided by NSX Advanced Load Balancer, register NSX-T Cloud
// instances with VMware Cloud Director by consuming NsxtAlbImportableCloud.
type NsxtAlbCloud struct {
	NsxtAlbCloud *types.NsxtAlbCloud
	vcdClient    *VCDClient
}

// GetAllAlbClouds returns all configured NSX-T ALB Clouds
func (vcdClient *VCDClient) GetAllAlbClouds(queryParameters url.Values) ([]*NsxtAlbCloud, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Clouds require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbCloud
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtAlbCloud{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtAlbCloud types with client
	wrappedResponses := make([]*NsxtAlbCloud, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbCloud{
			NsxtAlbCloud: typeResponses[sliceIndex],
			vcdClient:    vcdClient,
		}
	}

	return wrappedResponses, nil
}

// GetAlbCloudByName returns NSX-T ALB Cloud by name
func (vcdClient *VCDClient) GetAlbCloudByName(name string) (*NsxtAlbCloud, error) {
	queryParameters := copyOrNewUrlValues(nil)
	queryParameters.Add("filter", "name=="+name)

	albClouds, err := vcdClient.GetAllAlbClouds(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error reading NSX-T ALB Cloud with Name '%s': %s", name, err)
	}

	if len(albClouds) == 0 {
		return nil, fmt.Errorf("%s could not find NSX-T ALB Cloud with Name '%s'", ErrorEntityNotFound, name)
	}

	if len(albClouds) > 1 {
		return nil, fmt.Errorf("found more than 1 NSX-T ALB Cloud with Name '%s'", name)
	}

	return albClouds[0], nil
}

// GetAlbCloudById returns NSX-T ALB Cloud by ID
//
// Note. This function uses server side filtering instead of directly querying endpoint with specified ID because such
// endpoint does not exist
func (vcdClient *VCDClient) GetAlbCloudById(id string) (*NsxtAlbCloud, error) {

	queryParameters := copyOrNewUrlValues(nil)
	queryParameters.Add("filter", "id=="+id)

	albCloud, err := vcdClient.GetAllAlbClouds(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error reading NSX-T ALB Cloud with ID '%s': %s", id, err)
	}

	if len(albCloud) == 0 {
		return nil, fmt.Errorf("%s could not find NSX-T ALB Cloud by ID '%s'", ErrorEntityNotFound, id)
	}

	return albCloud[0], nil
}

// CreateAlbCloud creates NSX-T ALB Cloud
func (vcdClient *VCDClient) CreateAlbCloud(albCloudConfig *types.NsxtAlbCloud) (*NsxtAlbCloud, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Clouds require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbCloud
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtAlbCloud{
		NsxtAlbCloud: &types.NsxtAlbCloud{},
		vcdClient:    vcdClient,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, albCloudConfig, returnObject.NsxtAlbCloud, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T ALB Cloud: %s", err)
	}

	return returnObject, nil
}

// Update is not supported in VCD 10.3 and older therefore this function remains commented
//
// Update updates existing NSX-T ALB Cloud with new supplied albCloudConfig configuration
//func (nsxtAlbCloud *NsxtAlbCloud) Update(albCloudConfig *types.NsxtAlbCloud) (*NsxtAlbCloud, error) {
//	client := nsxtAlbCloud.vcdClient.Client
//	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbCloud
//	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
//	if err != nil {
//		return nil, err
//	}
//
//	if albCloudConfig.ID == "" {
//		return nil, fmt.Errorf("cannot update NSX-T ALB Cloud without ID")
//	}
//
//	urlRef, err := client.OpenApiBuildEndpoint(endpoint, albCloudConfig.ID)
//	if err != nil {
//		return nil, err
//	}
//
//	responseAlbCloud := &NsxtAlbCloud{
//		NsxtAlbCloud: &types.NsxtAlbCloud{},
//		vcdClient:    nsxtAlbCloud.vcdClient,
//	}
//
//	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, albCloudConfig, responseAlbCloud.NsxtAlbCloud, nil)
//	if err != nil {
//		return nil, fmt.Errorf("error updating NSX-T ALB Cloud: %s", err)
//	}
//
//	return responseAlbCloud, nil
//}

// Delete removes NSX-T ALB Cloud configuration
func (nsxtAlbCloud *NsxtAlbCloud) Delete() error {
	client := nsxtAlbCloud.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbCloud
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if nsxtAlbCloud.NsxtAlbCloud.ID == "" {
		return fmt.Errorf("cannot delete NSX-T ALB Cloud without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, nsxtAlbCloud.NsxtAlbCloud.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T ALB Cloud: %s", err)
	}

	return nil
}
