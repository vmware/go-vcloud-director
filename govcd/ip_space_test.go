//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_IpSpacePublic(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointIpSpaces)

	ipSpaceConfig := &types.IpSpace{
		Name:                      check.TestName(),
		IPSpaceInternalScope:      []string{"22.0.0.0/24"},
		IPSpaceExternalScope:      "200.0.0.1/24",
		Type:                      types.IpSpacePublic,
		RouteAdvertisementEnabled: false,
		IPSpacePrefixes: []types.IPSpacePrefixes{
			{
				DefaultQuotaForPrefixLength: -1,
				IPPrefixSequence: []types.IPPrefixSequence{
					{
						StartingPrefixIPAddress: "22.0.0.200",
						PrefixLength:            31,
						TotalPrefixCount:        3,
					},
				},
			},
			{
				DefaultQuotaForPrefixLength: 2,
				IPPrefixSequence: []types.IPPrefixSequence{
					{
						StartingPrefixIPAddress: "22.0.0.100",
						PrefixLength:            30,
						TotalPrefixCount:        3,
					},
				},
			},
		},
		IPSpaceRanges: types.IPSpaceRanges{
			DefaultFloatingIPQuota: 3,
			IPRanges: []types.IpSpaceRangeValues{
				{
					StartIPAddress: "22.0.0.10",
					EndIPAddress:   "22.0.0.30",
				},
				{
					StartIPAddress: "22.0.0.32",
					EndIPAddress:   "22.0.0.34",
				},
			},
		},
	}

	ipSpaceChecks(vcd, check, ipSpaceConfig)
}

func (vcd *TestVCD) Test_IpSpaceShared(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointIpSpaces)

	ipSpaceConfig := &types.IpSpace{
		Name:                      check.TestName(),
		IPSpaceInternalScope:      []string{"22.0.0.1/24"},
		IPSpaceExternalScope:      "200.0.0.1/24",
		Type:                      types.IpSpaceShared,
		RouteAdvertisementEnabled: false,
		IPSpacePrefixes: []types.IPSpacePrefixes{
			{
				DefaultQuotaForPrefixLength: -1,
				IPPrefixSequence: []types.IPPrefixSequence{
					{
						StartingPrefixIPAddress: "22.0.0.200",
						PrefixLength:            31,
						TotalPrefixCount:        3,
					},
				},
			},
			{
				DefaultQuotaForPrefixLength: 2,
				IPPrefixSequence: []types.IPPrefixSequence{
					{
						StartingPrefixIPAddress: "22.0.0.100",
						PrefixLength:            30,
						TotalPrefixCount:        3,
					},
				},
			},
		},
		IPSpaceRanges: types.IPSpaceRanges{
			DefaultFloatingIPQuota: 3,
			IPRanges: []types.IpSpaceRangeValues{
				{
					StartIPAddress: "22.0.0.10",
					EndIPAddress:   "22.0.0.30",
				},
				{
					StartIPAddress: "22.0.0.32",
					EndIPAddress:   "22.0.0.34",
				},
			},
		},
	}
	ipSpaceChecks(vcd, check, ipSpaceConfig)
}

func (vcd *TestVCD) Test_IpSpacePrivate(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointIpSpaces)

	ipSpaceConfig := &types.IpSpace{
		OrgRef: &types.OpenApiReference{
			ID: vcd.org.Org.ID, // Private IP Space requires Org
		},
		Name:                      check.TestName(),
		IPSpaceInternalScope:      []string{"22.0.0.1/24"},
		IPSpaceExternalScope:      "200.0.0.1/24",
		Type:                      types.IpSpacePrivate,
		RouteAdvertisementEnabled: false,
		IPSpacePrefixes: []types.IPSpacePrefixes{
			{
				DefaultQuotaForPrefixLength: -1,
				IPPrefixSequence: []types.IPPrefixSequence{
					{
						StartingPrefixIPAddress: "22.0.0.200",
						PrefixLength:            31,
						TotalPrefixCount:        3,
					},
				},
			},
			{
				DefaultQuotaForPrefixLength: 2,
				IPPrefixSequence: []types.IPPrefixSequence{
					{
						StartingPrefixIPAddress: "22.0.0.100",
						PrefixLength:            30,
						TotalPrefixCount:        3,
					},
				},
			},
		},
		IPSpaceRanges: types.IPSpaceRanges{
			DefaultFloatingIPQuota: 3,
			IPRanges: []types.IpSpaceRangeValues{
				{
					StartIPAddress: "22.0.0.10",
					EndIPAddress:   "22.0.0.30",
				},
				{
					StartIPAddress: "22.0.0.32",
					EndIPAddress:   "22.0.0.34",
				},
			},
		},
	}

	ipSpaceChecks(vcd, check, ipSpaceConfig)
}

func ipSpaceChecks(vcd *TestVCD, check *C, ipSpaceConfig *types.IpSpace) {
	createdIpSpace, err := vcd.client.CreateIpSpace(ipSpaceConfig)
	check.Assert(err, IsNil)
	check.Assert(createdIpSpace, NotNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces + createdIpSpace.IpSpace.ID
	AddToCleanupListOpenApi(createdIpSpace.IpSpace.Name, check.TestName(), openApiEndpoint)

	// Get by ID
	byId, err := vcd.client.GetIpSpaceById(createdIpSpace.IpSpace.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.IpSpace, DeepEquals, createdIpSpace.IpSpace)

	// Get by Name
	byName, err := vcd.client.GetIpSpaceByName(createdIpSpace.IpSpace.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.IpSpace, DeepEquals, createdIpSpace.IpSpace)

	// Get all and make sure it is found
	allIpSpaces, err := vcd.client.GetAllIpSpaceSummaries(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allIpSpaces) > 0, Equals, true)
	var found bool
	for i := range allIpSpaces {
		if allIpSpaces[i].IpSpace.ID == byId.IpSpace.ID {
			found = true
			break
		}
	}
	check.Assert(found, Equals, true)

	// If an Org is assigned - attempt to lookup by name and Org ID
	if byId.IpSpace.OrgRef != nil && byId.IpSpace.OrgRef.ID != "" {
		byNameAndOrgId, err := vcd.client.GetIpSpaceByNameAndOrgId(byId.IpSpace.Name, byId.IpSpace.OrgRef.ID)
		check.Assert(err, IsNil)
		check.Assert(byNameAndOrgId, NotNil)
		check.Assert(byNameAndOrgId.IpSpace, DeepEquals, createdIpSpace.IpSpace)

	}

	// Check an update
	ipSpaceConfig.RouteAdvertisementEnabled = true
	ipSpaceConfig.IPSpaceInternalScope = append(ipSpaceConfig.IPSpaceInternalScope, "32.0.0.0/24")

	updatedIpSpace, err := createdIpSpace.Update(ipSpaceConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedIpSpace, NotNil)
	check.Assert(len(ipSpaceConfig.IPSpaceInternalScope), Equals, len(updatedIpSpace.IpSpace.IPSpaceInternalScope))

	if vcd.client.Client.APIVCDMaxVersionIs(">= 38.0") {
		fmt.Println("# Testing NAT and Firewall rule autocreation flags for VCD 10.5.0+")
		ipSpaceConfig.Name = check.TestName() + "-GatewayServiceConfig"
		ipSpaceConfig.DefaultGatewayServiceConfig = &types.IpSpaceDefaultGatewayServiceConfig{
			EnableDefaultFirewallRuleCreation: true,
			EnableDefaultNoSnatRuleCreation:   true,
			EnableDefaultSnatRuleCreation:     true,
		}

		updatedIpSpace, err = updatedIpSpace.Update(ipSpaceConfig)
		check.Assert(err, IsNil)
		check.Assert(updatedIpSpace.IpSpace.DefaultGatewayServiceConfig, DeepEquals, ipSpaceConfig.DefaultGatewayServiceConfig)
	}

	err = createdIpSpace.Delete()
	check.Assert(err, IsNil)

	// Check that the entity is not found
	notFoundById, err := vcd.client.GetIpSpaceById(byId.IpSpace.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundById, IsNil)
}
