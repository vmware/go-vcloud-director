// +build vm functional ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
* Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func init() {
	testingTags["vm"] = "vm_test.go"
}

func (vcd *TestVCD) findFirstVm(vapp VApp) (types.VM, string) {
	for _, vm := range vapp.VApp.Children.VM {
		if vm.Name != "" {
			return *vm, vm.Name
		}
	}
	return types.VM{}, ""
}

func (vcd *TestVCD) findFirstVapp() VApp {
	client := vcd.client
	config := vcd.config
	org, err := client.GetOrgByName(config.VCD.Org)
	if err != nil {
		fmt.Println(err)
		return VApp{}
	}
	vdc, err := org.GetVDCByName(config.VCD.Vdc, false)
	if err != nil {
		fmt.Println(err)
		return VApp{}
	}
	wantedVapp := vcd.vapp.VApp.Name
	vappName := ""
	for _, res := range vdc.Vdc.ResourceEntities {
		for _, item := range res.ResourceEntity {
			// Finding a named vApp, if it was defined in config
			if wantedVapp != "" {
				if item.Name == wantedVapp {
					vappName = item.Name
					break
				}
			} else {
				// Otherwise, we get the first vApp from the vDC list
				if item.Type == "application/vnd.vmware.vcloud.vApp+xml" {
					vappName = item.Name
					break
				}
			}
		}
	}
	if wantedVapp == "" {
		return VApp{}
	}
	vapp, _ := vdc.GetVAppByName(vappName, false)
	return *vapp
}

// Ensure vApp is suitable for VM test
// Some VM tests may fail if vApp is not powered on, so VM tests can call this function to ensure the vApp is suitable for VM tests
func (vcd *TestVCD) ensureVappIsSuitableForVMTest(vapp VApp) error {
	status, err := vapp.GetStatus()

	if err != nil {
		return err
	}

	// If vApp is not powered on (status = 4), power on vApp
	if status != types.VAppStatuses[4] {
		task, err := vapp.PowerOn()
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

// Ensure VM is suitable for VM test
// Please call ensureVappAvailableForVMTest first to power on the vApp because this function cannot handle VM in suspension state due to lack of VM APIs (e.g. discard VM suspend API)
// Some VM tests may fail if VM is not powered on or powered off, so VM tests can call this function to ensure the VM is suitable for VM tests
func (vcd *TestVCD) ensureVMIsSuitableForVMTest(vm *VM) error {
	// if the VM is not powered on (status = 4) or not powered off, wait for the VM power on
	// wait for around 1 min
	valid := false
	for i := 0; i < 6; i++ {
		status, err := vm.GetStatus()
		if err != nil {
			return err
		}

		// If the VM is powered on (status = 4)
		if status == types.VAppStatuses[4] {
			// Prevent affect Test_ChangeMemorySize
			// because TestVCD.Test_AttachedVMDisk is run before Test_ChangeMemorySize and Test_ChangeMemorySize will fail the test if the VM is powered on,
			task, err := vm.PowerOff()
			if err != nil {
				return err
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return err
			}
		}

		// If the VM is powered on (status = 4) or powered off (status = 8)
		if status == types.VAppStatuses[4] || status == types.VAppStatuses[8] {
			valid = true
		}

		// If 1st to 5th attempt is completed, sleep 10 seconds and try again
		// The last attempt will exit this for loop immediately, so no need to sleep
		if i < 5 {
			time.Sleep(time.Second * 10)
		}
	}

	if !valid {
		return errors.New("the VM is not powered on or powered off")
	}

	return nil
}

func (vcd *TestVCD) Test_FindVMByHREF(check *C) {
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
		Size:        vcd.config.VCD.Disk.Size,
		Description: TestVMAttachOrDetachDisk,
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

	// Discard vApp suspension
	// Disk attach and detach operations are not working if vApp is suspended
	err := vcd.ensureVappIsSuitableForVMTest(vapp)
	check.Assert(err, IsNil)
	err = vcd.ensureVMIsSuitableForVMTest(vm)
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
		Name:        TestVMDetachDisk,
		Size:        vcd.config.VCD.Disk.Size,
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

}

// Test Insert or Eject Media for VM
func (vcd *TestVCD) Test_HandleInsertOrEjectMedia(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	itemName := "TestHandleInsertOrEjectMedia"

	// Find VApp
	if vcd.vapp.VApp == nil {
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

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_HandleInsertOrEjectMedia")

	media, err := FindMediaAsCatalogItem(vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)
	check.Assert(media, Not(Equals), CatalogItem{})

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

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	// Skipping this test due to a bug in vCD. VM refresh status returns old state, though eject task is finished.
	if vcd.client.APIVCDMaxVersionIs(">= 32.0, <= 33.0") {
		check.Skip("Skipping test because this vCD version has a bug")
	}

	itemName := "TestInsertOrEjectMedia"

	// Find VApp
	if vcd.vapp.VApp == nil {
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

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_InsertOrEjectMedia")

	media, err := FindMediaAsCatalogItem(vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)
	check.Assert(media, Not(Equals), CatalogItem{})

	// Insert Media
	insertMediaTask, err := vm.insertOrEjectMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.CatalogItem.Entity.HREF,
			Name: media.CatalogItem.Entity.Name,
			ID:   media.CatalogItem.Entity.ID,
			Type: media.CatalogItem.Entity.Type,
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
			HREF: media.CatalogItem.Entity.HREF,
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
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	itemName := "TestAnswerVmQuestion"

	// Find VApp
	if vcd.vapp.VApp == nil {
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

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_AnswerVmQuestion")

	media, err := FindMediaAsCatalogItem(vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)
	check.Assert(media, Not(Equals), CatalogItem{})

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

	currentCpus := 0
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
	cpuCount := 4

	task, err := vm.ChangeCPUCountWithCore(cpuCount, &cores)
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
	task, err = vm.ChangeCPUCountWithCore(currentCpus, &currentCores)
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
	check.Assert(err, Not(IsNil))

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

// Test_VMGetGuestCustomizationSection retrieves guest customization and checks if properties are right.
func (vcd *TestVCD) Test_VMGetGuestCustomizationSection(check *C) {
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

// Test gathering VM virtual hardware items
func (vcd *TestVCD) Test_GetVirtualHardwareSection(check *C) {
	itemName := "TestGetVirtualHardwareSection"

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}

	fmt.Printf("Running: %s\n", itemName)

	// Find VM
	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	// Preform check of virtual hardware section
	section, err := vm.GetVirtualHardwareSection()
	check.Assert(err, IsNil)

	// Check that section.Info is not Nil, as its the only field that may not be omitted when marshalled
	check.Assert(section.Info, NotNil)

	// Check that section.Item is not Nil before looping over it
	check.Assert(section.Item, NotNil)

	// Loop over the Items to ensure
	for _, item := range section.Item {
		check.Assert(item.ResourceType, NotNil)
		check.Assert(item.ResourceSubType, NotNil)

		if item.ResourceType == 10 {
			check.Assert(item.Address, Matches, "^[a-fA-F0-9:]{17}|[a-fA-F0-9]{12}$")
		}
	}
}
