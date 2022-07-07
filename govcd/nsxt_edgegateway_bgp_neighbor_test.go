//go:build network || nsxt || functional || openapi || ALL
// +build network nsxt functional openapi ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxEdgeBgpNeighbor(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeBgpNeighbor)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Switch Edge Gateway to use dedicated uplink for the time of this test and then turn it off
	err = switchEdgeGatewayDedication(edge, true) // Turn on Dedicated Tier 0 gateway
	check.Assert(err, IsNil)
	defer switchEdgeGatewayDedication(edge, false) // Turn off Dedicated Tier 0 gateway

	// Create a new BGP IP Prefix List
	BgpNeighbor := &types.EdgeBgpNeighbor{
		NeighborAddress:        "11.11.11.11",
		RemoteASNumber:         "64123",
		KeepAliveTimer:         80,
		HoldDownTimer:          241,
		NeighborPassword:       "iQuee-ph2phe",
		AllowASIn:              true,
		GracefulRestartMode:    "HELPER_ONLY",
		IpAddressTypeFiltering: "DISABLED",
	}

	bgpIpPrefix, err := edge.CreateBgpNeighbor(BgpNeighbor)
	check.Assert(err, IsNil)
	check.Assert(bgpIpPrefix, NotNil)

	// Get all BGP IP Prefix Lists
	BgpNeighbors, err := edge.GetAllBgpNeighbors(nil)
	check.Assert(err, IsNil)
	check.Assert(BgpNeighbors, NotNil)
	check.Assert(len(BgpNeighbors), Equals, 1)
	check.Assert(BgpNeighbors[0].EdgeBgpNeighbor.NeighborAddress, Equals, BgpNeighbor.NeighborAddress)

	// Get By Neighbor IP Address
	bgpPrefixListByName, err := edge.GetBgpNeighborByIp(BgpNeighbor.NeighborAddress)
	check.Assert(err, IsNil)
	check.Assert(bgpPrefixListByName, NotNil)

	// Get By Id
	bgpPrefixListById, err := edge.GetBgpNeighborById(bgpIpPrefix.EdgeBgpNeighbor.ID)
	check.Assert(err, IsNil)
	check.Assert(bgpPrefixListById, NotNil)

	// Update
	BgpNeighbor.NeighborAddress = "12.12.12.12"
	BgpNeighbor.ID = BgpNeighbors[0].EdgeBgpNeighbor.ID

	updatedBgpNeighbor, err := BgpNeighbors[0].Update(BgpNeighbor)
	check.Assert(err, IsNil)
	check.Assert(updatedBgpNeighbor, NotNil)

	check.Assert(updatedBgpNeighbor.EdgeBgpNeighbor.ID, Equals, BgpNeighbors[0].EdgeBgpNeighbor.ID)

	// Delete
	err = BgpNeighbors[0].Delete()
	check.Assert(err, IsNil)

	// Try to get once again and ensure it is not there
	notFoundByName, err := edge.GetBgpNeighborByIp(BgpNeighbor.NeighborAddress)
	check.Assert(err, NotNil)
	check.Assert(notFoundByName, IsNil)

	notFoundById, err := edge.GetBgpNeighborById(bgpIpPrefix.EdgeBgpNeighbor.ID)
	check.Assert(err, NotNil)
	check.Assert(notFoundById, IsNil)

}

func switchEdgeGatewayDedication(edge *NsxtEdgeGateway, isDedicated bool) error {
	edge.EdgeGateway.EdgeGatewayUplinks[0].Dedicated = isDedicated
	_, err := edge.Update(edge.EdgeGateway)

	return err
}
