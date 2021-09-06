//go:build alb || functional || ALL
// +build alb functional ALL

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

	// Cleanup
	err = createdAlbCloud.Delete()
	check.Assert(err, IsNil)

	err = albController.Delete()
	check.Assert(err, IsNil)
}
