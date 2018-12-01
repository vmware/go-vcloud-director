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
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is <= 0")
	}

	// Find VM
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	vmType, vmName := vcd.find_first_vm(vcd.find_first_vapp())
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestVMAttachOrDetachDisk,
		Size:        vcd.config.VCD.Disk.Size,
		Description: TestVMAttachOrDetachDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	tasks, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	diskHREF := ""
	for _, taskItem := range tasks {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)

		if task.Task.Owner.Type == types.MimeDisk {
			diskHREF = task.Task.Owner.HREF
		}
	}

	// Verify created disk
	check.Assert(diskHREF, Not(Equals), "")
	disk, err := vcd.vdc.FindDiskByHREF(diskHREF)
	check.Assert(err, IsNil)
	check.Assert(disk.Disk.Name, Equals, diskCreateParamsDisk.Name)
	check.Assert(disk.Disk.Size, Equals, diskCreateParamsDisk.Size)
	check.Assert(disk.Disk.Description, Equals, diskCreateParamsDisk.Description)

	// Attach disk
	attachDiskTask, err := vm.attachOrDetachDisk(&types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	}, types.RelDiskAttach)
	check.Assert(err, IsNil)

	err = attachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Get attached VM
	vmRef, err := disk.AttachedVM()
	check.Assert(err, IsNil)
	check.Assert(vmRef, NotNil)
	check.Assert(vmRef.Name, Equals, vm.VM.Name)

	// Detach disk
	detachDiskTask, err := vm.attachOrDetachDisk(&types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	}, types.RelDiskDetach)
	check.Assert(err, IsNil)

	err = detachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Clean up
	PrependToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_VMAttachOrDetachDisk")
}

// Test attach disk to VM
func (vcd *TestVCD) Test_VMAttachDisk(check *C) {
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
		Name:        TestVMAttachDisk,
		Size:        vcd.config.VCD.Disk.Size,
		Description: TestVMAttachDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	tasks, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	diskHREF := ""
	for _, taskItem := range tasks {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)

		if task.Task.Owner.Type == types.MimeDisk {
			diskHREF = task.Task.Owner.HREF
		}
	}

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

	// Clean up
	PrependToCleanupList(fmt.Sprintf("%s|%s", disk.Disk.Name, disk.Disk.HREF), "disk", "", "Test_VMAttachDisk")
}

// Test detach disk from VM
func (vcd *TestVCD) Test_VMDetachDisk(check *C) {

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
		Name:        TestVMDetachDisk,
		Size:        vcd.config.VCD.Disk.Size,
		Description: TestVMDetachDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	tasks, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	// Wait for disk creation complete
	task := NewTask(vcd.vdc.client)
	diskHREF := ""
	for _, taskItem := range tasks {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)

		if task.Task.Owner.Type == types.MimeDisk {
			diskHREF = task.Task.Owner.HREF
		}
	}

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

	// Clean up
	PrependToCleanupList(fmt.Sprintf("%s|%s", diskCreateParamsDisk.Name, disk.Disk.HREF), "disk", "", "Test_VMDetachDisk")
}
