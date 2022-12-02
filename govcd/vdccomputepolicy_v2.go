package govcd

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// VdcComputePolicyV2 defines a VDC Compute Policy, which can be a VM Sizing Policy, a VM Placement Policy or a vGPU Policy.
type VdcComputePolicyV2 struct {
	VdcComputePolicyV2 *types.VdcComputePolicyV2
	Href               string
	client             *Client
}

// GetVdcComputePolicyV2ById retrieves VDC Compute Policy (V2) by given ID
func (client *VCDClient) GetVdcComputePolicyV2ById(id string) (*VdcComputePolicyV2, error) {
	return getVdcComputePolicyV2ById(client, id)
}

// getVdcComputePolicyV2ById retrieves VDC Compute Policy (V2) by given ID
func getVdcComputePolicyV2ById(client *VCDClient, id string) (*VdcComputePolicyV2, error) {
	endpoint := types.OpenApiPathVersion2_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := client.Client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty VDC id")
	}

	urlRef, err := client.Client.OpenApiBuildEndpoint(endpoint, id)

	if err != nil {
		return nil, err
	}

	vdcComputePolicy := &VdcComputePolicyV2{
		VdcComputePolicyV2: &types.VdcComputePolicyV2{},
		Href:               urlRef.String(),
		client:             &client.Client,
	}

	err = client.Client.OpenApiGetItem(minimumApiVersion, urlRef, nil, vdcComputePolicy.VdcComputePolicyV2, nil)
	if err != nil {
		return nil, err
	}

	return vdcComputePolicy, nil
}

// GetAllVdcComputePoliciesV2 retrieves all VDC Compute Policies (V2) using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
func (client *VCDClient) GetAllVdcComputePoliciesV2(queryParameters url.Values) ([]*VdcComputePolicyV2, error) {
	return getAllVdcComputePoliciesV2(client, queryParameters)
}

// getAllVdcComputePolicies retrieves all VDC Compute Policies (V2) using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
func getAllVdcComputePoliciesV2(client *VCDClient, queryParameters url.Values) ([]*VdcComputePolicyV2, error) {
	endpoint := types.OpenApiPathVersion2_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := client.Client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	responses := []*types.VdcComputePolicyV2{{}}

	err = client.Client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses, nil)
	if err != nil {
		return nil, err
	}

	var wrappedVdcComputePolicies []*VdcComputePolicyV2
	for _, response := range responses {
		wrappedVdcComputePolicy := &VdcComputePolicyV2{
			client:             &client.Client,
			VdcComputePolicyV2: response,
		}
		wrappedVdcComputePolicies = append(wrappedVdcComputePolicies, wrappedVdcComputePolicy)
	}

	return wrappedVdcComputePolicies, nil
}

// CreateVdcComputePolicyV2 creates a new VDC Compute Policy (V2) using OpenAPI endpoint
func (client *VCDClient) CreateVdcComputePolicyV2(newVdcComputePolicy *types.VdcComputePolicyV2) (*VdcComputePolicyV2, error) {
	endpoint := types.OpenApiPathVersion2_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := client.Client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnVdcComputePolicy := &VdcComputePolicyV2{
		VdcComputePolicyV2: &types.VdcComputePolicyV2{},
		client:             &client.Client,
	}

	err = client.Client.OpenApiPostItem(minimumApiVersion, urlRef, nil, newVdcComputePolicy, returnVdcComputePolicy.VdcComputePolicyV2, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating VDC Compute Policy: %s", getFriendlyErrorIfVmPlacementPolicyAlreadyExists(newVdcComputePolicy.Name, err))
	}

	return returnVdcComputePolicy, nil
}

// Update existing VDC Compute Policy (V2)
func (vdcComputePolicy *VdcComputePolicyV2) Update() (*VdcComputePolicyV2, error) {
	endpoint := types.OpenApiPathVersion2_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := vdcComputePolicy.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if vdcComputePolicy.VdcComputePolicyV2.ID == "" {
		return nil, fmt.Errorf("cannot update VDC Compute Policy without ID")
	}

	urlRef, err := vdcComputePolicy.client.OpenApiBuildEndpoint(endpoint, vdcComputePolicy.VdcComputePolicyV2.ID)
	if err != nil {
		return nil, err
	}

	returnVdcComputePolicy := &VdcComputePolicyV2{
		VdcComputePolicyV2: &types.VdcComputePolicyV2{},
		client:             vdcComputePolicy.client,
	}

	err = vdcComputePolicy.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, vdcComputePolicy.VdcComputePolicyV2, returnVdcComputePolicy.VdcComputePolicyV2, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating VDC Compute Policy: %s", err)
	}

	return returnVdcComputePolicy, nil
}

// Delete deletes VDC Compute Policy (V2)
func (vdcComputePolicy *VdcComputePolicyV2) Delete() error {
	endpoint := types.OpenApiPathVersion2_0_0 + types.OpenApiEndpointVdcComputePolicies
	minimumApiVersion, err := vdcComputePolicy.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if vdcComputePolicy.VdcComputePolicyV2.ID == "" {
		return fmt.Errorf("cannot delete VDC Compute Policy without id")
	}

	urlRef, err := vdcComputePolicy.client.OpenApiBuildEndpoint(endpoint, vdcComputePolicy.VdcComputePolicyV2.ID)
	if err != nil {
		return err
	}

	err = vdcComputePolicy.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting VDC Compute Policy: %s", err)
	}

	return nil
}

// GetAllAssignedVdcComputePoliciesV2 retrieves all VDC assigned Compute Policies (V2) using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
func (vdc *AdminVdc) GetAllAssignedVdcComputePoliciesV2(queryParameters url.Values) ([]*VdcComputePolicyV2, error) {
	return getAllAssignedVdcComputePoliciesV2(vdc.client, vdc.AdminVdc.ID, queryParameters)
}

// GetAllAssignedVdcComputePoliciesV2 retrieves all VDC assigned Compute Policies (V2) using OpenAPI endpoint and the mandatory VDC identifier.
// Query parameters can be supplied to perform additional filtering
func (vcdClient *VCDClient) GetAllAssignedVdcComputePoliciesV2(vdcId string, queryParameters url.Values) ([]*VdcComputePolicyV2, error) {
	return getAllAssignedVdcComputePoliciesV2(&vcdClient.Client, vdcId, queryParameters)
}

// getAllAssignedVdcComputePoliciesV2 retrieves all VDC assigned Compute Policies (V2) using OpenAPI endpoint and the mandatory VDC identifier.
// Query parameters can be supplied to perform additional filtering
func getAllAssignedVdcComputePoliciesV2(client *Client, vdcId string, queryParameters url.Values) ([]*VdcComputePolicyV2, error) {
	if strings.TrimSpace(vdcId) == "" {
		return nil, fmt.Errorf("VDC ID is mandatory to retrieve its assigned VDC Compute Policies")
	}

	endpoint := types.OpenApiPathVersion2_0_0 + types.OpenApiEndpointVdcAssignedComputePolicies
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcId))
	if err != nil {
		return nil, err
	}

	responses := []*types.VdcComputePolicyV2{{}}

	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses, nil)
	if err != nil {
		return nil, err
	}

	var wrappedVdcComputePolicies []*VdcComputePolicyV2
	for _, response := range responses {
		wrappedVdcComputePolicy := &VdcComputePolicyV2{
			client:             client,
			VdcComputePolicyV2: response,
		}
		wrappedVdcComputePolicies = append(wrappedVdcComputePolicies, wrappedVdcComputePolicy)
	}

	return wrappedVdcComputePolicies, nil
}

// SetAssignedComputePolicies assign(set) Compute Policies to the receiver VDC.
func (vdc *AdminVdc) SetAssignedComputePolicies(computePolicyReferences types.VdcComputePolicyReferences) (*types.VdcComputePolicyReferences, error) {
	util.Logger.Printf("[TRACE] Set Compute Policies started")

	if !vdc.client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires System Administrator privileges")
	}

	adminVdcPolicyHREF, err := url.ParseRequestURI(vdc.AdminVdc.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parsing VDC URL: %s", err)
	}

	vdcId, err := GetUuidFromHref(vdc.AdminVdc.HREF, true)
	if err != nil {
		return nil, fmt.Errorf("unable to get vdc ID from HREF: %s", err)
	}
	adminVdcPolicyHREF.Path = "/api/admin/vdc/" + vdcId + "/computePolicies"

	returnedVdcComputePolicies := &types.VdcComputePolicyReferences{}
	computePolicyReferences.Xmlns = types.XMLNamespaceVCloud

	_, err = vdc.client.ExecuteRequest(adminVdcPolicyHREF.String(), http.MethodPut,
		types.MimeVdcComputePolicyReferences, "error setting Compute Policies for VDC: %s", computePolicyReferences, returnedVdcComputePolicies)
	if err != nil {
		return nil, err
	}

	return returnedVdcComputePolicies, nil
}

// getFriendlyErrorIfVmPlacementPolicyAlreadyExists is intended to be used when a VM Placement Policy already exists, and
// we try to create another one with the same name. When this happens, VCD discloses a lot of unnecessary information to the user that is
// hard to read and understand, so this function simplifies the message.
// Note: This function should not be needed anymore once VCD 10.4.0 is discontinued (this issue is fixed in 10.4.1).
func getFriendlyErrorIfVmPlacementPolicyAlreadyExists(vmPlacementPolicyName string, err error) error {
	if err != nil && strings.Contains(err.Error(), "already exists") && strings.Contains(err.Error(), "duplicate key") {
		return fmt.Errorf("VM Placement Policy with name '%s' already exists", vmPlacementPolicyName)
	}
	return err
}
