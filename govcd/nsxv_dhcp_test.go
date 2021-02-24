// +build nsxv functional ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxvDhcp does the following:
// 1. pre-creates Org VDC routed network using Openapi
// 2. Creates 3 DHCP pools in the scope of this Org VDC Routed network by using UpdateDhcpPoolsAndBindings()
// 3. Updates 3 DHCP pools  in the scope of this Org VDC Routed network by using UpdateDhcpPools()
// 4. Removes DHCP pools
func (vcd *TestVCD) Test_NsxvDhcp(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}

	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	// Pre-create Org VDC routed network using OpenAPI to try and attach DHCP pools to it in this test
	orgVdcNet := createOrgVdcRoutedNet(check, edge.EdgeGateway.ID, vcd.vdc)
	defer func() {
		orgVdcNet.Delete()
	}()

	// Retrieve any binding to store them internally
	currentDhcpPool, err := edge.GetDhcpPoolsAndBindings()
	check.Assert(err, IsNil)

	poolDef1 := types.EdgeDhcpIpPool{
		AutoConfigureDNS:    false,
		DefaultGateway:      "22.1.1.1",
		DomainName:          "asd.hostname",
		LeaseTime:           "2000",
		SubnetMask:          "255.255.255.0",
		IpRange:             "22.1.1.242-22.1.1.243",
		PrimaryNameServer:   "8.8.8.8",
		SecondaryNameServer: "8.8.4.4",
	}

	poolDef2 := types.EdgeDhcpIpPool{
		AutoConfigureDNS: true,
		DefaultGateway:   "22.1.1.1",
		LeaseTime:        "infinite",
		SubnetMask:       "255.255.255.0",
		IpRange:          "22.1.1.244-22.1.1.245",
	}

	poolDef3 := types.EdgeDhcpIpPool{
		AutoConfigureDNS: false,
		DefaultGateway:   "22.1.1.1",
		SubnetMask:       "255.255.255.0",
		IpRange:          "22.1.1.200-22.1.1.201",
	}

	dhcpPoolConfig := &types.EdgeDhcp{
		Enabled: true,
		// Feed in the same static bindings to not override existing customer settings
		StaticBindings: currentDhcpPool.StaticBindings,
		EdgeDhcpIpPools: &types.EdgeDhcpIpPools{
			EdgeDhcpIpPool: []types.EdgeDhcpIpPool{
				poolDef1,
				poolDef2,
				poolDef3,
			},
		},
	}

	returnedDhcpPool, err := edge.UpdateDhcpPoolsAndBindings(dhcpPoolConfig)
	check.Assert(err, IsNil)

	check.Assert(len(returnedDhcpPool.EdgeDhcpIpPools.EdgeDhcpIpPool) > 0, Equals, true)
	check.Assert(returnedDhcpPool.Enabled, Equals, true)
	check.Assert(findDhcpPoolByIpRange(returnedDhcpPool.EdgeDhcpIpPools.EdgeDhcpIpPool, poolDef1.IpRange), Equals, poolDef1)
	check.Assert(findDhcpPoolByIpRange(returnedDhcpPool.EdgeDhcpIpPools.EdgeDhcpIpPool, poolDef2.IpRange), Equals, poolDef2)

	// poolDef3 definition did not specify lease time, but VCD returns 864000 by default therefore it must be injected
	// before comparison
	poolDef3.LeaseTime = "86400"
	check.Assert(findDhcpPoolByIpRange(returnedDhcpPool.EdgeDhcpIpPools.EdgeDhcpIpPool, poolDef3.IpRange), Equals, poolDef3)

	// Update by using only UpdateDhcpPools and see if the settings are the same
	returnedDhcpPool2, err := edge.UpdateDhcpPools(dhcpPoolConfig)
	check.Assert(err, IsNil)

	check.Assert(len(returnedDhcpPool2.EdgeDhcpIpPools.EdgeDhcpIpPool) > 0, Equals, true)
	check.Assert(returnedDhcpPool.Enabled, Equals, true)
	check.Assert(findDhcpPoolByIpRange(returnedDhcpPool2.EdgeDhcpIpPools.EdgeDhcpIpPool, poolDef1.IpRange), Equals, poolDef1)
	check.Assert(findDhcpPoolByIpRange(returnedDhcpPool2.EdgeDhcpIpPools.EdgeDhcpIpPool, poolDef2.IpRange), Equals, poolDef2)

	err = edge.ResetDhcpPools()
	check.Assert(err, IsNil)
}

// findDhcpPoolByIpRange helps to find particular DHCP pool in unordered list by ipRange (which does not allow
// duplicates)
func findDhcpPoolByIpRange(sliceOfPools []types.EdgeDhcpIpPool, ipRange string) types.EdgeDhcpIpPool {
	for index := range sliceOfPools {
		if sliceOfPools[index].IpRange == ipRange {
			return sliceOfPools[index]
		}
	}
	return types.EdgeDhcpIpPool{}
}

func createOrgVdcRoutedNet(check *C, edgeGatewayId string, vdc *Vdc) *OpenApiOrgVdcNetwork {
	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		OrgVdc:      &types.OpenApiReference{ID: vdc.Vdc.ID},

		NetworkType: types.OrgVdcNetworkTypeRouted,

		// Connection is used for "routed" network
		Connection: &types.Connection{
			RouterRef: types.OpenApiReference{
				ID: edgeGatewayId,
			},
			ConnectionType: "INTERNAL",
		},
		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "22.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "foo.bar",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "22.1.1.20",
								EndAddress:   "22.1.1.30",
							},
						}},
				},
			},
		},
	}

	orgVdcNet, err := vdc.CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig)
	check.Assert(err, IsNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + orgVdcNet.OpenApiOrgVdcNetwork.ID
	AddToCleanupListOpenApi(orgVdcNet.OpenApiOrgVdcNetwork.Name, check.TestName(), openApiEndpoint)

	return orgVdcNet
}
