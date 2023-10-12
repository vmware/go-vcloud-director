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

	dnsConfig, err := edge.GetNsxtEdgeGatewayDns()
	check.Assert(err, IsNil)
	check.Assert(dnsConfig.NsxtEdgeGatewayDns.Enabled, Equals, false)
	check.Assert(dnsConfig.NsxtEdgeGatewayDns.Version.Version, Equals, 0)
}
