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

	disabledDns, err := edge.GetDnsConfig()
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
	check.Assert(err, IsNil)
	dnsConfig := enabledDns.NsxtEdgeGatewayDns
	check.Assert(dnsConfig.Enabled, Equals, true)
	check.Assert(dnsConfig.DefaultForwarderZone.DisplayName, Equals, "test")
	check.Assert(len(dnsConfig.DefaultForwarderZone.UpstreamServers), Equals, 2)
	check.Assert(len(dnsConfig.ConditionalForwarderZones), Equals, 1)
	check.Assert(dnsConfig.ConditionalForwarderZones[0].DisplayName, Equals, "test-conditional")

	updatedDnsConfig := &types.NsxtEdgeGatewayDns{
		Enabled: true,
		DefaultForwarderZone: &types.NsxtDnsForwarderZoneConfig{
			DisplayName: "test",
			UpstreamServers: []string{
				"1.2.3.5",
				"2.3.4.6",
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
			{
				DisplayName: "test-conditional-2",
				UpstreamServers: []string{
					"1.2.3.4",
					"4.3.2.1",
				},
				DnsDomainNames: []string{
					"example.com",
				},
			},
		},
	}
	updatedDns, err := enabledDns.Update(updatedDnsConfig)
	updatedDnsConfig = updatedDns.NsxtEdgeGatewayDns
	check.Assert(err, IsNil)
	check.Assert(updatedDnsConfig.Enabled, Equals, true)
	check.Assert(updatedDnsConfig.DefaultForwarderZone.DisplayName, Equals, "test")
	check.Assert(len(updatedDnsConfig.DefaultForwarderZone.UpstreamServers), Equals, 3)
	conditionalZones := updatedDnsConfig.ConditionalForwarderZones
	check.Assert(len(conditionalZones), Equals, 2)
	// Flip the asserts in both cases of conditional zones arrays returned
	if conditionalZones[0].DisplayName == "test-conditional" {
		check.Assert(len(conditionalZones[0].UpstreamServers), Equals, 2)
		check.Assert(len(conditionalZones[0].DnsDomainNames), Equals, 3)
		check.Assert(len(conditionalZones[1].UpstreamServers), Equals, 2)
		check.Assert(len(conditionalZones[1].DnsDomainNames), Equals, 1)
	} else {
		check.Assert(len(conditionalZones[1].UpstreamServers), Equals, 2)
		check.Assert(len(conditionalZones[1].DnsDomainNames), Equals, 3)
		check.Assert(len(conditionalZones[0].UpstreamServers), Equals, 2)
		check.Assert(len(conditionalZones[0].DnsDomainNames), Equals, 1)
	}

	err = enabledDns.Delete()
	check.Assert(err, IsNil)

	deletedDns, err := edge.GetDnsConfig()
	check.Assert(err, IsNil)
	check.Assert(deletedDns.NsxtEdgeGatewayDns.Enabled, Equals, false)
}
