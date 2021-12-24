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
// More generic creation method available also - CreateVdcGroup
func (adminOrg *AdminOrg) CreateNsxtVdcGroup(name, description, startingVdcId string, participatingVdcIds []string) (*VdcGroup, error) {
	participatingVdcs, err := composeParticipatingOrgVdcs(adminOrg, startingVdcId, participatingVdcIds)
	if err != nil {
		return nil, err
	}

	vdcGroupConfig := &types.VdcGroup{}
	vdcGroupConfig.OrgId = adminOrg.orgId()
	vdcGroupConfig.Name = name
	vdcGroupConfig.Description = description
	vdcGroupConfig.ParticipatingOrgVdcs = participatingVdcs
	vdcGroupConfig.LocalEgress = false
	vdcGroupConfig.UniversalNetworkingEnabled = false
	vdcGroupConfig.NetworkProviderType = "NSX_T"
	vdcGroupConfig.Type = "LOCAL"
	vdcGroupConfig.ParticipatingOrgVdcs = participatingVdcs
	return adminOrg.CreateVdcGroup(vdcGroupConfig)
}

// composeParticipatingOrgVdcs converts fetched candidate VDCs to []types.ParticipatingOrgVdcs
// returns error also in case participatingVdcId not found as candidate VDC.
func composeParticipatingOrgVdcs(adminOrg *AdminOrg, startingVdcId string, participatingVdcIds []string) ([]types.ParticipatingOrgVdcs, error) {
	candidateVdcs, err := adminOrg.GetAllNsxtVdcGroupCandidates(startingVdcId, nil)
	if err != nil {
		return nil, err
	}
	participatingVdcs := []types.ParticipatingOrgVdcs{}
	var foundParticipatingVdcsIds []string
	for _, candidateVdc := range candidateVdcs {
		if contains(candidateVdc.Id, participatingVdcIds) {
			participatingVdcs = append(participatingVdcs, types.ParticipatingOrgVdcs{
				OrgRef:  candidateVdc.OrgRef,
				SiteRef: candidateVdc.SiteRef,
				VdcRef: types.OpenApiReference{
					ID: candidateVdc.Id,
				},
				FaultDomainTag:       candidateVdc.FaultDomainTag,
				NetworkProviderScope: candidateVdc.NetworkProviderScope,
			})
			foundParticipatingVdcsIds = append(foundParticipatingVdcsIds, candidateVdc.Id)
		}
	}

	if len(participatingVdcs) != len(participatingVdcIds) {
		var notFoundVdcs []string
		for _, participatingVdcId := range participatingVdcIds {
			if !contains(participatingVdcId, foundParticipatingVdcsIds) {
				notFoundVdcs = append(notFoundVdcs, participatingVdcId)
			}
		}
		return nil, fmt.Errorf("VDC IDs are not found as Candidate VDCs: %s", notFoundVdcs)
	}

	return participatingVdcs, nil
}

// contains tells whether slice of string contains item.
func contains(item string, slice []string) bool {
	for _, n := range slice {
		if item == n {
			return true
		}
	}
	return false
}

// CreateVdcGroup create VDC group with provided VDC ref.
// Only supports NSX-T VDCs.
func (adminOrg *AdminOrg) CreateVdcGroup(vdcGroup *types.VdcGroup) (*VdcGroup, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	return createVdcGroup(adminOrg, vdcGroup, getTenantContextHeader(tenantContext))
}

// createVdcGroup create VDC group with provided VDC ref.
// Only supports NSX-T VDCs.
func createVdcGroup(adminOrg *AdminOrg, vdcGroup *types.VdcGroup,
	additionalHeader map[string]string) (*VdcGroup, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups
	apiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponse := &VdcGroup{
		VdcGroup: &types.VdcGroup{},
		client:   adminOrg.client,
		Href:     urlRef.String(),
		parent:   adminOrg,
	}

	err = adminOrg.client.OpenApiPostItem(apiVersion, urlRef, nil,
		vdcGroup, typeResponse.VdcGroup, additionalHeader)
	if err != nil {
		return nil, err
	}

	return typeResponse, nil
}

// GetAllNsxtVdcGroupCandidates returns NSXT candidate VDCs for VDC group
func (adminOrg *AdminOrg) GetAllNsxtVdcGroupCandidates(startingVdcId string, queryParameters url.Values) ([]*types.CandidateVdc, error) {
	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("_context==LOCAL", queryParams)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", startingVdcId), queryParams)
	queryParams.Add("filterEncoded", "true")
	queryParams.Add("links", "true")
	return adminOrg.GetAllVdcGroupCandidates(queryParams)
}

// GetAllVdcGroupCandidates returns candidate VDCs for VDC group
func (adminOrg *AdminOrg) GetAllVdcGroupCandidates(queryParameters url.Values) ([]*types.CandidateVdc, error) {
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

	responses := []*types.CandidateVdc{}
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
	slowSearch, params, err := shouldDoSlowSearch("name", name, adminOrg.client)
	if err != nil {
		return nil, err
	}

	var foundVdcGroups []*VdcGroup
	vdcGroups, err := adminOrg.GetAllVdcGroups(params)
	if err != nil {
		return nil, err
	}
	if len(vdcGroups) == 0 {
		return nil, ErrorEntityNotFound
	}
	foundVdcGroups = append(foundVdcGroups, vdcGroups[0])

	if slowSearch {
		foundVdcGroups = nil
		for _, vdcGroup := range vdcGroups {
			if vdcGroup.VdcGroup.Name == name {
				foundVdcGroups = append(foundVdcGroups, vdcGroup)
			}
		}
		if len(foundVdcGroups) == 0 {
			return nil, ErrorEntityNotFound
		}
		if len(foundVdcGroups) > 1 {
			return nil, fmt.Errorf("more than one VDC group found with name '%s'", name)
		}
	}

	if len(vdcGroups) > 1 && !slowSearch {
		return nil, fmt.Errorf("more than one VDC group found with name '%s'", name)
	}

	return foundVdcGroups[0], nil
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

// Update updates existing Vdc group. Allows changing only name and description and participating VCDs
// Not restrictive update method also available - GenericUpdate
func (vdcGroup *VdcGroup) Update(name, description string, participatingOrgVddIs []string) (*VdcGroup, error) {

	vdcGroup.VdcGroup.Name = name
	vdcGroup.VdcGroup.Description = description

	participatingOrgVdcs, err := composeParticipatingOrgVdcs(vdcGroup.parent.fullObject().(*AdminOrg), vdcGroup.VdcGroup.Id, participatingOrgVddIs)
	if err != nil {
		return nil, err
	}
	vdcGroup.VdcGroup.ParticipatingOrgVdcs = participatingOrgVdcs

	return vdcGroup.GenericUpdate()
}

// GenericUpdate updates existing Vdc group. API allows changing only name and description and participating VCDs
func (vdcGroup *VdcGroup) GenericUpdate() (*VdcGroup, error) {
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
		Href:     vdcGroup.Href,
		parent:   vdcGroup.parent,
	}

	err = vdcGroup.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, vdcGroup.VdcGroup,
		returnVdcGroup.VdcGroup, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, fmt.Errorf("error updating VDC group: %s", err)
	}

	return returnVdcGroup, nil
}

// UpdateDfwPolicies updates distributed firewall policies
func (vdcGroup *VdcGroup) UpdateDfwPolicies(dfwPolicies types.DfwPolicies) (*VdcGroup, error) {
	tenantContext, err := vdcGroup.getTenantContext()
	if err != nil {
		return nil, err
	}
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwPolicies
	minimumApiVersion, err := vdcGroup.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if vdcGroup.VdcGroup.Id == "" {
		return nil, fmt.Errorf("cannot update VDC group Dfw policies without id")
	}

	urlRef, err := vdcGroup.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id))
	if err != nil {
		return nil, err
	}

	err = vdcGroup.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, dfwPolicies,
		nil, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, fmt.Errorf("error updating VDC group Dfw policies: %s", err)
	}

	adminOrg := vdcGroup.parent.fullObject().(*AdminOrg)
	return adminOrg.GetVdcGroupById(vdcGroup.VdcGroup.Id)
}

// UpdateDefaultDfwPolicies updates distributed firewall default policies
func (vdcGroup *VdcGroup) UpdateDefaultDfwPolicies(defaultDfwPolicies types.DefaultPolicy) (*VdcGroup, error) {
	tenantContext, err := vdcGroup.getTenantContext()
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwDefaultPolicies
	minimumApiVersion, err := vdcGroup.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if vdcGroup.VdcGroup.Id == "" {
		return nil, fmt.Errorf("cannot update VDC group default DFW policies without id")
	}

	urlRef, err := vdcGroup.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id))
	if err != nil {
		return nil, err
	}

	err = vdcGroup.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, defaultDfwPolicies,
		nil, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, fmt.Errorf("error updating VDC group default DFW policies: %s", err)
	}

	adminOrg := vdcGroup.parent.fullObject().(*AdminOrg)
	return adminOrg.GetVdcGroupById(vdcGroup.VdcGroup.Id)
}

// ActivateDfw activates distributed firewall
func (vdcGroup *VdcGroup) ActivateDfw() (*VdcGroup, error) {
	return vdcGroup.UpdateDfwPolicies(types.DfwPolicies{
		Enabled: true,
	})
}

// DeactivateDfw deactivates distributed firewall
func (vdcGroup *VdcGroup) DeactivateDfw() (*VdcGroup, error) {
	return vdcGroup.UpdateDfwPolicies(types.DfwPolicies{
		Enabled: false,
	})
}

// GetDfwPolicies retrieves all distributed firewall policies
func (vdcGroup *VdcGroup) GetDfwPolicies() (*types.DfwPolicies, error) {
	tenantContext, err := vdcGroup.getTenantContext()
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwPolicies
	minimumApiVersion, err := vdcGroup.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vdcGroup.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id))
	if err != nil {
		return nil, err
	}

	response := types.DfwPolicies{}
	err = vdcGroup.client.OpenApiGetItem(minimumApiVersion, urlRef, nil, &response, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// EnableDefaultPolicy activates default dfw policy
func (vdcGroup *VdcGroup) EnableDefaultPolicy() (*VdcGroup, error) {
	dfwPolicies, err := vdcGroup.GetDfwPolicies()
	if err != nil {
		return nil, err
	}

	if dfwPolicies.DefaultPolicy == nil {
		return nil, fmt.Errorf("DFW has to be enabled before changing  Default policy")
	}
	dfwPolicies.DefaultPolicy.Enabled = takeBoolPointer(true)
	return vdcGroup.UpdateDefaultDfwPolicies(*dfwPolicies.DefaultPolicy)
}

// DisableDefaultPolicy deactivates default dfw policy
func (vdcGroup *VdcGroup) DisableDefaultPolicy() (*VdcGroup, error) {
	dfwPolicies, err := vdcGroup.GetDfwPolicies()
	if err != nil {
		return nil, err
	}

	if dfwPolicies.DefaultPolicy == nil {
		return nil, fmt.Errorf("DFW has to be enabled before changing Default policy")
	}
	dfwPolicies.DefaultPolicy.Enabled = takeBoolPointer(false)
	return vdcGroup.UpdateDefaultDfwPolicies(*dfwPolicies.DefaultPolicy)
}
