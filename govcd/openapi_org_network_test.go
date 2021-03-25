// +build network nsxt functional openapi ALL

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
	skipOpenApiEndpointTest(ctx, vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	skipNoNsxtConfiguration(vcd, check)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OrgVdc:      &types.OpenApiReference{ID: vcd.nsxtVdc.Vdc.ID},

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

	runOpenApiOrgVdcNetworkTest(check, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeIsolated, nil)
}

func (vcd *TestVCD) Test_NsxtOrgVdcNetworkRouted(check *C) {
	skipOpenApiEndpointTest(ctx, vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	skipNoNsxtConfiguration(vcd, check)

	egw, err := vcd.org.GetNsxtEdgeGatewayByName(ctx, vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OrgVdc:      &types.OpenApiReference{ID: vcd.nsxtVdc.Vdc.ID},

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

	runOpenApiOrgVdcNetworkTest(check, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeRouted, nsxtRoutedDhcpConfig)
}

func (vcd *TestVCD) Test_NsxtOrgVdcNetworkImported(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipOpenApiEndpointTest(ctx, vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	skipNoNsxtConfiguration(vcd, check)

	if vcd.config.VCD.Nsxt.NsxtImportSegment == "" {
		check.Skip("Unused NSX-T segment was not provided")
	}

	logicalSwitch, err := vcd.nsxtVdc.GetNsxtImportableSwitchByName(ctx, vcd.config.VCD.Nsxt.NsxtImportSegment)
	check.Assert(err, IsNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OrgVdc:      &types.OpenApiReference{ID: vcd.nsxtVdc.Vdc.ID},

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

	runOpenApiOrgVdcNetworkTest(check, vcd.nsxtVdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeOpaque, nil)

}

func (vcd *TestVCD) Test_NsxvOrgVdcNetworkIsolated(check *C) {
	skipOpenApiEndpointTest(ctx, vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OrgVdc:      &types.OpenApiReference{ID: vcd.vdc.Vdc.ID},

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

	runOpenApiOrgVdcNetworkTest(check, vcd.vdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeIsolated, nil)
}

func (vcd *TestVCD) Test_NsxvOrgVdcNetworkRouted(check *C) {
	skipOpenApiEndpointTest(ctx, vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)

	nsxvEdgeGateway, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, true)
	check.Assert(err, IsNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OrgVdc:      &types.OpenApiReference{ID: vcd.vdc.Vdc.ID},

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

	runOpenApiOrgVdcNetworkTest(check, vcd.vdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeRouted, nil)
}

func (vcd *TestVCD) Test_NsxvOrgVdcNetworkDirect(check *C) {
	skipOpenApiEndpointTest(ctx, vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgVdcNetworks)
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	// Using legacy API lookup function because GetExternalNetworkV2ByName does not support VCD 9.7
	externalNetwork, err := vcd.client.GetExternalNetworkByName(ctx, vcd.config.VCD.ExternalNetwork)

	check.Assert(err, IsNil)

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OrgVdc:      &types.OpenApiReference{ID: vcd.vdc.Vdc.ID},

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

	runOpenApiOrgVdcNetworkTest(check, vcd.vdc, orgVdcNetworkConfig, types.OrgVdcNetworkTypeDirect, nil)
}

func runOpenApiOrgVdcNetworkTest(check *C, vdc *Vdc, orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork, extpectNetworkType string, dhcpFunc dhcpConfigFunc) {
	orgVdcNet, err := vdc.CreateOpenApiOrgVdcNetwork(ctx, orgVdcNetworkConfig)
	check.Assert(err, IsNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + orgVdcNet.OpenApiOrgVdcNetwork.ID
	AddToCleanupListOpenApi(orgVdcNet.OpenApiOrgVdcNetwork.Name, check.TestName(), openApiEndpoint)

	check.Assert(orgVdcNet.GetType(), Equals, extpectNetworkType)

	// Check it can be found
	orgVdcNetByIdInVdc, err := vdc.GetOpenApiOrgVdcNetworkById(ctx, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(err, IsNil)
	orgVdcNetByName, err := vdc.GetOpenApiOrgVdcNetworkByName(ctx, orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(err, IsNil)

	check.Assert(orgVdcNetByIdInVdc.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(orgVdcNetByName.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)

	// Retrieve all networks in VDC and expect newly created network to be there
	var foundNetInVdc bool
	allOrgVdcNets, err := vdc.GetAllOpenApiOrgVdcNetworks(ctx, nil)
	check.Assert(err, IsNil)
	for _, net := range allOrgVdcNets {
		if net.OpenApiOrgVdcNetwork.ID == orgVdcNet.OpenApiOrgVdcNetwork.ID {
			foundNetInVdc = true
		}
	}
	check.Assert(foundNetInVdc, Equals, true)

	// Update
	orgVdcNet.OpenApiOrgVdcNetwork.Description = check.TestName() + "updated description"
	updatedOrgVdcNet, err := orgVdcNet.Update(ctx, orgVdcNet.OpenApiOrgVdcNetwork)
	check.Assert(err, IsNil)

	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.Name, Equals, orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.ID, Equals, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(updatedOrgVdcNet.OpenApiOrgVdcNetwork.Description, Equals, orgVdcNet.OpenApiOrgVdcNetwork.Description)

	// Configure DHCP if specified
	if dhcpFunc != nil {
		dhcpFunc(check, vdc, updatedOrgVdcNet.OpenApiOrgVdcNetwork.ID)
	}
	// Delete
	err = orgVdcNet.Delete(ctx)
	check.Assert(err, IsNil)

	// Test again if it was deleted and expect it to contain ErrorEntityNotFound
	_, err = vdc.GetOpenApiOrgVdcNetworkByName(ctx, orgVdcNet.OpenApiOrgVdcNetwork.Name)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vdc.GetOpenApiOrgVdcNetworkById(ctx, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
}

type dhcpConfigFunc func(check *C, vdc *Vdc, orgNetId string)

func nsxtRoutedDhcpConfig(check *C, vdc *Vdc, orgNetId string) {
	dhcpDefinition := &types.OpenApiOrgVdcNetworkDhcp{
		Enabled: takeBoolPointer(true),
		DhcpPools: []types.OpenApiOrgVdcNetworkDhcpPools{
			{
				Enabled: takeBoolPointer(true),
				IPRange: types.OpenApiOrgVdcNetworkDhcpIpRange{
					StartAddress: "2.1.1.200",
					EndAddress:   "2.1.1.201",
				},
			},
		},
	}
	updatedDhcp, err := vdc.UpdateOpenApiOrgVdcNetworkDhcp(ctx, orgNetId, dhcpDefinition)
	check.Assert(err, IsNil)

	check.Assert(dhcpDefinition, DeepEquals, updatedDhcp.OpenApiOrgVdcNetworkDhcp)

	// VCD Versions before 10.2 do not allow to perform "DELETE" on DHCP pool
	// To remove DHCP configuration one must remove Org VDC network itself.
	if vdc.client.APIVCDMaxVersionIs(ctx, ">= 35.0") {
		err = vdc.DeleteOpenApiOrgVdcNetworkDhcp(ctx, orgNetId)
		check.Assert(err, IsNil)
	}
}
