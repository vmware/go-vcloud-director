//go:build network || nsxt || functional || openapi || ALL
// +build network nsxt functional openapi ALL

package govcd

import (
	"fmt"
	"net/netip"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtEdgeCreate(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	nsxvVdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(nsxvVdc, NotNil)
	nsxtVdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	if ContainsNotFound(err) {
		check.Skip(fmt.Sprintf("No NSX-T VDC (%s) found - skipping test", vcd.config.VCD.Nsxt.Vdc))
	}
	check.Assert(err, IsNil)
	check.Assert(nsxtVdc, NotNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(nsxtExternalNetwork, NotNil)

	egwDefinition := &types.OpenAPIEdgeGateway{
		Name:        "nsx-t-edge",
		Description: "nsx-t-edge-description",
		OrgVdc: &types.OpenApiReference{
			ID: nsxtVdc.Vdc.ID,
		},
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{{
			UplinkID: nsxtExternalNetwork.ExternalNetwork.ID,
			Subnets: types.OpenAPIEdgeGatewaySubnets{Values: []types.OpenAPIEdgeGatewaySubnetValue{{
				Gateway:      "1.1.1.1",
				PrefixLength: 24,
				Enabled:      true,
			}}},
			Connected: true,
			Dedicated: false,
		}},
	}

	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	AddToCleanupListOpenApi(createdEdge.EdgeGateway.Name, check.TestName(), openApiEndpoint)

	createdEdge.EdgeGateway.Name = "renamed-edge"
	updatedEdge, err := createdEdge.Update(createdEdge.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(updatedEdge.EdgeGateway.Name, Equals, "renamed-edge")

	// FIQL filtering test
	queryParams := url.Values{}
	queryParams.Add("filter", "name==renamed-edge")
	//
	egws, err := adminOrg.GetAllNsxtEdgeGateways(queryParams)
	check.Assert(err, IsNil)
	check.Assert(len(egws) >= 1, Equals, true)

	// Lookup using different available methods
	e1, err := adminOrg.GetNsxtEdgeGatewayByName(updatedEdge.EdgeGateway.Name)
	check.Assert(err, IsNil)
	check.Assert(e1, NotNil)
	e2, err := org.GetNsxtEdgeGatewayByName(updatedEdge.EdgeGateway.Name)
	check.Assert(err, IsNil)
	check.Assert(e2, NotNil)
	e3, err := nsxtVdc.GetNsxtEdgeGatewayByName(updatedEdge.EdgeGateway.Name)
	check.Assert(err, IsNil)
	check.Assert(e3, NotNil)
	e4, err := adminOrg.GetNsxtEdgeGatewayById(updatedEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(e4, NotNil)
	e5, err := org.GetNsxtEdgeGatewayById(updatedEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(e5, NotNil)
	e6, err := nsxtVdc.GetNsxtEdgeGatewayById(updatedEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(e6, NotNil)

	// Try to search for NSX-T edge gateway in NSX-V VDC and expect it to be not found
	expectNil, err := nsxvVdc.GetNsxtEdgeGatewayByName(updatedEdge.EdgeGateway.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(expectNil, IsNil)
	expectNil, err = nsxvVdc.GetNsxtEdgeGatewayById(updatedEdge.EdgeGateway.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(expectNil, IsNil)

	// Ensure all methods found the same edge gateway
	check.Assert(e1.EdgeGateway.ID, Equals, e2.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e3.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e4.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e5.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e6.EdgeGateway.ID)

	// Try out GetUsedIpAddresses function
	usedIPs, err := updatedEdge.GetUsedIpAddresses(nil)
	check.Assert(err, IsNil)
	check.Assert(usedIPs, NotNil)

	ipAddr, err := updatedEdge.GetUnusedExternalIPAddresses(1, netip.Prefix{}, false)
	// Expect an error as no ranges were assigned
	check.Assert(err, NotNil)
	check.Assert(ipAddr, DeepEquals, []netip.Addr(nil))

	// Try a refresh operation
	err = updatedEdge.Refresh()
	check.Assert(err, IsNil)
	check.Assert(updatedEdge.EdgeGateway.ID, Equals, e1.EdgeGateway.ID)

	// Cleanup
	err = updatedEdge.Delete()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_NsxtEdgeVdcGroup(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(nsxtExternalNetwork, NotNil)

	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)

	egwDefinition := &types.OpenAPIEdgeGateway{
		Name:        "nsx-t-edge",
		Description: "nsx-t-edge-description",
		OwnerRef: &types.OpenApiReference{
			ID: vdc.Vdc.ID,
		},
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{{
			UplinkID: nsxtExternalNetwork.ExternalNetwork.ID,
			Subnets: types.OpenAPIEdgeGatewaySubnets{Values: []types.OpenAPIEdgeGatewaySubnetValue{{
				Gateway:      "1.1.1.1",
				PrefixLength: 24,
				Enabled:      true,
			}}},
			Connected: true,
			Dedicated: false,
		}},
	}

	// Create Edge Gateway in VDC
	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdc:.*`)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	PrependToCleanupListOpenApi(createdEdge.EdgeGateway.Name, check.TestName(), openApiEndpoint)

	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Equals, egwDefinition.OwnerRef.ID)

	// Move Edge Gateway to VDC Group
	movedGateway, err := createdEdge.MoveToVdcOrVdcGroup(vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Equals, vdcGroup.VdcGroup.Id)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdcGroup:.*`)

	// Get by name and owner ID
	edgeByNameAndOwnerId, err := org.GetNsxtEdgeGatewayByNameAndOwnerId(createdEdge.EdgeGateway.Name, vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)

	// Check lookup of Edge Gateways in VDC Groups
	edgeInVdcGroup, err := vdcGroup.GetNsxtEdgeGatewayByName(createdEdge.EdgeGateway.Name)
	check.Assert(err, IsNil)

	// Ensure both methods for retrieval get the same object
	check.Assert(edgeByNameAndOwnerId.EdgeGateway, DeepEquals, edgeInVdcGroup.EdgeGateway)

	// Masking known variables that have change for deep check
	edgeInVdcGroup.EdgeGateway.OwnerRef.Name = ""
	check.Assert(edgeInVdcGroup.EdgeGateway, DeepEquals, createdEdge.EdgeGateway)

	// Move Edge Gateway back to VDC from VDC Group
	egwDefinition.OwnerRef.ID = vdc.Vdc.ID
	egwDefinition.ID = movedGateway.EdgeGateway.ID

	movedBackToVdcEdge, err := movedGateway.Update(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(movedBackToVdcEdge.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdc:.*`)

	// Ignore known to differ fields, but check that whole Edge Gateway structure remains the same
	// as it is important to perform update operations without impacting configuration itself

	// Fields to ignore on both sides
	createdEdge.EdgeGateway.OrgVdc = movedBackToVdcEdge.EdgeGateway.OrgVdc
	createdEdge.EdgeGateway.OwnerRef = movedBackToVdcEdge.EdgeGateway.OwnerRef
	check.Assert(movedBackToVdcEdge.EdgeGateway, DeepEquals, createdEdge.EdgeGateway)

	// Remove Edge Gateway
	err = movedBackToVdcEdge.Delete()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_NsxtEdgeGatewayUsedAndUnusedIPs(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	nsxvVdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(nsxvVdc, NotNil)
	nsxtVdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	if ContainsNotFound(err) {
		check.Skip(fmt.Sprintf("No NSX-T VDC (%s) found - skipping test", vcd.config.VCD.Nsxt.Vdc))
	}
	check.Assert(err, IsNil)
	check.Assert(nsxtVdc, NotNil)

	// NSX-T details
	man, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	nsxtManagerId, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", extractUuid(man[0].HREF))
	check.Assert(err, IsNil)

	tier0RouterVrf, err := vcd.client.GetImportableNsxtTier0RouterByName(vcd.config.VCD.Nsxt.Tier0router, nsxtManagerId)
	check.Assert(err, IsNil)
	backingId := tier0RouterVrf.NsxtTier0Router.ID

	netNsxt := &types.ExternalNetworkV2{
		Name: check.TestName(),
		Subnets: types.ExternalNetworkV2Subnets{Values: []types.ExternalNetworkV2Subnet{
			{
				Gateway:      "1.1.1.1",
				PrefixLength: 24,
				IPRanges: types.ExternalNetworkV2IPRanges{Values: []types.ExternalNetworkV2IPRange{
					{
						StartAddress: "1.1.1.3",
						EndAddress:   "1.1.1.25",
					},
				}},
				Enabled: true,
			},
		}},
		NetworkBackings: types.ExternalNetworkV2Backings{Values: []types.ExternalNetworkV2Backing{
			{
				BackingID: backingId,
				NetworkProvider: types.NetworkProvider{
					ID: nsxtManagerId,
				},
				BackingTypeValue: types.ExternalNetworkBackingTypeNsxtTier0Router,
			},
		}},
	}
	createdNet, err := CreateExternalNetworkV2(vcd.client, netNsxt)
	check.Assert(err, IsNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks + createdNet.ExternalNetwork.ID
	AddToCleanupListOpenApi(createdNet.ExternalNetwork.Name, check.TestName(), openApiEndpoint)

	egwDefinition := &types.OpenAPIEdgeGateway{
		Name: check.TestName(),
		OrgVdc: &types.OpenApiReference{
			ID: nsxtVdc.Vdc.ID,
		},
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{{
			UplinkID: createdNet.ExternalNetwork.ID,
			Subnets: types.OpenAPIEdgeGatewaySubnets{Values: []types.OpenAPIEdgeGatewaySubnetValue{{
				Gateway:      createdNet.ExternalNetwork.Subnets.Values[0].Gateway,
				PrefixLength: createdNet.ExternalNetwork.Subnets.Values[0].PrefixLength,
				Enabled:      true,
				IPRanges: &types.OpenApiIPRanges{
					Values: []types.OpenApiIPRangeValues{
						{
							StartAddress: createdNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].StartAddress,
							EndAddress:   createdNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
						},
					},
				},
			}}},
			Connected: true,
			Dedicated: false,
		}},
	}

	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)
	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	PrependToCleanupListOpenApi(createdEdge.EdgeGateway.Name, check.TestName(), openApiEndpoint)

	// Try out GetUsedIpAddresses function
	usedIPs, err := createdEdge.GetUsedIpAddresses(nil)
	check.Assert(err, IsNil)
	check.Assert(usedIPs, NotNil)

	// Edge Gateway always allocates 1 IP as its primary
	check.Assert(usedIPs[0].IPAddress, Equals, "1.1.1.3")
	check.Assert(usedIPs[0].Category, Equals, "PRIMARY_IP")

	// Attempt to get 1 unallocated IP
	ipAddr, err := createdEdge.GetUnusedExternalIPAddresses(1, netip.Prefix{}, false)
	check.Assert(err, IsNil)
	ipsCompared := compareEachIpElementAndOrder(ipAddr, []netip.Addr{netip.MustParseAddr("1.1.1.4")})
	check.Assert(ipsCompared, Equals, true)

	// Attempt to get 10 unallocated IPs
	ipAddr, err = createdEdge.GetUnusedExternalIPAddresses(10, netip.Prefix{}, true)
	check.Assert(err, IsNil)
	ipsCompared = compareEachIpElementAndOrder(ipAddr, []netip.Addr{
		netip.MustParseAddr("1.1.1.4"),
		netip.MustParseAddr("1.1.1.5"),
		netip.MustParseAddr("1.1.1.6"),
		netip.MustParseAddr("1.1.1.7"),
		netip.MustParseAddr("1.1.1.8"),
		netip.MustParseAddr("1.1.1.9"),
		netip.MustParseAddr("1.1.1.10"),
		netip.MustParseAddr("1.1.1.11"),
		netip.MustParseAddr("1.1.1.12"),
		netip.MustParseAddr("1.1.1.13"),
	})
	check.Assert(ipsCompared, Equals, true)

	// Attempt to get IP but filter it off by prefix
	ipAddr, err = createdEdge.GetUnusedExternalIPAddresses(1, netip.MustParsePrefix("192.168.1.1/24"), false)
	// Expect an error because Edge Gateway does not have IPs from required subnet 192.168.1.1/24
	check.Assert(err, NotNil)
	check.Assert(ipAddr, IsNil)

	// Attempt to get all unused IPs
	allIps, err := createdEdge.GetAllUnusedExternalIPAddresses(true)
	check.Assert(err, IsNil)
	ipsCompared = compareEachIpElementAndOrder(allIps, []netip.Addr{
		netip.MustParseAddr("1.1.1.4"),
		netip.MustParseAddr("1.1.1.5"),
		netip.MustParseAddr("1.1.1.6"),
		netip.MustParseAddr("1.1.1.7"),
		netip.MustParseAddr("1.1.1.8"),
		netip.MustParseAddr("1.1.1.9"),
		netip.MustParseAddr("1.1.1.10"),
		netip.MustParseAddr("1.1.1.11"),
		netip.MustParseAddr("1.1.1.12"),
		netip.MustParseAddr("1.1.1.13"),
		netip.MustParseAddr("1.1.1.14"),
		netip.MustParseAddr("1.1.1.15"),
		netip.MustParseAddr("1.1.1.16"),
		netip.MustParseAddr("1.1.1.17"),
		netip.MustParseAddr("1.1.1.18"),
		netip.MustParseAddr("1.1.1.19"),
		netip.MustParseAddr("1.1.1.20"),
		netip.MustParseAddr("1.1.1.21"),
		netip.MustParseAddr("1.1.1.22"),
		netip.MustParseAddr("1.1.1.23"),
		netip.MustParseAddr("1.1.1.24"),
		netip.MustParseAddr("1.1.1.25"),
	})
	check.Assert(ipsCompared, Equals, true)
	check.Assert(len(allIps), Equals, 22)

	// Verify that GetAllocatedIpCount returns correct number of allocated IPs
	//GetAllocatedIpCount
	totalAllocationIpCount, err := createdEdge.GetAllocatedIpCount(true)
	check.Assert(err, IsNil)
	check.Assert(totalAllocationIpCount, NotNil)
	check.Assert(*totalAllocationIpCount, Equals, 23) // 22 unused IPs + 1 primary

	// Cleanup
	err = createdEdge.Delete()
	check.Assert(err, IsNil)

	err = createdNet.Delete()
	check.Assert(err, IsNil)
}

// compareEachIpElementAndOrder performs comparison of IPs in a slice as default check.Assert
// functions is not able to perform this comparison
func compareEachIpElementAndOrder(ipSlice1, ipSlice2 []netip.Addr) bool {
	if len(ipSlice1) != len(ipSlice2) {
		return false
	}

	for index := range ipSlice1 {
		if ipSlice1[index] != ipSlice2[index] {
			return false
		}
	}

	return true
}
