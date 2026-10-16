// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelNetworkContextProfile = "NSX-T Network Context Profile"

// NsxtNetworkContextProfile contains a structure for managing user-defined NSX-T Network Context
// Profiles. SYSTEM scoped profiles are built-in and read-only; profiles with PROVIDER and TENANT
// scopes can be managed by this structure
type NsxtNetworkContextProfile struct {
	NsxtNetworkContextProfile *types.NsxtNetworkContextProfile
	VCDClient                 *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (n NsxtNetworkContextProfile) wrap(inner *types.NsxtNetworkContextProfile) *NsxtNetworkContextProfile {
	n.NsxtNetworkContextProfile = inner
	return &n
}

// CreateNetworkContextProfile creates a user-defined Network Context Profile that can be
// referenced in Distributed Firewall rules.
//
// Profiles with TENANT scope require both OrgRef and ContextEntityID (an Org VDC or VDC Group ID)
// to be set. Profiles with PROVIDER scope are visible to all tenants and require System
// administrator privileges
func (vcdClient *VCDClient) CreateNetworkContextProfile(config *types.NsxtNetworkContextProfile) (*NsxtNetworkContextProfile, error) {
	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles,
		entityLabel: labelNetworkContextProfile,
	}
	outerType := NsxtNetworkContextProfile{VCDClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetNetworkContextProfileById retrieves a Network Context Profile of any scope by its ID
func (vcdClient *VCDClient) GetNetworkContextProfileById(id string) (*NsxtNetworkContextProfile, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles,
		endpointParams: []string{id},
		entityLabel:    labelNetworkContextProfile,
	}

	outerType := NsxtNetworkContextProfile{VCDClient: vcdClient}
	return getOuterEntity[NsxtNetworkContextProfile, types.NsxtNetworkContextProfile](&vcdClient.Client, outerType, c)
}

// Update updates a user-defined Network Context Profile. Only profiles with PROVIDER or TENANT
// scope can be updated
func (profile *NsxtNetworkContextProfile) Update(config *types.NsxtNetworkContextProfile) (*NsxtNetworkContextProfile, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles,
		endpointParams: []string{config.ID},
		entityLabel:    labelNetworkContextProfile,
	}
	outerType := NsxtNetworkContextProfile{VCDClient: profile.VCDClient}
	return updateOuterEntity(&profile.VCDClient.Client, outerType, c, config)
}

// Delete removes a user-defined Network Context Profile
func (profile *NsxtNetworkContextProfile) Delete() error {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles,
		endpointParams: []string{profile.NsxtNetworkContextProfile.ID},
		entityLabel:    labelNetworkContextProfile,
	}
	return deleteEntityById(&profile.VCDClient.Client, c)
}

// GetAllNetworkContextProfiles retrieves a slice of types.NsxtNetworkContextProfile
// This function requires at least a filter value for 'context_id' which can be one of:
// * Org VDC ID - to get Network Context Profiles scoped for VDC
// * Network provider ID - to get Network Context Profiles scoped for attached NSX-T environment
// * VDC Group ID - to get Network Context Profiles scoped for attached NSX-T environment
func GetAllNetworkContextProfiles(client *Client, queryParameters url.Values) ([]*types.NsxtNetworkContextProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtNetworkContextProfile{}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetNetworkContextProfilesByScopeAndName retrieves a single NSX-T Network Context Profile by name
// and context ID. All fields - name, scope and contextId are mandatory
//
// contextId is mandatory and can be one off:
// * Org VDC ID - to get Network Context Profiles scoped for VDC
// * Network provider ID - to get Network Context Profiles scoped for attached NSX-T environment
// * VDC Group ID - to get Network Context Profiles scoped for attached NSX-T environment
//
// scope can be one off:
// * SYSTEM
// * PROVIDER
// * TENANT
func GetNetworkContextProfilesByNameScopeAndContext(client *Client, name, scope, contextId string) (*types.NsxtNetworkContextProfile, error) {
	if name == "" || contextId == "" || scope == "" {
		return nil, fmt.Errorf("error - 'name', 'scope' and 'contextId' must be specified")
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams.Add("filter", fmt.Sprintf("name==%s", name))
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", contextId), queryParams)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("scope==%s", scope), queryParams)

	allProfiles, err := GetAllNetworkContextProfiles(client, queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Network Context Profiles by name '%s', scope '%s' and context ID '%s': %s ",
			name, scope, contextId, err)
	}

	return returnSingleNetworkContextProfile(allProfiles)
}

func returnSingleNetworkContextProfile(allProfiles []*types.NsxtNetworkContextProfile) (*types.NsxtNetworkContextProfile, error) {
	if len(allProfiles) > 1 {
		return nil, fmt.Errorf("got more than 1 NSX-T Network Context Profile %d", len(allProfiles))
	}

	if len(allProfiles) < 1 {
		return nil, fmt.Errorf("%s: got 0 NSX-T Network Context Profiles", ErrorEntityNotFound)
	}

	return allProfiles[0], nil
}
