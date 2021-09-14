//go:build nsxt || alb || functional || ALL
// +build nsxt alb functional ALL

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

	createdSeGroup, err := vcd.client.CreateNsxtAlbServiceEngineGroup(albSeGroup)
	check.Assert(err, IsNil)

	check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.ID != "", Equals, true)
	check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.Name, Equals, albSeGroup.Name)
	check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.Description, Equals, albSeGroup.Description)
	check.Assert(createdSeGroup.NsxtAlbServiceEngineGroup.ReservationType, Equals, albSeGroup.ReservationType)

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
	updatedSeGroup, err := createdSeGroup.Update(createdSeGroup.NsxtAlbServiceEngineGroup)
	check.Assert(err, IsNil)

	check.Assert(updatedSeGroup.NsxtAlbServiceEngineGroup, DeepEquals, createdSeGroup.NsxtAlbServiceEngineGroup)

	// Cleanup
	err = createdSeGroup.Delete()
	check.Assert(err, IsNil)

	err = createdAlbCloud.Delete()
	check.Assert(err, IsNil)

	err = controller.Delete()
	check.Assert(err, IsNil)
}
