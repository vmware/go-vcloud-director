//go:build disk || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Test init independent disk struct
func (vcd *TestVCD) Test_NewDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	disk := NewDisk(&vcd.client.Client)
	check.Assert(disk, NotNil)
}

// Test create independent disk
func (vcd *TestVCD) Test_CreateDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestCreateDisk,
		SizeMb:      11,
		Description: TestCreateDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.SizeMb, Equals, diskCreateParamsDisk.SizeMb)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)
	if vcd.client.Client.APIVCDMaxVersionIs(">= 36") {
		check.Assert(disk.Disk.UUID, Not(Equals), "")
		check.Assert(disk.Disk.SharingType, Equals, "None")
		check.Assert(disk.Disk.Encrypted, Equals, false)
	}

}

// Test update independent disk
func (vcd *TestVCD) Test_UpdateDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestUpdateDisk,
		SizeMb:      99,
		Description: TestUpdateDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.SizeMb, Equals, diskCreateParamsDisk.SizeMb)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        TestUpdateDisk,
		SizeMb:      102,
		Description: TestUpdateDisk + "_Update",
	}

	updateTask, err := disk.Update(newDiskInfo)
	check.Assert(err, IsNil)
	err = updateTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Refresh disk info, getting updated info
	err = disk.Refresh()
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, newDiskInfo.Name)
	check.Assert(disk.Disk.SizeMb, Equals, newDiskInfo.SizeMb)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

}

// Test delete independent disk
func (vcd *TestVCD) Test_DeleteDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	var err error

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestDeleteDisk,
		SizeMb:      1,
		Description: TestDeleteDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.SizeMb, Equals, diskCreateParamsDisk.SizeMb)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Delete disk
	deleteDiskTask, err := disk.Delete()
	check.Assert(err, IsNil)

	err = deleteDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// Test refresh independent disk info
func (vcd *TestVCD) Test_RefreshDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestRefreshDisk,
		SizeMb:      43,
		Description: TestRefreshDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.SizeMb, Equals, diskCreateParamsDisk.SizeMb)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        TestRefreshDisk,
		SizeMb:      43,
		Description: TestRefreshDisk + "_Update",
	}

	updateTask, err := disk.Update(newDiskInfo)
	check.Assert(err, IsNil)
	err = updateTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Refresh disk info, getting updated info
	err = disk.Refresh()
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, newDiskInfo.Name)
	check.Assert(disk.Disk.SizeMb, Equals, newDiskInfo.SizeMb)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

}

// Test find disk attached VM
func (vcd *TestVCD) Test_AttachedVMDisk(check *C) {
	if vcd.skipVappTests {
		check.Skip("skipping test because vApp wasn't properly created")
	}

	// Find VM
	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Ensure vApp and VM are suitable for this test
	// Disk attach and detach operations are not working if VM is suspended
	err := vcd.ensureVappIsSuitableForVMTest(vapp)
	check.Assert(err, IsNil)
	err = vcd.ensureVMIsSuitableForVMTest(vm)
	check.Assert(err, IsNil)

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestAttachedVMDisk,
		SizeMb:      210,
		Description: TestAttachedVMDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.SizeMb, Equals, diskCreateParamsDisk.SizeMb)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Attach disk
	attachDiskTask, err := vm.AttachDisk(&types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	})
	check.Assert(err, IsNil)

	err = attachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Get attached VM
	vmRef, err := disk.AttachedVM()
	check.Assert(err, IsNil)
	check.Assert(vmRef, NotNil)
	check.Assert(vmRef.Name, Equals, vm.VM.Name)

	// Get attached VM
	vmHrefs, err := disk.GetAttachedVmsHrefs()
	check.Assert(err, IsNil)
	check.Assert(vmHrefs, NotNil)
	check.Assert(len(vmHrefs), Equals, 1)
	check.Assert(vmHrefs[0], Equals, vm.VM.HREF)

	// Detach disk
	err = vcd.detachIndependentDisk(Disk{disk.Disk, &vcd.client.Client})
	check.Assert(err, IsNil)
}

// Test find Disk by Href in VDC struct
func (vcd *TestVCD) Test_VdcFindDiskByHREF(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestVdcFindDiskByHREF,
		SizeMb:      2,
		Description: TestVdcFindDiskByHREF,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.FindDiskByHREF(diskHREF)

	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.SizeMb, Equals, diskCreateParamsDisk.SizeMb)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

}

// Test find disk by href and vdc client
func (vcd *TestVCD) Test_FindDiskByHREF(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestFindDiskByHREF,
		SizeMb:      3,
		Description: TestFindDiskByHREF,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.FindDiskByHREF(diskHREF)

	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.SizeMb, Equals, diskCreateParamsDisk.SizeMb)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Find disk by href
	foundDisk, err := FindDiskByHREF(vcd.vdc.client, disk.Disk.HREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, foundDisk.Disk.Name)

}

// Independent disk integration test
func (vcd *TestVCD) Test_Disk(check *C) {
	if vcd.skipVappTests {
		check.Skip("skipping test because vApp wasn't properly created")
	}

	// Find VM
	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Ensure vApp and VM are suitable for this test
	// Disk attach and detach operations are not working if VM is suspended
	err := vcd.ensureVappIsSuitableForVMTest(vapp)
	check.Assert(err, IsNil)
	err = vcd.ensureVMIsSuitableForVMTest(vm)
	check.Assert(err, IsNil)

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestDisk,
		SizeMb:      14,
		Description: TestDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.SizeMb, Equals, diskCreateParamsDisk.SizeMb)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Attach disk
	attachDiskTask, err := vm.AttachDisk(&types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	})
	check.Assert(err, IsNil)

	err = attachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Get attached VM
	vmRef, err := disk.AttachedVM()
	check.Assert(err, IsNil)
	check.Assert(vmRef, NotNil)
	check.Assert(vmRef.Name, Equals, vm.VM.Name)

	// Detach disk
	detachDiskTask, err := vm.DetachDisk(&types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	})
	check.Assert(err, IsNil)

	err = detachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        TestDisk,
		SizeMb:      41,
		Description: TestDisk + "_Update",
	}

	updateTask, err := disk.Update(newDiskInfo)
	check.Assert(err, IsNil)

	err = updateTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Refresh disk Info
	err = disk.Refresh()
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, newDiskInfo.Name)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

}

// Test query disk
func (vcd *TestVCD) Test_QueryDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	name := "TestQueryDisk"

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        name,
		SizeMb:      1,
		Description: name,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	diskRecord, err := vcd.vdc.QueryDisk(name)

	check.Assert(err, IsNil)
	check.Assert(diskRecord.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(diskRecord.Disk.SizeMb, Equals, int64(diskCreateParamsDisk.SizeMb))
	check.Assert(diskRecord.Disk.Description, Equals, diskCreateParamsDisk.Description)
}

// Test query disk
func (vcd *TestVCD) Test_QueryDisks(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	name := "TestQueryDisks"

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        name,
		SizeMb:      22,
		Description: name,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// create second disk with same name
	task, err = vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF = task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	diskRecords, err := vcd.vdc.QueryDisks(name)

	check.Assert(err, IsNil)
	check.Assert(len(*diskRecords), Equals, 2)
	check.Assert((*diskRecords)[0].Name, Equals, diskCreateParamsDisk.Name)
	check.Assert((*diskRecords)[0].SizeMb, Equals, int64(diskCreateParamsDisk.SizeMb))
	if vcd.client.Client.APIVCDMaxVersionIs(">= 36") {
		check.Assert((*diskRecords)[0].UUID, Not(Equals), "")
		check.Assert((*diskRecords)[0].SharingType, Equals, "None")
		check.Assert((*diskRecords)[0].Encrypted, Equals, false)
	}
}

// Tests Disk list retrieval by name, by ID
func (vcd *TestVCD) Test_GetDisks(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_GetDisk: VDC name not given")
		return
	}

	diskName := "TestGetDisk"
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        diskName,
		SizeMb:      12,
		Description: diskName + "Description",
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	err = vcd.vdc.Refresh()
	check.Assert(err, IsNil)
	diskList, err := vcd.vdc.GetDisksByName(diskName, false)
	check.Assert(err, IsNil)
	check.Assert(diskList, NotNil)
	check.Assert(len(*diskList), Equals, 1)
	check.Assert((*diskList)[0].Disk.Name, Equals, diskName)
	check.Assert((*diskList)[0].Disk.Description, Equals, diskName+"Description")

	disk, err := vcd.vdc.GetDiskById((*diskList)[0].Disk.Id, false)
	check.Assert(err, IsNil)
	check.Assert(disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, diskName)
	check.Assert(disk.Disk.Description, Equals, diskName+"Description")

	diskList, err = vcd.vdc.GetDisksByName("INVALID", false)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(diskList, IsNil)

	// test two disk with same name
	task, err = vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF = task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	err = vcd.vdc.Refresh()
	check.Assert(err, IsNil)
	diskList, err = vcd.vdc.GetDisksByName(diskName, false)
	check.Assert(err, IsNil)
	check.Assert(diskList, NotNil)
	check.Assert(len(*diskList), Equals, 2)

}

// Tests Disk list retrieval by name, by ID
func (vcd *TestVCD) Test_GetDiskByHref(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_GetDisk: VDC name not given")
		return
	}

	diskName := "TestGetDiskByHref"
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        diskName,
		SizeMb:      2048,
		Description: diskName + "Description",
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	disk, err := vcd.vdc.GetDiskByHref(diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, diskName)
	check.Assert(disk.Disk.Description, Equals, diskName+"Description")

	// Creating HREF with fake UUID
	uuid, err := GetUuidFromHref(diskHREF, true)
	check.Assert(err, IsNil)
	invalidDiskHREF := strings.ReplaceAll(diskHREF, uuid, "1abcbdb3-1111-1111-a1c2-85d261e22fcf")
	disk, err = vcd.vdc.GetDiskByHref(invalidDiskHREF)
	check.Assert(err, NotNil)
	if vcd.client.Client.IsSysAdmin {
		check.Assert(IsNotFound(err), Equals, true)
	} else {
		// The errors returned for non-existing disk are different for system administrator and org user
		check.Assert(strings.Contains(err.Error(), "API Error: 403:"), Equals, true)
	}
	check.Assert(disk, IsNil)
}
