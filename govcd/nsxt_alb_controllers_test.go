//go:build nsxt || alb || functional || ALL

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// Test_NsxtAlbController tests out NSX-T ALB Controller capabilities
func (vcd *TestVCD) Test_NsxtAlbController(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtAlbConfiguration(vcd, check)

	newController := spawnAlbController(vcd, check)

	// Get by Url
	controllerByUrl, err := vcd.client.GetAlbControllerByUrl(newController.NsxtAlbController.Url)
	check.Assert(err, IsNil)

	// Get by Name
	controllerByName, err := vcd.client.GetAlbControllerByName(controllerByUrl.NsxtAlbController.Name)
	check.Assert(err, IsNil)
	check.Assert(controllerByName.NsxtAlbController.ID, Equals, controllerByUrl.NsxtAlbController.ID)

	// Get by ID
	controllerById, err := vcd.client.GetAlbControllerById(controllerByUrl.NsxtAlbController.ID)
	check.Assert(err, IsNil)
	check.Assert(controllerById.NsxtAlbController.ID, Equals, controllerByName.NsxtAlbController.ID)

	// Get all Controllers and expect to find at least the known one
	allControllers, err := vcd.client.GetAllAlbControllers(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allControllers) > 0, Equals, true)
	var foundController bool
	for controllerIndex := range allControllers {
		if allControllers[controllerIndex].NsxtAlbController.ID == controllerByUrl.NsxtAlbController.ID {
			foundController = true
		}
	}
	check.Assert(foundController, Equals, true)

	// Check filtering for GetAllAlbControllers works
	filter := url.Values{}
	filter.Add("filter", "name=="+controllerByUrl.NsxtAlbController.Name)
	filteredControllers, err := vcd.client.GetAllAlbControllers(nil)
	check.Assert(err, IsNil)
	check.Assert(len(filteredControllers), Equals, 1)
	check.Assert(filteredControllers[0].NsxtAlbController.ID, Equals, controllerByUrl.NsxtAlbController.ID)

	// Test update of ALB controller
	updateControllerDef := &types.NsxtAlbController{
		ID:          controllerByUrl.NsxtAlbController.ID,
		Name:        controllerByUrl.NsxtAlbController.Name + "-update",
		Description: "Description set",
		Url:         vcd.config.VCD.Nsxt.NsxtAlbControllerUrl,
		Username:    vcd.config.VCD.Nsxt.NsxtAlbControllerUser,
		Password:    vcd.config.VCD.Nsxt.NsxtAlbControllerPassword,
		LicenseType: "BASIC", // Not used since v37.0
	}
	updatedController, err := controllerByUrl.Update(updateControllerDef)
	check.Assert(err, IsNil)
	check.Assert(updatedController.NsxtAlbController.Name, Equals, updateControllerDef.Name)
	check.Assert(updatedController.NsxtAlbController.Description, Equals, updateControllerDef.Description)
	check.Assert(updatedController.NsxtAlbController.Url, Equals, updateControllerDef.Url)
	check.Assert(updatedController.NsxtAlbController.Username, Equals, updateControllerDef.Username)
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.0") {
		check.Assert(updatedController.NsxtAlbController.LicenseType, Equals, updateControllerDef.LicenseType)
	}

	// Revert settings to original ones
	_, err = controllerByUrl.Update(controllerByUrl.NsxtAlbController)
	check.Assert(err, IsNil)

	// Remove and make sure it is not found
	err = updatedController.Delete()
	check.Assert(err, IsNil)

	// Try to find controller and expect an
	_, err = vcd.client.GetAlbControllerByName(controllerByUrl.NsxtAlbController.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
}

// spawnAlbController is a helper function to spawn NSX-T ALB Controller instance from defined config
func spawnAlbController(vcd *TestVCD, check *C) *NsxtAlbController {
	skipNoNsxtAlbConfiguration(vcd, check)

	newControllerDef := &types.NsxtAlbController{
		Name:        check.TestName(),
		Url:         vcd.config.VCD.Nsxt.NsxtAlbControllerUrl,
		Username:    vcd.config.VCD.Nsxt.NsxtAlbControllerUser,
		Password:    vcd.config.VCD.Nsxt.NsxtAlbControllerPassword,
		LicenseType: "ENTERPRISE", // Not used since v37.0
	}

	newController, err := vcd.client.CreateNsxtAlbController(newControllerDef)
	check.Assert(err, IsNil)
	check.Assert(newController.NsxtAlbController.ID, Not(Equals), "")

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController + newController.NsxtAlbController.ID
	AddToCleanupListOpenApi(newController.NsxtAlbController.Name, check.TestName(), openApiEndpoint)

	return newController
}
