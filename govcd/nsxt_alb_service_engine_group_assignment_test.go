//go:build nsxt || alb || functional || ALL

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAllEdgeAlbServiceEngineGroupAssignmentsDedicated(check *C) {
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
	enabledSettings, err := edge.UpdateAlbSettings(albSettingsConfig)
	check.Assert(err, IsNil)
	check.Assert(enabledSettings.Enabled, Equals, true)
	PrependToCleanupList("OpenApiEntityAlbSettingsDisable", "OpenApiEntityAlbSettingsDisable", edge.EdgeGateway.Name, check.TestName())

	serviceEngineGroupAssignmentConfig := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: edge.EdgeGateway.ID},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: seGroup.NsxtAlbServiceEngineGroup.ID},
	}
	assignment, err := vcd.client.CreateAlbServiceEngineGroupAssignment(serviceEngineGroupAssignmentConfig)
	check.Assert(err, IsNil)
	check.Assert(assignment.NsxtAlbServiceEngineGroupAssignment.ID, Not(Equals), "")
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments + assignment.NsxtAlbServiceEngineGroupAssignment.ID
	PrependToCleanupListOpenApi(assignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name, check.TestName(), openApiEndpoint)

	// Get By ID
	assignmentById, err := vcd.client.GetAlbServiceEngineGroupAssignmentById(assignment.NsxtAlbServiceEngineGroupAssignment.ID)
	check.Assert(err, IsNil)
	check.Assert(assignmentById.NsxtAlbServiceEngineGroupAssignment, DeepEquals, assignment.NsxtAlbServiceEngineGroupAssignment)

	// Get By Name
	assignmentByName, err := vcd.client.GetAlbServiceEngineGroupAssignmentByName(assignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name)
	check.Assert(err, IsNil)
	check.Assert(assignmentByName.NsxtAlbServiceEngineGroupAssignment, DeepEquals, assignment.NsxtAlbServiceEngineGroupAssignment)

	// Filtered by name and Edge Gateway ID
	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("gatewayRef.id==%s", edge.EdgeGateway.ID))
	filteredAssignmentByName, err := vcd.client.GetFilteredAlbServiceEngineGroupAssignmentByName(assignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name, queryParams)
	check.Assert(err, IsNil)
	check.Assert(filteredAssignmentByName.NsxtAlbServiceEngineGroupAssignment, DeepEquals, filteredAssignmentByName.NsxtAlbServiceEngineGroupAssignment)

	// Get all
	allAssignments, err := vcd.client.GetAllAlbServiceEngineGroupAssignments(nil)
	check.Assert(err, IsNil)
	var foundAssignment bool
	for i := range allAssignments {
		if allAssignments[i].NsxtAlbServiceEngineGroupAssignment.ID == assignment.NsxtAlbServiceEngineGroupAssignment.ID {
			foundAssignment = true
		}
	}
	check.Assert(foundAssignment, Equals, true)

	assignment.NsxtAlbServiceEngineGroupAssignment.MaxVirtualServices = addrOf(50)
	assignment.NsxtAlbServiceEngineGroupAssignment.MinVirtualServices = addrOf(30)
	// Expect an error because "DEDICATED" service engine group does not support specifying virtual services
	updatedAssignment, err := assignment.Update(assignment.NsxtAlbServiceEngineGroupAssignment)
	check.Assert(err, NotNil)
	check.Assert(updatedAssignment, IsNil)

	// Perform immediate cleanups
	err = assignment.Delete()
	check.Assert(err, IsNil)
	err = edge.DisableAlb()
	check.Assert(err, IsNil)
	err = seGroup.Delete()
	check.Assert(err, IsNil)
	err = cloud.Delete()
	check.Assert(err, IsNil)
	err = controller.Delete()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_GetAllEdgeAlbServiceEngineGroupAssignmentsShared(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtAlbConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAlbEdgeGateway)

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
	PrependToCleanupList(check.TestName()+"-ALB-settings", "OpenApiEntityAlbSettingsDisable", edge.EdgeGateway.Name, check.TestName())

	serviceEngineGroupAssignmentConfig := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: edge.EdgeGateway.ID},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: seGroup.NsxtAlbServiceEngineGroup.ID},
		MaxVirtualServices:    addrOf(89),
		MinVirtualServices:    addrOf(20),
	}
	assignment, err := vcd.client.CreateAlbServiceEngineGroupAssignment(serviceEngineGroupAssignmentConfig)
	check.Assert(err, IsNil)
	check.Assert(assignment.NsxtAlbServiceEngineGroupAssignment.ID, Not(Equals), "")
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments + assignment.NsxtAlbServiceEngineGroupAssignment.ID
	PrependToCleanupListOpenApi(assignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name, check.TestName(), openApiEndpoint)

	// Get By ID
	assignmentById, err := vcd.client.GetAlbServiceEngineGroupAssignmentById(assignment.NsxtAlbServiceEngineGroupAssignment.ID)
	check.Assert(err, IsNil)
	check.Assert(assignmentById.NsxtAlbServiceEngineGroupAssignment, DeepEquals, assignment.NsxtAlbServiceEngineGroupAssignment)

	// Get By Name
	assignmentByName, err := vcd.client.GetAlbServiceEngineGroupAssignmentByName(assignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name)
	check.Assert(err, IsNil)
	check.Assert(assignmentByName.NsxtAlbServiceEngineGroupAssignment, DeepEquals, assignment.NsxtAlbServiceEngineGroupAssignment)

	// Filtered by name and Edge Gateway ID
	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("gatewayRef.id==%s", edge.EdgeGateway.ID))
	filteredAssignmentByName, err := vcd.client.GetFilteredAlbServiceEngineGroupAssignmentByName(assignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name, queryParams)
	check.Assert(err, IsNil)
	check.Assert(filteredAssignmentByName.NsxtAlbServiceEngineGroupAssignment, DeepEquals, filteredAssignmentByName.NsxtAlbServiceEngineGroupAssignment)

	// Get all
	allAssignments, err := vcd.client.GetAllAlbServiceEngineGroupAssignments(nil)
	check.Assert(err, IsNil)
	var foundAssignment bool
	for i := range allAssignments {
		if allAssignments[i].NsxtAlbServiceEngineGroupAssignment.ID == assignment.NsxtAlbServiceEngineGroupAssignment.ID {
			foundAssignment = true
		}
	}
	check.Assert(foundAssignment, Equals, true)

	assignment.NsxtAlbServiceEngineGroupAssignment.MaxVirtualServices = addrOf(50)
	assignment.NsxtAlbServiceEngineGroupAssignment.MinVirtualServices = addrOf(30)

	updatedAssignment, err := assignment.Update(assignment.NsxtAlbServiceEngineGroupAssignment)
	check.Assert(err, IsNil)
	check.Assert(updatedAssignment, NotNil)

	// Perform immediate cleanups
	err = assignment.Delete()
	check.Assert(err, IsNil)
	err = edge.DisableAlb()
	check.Assert(err, IsNil)
	err = seGroup.Delete()
	check.Assert(err, IsNil)
	err = cloud.Delete()
	check.Assert(err, IsNil)
	err = controller.Delete()
	check.Assert(err, IsNil)
}
