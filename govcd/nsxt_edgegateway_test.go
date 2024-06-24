//go:build network || nsxt || functional || openapi || ALL

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
	vcd.skipIfNotSysAdmin(check)

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

	edgeCluster, err := nsxtVdc.GetNsxtEdgeClusterByName(vcd.config.VCD.Nsxt.NsxtEdgeCluster)
	check.Assert(err, IsNil)
	check.Assert(edgeCluster, NotNil)

	createdEdge.EdgeGateway.EdgeClusterConfig = &types.OpenAPIEdgeGatewayEdgeClusterConfig{
		PrimaryEdgeCluster: types.OpenAPIEdgeGatewayEdgeCluster{
			BackingID: edgeCluster.NsxtEdgeCluster.ID,
		},
	}
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
	vcd.skipIfNotSysAdmin(check)

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

	// Remove VDC Group
	err = vdcGroup.Delete()
	check.Assert(err, IsNil)

	// Remove VDC
	err = vdc.DeleteWait(true, true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_NsxtEdgeGatewayUsedAndUnusedIPs(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)
	vcd.skipIfNotSysAdmin(check)

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

	// Get used and unused IP counts
	usedIpCount, unusedIpCount, err := createdEdge.GetUsedAndUnusedExternalIPAddressCountWithLimit(false, 5)
	check.Assert(err, IsNil)
	check.Assert(unusedIpCount, Equals, int64(4))
	check.Assert(usedIpCount, Equals, int64(1))

	// Verify that GetAllocatedIpCount returns correct number of allocated IPs
	totalAllocationIpCount, err := createdEdge.GetAllocatedIpCount(true)
	check.Assert(err, IsNil)
	check.Assert(totalAllocationIpCount, NotNil)
	check.Assert(totalAllocationIpCount, Equals, 23) // 22 unused IPs + 1 primary

	// Try to deallocate more IPs than allocated
	failedDeallocationIpCount, err := createdEdge.QuickDeallocateIpCount(24)
	check.Assert(err, NotNil)
	check.Assert(failedDeallocationIpCount, IsNil)

	// Check that failed deallocation did not change the number of allocated IPs
	allocatedIpCountAfterFailedDeallocation, err := createdEdge.GetAllocatedIpCount(true)
	check.Assert(err, IsNil)
	check.Assert(allocatedIpCountAfterFailedDeallocation, NotNil)
	check.Assert(allocatedIpCountAfterFailedDeallocation, Equals, 23) // 22 unused IPs + 1 primary

	// Try to deallocate all IPs including primary. Expect a failure as an Edge Gateway must always
	// have a primary IP
	failedDeallocationIpCount, err = createdEdge.QuickDeallocateIpCount(23)
	check.Assert(err, NotNil)
	check.Assert(failedDeallocationIpCount, IsNil)

	// Deallocate 22 IP addresses
	deallocatedEdge, err := createdEdge.QuickDeallocateIpCount(22)
	check.Assert(err, IsNil)

	allocatedIpCountAfterDeallocation, err := deallocatedEdge.GetAllocatedIpCount(true)
	check.Assert(err, IsNil)
	check.Assert(allocatedIpCountAfterDeallocation, NotNil)
	check.Assert(allocatedIpCountAfterDeallocation, Equals, 1) // 1 primary

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

// Test_NsxtEdgeQoS tests QoS config (NSX-T Edge Gateway Rate Limiting) retrieval and update
func (vcd *TestVCD) Test_NsxtEdgeQoS(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointQosProfiles)
	if vcd.config.VCD.Nsxt.GatewayQosProfile == "" {
		check.Skip("No NSX-T Edge Gateway QoS Profile configured")
	}

	// Get QoS profile to use
	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers), Equals, 1)

	uuid, err := GetUuidFromHref(nsxtManagers[0].HREF, true)
	check.Assert(err, IsNil)
	urn, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", uuid)
	check.Assert(err, IsNil)

	qosProfile, err := vcd.client.GetNsxtEdgeGatewayQosProfileByDisplayName(urn, vcd.config.VCD.Nsxt.GatewayQosProfile)
	check.Assert(err, IsNil)
	check.Assert(qosProfile, NotNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Fetch current QoS config
	qosConfig, err := edge.GetQoS()
	check.Assert(err, IsNil)
	check.Assert(qosConfig, NotNil)
	check.Assert(qosConfig.EgressProfile, IsNil)
	check.Assert(qosConfig.IngressProfile, IsNil)

	// Create new QoS config
	newQosConfig := &types.NsxtEdgeGatewayQos{
		IngressProfile: &types.OpenApiReference{ID: qosProfile.NsxtEdgeGatewayQosProfile.ID},
		EgressProfile:  &types.OpenApiReference{ID: qosProfile.NsxtEdgeGatewayQosProfile.ID},
	}

	// Update QoS config
	updatedEdgeQosConfig, err := edge.UpdateQoS(newQosConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedEdgeQosConfig, NotNil)

	// Check that updates were applied
	check.Assert(updatedEdgeQosConfig.EgressProfile.ID, Equals, newQosConfig.EgressProfile.ID)
	check.Assert(updatedEdgeQosConfig.IngressProfile.ID, Equals, newQosConfig.IngressProfile.ID)

	// Remove QoS config
	updatedEdgeQosConfig, err = edge.UpdateQoS(&types.NsxtEdgeGatewayQos{})
	check.Assert(err, IsNil)
	check.Assert(updatedEdgeQosConfig, NotNil)
}

func (vcd *TestVCD) Test_NsxtEdgeDhcpForwarder(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayDhcpForwarder)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)
	AddToCleanupList(vcd.config.VCD.Nsxt.EdgeGateway, "nsxtDhcpForwarder", vcd.config.VCD.Org, check.TestName())

	// Fetch current DHCP forwarder config
	dhcpForwarderConfig, err := edge.GetDhcpForwarder()
	check.Assert(err, IsNil)
	check.Assert(dhcpForwarderConfig.Enabled, Equals, false)
	check.Assert(dhcpForwarderConfig.DhcpServers, DeepEquals, []string(nil))

	// Create new DHCP Forwarder config
	testDhcpServers := []string{
		"1.1.1.1",
		"192.168.2.254",
		"fe80::abcd",
	}

	newDhcpForwarderConfig := &types.NsxtEdgeGatewayDhcpForwarder{
		Enabled:     true,
		DhcpServers: testDhcpServers,
	}

	// Update DHCP forwarder config
	updatedEdgeDhcpForwarderConfig, err := edge.UpdateDhcpForwarder(newDhcpForwarderConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedEdgeDhcpForwarderConfig, NotNil)

	// Check that updates were applied
	check.Assert(updatedEdgeDhcpForwarderConfig.Enabled, Equals, true)
	check.Assert(updatedEdgeDhcpForwarderConfig.DhcpServers, DeepEquals, testDhcpServers)

	// remove the last dhcp server from the list
	testDhcpServers = testDhcpServers[0:2]
	newDhcpForwarderConfig.DhcpServers = testDhcpServers

	updatedEdgeDhcpForwarderConfig, err = edge.UpdateDhcpForwarder(newDhcpForwarderConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedEdgeDhcpForwarderConfig, NotNil)

	// Check that updates were applied
	check.Assert(updatedEdgeDhcpForwarderConfig.Enabled, Equals, true)
	check.Assert(updatedEdgeDhcpForwarderConfig.DhcpServers, DeepEquals, testDhcpServers)

	// Add servers to the list
	testDhcpServers = append(testDhcpServers, "192.254.0.2")
	newDhcpForwarderConfig.DhcpServers = testDhcpServers

	updatedEdgeDhcpForwarderConfig, err = edge.UpdateDhcpForwarder(newDhcpForwarderConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedEdgeDhcpForwarderConfig, NotNil)

	// Check that updates were applied
	check.Assert(updatedEdgeDhcpForwarderConfig.Enabled, Equals, true)
	check.Assert(updatedEdgeDhcpForwarderConfig.DhcpServers, DeepEquals, testDhcpServers)

	// Disable DHCP forwarder config
	newDhcpForwarderConfig.Enabled = false

	// Update DHCP forwarder config
	updatedEdgeDhcpForwarderConfig, err = edge.UpdateDhcpForwarder(newDhcpForwarderConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedEdgeDhcpForwarderConfig, NotNil)

	// Check that updates were applied
	check.Assert(updatedEdgeDhcpForwarderConfig.Enabled, Equals, false)
	check.Assert(updatedEdgeDhcpForwarderConfig.DhcpServers, DeepEquals, testDhcpServers)

	_, err = edge.UpdateDhcpForwarder(&types.NsxtEdgeGatewayDhcpForwarder{})
	check.Assert(err, IsNil)
}

// Test_NsxtEdgeSlaacProfile tests SLAAC profile (NSX-T Edge Gateway DHCPv6) retrieval and update
func (vcd *TestVCD) Test_NsxtEdgeSlaacProfile(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewaySlaacProfile)

	edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)
	AddToCleanupList(vcd.config.VCD.Nsxt.EdgeGateway, "slaacProfile", vcd.config.VCD.Org, check.TestName())

	// Fetch current SLAAC Profile
	slaacProfile, err := edge.GetSlaacProfile()
	check.Assert(err, IsNil)
	check.Assert(slaacProfile, NotNil)
	check.Assert(slaacProfile.Enabled, Equals, false)

	// Create new SLAAC config in SLAAC mode
	newSlaacProfile := &types.NsxtEdgeGatewaySlaacProfile{
		Enabled: true,
		Mode:    "SLAAC",
		DNSConfig: types.NsxtEdgeGatewaySlaacProfileDNSConfig{
			DNSServerIpv6Addresses: []string{"2001:4860:4860::8888", "2001:4860:4860::8844"},
			DomainNames:            []string{"non-existing.org.tld", "fake.org.tld"},
		},
	}

	// Update SLAAC profile
	updatedSlaacProfile, err := edge.UpdateSlaacProfile(newSlaacProfile)
	check.Assert(err, IsNil)
	check.Assert(updatedSlaacProfile, NotNil)
	check.Assert(updatedSlaacProfile, DeepEquals, newSlaacProfile)

	// Create new SLAAC config in DHCPv6 mode
	newSlaacProfileDhcpv6 := &types.NsxtEdgeGatewaySlaacProfile{
		Enabled: true,
		Mode:    "DHCPv6",
		DNSConfig: types.NsxtEdgeGatewaySlaacProfileDNSConfig{
			DNSServerIpv6Addresses: []string{},
			DomainNames:            []string{},
		},
	}

	// Update SLAAC profile
	updatedSlaacProfileDhcpv6, err := edge.UpdateSlaacProfile(newSlaacProfileDhcpv6)
	check.Assert(err, IsNil)
	check.Assert(updatedSlaacProfileDhcpv6, NotNil)
	check.Assert(updatedSlaacProfileDhcpv6, DeepEquals, newSlaacProfileDhcpv6)

	// Cleanup
	updatedSlaacProfile, err = edge.UpdateSlaacProfile(&types.NsxtEdgeGatewaySlaacProfile{Enabled: false, Mode: "DISABLED"})
	check.Assert(err, IsNil)
	check.Assert(updatedSlaacProfile, NotNil)
}

// Test_NsxtEdgeCreateWithT0AndExternalNetworks checks that IP Allocation counts and External
// Network attachment works well with NSX-T T0 Gateway backed external network
func (vcd *TestVCD) Test_NsxtEdgeCreateWithT0AndExternalNetworks(check *C) {
	test_NsxtEdgeCreateWithExternalNetworks(vcd, check, vcd.config.VCD.Nsxt.Tier0router, types.ExternalNetworkBackingTypeNsxtTier0Router)
}

// Test_NsxtEdgeCreateWithT0VrfAndExternalNetworks checks that IP Allocation counts and External
// Network attachment works well with NSX-T T0 VRF Gateway backed external network
func (vcd *TestVCD) Test_NsxtEdgeCreateWithT0VrfAndExternalNetworks(check *C) {
	test_NsxtEdgeCreateWithExternalNetworks(vcd, check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtVrfTier0Router)
}

func test_NsxtEdgeCreateWithExternalNetworks(vcd *TestVCD, check *C, backingRouter, backingRouterType string) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("Segment Backed External Network uplinks are supported in VCD 10.4.1+")
	}

	if vcd.config.VCD.Nsxt.NsxtImportSegment == "" || vcd.config.VCD.Nsxt.NsxtImportSegment2 == "" {
		check.Skip("NSX-T Imported Segments are not configured")
	}

	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)
	vcd.skipIfNotSysAdmin(check)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	nsxtVdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	if ContainsNotFound(err) {
		check.Skip(fmt.Sprintf("No NSX-T VDC (%s) found - skipping test", vcd.config.VCD.Nsxt.Vdc))
	}
	check.Assert(err, IsNil)
	check.Assert(nsxtVdc, NotNil)

	// Setup 2 NSX-T Segment backed External Networks and 1 T0 or T0 VRF backed networks
	nsxtManager, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	nsxtManagerId, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", extractUuid(nsxtManager[0].HREF))
	check.Assert(err, IsNil)

	//	T0 backed external network
	backingExtNet := getBackingIdByNameAndType(check, backingRouter, backingRouterType, vcd, nsxtManagerId)
	nsxtExternalNetworkCfg := t0vrfBackedExternalNetworkConfig(vcd, check.TestName()+"-t0", "89.1.1", backingRouterType, backingExtNet, nsxtManagerId)
	nsxtExternalNetwork, err := CreateExternalNetworkV2(vcd.client, nsxtExternalNetworkCfg)
	check.Assert(err, IsNil)
	check.Assert(nsxtExternalNetwork, NotNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks + nsxtExternalNetwork.ExternalNetwork.ID
	AddToCleanupListOpenApi(nsxtExternalNetwork.ExternalNetwork.Name, check.TestName(), openApiEndpoint)

	// First NSX-T Segment backed network
	backingId1 := getBackingIdByNameAndType(check, vcd.config.VCD.Nsxt.NsxtImportSegment, types.ExternalNetworkBackingTypeNsxtSegment, vcd, nsxtManagerId)
	segmentBackedNet1Cfg := t0vrfBackedExternalNetworkConfig(vcd, check.TestName()+"-1", "1.1.1", types.ExternalNetworkBackingTypeNsxtSegment, backingId1, nsxtManagerId)
	segmentBackedNet1, err := CreateExternalNetworkV2(vcd.client, segmentBackedNet1Cfg)
	check.Assert(err, IsNil)
	check.Assert(segmentBackedNet1, NotNil)
	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks + segmentBackedNet1.ExternalNetwork.ID
	AddToCleanupListOpenApi(segmentBackedNet1.ExternalNetwork.Name, check.TestName(), openApiEndpoint)

	// Second NSX-T Segment backed network
	backingId2 := getBackingIdByNameAndType(check, vcd.config.VCD.Nsxt.NsxtImportSegment2, types.ExternalNetworkBackingTypeNsxtSegment, vcd, nsxtManagerId)
	segmentBackedNet2Cfg := t0vrfBackedExternalNetworkConfig(vcd, check.TestName()+"-2", "4.4.4", types.ExternalNetworkBackingTypeNsxtSegment, backingId2, nsxtManagerId)
	segmentBackedNet2, err := CreateExternalNetworkV2(vcd.client, segmentBackedNet2Cfg)
	check.Assert(err, IsNil)
	check.Assert(segmentBackedNet2, NotNil)
	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks + segmentBackedNet2.ExternalNetwork.ID
	AddToCleanupListOpenApi(segmentBackedNet1.ExternalNetwork.Name, check.TestName(), openApiEndpoint)
	// Setup 2 NSX-T Segment backed External Networks and 1 T0 or T0 VRF backed networks

	egwDefinition := &types.OpenAPIEdgeGateway{
		Name:        "nsx-t-edge",
		Description: "nsx-t-edge-description",
		OrgVdc: &types.OpenApiReference{
			ID: nsxtVdc.Vdc.ID,
		},
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{
			{
				UplinkID: nsxtExternalNetwork.ExternalNetwork.ID,
				Subnets: types.OpenAPIEdgeGatewaySubnets{Values: []types.OpenAPIEdgeGatewaySubnetValue{{
					Gateway:      "5.1.1.1",
					PrefixLength: 24,
					Enabled:      true,
				}}},
				Connected: true,
				Dedicated: false,
			},
			{
				UplinkID: segmentBackedNet1.ExternalNetwork.ID,
				Subnets: types.OpenAPIEdgeGatewaySubnets{Values: []types.OpenAPIEdgeGatewaySubnetValue{{
					Gateway:              "1.1.1.1",
					PrefixLength:         24,
					Enabled:              true,
					AutoAllocateIPRanges: true,
					PrimaryIP:            "1.1.1.5",
					TotalIPCount:         addrOf(4),
				}}},
				Connected: true,
				Dedicated: false,
			},
			{
				UplinkID: segmentBackedNet2.ExternalNetwork.ID,
				Subnets: types.OpenAPIEdgeGatewaySubnets{Values: []types.OpenAPIEdgeGatewaySubnetValue{{
					Gateway:              "4.4.4.1",
					PrefixLength:         24,
					Enabled:              true,
					AutoAllocateIPRanges: true,
					TotalIPCount:         addrOf(7),
				}}},
				Connected: true,
				Dedicated: false,
			},
		},
	}

	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)
	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	PrependToCleanupListOpenApi(createdEdge.EdgeGateway.Name, check.TestName(), openApiEndpoint)

	// Retrieve edge gateway
	retrievedEdge, err := adminOrg.GetNsxtEdgeGatewayById(createdEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(retrievedEdge, NotNil)

	// Check IP allocation in NSX-T Segment backed networks
	totalAllocatedIpCountSegmentBacked, err := retrievedEdge.GetAllocatedIpCountByUplinkType(false, types.ExternalNetworkBackingTypeNsxtSegment)
	check.Assert(err, IsNil)
	check.Assert(totalAllocatedIpCountSegmentBacked, Equals, (4 + 7))

	// Check IP allocation in NSX-T T0 backed networks
	totalAllocatedIpCountT0backed, err := retrievedEdge.GetAllocatedIpCountByUplinkType(false, backingRouterType)
	check.Assert(err, IsNil)
	check.Assert(totalAllocatedIpCountT0backed, Equals, 1)

	totalAllocatedIpCountForPrimaryUplink, err := retrievedEdge.GetPrimaryNetworkAllocatedIpCount(false)
	check.Assert(err, IsNil)
	check.Assert(totalAllocatedIpCountForPrimaryUplink, Equals, 1)

	// Check IP allocation for all subnets
	totalAllocatedIpCount, err := retrievedEdge.GetAllocatedIpCount(false)
	check.Assert(err, IsNil)
	check.Assert(totalAllocatedIpCount, Equals, (1 + 4 + 7))

	createdEdge.EdgeGateway.Name = check.TestName() + "-renamed-edge"
	updatedEdge, err := createdEdge.Update(createdEdge.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(updatedEdge.EdgeGateway.Name, Equals, createdEdge.EdgeGateway.Name)

	// Check IP allocation in NSX-T Segment backed networks
	totalAllocatedIpCountSegmentBacked, err = updatedEdge.GetAllocatedIpCountByUplinkType(false, types.ExternalNetworkBackingTypeNsxtSegment)
	check.Assert(err, IsNil)
	check.Assert(totalAllocatedIpCountSegmentBacked, Equals, (4 + 7))

	// Check IP allocation in NSX-T T0 backed networks
	totalAllocatedIpCountT0backed, err = updatedEdge.GetAllocatedIpCountByUplinkType(false, backingRouterType)
	check.Assert(err, IsNil)
	check.Assert(totalAllocatedIpCountT0backed, Equals, 1)

	// Check IP allocation for all subnets
	totalAllocatedIpCount, err = updatedEdge.GetAllocatedIpCount(false)
	check.Assert(err, IsNil)
	check.Assert(totalAllocatedIpCount, Equals, (1 + 4 + 7))

	// Cleanup
	err = updatedEdge.Delete()
	check.Assert(err, IsNil)

	err = segmentBackedNet2.Delete()
	check.Assert(err, IsNil)

	err = segmentBackedNet1.Delete()
	check.Assert(err, IsNil)

	err = nsxtExternalNetwork.Delete()
	check.Assert(err, IsNil)

}

func t0vrfBackedExternalNetworkConfig(vcd *TestVCD, name, ipPrefix string, backingType, backingId, NetworkProviderId string) *types.ExternalNetworkV2 {
	net := &types.ExternalNetworkV2{
		Name: name,
		Subnets: types.ExternalNetworkV2Subnets{Values: []types.ExternalNetworkV2Subnet{
			{
				Gateway:      ipPrefix + ".1",
				PrefixLength: 24,
				IPRanges: types.ExternalNetworkV2IPRanges{Values: []types.ExternalNetworkV2IPRange{
					{
						StartAddress: ipPrefix + ".3",
						EndAddress:   ipPrefix + ".50",
					},
				}},
				Enabled: true,
			},
		}},
		NetworkBackings: types.ExternalNetworkV2Backings{Values: []types.ExternalNetworkV2Backing{
			{
				BackingID: backingId,
				NetworkProvider: types.NetworkProvider{
					ID: NetworkProviderId,
				},
				BackingTypeValue: backingType,
			},
		}},
	}

	return net
}
