//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxEdgeBgpConfiguration(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeBgpConfig)

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

	// Get and store existing BGP configuration
	bgpConfig, err := edge.GetBgpConfiguration()
	check.Assert(err, IsNil)
	check.Assert(bgpConfig, NotNil)

	// Disable BGP
	err = edge.DisableBgpConfiguration()
	check.Assert(err, IsNil)

	newBgpConfig := &types.EdgeBgpConfig{
		Enabled:       true,
		Ecmp:          true,
		LocalASNumber: "65420",
		GracefulRestart: &types.EdgeBgpGracefulRestartConfig{
			StaleRouteTimer: 190,
			RestartTimer:    600,
			Mode:            "HELPER_ONLY",
		},
	}

	updatedBgpConfig, err := edge.UpdateBgpConfiguration(newBgpConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedBgpConfig, NotNil)

	newBgpConfig.Version = updatedBgpConfig.Version // Version is constantly iterated and cant be checked
	check.Assert(updatedBgpConfig, DeepEquals, newBgpConfig)

	// Check "disable" function which keeps all fields the same, but sets "Enabled: false"
	err = edge.DisableBgpConfiguration()
	check.Assert(err, IsNil)

	bgpConfigAfterDisabling, err := edge.GetBgpConfiguration()
	check.Assert(err, IsNil)
	check.Assert(bgpConfig, NotNil)
	check.Assert(bgpConfigAfterDisabling.Enabled, Equals, false)

}

func switchEdgeGatewayDedication(edge *NsxtEdgeGateway, isDedicated bool) error {
	edge.EdgeGateway.EdgeGatewayUplinks[0].Dedicated = isDedicated
	_, err := edge.Update(edge.EdgeGateway)

	return err
}
