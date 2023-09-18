//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"
	"regexp"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxtL2VpnTunnel tests NSX-T Edge Gateway L2 VPN Tunnels.
// 1. It creates/gets/updates/deletes a SERVER type Tunnel, also the peer code (encoded configuration of the tunnel)
// is saved for creation of CLIENT type tunnel.
// 2. Creates/gets/updates/deletes a CLIENT type Tunnel
func (vcd *TestVCD) Test_NsxtL2VpnTunnel(check *C) {
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

	// SERVER Tunnel part
	serverTunnelParams := &types.NsxtL2VpnTunnel{
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

	serverTunnel, err := edge.CreateL2VpnTunnel(serverTunnelParams)
	check.Assert(err, IsNil)
	check.Assert(serverTunnel, NotNil)
	AddToCleanupListOpenApi(serverTunnel.NsxtL2VpnTunnel.ID, check.TestName(),
		fmt.Sprintf(types.OpenApiPathVersion1_0_0+
			types.OpenApiEndpointEdgeGatewayL2VpnTunnel+
			serverTunnel.NsxtL2VpnTunnel.ID, edge.EdgeGateway.ID))

	// Save the peer code to create a Client tunnel for testing
	peerCode := serverTunnel.NsxtL2VpnTunnel.PeerCode

	check.Assert(serverTunnel.NsxtL2VpnTunnel.Name, Equals, check.TestName())
	check.Assert(serverTunnel.NsxtL2VpnTunnel.Description, Equals, check.TestName())
	check.Assert(serverTunnel.NsxtL2VpnTunnel.SessionMode, Equals, "SERVER")
	check.Assert(serverTunnel.NsxtL2VpnTunnel.Enabled, Equals, true)
	check.Assert(serverTunnel.NsxtL2VpnTunnel.LocalEndpointIp, Equals, localEndpointIp[0].IPAddress)
	check.Assert(serverTunnel.NsxtL2VpnTunnel.RemoteEndpointIp, Equals, "1.1.1.1")
	check.Assert(serverTunnel.NsxtL2VpnTunnel.ConnectorInitiationMode, Equals, "ON_DEMAND")
	check.Assert(serverTunnel.NsxtL2VpnTunnel.PreSharedKey, Equals, check.TestName())

	fetchedServerTunnel, err := edge.GetL2VpnTunnelById(serverTunnel.NsxtL2VpnTunnel.ID)
	check.Assert(err, IsNil)
	check.Assert(fetchedServerTunnel, DeepEquals, serverTunnel)

	updatedServerTunnelParams := serverTunnelParams
	updatedServerTunnelParams.ConnectorInitiationMode = "INITIATOR"
	updatedServerTunnelParams.RemoteEndpointIp = "2.2.2.2"
	updatedServerTunnelParams.TunnelInterface = "192.168.0.1/24"

	updatedServerTunnel, err := serverTunnel.Update(updatedServerTunnelParams)
	check.Assert(err, IsNil)
	check.Assert(updatedServerTunnel, NotNil)

	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.Name, Equals, check.TestName())
	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.Description, Equals, check.TestName())
	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.SessionMode, Equals, "SERVER")
	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.Enabled, Equals, true)
	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.LocalEndpointIp, Equals, localEndpointIp[0].IPAddress)
	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.RemoteEndpointIp, Equals, "2.2.2.2")
	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.TunnelInterface, Equals, "192.168.0.1/24")
	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.ConnectorInitiationMode, Equals, "INITIATOR")
	check.Assert(updatedServerTunnel.NsxtL2VpnTunnel.PreSharedKey, Equals, check.TestName())

	err = updatedServerTunnel.Delete()
	check.Assert(err, IsNil)

	deletedServerTunnel, err := edge.GetL2VpnTunnelById(serverTunnel.NsxtL2VpnTunnel.ID)
	check.Assert(err, NotNil)
	check.Assert(deletedServerTunnel, IsNil)

	// CLIENT Tunnel part
	clientTunnelParams := &types.NsxtL2VpnTunnel{
		Name:             check.TestName(),
		Description:      check.TestName(),
		SessionMode:      "CLIENT",
		Enabled:          true,
		LocalEndpointIp:  localEndpointIp[0].IPAddress,
		RemoteEndpointIp: "1.1.1.1",
		PreSharedKey:     check.TestName(),
		PeerCode:         peerCode,
		StretchedNetworks: []types.EdgeL2VpnStretchedNetwork{
			{
				NetworkRef: types.OpenApiReference{
					Name: network.OrgVDCNetwork.Name,
					ID:   network.OrgVDCNetwork.ID,
				},
				TunnelID: 1,
			},
		},
		Logging: false,
	}

	clientTunnel, err := edge.CreateL2VpnTunnel(clientTunnelParams)
	check.Assert(err, IsNil)
	check.Assert(clientTunnel, NotNil)
	AddToCleanupListOpenApi(clientTunnel.NsxtL2VpnTunnel.ID, check.TestName(),
		fmt.Sprintf(types.OpenApiPathVersion1_0_0+
			types.OpenApiEndpointEdgeGatewayL2VpnTunnel+
			clientTunnel.NsxtL2VpnTunnel.ID, edge.EdgeGateway.ID))

	check.Assert(clientTunnel.NsxtL2VpnTunnel.Name, Equals, check.TestName())
	check.Assert(clientTunnel.NsxtL2VpnTunnel.Description, Equals, check.TestName())
	check.Assert(clientTunnel.NsxtL2VpnTunnel.SessionMode, Equals, "CLIENT")
	check.Assert(clientTunnel.NsxtL2VpnTunnel.Enabled, Equals, true)
	check.Assert(clientTunnel.NsxtL2VpnTunnel.LocalEndpointIp, Equals, localEndpointIp[0].IPAddress)
	check.Assert(clientTunnel.NsxtL2VpnTunnel.RemoteEndpointIp, Equals, "1.1.1.1")
	check.Assert(clientTunnel.NsxtL2VpnTunnel.PreSharedKey, Equals, check.TestName())

	fetchedClientTunnel, err := edge.GetL2VpnTunnelById(clientTunnel.NsxtL2VpnTunnel.ID)
	check.Assert(err, IsNil)
	check.Assert(fetchedClientTunnel, DeepEquals, clientTunnel)

	updatedClientTunnelParams := clientTunnelParams
	updatedClientTunnelParams.RemoteEndpointIp = "2.2.2.2"

	updatedClientTunnel, err := clientTunnel.Update(updatedClientTunnelParams)
	check.Assert(err, IsNil)
	check.Assert(updatedClientTunnel, NotNil)

	check.Assert(updatedClientTunnel.NsxtL2VpnTunnel.Name, Equals, check.TestName())
	check.Assert(updatedClientTunnel.NsxtL2VpnTunnel.Description, Equals, check.TestName())
	check.Assert(updatedClientTunnel.NsxtL2VpnTunnel.SessionMode, Equals, "CLIENT")
	check.Assert(updatedClientTunnel.NsxtL2VpnTunnel.Enabled, Equals, true)
	check.Assert(updatedClientTunnel.NsxtL2VpnTunnel.LocalEndpointIp, Equals, localEndpointIp[0].IPAddress)
	check.Assert(updatedClientTunnel.NsxtL2VpnTunnel.RemoteEndpointIp, Equals, "2.2.2.2")
	check.Assert(updatedClientTunnel.NsxtL2VpnTunnel.PreSharedKey, Equals, check.TestName())

	// There is an unexpected error in all versions up to 10.5.0, it happens
	// when a L2 VPN Tunnel is created in CLIENT mode, has atleast one Org VDC
	// network attached, and is updated in any way. After that, to delete the tunnel,
	// one needs to send a DELETE request twice (and get an error on the first attempt)
	// or de-attach the Org Networks from the Tunnel and send the DELETE request
	err = updatedClientTunnel.Delete()
	errorMessage := regexp.MustCompile("error code 500030")
	check.Assert(errorMessage.MatchString(err.Error()), Equals, true)

	err = updatedClientTunnel.Delete()
	check.Assert(err, IsNil)

	deletedClientTunnel, err := edge.GetL2VpnTunnelById(clientTunnel.NsxtL2VpnTunnel.ID)
	check.Assert(err, NotNil)
	check.Assert(deletedClientTunnel, IsNil)
}
