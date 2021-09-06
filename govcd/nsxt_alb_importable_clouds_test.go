//go:build alb || functional || ALL
// +build alb functional ALL

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

// Test_GetAllAlbImportableClouds tests if if there is at least one importable cloud available
func (vcd *TestVCD) Test_GetAllAlbImportableClouds(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtAlbConfiguration(vcd, check)

	albController := spawnAlbController(vcd, check)

	controllers, err := vcd.client.GetAllAlbControllers(nil)
	check.Assert(err, IsNil)
	check.Assert(len(controllers) > 0, Equals, true)

	// Test client function with explicit ALB Controller ID requirement
	clientImportableClouds, err := vcd.client.GetAllAlbImportableClouds(controllers[0].NsxtAlbController.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(len(clientImportableClouds) > 0, Equals, true)

	// Test function attached directly to NsxtAlbController
	controllerImportableClouds, err := controllers[0].GetAllAlbImportableClouds(nil)
	check.Assert(err, IsNil)
	check.Assert(len(controllerImportableClouds) > 0, Equals, true)

	// Cleanup
	err = albController.Delete()
	check.Assert(err, IsNil)
}
