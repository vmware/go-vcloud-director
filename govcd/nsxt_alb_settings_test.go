//go:build nsxt || alb || functional || ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAlbSettings(check *C) {
	skipNoNsxtAlbConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAlbEdgeGateway)

	edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	albSettings, err := edge.GetAlbSettings()
	check.Assert(err, IsNil)
	check.Assert(albSettings, NotNil)
	check.Assert(albSettings.Enabled, Equals, false)
}

func (vcd *TestVCD) Test_UpdateAlbSettings(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtAlbConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAlbEdgeGateway)

	controller, cloud, seGroup := spawnAlbControllerCloudServiceEngineGroup(vcd, check, "DEDICATED")
	edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Enable ALB on Edge Gateway with default ServiceNetworkDefinition
	albSettingsConfig := &types.NsxtAlbConfig{
		Enabled: true,
	}

	// Field is only available when using API version v37.0 onwards
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.0") {
		albSettingsConfig.SupportedFeatureSet = "STANDARD"
	}

	enabledSettings, err := edge.UpdateAlbSettings(albSettingsConfig)
	check.Assert(err, IsNil)
	check.Assert(enabledSettings.Enabled, Equals, true)
	check.Assert(enabledSettings.ServiceNetworkDefinition, Equals, "192.168.255.1/25")
	// Field is only available when using API version v37.0 onwards
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.0") {
		check.Assert(enabledSettings.SupportedFeatureSet, Equals, "STANDARD")
	}
	PrependToCleanupList("", "OpenApiEntityAlbSettingsDisable", edge.EdgeGateway.Name, check.TestName())

	// Disable ALB on Edge Gateway
	albSettingsConfig.Enabled = false
	disabledSettings, err := edge.UpdateAlbSettings(albSettingsConfig)
	check.Assert(err, IsNil)
	check.Assert(disabledSettings.Enabled, Equals, false)

	// Enable ALB on Edge Gateway with custom ServiceNetworkDefinition
	albSettingsConfig.Enabled = true
	albSettingsConfig.ServiceNetworkDefinition = "93.93.11.1/25"
	enabledSettingsCustomServiceDefinition, err := edge.UpdateAlbSettings(albSettingsConfig)
	check.Assert(err, IsNil)
	check.Assert(enabledSettingsCustomServiceDefinition.Enabled, Equals, true)
	check.Assert(enabledSettingsCustomServiceDefinition.ServiceNetworkDefinition, Equals, "93.93.11.1/25")

	// Disable ALB on Edge Gateway
	err = edge.DisableAlb()
	check.Assert(err, IsNil)

	// Enable IPv6 service network definition (VCD 10.4.0+)
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.0") {
		printVerbose("Enabling IPv6 service network definition for VCD 10.4.0+\n")
		albSettingsConfig.Ipv6ServiceNetworkDefinition = "2001:0db8:85a3:0000:0000:8a2e:0370:7334/120"
		enabledSettingsIpv6ServiceDefinition, err := edge.UpdateAlbSettings(albSettingsConfig)
		check.Assert(err, IsNil)
		check.Assert(enabledSettingsIpv6ServiceDefinition.Ipv6ServiceNetworkDefinition, Equals, "2001:0db8:85a3:0000:0000:8a2e:0370:7334/120")
		err = edge.DisableAlb()
		check.Assert(err, IsNil)
	}

	// Enable Transparent mode (VCD 10.4.1+)
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.1") {
		printVerbose("Enabling Transparent mode for VCD 10.4.1+\n")
		albSettingsConfig.TransparentModeEnabled = takeBoolPointer(true)
		enabledSettingsTransparentMode, err := edge.UpdateAlbSettings(albSettingsConfig)
		check.Assert(err, IsNil)
		check.Assert(*enabledSettingsTransparentMode.TransparentModeEnabled, Equals, true)
		err = edge.DisableAlb()
		check.Assert(err, IsNil)
	}

	albSettings, err := edge.GetAlbSettings()
	check.Assert(err, IsNil)
	check.Assert(albSettings, NotNil)
	check.Assert(albSettings.Enabled, Equals, false)

	// Remove objects
	err = seGroup.Delete()
	check.Assert(err, IsNil)
	err = cloud.Delete()
	check.Assert(err, IsNil)
	err = controller.Delete()
	check.Assert(err, IsNil)
}
