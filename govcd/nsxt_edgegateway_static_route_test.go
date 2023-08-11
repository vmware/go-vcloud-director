//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxEdgeStaticRoute(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayStaticRoutes)
	vcd.skipIfNotSysAdmin(check)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Switch Edge Gateway to use dedicated uplink for the time of this test and then turn it off
	err = switchEdgeGatewayDedication(edge, true) // Turn on Dedicated Tier 0 gateway
	check.Assert(err, IsNil)
	defer func() {
		err = switchEdgeGatewayDedication(edge, false)
		check.Assert(err, IsNil)
	}()

	// Get Org VDC routed network
	orgVdcNet, err := nsxtVdc.GetOpenApiOrgVdcNetworkByName(vcd.config.VCD.Nsxt.RoutedNetwork)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNet, NotNil)

	staticRouteConfig := &types.NsxtEdgeGatewayStaticRoute{
		Name:        check.TestName(),
		Description: "description",
		NetworkCidr: "1.1.1.0/24",
		NextHops: []types.NsxtEdgeGatewayStaticRouteNextHops{
			{
				IPAddress:     orgVdcNet.OpenApiOrgVdcNetwork.Subnets.Values[0].Gateway,
				AdminDistance: 4,
				Scope: &types.NsxtEdgeGatewayStaticRouteNextHopScope{
					ID:        orgVdcNet.OpenApiOrgVdcNetwork.ID,
					ScopeType: "NETWORK",
				},
			},
		},
	}

	staticRoute, err := edge.CreateStaticRoute(staticRouteConfig)
	check.Assert(err, IsNil)
	check.Assert(staticRoute, NotNil)

	// Get all BGP IP Prefix Lists
	staticRouteList, err := edge.GetAllStaticRoutes(nil)
	check.Assert(err, IsNil)
	check.Assert(staticRouteList, NotNil)
	check.Assert(len(staticRouteList), Equals, 1)
	check.Assert(staticRouteList[0].NsxtEdgeGatewayStaticRoute.Name, Equals, staticRoute.NsxtEdgeGatewayStaticRoute.Name)

	// Get By Name
	staticRouteByName, err := edge.GetStaticRouteByName(staticRoute.NsxtEdgeGatewayStaticRoute.Name)
	check.Assert(err, IsNil)
	check.Assert(staticRouteByName, NotNil)

	// Get By Id
	staticRouteById, err := edge.GetStaticRouteById(staticRoute.NsxtEdgeGatewayStaticRoute.ID)
	check.Assert(err, IsNil)
	check.Assert(staticRouteById, NotNil)

	// Get By Network CIDR
	staticRouteByNetworkCidr, err := edge.GetStaticRouteByNetworkCidr(staticRoute.NsxtEdgeGatewayStaticRoute.NetworkCidr)
	check.Assert(err, IsNil)
	check.Assert(staticRouteByNetworkCidr, NotNil)

	// Update
	staticRouteConfig.Name = check.TestName() + "-updated"
	staticRouteConfig.Description = "test-description-updated"
	staticRouteConfig.ID = staticRouteByNetworkCidr.NsxtEdgeGatewayStaticRoute.ID

	// staticRoute
	updatedStaticRoute, err := staticRoute.Update(staticRouteConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedStaticRoute, NotNil)

	check.Assert(updatedStaticRoute.NsxtEdgeGatewayStaticRoute.ID, Equals, staticRouteByName.NsxtEdgeGatewayStaticRoute.ID)

	// Delete
	err = staticRoute.Delete()
	check.Assert(err, IsNil)

	// Try to get once again and ensure it is not there
	notFoundByName, err := edge.GetStaticRouteByName(staticRoute.NsxtEdgeGatewayStaticRoute.Name)
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)

	notFoundById, err := edge.GetStaticRouteById(staticRoute.NsxtEdgeGatewayStaticRoute.ID)
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundById, IsNil)

	notFoundByCidr, err := edge.GetStaticRouteByNetworkCidr(staticRoute.NsxtEdgeGatewayStaticRoute.NetworkCidr)
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByCidr, IsNil)
}
