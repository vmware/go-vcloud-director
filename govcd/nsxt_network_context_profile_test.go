//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
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
