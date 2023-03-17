//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtIpSecVpn(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	ipSecDef := &types.NsxtIpSecVpnTunnel{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		Enabled:     true,
		LocalEndpoint: types.NsxtIpSecVpnTunnelLocalEndpoint{
			LocalAddress:  edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP,
			LocalNetworks: []string{"10.10.10.0/24"},
		},
		RemoteEndpoint: types.NsxtIpSecVpnTunnelRemoteEndpoint{
			RemoteId:       "192.168.140.1",
			RemoteAddress:  "192.168.140.1",
			RemoteNetworks: []string{"20.20.20.0/24"},
		},
		PreSharedKey: "PSK-Sec",
		SecurityType: "DEFAULT",
		Logging:      true,
	}

	runIpSecVpnTests(check, edge, ipSecDef)

}

func (vcd *TestVCD) Test_NsxtIpSecVpnCustomSecurityProfile(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	ipSecDef := &types.NsxtIpSecVpnTunnel{
		Name:               check.TestName(),
		Description:        check.TestName() + "-description",
		Enabled:            true,
		AuthenticationMode: types.NsxtIpSecVpnAuthenticationModePSK, // Default value even when it is unset
		LocalEndpoint: types.NsxtIpSecVpnTunnelLocalEndpoint{
			LocalAddress:  edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP,
			LocalNetworks: []string{"10.10.10.0/24"},
		},
		RemoteEndpoint: types.NsxtIpSecVpnTunnelRemoteEndpoint{
			RemoteId:       "192.168.140.1",
			RemoteAddress:  "192.168.140.1",
			RemoteNetworks: []string{"20.20.20.0/24"},
		},
		PreSharedKey: "PSK-Sec",
		SecurityType: "DEFAULT",
		Logging:      false,
	}

	createdIpSecVpn, err := edge.CreateIpSecVpnTunnel(ipSecDef)
	check.Assert(err, IsNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + fmt.Sprintf(types.OpenApiEndpointIpSecVpnTunnel, createdIpSecVpn.edgeGatewayId) + createdIpSecVpn.NsxtIpSecVpn.ID
	AddToCleanupListOpenApi(createdIpSecVpn.NsxtIpSecVpn.Name, check.TestName(), openApiEndpoint)

	// Customize Security Profile
	secProfile := &types.NsxtIpSecVpnTunnelSecurityProfile{
		SecurityType: "CUSTOM",
		IkeConfiguration: types.NsxtIpSecVpnTunnelProfileIkeConfiguration{
			IkeVersion:           "IKE_V2",
			EncryptionAlgorithms: []string{"AES_128"},
			DigestAlgorithms:     []string{"SHA2_256"},
			DhGroups:             []string{"GROUP14"},
			SaLifeTime:           takeIntAddress(86400),
		},
		TunnelConfiguration: types.NsxtIpSecVpnTunnelProfileTunnelConfiguration{
			PerfectForwardSecrecyEnabled: true,
			DfPolicy:                     "CLEAR",
			EncryptionAlgorithms:         []string{"AES_256"},
			DigestAlgorithms:             []string{"SHA2_256"},
			DhGroups:                     []string{"GROUP14"},
			SaLifeTime:                   takeIntAddress(3600),
		},
		DpdConfiguration: types.NsxtIpSecVpnTunnelProfileDpdConfiguration{ProbeInterval: 3},
	}
	setSecProfile, err := createdIpSecVpn.UpdateTunnelConnectionProperties(secProfile)
	check.Assert(err, IsNil)
	check.Assert(setSecProfile, DeepEquals, secProfile)

	// Check if status endpoint works properly, but cannot rely on returned status as it is not immediately returned and
	// it can hold on for a long time before available. At least validate that this function does not return error.
	_, err = createdIpSecVpn.GetStatus()
	check.Assert(err, IsNil)

	//Latest Version
	latestSecProfile, err := edge.GetIpSecVpnTunnelById(createdIpSecVpn.NsxtIpSecVpn.ID)
	check.Assert(err, IsNil)

	// Reset security profile to default
	latestSecProfile.NsxtIpSecVpn.SecurityType = "DEFAULT"
	updatedIpSecVpn, err := createdIpSecVpn.Update(latestSecProfile.NsxtIpSecVpn)
	check.Assert(err, IsNil)
	// All fields should be the same, except version
	latestSecProfile.NsxtIpSecVpn.Version = updatedIpSecVpn.NsxtIpSecVpn.Version
	check.Assert(updatedIpSecVpn.NsxtIpSecVpn, DeepEquals, latestSecProfile.NsxtIpSecVpn)

	// Remove object
	err = createdIpSecVpn.Delete()
	check.Assert(err, IsNil)
}

// Test_NsxtIpSecVpnUniqueness checks that uniqueness is enforced at API level on LocalAddress+RemoteAddress by creating
// two IPsec VPN tunnels with different field values but the same LocalAddress and RemoteAddress without any other
// fields clashing.
func (vcd *TestVCD) Test_NsxtIpSecVpnUniqueness(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	ipSecDef := &types.NsxtIpSecVpnTunnel{
		Name:        check.TestName(),
		Description: check.TestName() + "-description",
		Enabled:     true,
		LocalEndpoint: types.NsxtIpSecVpnTunnelLocalEndpoint{
			LocalAddress:  edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP,
			LocalNetworks: []string{"10.10.10.0/24"},
		},
		RemoteEndpoint: types.NsxtIpSecVpnTunnelRemoteEndpoint{
			RemoteId:       "192.168.170.1",
			RemoteAddress:  "192.168.170.1",
			RemoteNetworks: []string{"20.20.20.0/24"},
		},
		PreSharedKey: "PSK-Sec",
		SecurityType: "DEFAULT",
		Logging:      true,
	}

	// Create first IPsec VPN Tunnel
	createdIpSecVpn, err := edge.CreateIpSecVpnTunnel(ipSecDef)
	check.Assert(err, IsNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + fmt.Sprintf(types.OpenApiEndpointIpSecVpnTunnel, createdIpSecVpn.edgeGatewayId) + createdIpSecVpn.NsxtIpSecVpn.ID
	AddToCleanupListOpenApi(createdIpSecVpn.NsxtIpSecVpn.Name, check.TestName(), openApiEndpoint)

	// Try to create second IPsec VPN Tunnel with the same localAddress and RemoteAddress and expect an error
	ipSecDef2 := &types.NsxtIpSecVpnTunnel{
		Name:        check.TestName() + "2",
		Description: check.TestName() + "-description2",
		Enabled:     true,
		LocalEndpoint: types.NsxtIpSecVpnTunnelLocalEndpoint{
			LocalAddress:  edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP,
			LocalNetworks: []string{"40.10.10.0/24"},
		},
		RemoteEndpoint: types.NsxtIpSecVpnTunnelRemoteEndpoint{
			RemoteId:       "192.168.170.1",
			RemoteAddress:  "192.168.170.1",
			RemoteNetworks: []string{"50.20.20.0/24"},
		},
		PreSharedKey: "PSK-Sec",
		SecurityType: "DEFAULT",
		Logging:      true,
	}

	// Ensure that the IsEqual matches those definitions as equal ones
	check.Assert(createdIpSecVpn.IsEqualTo(ipSecDef2), Equals, true)

	createdIpSecVpn2, err := edge.CreateIpSecVpnTunnel(ipSecDef2)
	check.Assert(err.Error(), Matches, ".*IPSec VPN Tunnel with local address .* and remote address .* is already in use.*")
	check.Assert(createdIpSecVpn2, IsNil)

	// Removing the first IPsec VPN tunnel
	err = createdIpSecVpn.Delete()
	check.Assert(err, IsNil)
}

func runIpSecVpnTests(check *C, edge *NsxtEdgeGateway, ipSecDef *types.NsxtIpSecVpnTunnel) {
	createdIpSecVpn, err := edge.CreateIpSecVpnTunnel(ipSecDef)
	check.Assert(err, IsNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + fmt.Sprintf(types.OpenApiEndpointIpSecVpnTunnel, createdIpSecVpn.edgeGatewayId) + createdIpSecVpn.NsxtIpSecVpn.ID
	AddToCleanupListOpenApi(createdIpSecVpn.NsxtIpSecVpn.Name, check.TestName(), openApiEndpoint)

	foundIpSecVpnById, err := edge.GetIpSecVpnTunnelById(createdIpSecVpn.NsxtIpSecVpn.ID)
	check.Assert(err, IsNil)
	check.Assert(foundIpSecVpnById.NsxtIpSecVpn, DeepEquals, createdIpSecVpn.NsxtIpSecVpn)

	foundIpSecVpnByName, err := edge.GetIpSecVpnTunnelByName(createdIpSecVpn.NsxtIpSecVpn.Name)
	check.Assert(err, IsNil)
	check.Assert(foundIpSecVpnByName.NsxtIpSecVpn, DeepEquals, createdIpSecVpn.NsxtIpSecVpn)
	check.Assert(foundIpSecVpnByName.NsxtIpSecVpn, DeepEquals, foundIpSecVpnById.NsxtIpSecVpn)

	check.Assert(createdIpSecVpn.NsxtIpSecVpn.ID, Not(Equals), "")

	ipSecDef.Name = check.TestName() + "-updated"
	ipSecDef.RemoteEndpoint.RemoteAddress = "192.168.40.1"
	ipSecDef.ID = createdIpSecVpn.NsxtIpSecVpn.ID

	updatedIpSecVpn, err := createdIpSecVpn.Update(ipSecDef)
	check.Assert(err, IsNil)
	check.Assert(updatedIpSecVpn.NsxtIpSecVpn.Name, Equals, ipSecDef.Name)
	check.Assert(updatedIpSecVpn.NsxtIpSecVpn.ID, Equals, ipSecDef.ID)
	check.Assert(updatedIpSecVpn.NsxtIpSecVpn.RemoteEndpoint.RemoteAddress, Equals, ipSecDef.RemoteEndpoint.RemoteAddress)

	err = createdIpSecVpn.Delete()
	check.Assert(err, IsNil)

	// Ensure rule does not exist in the list
	allVpnConfigs, err := edge.GetAllIpSecVpnTunnels(nil)
	check.Assert(err, IsNil)
	for _, vpnConfig := range allVpnConfigs {
		check.Assert(vpnConfig.IsEqualTo(updatedIpSecVpn.NsxtIpSecVpn), Equals, false)
	}
}

func (vcd *TestVCD) Test_NsxtIpSecVpnCertificateAuth(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Upload Certificates to use in the test
	aliasForPrivateKey := check.TestName() + "cert-with-private-key"
	privateKeyPassphrase := "test"
	certificateWithPrivateKeyConfig := &types.CertificateLibraryItem{
		Alias:                aliasForPrivateKey,
		Certificate:          certificate,
		PrivateKey:           privateKey,
		PrivateKeyPassphrase: privateKeyPassphrase,
	}

	certWithKey, err := adminOrg.AddCertificateToLibrary(certificateWithPrivateKeyConfig)
	check.Assert(err, IsNil)
	openApiEndpoint, err := getEndpointByVersion(&vcd.client.Client)
	check.Assert(err, IsNil)
	check.Assert(openApiEndpoint, NotNil)
	PrependToCleanupListOpenApi(certWithKey.CertificateLibrary.Alias, check.TestName(),
		openApiEndpoint+certWithKey.CertificateLibrary.Id)

	// Upload CA Certificate to use in the test
	aliasForCaCertificate := check.TestName() + "ca-certificate"
	caCertificateConfig := &types.CertificateLibraryItem{
		Alias:       aliasForCaCertificate,
		Certificate: rootCaCertificate,
	}

	caCert, err := adminOrg.AddCertificateToLibrary(caCertificateConfig)
	check.Assert(err, IsNil)
	PrependToCleanupListOpenApi(caCert.CertificateLibrary.Alias, check.TestName(),
		openApiEndpoint+caCert.CertificateLibrary.Id)

	// Create IPSec VPN configuration with certificate authentication mode
	ipSecDef := &types.NsxtIpSecVpnTunnel{
		Name:               check.TestName(),
		Description:        check.TestName() + "-description",
		Enabled:            true,
		AuthenticationMode: types.NsxtIpSecVpnAuthenticationModeCertificate,
		CertificateRef: &types.OpenApiReference{
			ID: certWithKey.CertificateLibrary.Id,
		},
		CaCertificateRef: &types.OpenApiReference{
			ID: caCert.CertificateLibrary.Id,
		},

		LocalEndpoint: types.NsxtIpSecVpnTunnelLocalEndpoint{
			LocalAddress:  edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP,
			LocalNetworks: []string{"10.10.10.0/24"},
		},
		RemoteEndpoint: types.NsxtIpSecVpnTunnelRemoteEndpoint{
			RemoteId:       "custom-remote-id",
			RemoteAddress:  "192.168.140.1",
			RemoteNetworks: []string{"20.20.20.0/24"},
		},
		SecurityType: "DEFAULT",
		Logging:      true,
	}

	runIpSecVpnTests(check, edge, ipSecDef)

	// cleanup uploaded certificates
	err = certWithKey.Delete()
	check.Assert(err, IsNil)
	err = caCert.Delete()
	check.Assert(err, IsNil)
}
