//go:build nsxt || alb || functional || ALL
// +build nsxt alb functional ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAllEdgeAlbServiceEngineGroupAssignmentsDedicatedSe(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtAlbConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayAlb)

	controller, cloud, seGroup := spawnAlbControllerCloudServiceEngineGroup(vcd, check, "DEDICATED")
	edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Enable ALB on Edge Gateway with default ServiceNetworkDefinition
	albSettingsConfig := &types.NsxtAlbConfig{
		Enabled: true,
	}
	enabledSettings, err := edge.UpdateAlbSettings(albSettingsConfig)
	check.Assert(err, IsNil)
	check.Assert(enabledSettings.Enabled, Equals, true)
	PrependToCleanupList("OpenApiEntityAlbSettingsDisable", "OpenApiEntityAlbSettingsDisable", edge.EdgeGateway.Name, check.TestName())

	///////////
	a := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: edge.EdgeGateway.ID},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: seGroup.NsxtAlbServiceEngineGroup.ID},
	}
	assignment, err := vcd.client.CreateAlbServiceEngineGroupAssignment(a)
	check.Assert(err, IsNil)
	check.Assert(assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ID, Not(Equals), "")
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments + assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ID
	PrependToCleanupListOpenApi(assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name, check.TestName(), openApiEndpoint)

	// Get By ID
	assignmentById, err := vcd.client.GetAlbServiceEngineGroupAssignmentById(assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ID)
	check.Assert(err, IsNil)
	check.Assert(assignmentById.NsxtEdgeAlbServiceEngineGroupAssignment, DeepEquals, assignment.NsxtEdgeAlbServiceEngineGroupAssignment)

	// Get By Name
	assignmentByName, err := vcd.client.GetAlbServiceEngineGroupAssignmentByName(assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name)
	check.Assert(err, IsNil)
	check.Assert(assignmentByName.NsxtEdgeAlbServiceEngineGroupAssignment, DeepEquals, assignment.NsxtEdgeAlbServiceEngineGroupAssignment)

	// Get all
	allAssignments, err := vcd.client.GetAllAlbServiceEngineGroupAssignments(nil)
	check.Assert(err, IsNil)
	var foundAssignment bool
	for i := range allAssignments {
		if allAssignments[i].NsxtEdgeAlbServiceEngineGroupAssignment.ID == assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ID {
			foundAssignment = true
		}
	}
	check.Assert(foundAssignment, Equals, true)

	assignment.NsxtEdgeAlbServiceEngineGroupAssignment.MaxVirtualServices = takeIntAddress(50)
	assignment.NsxtEdgeAlbServiceEngineGroupAssignment.MinVirtualServices = takeIntAddress(30)
	// Expect an error because "DEDICATED" service engine group does not support specifying virtual services
	updatedAssignment, err := assignment.Update(assignment.NsxtEdgeAlbServiceEngineGroupAssignment)
	check.Assert(err, NotNil)
	check.Assert(updatedAssignment, IsNil)

	err = assignment.Delete()
	check.Assert(err, IsNil)

	///////////

	err = edge.DisableAlb()
	check.Assert(err, IsNil)

	// Remove objects
	err = seGroup.Delete()
	check.Assert(err, IsNil)
	err = cloud.Delete()
	check.Assert(err, IsNil)
	err = controller.Delete()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_GetAllEdgeAlbServiceEngineGroupAssignmentsSharedSe(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtAlbConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGatewayAlb)

	controller, cloud, seGroup := spawnAlbControllerCloudServiceEngineGroup(vcd, check, "SHARED")
	edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Enable ALB on Edge Gateway with default ServiceNetworkDefinition
	albSettingsConfig := &types.NsxtAlbConfig{
		Enabled: true,
	}
	enabledSettings, err := edge.UpdateAlbSettings(albSettingsConfig)
	check.Assert(err, IsNil)
	check.Assert(enabledSettings.Enabled, Equals, true)
	PrependToCleanupList("", "OpenApiEntityAlbSettingsDisable", edge.EdgeGateway.Name, check.TestName())

	///////////
	a := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: edge.EdgeGateway.ID},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: seGroup.NsxtAlbServiceEngineGroup.ID},
		MaxVirtualServices:    takeIntAddress(89),
		MinVirtualServices:    takeIntAddress(20),
	}
	assignment, err := vcd.client.CreateAlbServiceEngineGroupAssignment(a)
	check.Assert(err, IsNil)
	check.Assert(assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ID, Not(Equals), "")
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments + assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ID
	PrependToCleanupListOpenApi(assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name, check.TestName(), openApiEndpoint)

	// Get By ID
	assignmentById, err := vcd.client.GetAlbServiceEngineGroupAssignmentById(assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ID)
	check.Assert(err, IsNil)
	check.Assert(assignmentById.NsxtEdgeAlbServiceEngineGroupAssignment, DeepEquals, assignment.NsxtEdgeAlbServiceEngineGroupAssignment)

	// Get By Name
	assignmentByName, err := vcd.client.GetAlbServiceEngineGroupAssignmentByName(assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name)
	check.Assert(err, IsNil)
	check.Assert(assignmentByName.NsxtEdgeAlbServiceEngineGroupAssignment, DeepEquals, assignment.NsxtEdgeAlbServiceEngineGroupAssignment)

	// Get all
	allAssignments, err := vcd.client.GetAllAlbServiceEngineGroupAssignments(nil)
	check.Assert(err, IsNil)
	var foundAssignment bool
	for i := range allAssignments {
		if allAssignments[i].NsxtEdgeAlbServiceEngineGroupAssignment.ID == assignment.NsxtEdgeAlbServiceEngineGroupAssignment.ID {
			foundAssignment = true
		}
	}
	check.Assert(foundAssignment, Equals, true)

	assignment.NsxtEdgeAlbServiceEngineGroupAssignment.MaxVirtualServices = takeIntAddress(50)
	assignment.NsxtEdgeAlbServiceEngineGroupAssignment.MinVirtualServices = takeIntAddress(30)
	// Expect an error because "DEDICATED" service engine group does not support specifying virtual services
	updatedAssignment, err := assignment.Update(assignment.NsxtEdgeAlbServiceEngineGroupAssignment)
	check.Assert(err, IsNil)
	check.Assert(updatedAssignment, NotNil)

	err = assignment.Delete()
	check.Assert(err, IsNil)

	///////////

	err = edge.DisableAlb()
	check.Assert(err, IsNil)

	// Remove objects
	err = seGroup.Delete()
	check.Assert(err, IsNil)
	err = cloud.Delete()
	check.Assert(err, IsNil)
	err = controller.Delete()
	check.Assert(err, IsNil)
}
