//go:build vm || functional || ALL

/*
* Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
* Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
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

	err = vm.Refresh()
	check.Assert(err, IsNil)

	// Expecting outcome to be 'false', but VCD sometimes lags to report that media is ejected in VM
	// state (even after ejection task is finished).
	// In such cases, do additional VM refresh and try to see if the structure has no media anymore
	mediaInjected := isMediaInjected(vm.VM.VirtualHardwareSection.Item)
	retryCount := 5
	if mediaInjected {
		// Run a few more attempts every second to see if the VM finally shows no media being
		// present in VirtualHardwareItem structure
		for a := 0; a < retryCount; a++ {
			fmt.Printf("attempt %d\n", a)
			err = vm.Refresh()
			if err != nil {
				fmt.Printf("error refreshing VM: %s\n", err)
			}
			retryIsMediaInjected := isMediaInjected(vm.VM.VirtualHardwareSection.Item)
			if !retryIsMediaInjected {
				fmt.Printf("media error recovered at %d refresh attempt\n", a)
				check.SucceedNow()
			}
			time.Sleep(time.Second)
		}
		check.Errorf("error was not recovered after %d attempts - ejection FAILED", retryCount)
	}
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

	// Remove catalog item so far other tests don't fail
	task, err := media.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
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

	_, vm := createNsxtVAppAndVm(vcd, check)

	nestingStatus := vm.VM.NestedHypervisorEnabled
	check.Assert(nestingStatus, Equals, false)

	// PowerOn
	task, err := vm.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Try to change the setting on powered on VM to fail
	_, err = vm.ToggleHardwareVirtualization(true)
	check.Assert(err, ErrorMatches, ".*hardware virtualization can be changed from powered off state.*")

	// Undeploy, so the VM goes to POWERED_OFF state instead of PARTIALLY_POWERED_OFF
	task, err = vm.Undeploy()
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

	err = deleteNsxtVapp(vcd, check.TestName())
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_VMPowerOnPowerOff(check *C) {
	_, vm := createNsxtVAppAndVm(vcd, check)

	// Ensure VM is not powered on
	vmStatus, err := vm.GetStatus()
	check.Assert(err, IsNil)
	if vmStatus != "POWERED_OFF" && vmStatus != "PARTIALLY_POWERED_OFF" {
		fmt.Printf("VM status: %s, powering off", vmStatus)
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

	task, err = vm.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	vmStatus, err = vm.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vmStatus == "POWERED_OFF" || vmStatus == "PARTIALLY_POWERED_OFF", Equals, true)

	err = deleteNsxtVapp(vcd, check.TestName())
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_VmShutdown(check *C) {
	vapp, vm := createNsxtVAppAndVm(vcd, check)

	vdc, err := vm.GetParentVdc()
	check.Assert(err, IsNil)

	// Ensure VM is not powered on
	vmStatus, err := vm.GetStatus()
	check.Assert(err, IsNil)
	fmt.Println("VM status: ", vmStatus)

	if vmStatus != "POWERED_ON" {
		task, err := vm.PowerOn()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		check.Assert(task.Task.Status, Equals, "success")
		err = vm.Refresh()
		check.Assert(err, IsNil)
		vmStatus, err = vm.GetStatus()
		check.Assert(err, IsNil)
		fmt.Println("VM status: ", vmStatus)
	}

	timeout := time.Minute * 5 // Avoiding infinite loops
	startTime := time.Now()
	elapsed := time.Since(startTime)
	gcStatus := ""
	statusFound := false
	// Wait until Guest Tools gets to `REBOOT_PENDING` or `GC_COMPLETE` as there is no real way to
	// check if VM has Guest Tools operating
	for elapsed < timeout {
		err = vm.Refresh()
		check.Assert(err, IsNil)

		vmQuery, err := vdc.QueryVM(vapp.VApp.Name, vm.VM.Name)
		check.Assert(err, IsNil)

		gcStatus = vmQuery.VM.GcStatus
		printVerbose("VM Tools Status: %s (%s)\n", vmQuery.VM.GcStatus, elapsed)
		if vmQuery.VM.GcStatus == "GC_COMPLETE" || vmQuery.VM.GcStatus == "REBOOT_PENDING" {
			statusFound = true
			break
		}

		time.Sleep(5 * time.Second)
		elapsed = time.Since(startTime)
	}
	fmt.Printf("VM Tools Status: %s (%s)\n", gcStatus, elapsed)
	check.Assert(statusFound, Equals, true)

	printVerbose("Shutting down VM:\n")

	task, err := vm.Shutdown()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	newStatus, err := vm.GetStatus()
	check.Assert(err, IsNil)
	printVerbose("New VM status: %s\n", newStatus)
	check.Assert(newStatus, Equals, "POWERED_OFF")

	err = deleteNsxtVapp(vcd, check.TestName())
	check.Assert(err, IsNil)
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

	fmt.Printf("Running: %s\n", check.TestName())

	_, vm := createNsxtVAppAndVm(vcd, check)

	// It may be that prebuilt VM was not booted before in the test vApp and it would still have
	// a guest customization status 'GC_PENDING'. This is because initially VM has this flag set
	// but while this flag is here the test cannot actually check if vm.PowerOnAndForceCustomization()
	// gives any effect therefore we must "wait through" initial guest customization if it is in
	// 'GC_PENDING' state.
	custStatus, err := vm.GetGuestCustomizationStatus()
	check.Assert(err, IsNil)

	vmStatus, err := vm.GetStatus()
	check.Assert(err, IsNil)
	if custStatus == types.GuestCustStatusPending {
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

	err = deleteNsxtVapp(vcd, check.TestName())
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
	err = deleteVapp(vcd, vmName)
	check.Assert(err, IsNil)
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
	err = deleteVapp(vcd, vmName)
	check.Assert(err, IsNil)
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

	// verify VM description still available - test for bugfix #418
	description := "Test_UpdateInternalDisk_Description"
	vm, err = vm.UpdateVmSpecSection(vm.VM.VmSpecSection, description)
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

	// verify VM description still available - test for bugfix #418
	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(vm.VM.Description, Equals, description)

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
	err = deleteVapp(vcd, vmName)
	check.Assert(err, IsNil)
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

	vapp, err := deployVappForTest(vcd, "Test_AddNewEmptyVMMultiNIC")
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
		ThinProvisioned:   addrOf(true),
		OverrideVmDefault: true}

	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      "Test_AddNewEmptyVMMultiNIC",
			NetworkConnectionSection:  desiredNetConfig,
			Description:               "created by Test_AddNewEmptyVMMultiNIC",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          addrOf(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           addrOf(2),
				NumCoresPerSocket: addrOf(1),
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

	// #nosec G101 -- Not a credential
	vmName := "Test_UpdateVmSpecSection"
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}

	vdc, _, vappTemplate, vapp, desiredNetConfig, err := vcd.createAndGetResourcesForVmCreation(check, vmName)
	check.Assert(err, IsNil)

	vm, err := spawnVM("FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", false)
	check.Assert(err, IsNil)

	vmSpecSection := vm.VM.VmSpecSection
	osType := "sles10_64Guest"
	vmSpecSection.OsType = osType
	vmSpecSection.NumCpus = addrOf(4)
	vmSpecSection.NumCoresPerSocket = addrOf(2)
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
	err = deleteVapp(vcd, vmName)
	check.Assert(err, IsNil)
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

	vm, err := spawnVM("FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", false)
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
	err = deleteVapp(vcd, vmName)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_AddNewEmptyVMWithVmComputePolicyAndUpdate(check *C) {
	vcd.skipIfNotSysAdmin(check)
	vapp, err := deployVappForTest(vcd, "Test_AddNewEmptyVMWithVmComputePolicy")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	newComputePolicy := &VdcComputePolicy{
		client: vcd.org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:        check.TestName() + "_empty",
			Description: addrOf("Empty policy created by test"),
		},
	}

	newComputePolicy2 := &VdcComputePolicy{
		client: vcd.org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:        check.TestName() + "_memory",
			Description: addrOf("Empty policy created by test 2"),
			Memory:      addrOf(2048),
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

	createdPolicy, err := adminOrg.client.CreateVdcComputePolicy(newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)

	createdPolicy2, err := adminOrg.client.CreateVdcComputePolicy(newComputePolicy2.VdcComputePolicy)
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
		ThinProvisioned:   addrOf(true),
		OverrideVmDefault: true}

	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      "Test_AddNewEmptyVMWithVmComputePolicy",
			Description:               "created by Test_AddNewEmptyVMWithVmComputePolicy",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          addrOf(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           addrOf(2),
				NumCoresPerSocket: addrOf(1),
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

	vapp, err := deployVappForTest(vcd, "Test_VMUpdateStorageProfile")
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

func (vcd *TestVCD) Test_VMUpdateComputePolicies(check *C) {
	vcd.skipIfNotSysAdmin(check)
	providerVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)
	check.Assert(providerVdc, NotNil)

	vmGroup, err := vcd.client.GetVmGroupByNameAndProviderVdcUrn(vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup, providerVdc.ProviderVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(vmGroup, NotNil)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.nsxtVdc.Vdc.Name, false)
	if adminVdc == nil || err != nil {
		vcd.infoCleanup(notFoundMsg, "vdc", vcd.nsxtVdc.Vdc.Name)
	}

	// Create some Compute Policies
	var placementPolicies []*VdcComputePolicyV2
	var sizingPolicies []*VdcComputePolicyV2
	numberOfPolicies := 2
	for i := 0; i < numberOfPolicies; i++ {
		sizingPolicyName := fmt.Sprintf("%s_Sizing%d", check.TestName(), i+1)
		placementPolicyName := fmt.Sprintf("%s_Placement%d", check.TestName(), i+1)

		sizingPolicies = append(sizingPolicies, &VdcComputePolicyV2{
			VdcComputePolicyV2: &types.VdcComputePolicyV2{
				VdcComputePolicy: types.VdcComputePolicy{
					Name:         sizingPolicyName,
					Description:  addrOf("Empty sizing policy created by test"),
					IsSizingOnly: true,
				},
				PolicyType: "VdcVmPolicy",
			},
		})

		placementPolicies = append(placementPolicies, &VdcComputePolicyV2{
			VdcComputePolicyV2: &types.VdcComputePolicyV2{
				VdcComputePolicy: types.VdcComputePolicy{
					Name:         placementPolicyName,
					Description:  addrOf("Empty placement policy created by test"),
					IsSizingOnly: false,
				},
				PolicyType: "VdcVmPolicy",
				PvdcNamedVmGroupsMap: []types.PvdcNamedVmGroupsMap{
					{
						NamedVmGroups: []types.OpenApiReferences{{
							{
								Name: vmGroup.VmGroup.Name,
								ID:   fmt.Sprintf("urn:vcloud:namedVmGroup:%s", vmGroup.VmGroup.NamedVmGroupId),
							},
						}},
						Pvdc: types.OpenApiReference{
							Name: providerVdc.ProviderVdc.Name,
							ID:   providerVdc.ProviderVdc.ID,
						},
					},
				},
			},
		})

		sizingPolicies[i], err = vcd.client.CreateVdcComputePolicyV2(sizingPolicies[i].VdcComputePolicyV2)
		check.Assert(err, IsNil)
		AddToCleanupList(sizingPolicies[i].VdcComputePolicyV2.ID, "vdcComputePolicy", vcd.org.Org.Name, sizingPolicyName)

		placementPolicies[i], err = vcd.client.CreateVdcComputePolicyV2(placementPolicies[i].VdcComputePolicyV2)
		check.Assert(err, IsNil)
		AddToCleanupList(placementPolicies[i].VdcComputePolicyV2.ID, "vdcComputePolicy", vcd.org.Org.Name, placementPolicyName)
	}

	vdcComputePolicyHref, err := adminOrg.client.OpenApiBuildEndpoint(types.OpenApiPathVersion2_0_0, types.OpenApiEndpointVdcComputePolicies)
	check.Assert(err, IsNil)

	// Add the created compute policies to the ones that the VDC has already assigned
	alreadyAssignedPolicies, err := adminVdc.GetAllAssignedVdcComputePoliciesV2(nil)
	check.Assert(err, IsNil)
	var allComputePoliciesToAssign []*types.Reference
	for _, alreadyAssignedPolicy := range alreadyAssignedPolicies {
		allComputePoliciesToAssign = append(allComputePoliciesToAssign, &types.Reference{HREF: vdcComputePolicyHref.String() + alreadyAssignedPolicy.VdcComputePolicyV2.ID})
	}
	for i := 0; i < numberOfPolicies; i++ {
		allComputePoliciesToAssign = append(allComputePoliciesToAssign, &types.Reference{HREF: vdcComputePolicyHref.String() + sizingPolicies[i].VdcComputePolicyV2.ID})
		allComputePoliciesToAssign = append(allComputePoliciesToAssign, &types.Reference{HREF: vdcComputePolicyHref.String() + placementPolicies[i].VdcComputePolicyV2.ID})
	}

	assignedVdcComputePolicies, err := adminVdc.SetAssignedComputePolicies(types.VdcComputePolicyReferences{VdcComputePolicyReference: allComputePoliciesToAssign})
	check.Assert(err, IsNil)
	check.Assert(len(alreadyAssignedPolicies)+numberOfPolicies*2, Equals, len(assignedVdcComputePolicies.VdcComputePolicyReference))

	vapp, vm := createNsxtVAppAndVm(vcd, check)
	check.Assert(vapp, NotNil)
	check.Assert(vm, NotNil)

	// Update all Compute Policies: Sizing and Placement
	check.Assert(err, IsNil)
	vm, err = vm.UpdateComputePolicyV2(sizingPolicies[0].VdcComputePolicyV2.ID, placementPolicies[0].VdcComputePolicyV2.ID, "")
	check.Assert(err, IsNil)
	check.Assert(vm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, sizingPolicies[0].VdcComputePolicyV2.ID)
	check.Assert(vm.VM.ComputePolicy.VmPlacementPolicy.ID, Equals, placementPolicies[0].VdcComputePolicyV2.ID)

	// Update Sizing policy only
	vm, err = vm.UpdateComputePolicyV2(sizingPolicies[0].VdcComputePolicyV2.ID, placementPolicies[1].VdcComputePolicyV2.ID, "")
	check.Assert(err, IsNil)
	check.Assert(vm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, sizingPolicies[0].VdcComputePolicyV2.ID)
	check.Assert(vm.VM.ComputePolicy.VmPlacementPolicy.ID, Equals, placementPolicies[1].VdcComputePolicyV2.ID)

	// Update Placement policy only
	vm, err = vm.UpdateComputePolicyV2(sizingPolicies[1].VdcComputePolicyV2.ID, placementPolicies[1].VdcComputePolicyV2.ID, "")
	check.Assert(err, IsNil)
	check.Assert(vm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, sizingPolicies[1].VdcComputePolicyV2.ID)
	check.Assert(vm.VM.ComputePolicy.VmPlacementPolicy.ID, Equals, placementPolicies[1].VdcComputePolicyV2.ID)

	// Remove Placement Policy
	vm, err = vm.UpdateComputePolicyV2(sizingPolicies[1].VdcComputePolicyV2.ID, "", "")
	check.Assert(err, IsNil)
	check.Assert(vm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, sizingPolicies[1].VdcComputePolicyV2.ID)
	check.Assert(vm.VM.ComputePolicy.VmPlacementPolicy, IsNil)

	// Remove Sizing Policy
	vm, err = vm.UpdateComputePolicyV2("", placementPolicies[1].VdcComputePolicyV2.ID, "")
	check.Assert(err, IsNil)
	check.Assert(vm.VM.ComputePolicy.VmSizingPolicy, IsNil)
	check.Assert(vm.VM.ComputePolicy.VmPlacementPolicy.ID, Equals, placementPolicies[1].VdcComputePolicyV2.ID)

	// Try to remove both, it should fail
	_, err = vm.UpdateComputePolicyV2("", "", "")
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), "either sizing policy ID or placement policy ID is needed"))

	// Clean VM
	task, err := vapp.Undeploy()
	check.Assert(err, IsNil)
	check.Assert(task, Not(Equals), Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	check.Assert(task, Not(Equals), Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Cleanup assigned compute policies
	var beforeTestPolicyReferences []*types.Reference
	for _, assignedPolicy := range alreadyAssignedPolicies {
		beforeTestPolicyReferences = append(beforeTestPolicyReferences, &types.Reference{HREF: vdcComputePolicyHref.String() + assignedPolicy.VdcComputePolicyV2.ID})
	}

	_, err = adminVdc.SetAssignedComputePolicies(types.VdcComputePolicyReferences{VdcComputePolicyReference: beforeTestPolicyReferences})
	check.Assert(err, IsNil)
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
				Modified:          addrOf(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           addrOf(1),
				NumCoresPerSocket: addrOf(1),
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
							ThinProvisioned:   addrOf(true),
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

	vmName := "testStandaloneTemplate"
	vmDescription := "Standalone VM"
	params := types.InstantiateVmTemplateParams{
		Xmlns:            types.XMLNamespaceVCloud,
		Name:             vmName,
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
			StorageProfile: nil,
			VmCapabilities: nil,
			VmGeneralParams: &types.VMGeneralParams{
				Name:               vmName,
				Description:        vmDescription,
				NeedsCustomization: false,
				RegenerateBiosUuid: false,
			},
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
	check.Assert(vm.VM.Name, Equals, vmName)
	check.Assert(vm.VM.Description, Equals, vmDescription)

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

func (vcd *TestVCD) Test_VMChangeCPU(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	currentCpus := existingVm.VmSpecSection.NumCpus
	currentCores := existingVm.VmSpecSection.NumCoresPerSocket

	check.Assert(0, Not(Equals), currentCpus)
	check.Assert(0, Not(Equals), currentCores)

	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	cores := 2
	cpuCount := 4

	err = vm.ChangeCPU(cpuCount, cores)
	check.Assert(err, IsNil)

	check.Assert(*vm.VM.VmSpecSection.NumCpus, Equals, cpuCount)
	check.Assert(*vm.VM.VmSpecSection.NumCoresPerSocket, Equals, cores)

	// return to previous value
	err = vm.ChangeCPU(*currentCpus, *currentCores)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_VMChangeCPUAndCoreCount(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}

	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	currentCpus := existingVm.VmSpecSection.NumCpus
	currentCores := existingVm.VmSpecSection.NumCoresPerSocket

	check.Assert(0, Not(Equals), currentCpus)
	check.Assert(0, Not(Equals), currentCores)

	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	cores := 2
	cpuCount := 4

	err = vm.ChangeCPUAndCoreCount(&cpuCount, &cores)
	check.Assert(err, IsNil)

	check.Assert(*vm.VM.VmSpecSection.NumCpus, Equals, cpuCount)
	check.Assert(*vm.VM.VmSpecSection.NumCoresPerSocket, Equals, cores)

	// Try changing only CPU count and seeing if coreCount remains the same
	newCpuCount := 2
	err = vm.ChangeCPUAndCoreCount(&newCpuCount, nil)
	check.Assert(err, IsNil)

	check.Assert(*vm.VM.VmSpecSection.NumCpus, Equals, newCpuCount)
	check.Assert(*vm.VM.VmSpecSection.NumCoresPerSocket, Equals, cores)

	// Change only core count and check that CPU count remains as it was
	newCoreCount := 1
	err = vm.ChangeCPUAndCoreCount(nil, &newCoreCount)
	check.Assert(err, IsNil)

	check.Assert(*vm.VM.VmSpecSection.NumCpus, Equals, newCpuCount)
	check.Assert(*vm.VM.VmSpecSection.NumCoresPerSocket, Equals, newCoreCount)

	// return to previous value
	err = vm.ChangeCPUAndCoreCount(currentCpus, currentCores)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_VMChangeMemory(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	check.Assert(existingVm.VmSpecSection.MemoryResourceMb, Not(IsNil))

	currentMemory := existingVm.VmSpecSection.MemoryResourceMb.Configured
	check.Assert(0, Not(Equals), currentMemory)

	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	err = vm.ChangeMemory(2304)
	check.Assert(err, IsNil)

	check.Assert(existingVm.VmSpecSection.MemoryResourceMb, Not(IsNil))
	check.Assert(vm.VM.VmSpecSection.MemoryResourceMb.Configured, Equals, int64(2304))

	// return to previous value
	err = vm.ChangeMemory(currentMemory)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_AddRawVm(check *C) {
	vapp, vm := createNsxtVAppAndVm(vcd, check)
	check.Assert(vapp, NotNil)
	check.Assert(vm, NotNil)

	// Check that vApp did not lose its state
	vappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vappStatus, Equals, "MIXED") //vApp is powered on, but the VM within is powered off
	check.Assert(vapp.VApp.Name, Equals, check.TestName())
	check.Assert(vapp.VApp.Description, Equals, check.TestName())

	// Check that VM is not powered on
	vmStatus, err := vm.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vmStatus, Equals, "POWERED_OFF")

	// Cleanup
	task, err := vapp.Undeploy()
	check.Assert(err, IsNil)
	check.Assert(task, Not(Equals), Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	check.Assert(task, Not(Equals), Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

func createNsxtVAppAndVm(vcd *TestVCD, check *C) (*VApp, *VM) {
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.NsxtBackedCatalogName, false)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)
	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.NsxtCatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catitem, NotNil)
	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vapptemplate.VAppTemplate.Children.VM[0].HREF, NotNil)

	vapp, err := vcd.nsxtVdc.CreateRawVApp(check.TestName(), check.TestName())
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)
	// After a successful creation, the entity is added to the cleanup list.
	AddToCleanupList(vapp.VApp.Name, "vapp", vcd.nsxtVdc.Vdc.Name, check.TestName())

	// Check that vApp is powered-off
	vappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vappStatus, Equals, "RESOLVED")

	task, err := vapp.PowerOn()
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	vappStatus, err = vapp.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vappStatus, Equals, "POWERED_ON")

	// Once the operation is successful, we won't trigger a failure
	// until after the vApp deletion
	check.Check(vapp.VApp.Name, Equals, check.TestName())
	check.Check(vapp.VApp.Description, Equals, check.TestName())

	// Construct VM
	vmDef := &types.ReComposeVAppParams{
		Ovf:              types.XMLNamespaceOVF,
		Xsi:              types.XMLNamespaceXSI,
		Xmlns:            types.XMLNamespaceVCloud,
		AllEULAsAccepted: true,
		// Deploy:           false,
		Name: vapp.VApp.Name,
		// PowerOn: false, // Not touching power state at this phase
		SourcedItem: &types.SourcedCompositionItemParam{
			Source: &types.Reference{
				HREF: vapptemplate.VAppTemplate.Children.VM[0].HREF,
				Name: check.TestName() + "-vm-tmpl",
			},
			VMGeneralParams: &types.VMGeneralParams{
				Description: "test-vm-description",
			},
			InstantiationParams: &types.InstantiationParams{
				NetworkConnectionSection: &types.NetworkConnectionSection{},
			},
		},
	}
	vm, err := vapp.AddRawVM(vmDef)
	check.Assert(err, IsNil)
	check.Assert(vm, NotNil)
	check.Assert(vm.VM.Name, Equals, vmDef.SourcedItem.Source.Name)

	// Refresh vApp to have latest state
	err = vapp.Refresh()
	check.Assert(err, IsNil)

	return vapp, vm
}

func (vcd *TestVCD) Test_GetOvfEnvironment(check *C) {

	version, err := vcd.client.Client.GetVcdShortVersion()
	check.Assert(err, IsNil)
	if version == "10.5.0" {
		check.Skip("There is a known bug with the OVF environment on 10.5.0")
	}

	_, vm := createNsxtVAppAndVm(vcd, check)
	check.Assert(vm, NotNil)

	task, err := vm.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Read ovfenv when VM is started
	ovfenv, err := vm.GetEnvironment()
	check.Assert(err, IsNil)
	check.Assert(ovfenv, NotNil)

	// Provides information from the virtualization platform like VM moref
	check.Assert(strings.Contains(ovfenv.VCenterId, "vm-"), Equals, true)

	// Check virtualization platform Vendor
	check.Assert(ovfenv.PlatformSection, NotNil)
	check.Assert(ovfenv.PlatformSection.Vendor, Equals, "VMware, Inc.")

	// Check guest operating system level configuration for hostname
	check.Assert(ovfenv.PropertySection, NotNil)
	for _, p := range ovfenv.PropertySection.Properties {
		if p.Key == "vCloud_computerName" {
			check.Assert(p.Value, Not(Equals), "")
		}
	}
	check.Assert(ovfenv.EthernetAdapterSection, NotNil)
	for _, p := range ovfenv.EthernetAdapterSection.Adapters {
		check.Assert(p.Mac, Not(Equals), "")
	}

	err = deleteNsxtVapp(vcd, check.TestName())
	check.Assert(err, IsNil)
}
