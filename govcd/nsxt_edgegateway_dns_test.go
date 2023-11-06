//go:build ALL || openapi || functional || nsxt

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtEdgeGatewayDns(check *C) {
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayDns)
	skipNoNsxtConfiguration(vcd, check)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)
	AddToCleanupList(vcd.config.VCD.Nsxt.EdgeGateway, "nsxtEdgeGatewayDns", vcd.config.VCD.Org, check.TestName())

	disabledDns, err := edge.GetNsxtEdgeGatewayDns()
	check.Assert(err, IsNil)
	check.Assert(disabledDns.NsxtEdgeGatewayDns.Enabled, Equals, false)

	enabledDnsConfig := &types.NsxtEdgeGatewayDns{
		Enabled: true,
		DefaultForwarderZone: &types.NsxtDnsForwarderZoneConfig{
			DisplayName: "test",
			UpstreamServers: []string{
				"1.2.3.4",
				"2.3.4.5",
			},
		},
		ConditionalForwarderZones: []*types.NsxtDnsForwarderZoneConfig{
			{
				DisplayName: "test-conditional",
				UpstreamServers: []string{
					"5.5.5.5",
					"2.3.4.1",
				},
				DnsDomainNames: []string{
					"test.com",
					"abc.com",
					"example.org",
				},
			},
		},
	}

	enabledDns, err := disabledDns.Update(enabledDnsConfig)
	enabledDnsConfig = enabledDns.NsxtEdgeGatewayDns
	check.Assert(err, IsNil)
	check.Assert(enabledDnsConfig.Enabled, Equals, true)
	check.Assert(enabledDnsConfig.DefaultForwarderZone.DisplayName, Equals, "test")
	check.Assert(len(enabledDnsConfig.DefaultForwarderZone.UpstreamServers), Equals, 2)
	check.Assert(len(enabledDnsConfig.ConditionalForwarderZones), Equals, 1)
	check.Assert(enabledDnsConfig.ConditionalForwarderZones[0].DisplayName, Equals, "test-conditional")

	err = enabledDns.Delete()
	check.Assert(err, IsNil)

	deletedDns, err := edge.GetNsxtEdgeGatewayDns()
	check.Assert(err, IsNil)
	check.Assert(deletedDns.NsxtEdgeGatewayDns.Enabled, Equals, false)
	extNet := createExternalNetwork(vcd, check)
	ipSpacesEnabledEgw := createNsxtEdgeGateway(vcd, check, extNet.ExternalNetwork.ID)
	ipSpacesDns, err := ipSpacesEnabledEgw.GetNsxtEdgeGatewayDns()
	check.Assert(err, IsNil)
	check.Assert(ipSpacesDns.NsxtEdgeGatewayDns, Equals, true)

}
