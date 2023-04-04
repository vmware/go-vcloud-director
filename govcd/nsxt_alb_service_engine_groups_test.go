//go:build nsxt || alb || functional || ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAllAlbServiceEngineGroups(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	controller, createdAlbCloud := spawnAlbControllerAndCloud(vcd, check)

	importableSeGroups, err := vcd.client.GetAllAlbImportableServiceEngineGroups(createdAlbCloud.NsxtAlbCloud.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(len(importableSeGroups) > 0, Equals, true)

	albSeGroup := &types.NsxtAlbServiceEngineGroup{
		Name:            check.TestName() + "SE-group",
		Description:     "Service Engine Group created by " + check.TestName(),
		ReservationType: "DEDICATED",
		ServiceEngineGroupBacking: types.ServiceEngineGroupBacking{
			BackingId: importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.ID,
			LoadBalancerCloudRef: &types.OpenApiReference{
				ID: createdAlbCloud.NsxtAlbCloud.ID,
			},
		},
	}

	// Field is only available when using API version v37.0 onwards
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.0") {
		albSeGroup.SupportedFeatureSet = "PREMIUM"
	}

	createdSeGroup, err := vcd.client.CreateNsxtAlbServiceEngineGroup(albSeGroup)
	check.Assert(err, IsNil)

	check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.ID != "", Equals, true)
	check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.Name, Equals, albSeGroup.Name)
	check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.Description, Equals, albSeGroup.Description)
	check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.ReservationType, Equals, albSeGroup.ReservationType)
	// Field is only populated in responses when using API version v37.0 onwards
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.0") {
		check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.SupportedFeatureSet, Equals, albSeGroup.SupportedFeatureSet)
	}

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups + createdSeGroup.NsxtAlbServiceEngineGroup.ID
	AddToCleanupListOpenApi(createdSeGroup.NsxtAlbServiceEngineGroup.Name, check.TestName(), openApiEndpoint)

	// Sync
	err = createdSeGroup.Sync()
	check.Assert(err, IsNil)

	// Find by Name
	seGroupByName, err := vcd.client.GetAlbServiceEngineGroupByName("", createdSeGroup.NsxtAlbServiceEngineGroup.Name)
	check.Assert(err, IsNil)
	check.Assert(seGroupByName, NotNil)

	// Find by ID
	seGroupById, err := vcd.client.GetAlbServiceEngineGroupById(createdSeGroup.NsxtAlbServiceEngineGroup.ID)
	check.Assert(err, IsNil)
	check.Assert(seGroupById, NotNil)

	check.Assert(seGroupByName.NsxtAlbServiceEngineGroup.ID, Equals, createdSeGroup.NsxtAlbServiceEngineGroup.ID)
	check.Assert(seGroupById.NsxtAlbServiceEngineGroup.ID, Equals, createdSeGroup.NsxtAlbServiceEngineGroup.ID)

	// Test update
	createdSeGroup.NsxtAlbServiceEngineGroup.Name = createdSeGroup.NsxtAlbServiceEngineGroup.Name + "updated"
	// Field is only available when using API version v37.0 onwards
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.0") {
		albSeGroup.SupportedFeatureSet = "STANDARD"
	}
	updatedSeGroup, err := createdSeGroup.Update(createdSeGroup.NsxtAlbServiceEngineGroup)
	check.Assert(err, IsNil)

	// SupportedFeatureSet is a field only available since v37.0, in that case we ignore it in the following DeepEquals
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.0") {
		updatedSeGroup.NsxtAlbServiceEngineGroup.SupportedFeatureSet = createdSeGroup.NsxtAlbServiceEngineGroup.SupportedFeatureSet
	}
	check.Assert(updatedSeGroup.NsxtAlbServiceEngineGroup, DeepEquals, createdSeGroup.NsxtAlbServiceEngineGroup)

	// Cleanup
	err = createdSeGroup.Delete()
	check.Assert(err, IsNil)

	err = createdAlbCloud.Delete()
	check.Assert(err, IsNil)

	err = controller.Delete()
	check.Assert(err, IsNil)
}

// spawnAlbControllerCloudServiceEngineGroup is a helper function to spawn NSX-T ALB Controller, ALB Cloud, and ALB
// Service Engine Group
func spawnAlbControllerCloudServiceEngineGroup(vcd *TestVCD, check *C, seGroupReservationType string) (*NsxtAlbController, *NsxtAlbCloud, *NsxtAlbServiceEngineGroup) {
	skipNoNsxtAlbConfiguration(vcd, check)

	albController, createdAlbCloud := spawnAlbControllerAndCloud(vcd, check)

	importableSeGroup, err := vcd.client.GetAlbImportableServiceEngineGroupByName(createdAlbCloud.NsxtAlbCloud.ID, vcd.config.VCD.Nsxt.NsxtAlbServiceEngineGroup)
	check.Assert(err, IsNil)

	albSeGroup := &types.NsxtAlbServiceEngineGroup{
		Name:            check.TestName() + "SE-group",
		Description:     "Service Engine Group created by " + check.TestName(),
		ReservationType: seGroupReservationType,
		ServiceEngineGroupBacking: types.ServiceEngineGroupBacking{
			BackingId: importableSeGroup.NsxtAlbImportableServiceEngineGroups.ID,
			LoadBalancerCloudRef: &types.OpenApiReference{
				ID: createdAlbCloud.NsxtAlbCloud.ID,
			},
		},
	}

	// Field is only available when using API version v37.0 onwards
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.0") {
		albSeGroup.SupportedFeatureSet = "PREMIUM"
	}

	createdSeGroup, err := vcd.client.CreateNsxtAlbServiceEngineGroup(albSeGroup)
	check.Assert(err, IsNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups + createdSeGroup.NsxtAlbServiceEngineGroup.ID
	PrependToCleanupListOpenApi(createdSeGroup.NsxtAlbServiceEngineGroup.Name, check.TestName(), openApiEndpoint)

	return albController, createdAlbCloud, createdSeGroup
}
