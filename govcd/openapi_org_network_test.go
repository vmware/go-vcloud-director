//go:build network || nsxt || functional || openapi || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtOrgVdcNetworkIsolated(check *C) {
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	skipNoNsxtConfiguration(vcd, check)
	vcd.skipIfNotSysAdmin(check) // this test uses GetNsxtEdgeClusterByName, which requires system administrator privileges

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OwnerRef:    &types.OpenApiReference{ID: vcd.nsxtVdc.Vdc.ID},
		NetworkType: types.OrgVdcNetworkTypeIsolated,
		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "2.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "bar.foo",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "2.1.1.20",
								EndAddress:   "2.1.1.30",
							},
							{
								StartAddress: "2.1.1.40",
								EndAddress:   "2.1.1.50",
							},
							{
								StartAddress: "2.1.1.88",
								EndAddress:   "2.1.1.92",
							},
						}},
				},
			},
		},
	}

	runOpenApiOrgVdcNetworkTestWithSegmentProfileTemplateEndpoint(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeIsolated)
	runOpenApiOrgVdcNetworkTestWithSegmentProfileTemplate(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeIsolated)
	runOpenApiOrgVdcNetworkTest(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeIsolated, []dhcpConfigFunc{nsxtDhcpConfigNetworkMode})
	runOpenApiOrgVdcNetworkWithVdcGroupTest(check, vcd, orgVdcNetworkConfig, types.OrgVdcNetworkTypeIsolated, []dhcpConfigFunc{nsxtDhcpConfigNetworkMode})
}

func (vcd *TestVCD) Test_NsxtOrgVdcNetworkRouted(check *C) {
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	skipNoNsxtConfiguration(vcd, check)
	vcd.skipIfNotSysAdmin(check) // this test uses GetNsxtEdgeClusterByName, which requires system administrator privileges

	egw, err := vcd.org.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OwnerRef:    &types.OpenApiReference{ID: vcd.nsxtVdc.Vdc.ID},
		NetworkType: types.OrgVdcNetworkTypeRouted,

		// Connection is used for "routed" network
		Connection: &types.Connection{
			RouterRef: types.OpenApiReference{
				ID: egw.EdgeGateway.ID,
			},
			ConnectionType: "INTERNAL",
		},
		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "2.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "foo.bar",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "2.1.1.20",
								EndAddress:   "2.1.1.30",
							},
							{
								StartAddress: "2.1.1.40",
								EndAddress:   "2.1.1.50",
							},
							{
								StartAddress: "2.1.1.60",
								EndAddress:   "2.1.1.62",
							}, {
								StartAddress: "2.1.1.72",
								EndAddress:   "2.1.1.74",
							}, {
								StartAddress: "2.1.1.84",
								EndAddress:   "2.1.1.85",
							},
						}},
				},
			},
		},
	}

	runOpenApiOrgVdcNetworkTestWithSegmentProfileTemplateEndpoint(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeRouted)
	runOpenApiOrgVdcNetworkTestWithSegmentProfileTemplate(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeRouted)
	runOpenApiOrgVdcNetworkTest(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeRouted, []dhcpConfigFunc{nsxtRoutedDhcpConfigEdgeMode, nsxtDhcpConfigNetworkMode})
	runOpenApiOrgVdcNetworkWithVdcGroupTest(check, vcd, orgVdcNetworkConfig, types.OrgVdcNetworkTypeRouted, []dhcpConfigFunc{nsxtRoutedDhcpConfigEdgeMode, nsxtDhcpConfigNetworkMode})
}

func (vcd *TestVCD) Test_NsxtOrgVdcNetworkImportedNsxtLogicalSwitch(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	skipNoNsxtConfiguration(vcd, check)

	if vcd.config.VCD.Nsxt.NsxtImportSegment == "" {
		check.Skip("Unused NSX-T segment was not provided")
	}

	logicalSwitch, err := vcd.nsxtVdc.GetNsxtImportableSwitchByName(vcd.config.VCD.Nsxt.NsxtImportSegment)
	check.Assert(err, IsNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",

		// On v35.0 orgVdc is not supported anymore. Using ownerRef instead.
		OwnerRef: &types.OpenApiReference{ID: vcd.nsxtVdc.Vdc.ID},

		NetworkType: types.OrgVdcNetworkTypeOpaque,
		// BackingNetworkId contains NSX-T logical switch ID for Imported networks
		BackingNetworkId: logicalSwitch.NsxtImportableSwitch.ID,

		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "2.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "foo.bar",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "2.1.1.20",
								EndAddress:   "2.1.1.30",
							},
							{
								StartAddress: "2.1.1.40",
								EndAddress:   "2.1.1.50",
							},
						}},
				},
			},
		},
	}

	runOpenApiOrgVdcNetworkTestWithSegmentProfileTemplateEndpoint(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeOpaque)
	runOpenApiOrgVdcNetworkTestWithSegmentProfileTemplate(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeOpaque)
	runOpenApiOrgVdcNetworkTest(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeOpaque, nil)
	runOpenApiOrgVdcNetworkWithVdcGroupTest(check, vcd, orgVdcNetworkConfig, types.OrgVdcNetworkTypeOpaque, nil)
}

func (vcd *TestVCD) Test_NsxtOrgVdcNetworkImportedDistributedVirtualPortGroup(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	skipNoNsxtConfiguration(vcd, check)

	if vcd.config.VCD.Nsxt.NsxtDvpg == "" {
		check.Skip("Distributed Virtual Port Group was not provided")
	}

	dvpg, err := vcd.nsxtVdc.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.NsxtDvpg)
	check.Assert(err, IsNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",

		OwnerRef: &types.OpenApiReference{ID: vcd.nsxtVdc.Vdc.ID},

		NetworkType: types.OrgVdcNetworkTypeOpaque,
		// BackingNetworkId contains Distributed Virtual Port Group ID for Imported networks
		BackingNetworkId:   dvpg.VcenterImportableDvpg.BackingRef.ID,
		BackingNetworkType: types.OrgVdcNetworkBackingTypeDvPortgroup,

		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "2.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "foo.bar",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "2.1.1.20",
								EndAddress:   "2.1.1.30",
							},
							{
								StartAddress: "2.1.1.40",
								EndAddress:   "2.1.1.50",
							},
						}},
				},
			},
		},
	}

	// Org VDC network backed by Distributed Virtual Port Group can only be created in VDC (not VDC Group)
	runOpenApiOrgVdcNetworkTest(check, vcd, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeOpaque, nil)
}

func (vcd *TestVCD) Test_NsxvOrgVdcNetworkIsolated(check *C) {
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",

		// On v35.0 orgVdc is not supported anymore. Using ownerRef instead.
		OwnerRef: &types.OpenApiReference{ID: vcd.vdc.Vdc.ID},

		NetworkType: types.OrgVdcNetworkTypeIsolated,
		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "4.1.1.1",
					PrefixLength: 25,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "bar.foo",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "4.1.1.20",
								EndAddress:   "4.1.1.30",
							},
							{
								StartAddress: "4.1.1.40",
								EndAddress:   "4.1.1.50",
							},
							{
								StartAddress: "4.1.1.88",
								EndAddress:   "4.1.1.92",
							},
						}},
				},
			},
		},
	}

	runOpenApiOrgVdcNetworkTest(check, vcd, vcd.vdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeIsolated, nil)
}

func (vcd *TestVCD) Test_NsxvOrgVdcNetworkRouted(check *C) {
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)

	nsxvEdgeGateway, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, true)
	check.Assert(err, IsNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",

		// On v35.0 orgVdc is not supported anymore. Using ownerRef instead.
		OwnerRef: &types.OpenApiReference{ID: vcd.vdc.Vdc.ID},

		NetworkType: types.OrgVdcNetworkTypeRouted,

		// Connection is used for "routed" network
		Connection: &types.Connection{
			RouterRef: types.OpenApiReference{
				ID: nsxvEdgeGateway.EdgeGateway.ID,
			},
			ConnectionType: "INTERNAL",
		},
		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "2.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "foo.bar",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "2.1.1.20",
								EndAddress:   "2.1.1.30",
							},
							{
								StartAddress: "2.1.1.40",
								EndAddress:   "2.1.1.50",
							},
							{
								StartAddress: "2.1.1.60",
								EndAddress:   "2.1.1.62",
							}, {
								StartAddress: "2.1.1.72",
								EndAddress:   "2.1.1.74",
							}, {
								StartAddress: "2.1.1.84",
								EndAddress:   "2.1.1.85",
							},
						}},
				},
			},
		},
	}

	runOpenApiOrgVdcNetworkTest(check, vcd, vcd.vdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeRouted, nil)
}

func (vcd *TestVCD) Test_NsxvOrgVdcNetworkDirect(check *C) {
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	externalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(externalNetwork, NotNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",

		// On v35.0 orgVdc is not supported anymore. Using ownerRef instead.
		OwnerRef: &types.OpenApiReference{ID: vcd.vdc.Vdc.ID},

		NetworkType:   types.OrgVdcNetworkTypeDirect,
		ParentNetwork: &types.OpenApiReference{ID: externalNetwork.ExternalNetwork.ID},

		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "2.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "foo.bar",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "2.1.1.20",
								EndAddress:   "2.1.1.30",
							},
							{
								StartAddress: "2.1.1.40",
								EndAddress:   "2.1.1.50",
							},
							{
								StartAddress: "2.1.1.60",
								EndAddress:   "2.1.1.62",
							}, {
								StartAddress: "2.1.1.72",
								EndAddress:   "2.1.1.74",
							}, {
								StartAddress: "2.1.1.84",
								EndAddress:   "2.1.1.85",
							},
						}},
				},
			},
		},
	}

	runOpenApiOrgVdcNetworkTest(check, vcd, vcd.vdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeDirect, nil)
}

func runOpenApiOrgVdcNetworkTest(check *C, vcd *TestVCD, vdc *Vdc, orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork, expectNetworkType string, dhcpFunc []dhcpConfigFunc) {
	orgVdcNet, err := vdc.CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig)
	check.Assert(err, IsNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + orgVdcNet.OpenApiOrgVdcNetwork.ID
	AddToCleanupListOpenApi(orgVdcNet.OpenApiOrgVdcNetwork.Name, check.TestName(), openApiEndpoint)

	check.Assert(orgVdcNet.GetType(), Equals, expectNetworkType)

	// Check it can be found
	orgVdcNetByIdInVdc, err := vdc.GetOpenApiOrgVdcNetworkById(orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(err, IsNil)
	orgVdcNetByName, err := vdc.GetOpenApiOrgVdcNetworkByName(orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(err, IsNil)

	check.Assert(orgVdcNetByIdInVdc.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(orgVdcNetByName.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)

	// Retrieve all networks in VDC and expect newly created network to be there
	var foundNetInVdc bool
	allOrgVdcNets, err := vdc.GetAllOpenApiOrgVdcNetworks(nil)
	check.Assert(err, IsNil)
	for _, net := range allOrgVdcNets {
		if net.OpenApiOrgVdcNetwork.ID == orgVdcNet.OpenApiOrgVdcNetwork.ID {
			foundNetInVdc = true
		}
	}
	check.Assert(foundNetInVdc, Equals, true)

	// Update
	orgVdcNet.OpenApiOrgVdcNetwork.Description = check.TestName() + "updated description"
	updatedOrgVdcNet, err := orgVdcNet.Update(orgVdcNet.OpenApiOrgVdcNetwork)
	check.Assert(err, IsNil)

	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.Name, Equals, orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.Description, Equals, orgVdcNet.OpenApiOrgVdcNetwork.Description)

	// Configure DHCP if specified
	for i := range dhcpFunc {
		dhcpFunc[i](check, vcd, vdc, updatedOrgVdcNet.OpenApiOrgVdcNetwork.ID)
	}
	// Delete
	err = orgVdcNet.Delete()
	check.Assert(err, IsNil)

	// Test again if it was deleted and expect it to contain ErrorEntityNotFound
	_, err = vdc.GetOpenApiOrgVdcNetworkByName(orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vdc.GetOpenApiOrgVdcNetworkById(orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
}

type dhcpConfigFunc func(check *C, vcd *TestVCD, vdc *Vdc, orgNetId string)

func nsxtRoutedDhcpConfigEdgeMode(check *C, vcd *TestVCD, vdc *Vdc, orgNetId string) {
	printVerbose("## Testing DHCP in EDGE mode\n")
	dhcpDefinition := &types.OpenApiOrgVdcNetworkDhcp{
		Enabled: addrOf(true),
		DhcpPools: []types.OpenApiOrgVdcNetworkDhcpPools{
			{
				Enabled: addrOf(true),
				IPRange: types.OpenApiOrgVdcNetworkDhcpIpRange{
					StartAddress: "2.1.1.200",
					EndAddress:   "2.1.1.201",
				},
			},
		},
		DnsServers: []string{
			"8.8.8.8",
			"8.8.4.4",
		},
	}

	// In API versions lower than 36.1, dnsServers list does not exist
	if vdc.client.APIVCDMaxVersionIs("< 36.1") {
		dhcpDefinition.DnsServers = nil
	}

	orgVdcNetwork, err := vcd.org.GetOpenApiOrgVdcNetworkById(orgNetId)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	// Check that DHCP is not enabled
	check.Assert(orgVdcNetwork.IsDhcpEnabled(), Equals, false)

	updatedDhcp, err := vdc.UpdateOpenApiOrgVdcNetworkDhcp(orgNetId, dhcpDefinition)
	check.Assert(err, IsNil)

	// Check that DHCP is enabled
	check.Assert(orgVdcNetwork.IsDhcpEnabled(), Equals, true)
	check.Assert(dhcpDefinition, DeepEquals, updatedDhcp.OpenApiOrgVdcNetworkDhcp)

	if orgVdcNetwork.client.APIVCDMaxVersionIs(">= 36.1") {
		printVerbose("### Testing DHCP Bindings - only supported in 10.3.1+\n")
		testNsxtDhcpBinding(check, vcd, orgVdcNetwork)
	}

	err = vdc.DeleteOpenApiOrgVdcNetworkDhcp(orgNetId)
	check.Assert(err, IsNil)

	updatedDhcp2, err := orgVdcNetwork.UpdateDhcp(dhcpDefinition)
	check.Assert(err, IsNil)
	check.Assert(updatedDhcp2, NotNil)
	check.Assert(dhcpDefinition, DeepEquals, updatedDhcp2.OpenApiOrgVdcNetworkDhcp)

	err = orgVdcNetwork.DeletNetworkDhcp()
	check.Assert(err, IsNil)

	deletedDhcp, err := orgVdcNetwork.GetOpenApiOrgVdcNetworkDhcp()
	check.Assert(err, IsNil)
	check.Assert(len(deletedDhcp.OpenApiOrgVdcNetworkDhcp.DhcpPools), Equals, 0)
	check.Assert(len(deletedDhcp.OpenApiOrgVdcNetworkDhcp.DnsServers), Equals, 0)

	// Check that DHCP is not enabled
	check.Assert(orgVdcNetwork.IsDhcpEnabled(), Equals, false)
}

// nsxtDhcpConfigNetworkMode checks DHCP functionality in NETWORK mode.
// It requires that Edge Cluster is set at VDC level therefore this function does it for the
// duration of this test and restores it back
func nsxtDhcpConfigNetworkMode(check *C, vcd *TestVCD, vdc *Vdc, orgNetId string) {
	// Only supported in 10.3.1+
	if vdc.client.APIVCDMaxVersionIs("< 36.1") {
		return
	}

	printVerbose("## Testing DHCP in NETWORK mode\n")

	// DHCP in NETWORK mode requires Edge Cluster to be set for VDC and cleaned up afterward
	edgeCluster, err := vdc.GetNsxtEdgeClusterByName(vcd.config.VCD.Nsxt.NsxtEdgeCluster)
	check.Assert(err, IsNil)
	vdcNetworkProfile := &types.VdcNetworkProfile{
		ServicesEdgeCluster: &types.VdcNetworkProfileServicesEdgeCluster{
			BackingID: edgeCluster.NsxtEdgeCluster.ID,
		},
	}
	_, err = vdc.UpdateVdcNetworkProfile(vdcNetworkProfile)
	check.Assert(err, IsNil)
	defer func() {
		err := vdc.DeleteVdcNetworkProfile()
		if err != nil {
			check.Errorf("error cleaning up VDC Network Profile: %s", err)
		}
	}()

	dhcpDefinition := &types.OpenApiOrgVdcNetworkDhcp{
		Enabled:   addrOf(true),
		Mode:      "NETWORK",
		IPAddress: "2.1.1.252",
		DhcpPools: []types.OpenApiOrgVdcNetworkDhcpPools{
			{
				Enabled: addrOf(true),
				IPRange: types.OpenApiOrgVdcNetworkDhcpIpRange{
					StartAddress: "2.1.1.200",
					EndAddress:   "2.1.1.201",
				},
			},
		},
		DnsServers: []string{
			"8.8.8.8",
			"8.8.4.4",
		},
	}

	orgVdcNetwork, err := vcd.org.GetOpenApiOrgVdcNetworkById(orgNetId)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	// Check that DHCP is not enabled
	check.Assert(orgVdcNetwork.IsDhcpEnabled(), Equals, false)

	updatedDhcp, err := vdc.UpdateOpenApiOrgVdcNetworkDhcp(orgNetId, dhcpDefinition)
	check.Assert(err, IsNil)

	// Check that DHCP is enabled
	check.Assert(orgVdcNetwork.IsDhcpEnabled(), Equals, true)
	check.Assert(dhcpDefinition, DeepEquals, updatedDhcp.OpenApiOrgVdcNetworkDhcp)

	if orgVdcNetwork.client.APIVCDMaxVersionIs(">= 36.1") {
		printVerbose("### Testing DHCP Bindings - only supported in 10.3.1+\n")
		testNsxtDhcpBinding(check, vcd, orgVdcNetwork)
	}

	err = vdc.DeleteOpenApiOrgVdcNetworkDhcp(orgNetId)
	check.Assert(err, IsNil)

	updatedDhcp2, err := orgVdcNetwork.UpdateDhcp(dhcpDefinition)
	check.Assert(err, IsNil)
	check.Assert(updatedDhcp2, NotNil)

	check.Assert(dhcpDefinition, DeepEquals, updatedDhcp2.OpenApiOrgVdcNetworkDhcp)

	err = orgVdcNetwork.DeletNetworkDhcp()
	check.Assert(err, IsNil)

	deletedDhcp, err := orgVdcNetwork.GetOpenApiOrgVdcNetworkDhcp()
	check.Assert(err, IsNil)
	check.Assert(len(deletedDhcp.OpenApiOrgVdcNetworkDhcp.DhcpPools), Equals, 0)
	check.Assert(len(deletedDhcp.OpenApiOrgVdcNetworkDhcp.DnsServers), Equals, 0)

	// Check that DHCP is not enabled
	check.Assert(orgVdcNetwork.IsDhcpEnabled(), Equals, false)
}

func runOpenApiOrgVdcNetworkWithVdcGroupTest(check *C, vcd *TestVCD, orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork, expectNetworkType string, dhcpFunc []dhcpConfigFunc) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(nsxtExternalNetwork, NotNil)

	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)
	egwDefinition := &types.OpenAPIEdgeGateway{
		Name:        "nsx-for-org-network-edge",
		Description: "nsx-for-org-network-edge-description",
		OwnerRef: &types.OpenApiReference{
			ID: vdc.Vdc.ID,
		},
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{{
			UplinkID: nsxtExternalNetwork.ExternalNetwork.ID,
			Subnets: types.OpenAPIEdgeGatewaySubnets{Values: []types.OpenAPIEdgeGatewaySubnetValue{{
				Gateway:      "10.10.10.10",
				PrefixLength: 24,
				Enabled:      true,
			}}},
			Connected: true,
			Dedicated: false,
		}},
	}

	// Create Edge Gateway in VDC Group
	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdc:.*`)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	PrependToCleanupListOpenApi(createdEdge.EdgeGateway.Name, check.TestName(), openApiEndpoint)

	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Equals, egwDefinition.OwnerRef.ID)

	movedGateway, err := createdEdge.MoveToVdcOrVdcGroup(vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Equals, vdcGroup.VdcGroup.Id)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdcGroup:.*`)

	orgVdcNetworkConfig.OwnerRef.ID = vdcGroup.VdcGroup.Id
	if orgVdcNetworkConfig.Connection != nil {
		orgVdcNetworkConfig.Connection.RouterRef.ID = movedGateway.EdgeGateway.ID
	}
	orgVdcNet, err := vdcGroup.CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig)
	check.Assert(err, IsNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + orgVdcNet.OpenApiOrgVdcNetwork.ID
	AddToCleanupListOpenApi(orgVdcNet.OpenApiOrgVdcNetwork.Name, check.TestName(), openApiEndpoint)

	check.Assert(orgVdcNet.GetType(), Equals, expectNetworkType)

	// Check it can be found
	orgVdcNetByIdInVdc, err := vdcGroup.GetOpenApiOrgVdcNetworkById(orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetByIdInVdc, NotNil)
	orgVdcNetByName, err := vdcGroup.GetOpenApiOrgVdcNetworkByName(orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetByName, NotNil)

	check.Assert(orgVdcNetByIdInVdc.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(orgVdcNetByName.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)

	// Retrieve all networks in VDC and expect newly created network to be there
	var foundNetInVdc bool
	allOrgVdcNets, err := vdcGroup.GetAllOpenApiOrgVdcNetworks(nil)
	check.Assert(err, IsNil)
	for _, net := range allOrgVdcNets {
		if net.OpenApiOrgVdcNetwork.ID == orgVdcNet.OpenApiOrgVdcNetwork.ID {
			foundNetInVdc = true
		}
	}
	check.Assert(foundNetInVdc, Equals, true)

	// Update
	orgVdcNet.OpenApiOrgVdcNetwork.Description = check.TestName() + "updated description"
	updatedOrgVdcNet, err := orgVdcNet.Update(orgVdcNet.OpenApiOrgVdcNetwork)
	check.Assert(err, IsNil)

	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.Name, Equals, orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.Description, Equals, orgVdcNet.OpenApiOrgVdcNetwork.Description)

	// Configure DHCP if specified
	for i := range dhcpFunc {
		dhcpFunc[i](check, vcd, vdc, updatedOrgVdcNet.OpenApiOrgVdcNetwork.ID)
	}
	// Delete
	err = orgVdcNet.Delete()
	check.Assert(err, IsNil)

	// Test again if it was deleted and expect it to contain ErrorEntityNotFound
	_, err = vdcGroup.GetOpenApiOrgVdcNetworkByName(orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vdcGroup.GetOpenApiOrgVdcNetworkById(orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(ContainsNotFound(err), Equals, true)

	//cleanup
	err = movedGateway.Delete()
	check.Assert(err, IsNil)

	// Remove VDC group and VDC
	err = vdcGroup.Delete()
	check.Assert(err, IsNil)
	task, err := vdc.Delete(true, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

func testNsxtDhcpBinding(check *C, vcd *TestVCD, orgNet *OpenApiOrgVdcNetwork) {
	// define DHCP binding configuration
	dhcpBindingConfig := &types.OpenApiOrgVdcNetworkDhcpBinding{
		Name:        check.TestName() + "-dhcp-binding",
		Description: "dhcp binding description",
		IpAddress:   "2.1.1.231",
		MacAddress:  "00:11:22:33:44:55",
		BindingType: types.NsxtDhcpBindingTypeIpv4,
		DhcpV4BindingConfig: &types.DhcpV4BindingConfig{
			HostName:         "dhcp-binding-hostname",
			GatewayIPAddress: "2.1.1.244",
		},
	}

	// create DHCP binding
	createdDhcpBinding, err := orgNet.CreateOpenApiOrgVdcNetworkDhcpBinding(dhcpBindingConfig)
	check.Assert(err, IsNil)
	check.Assert(createdDhcpBinding, NotNil)

	// Add binding to cleanup list
	openApiEndpoint := fmt.Sprintf(types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworksDhcpBindings+"%s",
		orgNet.OpenApiOrgVdcNetwork.ID, createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)
	PrependToCleanupListOpenApi(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.Name, check.TestName(), openApiEndpoint)

	// Validate DHCP binding fields
	check.Assert(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.Name, Equals, dhcpBindingConfig.Name)
	check.Assert(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.Description, Equals, dhcpBindingConfig.Description)
	check.Assert(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.IpAddress, Equals, dhcpBindingConfig.IpAddress)
	check.Assert(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.MacAddress, Equals, dhcpBindingConfig.MacAddress)
	check.Assert(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.BindingType, Equals, dhcpBindingConfig.BindingType)

	// Get DHCP binding by ID
	getDhcpBinding, err := orgNet.GetOpenApiOrgVdcNetworkDhcpBindingById(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)
	check.Assert(err, IsNil)
	check.Assert(getDhcpBinding, NotNil)
	check.Assert(getDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.Name, Equals, dhcpBindingConfig.Name)

	// Get DHCP binding by Name
	getDhcpBindingByName, err := orgNet.GetOpenApiOrgVdcNetworkDhcpBindingByName(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.Name)
	check.Assert(err, IsNil)
	check.Assert(getDhcpBindingByName, NotNil)
	check.Assert(getDhcpBindingByName.OpenApiOrgVdcNetworkDhcpBinding.ID, Equals, createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)

	// Get all DHCP bindings
	allDhcpBindings, err := orgNet.GetAllOpenApiOrgVdcNetworkDhcpBindings(nil)
	check.Assert(err, IsNil)
	check.Assert(allDhcpBindings, NotNil)
	check.Assert(len(allDhcpBindings), Equals, 1)
	check.Assert(allDhcpBindings[0].OpenApiOrgVdcNetworkDhcpBinding.ID, Equals, createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)

	// Update DHCP binding
	dhcpBindingConfig.Description = "updated description"
	dhcpBindingConfig.IpAddress = "2.1.1.232"
	dhcpBindingConfig.MacAddress = "00:11:22:33:33:33"
	dhcpBindingConfig.ID = createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID

	updatedDhcpBinding, err := createdDhcpBinding.Update(dhcpBindingConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedDhcpBinding, NotNil)
	check.Assert(updatedDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.Description, Equals, dhcpBindingConfig.Description)
	check.Assert(updatedDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.IpAddress, Equals, dhcpBindingConfig.IpAddress)
	check.Assert(updatedDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.MacAddress, Equals, dhcpBindingConfig.MacAddress)

	// Attempt to refresh originally created binding and see if it got these new updates values as well
	err = createdDhcpBinding.Refresh()
	check.Assert(err, IsNil)
	check.Assert(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.Description, Equals, dhcpBindingConfig.Description)
	check.Assert(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.IpAddress, Equals, dhcpBindingConfig.IpAddress)
	check.Assert(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.MacAddress, Equals, dhcpBindingConfig.MacAddress)

	// Delete DHCP binding
	err = createdDhcpBinding.Delete()
	check.Assert(err, IsNil)

	// Ensure the binding is removed
	bindingShouldBeNil, err := orgNet.GetOpenApiOrgVdcNetworkDhcpBindingById(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)
	check.Assert(err, NotNil)
	check.Assert(bindingShouldBeNil, IsNil)
}

func runOpenApiOrgVdcNetworkTestWithSegmentProfileTemplateEndpoint(check *C, vcd *TestVCD, vdc *Vdc, orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork, expectNetworkType string) {
	printVerbose("## Testing Segment Profile assignment in explicit Segment Profile endpoint\n")

	nsxtManager, err := vcd.client.GetNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(nsxtManager, NotNil)
	nsxtManagerUrn, err := nsxtManager.Urn()
	check.Assert(err, IsNil)

	// Filter by NSX-T Manager
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("nsxTManagerRef.id==%s", nsxtManagerUrn), queryParams)

	// Lookup prerequisite profiles for Segment Profile template creation
	ipDiscoveryProfile, err := vcd.client.GetIpDiscoveryProfileByName(vcd.config.VCD.Nsxt.IpDiscoveryProfile, queryParams)
	check.Assert(err, IsNil)
	macDiscoveryProfile, err := vcd.client.GetMacDiscoveryProfileByName(vcd.config.VCD.Nsxt.MacDiscoveryProfile, queryParams)
	check.Assert(err, IsNil)
	spoofGuardProfile, err := vcd.client.GetSpoofGuardProfileByName(vcd.config.VCD.Nsxt.SpoofGuardProfile, queryParams)
	check.Assert(err, IsNil)
	qosProfile, err := vcd.client.GetQoSProfileByName(vcd.config.VCD.Nsxt.QosProfile, queryParams)
	check.Assert(err, IsNil)
	segmentSecurityProfile, err := vcd.client.GetSegmentSecurityProfileByName(vcd.config.VCD.Nsxt.SegmentSecurityProfile, queryParams)
	check.Assert(err, IsNil)

	orgVdcNet, err := vdc.CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig)
	check.Assert(err, IsNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + orgVdcNet.OpenApiOrgVdcNetwork.ID
	AddToCleanupListOpenApi(orgVdcNet.OpenApiOrgVdcNetwork.Name, check.TestName(), openApiEndpoint)

	// Set segment profiles explicitly without using templates
	entitySegmentProfileCfg := &types.OrgVdcNetworkSegmentProfiles{
		IPDiscoveryProfile:     &types.Reference{ID: ipDiscoveryProfile.ID},
		MacDiscoveryProfile:    &types.Reference{ID: macDiscoveryProfile.ID},
		SpoofGuardProfile:      &types.Reference{ID: spoofGuardProfile.ID},
		QosProfile:             &types.Reference{ID: qosProfile.ID},
		SegmentSecurityProfile: &types.Reference{ID: segmentSecurityProfile.ID},
	}

	updatedSegmentProfiles, err := orgVdcNet.UpdateSegmentProfile(entitySegmentProfileCfg)
	check.Assert(err, IsNil)
	check.Assert(updatedSegmentProfiles, NotNil)

	check.Assert(updatedSegmentProfiles.IPDiscoveryProfile.ID, Equals, ipDiscoveryProfile.ID)
	check.Assert(updatedSegmentProfiles.MacDiscoveryProfile.ID, Equals, macDiscoveryProfile.ID)
	check.Assert(updatedSegmentProfiles.SpoofGuardProfile.ID, Equals, spoofGuardProfile.ID)
	check.Assert(updatedSegmentProfiles.QosProfile.ID, Equals, qosProfile.ID)
	check.Assert(updatedSegmentProfiles.SegmentSecurityProfile.ID, Equals, segmentSecurityProfile.ID)

	retrievedSegmentProfile, err := orgVdcNet.GetSegmentProfile()
	check.Assert(err, IsNil)
	check.Assert(retrievedSegmentProfile, NotNil)

	check.Assert(retrievedSegmentProfile.IPDiscoveryProfile.ID, Equals, ipDiscoveryProfile.ID)
	check.Assert(retrievedSegmentProfile.MacDiscoveryProfile.ID, Equals, macDiscoveryProfile.ID)
	check.Assert(retrievedSegmentProfile.SpoofGuardProfile.ID, Equals, spoofGuardProfile.ID)
	check.Assert(retrievedSegmentProfile.QosProfile.ID, Equals, qosProfile.ID)
	check.Assert(retrievedSegmentProfile.SegmentSecurityProfile.ID, Equals, segmentSecurityProfile.ID)

	// Delete
	err = orgVdcNet.Delete()
	check.Assert(err, IsNil)
}

func runOpenApiOrgVdcNetworkTestWithSegmentProfileTemplate(check *C, vcd *TestVCD, vdc *Vdc, orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork, expectNetworkType string) {
	printVerbose("## Testing Segment Profile Template assignment in Org VDC Network Structure\n")
	// Precreate two segment profile templates
	spt1 := preCreateSegmentProfileTemplate(vcd, check, "1")
	spt2 := preCreateSegmentProfileTemplate(vcd, check, "2")
	orgVdcNetworkConfig.SegmentProfileTemplate = &types.OpenApiReference{ID: spt1.NsxtSegmentProfileTemplate.ID}
	defer func() { // Cleanup Segment Profile Template configuration to prevent altering other tests
		orgVdcNetworkConfig.SegmentProfileTemplate = nil
	}()

	orgVdcNet, err := vdc.CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig)
	check.Assert(err, IsNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + orgVdcNet.OpenApiOrgVdcNetwork.ID
	AddToCleanupListOpenApi(orgVdcNet.OpenApiOrgVdcNetwork.Name, check.TestName(), openApiEndpoint)

	// Segment Profile Templates are not returned in GET by API definition which is not convenient,
	// but this check will work as a fuse to detect if anything changes in that regard in future
	check.Assert(orgVdcNet.OpenApiOrgVdcNetwork.SegmentProfileTemplate, IsNil)

	// Retrieve Segment Profile Template using its dedicated endpoint
	segmentProfileConfig, err := orgVdcNet.GetSegmentProfile()
	check.Assert(err, IsNil)
	check.Assert(segmentProfileConfig, NotNil)
	check.Assert(segmentProfileConfig.SegmentProfileTemplate.TemplateRef.ID, Equals, spt1.NsxtSegmentProfileTemplate.ID)

	// Update Segment Profile Template
	orgVdcNet.OpenApiOrgVdcNetwork.SegmentProfileTemplate = &types.OpenApiReference{ID: spt2.NsxtSegmentProfileTemplate.ID}

	updatedOrgVdcNet, err := orgVdcNet.Update(orgVdcNet.OpenApiOrgVdcNetwork)
	check.Assert(err, IsNil)

	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.Name, Equals, orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.Description, Equals, orgVdcNet.OpenApiOrgVdcNetwork.Description)

	// Segment Profile Templates are not returned in GET by API definition which is not convenient,
	// but this check will work as a fuse to detect if anything changes in that regard in future
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.SegmentProfileTemplate, IsNil)
	// Retrieve Segment Profile Template using its dedicated endpoint
	segmentProfileConfig, err = orgVdcNet.GetSegmentProfile()
	check.Assert(err, IsNil)
	check.Assert(segmentProfileConfig, NotNil)
	check.Assert(segmentProfileConfig.SegmentProfileTemplate.TemplateRef.ID, Equals, spt2.NsxtSegmentProfileTemplate.ID)

	// Delete
	err = orgVdcNet.Delete()
	check.Assert(err, IsNil)

	// Delete Segment Profile Templates

	err = spt1.Delete()
	check.Assert(err, IsNil)
	err = spt2.Delete()
	check.Assert(err, IsNil)
}

func preCreateSegmentProfileTemplate(vcd *TestVCD, check *C, sptNameSuffix string) *NsxtSegmentProfileTemplate {
	skipNoNsxtConfiguration(vcd, check)
	vcd.skipIfNotSysAdmin(check)

	nsxtManager, err := vcd.client.GetNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(nsxtManager, NotNil)
	nsxtManagerUrn, err := nsxtManager.Urn()
	check.Assert(err, IsNil)

	// Filter by NSX-T Manager
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("nsxTManagerRef.id==%s", nsxtManagerUrn), queryParams)

	// Lookup prerequisite profiles for Segment Profile template creation
	ipDiscoveryProfile, err := vcd.client.GetIpDiscoveryProfileByName(vcd.config.VCD.Nsxt.IpDiscoveryProfile, queryParams)
	check.Assert(err, IsNil)
	macDiscoveryProfile, err := vcd.client.GetMacDiscoveryProfileByName(vcd.config.VCD.Nsxt.MacDiscoveryProfile, queryParams)
	check.Assert(err, IsNil)
	spoofGuardProfile, err := vcd.client.GetSpoofGuardProfileByName(vcd.config.VCD.Nsxt.SpoofGuardProfile, queryParams)
	check.Assert(err, IsNil)
	qosProfile, err := vcd.client.GetQoSProfileByName(vcd.config.VCD.Nsxt.QosProfile, queryParams)
	check.Assert(err, IsNil)
	segmentSecurityProfile, err := vcd.client.GetSegmentSecurityProfileByName(vcd.config.VCD.Nsxt.SegmentSecurityProfile, queryParams)
	check.Assert(err, IsNil)

	config := &types.NsxtSegmentProfileTemplate{
		Name:                   check.TestName() + "-" + sptNameSuffix,
		Description:            check.TestName() + "-description",
		IPDiscoveryProfile:     &types.Reference{ID: ipDiscoveryProfile.ID},
		MacDiscoveryProfile:    &types.Reference{ID: macDiscoveryProfile.ID},
		QosProfile:             &types.Reference{ID: qosProfile.ID},
		SegmentSecurityProfile: &types.Reference{ID: segmentSecurityProfile.ID},
		SpoofGuardProfile:      &types.Reference{ID: spoofGuardProfile.ID},
		SourceNsxTManagerRef:   &types.OpenApiReference{ID: nsxtManager.NsxtManager.ID},
	}

	createdSegmentProfileTemplate, err := vcd.client.CreateSegmentProfileTemplate(config)
	check.Assert(err, IsNil)
	check.Assert(createdSegmentProfileTemplate, NotNil)

	// Add to cleanup list
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates + createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID
	AddToCleanupListOpenApi(config.Name, check.TestName(), openApiEndpoint)

	return createdSegmentProfileTemplate
}
