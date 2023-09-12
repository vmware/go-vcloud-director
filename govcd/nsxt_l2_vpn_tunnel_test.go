//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_L2VpnTunnels(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayL2VpnTunnel)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	_, err = edge.GetAllL2VpnTunnels(nil)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_CreateL2VpnTunnel(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayL2VpnTunnel)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	network, err := nsxtVdc.GetOrgVdcNetworkByName(vcd.config.VCD.Nsxt.RoutedNetwork, false)
	check.Assert(err, IsNil)

	// Get the auto-allocated IP of the Edge Gateway
	localEndpointIp, err := edge.GetUsedIpAddresses(nil)
	check.Assert(err, IsNil)

	vpnTunnel := &types.NsxtL2VpnTunnel{
		Name:                    check.TestName(),
		Description:             check.TestName(),
		SessionMode:             "SERVER",
		Enabled:                 true,
		LocalEndpointIp:         localEndpointIp[0].IPAddress,
		RemoteEndpointIp:        "1.1.1.1",
		TunnelInterface:         "",
		ConnectorInitiationMode: "ON_DEMAND",
		PreSharedKey:            check.TestName(),
		StretchedNetworks: []types.EdgeL2VpnStretchedNetwork{
			{
				NetworkRef: types.OpenApiReference{
					Name: network.OrgVDCNetwork.Name,
					ID:   network.OrgVDCNetwork.ID,
				},
			},
		},
		Logging: false,
	}

	tunnel, err := edge.CreateL2VpnTunnel(vpnTunnel)
	check.Assert(err, IsNil)
	check.Assert(tunnel, NotNil)

	check.Assert(tunnel.NsxtL2VpnTunnel.Name, Equals, check.TestName())
	check.Assert(tunnel.NsxtL2VpnTunnel.Description, Equals, check.TestName())
	check.Assert(tunnel.NsxtL2VpnTunnel.SessionMode, Equals, "SERVER")
	check.Assert(tunnel.NsxtL2VpnTunnel.Enabled, Equals, true)
	check.Assert(tunnel.NsxtL2VpnTunnel.LocalEndpointIp, Equals, localEndpointIp[0].IPAddress)
	check.Assert(tunnel.NsxtL2VpnTunnel.RemoteEndpointIp, Equals, "1.1.1.1")
	check.Assert(tunnel.NsxtL2VpnTunnel.ConnectorInitiationMode, Equals, "ON_DEMAND")
	check.Assert(tunnel.NsxtL2VpnTunnel.PreSharedKey, Equals, check.TestName())

	err = tunnel.Delete()
	check.Assert(err, IsNil)
}
