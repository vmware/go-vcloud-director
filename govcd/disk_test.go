// +build disk functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"

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
	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("skipping test because disk size is 0")
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

	ctx := context.Background()
	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(ctx, diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

}

// Test update independent disk
func (vcd *TestVCD) Test_UpdateDisk(check *C) {
	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("skipping test because disk size is 0")
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

	ctx := context.Background()
	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(ctx, diskHREF)
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

	updateTask, err := disk.Update(ctx, newDiskInfo)
	check.Assert(err, IsNil)
	err = updateTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Refresh disk info, getting updated info
	err = disk.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, newDiskInfo.Name)
	check.Assert(disk.Disk.Size, Equals, newDiskInfo.Size)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

}

// Test delete independent disk
func (vcd *TestVCD) Test_DeleteDisk(check *C) {
	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("skipping test because disk size is 0")
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

	ctx := context.Background()
	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(ctx, diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Delete disk
	deleteDiskTask, err := disk.Delete(ctx)
	check.Assert(err, IsNil)

	err = deleteDiskTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
}

// Test refresh independent disk info
func (vcd *TestVCD) Test_RefreshDisk(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is 0")
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

	ctx := context.Background()
	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(ctx, diskHREF)
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

	updateTask, err := disk.Update(ctx, newDiskInfo)
	check.Assert(err, IsNil)
	err = updateTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Refresh disk info, getting updated info
	err = disk.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, newDiskInfo.Name)
	check.Assert(disk.Disk.Size, Equals, newDiskInfo.Size)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

}

// Test find disk attached VM
func (vcd *TestVCD) Test_AttachedVMDisk(check *C) {

	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is 0")
	}

	if vcd.skipVappTests {
		check.Skip("skipping test because vApp wasn't properly created")
	}
	ctx := context.Background()

	// Find VM
	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Ensure vApp and VM are suitable for this test
	// Disk attach and detach operations are not working if VM is suspended
	err := vcd.ensureVappIsSuitableForVMTest(ctx, vapp)
	check.Assert(err, IsNil)
	err = vcd.ensureVMIsSuitableForVMTest(ctx, vm)
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

	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(ctx, diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Attach disk
	attachDiskTask, err := vm.AttachDisk(ctx, &types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	})
	check.Assert(err, IsNil)

	err = attachDiskTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Get attached VM
	vmRef, err := disk.AttachedVM(ctx)
	check.Assert(err, IsNil)
	check.Assert(vmRef, NotNil)
	check.Assert(vmRef.Name, Equals, vm.VM.Name)

	// Detach disk
	err = vcd.detachIndependentDisk(ctx, Disk{disk.Disk, &vcd.client.Client})
	check.Assert(err, IsNil)
}

// Test find Disk by Href in VDC struct
func (vcd *TestVCD) Test_VdcFindDiskByHREF(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is 0")
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
	ctx := context.Background()

	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.FindDiskByHREF(ctx, diskHREF)

	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

}

// Test find disk by href and vdc client
func (vcd *TestVCD) Test_FindDiskByHREF(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is 0")
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
	ctx := context.Background()

	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.FindDiskByHREF(ctx, diskHREF)

	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Find disk by href
	foundDisk, err := FindDiskByHREF(ctx, vcd.vdc.client, disk.Disk.HREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, foundDisk.Disk.Name)

}

// Independent disk integration test
func (vcd *TestVCD) Test_Disk(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is 0")
	}

	if vcd.config.VCD.Disk.SizeForUpdate <= 0 {
		check.Skip("skipping test because disk update size is <= 0")
	}

	if vcd.skipVappTests {
		check.Skip("skipping test because vApp wasn't properly created")
	}
	ctx := context.Background()

	// Find VM
	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Ensure vApp and VM are suitable for this test
	// Disk attach and detach operations are not working if VM is suspended
	err := vcd.ensureVappIsSuitableForVMTest(ctx, vapp)
	check.Assert(err, IsNil)
	err = vcd.ensureVMIsSuitableForVMTest(ctx, vm)
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

	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.GetDiskByHref(ctx, diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Attach disk
	attachDiskTask, err := vm.AttachDisk(ctx, &types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	})
	check.Assert(err, IsNil)

	err = attachDiskTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Get attached VM
	vmRef, err := disk.AttachedVM(ctx)
	check.Assert(err, IsNil)
	check.Assert(vmRef, NotNil)
	check.Assert(vmRef.Name, Equals, vm.VM.Name)

	// Detach disk
	detachDiskTask, err := vm.DetachDisk(ctx, &types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	})
	check.Assert(err, IsNil)

	err = detachDiskTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        TestDisk,
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
		Description: TestDisk + "_Update",
	}

	updateTask, err := disk.Update(ctx, newDiskInfo)
	check.Assert(err, IsNil)

	err = updateTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Refresh disk Info
	err = disk.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, newDiskInfo.Name)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

}

// Test query disk
func (vcd *TestVCD) Test_QueryDisk(check *C) {

	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is 0")
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

	ctx := context.Background()
	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	diskRecord, err := vcd.vdc.QueryDisk(ctx, name)

	check.Assert(err, IsNil)
	check.Assert(diskRecord.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(diskRecord.Disk.SizeB, Equals, diskCreateParamsDisk.Size)
	check.Assert(diskRecord.Disk.Description, Equals, diskCreateParamsDisk.Description)
}

// Test query disk
func (vcd *TestVCD) Test_QueryDisks(check *C) {

	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is 0")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	name := "TestQueryDisks"

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: name,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}
	ctx := context.Background()

	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// create second disk with same name
	task, err = vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF = task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	diskRecords, err := vcd.vdc.QueryDisks(ctx, name)

	check.Assert(err, IsNil)
	check.Assert(len(*diskRecords), Equals, 2)
	check.Assert((*diskRecords)[0].Name, Equals, diskCreateParamsDisk.Name)
	check.Assert((*diskRecords)[0].SizeB, Equals, int64(diskCreateParamsDisk.Size))
}

// Tests Disk list retrieval by name, by ID
func (vcd *TestVCD) Test_GetDisks(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("skipping test because disk size is 0")
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
	ctx := context.Background()

	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	err = vcd.vdc.Refresh(ctx)
	check.Assert(err, IsNil)
	diskList, err := vcd.vdc.GetDisksByName(ctx, diskName, false)
	check.Assert(err, IsNil)
	check.Assert(diskList, NotNil)
	check.Assert(len(*diskList), Equals, 1)
	check.Assert((*diskList)[0].Disk.Name, Equals, diskName)
	check.Assert((*diskList)[0].Disk.Description, Equals, diskName+"Description")

	disk, err := vcd.vdc.GetDiskById(ctx, (*diskList)[0].Disk.Id, false)
	check.Assert(err, IsNil)
	check.Assert(disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, diskName)
	check.Assert(disk.Disk.Description, Equals, diskName+"Description")

	diskList, err = vcd.vdc.GetDisksByName(ctx, "INVALID", false)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(diskList, IsNil)

	// test two disk with same name
	task, err = vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF = task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	err = vcd.vdc.Refresh(ctx)
	check.Assert(err, IsNil)
	diskList, err = vcd.vdc.GetDisksByName(ctx, diskName, false)
	check.Assert(err, IsNil)
	check.Assert(diskList, NotNil)
	check.Assert(len(*diskList), Equals, 2)

}

// Tests Disk list retrieval by name, by ID
func (vcd *TestVCD) Test_GetDiskByHref(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Disk.Size == 0 {
		check.Skip("skipping test because disk size is 0")
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

	ctx := context.Background()
	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	// Wait for disk creation complete
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	disk, err := vcd.vdc.GetDiskByHref(ctx, diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, diskName)
	check.Assert(disk.Disk.Description, Equals, diskName+"Description")

	invalidDiskHREF := diskHREF + "1"
	disk, err = vcd.vdc.GetDiskByHref(ctx, invalidDiskHREF)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(disk, IsNil)
}
