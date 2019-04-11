/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"regexp"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Tests the helper function getParentVDC with the vapp
// created at the start of testing
func (vcd *TestVCD) TestGetParentVDC(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	vapp, err := vcd.vdc.FindVAppByName(vcd.vapp.VApp.Name)
	check.Assert(err, IsNil)

	vdc, err := vapp.getParentVDC()

	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.vdc.Vdc.Name)
}

func (vcd *TestVCD) createTestVapp(name string) (VApp, error) {
	// Populate OrgVDCNetwork
	networks := []*types.OrgVDCNetwork{}
	net, err := vcd.vdc.FindVDCNetwork(vcd.config.VCD.Networks[0])
	if err != nil {
		return VApp{}, fmt.Errorf("error finding network : %v", err)
	}
	networks = append(networks, net.OrgVDCNetwork)
	// Populate Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	if err != nil || cat == (Catalog{}) {
		return VApp{}, fmt.Errorf("error finding catalog : %v", err)
	}
	// Populate Catalog Item
	catitem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.CatalogItem)
	if err != nil {
		return VApp{}, fmt.Errorf("error finding catalog item : %v", err)
	}
	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	if err != nil {
		return VApp{}, fmt.Errorf("error finding vapptemplate : %v", err)
	}
	// Get StorageProfileReference
	storageprofileref, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	if err != nil {
		return VApp{}, fmt.Errorf("error finding storage profile: %v", err)
	}
	// Compose VApp
	task, err := vcd.vdc.ComposeVApp(networks, vapptemplate, storageprofileref, name, "description", true)
	if err != nil {
		return VApp{}, fmt.Errorf("error composing vapp: %v", err)
	}
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(name, "vapp", "", "createTestVapp")
	err = task.WaitTaskCompletion()
	if err != nil {
		return VApp{}, fmt.Errorf("error composing vapp: %v", err)
	}
	// Get VApp
	vapp, err := vcd.vdc.FindVAppByName(name)
	if err != nil {
		return VApp{}, fmt.Errorf("error getting vapp: %v", err)
	}
	return vapp, err
}

// Tests Powering On and Powering Off a VApp. Also tests Deletion
// of a VApp
func (vcd *TestVCD) Test_PowerOn(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_Reboot(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vcd.vapp.Reboot()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

func (vcd *TestVCD) Test_BlockWhileStatus(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	initialVappStatus, err := vcd.vapp.GetStatus()
	check.Assert(err, IsNil)

	// Trigger power on in a few seconds after we already wait for status change
	// to simulate system lag
	powerOnResponse := make(chan struct {
		t   Task
		err error
	}, 1)

	go func() {
		time.Sleep(2 * time.Second)
		task, err := vcd.vapp.PowerOn()
		powerOnResponse <- struct {
			t   Task
			err error
		}{task, err}
	}()

	// This must timeout as the timeout is zero
	errMustTimeout := vcd.vapp.BlockWhileStatus(initialVappStatus, 0)
	check.Assert(errMustTimeout, ErrorMatches, "timed out waiting for vApp to exit state .* after .* seconds")

	// This must wait until vApp changes status from initialVappStatus
	err = vcd.vapp.BlockWhileStatus(initialVappStatus, vcd.vapp.client.MaxRetryTimeout)
	check.Assert(err, IsNil)

	// Collect back status response from PowerOn goroutine
	resp := <-powerOnResponse
	check.Assert(resp.err, IsNil)
	err = resp.t.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(resp.t.Task.Status, Equals, "success")

	// Clean up and leave it down
	task, err := vcd.vapp.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking if the ovf was set properly
func (vcd *TestVCD) Test_SetOvf(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	var test map[string]string
	test = make(map[string]string)
	test["guestinfo.hostname"] = "testhostname"
	task, err := vcd.vapp.SetOvf(test)

	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

func (vcd *TestVCD) Test_AddMetadataOnVapp(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Add metadata
	task, err := vcd.vapp.AddMetadata("key", "value")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Check if metadata was added correctly
	metadata, err := vcd.vapp.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata.MetadataEntry[0].Key, Equals, "key")
	check.Assert(metadata.MetadataEntry[0].TypedValue.Value, Equals, "value")
}

func (vcd *TestVCD) Test_DeleteMetadataOnVapp(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Add metadata
	task, err := vcd.vapp.AddMetadata("key2", "value2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Remove metadata
	task, err = vcd.vapp.DeleteMetadata("key2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	metadata, err := vcd.vapp.GetMetadata()
	check.Assert(err, IsNil)
	for _, k := range metadata.MetadataEntry {
		if k.Key == "key2" {
			check.Errorf("metadata.MetadataEntry should not contain key: %s", k)
		}
	}
}

// TODO: Add a check checking if the customization script ran
func (vcd *TestVCD) Test_RunCustomizationScript(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Run Script on Test Vapp
	task, err := vcd.vapp.RunCustomizationScript("computername", "this is my script")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking if the cpu count did change
func (vcd *TestVCD) Test_ChangeCPUcount(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.ChangeCPUCount(1)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking if the cpu count and cores did change
func (vcd *TestVCD) Test_ChangeCPUCountWithCore(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	currentCpus := 0
	currentCores := 0

	// save current values
	if nil != vcd.vapp.VApp.Children.VM[0] && nil != vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection && nil != vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection.Item {
		for _, item := range vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection.Item {
			if item.ResourceType == 3 {
				currentCpus = item.VirtualQuantity
				currentCores = item.CoresPerSocket
				break
			}
		}
	}

	cores := 2
	cpuCount := 4
	task, err := vcd.vapp.ChangeCPUCountWithCore(cpuCount, &cores)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	foundItem := false
	if nil != vcd.vapp.VApp.Children.VM[0] && nil != vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection && nil != vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection.Item {
		for _, item := range vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection.Item {
			if item.ResourceType == 3 {
				check.Assert(item.CoresPerSocket, Equals, cores)
				check.Assert(item.VirtualQuantity, Equals, cpuCount)
				foundItem = true
				break
			}
		}
		check.Assert(foundItem, Equals, true)
	}

	// return tu previous value
	task, err = vcd.vapp.ChangeCPUCountWithCore(currentCpus, &currentCores)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking if the vapp uses the new memory size
func (vcd *TestVCD) Test_ChangeMemorySize(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.ChangeMemorySize(512)

	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking the if the vapp uses the new storage profile
func (vcd *TestVCD) Test_ChangeStorageProfile(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	if vcd.config.VCD.StorageProfile.SP2 == "" {
		check.Skip("Skipping test because second storage profile not given")
	}
	task, err := vcd.vapp.ChangeStorageProfile(vcd.config.VCD.StorageProfile.SP2)
	errStr := fmt.Sprintf("%v", err)

	re := regexp.MustCompile(`error retrieving storage profile`)
	if re.MatchString(errStr) {
		check.Skip("Skipping test because second storage profile not found")
	}
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// TODO: Add a check checking the vm name
func (vcd *TestVCD) Test_ChangeVMName(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.ChangeVMName("My-vm")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_Reset(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vcd.vapp.Reset()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_Suspend(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vcd.vapp.Suspend()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_Shutdown(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vcd.vapp.Shutdown()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

func (vcd *TestVCD) Test_Deploy(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Deploy
	task, err := vcd.vapp.Deploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Find out if there is a way to check if the vapp is on without
// powering it on.
func (vcd *TestVCD) Test_PowerOff(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	task, err := vcd.vapp.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vcd.vapp.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: EVENTUALLY REMOVE THIS REDEPLOY
func (vcd *TestVCD) Test_Undeploy(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Check if the vapp has been deployed yet
	err := vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	if !vcd.vapp.VApp.Deployed {
		task, err := vcd.vapp.Deploy()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}
	// Undeploy
	task, err := vcd.vapp.Undeploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	// Deploy
	// For some reason it will not work without redeploying
	// TODO: EVENTUALLY REMOVE THIS REDEPLOY
	task, err = vcd.vapp.Deploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_AddAndRemoveIsolatedNetwork(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	// Add Metadata
	networkName := "AddAndRemoveIsolatedNetworkTest"
	const gateway = "192.168.0.1"
	const netmask = "255.255.255.0"
	const dns1 = "8.8.8.8"
	const dns2 = "1.1.1.1"
	const dnsSuffix = "biz.biz"
	const startAddress = "192.168.0.10"
	const endAddress = "192.168.0.20"
	const dhcpStartAddress = "192.168.0.30"
	const dhcpEndAddress = "192.168.0.40"
	const maxLeaseTime = 3500
	const defaultLeaseTime = 2400
	var guestVlanAllowed = true
	task, err := vcd.vapp.AddIsolatedNetwork(&VappNetworkSettings{
		Name:             networkName,
		Gateway:          gateway,
		NetMask:          netmask,
		DNS1:             dns1,
		DNS2:             dns2,
		DNSSuffix:        dnsSuffix,
		StaticIPRanges:   []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		GuestVLANAllowed: &guestVlanAllowed,
		DhcpSettings:     &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
	})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	networkConfig, err := vcd.vapp.GetNetworkConfig()
	check.Assert(err, IsNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range networkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}

	check.Assert(networkFound.Configuration.IPScopes.IPScope.Gateway, Equals, gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope.Netmask, Equals, netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope.DNS1, Equals, dns1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope.DNS2, Equals, dns2)
	check.Assert(networkFound.Configuration.IPScopes.IPScope.DNSSuffix, Equals, dnsSuffix)
	check.Assert(networkFound.Configuration.IPScopes.IPScope.IPRanges.IPRange[0].StartAddress, Equals, startAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope.IPRanges.IPRange[0].EndAddress, Equals, endAddress)

	check.Assert(networkFound.Configuration.Features.DhcpService.IsEnabled, Equals, true)
	check.Assert(networkFound.Configuration.Features.DhcpService.MaxLeaseTime, Equals, maxLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.DefaultLeaseTime, Equals, defaultLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.StartAddress, Equals, dhcpStartAddress)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.EndAddress, Equals, dhcpEndAddress)

	task, err = vcd.vapp.RemoveIsolatedNetwork(networkName)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	networkConfig, err = vcd.vapp.GetNetworkConfig()
	check.Assert(err, IsNil)

	isExist := false
	for _, networkConfig := range networkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			isExist = true
		}
	}
	check.Assert(isExist, Equals, false)
}
