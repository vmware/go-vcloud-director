/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/types/v56"
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

	// Defer prepend the disk info to cleanup list until the function returns
	defer PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, diskHREF), "disk", "", check.TestName())

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

	// Defer prepend the disk info to cleanup list until the function returns
	defer PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, diskHREF), "disk", "", check.TestName())

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

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        TestUpdateDisk,
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
		Description: TestUpdateDisk + "_Update",
	}

	updateTask, err := disk.Update(newDiskInfo)
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

	// Defer prepend the disk info to cleanup list until the function returns
	defer func() {
		if err != nil {
			PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, diskHREF), "disk", "", check.TestName())
		}
	}()

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

	// Defer prepend the disk info to cleanup list until the function returns
	defer PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, diskHREF), "disk", "", check.TestName())

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

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        TestRefreshDisk,
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
		Description: TestRefreshDisk + "_Update",
	}

	updateTask, err := disk.Update(newDiskInfo)
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
	vmType, vmName := vcd.find_first_vm(vcd.find_first_vapp())
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

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

	// Defer prepend the disk info to cleanup list until the function returns
	defer PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, diskHREF), "disk", "", check.TestName())

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

	// Defer prepend the disk info to cleanup list until the function returns
	defer PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, diskHREF), "disk", "", check.TestName())

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

	// Defer prepend the disk info to cleanup list until the function returns
	defer PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, diskHREF), "disk", "", check.TestName())

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
	vmType, vmName := vcd.find_first_vm(vcd.find_first_vapp())
	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

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

	// Defer prepend the disk info to cleanup list until the function returns
	defer PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, diskHREF), "disk", "", check.TestName())

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
