/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// IpSpaceOrgAssignment handles Custom Quotas (name in UI) for a particular Org. They complement
// default quotas which are being set in IP Space itself.
// The behavior of IpSpaceOrgAssignment is specific - whenever an NSX-T Edge Gateway backed by
// Provider gateway using IP Spaces is being created - Org Assignment is created implicitly. One can
// look up that assignment by IP Space and Org to update `types.IpSpaceOrgAssignment.CustomQuotas`
// field
type IpSpaceOrgAssignment struct {
	IpSpaceOrgAssignment *types.IpSpaceOrgAssignment
	IpSpaceId            string

	vcdClient *VCDClient
}

// GetAllOrgAssignments retrieves all IP Space Org assignments within an IP Space
//
// Note. Org assignments are implicitly created after NSX-T Edge Gateway backed by Provider gateway
// using IP Spaces is being created.
func (ipSpace *IpSpace) GetAllOrgAssignments(queryParameters url.Values) ([]*IpSpaceOrgAssignment, error) {
	client := ipSpace.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceOrgAssignments
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	queryParams := queryParameterFilterAnd(fmt.Sprintf("ipSpaceRef.id==%s", ipSpace.IpSpace.ID), queryParameters)
	typeResponses := []*types.IpSpaceOrgAssignment{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into IpSpaceOrgAssignment types with client
	results := make([]*IpSpaceOrgAssignment, len(typeResponses))
	for sliceIndex := range typeResponses {
		results[sliceIndex] = &IpSpaceOrgAssignment{
			IpSpaceOrgAssignment: typeResponses[sliceIndex],
			IpSpaceId:            ipSpace.IpSpace.ID,
			vcdClient:            ipSpace.vcdClient,
		}
	}

	return results, nil
}

// GetOrgAssignmentById retrieves IP Space Org Assignment with a given ID
func (ipSpace *IpSpace) GetOrgAssignmentById(id string) (*IpSpaceOrgAssignment, error) {
	if id == "" {
		return nil, fmt.Errorf("IP Space Org Assignment lookup requires ID")
	}

	client := ipSpace.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceOrgAssignments
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	response := &IpSpaceOrgAssignment{
		IpSpaceOrgAssignment: &types.IpSpaceOrgAssignment{},
		IpSpaceId:            ipSpace.IpSpace.ID,
		vcdClient:            ipSpace.vcdClient,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, response.IpSpaceOrgAssignment, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetOrgAssignmentById retrieves IP Space Org Assignment with a given Org Name
func (ipSpace *IpSpace) GetOrgAssignmentByOrgName(orgName string) (*IpSpaceOrgAssignment, error) {
	if orgName == "" {
		return nil, fmt.Errorf("name of Org is required")
	}
	queryParams := queryParameterFilterAnd(fmt.Sprintf("orgRef.name==%s", orgName), nil)
	results, err := ipSpace.GetAllOrgAssignments(queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving IP Space Org Assignments by Org Name: %s", err)
	}

	singleResult, err := oneOrError("Org Name", orgName, results)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetOrgAssignmentById retrieves IP Space Org Assignment with a given Org ID
func (ipSpace *IpSpace) GetOrgAssignmentByOrgId(orgId string) (*IpSpaceOrgAssignment, error) {
	if orgId == "" {
		return nil, fmt.Errorf("organization ID is required")
	}
	queryParams := queryParameterFilterAnd(fmt.Sprintf("orgRef.id==%s", orgId), nil)
	results, err := ipSpace.GetAllOrgAssignments(queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving IP Space Org Assignments by Org ID: %s", err)
	}

	singleResult, err := oneOrError("Org ID", orgId, results)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// Update Org Assignment
func (ipSpaceOrgAssignment *IpSpaceOrgAssignment) Update(ipSpaceOrgAssignmentConfig *types.IpSpaceOrgAssignment) (*IpSpaceOrgAssignment, error) {
	client := ipSpaceOrgAssignment.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceOrgAssignments
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	ipSpaceOrgAssignmentConfig.ID = ipSpaceOrgAssignment.IpSpaceOrgAssignment.ID
	urlRef, err := client.OpenApiBuildEndpoint(endpoint, ipSpaceOrgAssignmentConfig.ID)
	if err != nil {
		return nil, err
	}

	result := &IpSpaceOrgAssignment{
		IpSpaceOrgAssignment: &types.IpSpaceOrgAssignment{},
		IpSpaceId:            ipSpaceOrgAssignment.IpSpaceId,
		vcdClient:            ipSpaceOrgAssignment.vcdClient,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, ipSpaceOrgAssignmentConfig, result.IpSpaceOrgAssignment, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating IP Space Org Assignment: %s", err)
	}

	return result, nil
}
