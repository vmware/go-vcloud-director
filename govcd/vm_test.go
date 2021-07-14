// +build vm functional ALL

/*
* Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
* Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kr/pretty"
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func init() {
	testingTags["vm"] = "vm_test.go"
}

func (vcd *TestVCD) Test_FindVMByHREF(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}

	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.findFirstVapp()
	if vapp.VApp.Name == "" {
		check.Skip("Disabled: No suitable vApp found in vDC")
	}
	vm, vmName := vcd.findFirstVm(vapp)
	if vm.Name == "" {
		check.Skip("Disabled: No suitable VM found in vDC")
	}

	newVM, err := vcd.client.Client.GetVMByHref(vm.HREF)
	check.Assert(err, IsNil)
	check.Assert(newVM.VM.Name, Equals, vmName)
	check.Assert(newVM.VM.VirtualHardwareSection.Item, NotNil)
}

// Test attach disk to VM and detach disk from VM
func (vcd *TestVCD) Test_VMAttachOrDetachDisk(check *C) {
	// Find VM
	if vcd.vapp != nil && vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

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
		Name:        TestVMAttachOrDetachDisk,
		SizeMb:      1,
		Description: TestVMAttachOrDetachDisk,
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

}

// Test attach disk to VM
func (vcd *TestVCD) Test_VMAttachDisk(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
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

	// Discard vApp suspension
	// Disk attach and detach operations are not working if vApp is suspended
	err := vcd.ensureVappIsSuitableForVMTest(vapp)
	check.Assert(err, IsNil)
	err = vcd.ensureVMIsSuitableForVMTest(vm)
	check.Assert(err, IsNil)

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestVMAttachDisk,
		SizeMb:      1,
		Description: TestVMAttachDisk,
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

	// Cleanup: Detach disk
	detachDiskTask, err := vm.attachOrDetachDisk(&types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	}, types.RelDiskDetach)
	check.Assert(err, IsNil)

	err = detachDiskTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

}

// Test detach disk from VM
func (vcd *TestVCD) Test_VMDetachDisk(check *C) {
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
		Name:        TestVMDetachDisk,
		SizeMb:      1,
		Description: TestVMDetachDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	// Defer prepend the disk info to cleanup list until the function returns
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

}

// Test Insert or Eject Media for VM
func (vcd *TestVCD) Test_HandleInsertOrEjectMedia(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	itemName := "TestHandleInsertOrEjectMedia"

	// Find VApp
	if vcd.vapp != nil && vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Upload Media
	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_HandleInsertOrEjectMedia")

	catalog, err = vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	media, err := catalog.GetMediaByName(itemName, false)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)

	insertMediaTask, err := vm.HandleInsertMedia(vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	err = insertMediaTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, true)

	vm, err = vm.HandleEjectMediaAndAnswer(vcd.org, vcd.config.VCD.Catalog.Name, itemName, true)
	check.Assert(err, IsNil)

	//verify
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, false)
}

// Test Insert or Eject Media for VM
func (vcd *TestVCD) Test_InsertOrEjectMedia(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	// Skipping this test due to a bug in vCD. VM refresh status returns old state, though eject task is finished.
	if vcd.client.Client.APIVCDMaxVersionIs(">= 32.0, <= 33.0") {
		check.Skip("Skipping test because this vCD version has a bug")
	}

	itemName := "TestInsertOrEjectMedia"

	// Find VApp
	if vcd.vapp != nil && vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Upload Media
	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_InsertOrEjectMedia")

	catalog, err = vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	media, err := catalog.GetMediaByName(itemName, false)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)

	// Insert Media
	insertMediaTask, err := vm.insertOrEjectMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.Media.HREF,
			Name: media.Media.Name,
			ID:   media.Media.ID,
			Type: media.Media.Type,
		},
	}, types.RelMediaInsertMedia)
	check.Assert(err, IsNil)

	err = insertMediaTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)

	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, true)

	// Insert Media
	ejectMediaTask, err := vm.insertOrEjectMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.Media.HREF,
		},
	}, types.RelMediaEjectMedia)
	check.Assert(err, IsNil)

	err = ejectMediaTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, false)
}

// Test Insert or Eject Media for VM
func (vcd *TestVCD) Test_AnswerVmQuestion(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	itemName := "TestAnswerVmQuestion"

	// Find VApp
	if vcd.vapp != nil && vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Upload Media
	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_AnswerVmQuestion")

	catalog, err = vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	media, err := catalog.GetMediaByName(itemName, false)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)

	err = vm.Refresh()
	check.Assert(err, IsNil)

	insertMediaTask, err := vm.HandleInsertMedia(vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	err = insertMediaTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, true)

	ejectMediaTask, err := vm.HandleEjectMedia(vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	for i := 0; i < 10; i++ {
		question, err := vm.GetQuestion()
		check.Assert(err, IsNil)

		if question.QuestionId != "" && strings.Contains(question.Question, "Disconnect anyway and override the lock?") {
			err = vm.AnswerQuestion(question.QuestionId, 0)
			check.Assert(err, IsNil)
		}
		time.Sleep(time.Second * 3)
	}

	err = ejectMediaTask.Task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, false)
}

func (vcd *TestVCD) Test_VMChangeCPUCountWithCore(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	currentCpus := int64(0)
	currentCores := 0

	// save current values
	if nil != vcd.vapp.VApp.Children.VM[0] && nil != vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection && nil != vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection.Item {
		for _, item := range vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection.Item {
			if item.ResourceType == types.ResourceTypeProcessor {
				currentCpus = item.VirtualQuantity
				currentCores = item.CoresPerSocket
				break
			}
		}
	}

	check.Assert(0, Not(Equals), currentCpus)
	check.Assert(0, Not(Equals), currentCores)

	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	cores := 2
	cpuCount := int64(4)

	task, err := vm.ChangeCPUCountWithCore(int(cpuCount), &cores)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vm.Refresh()
	check.Assert(err, IsNil)
	foundItem := false
	if nil != vm.VM.VirtualHardwareSection.Item {
		for _, item := range vm.VM.VirtualHardwareSection.Item {
			if item.ResourceType == types.ResourceTypeProcessor {
				check.Assert(item.CoresPerSocket, Equals, cores)
				check.Assert(item.VirtualQuantity, Equals, cpuCount)
				foundItem = true
				break
			}
		}
		check.Assert(foundItem, Equals, true)
	}

	// return to previous value
	task, err = vm.ChangeCPUCountWithCore(int(currentCpus), &currentCores)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_VMToggleHardwareVirtualization(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	// Default nesting status should be false
	nestingStatus := existingVm.NestedHypervisorEnabled
	check.Assert(nestingStatus, Equals, false)

	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	// PowerOn
	task, err := vm.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Try to change the setting on powered on VM to fail
	_, err = vm.ToggleHardwareVirtualization(true)
	check.Assert(err, ErrorMatches, ".*hardware virtualization can be changed from powered off state.*")

	// PowerOf
	task, err = vm.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Perform steps on powered off VM
	task, err = vm.ToggleHardwareVirtualization(true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(vm.VM.NestedHypervisorEnabled, Equals, true)

	task, err = vm.ToggleHardwareVirtualization(false)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(vm.VM.NestedHypervisorEnabled, Equals, false)
}

func (vcd *TestVCD) Test_VMPowerOnPowerOff(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	// Ensure VM is not powered on
	vmStatus, err := vm.GetStatus()
	check.Assert(err, IsNil)
	if vmStatus != "POWERED_OFF" {
		task, err := vm.PowerOff()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		check.Assert(task.Task.Status, Equals, "success")
	}

	task, err := vm.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	err = vm.Refresh()
	check.Assert(err, IsNil)
	vmStatus, err = vm.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vmStatus, Equals, "POWERED_ON")

	// Power off again
	task, err = vm.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	err = vm.Refresh()
	check.Assert(err, IsNil)
	vmStatus, err = vm.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vmStatus, Equals, "POWERED_OFF")
}

func (vcd *TestVCD) Test_GetNetworkConnectionSection(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	networkBefore, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)

	err = vm.UpdateNetworkConnectionSection(networkBefore)
	check.Assert(err, IsNil)

	networkAfter, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)

	// Filter out always differing fields and do deep comparison of objects
	networkBefore.Link = &types.Link{}
	networkAfter.Link = &types.Link{}
	check.Assert(networkAfter, DeepEquals, networkBefore)

}

// Test_PowerOnAndForceCustomization uses the VM from TestSuite and forces guest customization
// in addition to the one which is triggered on first boot. It waits until the initial guest
// customization after first power on is finished because it is inherited from the template.
// After this initial wait it Undeploys VM and triggers a second customization and again waits until guest
// customization status exits "GC_PENDING" state to succeed the test.
// This test relies on longer timeouts in BlockWhileGuestCustomizationStatus because VMs take a lengthy time
// to boot up and report customization done.
func (vcd *TestVCD) Test_PowerOnAndForceCustomization(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}

	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm, err := vcd.client.Client.GetVMByHref(vmType.HREF)
	check.Assert(err, IsNil)

	// It may be that prebuilt VM was not booted before in the test vApp and it would still have
	// a guest customization status 'GC_PENDING'. This is because initially VM has this flag set
	// but while this flag is here the test cannot actually check if vm.PowerOnAndForceCustomization()
	// gives any effect therefore we must "wait through" initial guest customization if it is in
	// 'GC_PENDING' state.
	custStatus, err := vm.GetGuestCustomizationStatus()
	check.Assert(err, IsNil)
	if custStatus == types.GuestCustStatusPending {
		vmStatus, err := vm.GetStatus()
		check.Assert(err, IsNil)
		// If VM is POWERED OFF - let's power it on before waiting for its status to change
		if vmStatus == "POWERED_OFF" {
			task, err := vm.PowerOn()
			check.Assert(err, IsNil)
			err = task.WaitTaskCompletion()
			check.Assert(err, IsNil)
			check.Assert(task.Task.Status, Equals, "success")
		}

		err = vm.BlockWhileGuestCustomizationStatus(types.GuestCustStatusPending, 300)
		check.Assert(err, IsNil)
	}

	// Check that VM is deployed
	vmIsDeployed, err := vm.IsDeployed()
	check.Assert(err, IsNil)
	check.Assert(vmIsDeployed, Equals, true)

	// Try to force operation on deployed VM and expect an error
	err = vm.PowerOnAndForceCustomization()
	check.Assert(err, NotNil)

	// VM _must_ be un-deployed because PowerOnAndForceCustomization task will never finish (and
	// probably not triggered) if it is not un-deployed.
	task, err := vm.Undeploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Check that VM is un-deployed
	vmIsDeployed, err = vm.IsDeployed()
	check.Assert(err, IsNil)
	check.Assert(vmIsDeployed, Equals, false)

	err = vm.PowerOnAndForceCustomization()
	check.Assert(err, IsNil)

	// Ensure that VM has the status set to "GC_PENDING" after forced re-customization
	recustomizedVmStatus, err := vm.GetGuestCustomizationStatus()
	check.Assert(err, IsNil)
	check.Assert(recustomizedVmStatus, Equals, types.GuestCustStatusPending)

	// Check that VM is deployed
	vmIsDeployed, err = vm.IsDeployed()
	check.Assert(err, IsNil)
	check.Assert(vmIsDeployed, Equals, true)

	// Wait until the VM exists GC_PENDING status again. At the moment this is the only simple way
	// to see that the customization really worked as there is no API in vCD to execute remote
	// commands on guest VMs
	err = vm.BlockWhileGuestCustomizationStatus(types.GuestCustStatusPending, 300)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_BlockWhileGuestCustomizationStatus(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}

	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	// Attempt to set invalid timeout values and expect validation error
	err = vm.BlockWhileGuestCustomizationStatus(types.GuestCustStatusPending, 0)
	check.Assert(err, ErrorMatches, "timeOutAfterSeconds must be in range 4<X<7200")
	err = vm.BlockWhileGuestCustomizationStatus(types.GuestCustStatusPending, 4)
	check.Assert(err, ErrorMatches, "timeOutAfterSeconds must be in range 4<X<7200")
	err = vm.BlockWhileGuestCustomizationStatus(types.GuestCustStatusPending, -30)
	check.Assert(err, ErrorMatches, "timeOutAfterSeconds must be in range 4<X<7200")
	err = vm.BlockWhileGuestCustomizationStatus(types.GuestCustStatusPending, 7201)
	check.Assert(err, ErrorMatches, "timeOutAfterSeconds must be in range 4<X<7200")

	vmCustStatus, err := vm.GetGuestCustomizationStatus()
	check.Assert(err, IsNil)

	// Use current value to trigger timeout
	err = vm.BlockWhileGuestCustomizationStatus(vmCustStatus, 5)
	check.Assert(err, ErrorMatches, "timed out waiting for VM guest customization status to exit state GC_PENDING after 5 seconds")

	// Use unreal value to trigger instant unblocking
	err = vm.BlockWhileGuestCustomizationStatus("invalid_GC_STATUS", 5)
	check.Assert(err, IsNil)
}

// Test_VMSetProductSectionList sets product section, retrieves it and deeply matches if properties
// were properly set using a propertyTester helper.
func (vcd *TestVCD) Test_VMSetProductSectionList(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)
	propertyTester(vcd, check, vm)
}

// Test_VMSetGetGuestCustomizationSection sets and when retrieves guest customization and checks if properties are right.
func (vcd *TestVCD) Test_VMSetGetGuestCustomizationSection(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)
	guestCustomizationPropertyTester(vcd, check, vm)
}

// Test create internal disk For VM
func (vcd *TestVCD) Test_AddInternalDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// In general VM internal disks works with Org users, but due we need change VDC fast provisioning value, we have to be sys admins
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	vmName := "Test_AddInternalDisk"

	vm, storageProfile, diskSettings, diskId, previousProvisioningValue, err := vcd.createInternalDisk(check, vmName, 1)
	check.Assert(err, IsNil)

	//verify
	disk, err := vm.GetInternalDiskById(diskId, true)
	check.Assert(err, IsNil)

	check.Assert(disk.StorageProfile.HREF, Equals, storageProfile.HREF)
	check.Assert(disk.StorageProfile.ID, Equals, storageProfile.ID)
	check.Assert(disk.AdapterType, Equals, diskSettings.AdapterType)
	check.Assert(*disk.ThinProvisioned, Equals, *diskSettings.ThinProvisioned)
	check.Assert(*disk.Iops, Equals, *diskSettings.Iops)
	check.Assert(disk.SizeMb, Equals, diskSettings.SizeMb)
	check.Assert(disk.UnitNumber, Equals, diskSettings.UnitNumber)
	check.Assert(disk.BusNumber, Equals, diskSettings.BusNumber)
	check.Assert(disk.AdapterType, Equals, diskSettings.AdapterType)

	//cleanup
	err = vm.DeleteInternalDisk(diskId)
	check.Assert(err, IsNil)

	// disable fast provisioning if needed
	updateVdcFastProvisioning(vcd, check, previousProvisioningValue)

	// delete Vapp early to avoid env capacity issue
	deleteVapp(vcd, vmName)
}

// createInternalDisk Finds available VM and creates internal Disk in it.
// returns VM, storage profile, disk settings, disk id and error.
func (vcd *TestVCD) createInternalDisk(check *C, vmName string, busNumber int) (*VM, types.Reference, *types.DiskSettings, string, string, error) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}
	if vcd.config.VCD.StorageProfile.SP1 == "" {
		check.Skip("No Storage Profile given for VDC tests")
	}

	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("No Catalog name given for VDC tests")
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("No Catalog item given for VDC tests")
	}

	// disables fast provisioning if needed
	previousVdcFastProvisioningValue := updateVdcFastProvisioning(vcd, check, "disable")
	AddToCleanupList(previousVdcFastProvisioningValue, "fastProvisioning", vcd.config.VCD.Org+"|"+vcd.config.VCD.Vdc, "createInternalDisk")

	vdc, _, vappTemplate, vapp, desiredNetConfig, err := vcd.createAndGetResourcesForVmCreation(check, vmName)
	check.Assert(err, IsNil)

	vm, err := spawnVM("FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", true)
	check.Assert(err, IsNil)

	storageProfile, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)
	isThinProvisioned := true
	iops := int64(0)
	diskSettings := &types.DiskSettings{
		SizeMb:            1024,
		UnitNumber:        0,
		BusNumber:         busNumber,
		AdapterType:       "4",
		ThinProvisioned:   &isThinProvisioned,
		StorageProfile:    &storageProfile,
		OverrideVmDefault: true,
		Iops:              &iops,
	}

	diskId, err := vm.AddInternalDisk(diskSettings)
	check.Assert(err, IsNil)
	check.Assert(diskId, NotNil)
	return &vm, storageProfile, diskSettings, diskId, previousVdcFastProvisioningValue, err
}

// updateVdcFastProvisioning Enables or Disables fast provisioning if needed
func updateVdcFastProvisioning(vcd *TestVCD, check *C, enable string) string {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.config.VCD.Vdc, true)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)

	vdcFastProvisioningValue := "disabled"
	if *adminVdc.AdminVdc.UsesFastProvisioning {
		vdcFastProvisioningValue = "enable"
	}

	if *adminVdc.AdminVdc.UsesFastProvisioning && enable == "enable" {
		return vdcFastProvisioningValue
	}

	if !*adminVdc.AdminVdc.UsesFastProvisioning && enable != "enable" {
		return vdcFastProvisioningValue
	}
	valuePt := false
	if enable == "enable" {
		valuePt = true
	}
	adminVdc.AdminVdc.UsesFastProvisioning = &valuePt
	_, err = adminVdc.Update()
	check.Assert(err, IsNil)
	return vdcFastProvisioningValue
}

// Test delete internal disk For VM
func (vcd *TestVCD) Test_DeleteInternalDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// In general VM internal disks works with Org users, but due we need change VDC fast provisioning value, we have to be sys admins
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	vmName := "Test_DeleteInternalDisk"

	vm, _, _, diskId, previousProvisioningValue, err := vcd.createInternalDisk(check, vmName, 2)
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)

	err = vm.DeleteInternalDisk(diskId)
	check.Assert(err, IsNil)

	disk, err := vm.GetInternalDiskById(diskId, true)
	check.Assert(err, Equals, ErrorEntityNotFound)
	check.Assert(disk, IsNil)

	// enable fast provisioning if needed
	updateVdcFastProvisioning(vcd, check, previousProvisioningValue)

	// delete Vapp early to avoid env capacity issue
	deleteVapp(vcd, vmName)
}

// Test update internal disk for VM which has independent disk
func (vcd *TestVCD) Test_UpdateInternalDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// In general VM internal disks works with Org users, but due we need change VDC fast provisioning value, we have to be sys admins
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	vmName := "Test_UpdateInternalDisk"
	vm, storageProfile, diskSettings, diskId, previousProvisioningValue, err := vcd.createInternalDisk(check, vmName, 1)
	check.Assert(err, IsNil)

	//verify
	disk, err := vm.GetInternalDiskById(diskId, true)
	check.Assert(err, IsNil)
	check.Assert(disk, NotNil)

	// increase new disk size
	vmSpecSection := vm.VM.VmSpecSection
	changeDiskSettings := vm.VM.VmSpecSection.DiskSection.DiskSettings
	for _, diskSettings := range changeDiskSettings {
		if diskSettings.DiskId == diskId {
			diskSettings.SizeMb = 2048
		}
	}

	vmSpecSection.DiskSection.DiskSettings = changeDiskSettings

	vmSpecSection, err = vm.UpdateInternalDisks(vmSpecSection)
	check.Assert(err, IsNil)
	check.Assert(vmSpecSection, NotNil)

	disk, err = vm.GetInternalDiskById(diskId, true)
	check.Assert(err, IsNil)
	check.Assert(disk, NotNil)

	//verify
	check.Assert(disk.StorageProfile.HREF, Equals, storageProfile.HREF)
	check.Assert(disk.StorageProfile.ID, Equals, storageProfile.ID)
	check.Assert(disk.AdapterType, Equals, diskSettings.AdapterType)
	check.Assert(*disk.ThinProvisioned, Equals, *diskSettings.ThinProvisioned)
	check.Assert(*disk.Iops, Equals, *diskSettings.Iops)
	check.Assert(disk.SizeMb, Equals, int64(2048))
	check.Assert(disk.UnitNumber, Equals, diskSettings.UnitNumber)
	check.Assert(disk.BusNumber, Equals, diskSettings.BusNumber)
	check.Assert(disk.AdapterType, Equals, diskSettings.AdapterType)

	// attach independent disk
	independentDisk, err := attachIndependentDisk(vcd, check)
	check.Assert(err, IsNil)

	//cleanup
	err = vm.DeleteInternalDisk(diskId)
	check.Assert(err, IsNil)
	detachIndependentDisk(vcd, check, independentDisk)

	// disable fast provisioning if needed
	updateVdcFastProvisioning(vcd, check, previousProvisioningValue)

	// delete Vapp early to avoid env capacity issue
	deleteVapp(vcd, vmName)
}

func attachIndependentDisk(vcd *TestVCD, check *C) (*Disk, error) {
	// Find VM
	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

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
		SizeMb:      1,
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
	return disk, err
}

func detachIndependentDisk(vcd *TestVCD, check *C, disk *Disk) {
	err := vcd.detachIndependentDisk(Disk{disk.Disk, &vcd.client.Client})
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_VmGetParentvAppAndVdc(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}

	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.findFirstVapp()
	if vapp.VApp.Name == "" {
		check.Skip("Disabled: No suitable vApp found in vDC")
	}
	vm, vmName := vcd.findFirstVm(vapp)
	if vm.Name == "" {
		check.Skip("Disabled: No suitable VM found in vDC")
	}

	newVM, err := vcd.client.Client.GetVMByHref(vm.HREF)
	check.Assert(err, IsNil)
	check.Assert(newVM.VM.Name, Equals, vmName)
	check.Assert(newVM.VM.VirtualHardwareSection.Item, NotNil)

	parentvApp, err := newVM.GetParentVApp()
	check.Assert(err, IsNil)
	check.Assert(parentvApp.VApp.HREF, Equals, vapp.VApp.HREF)

	parentVdc, err := newVM.GetParentVdc()
	check.Assert(err, IsNil)
	check.Assert(parentVdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
}

func (vcd *TestVCD) Test_AddNewEmptyVMMultiNIC(check *C) {

	config := vcd.config
	if config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	// Find VApp
	if vcd.vapp != nil && vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp, err := createVappForTest(vcd, "Test_AddNewEmptyVMMultiNIC")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	desiredNetConfig := &types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 0
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModePool,
			Network:                 config.VCD.Network.Net1,
			NetworkConnectionIndex:  0,
		},
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModeNone,
			Network:                 types.NoneNetwork,
			NetworkConnectionIndex:  1,
		})

	// Test with two different networks if we have them
	if config.VCD.Network.Net2 != "" {
		// Attach second vdc network to vApp
		vdcNetwork2, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net2, false)
		check.Assert(err, IsNil)
		_, err = vapp.AddOrgNetwork(&VappNetworkSettings{}, vdcNetwork2.OrgVDCNetwork, false)
		check.Assert(err, IsNil)

		desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
			&types.NetworkConnection{
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModePool,
				Network:                 config.VCD.Network.Net2,
				NetworkConnectionIndex:  2,
			},
		)
	} else {
		fmt.Println("Skipping adding another vdc network as network2 was not specified")
	}

	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	media, err := cat.GetMediaByName(vcd.config.Media.Media, false)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)

	var task Task
	var sp types.Reference
	var customSP = false

	if vcd.config.VCD.StorageProfile.SP1 != "" {
		sp, _ = vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	}

	newDisk := types.DiskSettings{
		AdapterType:       "5",
		SizeMb:            int64(16384),
		BusNumber:         0,
		UnitNumber:        0,
		ThinProvisioned:   takeBoolPointer(true),
		OverrideVmDefault: true}

	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      "Test_AddNewEmptyVMMultiNIC",
			NetworkConnectionSection:  desiredNetConfig,
			Description:               "created by Test_AddNewEmptyVMMultiNIC",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          takeBoolPointer(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           takeIntAddress(2),
				NumCoresPerSocket: takeIntAddress(1),
				CpuResourceMhz:    &types.CpuResourceMhz{Configured: 1},
				MemoryResourceMb:  &types.MemoryResourceMb{Configured: 1024},
				DiskSection:       &types.DiskSection{DiskSettings: []*types.DiskSettings{&newDisk}},
				HardwareVersion:   &types.HardwareVersion{Value: "vmx-13"}, // need support older version vCD
				VmToolsVersion:    "",
				VirtualCpuType:    "VM32",
				TimeSyncWithHost:  nil,
			},
			BootImage: &types.Media{HREF: media.Media.HREF, Name: media.Media.Name, ID: media.Media.ID},
		},
		AllEULAsAccepted: true,
	}

	createdVm, err := vapp.AddEmptyVm(requestDetails)
	check.Assert(err, IsNil)
	check.Assert(createdVm, NotNil)

	// Ensure network config was valid
	actualNetConfig, err := createdVm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)

	if customSP {
		check.Assert(createdVm.VM.StorageProfile.HREF, Equals, sp.HREF)
	}

	verifyNetworkConnectionSection(check, actualNetConfig, desiredNetConfig)

	// Cleanup
	err = vapp.RemoveVM(*createdVm)
	check.Assert(err, IsNil)

	// Ensure network is detached from vApp to avoid conflicts in other tests
	task, err = vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// Test update of VM Spec section
func (vcd *TestVCD) Test_UpdateVmSpecSection(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	vmName := "Test_UpdateVmSpecSection"
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}

	vdc, _, vappTemplate, vapp, desiredNetConfig, err := vcd.createAndGetResourcesForVmCreation(check, vmName)
	check.Assert(err, IsNil)

	vm, err := spawnVM("FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", true)
	check.Assert(err, IsNil)

	task, err := vm.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	vmSpecSection := vm.VM.VmSpecSection
	osType := "sles10_64Guest"
	vmSpecSection.OsType = osType
	vmSpecSection.NumCpus = takeIntAddress(4)
	vmSpecSection.NumCoresPerSocket = takeIntAddress(2)
	vmSpecSection.MemoryResourceMb = &types.MemoryResourceMb{Configured: 768}

	updatedVm, err := vm.UpdateVmSpecSection(vmSpecSection, "updateDescription")
	check.Assert(err, IsNil)
	check.Assert(updatedVm, NotNil)

	//verify
	check.Assert(updatedVm.VM.VmSpecSection.OsType, Equals, osType)
	check.Assert(*updatedVm.VM.VmSpecSection.NumCpus, Equals, 4)
	check.Assert(*updatedVm.VM.VmSpecSection.NumCoresPerSocket, Equals, 2)
	check.Assert(updatedVm.VM.VmSpecSection.MemoryResourceMb.Configured, Equals, int64(768))
	check.Assert(updatedVm.VM.Description, Equals, "updateDescription")

	// delete Vapp early to avoid env capacity issue
	deleteVapp(vcd, vmName)
}

func (vcd *TestVCD) Test_QueryVmList(check *C) {

	if vcd.skipVappTests {
		check.Skip("Test_QueryVmList needs an existing vApp to run")
		return
	}

	// Get the setUp vApp using traditional methods
	vapp, err := vcd.vdc.GetVAppByName(TestSetUpSuite, true)
	check.Assert(err, IsNil)
	vmName := ""
	for _, vm := range vapp.VApp.Children.VM {
		vmName = vm.Name
		break
	}
	if vmName == "" {
		check.Skip("No VM names found")
		return
	}

	for filter := range []types.VmQueryFilter{types.VmQueryFilterOnlyDeployed, types.VmQueryFilterAll} {
		list, err := vcd.vdc.QueryVmList(types.VmQueryFilter(filter))
		check.Assert(err, IsNil)
		check.Assert(list, NotNil)
		foundVm := false

		// Check the VM list for a known VM name
		for _, vm := range list {
			if vm.Name == vmName {
				foundVm = true
				break
			}
		}
		check.Assert(foundVm, Equals, true)
	}
}

// Test update of VM Capabilities
func (vcd *TestVCD) Test_UpdateVmCpuAndMemoryHotAdd(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	vmName := "Test_UpdateVmCpuAndMemoryHotAdd"
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}

	vdc, _, vappTemplate, vapp, desiredNetConfig, err := vcd.createAndGetResourcesForVmCreation(check, vmName)
	check.Assert(err, IsNil)

	vm, err := spawnVM("FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", true)
	check.Assert(err, IsNil)

	task, err := vm.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	check.Assert(vm.VM.VMCapabilities.MemoryHotAddEnabled, Equals, false)
	check.Assert(vm.VM.VMCapabilities.CPUHotAddEnabled, Equals, false)

	updatedVm, err := vm.UpdateVmCpuAndMemoryHotAdd(true, true)
	check.Assert(err, IsNil)
	check.Assert(updatedVm, NotNil)

	//verify
	check.Assert(updatedVm.VM.VMCapabilities.MemoryHotAddEnabled, Equals, true)
	check.Assert(updatedVm.VM.VMCapabilities.CPUHotAddEnabled, Equals, true)

	// delete Vapp early to avoid env capacity issue
	deleteVapp(vcd, vmName)
}

func (vcd *TestVCD) Test_AddNewEmptyVMWithVmComputePolicyAndUpdate(check *C) {
	vapp, err := createVappForTest(vcd, "Test_AddNewEmptyVMWithVmComputePolicy")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	newComputePolicy := &VdcComputePolicy{
		client: vcd.org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:        check.TestName() + "_empty",
			Description: "Empty policy created by test",
		},
	}

	newComputePolicy2 := &VdcComputePolicy{
		client: vcd.org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:        check.TestName() + "_memory",
			Description: "Empty policy created by test 2",
			Memory:      takeIntAddress(2048),
		},
	}

	// Create and assign compute policy
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.vdc.Vdc.Name, false)
	if adminVdc == nil || err != nil {
		vcd.infoCleanup(notFoundMsg, "vdc", vcd.vdc.Vdc.Name)
	}

	createdPolicy, err := adminOrg.CreateVdcComputePolicy(newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)

	createdPolicy2, err := adminOrg.CreateVdcComputePolicy(newComputePolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_AddNewEmptyVMWithVmComputePolicyAndUpdate")
	AddToCleanupList(createdPolicy2.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_AddNewEmptyVMWithVmComputePolicyAndUpdate")

	vdcComputePolicyHref, err := adminOrg.client.OpenApiBuildEndpoint(types.OpenApiPathVersion1_0_0, types.OpenApiEndpointVdcComputePolicies)
	check.Assert(err, IsNil)

	// Get policy to existing ones (can be only default one)
	allAssignedComputePolicies, err := adminVdc.GetAllAssignedVdcComputePolicies(nil)
	check.Assert(err, IsNil)
	var policyReferences []*types.Reference
	for _, assignedPolicy := range allAssignedComputePolicies {
		policyReferences = append(policyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + assignedPolicy.VdcComputePolicy.ID})
	}
	policyReferences = append(policyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + createdPolicy.VdcComputePolicy.ID})
	policyReferences = append(policyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + createdPolicy2.VdcComputePolicy.ID})

	assignedVdcComputePolicies, err := adminVdc.SetAssignedComputePolicies(types.VdcComputePolicyReferences{VdcComputePolicyReference: policyReferences})
	check.Assert(err, IsNil)
	check.Assert(len(allAssignedComputePolicies)+2, Equals, len(assignedVdcComputePolicies.VdcComputePolicyReference))
	// end

	var task Task
	var sp types.Reference
	var customSP = false

	if vcd.config.VCD.StorageProfile.SP1 != "" {
		sp, _ = vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	}

	newDisk := types.DiskSettings{
		AdapterType:       "5",
		SizeMb:            int64(16384),
		BusNumber:         0,
		UnitNumber:        0,
		ThinProvisioned:   takeBoolPointer(true),
		OverrideVmDefault: true}

	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      "Test_AddNewEmptyVMWithVmComputePolicy",
			Description:               "created by Test_AddNewEmptyVMWithVmComputePolicy",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          takeBoolPointer(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           takeIntAddress(2),
				NumCoresPerSocket: takeIntAddress(1),
				CpuResourceMhz:    &types.CpuResourceMhz{Configured: 1},
				MemoryResourceMb:  &types.MemoryResourceMb{Configured: 1024},
				MediaSection:      nil,
				DiskSection:       &types.DiskSection{DiskSettings: []*types.DiskSettings{&newDisk}},
				HardwareVersion:   &types.HardwareVersion{Value: "vmx-13"}, // need support older version vCD
				VmToolsVersion:    "",
				VirtualCpuType:    "VM32",
				TimeSyncWithHost:  nil,
			},
			ComputePolicy: &types.ComputePolicy{VmSizingPolicy: &types.Reference{HREF: vdcComputePolicyHref.String() + createdPolicy.VdcComputePolicy.ID}},
		},
		AllEULAsAccepted: true,
	}

	createdVm, err := vapp.AddEmptyVm(requestDetails)
	check.Assert(err, IsNil)
	check.Assert(createdVm, NotNil)
	check.Assert(createdVm.VM.ComputePolicy, NotNil)
	check.Assert(createdVm.VM.ComputePolicy.VmSizingPolicy, NotNil)
	check.Assert(createdVm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, createdPolicy.VdcComputePolicy.ID)

	if customSP {
		check.Assert(createdVm.VM.StorageProfile.HREF, Equals, sp.HREF)
	}

	updatedVm, err := createdVm.UpdateComputePolicy(createdPolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)
	check.Assert(updatedVm, NotNil)
	check.Assert(updatedVm.VM.ComputePolicy, NotNil)
	check.Assert(updatedVm.VM.ComputePolicy.VmSizingPolicy, NotNil)
	check.Assert(updatedVm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, createdPolicy2.VdcComputePolicy.ID)
	check.Assert(updatedVm.VM.VmSpecSection.MemoryResourceMb, NotNil)
	check.Assert(updatedVm.VM.VmSpecSection.MemoryResourceMb.Configured, Equals, int64(2048))

	// Cleanup
	err = vapp.RemoveVM(*createdVm)
	check.Assert(err, IsNil)

	// Ensure network is detached from vApp to avoid conflicts in other tests
	task, err = vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// cleanup assigned compute policy
	var beforeTestPolicyReferences []*types.Reference
	for _, assignedPolicy := range allAssignedComputePolicies {
		beforeTestPolicyReferences = append(beforeTestPolicyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + assignedPolicy.VdcComputePolicy.ID})
	}

	_, err = adminVdc.SetAssignedComputePolicies(types.VdcComputePolicyReferences{VdcComputePolicyReference: beforeTestPolicyReferences})
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_VMUpdateStorageProfile(check *C) {
	config := vcd.config
	if config.VCD.StorageProfile.SP1 == "" || config.VCD.StorageProfile.SP2 == "" {
		check.Skip("Skipping test because both storage profiles have to be configured")
	}

	vapp, err := createVappForTest(vcd, "Test_VMUpdateStorageProfile")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	var storageProfile types.Reference

	storageProfile, _ = vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)

	createdVm, err := makeEmptyVm(vapp, "Test_VMUpdateStorageProfile")
	check.Assert(err, IsNil)
	check.Assert(createdVm, NotNil)
	check.Assert(createdVm.VM.StorageProfile.HREF, Equals, storageProfile.HREF)

	storageProfile2, _ := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP2)
	updatedVm, err := createdVm.UpdateStorageProfile(storageProfile2.HREF)
	check.Assert(err, IsNil)
	check.Assert(updatedVm, NotNil)
	check.Assert(createdVm.VM.StorageProfile.HREF, Equals, storageProfile2.HREF)

	// Cleanup
	var task Task
	err = vapp.RemoveVM(*createdVm)
	check.Assert(err, IsNil)

	// Ensure network is detached from vApp to avoid conflicts in other tests
	task, err = vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) getNetworkConnection() *types.NetworkConnectionSection {

	if vcd.config.VCD.Network.Net1 == "" {
		return nil
	}
	return &types.NetworkConnectionSection{
		Info:                          "Network Configuration for VM",
		PrimaryNetworkConnectionIndex: 0,
		NetworkConnection: []*types.NetworkConnection{
			&types.NetworkConnection{
				Network:                 vcd.config.VCD.Network.Net1,
				NeedsCustomization:      false,
				NetworkConnectionIndex:  0,
				IPAddress:               "any",
				IsConnected:             true,
				IPAddressAllocationMode: "DHCP",
				NetworkAdapterType:      "VMXNET3",
			},
		},
		Link: nil,
	}
}

func (vcd *TestVCD) Test_CreateStandaloneVM(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	vdc, err := adminOrg.GetVDCByName(vcd.vdc.Vdc.Name, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	description := "created by " + check.TestName()
	params := types.CreateVmParams{
		Name:        "testStandaloneVm",
		PowerOn:     false,
		Description: description,
		CreateVm: &types.Vm{
			Name:                     "testStandaloneVm",
			VirtualHardwareSection:   nil,
			NetworkConnectionSection: vcd.getNetworkConnection(),
			VmSpecSection: &types.VmSpecSection{
				Modified:          takeBoolPointer(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           takeIntAddress(1),
				NumCoresPerSocket: takeIntAddress(1),
				CpuResourceMhz: &types.CpuResourceMhz{
					Configured: 0,
				},
				MemoryResourceMb: &types.MemoryResourceMb{
					Configured: 512,
				},
				DiskSection: &types.DiskSection{
					DiskSettings: []*types.DiskSettings{
						&types.DiskSettings{
							SizeMb:            1024,
							UnitNumber:        0,
							BusNumber:         0,
							AdapterType:       "5",
							ThinProvisioned:   takeBoolPointer(true),
							OverrideVmDefault: false,
						},
					},
				},

				HardwareVersion: &types.HardwareVersion{Value: "vmx-14"},
				VmToolsVersion:  "",
				VirtualCpuType:  "VM32",
			},
			GuestCustomizationSection: &types.GuestCustomizationSection{
				Info:         "Specifies Guest OS Customization Settings",
				ComputerName: "standalone1",
			},
		},
		Xmlns: types.XMLNamespaceVCloud,
	}
	vappList := vdc.GetVappList()
	vappNum := len(vappList)
	vm, err := vdc.CreateStandaloneVm(&params)
	check.Assert(err, IsNil)
	check.Assert(vm, NotNil)
	AddToCleanupList(vm.VM.ID, "standaloneVm", "", check.TestName())
	check.Assert(vm.VM.Description, Equals, description)

	_ = vdc.Refresh()
	vappList = vdc.GetVappList()
	check.Assert(len(vappList), Equals, vappNum+1)
	for _, vapp := range vappList {
		printVerbose("vapp: %s\n", vapp.Name)
	}
	err = vm.Delete()
	check.Assert(err, IsNil)
	_ = vdc.Refresh()
	vappList = vdc.GetVappList()
	check.Assert(len(vappList), Equals, vappNum)
}

func (vcd *TestVCD) Test_CreateStandaloneVMFromTemplate(check *C) {

	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("no catalog was defined")
	}
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("no catalog item was defined")
	}
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	vdc, err := adminOrg.GetVDCByName(vcd.vdc.Vdc.Name, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	catalog, err := adminOrg.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	catalogItem, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catalogItem, NotNil)

	vappTemplate, err := catalog.GetVappTemplateByHref(catalogItem.CatalogItem.Entity.HREF)
	check.Assert(err, IsNil)
	check.Assert(vappTemplate, NotNil)
	check.Assert(vappTemplate.VAppTemplate.Children, NotNil)
	check.Assert(len(vappTemplate.VAppTemplate.Children.VM), Not(Equals), 0)

	vmTemplate := vappTemplate.VAppTemplate.Children.VM[0]
	check.Assert(vmTemplate.HREF, Not(Equals), "")
	check.Assert(vmTemplate.ID, Not(Equals), "")
	check.Assert(vmTemplate.Type, Not(Equals), "")
	check.Assert(vmTemplate.Name, Not(Equals), "")

	params := types.InstantiateVmTemplateParams{
		Xmlns:            types.XMLNamespaceVCloud,
		Name:             "testStandaloneTemplate",
		PowerOn:          true,
		AllEULAsAccepted: true,
		SourcedVmTemplateItem: &types.SourcedVmTemplateParams{
			LocalityParams: nil,
			Source: &types.Reference{
				HREF: vmTemplate.HREF,
				ID:   vmTemplate.ID,
				Type: vmTemplate.Type,
				Name: vmTemplate.Name,
			},
			StorageProfile:                nil,
			VmCapabilities:                nil,
			VmGeneralParams:               nil,
			VmTemplateInstantiationParams: nil,
		},
	}
	vappList := vdc.GetVappList()
	vappNum := len(vappList)
	util.Logger.Printf("%# v", pretty.Formatter(params))
	vm, err := vdc.CreateStandaloneVMFromTemplate(&params)
	check.Assert(err, IsNil)
	check.Assert(vm, NotNil)
	AddToCleanupList(vm.VM.ID, "standaloneVm", "", check.TestName())

	_ = vdc.Refresh()
	vappList = vdc.GetVappList()
	check.Assert(len(vappList), Equals, vappNum+1)
	for _, vapp := range vappList {
		printVerbose("vapp: %s\n", vapp.Name)
	}

	err = vm.Delete()
	check.Assert(err, IsNil)
	_ = vdc.Refresh()
	vappList = vdc.GetVappList()
	check.Assert(len(vappList), Equals, vappNum)
}

func (vcd *TestVCD) Test_VmMoveToVapp(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Org not found in configuration")
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("VDC not found in configuration")
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Catalog not found in configuration")
	}
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Catalog item not found in configuration")
	}
	// Retrieve Org, VDC, and Catalog
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)

	// Upload a small catalog item
	startTime := time.Now()
	catalogItemName := check.TestName()
	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, catalogItemName, "upload from VM move test", 1024)
	check.Assert(err, IsNil)

	AddToCleanupList(catalogItemName, "catalogItem", vcd.org.Org.Name+"|"+catalog.Catalog.Name, check.TestName())
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	elapsed := time.Since(startTime)
	verbosePrintf("upload vApp template %s\n", elapsed)

	// Create vApps
	vappName1 := check.TestName() + "-1"
	vappDescription1 := "test compose raw vAppV2 - 1"
	vappName2 := check.TestName() + "-2"
	vappDescription2 := "test compose raw vAppV2 - 2"
	vappName3 := check.TestName() + "-3"
	vappDescription3 := "test compose raw vAppV2 - 3"
	vappName4 := check.TestName() + "-4"
	vappDescription4 := "test compose raw vAppV2 - 4"

	startTimeVapps := time.Now()
	vapp1, err := makeEmptyVapp(vdc, vappName1, vappDescription1)
	check.Assert(err, IsNil)
	AddToCleanupList(vappName1, "vapp", vdc.Vdc.Name, check.TestName())

	vapp2, err := makeEmptyVapp(vdc, vappName2, vappDescription2)
	check.Assert(err, IsNil)
	AddToCleanupList(vappName2, "vapp", vdc.Vdc.Name, check.TestName())

	vapp3, err := makeEmptyVapp(vdc, vappName3, vappDescription3)
	check.Assert(err, IsNil)
	AddToCleanupList(vappName3, "vapp", vdc.Vdc.Name, check.TestName())

	vapp4, err := makeEmptyVapp(vdc, vappName4, vappDescription4)
	check.Assert(err, IsNil)
	AddToCleanupList(vappName4, "vapp", vdc.Vdc.Name, check.TestName())

	elapsed = time.Since(startTimeVapps)
	verbosePrintf("vApps created %s\n", elapsed)

	// Create VM from internal catalog in vApp2
	startTimeVm2 := time.Now()
	vm2, err := makeEmptyVm(vapp2, vappName2)
	check.Assert(err, IsNil)

	elapsed = time.Since(startTimeVm2)
	verbosePrintf("empty VM created %s\n", elapsed)

	// Create VM from small template in vApp3
	startTimeVm3 := time.Now()
	item, err := catalog.GetCatalogItemByName(catalogItemName, true)
	check.Assert(err, IsNil)
	template, err := item.GetVAppTemplate()
	check.Assert(err, IsNil)
	task, err := vapp3.AddNewVM(vappName3, template, nil, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	vm3, err := vapp3.GetVMByName(vappName3, true)
	check.Assert(err, IsNil)

	elapsed = time.Since(startTimeVm3)
	verbosePrintf("VM from small template created %s\n", elapsed)

	templateName := vcd.config.VCD.Catalog.CatalogItem
	// Create VM from larger template in vApp4
	startTimeVm4 := time.Now()
	existingItem, err := catalog.GetCatalogItemByName(templateName, true)
	check.Assert(err, IsNil)
	existingTemplate, err := existingItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	task, err = vapp4.AddNewVM(vappName4, existingTemplate, nil, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	vm4, err := vapp4.GetVMByName(vappName4, true)
	check.Assert(err, IsNil)

	// Save VM IDs, to be checked back after moving
	vm2Id := vm2.VM.ID
	vm3Id := vm3.VM.ID
	vm4Id := vm4.VM.ID

	elapsed = time.Since(startTimeVm4)
	verbosePrintf("VM from larger template created %s\n", elapsed)

	// Check that each vApp has the initially expected number of VMs
	numOfVms := func(vapp *VApp) int {
		if vapp.VApp.Children != nil {
			return len(vapp.VApp.Children.VM)
		}
		return 0
	}

	check.Assert(numOfVms(vapp1), Equals, 0)
	check.Assert(numOfVms(vapp2), Equals, 1)
	check.Assert(numOfVms(vapp3), Equals, 1)
	check.Assert(numOfVms(vapp4), Equals, 1)

	// Move VM from vApp2 to vApp1
	startTimeMoveVm2 := time.Now()
	err = vm2.MoveToVapp(vapp1)
	check.Assert(err, IsNil)

	elapsed = time.Since(startTimeMoveVm2)
	verbosePrintf("moved empty VM %s\n", elapsed)

	// Move VM from vApp3 to vApp1
	startTimeMoveVm3 := time.Now()
	err = vm3.MoveToVapp(vapp1)
	check.Assert(err, IsNil)

	elapsed = time.Since(startTimeMoveVm3)
	verbosePrintf("Moved VM 3 (small template) %s\n", elapsed)

	// Move VM from vApp4 to vApp1
	startTimeMoveVm4 := time.Now()
	err = vm4.MoveToVapp(vapp1)
	check.Assert(err, IsNil)

	elapsed = time.Since(startTimeMoveVm4)
	verbosePrintf("Moved VM 4 (larger template) %s\n", elapsed)

	// Refresh the vApps after moving
	// vapp1 gets refreshed implicitly during the move.
	for _, vapp := range []*VApp{vapp2, vapp3, vapp4} {
		err = vapp.Refresh()
		check.Assert(err, IsNil)
	}

	// Make sure the destination vApp has all the VMs, while the source vApps are empty
	check.Assert(numOfVms(vapp1), Equals, 3)
	check.Assert(numOfVms(vapp2), Equals, 0)
	check.Assert(numOfVms(vapp3), Equals, 0)
	check.Assert(numOfVms(vapp4), Equals, 0)

	// Retrieve the moved VMs from the destination vApp
	newVm2, err := vapp1.GetVMByName(vappName2, false)
	check.Assert(err, IsNil)
	newVm3, err := vapp1.GetVMByName(vappName3, false)
	check.Assert(err, IsNil)
	newVm4, err := vapp1.GetVMByName(vappName4, false)
	check.Assert(err, IsNil)

	// Make sure that the VMs keep their IDs after moving
	check.Assert(vm2Id, Equals, newVm2.VM.ID)
	check.Assert(vm3Id, Equals, newVm3.VM.ID)
	check.Assert(vm4Id, Equals, newVm4.VM.ID)

	// Remove all vApps
	var tasks = make([]Task, 4)
	for i, vapp := range []*VApp{vapp1, vapp2, vapp3, vapp4} {
		tasks[i], err = vapp.Delete()
		check.Assert(err, IsNil)
	}
	for _, task := range tasks {
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	// Remove catalog  item
	err = item.Delete()
	check.Assert(err, IsNil)
}

/*
func (vcd *TestVCD) Test_VmParallelVapp(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Org not found in configuration")
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("VDC not found in configuration")
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Catalog not found in configuration")
	}
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Catalog item not found in configuration")
	}
	// Retrieve Org, VDC, and Catalog
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)

	baseVappName := check.TestName()
	baseVappDescription := "test parallel vApp"
	baseVmName := check.TestName() + "-vm"

	// Create vApps
	const numVapps = 10
	var vApps [numVapps]*VApp
	startTimeVapps := time.Now()
	for i := 0; i < numVapps; i++ {
		vappName := fmt.Sprintf("%s - %d", baseVappName, i)
		vappDescription := fmt.Sprintf("%s - %d", baseVappDescription, i)
		vApps[i], err = makeEmptyVapp(vdc, vappName, vappDescription)
		check.Assert(err, IsNil)
		AddToCleanupList(vappName, "vapp", vdc.Vdc.Name, check.TestName())
	}

	elapsed := time.Since(startTimeVapps)
	verbosePrintf("vApps created %s\n", elapsed)

	templateName := vcd.config.VCD.Catalog.CatalogItem
	if os.Getenv("vcd_template_name") != "" {
		templateName = os.Getenv("vcd_template_name")
	}

	var tasks [numVapps]Task

	startTimeVms := time.Now()
	existingItem, err := catalog.GetCatalogItemByName(templateName, true)
	check.Assert(err, IsNil)
	existingTemplate, err := existingItem.GetVAppTemplate()
	check.Assert(err, IsNil)

	for i := 1; i < numVapps; i++ {
		vmName := fmt.Sprintf("%s - %d", baseVmName, i)
		tasks[i], err = vApps[i].AddNewVM(vmName, existingTemplate, nil, true)
		check.Assert(err, IsNil)
	}

	elapsed = time.Since(startTimeVms)
	verbosePrintf("VM tasks started %s\n", elapsed)

	startTimeTaskCompletion := time.Now()

	for i := 1; i < numVapps; i++ {
		err = tasks[i].WaitTaskCompletion()
		check.Assert(err, IsNil)
	}
	elapsed = time.Since(startTimeTaskCompletion)
	verbosePrintf("VM tasks completed %s\n", elapsed)

	// Move all VMs to vApp1
	startTimeMoveVms := time.Now()
	var vms = make([]*VM, numVapps -1)
	for i := 1; i < numVapps; i++ {
		vmName := fmt.Sprintf("%s - %d", baseVmName, i)
		vm, err := vApps[i].GetVMByName(vmName, true)
		check.Assert(err, IsNil)
		vms[i-1] = vm
		//err = vm.MoveToVapp(vApps[0])
		//check.Assert(err, IsNil)
	}

	err = MoveVmsToVapp(vms, vApps[0])
	check.Assert(err, IsNil)
	elapsed = time.Since(startTimeMoveVms)
	verbosePrintf("moved VMs %s\n", elapsed)

	numOfVms := func(vapp *VApp) int {
		if vapp.VApp.Children != nil {
			return len(vapp.VApp.Children.VM)
		}
		return 0
	}
	// Refresh the vApps after moving
	// Make sure the destination vApp has all the VMs
	for i := 0; i < numVapps; i++ {
		err = vApps[i].Refresh()
		check.Assert(err, IsNil)
	}
	check.Assert(numOfVms(vApps[0]), Equals, numVapps-1)

	// Remove all vApps
	for i := 0; i < numVapps; i++ {
		tasks[i], err = vApps[i].Delete()
		check.Assert(err, IsNil)
	}
	for _, task := range tasks {
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}
}
*/

func (vcd *TestVCD) Test_VmSerialVapp(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Org not found in configuration")
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("VDC not found in configuration")
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Catalog not found in configuration")
	}
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Catalog item not found in configuration")
	}
	// Retrieve Org, VDC, and Catalog
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)

	vappName := check.TestName()
	vappDescription := "test parallel vApp"
	baseVmName := check.TestName() + "-vm"

	// Create vApp
	const numVms = 9
	startTimeVapp := time.Now()
	vapp, err := makeEmptyVapp(vdc, vappName, vappDescription)
	check.Assert(err, IsNil)
	AddToCleanupList(vappName, "vapp", vdc.Vdc.Name, check.TestName())

	elapsed := time.Since(startTimeVapp)
	verbosePrintf("vApp created %s\n", elapsed)
	templateName := vcd.config.VCD.Catalog.CatalogItem
	if os.Getenv("vcd_template_name") != "" {
		templateName = os.Getenv("vcd_template_name")
	}

	startTimeVms := time.Now()
	existingItem, err := catalog.GetCatalogItemByName(templateName, true)
	check.Assert(err, IsNil)
	existingTemplate, err := existingItem.GetVAppTemplate()
	check.Assert(err, IsNil)

	for i := 0; i < numVms; i++ {
		vmName := fmt.Sprintf("%s - %d", baseVmName, i)
		task, err := vapp.AddNewVM(vmName, existingTemplate, nil, true)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	elapsed = time.Since(startTimeVms)
	verbosePrintf("VM tasks completed %s\n", elapsed)

	// Remove vApp
	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_ParallelVm(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("Org not found in configuration")
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("VDC not found in configuration")
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Catalog not found in configuration")
	}
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Catalog item not found in configuration")
	}
	// Retrieve Org, VDC, and Catalog
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)

	baseVmName := check.TestName() + "-vm"
	templateName := vcd.config.VCD.Catalog.CatalogItem
	if os.Getenv("vcd_template_name") != "" {
		templateName = os.Getenv("vcd_template_name")
	}
	existingItem, err := catalog.GetCatalogItemByName(templateName, true)
	check.Assert(err, IsNil)

	vappTemplate, err := catalog.GetVappTemplateByHref(existingItem.CatalogItem.Entity.HREF)
	check.Assert(err, IsNil)
	check.Assert(vappTemplate, NotNil)
	check.Assert(vappTemplate.VAppTemplate.Children, NotNil)
	check.Assert(len(vappTemplate.VAppTemplate.Children.VM), Not(Equals), 0)

	vmTemplate := vappTemplate.VAppTemplate.Children.VM[0]
	check.Assert(vmTemplate.HREF, Not(Equals), "")
	check.Assert(vmTemplate.ID, Not(Equals), "")
	check.Assert(vmTemplate.Type, Not(Equals), "")
	check.Assert(vmTemplate.Name, Not(Equals), "")

	// Create VMs
	const numVms = 9
	var VMs = make([]*VM, numVms)
	var tasks [numVms]Task
	startTimeVms := time.Now()
	for i := 0; i < numVms; i++ {
		vmName := fmt.Sprintf("%s - %d", baseVmName, i)
		vmDescription := fmt.Sprintf(" standalone VM %s - %d", baseVmName, i)

		params := types.InstantiateVmTemplateParams{
			Xmlns:            types.XMLNamespaceVCloud,
			Name:             vmName,
			PowerOn:          false,
			Description:      vmDescription,
			AllEULAsAccepted: true,
			SourcedVmTemplateItem: &types.SourcedVmTemplateParams{
				LocalityParams: nil,
				Source: &types.Reference{
					HREF: vmTemplate.HREF,
					ID:   vmTemplate.ID,
					Type: vmTemplate.Type,
					Name: vmTemplate.Name,
				},
				StorageProfile:                nil,
				VmCapabilities:                nil,
				VmGeneralParams:               nil,
				VmTemplateInstantiationParams: nil,
			},
		}
		tasks[i], err = vdc.CreateStandaloneVMFromTemplateAsync(&params)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		AddToCleanupList(vmName, "standaloneVm", "", check.TestName())
	}

	elapsed := time.Since(startTimeVms)
	verbosePrintf("VM tasks started %s\n", elapsed)

	startTimeVmTasks := time.Now()
	for i := 0; i < numVms; i++ {
		vmName := fmt.Sprintf("%s - %d", baseVmName, i)
		err = tasks[i].WaitTaskCompletion()
		check.Assert(err, IsNil)
		VMs[i], err = vdc.QueryVmByName(vmName)
		check.Assert(err, IsNil)
	}

	elapsed = time.Since(startTimeVmTasks)
	verbosePrintf("VMs created %s\n", elapsed)

	startTimeMoveVms := time.Now()
	// Move all VMs to vApp
	vapp, err := makeEmptyVapp(vdc, check.TestName(), "destination vApp")
	check.Assert(err, IsNil)

	err = MoveVmsToVapp(VMs, vapp)
	check.Assert(err, IsNil)
	elapsed = time.Since(startTimeMoveVms)
	verbosePrintf("moved VMs %s\n", elapsed)

	numOfVms := func(vapp *VApp) int {
		if vapp.VApp.Children != nil {
			return len(vapp.VApp.Children.VM)
		}
		return 0
	}
	// Refresh the vApp after moving
	err = vapp.Refresh()
	check.Assert(err, IsNil)
	check.Assert(numOfVms(vapp), Equals, numVms)

	// Remove vApp
	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}
