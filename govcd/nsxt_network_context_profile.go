// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelNsxtNetworkContextProfile = "NSX-T Network Context Profile"

// NsxtNetworkContextProfile manages user-defined (TENANT or PROVIDER scoped) NSX-T Network Context
// Profiles. SYSTEM profiles are read-only.
type NsxtNetworkContextProfile struct {
	NsxtNetworkContextProfile *types.NsxtNetworkContextProfile
	vcdClient                 *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (p NsxtNetworkContextProfile) wrap(inner *types.NsxtNetworkContextProfile) *NsxtNetworkContextProfile {
	p.NsxtNetworkContextProfile = inner
	return &p
}

func nsxtNetworkContextProfileConfig(endpointParams ...string) crudConfig {
	return crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles,
		endpointParams: endpointParams,
		entityLabel:    labelNsxtNetworkContextProfile,
	}
}

// CreateNetworkContextProfile creates a user-defined NSX-T Network Context Profile. Scope,
// ContextEntityID (VDC, VDC Group or NSX-T Manager ID) and at least one attribute are required.
func (vcdClient *VCDClient) CreateNetworkContextProfile(config *types.NsxtNetworkContextProfile) (*NsxtNetworkContextProfile, error) {
	if config == nil {
		return nil, fmt.Errorf("%s configuration must not be nil", labelNsxtNetworkContextProfile)
	}
	outerType := NsxtNetworkContextProfile{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, nsxtNetworkContextProfileConfig(), config)
}

// GetNetworkContextProfileById retrieves a single NSX-T Network Context Profile by ID
func (vcdClient *VCDClient) GetNetworkContextProfileById(id string) (*NsxtNetworkContextProfile, error) {
	if id == "" {
		return nil, fmt.Errorf("%s ID must not be empty", labelNsxtNetworkContextProfile)
	}
	outerType := NsxtNetworkContextProfile{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, nsxtNetworkContextProfileConfig(id))
}

// Update updates a user-defined NSX-T Network Context Profile
func (p *NsxtNetworkContextProfile) Update(config *types.NsxtNetworkContextProfile) (*NsxtNetworkContextProfile, error) {
	if config == nil {
		return nil, fmt.Errorf("%s configuration must not be nil", labelNsxtNetworkContextProfile)
	}
	if p == nil || p.NsxtNetworkContextProfile == nil || p.NsxtNetworkContextProfile.ID == "" {
		return nil, fmt.Errorf("cannot update %s without ID", labelNsxtNetworkContextProfile)
	}
	outerType := NsxtNetworkContextProfile{vcdClient: p.vcdClient}
	return updateOuterEntity(&p.vcdClient.Client, outerType, nsxtNetworkContextProfileConfig(p.NsxtNetworkContextProfile.ID), config)
}

// Delete deletes a user-defined NSX-T Network Context Profile
func (p *NsxtNetworkContextProfile) Delete() error {
	if p == nil || p.NsxtNetworkContextProfile == nil || p.NsxtNetworkContextProfile.ID == "" {
		return fmt.Errorf("cannot delete %s without ID", labelNsxtNetworkContextProfile)
	}
	return deleteEntityById(&p.vcdClient.Client, nsxtNetworkContextProfileConfig(p.NsxtNetworkContextProfile.ID))
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
