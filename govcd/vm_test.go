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
	vapp, _ := vdc.FindVAppByName(vappName)
	return vapp
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
	vm, vm_name := vcd.findFirstVm(vapp)
	if vm.Name == "" {
		check.Skip("Disabled: No suitable VM found in vDC")
	}

	vm_href := vm.HREF
	new_vm, err := vcd.client.Client.FindVMByHREF(vm_href)

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
	catalog, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_HandleInsertOrEjectMedia")

	media, err := FindMediaAsCatalogItem(&vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)
	check.Assert(media, Not(Equals), CatalogItem{})

	insertMediaTask, err := vm.HandleInsertMedia(&vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	err = insertMediaTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, true)

	ejectMediaTask, err := vm.HandleEjectMedia(&vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	err = ejectMediaTask.WaitTaskCompletion(true)
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, false)
}

// Test Insert or Eject Media for VM
func (vcd *TestVCD) Test_InsertOrEjectMedia(check *C) {

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
	catalog, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_InsertOrEjectMedia")

	media, err := FindMediaAsCatalogItem(&vcd.org, vcd.config.VCD.Catalog.Name, itemName)
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

func (vcd *TestVCD) Test_AddMetadataOnVm(check *C) {
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

	// Add metadata
	task, err := vm.AddMetadata("key", "value")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Check if metadata was added correctly
	metadata, err := vm.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata.MetadataEntry[0].Key, Equals, "key")
	check.Assert(metadata.MetadataEntry[0].TypedValue.Value, Equals, "value")
}

func (vcd *TestVCD) Test_DeleteMetadataOnVm(check *C) {
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

	// Add metadata
	task, err := vm.AddMetadata("key2", "value2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Remove metadata
	task, err = vm.DeleteMetadata("key2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	metadata, err := vm.GetMetadata()
	check.Assert(err, IsNil)
	for _, k := range metadata.MetadataEntry {
		if k.Key == "key2" {
			check.Errorf("metadata.MetadataEntry should not contain key: %s", k)
		}
	}
}

// check resource subtype for specific value which means media is injected
func isMediaInjected(items []*types.VirtualHardwareItem) bool {
	isFound := false
	for _, hardwareItem := range items {
		if hardwareItem.ResourceSubType == types.VMsCDResourceSubType {
			isFound = true
			break
		}
	}
	return isFound
}

// Test Insert or Eject Media for VM
func (vcd *TestVCD) Test_AnswerVmQuestion(check *C) {

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
	catalog, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_AnswerVmQuestion")

	media, err := FindMediaAsCatalogItem(&vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)
	check.Assert(media, Not(Equals), CatalogItem{})

	err = vm.Refresh()
	check.Assert(err, IsNil)

	insertMediaTask, err := vm.HandleInsertMedia(&vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)

	err = insertMediaTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(isMediaInjected(vm.VM.VirtualHardwareSection.Item), Equals, true)

	ejectMediaTask, err := vm.HandleEjectMedia(&vcd.org, vcd.config.VCD.Catalog.Name, itemName)
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
	vmType, vmName := vcd.findFirstVm(vapp)
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

	vm, err := vcd.client.Client.FindVMByHREF(vmType.HREF)

	cores := 2
	cpuCount := 4

	task, err := vm.ChangeCPUCountWithCore(cpuCount, &cores)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
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
	check.Assert(task.Task.Status, Equals, "success")
}


func (vcd *TestVCD) Test_VMToggleHardwareVirtualization(check *C) {
	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	// Default nesting status should be false
	nestingStatus := vmType.NestedHypervisorEnabled
	check.Assert(nestingStatus, Equals, false)

	vm, err := vcd.client.Client.FindVMByHREF(vmType.HREF)

	// PowerOn
	task, err := vm.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")

	// Try to change the setting on powered on VM to fail
	_, err = vm.ToggleHardwareVirtualization(true)
	check.Assert(err, ErrorMatches, ".*hardware virtualization can be changed from powered off state.*")

	// PowerOf
	task, err = vm.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")

	// Perform steps on powered off VM
	task, err = vm.ToggleHardwareVirtualization(true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")

	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(vm.VM.NestedHypervisorEnabled, Equals, true)

	task, err = vm.ToggleHardwareVirtualization(false)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")

	err = vm.Refresh()
	check.Assert(err, IsNil)
	check.Assert(vm.VM.NestedHypervisorEnabled, Equals, false)
}

func (vcd *TestVCD) Test_VMPowerOnPowerOff(check *C) {
	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.client.Client.FindVMByHREF(vmType.HREF)
	check.Assert(err, IsNil)

	// Ensure VM is not powered on
	vmStatus, err := vm.GetStatus()
	if vmStatus != "POWERED_OFF" {
		task, err := vm.PowerOff()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(task.Task.Status, Equals, "success")
	}

	task, err := vm.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
	err = vm.Refresh()
	check.Assert(err, IsNil)
	vmStatus, err = vm.GetStatus()
	check.Assert(vmStatus, Equals, "POWERED_ON")

	// Power off again
	task, err = vm.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
	err = vm.Refresh()
	check.Assert(err, IsNil)
	vmStatus, err = vm.GetStatus()
	check.Assert(vmStatus, Equals, "POWERED_OFF")
}

// Test changing network configuration
func (vcd *TestVCD) Test_VMChangeNetworkConfig(check *C) {
	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	if len(vcd.config.VCD.Networks) >= 1 {
		vapp := vcd.findFirstVapp()
		vmType, vmName := vcd.findFirstVm(vapp)
		if vmName == "" {
			check.Skip("skipping test because no VM is found")
		}

		fmt.Printf("Running: %s part 1 change ip\n", check.TestName())

		vm := NewVM(&vcd.client.Client)
		vm.VM = &vmType

		var nets []map[string]interface{}
		n := make(map[string]interface{})

		// Find network
		net, err := vcd.vdc.FindVDCNetwork(vcd.config.VCD.Networks[0])

		check.Assert(err, IsNil)
		check.Assert(net, NotNil)
		check.Assert(net.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Networks[0])
		check.Assert(net.OrgVDCNetwork.HREF, Not(Equals), "")

		endAddress := net.OrgVDCNetwork.Configuration.IPScopes.IPScope.IPRanges.IPRange[0].EndAddress
		startAddress := net.OrgVDCNetwork.Configuration.IPScopes.IPScope.IPRanges.IPRange[0].StartAddress

		n["ip"] = endAddress
		n["ip_allocation_mode"] = "MANUAL"
		n["is_primary"] = true
		n["orgnetwork"] = net.OrgVDCNetwork.Name

		nets = append(nets, n)

		// Change ipaddress to endAddress
		task, err := vm.ChangeNetworkConfig(nets)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		check.Assert(task.Task.Status, Equals, "success")

		// Check if ipaddress is endAddress
		networkConnectionSection, err := vm.GetNetworkConnectionSection()
		check.Assert(err, IsNil)
		check.Assert(networkConnectionSection.NetworkConnection[0].Network, Equals, vcd.config.VCD.Networks[0])
		check.Assert(networkConnectionSection.NetworkConnection[0].IPAddressAllocationMode, Equals, n["ip_allocation_mode"])
		check.Assert(networkConnectionSection.NetworkConnection[0].IPAddress, Equals, endAddress)

		n["ip"] = startAddress

		// Change ipaddress to startAddress
		task, err = vm.ChangeNetworkConfig(nets)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		check.Assert(task.Task.Status, Equals, "success")

		// Check if ipaddress is startAddress
		networkConnectionSection, err = vm.GetNetworkConnectionSection()
		check.Assert(err, IsNil)
		check.Assert(networkConnectionSection.NetworkConnection[0].Network, Equals, vcd.config.VCD.Networks[0])
		check.Assert(networkConnectionSection.NetworkConnection[0].IPAddress, Equals, startAddress)
	}

	if len(vcd.config.VCD.Networks) >= 2 {
		vapp := vcd.findFirstVapp()
		vmType, vmName := vcd.findFirstVm(vapp)
		if vmName == "" {
			check.Skip("skipping test because no VM is found")
		}

		fmt.Printf("Running: %s part 2 change network adapter count\n", check.TestName())

		vm := NewVM(&vcd.client.Client)
		vm.VM = &vmType

		var net OrgVDCNetwork
		var nets []map[string]interface{}
		var err error

		// Find networks
		for index := range vcd.config.VCD.Networks {
			n := make(map[string]interface{})
			net, err = vcd.vdc.FindVDCNetwork(vcd.config.VCD.Networks[index])

			check.Assert(err, IsNil)
			check.Assert(net, NotNil)
			check.Assert(net.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Networks[index])
			check.Assert(net.OrgVDCNetwork.HREF, Not(Equals), "")

			n["ip"] = ""
			n["ip_allocation_mode"] = "POOL"
			n["is_primary"] = false
			if index == 0 {
				n["is_primary"] = true
			}
			n["orgnetwork"] = net.OrgVDCNetwork.Name

			// Add missing networks before continue changing them
			task, err := vapp.AppendNetworkConfig(net.OrgVDCNetwork)
			check.Assert(err, IsNil)
			if task != (Task{}) {
				err = task.WaitTaskCompletion()
				check.Assert(err, IsNil)
				task, err = vm.AppendNetworkConnection(net.OrgVDCNetwork)
				err = task.WaitTaskCompletion()
				check.Assert(err, IsNil)
			}

			nets = append(nets, n)
		}

		// Add networks we found
		task, err := vm.ChangeNetworkConfig(nets)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		check.Assert(task.Task.Status, Equals, "success")

		// Check if networks were added correctly
		networkConnectionSection, err := vm.GetNetworkConnectionSection()
		check.Assert(err, IsNil)
		for index := range vcd.config.VCD.Networks {
			check.Assert(networkConnectionSection.NetworkConnection[index].Network, Equals, vcd.config.VCD.Networks[index])
			check.Assert(networkConnectionSection.NetworkConnection[index].IPAddressAllocationMode, Equals, nets[index]["ip_allocation_mode"])
			check.Assert(networkConnectionSection.NetworkConnection[index].IPAddress, NotNil)
		}
		check.Assert(networkConnectionSection.PrimaryNetworkConnectionIndex, Equals, 0)
	}
}

// Test_updateNicParameters_multinic is meant to check functionality of a complicated
// code structure used in vm.ChangeNetworkConfig which is abstracted into
// vm.updateNicParameters() method so that it does not contain any API calls, but
// only adjust the object which is meant to be sent to API. Initially we hit a bug
// which occurred only when API returned NICs in random order.
func (vcd *TestVCD) Test_updateNicParameters_multiNIC(check *C) {

	// Mock VM struct
	c := Client{}
	vm := NewVM(&c)

	// Sample config which is rendered by .tf schema parsed
	tfCfg := []map[string]interface{}{
		map[string]interface{}{
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "POOL",
			"ip":                 "",
			"is_primary":         false,
		},
		map[string]interface{}{
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "DHCP",
			"ip":                 "",
			"is_primary":         true,
		},
		map[string]interface{}{
			"orgnetwork":         "multinic-net2",
			"ip_allocation_mode": "NONE",
			"ip":                 "",
			"is_primary":         false,
		},
	}

	// A sample NetworkConnectionSection object simulating API returning ordered list
	vcdConfig := types.NetworkConnectionSection{
		PrimaryNetworkConnectionIndex: 0,
		NetworkConnection: []*types.NetworkConnection{
			&types.NetworkConnection{
				Network:                 "multinic-net",
				NetworkConnectionIndex:  0,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:00",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
			&types.NetworkConnection{
				Network:                 "multinic-net",
				NetworkConnectionIndex:  1,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:01",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
			&types.NetworkConnection{
				Network:                 "multinic-net",
				NetworkConnectionIndex:  2,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:02",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
		},
	}

	// NIC configuration when API returns an ordered list
	vcdCfg := &vcdConfig
	vm.updateNicParameters(tfCfg, vcdCfg)

	// Test NIC updates when API returns an unordered list
	// Swap two &types.NetworkConnection so that it is not ordered correctly
	vcdConfig2 := vcdConfig
	vcdConfig2.NetworkConnection[2], vcdConfig2.NetworkConnection[0] = vcdConfig2.NetworkConnection[0], vcdConfig2.NetworkConnection[2]
	vcdCfg2 := &vcdConfig2
	vm.updateNicParameters(tfCfg, vcdCfg2)

	var tableTests = []struct {
		tfConfig []map[string]interface{}
		vcdConfig *types.NetworkConnectionSection
	}{
		{tfCfg, vcdCfg},	// Ordered NIC list
		{tfCfg, vcdCfg2},	// Unordered NIC list
	}

	for _, tableTest := range tableTests {

		// Check that primary interface is reset to 1 as hardcoded in tfCfg "is_primary" parameter
		check.Assert(vcdCfg.PrimaryNetworkConnectionIndex, Equals, 1)

		for loopIndex := range tableTest.vcdConfig.NetworkConnection {
			nicSlot := tableTest.vcdConfig.NetworkConnection[loopIndex].NetworkConnectionIndex

			check.Assert(tableTest.vcdConfig.NetworkConnection[loopIndex].IPAddressAllocationMode, Equals, tableTest.tfConfig[nicSlot]["ip_allocation_mode"].(string))
			check.Assert(tableTest.vcdConfig.NetworkConnection[loopIndex].IsConnected, Equals, true)
			check.Assert(tableTest.vcdConfig.NetworkConnection[loopIndex].NeedsCustomization, Equals, true)
			check.Assert(tableTest.vcdConfig.NetworkConnection[loopIndex].Network, Equals, tableTest.tfConfig[nicSlot]["orgnetwork"].(string))
		}
	}
}

// Test_updateNicParameters_singleNIC is meant to check functionality when single NIC
// is being configured and meant to check functionality so that the function is able
// to cover legacy scenarios when Terraform provider was able to create single IP only.
func (vcd *TestVCD) Test_updateNicParameters_singleNIC(check *C) {
	// Mock VM struct
	c := Client{}
	vm := NewVM(&c)

	tfCfgDHCP := []map[string]interface{}{
		map[string]interface{}{
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "",
			"ip":                 "dhcp",
		},
	}

	tfCfgAllocated := []map[string]interface{}{
		map[string]interface{}{
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "",
			"ip":                 "allocated",
		},
	}

	tfCfgNone := []map[string]interface{}{
		map[string]interface{}{
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "",
			"ip":                 "none",
		},
	}

	tfCfgManual := []map[string]interface{}{
		map[string]interface{}{
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "",
			"ip":                 "1.1.1.1",
		},
	}

	vcdConfig := types.NetworkConnectionSection{
		PrimaryNetworkConnectionIndex: 1,
		NetworkConnection: []*types.NetworkConnection{
			&types.NetworkConnection{
				Network:                 "singlenic-net",
				NetworkConnectionIndex:  0,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:00",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
		},
	}
	
	var tableTests = []struct {
		tfConfig []map[string]interface{}
		//vcdConfig types.NetworkConnectionSection
		expectedIPAddressAllocationMode string
		expectedIPAddress string
	}{
		{tfCfgDHCP,  types.IPAllocationModeDHCP, "Any"},
		{tfCfgAllocated,  types.IPAllocationModePool, "Any"},
		{tfCfgNone,  types.IPAllocationModeNone, "Any"},
		{tfCfgManual,  types.IPAllocationModeManual, tfCfgManual[0]["ip"].(string)},
	}

	for _, tableTest := range tableTests {
		vcdCfg := &vcdConfig
		vm.updateNicParameters(tableTest.tfConfig, vcdCfg)	// Execute parsing procedure

		// Check that primary interface is reset to 0 as it is the only one
		check.Assert(vcdCfg.PrimaryNetworkConnectionIndex, Equals, 0)

		check.Assert(vcdCfg.NetworkConnection[0].IPAddressAllocationMode, Equals, tableTest.expectedIPAddressAllocationMode)
		check.Assert(vcdCfg.NetworkConnection[0].IPAddress, Equals, tableTest.expectedIPAddress)
	}
}
