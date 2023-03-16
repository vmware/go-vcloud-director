//go:build nsxt || alb || functional || ALL

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAllAlbImportableServiceEngineGroups(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	albController, createdAlbCloud := spawnAlbControllerAndCloud(vcd, check)

	importableSeGroups, err := vcd.client.GetAllAlbImportableServiceEngineGroups(createdAlbCloud.NsxtAlbCloud.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(len(importableSeGroups) > 0, Equals, true)
	check.Assert(importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.ID != "", Equals, true)
	check.Assert(importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.DisplayName != "", Equals, true)
	check.Assert(importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.HaMode != "", Equals, true)

	// Get By Name
	impSeGrpByName, err := vcd.client.GetAlbImportableServiceEngineGroupByName(createdAlbCloud.NsxtAlbCloud.ID, importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.DisplayName)
	check.Assert(err, IsNil)
	// Get By ID
	impSeGrpById, err := vcd.client.GetAlbImportableServiceEngineGroupById(createdAlbCloud.NsxtAlbCloud.ID, importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.ID)
	check.Assert(err, IsNil)

	// Get By Name on parent Cloud
	cldImpSeGrpByName, err := createdAlbCloud.GetAlbImportableServiceEngineGroupByName(createdAlbCloud.NsxtAlbCloud.ID, importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.DisplayName)
	check.Assert(err, IsNil)
	// Get By ID on parent Cloud
	cldImpSeGrpById, err := createdAlbCloud.GetAlbImportableServiceEngineGroupById(createdAlbCloud.NsxtAlbCloud.ID, importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.ID)
	check.Assert(err, IsNil)

	check.Assert(impSeGrpByName.NsxtAlbImportableServiceEngineGroups, DeepEquals, importableSeGroups[0].NsxtAlbImportableServiceEngineGroups)
	check.Assert(impSeGrpByName.NsxtAlbImportableServiceEngineGroups, DeepEquals, impSeGrpById.NsxtAlbImportableServiceEngineGroups)
	check.Assert(impSeGrpByName.NsxtAlbImportableServiceEngineGroups, DeepEquals, cldImpSeGrpByName.NsxtAlbImportableServiceEngineGroups)
	check.Assert(impSeGrpByName.NsxtAlbImportableServiceEngineGroups, DeepEquals, cldImpSeGrpById.NsxtAlbImportableServiceEngineGroups)

	// Cleanup
	err = createdAlbCloud.Delete()
	check.Assert(err, IsNil)

	err = albController.Delete()
	check.Assert(err, IsNil)
}
