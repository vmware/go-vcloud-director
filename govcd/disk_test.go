// +build disk functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test init independent disk struct
func (vcd *TestVCD) Test_NewDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	disk := NewDisk(&vcd.client.Client)
	check.Assert(disk, NotNil)
}

// Test create independent disk
func (vcd *TestVCD) Test_CreateDisk(check *C) {
	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("Skipping test because disk size is <= 0")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestCreateDisk,
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

}

// Test update independent disk
func (vcd *TestVCD) Test_UpdateDisk(check *C) {
	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("skipping test because disk size is <= 0")
	}

	if vcd.config.VCD.Disk.SizeForUpdate <= 0 {
		check.Skip("skipping test because disk update size is <= 0")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestUpdateDisk,
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        TestUpdateDisk,
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
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
	check.Assert(disk.Disk.Size, Equals, newDiskInfo.Size)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

}

// Test delete independent disk
func (vcd *TestVCD) Test_DeleteDisk(check *C) {
	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("skipping test because disk size is <= 0")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	var err error

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestDeleteDisk,
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Delete disk
	deleteDiskTask, err := disk.Delete()
	check.Assert(err, IsNil)

	err = deleteDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// Test refresh independent disk info
func (vcd *TestVCD) Test_RefreshDisk(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is <= 0")
	}

	if vcd.config.VCD.Disk.SizeForUpdate <= 0 {
		check.Skip("skipping test because disk update size is <= 0")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestRefreshDisk,
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        TestRefreshDisk,
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
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
	check.Assert(disk.Disk.Size, Equals, newDiskInfo.Size)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

}

// Test find disk attached VM
func (vcd *TestVCD) Test_AttachedVMDisk(check *C) {

	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is <= 0")
	}

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
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
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
	err = vcd.detachIndependentDisk(Disk{disk.Disk, &vcd.client.Client})
	check.Assert(err, IsNil)
}

// Checks whether an independent disk is attached to a VM, and detaches it
func (vcd *TestVCD) detachIndependentDisk(disk Disk) error {

	// See if the disk is attached to the VM
	vmRef, err := disk.AttachedVM()
	if err != nil {
		return err
	}
	// If the disk is attached to the VM, detach disk from the VM
	if vmRef != nil {

		vm, err := vcd.client.Client.GetVMByHref(vmRef.HREF)
		if err != nil {
			return err
		}

		// Detach the disk from VM
		task, err := vm.DetachDisk(&types.DiskAttachOrDetachParams{
			Disk: &types.Reference{
				HREF: disk.Disk.HREF,
			},
		})
		if err != nil {
			return err
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return err
		}
	}
	return nil
}

// Test find Disk by Href in VDC struct
func (vcd *TestVCD) Test_VdcFindDiskByHREF(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is <= 0")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestVdcFindDiskByHREF,
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

}

// Test find disk by href and vdc client
func (vcd *TestVCD) Test_FindDiskByHREF(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("Skipping test because disk size is <= 0")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestFindDiskByHREF,
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Find disk by href
	foundDisk, err := FindDiskByHREF(vcd.vdc.client, disk.Disk.HREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, foundDisk.Disk.Name)

}

// Independent disk integration test
func (vcd *TestVCD) Test_Disk(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is <= 0")
	}

	if vcd.config.VCD.Disk.SizeForUpdate <= 0 {
		check.Skip("skipping test because disk update size is <= 0")
	}

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
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
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
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
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

	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("Skipping test because disk size is <= 0")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	name := "TestQueryDisk"

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        name,
		Size:        vcd.config.VCD.Disk.Size,
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
	check.Assert(diskRecord.Disk.SizeB, Equals, int64(diskCreateParamsDisk.Size))

	// vCD version >= 9.5. Earlier versions don't return Description
	if vcd.client.APIVCDMaxVersionIs(">= 31.0") {
		check.Assert(diskRecord.Disk.Description, Equals, diskCreateParamsDisk.Description)
	} else {
		fmt.Printf("%s: skipping disk description check (not available in vCD < 9.5) \n", check.TestName())
	}

}

// Tests Disk list retrieval by name, by ID
func (vcd *TestVCD) Test_GetDisks(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("Skipping test because disk size is <= 0")
	}

	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_GetDisk: VDC name not given")
		return
	}

	diskName := "TestGetDisk"
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        diskName,
		Size:        vcd.config.VCD.Disk.Size,
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

	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("Skipping test because disk size is <= 0")
	}

	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_GetDisk: VDC name not given")
		return
	}

	diskName := "TestGetDiskByHref"
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        diskName,
		Size:        vcd.config.VCD.Disk.Size,
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

	invalidDiskHREF := diskHREF + "1"
	disk, err = vcd.vdc.GetDiskByHref(invalidDiskHREF)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(disk, IsNil)
}
