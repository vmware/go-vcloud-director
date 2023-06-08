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

// IpSpaceIpAllocation
type IpSpaceIpAllocation struct {
	IpSpaceIpAllocation *types.IpSpaceIpAllocation
	IpSpaceId           string

	client *Client
	// Org context must be sent with requests
	parent organization
}

func (ipSpace *IpSpace) AllocateIp(orgId, orgName string, ipAllocationConfig *types.IpSpaceIpAllocationRequest) ([]types.IpSpaceIpAllocationRequestResult, error) {
	return allocateIpSpaceIp(&ipSpace.vcdClient.Client, orgId, orgName, ipSpace.IpSpace.ID, ipAllocationConfig)
}

func (org *Org) IpSpaceAllocateIp(ipSpaceId string, ipAllocationConfig *types.IpSpaceIpAllocationRequest) ([]types.IpSpaceIpAllocationRequestResult, error) {
	return allocateIpSpaceIp(org.client, org.Org.ID, org.Org.Name, ipSpaceId, ipAllocationConfig)
}

func (org *Org) GetAllIpSpaceAllocations(ipSpaceId string, queryParameters url.Values) ([]*IpSpaceIpAllocation, error) {
	client := org.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinksAllocations
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSpaceId))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.IpSpaceIpAllocation{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into DefinedEntityType types with client
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

// func (org *Org) GetIpSpaceAllocationByTypeAndName(ipSpaceId string, ipSpaceType string, queryParameters url.Values) (*IpSpaceIpAllocation, error) {
// 	queryParams := queryParameterFilterAnd(fmt.Sprintf("type==%s", ipSpaceType), queryParameters)

// 	results, err := org.GetAllIpSpaceAllocations(ipSpaceId, queryParams)
// 	if err != nil {
// 		return nil, fmt.Errorf("error retrieving IP allocations: %s", err)
// 	}

// 	singleResult, err := oneOrError("type", ipSpaceType, results)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return singleResult, nil
// }

func (org *Org) GetIpSpaceAllocationByTypeAndValue(ipSpaceId string, allocationType, value string, queryParameters url.Values) (*IpSpaceIpAllocation, error) {
	queryParams := queryParameterFilterAnd(fmt.Sprintf("value==%s;type==%s", value, allocationType), queryParameters)

	results, err := org.GetAllIpSpaceAllocations(ipSpaceId, queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving IP allocations: %s", err)
	}

	singleResult, err := oneOrError("value", value, results)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// ipAllocationType is mandatory. FLOATING_IP or IP_PREFIX
func (ipSpace *IpSpace) GetAllIpSpaceAllocations(ipAllocationType string, queryParameters url.Values) ([]*IpSpaceIpAllocation, error) {
	client := ipSpace.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinksAllocations
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSpace.IpSpace.ID))
	if err != nil {
		return nil, err
	}
	// Type is mandatory parameter
	queryParams := queryParameterFilterAnd(fmt.Sprintf("type==%s", ipAllocationType), queryParameters)

	typeResponses := []*types.IpSpaceIpAllocation{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into DefinedEntityType types with client
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

func (org *Org) GetIpSpaceAllocationById(ipSpaceId, allocationId string) (*IpSpaceIpAllocation, error) {
	client := org.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinksAllocations
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

func (ipSpaceAllocation *IpSpaceIpAllocation) Update(ipSpaceAllocationConfig *types.IpSpaceIpAllocation) (*IpSpaceIpAllocation, error) {
	client := ipSpaceAllocation.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinksAllocations
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

func (ipSpaceAllocation *IpSpaceIpAllocation) Delete() error {
	if ipSpaceAllocation == nil || ipSpaceAllocation.IpSpaceIpAllocation == nil || ipSpaceAllocation.IpSpaceIpAllocation.ID == "" {
		return fmt.Errorf("IP Space IP Allocation must have ID")
	}

	if ipSpaceAllocation.IpSpaceId == "" || ipSpaceAllocation.parent == nil {
		return fmt.Errorf("incomplete IpSpaceIpAllocation type")
	}

	client := ipSpaceAllocation.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinksAllocations
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
		return err
	}

	if err != nil {
		return fmt.Errorf("error deleting IP Space IP Allocation: %s", err)
	}

	return nil
}

func allocateIpSpaceIp(client *Client, orgId, orgName, ipSpaceId string, ipAllocationConfig *types.IpSpaceIpAllocationRequest) ([]types.IpSpaceIpAllocationRequestResult, error) {
	if orgId == "" || orgName == "" || ipSpaceId == "" {
		return nil, fmt.Errorf("IP Space must have all values Org ID, Org Name and IP Space ID populated")
	}

	// client := ipSpace.vcdClient.Client
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
	result := task.Task.Result.ResultContent.Text

	unmarshalStorage := []types.IpSpaceIpAllocationRequestResult{}
	err = json.Unmarshal([]byte(result), &unmarshalStorage)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling task result: %s", err)
	}

	return unmarshalStorage, nil
}
