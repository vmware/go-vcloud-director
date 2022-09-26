package govcd

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// VdcComputePolicy defines a VDC Compute Policy, which can be a VM Sizing Policy, a VM Placement Policy or a vGPU Policy.
// Deprecated: Use VdcComputePolicyV2 instead
type VdcComputePolicy struct {
	VdcComputePolicy *types.VdcComputePolicy
	Href             string
	client           *Client
}

// GetVdcComputePolicyById retrieves VDC compute policy by given ID
// Deprecated: Use VCDClient.GetVdcComputePolicyV2ById instead
func (client *Client) GetVdcComputePolicyById(id string) (*VdcComputePolicy, error) {
	return getVdcComputePolicyById(client, id)
}

// GetVdcComputePolicyById retrieves VDC compute policy by given ID
// Deprecated: use VCDClient.GetVdcComputePolicyV2ById
func (org *AdminOrg) GetVdcComputePolicyById(id string) (*VdcComputePolicy, error) {
	return getVdcComputePolicyById(org.client, id)
}

// GetVdcComputePolicyById retrieves VDC compute policy by given ID
// Deprecated: use VCDClient.GetVdcComputePolicyV2ById
func (org *Org) GetVdcComputePolicyById(id string) (*VdcComputePolicy, error) {
	return getVdcComputePolicyById(org.client, id)
}

// getVdcComputePolicyById retrieves VDC compute policy by given ID
// Deprecated: Use getVdcComputePolicyV2ById instead
func getVdcComputePolicyById(client *Client, id string) (*VdcComputePolicy, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty VDC id")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)

	if err != nil {
		return nil, err
	}

	vdcComputePolicy := &VdcComputePolicy{
		VdcComputePolicy: &types.VdcComputePolicy{},
		Href:             urlRef.String(),
		client:           client,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, vdcComputePolicy.VdcComputePolicy, nil)
	if err != nil {
		return nil, err
	}

	return vdcComputePolicy, nil
}

// GetAllVdcComputePolicies retrieves all VDC compute policies using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
// Deprecated: use VCDClient.GetAllVdcComputePoliciesV2
func (client *Client) GetAllVdcComputePolicies(queryParameters url.Values) ([]*VdcComputePolicy, error) {
	return getAllVdcComputePolicies(client, queryParameters)
}

// GetAllVdcComputePolicies retrieves all VDC compute policies using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
// Deprecated: use VCDClient.GetAllVdcComputePoliciesV2
func (org *AdminOrg) GetAllVdcComputePolicies(queryParameters url.Values) ([]*VdcComputePolicy, error) {
	return getAllVdcComputePolicies(org.client, queryParameters)
}

// GetAllVdcComputePolicies retrieves all VDC compute policies using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
// Deprecated: use VCDClient.GetAllVdcComputePoliciesV2
func (org *Org) GetAllVdcComputePolicies(queryParameters url.Values) ([]*VdcComputePolicy, error) {
	return getAllVdcComputePolicies(org.client, queryParameters)
}

// getAllVdcComputePolicies retrieves all VDC compute policies using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
// Deprecated: use getAllVdcComputePoliciesV2
func getAllVdcComputePolicies(client *Client, queryParameters url.Values) ([]*VdcComputePolicy, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	responses := []*types.VdcComputePolicy{{}}

	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses, nil)
	if err != nil {
		return nil, err
	}

	var wrappedVdcComputePolicies []*VdcComputePolicy
	for _, response := range responses {
		wrappedVdcComputePolicy := &VdcComputePolicy{
			client:           client,
			VdcComputePolicy: response,
		}
		wrappedVdcComputePolicies = append(wrappedVdcComputePolicies, wrappedVdcComputePolicy)
	}

	return wrappedVdcComputePolicies, nil
}

// CreateVdcComputePolicy creates a new VDC Compute Policy using OpenAPI endpoint
// Deprecated: use VCDClient.CreateVdcComputePolicyV2
func (org *AdminOrg) CreateVdcComputePolicy(newVdcComputePolicy *types.VdcComputePolicy) (*VdcComputePolicy, error) {
	return org.client.CreateVdcComputePolicy(newVdcComputePolicy)
}

// CreateVdcComputePolicy creates a new VDC Compute Policy using OpenAPI endpoint
// Deprecated: use VCDClient.CreateVdcComputePolicyV2
func (client *Client) CreateVdcComputePolicy(newVdcComputePolicy *types.VdcComputePolicy) (*VdcComputePolicy, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnVdcComputePolicy := &VdcComputePolicy{
		VdcComputePolicy: &types.VdcComputePolicy{},
		client:           client,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, newVdcComputePolicy, returnVdcComputePolicy.VdcComputePolicy, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating VDC compute policy: %s", err)
	}

	return returnVdcComputePolicy, nil
}

// Update existing VDC compute policy
// Deprecated: use VdcComputePolicyV2.Update
func (vdcComputePolicy *VdcComputePolicy) Update() (*VdcComputePolicy, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := vdcComputePolicy.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if vdcComputePolicy.VdcComputePolicy.ID == "" {
		return nil, fmt.Errorf("cannot update VDC compute policy without ID")
	}

	urlRef, err := vdcComputePolicy.client.OpenApiBuildEndpoint(endpoint, vdcComputePolicy.VdcComputePolicy.ID)
	if err != nil {
		return nil, err
	}

	returnVdcComputePolicy := &VdcComputePolicy{
		VdcComputePolicy: &types.VdcComputePolicy{},
		client:           vdcComputePolicy.client,
	}

	err = vdcComputePolicy.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, vdcComputePolicy.VdcComputePolicy, returnVdcComputePolicy.VdcComputePolicy, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating VDC compute policy: %s", err)
	}

	return returnVdcComputePolicy, nil
}

// Delete deletes VDC compute policy
// Deprecated: use VdcComputePolicyV2.Delete
func (vdcComputePolicy *VdcComputePolicy) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := vdcComputePolicy.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if vdcComputePolicy.VdcComputePolicy.ID == "" {
		return fmt.Errorf("cannot delete VDC compute policy without id")
	}

	urlRef, err := vdcComputePolicy.client.OpenApiBuildEndpoint(endpoint, vdcComputePolicy.VdcComputePolicy.ID)
	if err != nil {
		return err
	}

	err = vdcComputePolicy.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting VDC compute policy: %s", err)
	}

	return nil
}

// GetAllAssignedVdcComputePolicies retrieves all VDC assigned compute policies using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
// Deprecated: use AdminVdc.GetAllAssignedVdcComputePoliciesV2
func (vdc *AdminVdc) GetAllAssignedVdcComputePolicies(queryParameters url.Values) ([]*VdcComputePolicy, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcAssignedComputePolicies
	minimumApiVersion, err := vdc.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vdc.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdc.AdminVdc.ID))
	if err != nil {
		return nil, err
	}

	responses := []*types.VdcComputePolicy{{}}

	err = vdc.client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses, nil)
	if err != nil {
		return nil, err
	}

	var wrappedVdcComputePolicies []*VdcComputePolicy
	for _, response := range responses {
		wrappedVdcComputePolicy := &VdcComputePolicy{
			client:           vdc.client,
			VdcComputePolicy: response,
		}
		wrappedVdcComputePolicies = append(wrappedVdcComputePolicies, wrappedVdcComputePolicy)
	}

	return wrappedVdcComputePolicies, nil
}
