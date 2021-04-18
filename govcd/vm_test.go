// +build vm functional ALL

/*
* Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
* Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"
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
	ctx := context.Background()

	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.findFirstVapp(ctx)
	if vapp.VApp.Name == "" {
		check.Skip("Disabled: No suitable vApp found in vDC")
	}
	vm, vmName := vcd.findFirstVm(vapp)
	if vm.Name == "" {
		check.Skip("Disabled: No suitable VM found in vDC")
	}

	newVM, err := vcd.client.Client.GetVMByHref(ctx, vm.HREF)
	check.Assert(err, IsNil)
	check.Assert(newVM.VM.Name, Equals, vmName)
	check.Assert(newVM.VM.VirtualHardwareSection.Item, NotNil)
}

// Test attach disk to VM and detach disk from VM
func (vcd *TestVCD) Test_VMAttachOrDetachDisk(check *C) {
	if vcd.config.VCD.Disk.Size <= 0 {
		check.Skip("skipping test because disk size is 0")
	}

	// Find VM
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	ctx := context.Background()
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
		Name:        TestVMAttachOrDetachDisk,
		Size:        vcd.config.VCD.Disk.Size,
		Description: TestVMAttachOrDetachDisk,
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
	attachDiskTask, err := vm.attachOrDetachDisk(ctx, &types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	}, types.RelDiskAttach)
	check.Assert(err, IsNil)

	err = attachDiskTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Get attached VM
	vmRef, err := disk.AttachedVM(ctx)
	check.Assert(err, IsNil)
	check.Assert(vmRef, NotNil)
	check.Assert(vmRef.Name, Equals, vm.VM.Name)

	// Detach disk
	detachDiskTask, err := vm.attachOrDetachDisk(ctx, &types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	}, types.RelDiskDetach)
	check.Assert(err, IsNil)

	err = detachDiskTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

}

// Test attach disk to VM
func (vcd *TestVCD) Test_VMAttachDisk(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

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

	// Discard vApp suspension
	// Disk attach and detach operations are not working if vApp is suspended
	err := vcd.ensureVappIsSuitableForVMTest(ctx, vapp)
	check.Assert(err, IsNil)
	err = vcd.ensureVMIsSuitableForVMTest(ctx, vm)
	check.Assert(err, IsNil)

	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestVMAttachDisk,
		Size:        vcd.config.VCD.Disk.Size,
		Description: TestVMAttachDisk,
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

	// Cleanup: Detach disk
	detachDiskTask, err := vm.attachOrDetachDisk(ctx, &types.DiskAttachOrDetachParams{
		Disk: &types.Reference{
			HREF: disk.Disk.HREF,
		},
	}, types.RelDiskDetach)
	check.Assert(err, IsNil)

	err = detachDiskTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

}

// Test detach disk from VM
func (vcd *TestVCD) Test_VMDetachDisk(check *C) {

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
		Name:        TestVMDetachDisk,
		Size:        vcd.config.VCD.Disk.Size,
		Description: TestVMDetachDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(ctx, diskCreateParams)
	check.Assert(err, IsNil)

	check.Assert(task.Task.Owner.Type, Equals, types.MimeDisk)
	diskHREF := task.Task.Owner.HREF

	// Defer prepend the disk info to cleanup list until the function returns
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

}

// Test Insert or Eject Media for VM
func (vcd *TestVCD) Test_HandleInsertOrEjectMedia(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	itemName := "TestHandleInsertOrEjectMedia"

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	ctx := context.Background()
	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Upload Media
	catalog, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_HandleInsertOrEjectMedia")

	catalog, err = vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	media, err := catalog.GetMediaByName(ctx, itemName, false)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)

	insertMediaTask, err := vm.HandleInsertMedia(ctx, vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	err = insertMediaTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, true)

	vm, err = vm.HandleEjectMediaAndAnswer(ctx, vcd.org, vcd.config.VCD.Catalog.Name, itemName, true)
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
	ctx := context.Background()

	// Skipping this test due to a bug in vCD. VM refresh status returns old state, though eject task is finished.
	if vcd.client.Client.APIVCDMaxVersionIs(ctx, ">= 32.0, <= 33.0") {
		check.Skip("Skipping test because this vCD version has a bug")
	}

	itemName := "TestInsertOrEjectMedia"

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Upload Media
	catalog, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_InsertOrEjectMedia")

	catalog, err = vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	media, err := catalog.GetMediaByName(ctx, itemName, false)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)

	// Insert Media
	insertMediaTask, err := vm.insertOrEjectMedia(ctx, &types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.Media.HREF,
			Name: media.Media.Name,
			ID:   media.Media.ID,
			Type: media.Media.Type,
		},
	}, types.RelMediaInsertMedia)
	check.Assert(err, IsNil)

	err = insertMediaTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)

	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, true)

	// Insert Media
	ejectMediaTask, err := vm.insertOrEjectMedia(ctx, &types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.Media.HREF,
		},
	}, types.RelMediaEjectMedia)
	check.Assert(err, IsNil)

	err = ejectMediaTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh(ctx)
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
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	ctx := context.Background()
	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Upload Media
	catalog, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_AnswerVmQuestion")

	catalog, err = vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	media, err := catalog.GetMediaByName(ctx, itemName, false)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)

	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)

	insertMediaTask, err := vm.HandleInsertMedia(ctx, vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	err = insertMediaTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, true)

	ejectMediaTask, err := vm.HandleEjectMedia(ctx, vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	for i := 0; i < 10; i++ {
		question, err := vm.GetQuestion(ctx)
		check.Assert(err, IsNil)

		if question.QuestionId != "" && strings.Contains(question.Question, "Disconnect anyway and override the lock?") {
			err = vm.AnswerQuestion(ctx, question.QuestionId, 0)
			check.Assert(err, IsNil)
		}
		time.Sleep(time.Second * 3)
	}

	err = ejectMediaTask.Task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, false)
}

func (vcd *TestVCD) Test_VMChangeCPUCountWithCore(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	ctx := context.Background()
	vapp := vcd.findFirstVapp(ctx)
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

	vm, err := vcd.client.Client.GetVMByHref(ctx, existingVm.HREF)
	check.Assert(err, IsNil)

	cores := 2
	cpuCount := int64(4)

	task, err := vm.ChangeCPUCountWithCore(ctx, int(cpuCount), &cores)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vm.Refresh(ctx)
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
	task, err = vm.ChangeCPUCountWithCore(ctx, int(currentCpus), &currentCores)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_VMToggleHardwareVirtualization(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	ctx := context.Background()
	vapp := vcd.findFirstVapp(ctx)
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	// Default nesting status should be false
	nestingStatus := existingVm.NestedHypervisorEnabled
	check.Assert(nestingStatus, Equals, false)

	vm, err := vcd.client.Client.GetVMByHref(ctx, existingVm.HREF)
	check.Assert(err, IsNil)

	// PowerOn
	task, err := vm.PowerOn(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Try to change the setting on powered on VM to fail
	_, err = vm.ToggleHardwareVirtualization(ctx, true)
	check.Assert(err, ErrorMatches, ".*hardware virtualization can be changed from powered off state.*")

	// PowerOf
	task, err = vm.PowerOff(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Perform steps on powered off VM
	task, err = vm.ToggleHardwareVirtualization(ctx, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(vm.VM.NestedHypervisorEnabled, Equals, true)

	task, err = vm.ToggleHardwareVirtualization(ctx, false)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(vm.VM.NestedHypervisorEnabled, Equals, false)
}

func (vcd *TestVCD) Test_VMPowerOnPowerOff(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	ctx := context.Background()
	vapp := vcd.findFirstVapp(ctx)
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.client.Client.GetVMByHref(ctx, existingVm.HREF)
	check.Assert(err, IsNil)

	// Ensure VM is not powered on
	vmStatus, err := vm.GetStatus(ctx)
	check.Assert(err, IsNil)
	if vmStatus != "POWERED_OFF" {
		task, err := vm.PowerOff(ctx)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion(ctx)
		check.Assert(err, IsNil)
		check.Assert(task.Task.Status, Equals, "success")
	}

	task, err := vm.PowerOn(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)
	vmStatus, err = vm.GetStatus(ctx)
	check.Assert(err, IsNil)
	check.Assert(vmStatus, Equals, "POWERED_ON")

	// Power off again
	task, err = vm.PowerOff(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)
	vmStatus, err = vm.GetStatus(ctx)
	check.Assert(err, IsNil)
	check.Assert(vmStatus, Equals, "POWERED_OFF")
}

func (vcd *TestVCD) Test_GetNetworkConnectionSection(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	ctx := context.Background()
	vapp := vcd.findFirstVapp(ctx)
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm, err := vcd.client.Client.GetVMByHref(ctx, existingVm.HREF)
	check.Assert(err, IsNil)

	networkBefore, err := vm.GetNetworkConnectionSection(ctx)
	check.Assert(err, IsNil)

	err = vm.UpdateNetworkConnectionSection(ctx, networkBefore)
	check.Assert(err, IsNil)

	networkAfter, err := vm.GetNetworkConnectionSection(ctx)
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
	ctx := context.Background()
	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm, err := vcd.client.Client.GetVMByHref(ctx, vmType.HREF)
	check.Assert(err, IsNil)

	// It may be that prebuilt VM was not booted before in the test vApp and it would still have
	// a guest customization status 'GC_PENDING'. This is because initially VM has this flag set
	// but while this flag is here the test cannot actually check if vm.PowerOnAndForceCustomization()
	// gives any effect therefore we must "wait through" initial guest customization if it is in
	// 'GC_PENDING' state.
	custStatus, err := vm.GetGuestCustomizationStatus(ctx)
	check.Assert(err, IsNil)
	if custStatus == types.GuestCustStatusPending {
		vmStatus, err := vm.GetStatus(ctx)
		check.Assert(err, IsNil)
		// If VM is POWERED OFF - let's power it on before waiting for its status to change
		if vmStatus == "POWERED_OFF" {
			task, err := vm.PowerOn(ctx)
			check.Assert(err, IsNil)
			err = task.WaitTaskCompletion(ctx)
			check.Assert(err, IsNil)
			check.Assert(task.Task.Status, Equals, "success")
		}

		err = vm.BlockWhileGuestCustomizationStatus(ctx, types.GuestCustStatusPending, 300)
		check.Assert(err, IsNil)
	}

	// Check that VM is deployed
	vmIsDeployed, err := vm.IsDeployed(ctx)
	check.Assert(err, IsNil)
	check.Assert(vmIsDeployed, Equals, true)

	// Try to force operation on deployed VM and expect an error
	err = vm.PowerOnAndForceCustomization(ctx)
	check.Assert(err, NotNil)

	// VM _must_ be un-deployed because PowerOnAndForceCustomization task will never finish (and
	// probably not triggered) if it is not un-deployed.
	task, err := vm.Undeploy(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Check that VM is un-deployed
	vmIsDeployed, err = vm.IsDeployed(ctx)
	check.Assert(err, IsNil)
	check.Assert(vmIsDeployed, Equals, false)

	err = vm.PowerOnAndForceCustomization(ctx)
	check.Assert(err, IsNil)

	// Ensure that VM has the status set to "GC_PENDING" after forced re-customization
	recustomizedVmStatus, err := vm.GetGuestCustomizationStatus(ctx)
	check.Assert(err, IsNil)
	check.Assert(recustomizedVmStatus, Equals, types.GuestCustStatusPending)

	// Check that VM is deployed
	vmIsDeployed, err = vm.IsDeployed(ctx)
	check.Assert(err, IsNil)
	check.Assert(vmIsDeployed, Equals, true)

	// Wait until the VM exists GC_PENDING status again. At the moment this is the only simple way
	// to see that the customization really worked as there is no API in vCD to execute remote
	// commands on guest VMs
	err = vm.BlockWhileGuestCustomizationStatus(ctx, types.GuestCustStatusPending, 300)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_BlockWhileGuestCustomizationStatus(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}
	ctx := context.Background()
	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.findFirstVapp(ctx)
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm, err := vcd.client.Client.GetVMByHref(ctx, existingVm.HREF)
	check.Assert(err, IsNil)

	// Attempt to set invalid timeout values and expect validation error
	err = vm.BlockWhileGuestCustomizationStatus(ctx, types.GuestCustStatusPending, 0)
	check.Assert(err, ErrorMatches, "timeOutAfterSeconds must be in range 4<X<7200")
	err = vm.BlockWhileGuestCustomizationStatus(ctx, types.GuestCustStatusPending, 4)
	check.Assert(err, ErrorMatches, "timeOutAfterSeconds must be in range 4<X<7200")
	err = vm.BlockWhileGuestCustomizationStatus(ctx, types.GuestCustStatusPending, -30)
	check.Assert(err, ErrorMatches, "timeOutAfterSeconds must be in range 4<X<7200")
	err = vm.BlockWhileGuestCustomizationStatus(ctx, types.GuestCustStatusPending, 7201)
	check.Assert(err, ErrorMatches, "timeOutAfterSeconds must be in range 4<X<7200")

	vmCustStatus, err := vm.GetGuestCustomizationStatus(ctx)
	check.Assert(err, IsNil)

	// Use current value to trigger timeout
	err = vm.BlockWhileGuestCustomizationStatus(ctx, vmCustStatus, 5)
	check.Assert(err, ErrorMatches, "timed out waiting for VM guest customization status to exit state GC_PENDING after 5 seconds")

	// Use unreal value to trigger instant unblocking
	err = vm.BlockWhileGuestCustomizationStatus(ctx, "invalid_GC_STATUS", 5)
	check.Assert(err, IsNil)
}

// Test_VMSetProductSectionList sets product section, retrieves it and deeply matches if properties
// were properly set using a propertyTester helper.
func (vcd *TestVCD) Test_VMSetProductSectionList(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	ctx := context.Background()
	vapp := vcd.findFirstVapp(ctx)
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.client.Client.GetVMByHref(ctx, existingVm.HREF)
	check.Assert(err, IsNil)
	propertyTester(ctx, vcd, check, vm)
}

// Test_VMSetGetGuestCustomizationSection sets and when retrieves guest customization and checks if properties are right.
func (vcd *TestVCD) Test_VMSetGetGuestCustomizationSection(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	ctx := context.Background()
	vapp := vcd.findFirstVapp(ctx)
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.client.Client.GetVMByHref(ctx, existingVm.HREF)
	check.Assert(err, IsNil)
	guestCustomizationPropertyTester(ctx, vcd, check, vm)
}

// Test create internal disk For VM
func (vcd *TestVCD) Test_AddInternalDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// In general VM internal disks works with Org users, but due we need change VDC fast provisioning value, we have to be sys admins
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	ctx := context.Background()
	vmName := "Test_AddInternalDisk"

	vm, storageProfile, diskSettings, diskId, previousProvisioningValue, err := vcd.createInternalDisk(ctx, check, vmName, 1)
	check.Assert(err, IsNil)

	//verify
	disk, err := vm.GetInternalDiskById(ctx, diskId, true)
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
	err = vm.DeleteInternalDisk(ctx, diskId)
	check.Assert(err, IsNil)

	// disable fast provisioning if needed
	updateVdcFastProvisioning(ctx, vcd, check, previousProvisioningValue)

	// delete Vapp early to avoid env capacity issue
	deleteVapp(ctx, vcd, vmName)
}

// createInternalDisk Finds available VM and creates internal Disk in it.
// returns VM, storage profile, disk settings, disk id and error.
func (vcd *TestVCD) createInternalDisk(ctx context.Context, check *C, vmName string, busNumber int) (*VM, types.Reference, *types.DiskSettings, string, string, error) {
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
	previousVdcFastProvisioningValue := updateVdcFastProvisioning(ctx, vcd, check, "disable")
	AddToCleanupList(previousVdcFastProvisioningValue, "fastProvisioning", vcd.config.VCD.Org+"|"+vcd.config.VCD.Vdc, "createInternalDisk")

	vdc, _, vappTemplate, vapp, desiredNetConfig, err := vcd.createAndGetResourcesForVmCreation(ctx, check, vmName)
	check.Assert(err, IsNil)

	vm, err := spawnVM(ctx, "FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", true)
	check.Assert(err, IsNil)

	storageProfile, err := vcd.vdc.FindStorageProfileReference(ctx, vcd.config.VCD.StorageProfile.SP1)
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

	diskId, err := vm.AddInternalDisk(ctx, diskSettings)
	check.Assert(err, IsNil)
	check.Assert(diskId, NotNil)
	return &vm, storageProfile, diskSettings, diskId, previousVdcFastProvisioningValue, err
}

// updateVdcFastProvisioning Enables or Disables fast provisioning if needed
func updateVdcFastProvisioning(ctx context.Context, vcd *TestVCD, check *C, enable string) string {
	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(ctx, vcd.config.VCD.Vdc, true)
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
	_, err = adminVdc.Update(ctx)
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
	ctx := context.Background()
	vmName := "Test_DeleteInternalDisk"

	vm, _, _, diskId, previousProvisioningValue, err := vcd.createInternalDisk(ctx, check, vmName, 2)
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)

	err = vm.DeleteInternalDisk(ctx, diskId)
	check.Assert(err, IsNil)

	disk, err := vm.GetInternalDiskById(ctx, diskId, true)
	check.Assert(err, Equals, ErrorEntityNotFound)
	check.Assert(disk, IsNil)

	// enable fast provisioning if needed
	updateVdcFastProvisioning(ctx, vcd, check, previousProvisioningValue)

	// delete Vapp early to avoid env capacity issue
	deleteVapp(ctx, vcd, vmName)
}

// Test update internal disk for VM which has independent disk
func (vcd *TestVCD) Test_UpdateInternalDisk(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// In general VM internal disks works with Org users, but due we need change VDC fast provisioning value, we have to be sys admins
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	ctx := context.Background()
	vmName := "Test_UpdateInternalDisk"
	vm, storageProfile, diskSettings, diskId, previousProvisioningValue, err := vcd.createInternalDisk(ctx, check, vmName, 1)
	check.Assert(err, IsNil)

	//verify
	disk, err := vm.GetInternalDiskById(ctx, diskId, true)
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

	vmSpecSection, err = vm.UpdateInternalDisks(ctx, vmSpecSection)
	check.Assert(err, IsNil)
	check.Assert(vmSpecSection, NotNil)

	disk, err = vm.GetInternalDiskById(ctx, diskId, true)
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
	independentDisk, err := attachIndependentDisk(ctx, vcd, check)
	check.Assert(err, IsNil)

	//cleanup
	err = vm.DeleteInternalDisk(ctx, diskId)
	check.Assert(err, IsNil)
	detachIndependentDisk(ctx, vcd, check, independentDisk)

	// disable fast provisioning if needed
	updateVdcFastProvisioning(ctx, vcd, check, previousProvisioningValue)

	// delete Vapp early to avoid env capacity issue
	deleteVapp(ctx, vcd, vmName)
}

func attachIndependentDisk(ctx context.Context, vcd *TestVCD, check *C) (*Disk, error) {
	// Find VM
	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

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
	return disk, err
}

func detachIndependentDisk(ctx context.Context, vcd *TestVCD, check *C, disk *Disk) {
	err := vcd.detachIndependentDisk(ctx, Disk{disk.Disk, &vcd.client.Client})
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_VmGetParentvAppAndVdc(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}
	ctx := context.Background()

	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.findFirstVapp(ctx)
	if vapp.VApp.Name == "" {
		check.Skip("Disabled: No suitable vApp found in vDC")
	}
	vm, vmName := vcd.findFirstVm(vapp)
	if vm.Name == "" {
		check.Skip("Disabled: No suitable VM found in vDC")
	}

	newVM, err := vcd.client.Client.GetVMByHref(ctx, vm.HREF)
	check.Assert(err, IsNil)
	check.Assert(newVM.VM.Name, Equals, vmName)
	check.Assert(newVM.VM.VirtualHardwareSection.Item, NotNil)

	parentvApp, err := newVM.GetParentVApp(ctx)
	check.Assert(err, IsNil)
	check.Assert(parentvApp.VApp.HREF, Equals, vapp.VApp.HREF)

	parentVdc, err := newVM.GetParentVdc(ctx)
	check.Assert(err, IsNil)
	check.Assert(parentVdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
}

func (vcd *TestVCD) Test_AddNewEmptyVMMultiNIC(check *C) {

	config := vcd.config
	if config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}
	ctx := context.Background()

	vapp, err := createVappForTest(ctx, vcd, "Test_AddNewEmptyVMMultiNIC")
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
		vdcNetwork2, err := vcd.vdc.GetOrgVdcNetworkByName(ctx, vcd.config.VCD.Network.Net2, false)
		check.Assert(err, IsNil)
		_, err = vapp.AddOrgNetwork(ctx, &VappNetworkSettings{}, vdcNetwork2.OrgVDCNetwork, false)
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

	cat, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	media, err := cat.GetMediaByName(ctx, vcd.config.Media.Media, false)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)

	var task Task
	var sp types.Reference
	var customSP = false

	if vcd.config.VCD.StorageProfile.SP1 != "" {
		sp, _ = vcd.vdc.FindStorageProfileReference(ctx, vcd.config.VCD.StorageProfile.SP1)
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

	createdVm, err := vapp.AddEmptyVm(ctx, requestDetails)
	check.Assert(err, IsNil)
	check.Assert(createdVm, NotNil)

	// Ensure network config was valid
	actualNetConfig, err := createdVm.GetNetworkConnectionSection(ctx)
	check.Assert(err, IsNil)

	if customSP {
		check.Assert(createdVm.VM.StorageProfile.HREF, Equals, sp.HREF)
	}

	verifyNetworkConnectionSection(check, actualNetConfig, desiredNetConfig)

	// Cleanup
	err = vapp.RemoveVM(ctx, *createdVm)
	check.Assert(err, IsNil)

	// Ensure network is detached from vApp to avoid conflicts in other tests
	task, err = vapp.RemoveAllNetworks(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	task, err = vapp.Delete(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
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
	ctx := context.Background()

	vdc, _, vappTemplate, vapp, desiredNetConfig, err := vcd.createAndGetResourcesForVmCreation(ctx, check, vmName)
	check.Assert(err, IsNil)

	vm, err := spawnVM(ctx, "FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", true)
	check.Assert(err, IsNil)

	task, err := vm.PowerOff(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	vmSpecSection := vm.VM.VmSpecSection
	osType := "sles10_64Guest"
	vmSpecSection.OsType = osType
	vmSpecSection.NumCpus = takeIntAddress(4)
	vmSpecSection.NumCoresPerSocket = takeIntAddress(2)
	vmSpecSection.MemoryResourceMb = &types.MemoryResourceMb{Configured: 768}

	updatedVm, err := vm.UpdateVmSpecSection(ctx, vmSpecSection, "updateDescription")
	check.Assert(err, IsNil)
	check.Assert(updatedVm, NotNil)

	//verify
	check.Assert(updatedVm.VM.VmSpecSection.OsType, Equals, osType)
	check.Assert(*updatedVm.VM.VmSpecSection.NumCpus, Equals, 4)
	check.Assert(*updatedVm.VM.VmSpecSection.NumCoresPerSocket, Equals, 2)
	check.Assert(updatedVm.VM.VmSpecSection.MemoryResourceMb.Configured, Equals, int64(768))
	check.Assert(updatedVm.VM.Description, Equals, "updateDescription")

	// delete Vapp early to avoid env capacity issue
	deleteVapp(ctx, vcd, vmName)
}

func (vcd *TestVCD) Test_QueryVmList(check *C) {

	if vcd.skipVappTests {
		check.Skip("Test_QueryVmList needs an existing vApp to run")
		return
	}
	ctx := context.Background()

	// Get the setUp vApp using traditional methods
	vapp, err := vcd.vdc.GetVAppByName(ctx, TestSetUpSuite, true)
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
		list, err := vcd.vdc.QueryVmList(ctx, types.VmQueryFilter(filter))
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
	ctx := context.Background()

	vdc, _, vappTemplate, vapp, desiredNetConfig, err := vcd.createAndGetResourcesForVmCreation(ctx, check, vmName)
	check.Assert(err, IsNil)

	vm, err := spawnVM(ctx, "FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", true)
	check.Assert(err, IsNil)

	task, err := vm.PowerOff(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	check.Assert(vm.VM.VMCapabilities.MemoryHotAddEnabled, Equals, false)
	check.Assert(vm.VM.VMCapabilities.CPUHotAddEnabled, Equals, false)

	updatedVm, err := vm.UpdateVmCpuAndMemoryHotAdd(ctx, true, true)
	check.Assert(err, IsNil)
	check.Assert(updatedVm, NotNil)

	//verify
	check.Assert(updatedVm.VM.VMCapabilities.MemoryHotAddEnabled, Equals, true)
	check.Assert(updatedVm.VM.VMCapabilities.CPUHotAddEnabled, Equals, true)

	// delete Vapp early to avoid env capacity issue
	deleteVapp(ctx, vcd, vmName)
}

func (vcd *TestVCD) Test_AddNewEmptyVMWithVmComputePolicyAndUpdate(check *C) {
	ctx := context.Background()
	if vcd.client.Client.APIVCDMaxVersionIs(ctx, "< 33.0") {
		check.Skip(fmt.Sprintf("Test %s requires VCD 10.0 (API version 33) or higher", check.TestName()))
	}

	vapp, err := createVappForTest(ctx, vcd, "Test_AddNewEmptyVMWithVmComputePolicy")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	cat, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, true)
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
	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(ctx, vcd.vdc.Vdc.Name, false)
	if adminVdc == nil || err != nil {
		vcd.infoCleanup(notFoundMsg, "vdc", vcd.vdc.Vdc.Name)
	}

	createdPolicy, err := adminOrg.CreateVdcComputePolicy(ctx, newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)

	createdPolicy2, err := adminOrg.CreateVdcComputePolicy(ctx, newComputePolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_AddNewEmptyVMWithVmComputePolicyAndUpdate")
	AddToCleanupList(createdPolicy2.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_AddNewEmptyVMWithVmComputePolicyAndUpdate")

	vdcComputePolicyHref, err := adminOrg.client.OpenApiBuildEndpoint(types.OpenApiPathVersion1_0_0, types.OpenApiEndpointVdcComputePolicies)
	check.Assert(err, IsNil)

	// Get policy to existing ones (can be only default one)
	allAssignedComputePolicies, err := adminVdc.GetAllAssignedVdcComputePolicies(ctx, nil)
	check.Assert(err, IsNil)
	var policyReferences []*types.Reference
	for _, assignedPolicy := range allAssignedComputePolicies {
		policyReferences = append(policyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + assignedPolicy.VdcComputePolicy.ID})
	}
	policyReferences = append(policyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + createdPolicy.VdcComputePolicy.ID})
	policyReferences = append(policyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + createdPolicy2.VdcComputePolicy.ID})

	assignedVdcComputePolicies, err := adminVdc.SetAssignedComputePolicies(ctx, types.VdcComputePolicyReferences{VdcComputePolicyReference: policyReferences})
	check.Assert(err, IsNil)
	check.Assert(len(allAssignedComputePolicies)+2, Equals, len(assignedVdcComputePolicies.VdcComputePolicyReference))
	// end

	var task Task
	var sp types.Reference
	var customSP = false

	if vcd.config.VCD.StorageProfile.SP1 != "" {
		sp, _ = vcd.vdc.FindStorageProfileReference(ctx, vcd.config.VCD.StorageProfile.SP1)
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

	createdVm, err := vapp.AddEmptyVm(ctx, requestDetails)
	check.Assert(err, IsNil)
	check.Assert(createdVm, NotNil)
	check.Assert(createdVm.VM.ComputePolicy, NotNil)
	check.Assert(createdVm.VM.ComputePolicy.VmSizingPolicy, NotNil)
	check.Assert(createdVm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, createdPolicy.VdcComputePolicy.ID)

	if customSP {
		check.Assert(createdVm.VM.StorageProfile.HREF, Equals, sp.HREF)
	}

	updatedVm, err := createdVm.UpdateComputePolicy(ctx, createdPolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)
	check.Assert(updatedVm, NotNil)
	check.Assert(updatedVm.VM.ComputePolicy, NotNil)
	check.Assert(updatedVm.VM.ComputePolicy.VmSizingPolicy, NotNil)
	check.Assert(updatedVm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, createdPolicy2.VdcComputePolicy.ID)
	check.Assert(updatedVm.VM.VmSpecSection.MemoryResourceMb, NotNil)
	check.Assert(updatedVm.VM.VmSpecSection.MemoryResourceMb.Configured, Equals, int64(2048))

	// Cleanup
	err = vapp.RemoveVM(ctx, *createdVm)
	check.Assert(err, IsNil)

	// Ensure network is detached from vApp to avoid conflicts in other tests
	task, err = vapp.RemoveAllNetworks(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	task, err = vapp.Delete(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// cleanup assigned compute policy
	var beforeTestPolicyReferences []*types.Reference
	for _, assignedPolicy := range allAssignedComputePolicies {
		beforeTestPolicyReferences = append(beforeTestPolicyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + assignedPolicy.VdcComputePolicy.ID})
	}

	_, err = adminVdc.SetAssignedComputePolicies(ctx, types.VdcComputePolicyReferences{VdcComputePolicyReference: beforeTestPolicyReferences})
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_VMUpdateStorageProfile(check *C) {
	config := vcd.config
	if config.VCD.StorageProfile.SP1 == "" || config.VCD.StorageProfile.SP2 == "" {
		check.Skip("Skipping test because both storage profiles have to be configured")
	}
	ctx := context.Background()

	vapp, err := createVappForTest(ctx, vcd, "Test_VMUpdateStorageProfile")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	cat, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	var storageProfile types.Reference

	storageProfile, _ = vcd.vdc.FindStorageProfileReference(ctx, vcd.config.VCD.StorageProfile.SP1)

	createdVm, err := makeEmptyVm(ctx, vapp, "Test_VMUpdateStorageProfile")
	check.Assert(err, IsNil)
	check.Assert(createdVm, NotNil)
	check.Assert(createdVm.VM.StorageProfile.HREF, Equals, storageProfile.HREF)

	storageProfile2, _ := vcd.vdc.FindStorageProfileReference(ctx, vcd.config.VCD.StorageProfile.SP2)
	updatedVm, err := createdVm.UpdateStorageProfile(ctx, storageProfile2.HREF)
	check.Assert(err, IsNil)
	check.Assert(updatedVm, NotNil)
	check.Assert(createdVm.VM.StorageProfile.HREF, Equals, storageProfile2.HREF)

	// Cleanup
	var task Task
	err = vapp.RemoveVM(ctx, *createdVm)
	check.Assert(err, IsNil)

	// Ensure network is detached from vApp to avoid conflicts in other tests
	task, err = vapp.RemoveAllNetworks(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	task, err = vapp.Delete(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
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
	ctx := context.Background()
	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	vdc, err := adminOrg.GetVDCByName(ctx, vcd.vdc.Vdc.Name, false)
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
	vm, err := vdc.CreateStandaloneVm(ctx, &params)
	check.Assert(err, IsNil)
	check.Assert(vm, NotNil)
	AddToCleanupList(vm.VM.ID, "standaloneVm", "", check.TestName())
	check.Assert(vm.VM.Description, Equals, description)

	_ = vdc.Refresh(ctx)
	vappList = vdc.GetVappList()
	check.Assert(len(vappList), Equals, vappNum+1)
	for _, vapp := range vappList {
		printVerbose("vapp: %s\n", vapp.Name)
	}
	err = vm.Delete(ctx)
	check.Assert(err, IsNil)
	_ = vdc.Refresh(ctx)
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
	ctx := context.Background()
	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	vdc, err := adminOrg.GetVDCByName(ctx, vcd.vdc.Vdc.Name, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	catalog, err := adminOrg.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	catalogItem, err := catalog.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catalogItem, NotNil)

	vappTemplate, err := catalog.GetVappTemplateByHref(ctx, catalogItem.CatalogItem.Entity.HREF)
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
	vm, err := vdc.CreateStandaloneVMFromTemplate(ctx, &params)
	check.Assert(err, IsNil)
	check.Assert(vm, NotNil)
	AddToCleanupList(vm.VM.ID, "standaloneVm", "", check.TestName())

	_ = vdc.Refresh(ctx)
	vappList = vdc.GetVappList()
	check.Assert(len(vappList), Equals, vappNum+1)
	for _, vapp := range vappList {
		printVerbose("vapp: %s\n", vapp.Name)
	}

	err = vm.Delete(ctx)
	check.Assert(err, IsNil)
	_ = vdc.Refresh(ctx)
	vappList = vdc.GetVappList()
	check.Assert(len(vappList), Equals, vappNum)
}
