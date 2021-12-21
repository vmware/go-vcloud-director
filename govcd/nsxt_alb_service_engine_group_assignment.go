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

// NsxtAlbServiceEngineGroupAssignment handles Service Engine Group Assignment to NSX-T Edge Gateways
type NsxtAlbServiceEngineGroupAssignment struct {
	NsxtAlbServiceEngineGroupAssignment *types.NsxtAlbServiceEngineGroupAssignment
	vcdClient                           *VCDClient
}

func (vcdClient *VCDClient) GetAllAlbServiceEngineGroupAssignments(queryParameters url.Values) ([]*NsxtAlbServiceEngineGroupAssignment, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtAlbServiceEngineGroupAssignment{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*NsxtAlbServiceEngineGroupAssignment, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbServiceEngineGroupAssignment{
			NsxtAlbServiceEngineGroupAssignment: typeResponses[sliceIndex],
			vcdClient:                           vcdClient,
		}
	}

	return wrappedResponses, nil
}

func (vcdClient *VCDClient) GetAlbServiceEngineGroupAssignmentById(id string) (*NsxtAlbServiceEngineGroupAssignment, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	typeResponse := &types.NsxtAlbServiceEngineGroupAssignment{}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponse := &NsxtAlbServiceEngineGroupAssignment{
		NsxtAlbServiceEngineGroupAssignment: typeResponse,
		vcdClient:                           vcdClient,
	}

	return wrappedResponse, nil
}

func (vcdClient *VCDClient) GetAlbServiceEngineGroupAssignmentByName(name string) (*NsxtAlbServiceEngineGroupAssignment, error) {
	// Filtering by Service Engine Group name is not supported on API therefore filtering is done locally
	allServiceEngineGroupAssignments, err := vcdClient.GetAllAlbServiceEngineGroupAssignments(nil)
	if err != nil {
		return nil, err
	}

	var foundGroup *NsxtAlbServiceEngineGroupAssignment

	for _, serviceEngineGroupAssignment := range allServiceEngineGroupAssignments {
		if serviceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name == name {
			foundGroup = serviceEngineGroupAssignment
		}
	}

	if foundGroup == nil {
		return nil, ErrorEntityNotFound
	}

	return foundGroup, nil
}

// GetFilteredAlbServiceEngineGroupAssignmentByName will get all ALB Service Engine Group assignments based on filters
// provided in queryParameters additionally will filter by name locally because VCD does not support server side
// filtering by name.
func (vcdClient *VCDClient) GetFilteredAlbServiceEngineGroupAssignmentByName(name string, queryParameters url.Values) (*NsxtAlbServiceEngineGroupAssignment, error) {
	// Filtering by Service Engine Group name is not supported on API therefore filtering is done locally
	allServiceEngineGroupAssignments, err := vcdClient.GetAllAlbServiceEngineGroupAssignments(queryParameters)
	if err != nil {
		return nil, err
	}

	var foundGroup *NsxtAlbServiceEngineGroupAssignment

	for _, serviceEngineGroupAssignment := range allServiceEngineGroupAssignments {
		if serviceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name == name {
			foundGroup = serviceEngineGroupAssignment
		}
	}

	if foundGroup == nil {
		return nil, ErrorEntityNotFound
	}

	return foundGroup, nil
}

func (vcdClient *VCDClient) CreateAlbServiceEngineGroupAssignment(assignmentConfig *types.NsxtAlbServiceEngineGroupAssignment) (*NsxtAlbServiceEngineGroupAssignment, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Service Engine Group Assignment require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtAlbServiceEngineGroupAssignment{
		NsxtAlbServiceEngineGroupAssignment: &types.NsxtAlbServiceEngineGroupAssignment{},
		vcdClient:                           vcdClient,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, assignmentConfig, returnObject.NsxtAlbServiceEngineGroupAssignment, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T ALB Service Engine Group Assignment: %s", err)
	}

	return returnObject, nil
}

// Update updates existing ALB Service Engine Group Assignment with new supplied assignmentConfig configuration
func (nsxtEdgeAlbServiceEngineGroup *NsxtAlbServiceEngineGroupAssignment) Update(assignmentConfig *types.NsxtAlbServiceEngineGroupAssignment) (*NsxtAlbServiceEngineGroupAssignment, error) {
	client := nsxtEdgeAlbServiceEngineGroup.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if assignmentConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T ALB Service Engine Group Assignment without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, assignmentConfig.ID)
	if err != nil {
		return nil, err
	}

	responseAlbController := &NsxtAlbServiceEngineGroupAssignment{
		NsxtAlbServiceEngineGroupAssignment: &types.NsxtAlbServiceEngineGroupAssignment{},
		vcdClient:                           nsxtEdgeAlbServiceEngineGroup.vcdClient,
	}

	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, assignmentConfig, responseAlbController.NsxtAlbServiceEngineGroupAssignment, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T ALB Service Engine Group Assignment: %s", err)
	}

	return responseAlbController, nil
}

// Delete deletes NSX-T ALB Service Engine Group Assignment
func (nsxtEdgeAlbServiceEngineGroup *NsxtAlbServiceEngineGroupAssignment) Delete() error {
	client := nsxtEdgeAlbServiceEngineGroup.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if nsxtEdgeAlbServiceEngineGroup.NsxtAlbServiceEngineGroupAssignment.ID == "" {
		return fmt.Errorf("cannot delete NSX-T ALB Service Engine Group Assignment without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, nsxtEdgeAlbServiceEngineGroup.NsxtAlbServiceEngineGroupAssignment.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T ALB Service Engine Group Assignment: %s", err)
	}

	return nil
}
