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

// NsxtAlbServiceEngineGroup allows users to provide virtual service management capabilities to your tenants, import
// service engine groups to your VMware Cloud Director deployment.
// A service engine group is an isolation domain that also defines shared service engine properties, such as size,
// network access, and failover. Resources in a service engine group can be used for different virtual services,
// depending on your tenant needs. These resources cannot be shared between different service engine groups.
type NsxtAlbServiceEngineGroup struct {
	NsxtAlbServiceEngineGroup *types.NsxtAlbServiceEngineGroup
	vcdClient                 *VCDClient
}

// GetAllNsxtAlbServiceEngineGroups retrieves NSX-T ALB Service Engines with possible filters
func (vcdClient *VCDClient) GetAllNsxtAlbServiceEngineGroups(context string, queryParameters url.Values) ([]*NsxtAlbServiceEngineGroup, error) {
	client := vcdClient.Client

	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Service Engine Groups require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	queryParams := copyOrNewUrlValues(queryParameters)
	if context != "" {
		queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", context), queryParams)
	}
	typeResponses := []*types.NsxtAlbServiceEngineGroup{{}}

	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*NsxtAlbServiceEngineGroup, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbServiceEngineGroup{
			NsxtAlbServiceEngineGroup: typeResponses[sliceIndex],
			vcdClient:                 vcdClient,
		}
	}

	return wrappedResponses, nil
}

// GetAlbServiceEngineGroupByName returns NSX-T ALB Service Engine by Name
func (vcdClient *VCDClient) GetAlbServiceEngineGroupByName(context, name string) (*NsxtAlbServiceEngineGroup, error) {
	albClouds, err := vcdClient.GetAllNsxtAlbServiceEngineGroups(context, nil)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Service Engine Group by Name '%s': %s", name, err)
	}

	// Filtering by ID is not supported therefore it must be filtered on client side
	var foundResult bool
	var foundAlbCloud *NsxtAlbServiceEngineGroup
	for i, value := range albClouds {
		if albClouds[i].NsxtAlbServiceEngineGroup.Name == name {
			foundResult = true
			foundAlbCloud = value
			break
		}
	}

	if !foundResult {
		return nil, fmt.Errorf("%s: could not find NSX-T ALB Service Engine Group by Name %s", ErrorEntityNotFound, name)
	}

	return foundAlbCloud, nil
}

// GetAlbServiceEngineGroupById returns importable NSX-T ALB Clouds.
// Note. ID filtering is performed on client side
func (vcdClient *VCDClient) GetAlbServiceEngineGroupById(context, id string) (*NsxtAlbServiceEngineGroup, error) {
	albClouds, err := vcdClient.GetAllNsxtAlbServiceEngineGroups(context, nil)
	if err != nil {
		return nil, fmt.Errorf("error finding Alb Cloud by ID '%s': %s", id, err)
	}

	// Filtering by ID is not supported therefore it must be filtered on client side
	var foundResult bool
	var foundAlbCloud *NsxtAlbServiceEngineGroup
	for i, value := range albClouds {
		if albClouds[i].NsxtAlbServiceEngineGroup.ID == id {
			foundResult = true
			foundAlbCloud = value
		}
	}

	if !foundResult {
		return nil, fmt.Errorf("%s: could not find NSX-T ALB Service Engine Group by ID %s", ErrorEntityNotFound, id)
	}

	return foundAlbCloud, nil
}

func (vcdClient *VCDClient) CreateNsxtAlbServiceEngineGroup(albServiceEngineGroup *types.NsxtAlbServiceEngineGroup) (*NsxtAlbServiceEngineGroup, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Service Engine Groups require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtAlbServiceEngineGroup{
		NsxtAlbServiceEngineGroup: &types.NsxtAlbServiceEngineGroup{},
		vcdClient:                 vcdClient,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, albServiceEngineGroup, returnObject.NsxtAlbServiceEngineGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T ALB Service Engine Group: %s", err)
	}

	return returnObject, nil
}

//
//// Update updates existing ALB Controller with new supplied albControllerConfig configuration
func (nsxtAlbServiceEngineGroup *NsxtAlbServiceEngineGroup) Update(albSEGroupConfig *types.NsxtAlbServiceEngineGroup) (*NsxtAlbServiceEngineGroup, error) {
	client := nsxtAlbServiceEngineGroup.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if albSEGroupConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T ALB Service Engine Group without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, albSEGroupConfig.ID)
	if err != nil {
		return nil, err
	}

	responseAlbController := &NsxtAlbServiceEngineGroup{
		NsxtAlbServiceEngineGroup: &types.NsxtAlbServiceEngineGroup{},
		vcdClient:                 nsxtAlbServiceEngineGroup.vcdClient,
	}

	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, albSEGroupConfig, responseAlbController.NsxtAlbServiceEngineGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T ALB Service Engine Group: %s", err)
	}

	return responseAlbController, nil
}

func (nsxtAlbServiceEngineGroup *NsxtAlbServiceEngineGroup) Delete() error {
	client := nsxtAlbServiceEngineGroup.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if nsxtAlbServiceEngineGroup.NsxtAlbServiceEngineGroup.ID == "" {
		return fmt.Errorf("cannot delete NSX-T ALB Service Engine Group without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, nsxtAlbServiceEngineGroup.NsxtAlbServiceEngineGroup.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T ALB Service Engine Group: %s", err)
	}

	return nil
}
