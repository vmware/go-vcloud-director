/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 * Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) find_first_vm(vapp VApp) (types.VM, string) {
	for _, vm := range vapp.VApp.Children.VM {
		if vm.Name != "" {
			return *vm, vm.Name
		}
	}
	return types.VM{}, ""
}

func (vcd *TestVCD) find_first_vapp() VApp {
	client := vcd.client
	config := vcd.config
	org, err := GetOrgByName(client, config.VCD.Org)
	if err != nil {
		fmt.Println(err)
		return VApp{}
	}
	vdc, err := org.GetVdcByName(config.VCD.Vdc)
	if err != nil {
		fmt.Println(err)
		return VApp{}
	}
	wanted_vapp := vcd.vapp.VApp.Name
	vapp_name := ""
	for _, res := range vdc.Vdc.ResourceEntities {
		for _, item := range res.ResourceEntity {
			// Finding a named vApp, if it was defined in config
			if wanted_vapp != "" {
				if item.Name == wanted_vapp {
					vapp_name = item.Name
					break
				}
			} else {
				// Otherwise, we get the first vApp from the vDC list
				if item.Type == "application/vnd.vmware.vcloud.vApp+xml" {
					vapp_name = item.Name
					break
				}
			}
		}
	}
	if wanted_vapp == "" {
		return VApp{}
	}
	vapp, _ := vdc.FindVAppByName(vapp_name)
	return vapp
}

func (vcd *TestVCD) Test_FindVMByHREF(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}

	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.find_first_vapp()
	if vapp.VApp.Name == "" {
		check.Skip("Disabled: No suitable vApp found in vDC")
	}
	vm, vm_name := vcd.find_first_vm(vapp)
	if vm.Name == "" {
		check.Skip("Disabled: No suitable VM found in vDC")
	}

	vm_href := vm.HREF
	new_vm, err := vcd.client.FindVMByHREF(vm_href)

	check.Assert(err, IsNil)
	check.Assert(new_vm.VM.Name, Equals, vm_name)
	check.Assert(new_vm.VM.VirtualHardwareSection.Item, NotNil)
}

// Test attach disk to VM and detach disk from VM
func (vcd *TestVCD) Test_VMAttachOrDetachDisk(check *C) {
	// Create Disk
	diskCreateParamsDisk := &types.Disk{
		Name:        "TestDisk",
		Size:        10240,
		Description: "Test Disk Description",
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

	// Clean up - delete disk
	deleteDiskTask, err := disk.Delete()
	check.Assert(err, IsNil)

	err = deleteDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// Test attach disk to VM
func (vcd *TestVCD) Test_VMAttachDisk(check *C) {
	// Create Disk
	diskCreateParamsDisk := &types.Disk{
		Name:        "TestDisk",
		Size:        10240,
		Description: "Test Disk Description",
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

	// Clean up - detach disk
	detachDiskTask, err := vm.DetachDisk(&types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	})
	check.Assert(err, IsNil)

	err = detachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Clean up - delete disk
	deleteDiskTask, err := disk.Delete()
	check.Assert(err, IsNil)

	err = deleteDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// Test detach disk from VM
func (vcd *TestVCD) Test_VMDetachDisk(check *C) {
	// Create Disk
	diskCreateParamsDisk := &types.Disk{
		Name:        "TestDisk",
		Size:        10240,
		Description: "Test Disk Description",
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

	// Clean up - delete disk
	deleteDiskTask, err := disk.Delete()
	check.Assert(err, IsNil)

	err = deleteDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
}
