/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 * Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_Disk(check *C) {
	// Create Disk
	diskCreateParamsDisk := &types.DiskCreateParamsDisk{
		Name:        "TestDisk",
		Size:        10240,
		Description: "Test Disk Description",
	}

	disk, err := vcd.vdc.CreateDisk(diskCreateParamsDisk)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, disk.Disk.Size)
	check.Assert(disk.Disk.Description, Equals, disk.Disk.Description)

	task := NewTask(vcd.vdc.client)
	for _, taskItem := range disk.Disk.Tasks {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Find VM
	vmType, _ := vcd.find_first_vm(vcd.find_first_vapp())
	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Attach Disk
	attachDiskTask, err := vm.AttachDisk(disk)
	check.Assert(err, IsNil)

	err = attachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Get Attached VM
	vmRef, err := disk.AttachedVM()
	check.Assert(err, IsNil)
	check.Assert(vmRef, NotNil)
	check.Assert(vmRef.Name, Equals, vm.VM.Name)

	// Detach Disk
	detachDiskTask, err := vm.DetachDisk(disk)
	check.Assert(err, IsNil)

	err = detachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	newDiskInfo := &types.DiskType{
		Name:        "HelloDisk",
		Description: "Hello Disk Description",
	}

	// Update Disk
	updateTask, err := disk.Update(newDiskInfo)
	err = updateTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Refresh Disk Info
	err = disk.Refresh()
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, newDiskInfo.Name)
	check.Assert(disk.Disk.Description, Equals, newDiskInfo.Description)

	// Delete Disk
	deleteDiskTask, err := disk.Delete()
	check.Assert(err, IsNil)

	err = deleteDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
}
