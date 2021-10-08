//go:build nsxt || alb || functional || ALL
// +build nsxt alb functional ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAlbSettings(check *C) {
	skipNoNsxtAlbConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayAlb)

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
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayAlb)

	controller, cloud, seGroup := spawnAlbControllerCloudServiceEngineGroup(vcd, check)
	edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Enable ALB on Edge Gateway with default ServiceNetworkDefinition
	albSettingsConfig := &types.NsxtAlbConfig{
		Enabled: true,
	}
	enabledSettings, err := edge.UpdateAlbSettings(albSettingsConfig)
	check.Assert(err, IsNil)
	check.Assert(enabledSettings.Enabled, Equals, true)
	check.Assert(enabledSettings.ServiceNetworkDefinition, Equals, "192.168.255.1/25")

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

	// Disable ALB on Edge Gateway again and ensure it was disabled
	err = edge.DisableAlb()
	check.Assert(err, IsNil)

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
