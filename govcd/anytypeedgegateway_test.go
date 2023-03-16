//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_AnyTypeEdgeGateway(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	nsxtEdge, err := adminOrg.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(nsxtEdge, NotNil)

	nsxvEdge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(nsxvEdge, NotNil)

	// Retrieve both types of Edge Gateways using adminOrg structure (NSX-T and NSX-V) using the
	// common type AnyTypeEdgeGateway
	nsxtAnyTypeEdgeGateway, err := adminOrg.GetAnyTypeEdgeGatewayById(nsxtEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(nsxtAnyTypeEdgeGateway, NotNil)
	check.Assert(nsxtAnyTypeEdgeGateway.EdgeGateway, DeepEquals, nsxtEdge.EdgeGateway)

	nsxvAnyTypeEdgeGateway, err := adminOrg.GetAnyTypeEdgeGatewayById(nsxvEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(nsxvAnyTypeEdgeGateway, NotNil)

	// Structures for NSX-V Edge Gateway differ (because it uses XML API) therefore all fields
	// cannot be compared
	check.Assert(nsxvAnyTypeEdgeGateway.EdgeGateway.ID, DeepEquals, nsxvEdge.EdgeGateway.ID)

	// Retrieve both types of Edge Gateways using Org structure (NSX-T and NSX-V) using the
	// common type AnyTypeEdgeGateway
	nsxtOrgAnyTypeEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(nsxtEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(nsxtOrgAnyTypeEdgeGateway, NotNil)
	check.Assert(nsxtOrgAnyTypeEdgeGateway.EdgeGateway, DeepEquals, nsxtEdge.EdgeGateway)

	// Convert NSX-T backed AnyTypeEdgeGateway to NsxtEdgeGateway
	convertedGw, err := nsxtOrgAnyTypeEdgeGateway.GetNsxtEdgeGateway()
	check.Assert(err, IsNil)
	check.Assert(convertedGw, NotNil)
	check.Assert(convertedGw.EdgeGateway, DeepEquals, nsxtOrgAnyTypeEdgeGateway.EdgeGateway)

	// Structures for NSX-V Edge Gateway differ (because it uses XML API) therefore all fields
	// cannot be compared
	nsxvOrgAnyTypeEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(nsxvEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(nsxvOrgAnyTypeEdgeGateway, NotNil)
	check.Assert(nsxvOrgAnyTypeEdgeGateway.EdgeGateway.ID, DeepEquals, nsxvEdge.EdgeGateway.ID)

}
