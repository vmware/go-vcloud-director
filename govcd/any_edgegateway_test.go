//go:build network || nsxt || functional || openapi || ALL
// +build network nsxt functional openapi ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_AnyEdgeGateway(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtEdge, err := adminOrg.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	nsxvEdge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)

	// Retrieve both types of Edge Gateways using adminOrg structure (NSX-T and NSX-V) using the
	// common type AnyEdgeGateway
	nsxtAnyEdgeGateway, err := adminOrg.GetAnyEdgeGatewayById(nsxtEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(nsxtAnyEdgeGateway.EdgeGateway, DeepEquals, nsxtEdge.EdgeGateway)

	nsxvAnyEdgeGateway, err := adminOrg.GetAnyEdgeGatewayById(nsxvEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)

	// Structures for NSX-V Edge Gateway differ (because it uses XML API) therefore all fields
	// cannot be compared
	check.Assert(nsxvAnyEdgeGateway.EdgeGateway.ID, DeepEquals, nsxvEdge.EdgeGateway.ID)

	// Retrieve both types of Edge Gateways using Org structure (NSX-T and NSX-V) using the
	// common type AnyEdgeGateway
	nsxtOrgAnyEdgeGateway, err := org.GetAnyEdgeGatewayById(nsxtEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(nsxtOrgAnyEdgeGateway.EdgeGateway, DeepEquals, nsxtEdge.EdgeGateway)

	// Convert NSX-T backed AnyEdgeGateway to NsxtEdgeGateway
	convertedGw, err := nsxtOrgAnyEdgeGateway.GetNsxtEdgeGateway()
	check.Assert(err, IsNil)
	check.Assert(convertedGw.EdgeGateway, DeepEquals, nsxtOrgAnyEdgeGateway.EdgeGateway)

	// Structures for NSX-V Edge Gateway differ (because it uses XML API) therefore all fields
	// cannot be compared
	nsxvOrgAnyEdgeGateway, err := org.GetAnyEdgeGatewayById(nsxvEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(nsxvOrgAnyEdgeGateway.EdgeGateway.ID, DeepEquals, nsxvEdge.EdgeGateway.ID)

}
