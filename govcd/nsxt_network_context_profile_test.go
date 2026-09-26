//go:build network || nsxt || functional || openapi || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAllNetworkContextProfiles(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointNetworkContextProfiles)

	filteredTestGetAllNetworkContextProfiles(nil, &vcd.client.Client, check)

	// Test with SYSTEM scope
	queryParams := copyOrNewUrlValues(nil)
	queryParams.Add("filter", "scope==SYSTEM")
	filteredTestGetAllNetworkContextProfiles(queryParams, &vcd.client.Client, check)

	// Test with PROVIDER scope
	queryParams = copyOrNewUrlValues(nil)
	queryParams.Add("filter", "scope==PROVIDER")
	filteredTestGetAllNetworkContextProfiles(queryParams, &vcd.client.Client, check)

	// Test with TENANT scope
	queryParams = copyOrNewUrlValues(nil)
	queryParams.Add("filter", "scope==TENANT")
	filteredTestGetAllNetworkContextProfiles(queryParams, &vcd.client.Client, check)
}

func (vcd *TestVCD) Test_GetNetworkContextProfilesByNameScopeAndContext(check *C) {
	vcd.skipIfNotSysAdmin(check)
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointNetworkContextProfiles)

	// Expect error when fields are empty
	profiles, err := GetNetworkContextProfilesByNameScopeAndContext(&vcd.client.Client, "", "", "")
	check.Assert(err, NotNil)
	check.Assert(profiles, IsNil)

	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers), Equals, 1)
	uuid, err := GetUuidFromHref(nsxtManagers[0].HREF, true)
	check.Assert(err, IsNil)
	nsxtManagerUrn, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", uuid)
	check.Assert(err, IsNil)

	profiles, err = GetNetworkContextProfilesByNameScopeAndContext(&vcd.client.Client, "AMQP", "SYSTEM", nsxtManagerUrn)
	check.Assert(err, IsNil)
	check.Assert(profiles, NotNil)

	// VCD does not have PROVIDER Network Context Profiles by default
	profiles, err = GetNetworkContextProfilesByNameScopeAndContext(&vcd.client.Client, "AMQP", "PROVIDER", nsxtManagerUrn)
	check.Assert(err, NotNil)
	check.Assert(profiles, IsNil)

	// VCD does not have TENANT Network Context Profiles by default
	profiles, err = GetNetworkContextProfilesByNameScopeAndContext(&vcd.client.Client, "AMQP", "TENANT", nsxtManagerUrn)
	check.Assert(err, NotNil)
	check.Assert(profiles, IsNil)
}

func filteredTestGetAllNetworkContextProfiles(queryParams url.Values, client *Client, check *C) {
	profiles, err := GetAllNetworkContextProfiles(client, queryParams)
	check.Assert(err, IsNil)
	check.Assert(profiles, NotNil)
}

func (vcd *TestVCD) Test_NsxtNetworkContextProfileCRUD(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointNetworkContextProfiles)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	vdcGroup, err := adminOrg.GetVdcGroupByName(vcd.config.VCD.Nsxt.VdcGroup)
	check.Assert(err, IsNil)

	config := &types.NsxtNetworkContextProfile{
		Name:            check.TestName(),
		Description:     check.TestName() + "-description",
		Scope:           "TENANT",
		OrgRef:          &types.OpenApiReference{ID: adminOrg.AdminOrg.ID},
		ContextEntityID: vdcGroup.VdcGroup.Id,
		Attributes: []types.NsxtNetworkContextProfileAttributes{
			{
				Type:   "APP_ID",
				Values: []string{"HTTP", "SSL"},
			},
		},
	}

	createdProfile, err := vcd.client.CreateNetworkContextProfile(config)
	check.Assert(err, IsNil)
	check.Assert(createdProfile, NotNil)
	check.Assert(createdProfile.NsxtNetworkContextProfile.ID, Not(Equals), "")

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles + createdProfile.NsxtNetworkContextProfile.ID
	AddToCleanupListOpenApi(config.Name, check.TestName(), openApiEndpoint)

	// Get by ID and compare main fields
	retrievedProfile, err := vcd.client.GetNetworkContextProfileById(createdProfile.NsxtNetworkContextProfile.ID)
	check.Assert(err, IsNil)
	check.Assert(retrievedProfile.NsxtNetworkContextProfile.Name, Equals, config.Name)
	check.Assert(retrievedProfile.NsxtNetworkContextProfile.Scope, Equals, "TENANT")
	check.Assert(len(retrievedProfile.NsxtNetworkContextProfile.Attributes), Equals, 1)

	// Lookup with the pre-existing name based function must find the same entity
	byName, err := GetNetworkContextProfilesByNameScopeAndContext(&vcd.client.Client, config.Name, "TENANT", vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)
	check.Assert(byName.ID, Equals, createdProfile.NsxtNetworkContextProfile.ID)

	// Update description and attribute values
	updateConfig := &types.NsxtNetworkContextProfile{
		ID:              createdProfile.NsxtNetworkContextProfile.ID,
		Name:            config.Name,
		Description:     config.Description + "-updated",
		Scope:           config.Scope,
		OrgRef:          config.OrgRef,
		ContextEntityID: config.ContextEntityID,
		Attributes: []types.NsxtNetworkContextProfileAttributes{
			{
				Type:   "APP_ID",
				Values: []string{"SSL"},
				SubAttributes: []types.NsxtNetworkContextProfileSubAttribute{
					{
						Type:   "TLS_VERSION",
						Values: []string{"TLS_V12", "TLS_V13"},
					},
				},
			},
		},
	}
	updatedProfile, err := retrievedProfile.Update(updateConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedProfile.NsxtNetworkContextProfile.Description, Equals, updateConfig.Description)
	check.Assert(len(updatedProfile.NsxtNetworkContextProfile.Attributes), Equals, 1)
	check.Assert(len(updatedProfile.NsxtNetworkContextProfile.Attributes[0].Values), Equals, 1)

	// Delete and expect the entity to be gone
	err = updatedProfile.Delete()
	check.Assert(err, IsNil)

	notFoundProfile, err := vcd.client.GetNetworkContextProfileById(createdProfile.NsxtNetworkContextProfile.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundProfile, IsNil)
}
