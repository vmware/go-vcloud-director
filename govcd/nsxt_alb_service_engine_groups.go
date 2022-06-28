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

// NsxtAlbServiceEngineGroup provides virtual service management capabilities for tenants. This entity can be created
// by referencing a backing importable service engine group - NsxtAlbImportableServiceEngineGroups.
//
// A service engine group is an isolation domain that also defines shared service engine properties, such as size,
// network access, and failover. Resources in a service engine group can be used for different virtual services,
// depending on your tenant needs. These resources cannot be shared between different service engine groups.
type NsxtAlbServiceEngineGroup struct {
	NsxtAlbServiceEngineGroup *types.NsxtAlbServiceEngineGroup
	vcdClient                 *VCDClient
}

// GetAllAlbServiceEngineGroups retrieves NSX-T ALB Service Engines with possible filters
//
// Context is not mandatory for this resource. Supported contexts are:
// * Gateway ID (_context==gatewayId) - returns all Load Balancer Service Engine Groups that are accessible to the
// gateway.
// * Assignable Gateway ID (_context=gatewayId;_context==assignable) returns all Load Balancer Service Engine Groups
// that are assignable to the gateway. This filters out any Load Balancer Service Engine groups that are already
// assigned to the gateway or assigned to another gateway if the reservation type is 'DEDICATED’.
func (vcdClient *VCDClient) GetAllAlbServiceEngineGroups(context string, queryParameters url.Values) ([]*NsxtAlbServiceEngineGroup, error) {
	client := vcdClient.Client

	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Service Engine Groups require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
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
// Context is not mandatory for this resource. Supported contexts are:
// * Gateway ID (_context==gatewayId) - returns all Load Balancer Service Engine Groups that are accessible to the
// gateway.
// * Assignable Gateway ID (_context=gatewayId;_context==assignable) returns all Load Balancer Service Engine Groups
// that are assignable to the gateway. This filters out any Load Balancer Service Engine groups that are already
// assigned to the gateway or assigned to another gateway if the reservation type is 'DEDICATED’.
func (vcdClient *VCDClient) GetAlbServiceEngineGroupByName(optionalContext, name string) (*NsxtAlbServiceEngineGroup, error) {
	queryParams := copyOrNewUrlValues(nil)
	if optionalContext != "" {
		queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", optionalContext), queryParams)
	}
	queryParams.Add("filter", fmt.Sprintf("name==%s", name))

	albSeGroups, err := vcdClient.GetAllAlbServiceEngineGroups("", queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T ALB Service Engine Group By Name '%s': %s", name, err)
	}

	if len(albSeGroups) == 0 {
		return nil, fmt.Errorf("%s", ErrorEntityNotFound)
	}

	if len(albSeGroups) > 1 {
		return nil, fmt.Errorf("more than 1 NSX-T ALB Service Engine Group with Name '%s' found", name)
	}

	return albSeGroups[0], nil
}

// GetAlbServiceEngineGroupById returns importable NSX-T ALB Cloud by ID
func (vcdClient *VCDClient) GetAlbServiceEngineGroupById(id string) (*NsxtAlbServiceEngineGroup, error) {
	client := vcdClient.Client

	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Service Engine Groups require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	typeResponse := &types.NsxtAlbServiceEngineGroup{}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponse := &NsxtAlbServiceEngineGroup{
		NsxtAlbServiceEngineGroup: typeResponse,
		vcdClient:                 vcdClient,
	}

	return wrappedResponse, nil
}

func (vcdClient *VCDClient) CreateNsxtAlbServiceEngineGroup(albServiceEngineGroup *types.NsxtAlbServiceEngineGroup) (*NsxtAlbServiceEngineGroup, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Service Engine Groups require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
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

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, albServiceEngineGroup, returnObject.NsxtAlbServiceEngineGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T ALB Service Engine Group: %s", err)
	}

	return returnObject, nil
}

// Update updates existing ALB Controller with new supplied albControllerConfig configuration
func (nsxtAlbServiceEngineGroup *NsxtAlbServiceEngineGroup) Update(albSEGroupConfig *types.NsxtAlbServiceEngineGroup) (*NsxtAlbServiceEngineGroup, error) {
	client := nsxtAlbServiceEngineGroup.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
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

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, albSEGroupConfig, responseAlbController.NsxtAlbServiceEngineGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T ALB Service Engine Group: %s", err)
	}

	return responseAlbController, nil
}

// Delete deletes NSX-T ALB Service Engine Group configuration
func (nsxtAlbServiceEngineGroup *NsxtAlbServiceEngineGroup) Delete() error {
	client := nsxtAlbServiceEngineGroup.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
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

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T ALB Service Engine Group: %s", err)
	}

	return nil
}

// Sync syncs a specified Load Balancer Service Engine Group. It requests the HA mode and the maximum number of
// supported Virtual Services for this Service Engine Group from the Load Balancer, and updates vCD's local record of
// these properties.
func (nsxtAlbServiceEngineGroup *NsxtAlbServiceEngineGroup) Sync() error {
	client := nsxtAlbServiceEngineGroup.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if nsxtAlbServiceEngineGroup.NsxtAlbServiceEngineGroup.ID == "" {
		return fmt.Errorf("cannot sync NSX-T ALB Service Engine Group without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, nsxtAlbServiceEngineGroup.NsxtAlbServiceEngineGroup.ID, "/sync")
	if err != nil {
		return err
	}

	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error syncing NSX-T ALB Service Engine Group: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("sync task for NSX-T ALB Service Engine Group failed: %s", err)
	}

	return nil
}
