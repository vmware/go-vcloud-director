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
	vcd.skipIfNotSysAdmin(check)
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointNetworkContextProfiles)

	vdc, err := vcd.org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	// Validation errors
	_, err = vcd.client.CreateNetworkContextProfile(nil)
	check.Assert(err, NotNil)
	_, err = vcd.client.GetNetworkContextProfileById("")
	check.Assert(err, NotNil)

	config := &types.NsxtNetworkContextProfile{
		Name:            check.TestName(),
		Description:     "created by go-vcloud-director tests",
		Scope:           "TENANT",
		ContextEntityID: vdc.Vdc.ID,
		Attributes: []types.NsxtNetworkContextProfileAttributes{
			{Type: "APP_ID", Values: []string{"SSL"}},
		},
	}
	profile, err := vcd.client.CreateNetworkContextProfile(config)
	check.Assert(err, IsNil)
	check.Assert(profile.NsxtNetworkContextProfile.ID, Not(Equals), "")
	PrependToCleanupListOpenApi(profile.NsxtNetworkContextProfile.ID, check.TestName(),
		types.OpenApiPathVersion1_0_0+types.OpenApiEndpointNetworkContextProfiles+profile.NsxtNetworkContextProfile.ID)

	// Get by ID and by name through the existing lookup function
	byId, err := vcd.client.GetNetworkContextProfileById(profile.NsxtNetworkContextProfile.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.NsxtNetworkContextProfile.Name, Equals, config.Name)
	check.Assert(byId.NsxtNetworkContextProfile.Attributes[0].Type, Equals, "APP_ID")

	byName, err := GetNetworkContextProfilesByNameScopeAndContext(&vcd.client.Client, config.Name, "TENANT", vdc.Vdc.ID)
	check.Assert(err, IsNil)
	check.Assert(byName.ID, Equals, profile.NsxtNetworkContextProfile.ID)

	// Update
	update := *byId.NsxtNetworkContextProfile
	update.Description = "updated"
	update.Attributes = []types.NsxtNetworkContextProfileAttributes{
		{
			Type:   "APP_ID",
			Values: []string{"SSL"},
			SubAttributes: []types.NsxtNetworkContextProfileSubAttributes{
				{Type: "TLS_VERSION", Values: []string{"TLS_V12"}},
			},
		},
	}
	updated, err := byId.Update(&update)
	check.Assert(err, IsNil)
	check.Assert(updated.NsxtNetworkContextProfile.Description, Equals, "updated")
	check.Assert(len(updated.NsxtNetworkContextProfile.Attributes[0].SubAttributes), Equals, 1)

	// Delete
	err = updated.Delete()
	check.Assert(err, IsNil)
	_, err = vcd.client.GetNetworkContextProfileById(profile.NsxtNetworkContextProfile.ID)
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
}
