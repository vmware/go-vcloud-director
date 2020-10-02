// +build vapp functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"regexp"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func init() {
	testingTags["vapp"] = "vapp_test.go"
}

// Tests the helper function getParentVDC with the vapp
// created at the start of testing
func (vcd *TestVCD) TestGetParentVDC(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}
	vapp, err := vcd.vdc.GetVAppByName(vcd.vapp.VApp.Name, false)
	check.Assert(err, IsNil)

	vdc, err := vapp.getParentVDC()

	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.vdc.Vdc.Name)
}

// Tests Powering On and Powering Off a VApp. Also tests Deletion
// of a VApp
func (vcd *TestVCD) Test_PowerOn(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
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
		check.Skip("Skipping test because vApp was not successfully created at setup")
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
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}

	initialVappStatus, err := vcd.vapp.GetStatus()
	check.Assert(err, IsNil)

	// This must timeout as the timeout is zero and we are not changing vApp
	errMustTimeout := vcd.vapp.BlockWhileStatus(initialVappStatus, 0)
	check.Assert(errMustTimeout, ErrorMatches, "timed out waiting for vApp to exit state .* after .* seconds")

	task, err := vcd.vapp.PowerOn()
	check.Assert(err, IsNil)
	// This must wait until vApp changes status from initialVappStatus
	err = vcd.vapp.BlockWhileStatus(initialVappStatus, vcd.vapp.client.MaxRetryTimeout)
	check.Assert(err, IsNil)

	// Ensure the powerOn operation succeeded
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Clean up and leave it down
	task, err = vcd.vapp.PowerOff()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking if the ovf was set properly
func (vcd *TestVCD) Test_SetOvf(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}
	var test = make(map[string]string)
	test["guestinfo.hostname"] = "testhostname"
	task, err := vcd.vapp.SetOvf(test)

	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

}

// TODO: Add a check checking if the customization script ran
func (vcd *TestVCD) Test_RunCustomizationScript(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
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
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}
	task, err := vcd.vapp.ChangeCPUCount(1)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// TODO: Add a check checking if the cpu count and cores did change
func (vcd *TestVCD) Test_ChangeCPUCountWithCore(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
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

	cores := 2
	cpuCount := 4
	task, err := vcd.vapp.ChangeCPUCountWithCore(cpuCount, &cores)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	foundItem := false
	if nil != vcd.vapp.VApp.Children.VM[0] && nil != vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection && nil != vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection.Item {
		for _, item := range vcd.vapp.VApp.Children.VM[0].VirtualHardwareSection.Item {
			if item.ResourceType == types.ResourceTypeProcessor {
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
	check.Assert(err, IsNil)
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
	check.Assert(err, IsNil)
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
	errStr := fmt.Sprintf("%s", err)

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
	check.Assert(err, IsNil)
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

	vappNetworkSettings := &VappNetworkSettings{
		Name:             networkName,
		Gateway:          gateway,
		NetMask:          netmask,
		DNS1:             dns1,
		DNS2:             dns2,
		DNSSuffix:        dnsSuffix,
		GuestVLANAllowed: &guestVlanAllowed,
		StaticIPRanges:   []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:     &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
	}

	task, err := vcd.vapp.AddIsolatedNetwork(vappNetworkSettings)
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

	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, dns1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS2, Equals, dns2)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNSSuffix, Equals, dnsSuffix)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, startAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, endAddress)

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

// Test_AddNewVMNilNIC creates VM with nil network configuration
func (vcd *TestVCD) Test_AddNewVMNilNIC(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	// Populate Catalog
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catitem, NotNil)

	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	check.Assert(err, IsNil)

	vapp, err := createVappForTest(vcd, "Test_AddNewVMNilNIC")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)
	task, err := vapp.AddNewVM(check.TestName(), vapptemplate, nil, true)

	check.Assert(err, IsNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	vm, err := vapp.GetVMByName(check.TestName(), true)
	check.Assert(err, IsNil)

	// Cleanup the created VM
	err = vapp.RemoveVM(*vm)
	check.Assert(err, IsNil)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// Test_AddNewVMMultiNIC creates a new VM in vApp with multiple network cards
func (vcd *TestVCD) Test_AddNewVMMultiNIC(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	config := vcd.config
	if config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	// Populate Catalog
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catitem, NotNil)

	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	check.Assert(err, IsNil)

	vapp, err := createVappForTest(vcd, "Test_AddNewVMMultiNIC")
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

	var task Task
	var sp types.Reference
	var customSP = false

	if vcd.config.VCD.StorageProfile.SP1 != "" {
		sp, _ = vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	}

	// TODO: explore the feasibility of adding a test for either case (with or without storage profile).
	if sp.HREF != "" {
		if testVerbose {
			fmt.Printf("Custom storage profile found. Using AddNewVMWithStorage \n")
		}
		customSP = true
		task, err = vapp.AddNewVMWithStorageProfile(check.TestName(), vapptemplate, desiredNetConfig, &sp, true)
	} else {
		if testVerbose {
			fmt.Printf("Custom storage profile not found. Using AddNewVM\n")
		}
		task, err = vapp.AddNewVM(check.TestName(), vapptemplate, desiredNetConfig, true)
	}

	check.Assert(err, IsNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	vm, err := vapp.GetVMByName(check.TestName(), true)
	check.Assert(err, IsNil)

	// Ensure network config was valid
	actualNetConfig, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)

	if customSP {
		check.Assert(vm.VM.StorageProfile.HREF, Equals, sp.HREF)
	}

	verifyNetworkConnectionSection(check, actualNetConfig, desiredNetConfig)

	// Cleanup
	err = vapp.RemoveVM(*vm)
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

func (vcd *TestVCD) Test_RemoveAllNetworks(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	networkName := "Test_RemoveAllNetworks"
	networkName2 := "Test_RemoveAllNetworks2"
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

	vappNetworkSettings := &VappNetworkSettings{
		Name:           networkName,
		Gateway:        gateway,
		NetMask:        netmask,
		DNS1:           dns1,
		DNS2:           dns2,
		DNSSuffix:      dnsSuffix,
		StaticIPRanges: []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:   &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
	}

	vappNetworkSettings2 := &VappNetworkSettings{
		Name:             networkName2,
		Gateway:          gateway,
		NetMask:          netmask,
		DNS1:             dns1,
		DNS2:             dns2,
		DNSSuffix:        dnsSuffix,
		StaticIPRanges:   []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:     &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
		GuestVLANAllowed: &guestVlanAllowed,
	}

	_, err := vcd.vapp.CreateVappNetwork(vappNetworkSettings, nil)
	check.Assert(err, IsNil)

	_, err = vcd.vapp.CreateVappNetwork(vappNetworkSettings2, nil)
	check.Assert(err, IsNil)

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	networkConfig, err := vcd.vapp.GetNetworkConfig()
	check.Assert(err, IsNil)

	check.Assert(len(networkConfig.NetworkConfig), Equals, 2)

	task, err := vcd.vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	networkConfig, err = vcd.vapp.GetNetworkConfig()
	check.Assert(err, IsNil)

	hasNetworks := false
	for _, networkConfig := range networkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName || networkConfig.NetworkName == networkName2 {
			hasNetworks = true
		}

	}
	check.Assert(hasNetworks, Equals, false)
}

// Test_VappSetProductSectionList sets vApp product section, retrieves it and deeply matches if
// properties were properly set using a propertyTester helper.
func (vcd *TestVCD) Test_VappSetProductSectionList(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	vapp := vcd.findFirstVapp()
	propertyTester(vcd, check, &vapp)
}

// Tests VM retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_GetVM(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_GetVapp: Org name not given.")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_GetVapp: VDC name not given.")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	vapp := vcd.findFirstVapp()

	if vapp.VApp == nil || vapp.VApp.HREF == "" || vapp.client == nil {
		check.Skip("no suitable vApp found")
	}
	_, vmName := vcd.findFirstVm(vapp)

	if vmName == "" {
		check.Skip("no suitable VM found")
	}

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return vapp.GetVMByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return vapp.GetVMById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return vapp.GetVMByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "VApp",
		parentName:    vapp.VApp.Name,
		entityType:    "VM",
		entityName:    vmName,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

func (vcd *TestVCD) Test_AddAndRemoveIsolatedVappNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	vapp, err := createVappForTest(vcd, "Test_AddAndRemoveIsolatedVappNetwork")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	// Add Metadata
	networkName := "Test_AddAndRemoveIsolatedVappNetwork"
	description := "Created in test"
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

	vappNetworkSettings := &VappNetworkSettings{
		Name:             networkName,
		Gateway:          gateway,
		NetMask:          netmask,
		DNS1:             dns1,
		DNS2:             dns2,
		DNSSuffix:        dnsSuffix,
		StaticIPRanges:   []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:     &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
		GuestVLANAllowed: &guestVlanAllowed,
		Description:      description,
	}

	vappNetworkConfig, err := vapp.CreateVappNetwork(vappNetworkSettings, nil)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}

	check.Assert(networkFound.Description, Equals, description)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, dns1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS2, Equals, dns2)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNSSuffix, Equals, dnsSuffix)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, startAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, endAddress)

	check.Assert(networkFound.Configuration.Features.DhcpService.IsEnabled, Equals, true)
	check.Assert(networkFound.Configuration.Features.DhcpService.MaxLeaseTime, Equals, maxLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.DefaultLeaseTime, Equals, defaultLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.StartAddress, Equals, dhcpStartAddress)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.EndAddress, Equals, dhcpEndAddress)

	err = vapp.Refresh()
	check.Assert(err, IsNil)
	vappNetworkConfig, err = vapp.RemoveNetwork(networkName)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	isExist := false
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			isExist = true
		}
	}
	check.Assert(isExist, Equals, false)

	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_AddAndRemoveNatVappNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	vapp, err := createVappForTest(vcd, "Test_AddAndRemoveNatVappNetwork")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	// Add Metadata
	networkName := "Test_AddAndRemoveNatVappNetwork"
	description := "Created in test"
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
	var retainIpMacEnabled = true

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	vappNetworkSettings := &VappNetworkSettings{
		Name:               networkName,
		Gateway:            gateway,
		NetMask:            netmask,
		DNS1:               dns1,
		DNS2:               dns2,
		DNSSuffix:          dnsSuffix,
		StaticIPRanges:     []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:       &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
		GuestVLANAllowed:   &guestVlanAllowed,
		Description:        description,
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	vappNetworkConfig, err := vapp.CreateVappNetwork(vappNetworkSettings, orgVdcNetwork.OrgVDCNetwork)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}

	check.Assert(networkFound.Description, Equals, description)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, dns1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS2, Equals, dns2)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNSSuffix, Equals, dnsSuffix)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, startAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, endAddress)

	check.Assert(networkFound.Configuration.Features.DhcpService.IsEnabled, Equals, true)
	check.Assert(networkFound.Configuration.Features.DhcpService.MaxLeaseTime, Equals, maxLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.DefaultLeaseTime, Equals, defaultLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.StartAddress, Equals, dhcpStartAddress)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.EndAddress, Equals, dhcpEndAddress)

	check.Assert(*networkFound.Configuration.RetainNetInfoAcrossDeployments, Equals, retainIpMacEnabled)

	check.Assert(networkFound.Configuration.ParentNetwork.Name, Equals, orgVdcNetwork.OrgVDCNetwork.Name)

	err = vapp.Refresh()
	check.Assert(err, IsNil)
	vappNetworkConfig, err = vapp.RemoveNetwork(networkName)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	isExist := false
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			isExist = true
		}
	}
	check.Assert(isExist, Equals, false)

	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_UpdateVappNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	vapp, err := createVappForTest(vcd, "Test_UpdateVappNetwork")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	// Add
	networkName := "Test_UpdateVappNetwork"
	description := "Created in test"
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
	var retainIpMacEnabled = true

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	vappNetworkSettings := &VappNetworkSettings{
		Name:               networkName,
		Gateway:            gateway,
		NetMask:            netmask,
		DNS1:               dns1,
		DNS2:               dns2,
		DNSSuffix:          dnsSuffix,
		StaticIPRanges:     []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:       &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
		GuestVLANAllowed:   &guestVlanAllowed,
		Description:        description,
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	vappNetworkConfig, err := vapp.CreateVappNetwork(vappNetworkSettings, nil)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}

	check.Assert(networkFound.Description, Equals, description)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, dns1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS2, Equals, dns2)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNSSuffix, Equals, dnsSuffix)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, startAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, endAddress)

	check.Assert(networkFound.Configuration.Features.DhcpService.IsEnabled, Equals, true)
	check.Assert(networkFound.Configuration.Features.DhcpService.MaxLeaseTime, Equals, maxLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.DefaultLeaseTime, Equals, defaultLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.StartAddress, Equals, dhcpStartAddress)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.EndAddress, Equals, dhcpEndAddress)

	var emptyFirewallService *types.FirewallService
	check.Assert(networkFound.Configuration.Features.FirewallService, Equals, emptyFirewallService)
	check.Assert(*networkFound.Configuration.RetainNetInfoAcrossDeployments, Equals, retainIpMacEnabled)

	// Update
	updateNetworkName := "Test_UpdateVappNetworkUpdated"
	updateDescription := "Created in test Updated"
	const updateGateway = "192.168.0.1"
	const updateNetmask = "255.255.255.0"
	const updateDns1 = "8.8.8.7"
	const updateDns2 = "1.1.1.2"
	const updateDnsSuffix = "biz.biz2"
	const updateStartAddress = "192.168.0.11"
	const updateEndAddress = "192.168.0.21"
	const updateDhcpStartAddress = "192.168.0.31"
	const updateDhcpEndAddress = "192.168.0.41"
	const updateMaxLeaseTime = 3400
	const updateDefaultLeaseTime = 2300
	var updateGuestVlanAllowed = false
	var updateRetainIpMacEnabled = false

	uuid, err := GetUuidFromHref(networkFound.Link.HREF, false)
	check.Assert(err, IsNil)
	check.Assert(uuid, NotNil)

	updateVappNetworkSettings := &VappNetworkSettings{
		ID:                 uuid,
		Name:               updateNetworkName,
		Description:        updateDescription,
		Gateway:            updateGateway,
		NetMask:            updateNetmask,
		DNS1:               updateDns1,
		DNS2:               updateDns2,
		DNSSuffix:          updateDnsSuffix,
		StaticIPRanges:     []*types.IPRange{{StartAddress: updateStartAddress, EndAddress: updateEndAddress}},
		DhcpSettings:       &DhcpSettings{IsEnabled: true, MaxLeaseTime: updateMaxLeaseTime, DefaultLeaseTime: updateDefaultLeaseTime, IPRange: &types.IPRange{StartAddress: updateDhcpStartAddress, EndAddress: updateDhcpEndAddress}},
		GuestVLANAllowed:   &updateGuestVlanAllowed,
		RetainIpMacEnabled: &updateRetainIpMacEnabled,
	}

	vappNetworkConfig, err = vapp.UpdateNetwork(updateVappNetworkSettings, orgVdcNetwork.OrgVDCNetwork)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	networkFound = types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == updateNetworkName {
			networkFound = networkConfig
		}
	}

	check.Assert(networkFound.Description, Equals, updateDescription)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, updateGateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, updateNetmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, updateDns1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS2, Equals, updateDns2)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNSSuffix, Equals, updateDnsSuffix)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, updateStartAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, updateEndAddress)

	check.Assert(networkFound.Configuration.Features.DhcpService.IsEnabled, Equals, true)
	check.Assert(networkFound.Configuration.Features.DhcpService.MaxLeaseTime, Equals, updateMaxLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.DefaultLeaseTime, Equals, updateDefaultLeaseTime)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.StartAddress, Equals, updateDhcpStartAddress)
	check.Assert(networkFound.Configuration.Features.DhcpService.IPRange.EndAddress, Equals, updateDhcpEndAddress)

	check.Assert(*networkFound.Configuration.RetainNetInfoAcrossDeployments, Equals, updateRetainIpMacEnabled)

	check.Assert(networkFound.Configuration.ParentNetwork.Name, Equals, orgVdcNetwork.OrgVDCNetwork.Name)

	err = vapp.Refresh()
	check.Assert(err, IsNil)
	vappNetworkConfig, err = vapp.RemoveNetwork(updateNetworkName)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	isExist := false
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			isExist = true
		}
	}
	check.Assert(isExist, Equals, false)

	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_AddAndRemoveVappNetworkWithMinimumValues(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	vapp, err := createVappForTest(vcd, "Test_AddAndRemoveVappNetworkWithMinimumValues")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	// Add Metadata
	networkName := "Test_AddAndRemoveVappNetworkWithMinimumValues"
	const gateway = "192.168.0.1"
	const netmask = "255.255.255.0"

	vappNetworkSettings := &VappNetworkSettings{
		Name:    networkName,
		Gateway: gateway,
		NetMask: netmask,
	}

	vappNetworkConfig, err := vapp.CreateVappNetwork(vappNetworkSettings, nil)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}

	var ipRange *types.IPRanges
	var networkFeatures *types.NetworkFeatures
	var parentNetwork *types.Reference
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, "")
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS2, Equals, "")
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNSSuffix, Equals, "")
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges, Equals, ipRange)

	check.Assert(networkFound.Configuration.Features, Equals, networkFeatures)

	check.Assert(*networkFound.Configuration.RetainNetInfoAcrossDeployments, Equals, false)

	check.Assert(networkFound.Configuration.ParentNetwork, Equals, parentNetwork)

	err = vapp.Refresh()
	check.Assert(err, IsNil)
	vappNetworkConfig, err = vapp.RemoveNetwork(networkName)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	isExist := false
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			isExist = true
		}
	}
	check.Assert(isExist, Equals, false)

	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_AddAndRemoveOrgVappNetworkWithMinimumValues(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	vapp, err := createVappForTest(vcd, "Test_AddAndRemoveOrgVappNetworkWithMinimumValues")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	vappNetworkSettings := &VappNetworkSettings{}

	vappNetworkConfig, err := vapp.AddOrgNetwork(vappNetworkSettings, orgVdcNetwork.OrgVDCNetwork, false)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vcd.config.VCD.Network.Net1 {
			networkFound = networkConfig
		}
	}

	var networkFeatures *types.NetworkFeatures
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress)

	check.Assert(networkFound.Configuration.Features, Equals, networkFeatures)

	check.Assert(*networkFound.Configuration.RetainNetInfoAcrossDeployments, Equals, false)

	check.Assert(networkFound.Configuration.ParentNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	vappNetworkConfig, err = vapp.RemoveNetwork(vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	isExist := false
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vcd.config.VCD.Network.Net1 {
			isExist = true
		}
	}
	check.Assert(isExist, Equals, false)

	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_AddAndRemoveOrgVappNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	vapp, err := createVappForTest(vcd, "Test_AddAndRemoveOrgVappNetwork")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	var retainIpMacEnabled = true

	vappNetworkSettings := &VappNetworkSettings{
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	vappNetworkConfig, err := vapp.AddOrgNetwork(vappNetworkSettings, orgVdcNetwork.OrgVDCNetwork, true)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vcd.config.VCD.Network.Net1 {
			networkFound = networkConfig
		}
	}

	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress)

	check.Assert(*networkFound.Configuration.RetainNetInfoAcrossDeployments, Equals, retainIpMacEnabled)

	check.Assert(networkFound.Configuration.ParentNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	vappNetworkConfig, err = vapp.RemoveNetwork(vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	isExist := false
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vcd.config.VCD.Network.Net1 {
			isExist = true
		}
	}
	check.Assert(isExist, Equals, false)

	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_UpdateOrgVappNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	vapp, err := createVappForTest(vcd, "Test_UpdateOrgVappNetwork")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	var retainIpMacEnabled = true

	vappNetworkSettings := &VappNetworkSettings{
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	vappNetworkConfig, err := vapp.AddOrgNetwork(vappNetworkSettings, orgVdcNetwork.OrgVDCNetwork, true)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vcd.config.VCD.Network.Net1 {
			networkFound = networkConfig
		}
	}

	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress)

	check.Assert(*networkFound.Configuration.RetainNetInfoAcrossDeployments, Equals, retainIpMacEnabled)

	check.Assert(networkFound.Configuration.ParentNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	uuid, err := GetUuidFromHref(networkFound.Link.HREF, false)
	check.Assert(err, IsNil)
	check.Assert(uuid, NotNil)

	var updateRetainIpMacEnabled = false

	vappNetworkSettings = &VappNetworkSettings{
		ID:                 uuid,
		RetainIpMacEnabled: &updateRetainIpMacEnabled,
	}

	vappNetworkConfig, err = vapp.UpdateOrgNetwork(vappNetworkSettings, false)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vcd.config.VCD.Network.Net1 {
			networkFound = networkConfig
		}
	}

	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Gateway, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].Netmask, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].DNS1, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS1)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress)
	check.Assert(networkFound.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, orgVdcNetwork.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress)

	var emptyFirewallFeatures *types.NetworkFeatures
	check.Assert(networkFound.Configuration.Features, Equals, emptyFirewallFeatures)
	check.Assert(networkFound.Configuration.Features, Equals, emptyFirewallFeatures)
	check.Assert(*networkFound.Configuration.RetainNetInfoAcrossDeployments, Equals, updateRetainIpMacEnabled)

	check.Assert(networkFound.Configuration.ParentNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	vappNetworkConfig, err = vapp.RemoveNetwork(vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)

	isExist := false
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vcd.config.VCD.Network.Net1 {
			isExist = true
		}
	}
	check.Assert(isExist, Equals, false)

	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// Test_AddNewVMFromMultiVmTemplate creates VM from OVA holding a few VMs
func (vcd *TestVCD) Test_AddNewVMFromMultiVmTemplate(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	if vcd.config.OVA.OvaMultiVmPath == "" && vcd.config.VCD.Catalog.CatalogItemWithMultiVms == "" {
		check.Skip("skipping test because ovaMultiVmPath or catalogItemWithMultiVms has to be defined")
	}

	if vcd.config.VCD.Catalog.VmNameInMultiVmItem == "" {
		check.Skip("skipping test because vmNameInMultiVmItem is not defined")
	}

	// Populate Catalog
	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := vcd.config.VCD.Catalog.CatalogItemWithMultiVms
	if itemName == "" {
		check.Log("Using `OvaMultiVmPath` for test. Will upload to use it.")
		itemName = check.TestName()
		uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaMultiVmPath, itemName, "upload from test", 1024)
		check.Assert(err, IsNil)
		err = uploadTask.WaitTaskCompletion()
		check.Assert(err, IsNil)

		AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, check.TestName())
	} else {
		check.Log("Using `CatalogItemWithMultiVms` for test")
	}

	vmInTemplateRecord, err := vcd.vdc.QueryVappVmTemplate(vcd.config.VCD.Catalog.Name, itemName, vcd.config.VCD.Catalog.VmNameInMultiVmItem)
	check.Assert(err, IsNil)

	// Get VAppTemplate
	returnedVappTemplate, err := catalog.GetVappTemplateByHref(vmInTemplateRecord.HREF)
	check.Assert(err, IsNil)

	vapp, err := createVappForTest(vcd, "Test_AddNewVMFromMultiVmTemplate")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)
	task, err := vapp.AddNewVM(check.TestName(), *returnedVappTemplate, nil, true)

	check.Assert(err, IsNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	vm, err := vapp.GetVMByName(check.TestName(), true)
	check.Assert(err, IsNil)

	// Cleanup the created VM
	err = vapp.RemoveVM(*vm)
	check.Assert(err, IsNil)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

// Test_AddNewVMWithComputeCapacity creates a new VM in vApp with VM using compute capacity
func (vcd *TestVCD) Test_AddNewVMWithComputeCapacity(check *C) {

	if vcd.client.Client.APIVCDMaxVersionIs("< 33.0") {
		check.Skip(fmt.Sprintf("Test %s requires VCD 10.0 (API version 33) or higher", check.TestName()))
	}

	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	// Populate Catalog
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)

	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catitem, NotNil)

	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	check.Assert(err, IsNil)

	vapp, err := createVappForTest(vcd, "Test_AddNewVMWithComputeCapacity")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	// Create and assign compute policy
	newComputePolicy := &VdcComputePolicy{
		client: vcd.org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:        check.TestName() + "_empty",
			Description: "Empty policy created by test",
		},
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.vdc.Vdc.Name, false)
	if adminVdc == nil || err != nil {
		vcd.infoCleanup(notFoundMsg, "vdc", vcd.vdc.Vdc.Name)
	}

	createdPolicy, err := adminOrg.CreateVdcComputePolicy(newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_AddNewEmptyVMWithVmComputePolicy")

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

	assignedVdcComputePolicies, err := adminVdc.SetAssignedComputePolicies(types.VdcComputePolicyReferences{VdcComputePolicyReference: policyReferences})
	check.Assert(err, IsNil)
	check.Assert(len(allAssignedComputePolicies)+1, Equals, len(assignedVdcComputePolicies.VdcComputePolicyReference))
	// end

	var task Task

	task, err = vapp.AddNewVMWithComputePolicy(check.TestName(), vapptemplate, nil, nil, createdPolicy.VdcComputePolicy, true)

	check.Assert(err, IsNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	createdVm, err := vapp.GetVMByName(check.TestName(), true)
	check.Assert(err, IsNil)

	check.Assert(createdVm.VM.ComputePolicy, NotNil)
	check.Assert(createdVm.VM.ComputePolicy.VmSizingPolicy, NotNil)
	check.Assert(createdVm.VM.ComputePolicy.VmSizingPolicy.ID, Equals, createdPolicy.VdcComputePolicy.ID)

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
