/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

// VdcGroup is a structure defining a VdcGroup in Organization
type VdcGroup struct {
	VdcGroup *types.VdcGroup
	Href     string
	client   *Client
	parent   organization
}

// CreateNsxtVdcGroup create NSX-T VDC group with provided VDC IDs.
func (adminOrg *AdminOrg) CreateNsxtVdcGroup(name, description, startingVdcId string, participatingVdcIds []string) (*VdcGroup, error) {
	candidateVdcs, err := adminOrg.GetAllNsxtCandidateVdcs(startingVdcId, nil)
	if err != nil {
		return nil, err
	}
	participatingVdcs := []types.ParticipatingOrgVdcs{}
	vdcGroupConfig := &types.VdcGroup{}
	for _, candidateVdc := range *candidateVdcs {
		if containsInString(candidateVdc.Id, participatingVdcIds) {
			participatingVdcs = append(participatingVdcs, types.ParticipatingOrgVdcs{
				OrgRef:  candidateVdc.OrgRef,
				SiteRef: candidateVdc.SiteRef,
				VdcRef: types.OpenApiReference{
					ID: candidateVdc.Id,
				},
				FaultDomainTag:       candidateVdc.FaultDomainTag,
				NetworkProviderScope: candidateVdc.NetworkProviderScope,
			})
		}
	}
	vdcGroupConfig.OrgId = adminOrg.orgId()
	vdcGroupConfig.Name = name
	vdcGroupConfig.Description = description
	vdcGroupConfig.ParticipatingOrgVdcs = participatingVdcs
	vdcGroupConfig.LocalEgress = false
	vdcGroupConfig.UniversalNetworkingEnabled = false
	vdcGroupConfig.NetworkProviderType = "NSX_T"
	vdcGroupConfig.Type = "LOCAL"
	return adminOrg.CreateVdcGroup(vdcGroupConfig)
}

// containsInString tells whether slice of string contains item.
func containsInString(item string, slice []string) bool {
	for _, n := range slice {
		if item == n {
			return true
		}
	}
	return false
}

// CreateVdcGroup create VDC group with provided VDC ref.
// Only support NSX-T VDCs.
func (adminOrg *AdminOrg) CreateVdcGroup(vdcGroup *types.VdcGroup) (*VdcGroup, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	return createVdcGroup(adminOrg.client, vdcGroup, getTenantContextHeader(tenantContext))
}

// CreateVdcGroup create VDC group with provided VDC ref.
// Only support NSX-T VDCs.
func createVdcGroup(client *Client, vdcGroup *types.VdcGroup,
	additionalHeader map[string]string) (*VdcGroup, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponse := &VdcGroup{
		VdcGroup: &types.VdcGroup{},
		client:   client,
		Href:     urlRef.String(),
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil,
		vdcGroup, typeResponse.VdcGroup, additionalHeader)
	if err != nil {
		return nil, err
	}

	return typeResponse, nil
}

func (adminOrg *AdminOrg) GetAllNsxtCandidateVdcs(startingVdcId string, queryParameters url.Values) (*[]types.CandidateVdc, error) {
	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("_context==LOCAL", queryParams)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", startingVdcId), queryParams)
	queryParams.Add("filterEncoded", "true")
	queryParams.Add("links", "true")
	return adminOrg.GetAllCandidateVdcs(queryParams)
}

func (adminOrg *AdminOrg) GetAllCandidateVdcs(queryParameters url.Values) (*[]types.CandidateVdc, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsCandidateVdcs
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	responses := &[]types.CandidateVdc{}
	err = adminOrg.client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	return responses, nil
}

// Delete deletes VDC group
func (vdcGroup *VdcGroup) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups
	minimumApiVersion, err := vdcGroup.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if vdcGroup.VdcGroup.Id == "" {
		return fmt.Errorf("cannot delete VDC group without id")
	}

	urlRef, err := vdcGroup.client.OpenApiBuildEndpoint(endpoint, vdcGroup.VdcGroup.Id)
	if err != nil {
		return err
	}

	err = vdcGroup.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting VDC group: %s", err)
	}

	return nil
}

// GetAllVdcGroups retrieves all VDC groups. Query parameters can be supplied to perform additional filtering
func (adminOrg *AdminOrg) GetAllVdcGroups(queryParameters url.Values) ([]*VdcGroup, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	responses := []*types.VdcGroup{}
	err = adminOrg.client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	var wrappedVdcGroups []*VdcGroup
	for _, response := range responses {
		urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint, response.Id)
		if err != nil {
			return nil, err
		}
		wrappedVdcGroup := &VdcGroup{
			VdcGroup: response,
			client:   adminOrg.client,
			Href:     urlRef.String(),
			parent:   adminOrg,
		}
		wrappedVdcGroups = append(wrappedVdcGroups, wrappedVdcGroup)
	}

	return wrappedVdcGroups, nil
}

// GetVdcGroupByName retrieves VDC group by given name
// When the name contains commas, semicolons or asterisks, the encoding is rejected by the API in VCD 10.2 version.
// For this reason, when one or more commas, semicolons or asterisks are present we run the search brute force,
// by fetching all VDC groups and comparing the names. Yet, this not needed anymore in VCD 10.3 version.
// Also, url.QueryEscape as well as url.Values.Encode() both encode the space as a + character. So we use
// search brute force too. Reference to issue:
// https://github.com/golang/go/issues/4013
// https://github.com/czos/goamz/pull/11/files
func (adminOrg *AdminOrg) GetVdcGroupByName(name string) (*VdcGroup, error) {
	slowSearch, params, err := isShouldDoSlowSearch(name, adminOrg.client)
	if err != nil {
		return nil, err
	}

	vdcGroups, err := adminOrg.GetAllVdcGroups(params)
	if err != nil {
		return nil, err
	}
	if len(vdcGroups) == 0 {
		return nil, ErrorEntityNotFound
	}

	if slowSearch {
		for _, vdcGroup := range vdcGroups {
			if vdcGroup.VdcGroup.Name == name {
				return vdcGroup, nil
			}
		}
		return nil, ErrorEntityNotFound
	}

	if len(vdcGroups) > 1 {
		return nil, fmt.Errorf("more than one VDC group found with name '%s'", name)
	}
	return vdcGroups[0], nil
}

// GetVdcGroupById Returns VDC group using provided ID
func (adminOrg *AdminOrg) GetVdcGroupById(id string) (*VdcGroup, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty VDC group ID")
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	vdcGroup := &VdcGroup{
		VdcGroup: &types.VdcGroup{},
		client:   adminOrg.client,
		Href:     urlRef.String(),
		parent:   adminOrg,
	}

	err = adminOrg.client.OpenApiGetItem(minimumApiVersion, urlRef, nil, vdcGroup.VdcGroup, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	return vdcGroup, nil
}

// Update updates existing Vdc group. Allows changing only name and description
func (vdcGroup *VdcGroup) Update() (*VdcGroup, error) {
	tenantContext, err := vdcGroup.getTenantContext()
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups
	minimumApiVersion, err := vdcGroup.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if vdcGroup.VdcGroup.Id == "" {
		return nil, fmt.Errorf("cannot update VDC group without id")
	}

	urlRef, err := vdcGroup.client.OpenApiBuildEndpoint(endpoint, vdcGroup.VdcGroup.Id)
	if err != nil {
		return nil, err
	}

	returnVdcGroup := &VdcGroup{
		VdcGroup: &types.VdcGroup{},
		client:   vdcGroup.client,
	}

	err = vdcGroup.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, vdcGroup.VdcGroup,
		returnVdcGroup.VdcGroup, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, fmt.Errorf("error updating VDC group: %s", err)
	}

	return returnVdcGroup, nil
}
