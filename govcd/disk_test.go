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
	disk := NewDisk(&vcd.client.Client)
	check.Assert(disk, NotNil)
}

// Test create independent disk
func (vcd *TestVCD) Test_DiskCreate(check *C) {
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        vcd.config.VCD.Disk.Name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: vcd.config.VCD.Disk.Description,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Clean up
	AddToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_DiskCreate")
}

// Test update independent disk
func (vcd *TestVCD) Test_DiskUpdate(check *C) {
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        vcd.config.VCD.Disk.Name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: vcd.config.VCD.Disk.Description,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        vcd.config.VCD.Disk.NameForUpdate,
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
		Description: vcd.config.VCD.Disk.DescriptionForUpdate,
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

	// Clean up
	AddToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_DiskUpdate")
}

// Test delete independent disk
func (vcd *TestVCD) Test_DiskDelete(check *C) {
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        vcd.config.VCD.Disk.Name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: vcd.config.VCD.Disk.Description,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Delete disk
	deleteDiskTask, err := disk.Delete()
	check.Assert(err, IsNil)

	err = deleteDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// Test refresh independent disk info
func (vcd *TestVCD) Test_DiskRefresh(check *C) {

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        vcd.config.VCD.Disk.Name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: vcd.config.VCD.Disk.Description,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Update disk
	newDiskInfo := &types.Disk{
		Name:        vcd.config.VCD.Disk.NameForUpdate,
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
		Description: vcd.config.VCD.Disk.DescriptionForUpdate,
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

	// Clean up
	AddToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_DiskRefresh")
}

// Test find disk attached VM
func (vcd *TestVCD) Test_DiskAttachedVM(check *C) {

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        vcd.config.VCD.Disk.Name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: vcd.config.VCD.Disk.Description,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Find VM
	vmType, _ := vcd.find_first_vm(vcd.find_first_vapp())
	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

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

	// Clean up
	AddToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_DiskAttachedVM")
}

// Test find Disk by Href in VDC struct
func (vcd *TestVCD) Test_VdcFindDiskByHREF(check *C) {

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        vcd.config.VCD.Disk.Name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: vcd.config.VCD.Disk.Description,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Find disk by href
	foundDisk, err := vcd.vdc.FindDiskByHREF(disk.Disk.HREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, foundDisk.Disk.Name)

	// Clean up
	AddToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_VdcFindDiskByHREF")
}

// Test find disk by href and vdc client
func (vcd *TestVCD) Test_FindDiskByHREF(check *C) {
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        vcd.config.VCD.Disk.Name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: vcd.config.VCD.Disk.Description,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Find disk by href
	foundDisk, err := FindDiskByHREF(vcd.vdc.client, disk.Disk.HREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk, NotNil)
	check.Assert(disk.Disk.Name, Equals, foundDisk.Disk.Name)

	// Clean up
	AddToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_FindDiskByHREF")
}

// Independent disk integration test
func (vcd *TestVCD) Test_Disk(check *C) {
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        vcd.config.VCD.Disk.Name,
		Size:        vcd.config.VCD.Disk.Size,
		Description: vcd.config.VCD.Disk.Description,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Find VM
	vmType, _ := vcd.find_first_vm(vcd.find_first_vapp())
	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

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
		Name:        vcd.config.VCD.Disk.NameForUpdate,
		Size:        vcd.config.VCD.Disk.SizeForUpdate,
		Description: vcd.config.VCD.Disk.DescriptionForUpdate,
	}

	updateTask, err := disk.Update(newDiskInfo)
	err = updateTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Refresh disk Info
	err = disk.Refresh()
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, newDiskInfo.Name)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

	// Clean up
	AddToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_Disk")
}
