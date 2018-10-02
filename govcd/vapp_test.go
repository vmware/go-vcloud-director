/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/types/v56"

	. "gopkg.in/check.v1"
)

// Tests the helper function getParentVDC with the vapp
// created at the start of testing
func (vcd *TestVCD) TestGetParentVDC(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	vapp, err := vcd.vdc.FindVAppByName(vcd.vapp.VApp.Name)
	check.Assert(err, IsNil)

	vdc, err := vapp.getParentVDC()

	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.vdc.Vdc.Name)
}

func (vcd *TestVCD) createTestVapp(name string) (VApp, error) {
	// Populate OrgVDCNetwork
	networks := []*types.OrgVDCNetwork{}
	net, err := vcd.vdc.FindVDCNetwork(vcd.config.VCD.Network)
	if err != nil {
		return VApp{}, fmt.Errorf("error finding network : %v", err)
	}
	networks = append(networks, net.OrgVDCNetwork)
	// Populate Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	if err != nil || cat == (Catalog{}) {
		return VApp{}, fmt.Errorf("error finding catalog : %v", err)
	}
	// Populate Catalog Item
	catitem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.Catalogitem)
	if err != nil {
		return VApp{}, fmt.Errorf("error finding catalog item : %v", err)
	}
	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	if err != nil {
		return VApp{}, fmt.Errorf("error finding vapptemplate : %v", err)
	}
	// Get StorageProfileReference
	storageprofileref, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	if err != nil {
		return VApp{}, fmt.Errorf("error finding storage profile: %v", err)
	}
	// Compose VApp
	task, err := vcd.vdc.ComposeVApp(networks, vapptemplate, storageprofileref, name, "description", true)
	if err != nil {
		return VApp{}, fmt.Errorf("error composing vapp: %v", err)
	}
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(name, "vapp", "", "createTestVapp")
	err = task.WaitTaskCompletion()
	if err != nil {
		return VApp{}, fmt.Errorf("error composing vapp: %v", err)
	}
	// Get VApp
	vapp, err := vcd.vdc.FindVAppByName(name)
	if err != nil {
		return VApp{}, fmt.Errorf("error getting vapp: %v", err)
	}
	return vapp, err
}

// Tests Powering On and Powering Off a VApp. Also tests Deletion
// of a VApp
func (vcd *TestVCD) Test_PowerOn(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_Reboot(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, _ := vcd.vapp.PowerOn()
	_ = task.WaitTaskCompletion()
	task, err := vcd.vapp.Reboot()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

// TODO: Add a check checking if the ovf was set properly
func (vcd *TestVCD) Test_SetOvf(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	var test map[string]string
	test = make(map[string]string)
	test["guestinfo.hostname"] = "testhostname"
	task, err := vcd.vapp.SetOvf(test)

	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

// TODO: Add a check checking if the metadata was added to the vapp
func (vcd *TestVCD) Test_AddMetadata(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Add Metadata
	task, err := vcd.vapp.AddMetadata("key", "value")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

// TODO: Add a check checking if the customization script ran
func (vcd *TestVCD) Test_RunCustomizationScript(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Run Script on Test Vapp
	task, err := vcd.vapp.RunCustomizationScript("computername", "this is my script")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking if the cpu count did change
func (vcd *TestVCD) Test_ChangeCPUcount(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.ChangeCPUcount(1)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking if the vapp uses the new memory size
func (vcd *TestVCD) Test_ChangeMemorySize(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.ChangeMemorySize(512)

	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking the if the vapp uses the new storage profile
func (vcd *TestVCD) Test_ChangeStorageProfile(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	if vcd.config.VCD.StorageProfile.SP2 == "" {
		check.Skip("Skipping test because second storage profile not given")
	}
	task, err := vcd.vapp.ChangeStorageProfile(vcd.config.VCD.StorageProfile.SP2)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// TODO: Add a check checking the vm name
func (vcd *TestVCD) Test_ChangeVMName(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.ChangeVMName("My-vm")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_Reset(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, _ := vcd.vapp.PowerOn()
	_ = task.WaitTaskCompletion()
	task, err := vcd.vapp.Reset()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_Suspend(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, _ := vcd.vapp.PowerOn()
	_ = task.WaitTaskCompletion()
	task, err := vcd.vapp.Suspend()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_Shutdown(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, _ := vcd.vapp.PowerOn()
	_ = task.WaitTaskCompletion()
	task, err := vcd.vapp.Shutdown()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

func (vcd *TestVCD) Test_Deploy(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Deploy
	task, err := vcd.vapp.Deploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_PowerOff(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, _ := vcd.vapp.PowerOn()
	_ = task.WaitTaskCompletion()
	task, err := vcd.vapp.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: EVENTUALLY REMOVE THIS REDEPLOY
func (vcd *TestVCD) Test_Undeploy(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Check if the vapp has been deployed yet
	err := vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	if !vcd.vapp.VApp.Deployed {
		task, err := vcd.vapp.Deploy()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}
	// Undeploy
	task, err := vcd.vapp.Undeploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	// Deploy
	// For some reason it will not work without redeploying
	// TODO: EVENTUALLY REMOVE THIS REDEPLOY
	task, err = vcd.vapp.Deploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}
