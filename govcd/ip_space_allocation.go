/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// IpSpaceIpAllocation handles IP Space IP allocation requests
type IpSpaceIpAllocation struct {
	IpSpaceIpAllocation *types.IpSpaceIpAllocation
	IpSpaceId           string

	client *Client
	// Org context must be sent with requests
	parent organization
}

// AllocateIp performs IP Allocation request for a specific Org and returns the result
func (ipSpace *IpSpace) AllocateIp(orgId, orgName string, ipAllocationConfig *types.IpSpaceIpAllocationRequest) ([]types.IpSpaceIpAllocationRequestResult, error) {
	return allocateIpSpaceIp(&ipSpace.vcdClient.Client, orgId, orgName, ipSpace.IpSpace.ID, ipAllocationConfig)
}

// IpSpaceAllocateIp performs IP allocation request for a specific IP Space
func (org *Org) IpSpaceAllocateIp(ipSpaceId string, ipAllocationConfig *types.IpSpaceIpAllocationRequest) ([]types.IpSpaceIpAllocationRequestResult, error) {
	return allocateIpSpaceIp(org.client, org.Org.ID, org.Org.Name, ipSpaceId, ipAllocationConfig)
}

// GetIpSpaceAllocationByTypeAndValue retrieves IP Space allocation by its type and value
// allocationType can be 'FLOATING_IP' (types.IpSpaceIpAllocationTypeFloatingIp) or 'IP_PREFIX'
// (types.IpSpaceIpAllocationTypeIpPrefix)
func (org *Org) GetIpSpaceAllocationByTypeAndValue(ipSpaceId string, allocationType, value string, queryParameters url.Values) (*IpSpaceIpAllocation, error) {
	queryParams := queryParameterFilterAnd(fmt.Sprintf("value==%s;type==%s", value, allocationType), queryParameters)
	results, err := getAllIpSpaceAllocations(org.client, ipSpaceId, org, queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving IP allocations: %s", err)
	}

	singleResult, err := oneOrError("value", value, results)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetAllIpSpaceAllocations retrieves all IP Allocations for a particular IP Space
// allocationType can be 'FLOATING_IP' (types.IpSpaceIpAllocationTypeFloatingIp) or 'IP_PREFIX'
// (types.IpSpaceIpAllocationTypeIpPrefix)
func (ipSpace *IpSpace) GetAllIpSpaceAllocations(allocationType string, queryParameters url.Values) ([]*IpSpaceIpAllocation, error) {
	if allocationType == "" {
		return nil, fmt.Errorf("allocationType is mandatory and must be 'FLOATING_IP' or 'IP_PREFIX'")
	}

	client := ipSpace.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceIpAllocations
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSpace.IpSpace.ID))
	if err != nil {
		return nil, err
	}

	queryParams := queryParameterFilterAnd(fmt.Sprintf("type==%s", allocationType), queryParameters)
	typeResponses := []*types.IpSpaceIpAllocation{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into IpSpaceIpAllocation types with client
	results := make([]*IpSpaceIpAllocation, len(typeResponses))
	for sliceIndex := range typeResponses {
		results[sliceIndex] = &IpSpaceIpAllocation{
			IpSpaceIpAllocation: typeResponses[sliceIndex],
			client:              &client,
			IpSpaceId:           ipSpace.IpSpace.ID,
			parent: &Org{
				Org: &types.Org{
					ID:   typeResponses[sliceIndex].OrgRef.ID,
					Name: typeResponses[sliceIndex].OrgRef.Name},
			},
		}
	}

	return results, nil
}

// GetIpSpaceAllocationById retrieves IP Allocation in a given IP Space by IDs
func (org *Org) GetIpSpaceAllocationById(ipSpaceId, allocationId string) (*IpSpaceIpAllocation, error) {
	if ipSpaceId == "" || allocationId == "" {
		return nil, fmt.Errorf("ipSpaceId and allocationId cannot be empty")
	}
	client := org.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceIpAllocations
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSpaceId), allocationId)
	if err != nil {
		return nil, err
	}

	response := &IpSpaceIpAllocation{
		IpSpaceIpAllocation: &types.IpSpaceIpAllocation{},
		parent:              org,
		IpSpaceId:           ipSpaceId,
		client:              client,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, response.IpSpaceIpAllocation, nil)
	if err != nil {
		return nil, err
	}

	return response, nil

}

// Update updates IP Allocation with a given configuration
func (ipSpaceAllocation *IpSpaceIpAllocation) Update(ipSpaceAllocationConfig *types.IpSpaceIpAllocation) (*IpSpaceIpAllocation, error) {
	client := ipSpaceAllocation.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceIpAllocations
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSpaceAllocation.IpSpaceId), ipSpaceAllocation.IpSpaceIpAllocation.ID)
	if err != nil {
		return nil, err
	}

	returnIpSpaceAllocation := &IpSpaceIpAllocation{
		IpSpaceIpAllocation: &types.IpSpaceIpAllocation{},
		client:              client,
		parent:              ipSpaceAllocation.parent,
		IpSpaceId:           ipSpaceAllocation.IpSpaceId,
	}

	tenantContext, err := ipSpaceAllocation.getTenantContext()
	if err != nil {
		return nil, err
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, ipSpaceAllocationConfig, returnIpSpaceAllocation.IpSpaceIpAllocation, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, fmt.Errorf("error updating IP Space IP Allocation: %s", err)
	}

	return returnIpSpaceAllocation, nil
}

// Delete removes IP Allocation
func (ipSpaceAllocation *IpSpaceIpAllocation) Delete() error {
	if ipSpaceAllocation == nil || ipSpaceAllocation.IpSpaceIpAllocation == nil || ipSpaceAllocation.IpSpaceIpAllocation.ID == "" {
		return fmt.Errorf("IP Space IP Allocation must have ID")
	}

	if ipSpaceAllocation.IpSpaceId == "" || ipSpaceAllocation.parent == nil {
		return fmt.Errorf("incomplete IpSpaceIpAllocation type")
	}

	client := ipSpaceAllocation.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceIpAllocations
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	tenantContext, err := ipSpaceAllocation.getTenantContext()
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSpaceAllocation.IpSpaceId), ipSpaceAllocation.IpSpaceIpAllocation.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, getTenantContextHeader(tenantContext))
	if err != nil {
		return fmt.Errorf("error deleting IP Space IP Allocation: %s", err)
	}

	return nil
}

func allocateIpSpaceIp(client *Client, orgId, orgName, ipSpaceId string, ipAllocationConfig *types.IpSpaceIpAllocationRequest) ([]types.IpSpaceIpAllocationRequestResult, error) {
	if orgId == "" || orgName == "" || ipSpaceId == "" {
		return nil, fmt.Errorf("IP Space must have all values Org ID, Org Name and IP Space ID populated")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinksAllocate
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSpaceId))
	if err != nil {
		return nil, err
	}

	tenantContext := &TenantContext{
		OrgId:   orgId,
		OrgName: orgName,
	}

	task, err := client.OpenApiPostItemAsyncWithHeaders(apiVersion, urlRef, nil, ipAllocationConfig, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, fmt.Errorf("error triggering IP Allocation task for IP Space '%s': %s", ipSpaceId, err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error waiting for task completion: %s", err)
	}

	// Result of the task should contain a JSON with allocated IP details
	if task.Task == nil || task.Task.Result == nil || task.Task.Result.ResultContent.Text == "" {
		return nil, fmt.Errorf("error finding allocated IP result in task")
	}
	result := task.Task.Result.ResultContent.Text

	unmarshalStorage := []types.IpSpaceIpAllocationRequestResult{}
	err = json.Unmarshal([]byte(result), &unmarshalStorage)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling task result: %s", err)
	}

	return unmarshalStorage, nil
}

func getAllIpSpaceAllocations(client *Client, ipSpaceId string, org *Org, queryParameters url.Values) ([]*IpSpaceIpAllocation, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceIpAllocations
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSpaceId))
	if err != nil {
		return nil, err
	}

	tenantContext, err := org.getTenantContext()
	if err != nil {
		return nil, fmt.Errorf("error getting tenant context: %s", err)
	}

	typeResponses := []*types.IpSpaceIpAllocation{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into IpSpaceIpAllocation types with client
	results := make([]*IpSpaceIpAllocation, len(typeResponses))
	for sliceIndex := range typeResponses {
		results[sliceIndex] = &IpSpaceIpAllocation{
			IpSpaceIpAllocation: typeResponses[sliceIndex],
			client:              client,
			parent:              org,
			IpSpaceId:           ipSpaceId,
		}
	}

	return results, nil
}
