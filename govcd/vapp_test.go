// +build vapp functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func init() {
	testingTags["vapp"] = "vapp_test.go"
}

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
		check.Skip("Skipping test because vapp was not successfully created at setup")
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
	check.Assert(err, IsNil)
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
		Name:           networkName,
		Gateway:        gateway,
		NetMask:        netmask,
		DNS1:           dns1,
		DNS2:           dns2,
		DNSSuffix:      dnsSuffix,
		StaticIPRanges: []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:   &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
	}

	// vCD 8.20 does not support sending guestVlanAllowed
	if vcd.client.APIVCDMaxVersionIs("> 27.0") {
		vappNetworkSettings.GuestVLANAllowed = &guestVlanAllowed
	} else {
		fmt.Printf("Skipping GuestVLANAllowed parameter as it is not supported on vCD 8.20")
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

	vapp := vcd.findFirstVapp()

	task, err := vapp.AddNewVM(check.TestName(), vapptemplate, nil, true)

	check.Assert(err, IsNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	vdc, err := vapp.getParentVDC()
	check.Assert(err, IsNil)
	vm, err := vdc.FindVMByName(vapp, check.TestName())
	check.Assert(err, IsNil)

	// Cleanup the created VM
	err = vapp.RemoveVM(vm)
	check.Assert(err, IsNil)
}

// Test_AddNewVMMultiNIC creates a new VM in vApp with multiple network cards
func (vcd *TestVCD) Test_AddNewVMMultiNIC(check *C) {
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

	vapp := vcd.findFirstVapp()

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
		vdcNetwork2, err := vcd.vdc.FindVDCNetwork(vcd.config.VCD.Network.Net2)
		check.Assert(err, IsNil)
		orgvdcnetworks := []*types.OrgVDCNetwork{vdcNetwork2.OrgVDCNetwork}
		task, err := vapp.AddRAWNetworkConfig(orgvdcnetworks)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		check.Assert(task.Task.Status, Equals, "success")

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

	task, err := vapp.AddNewVM(check.TestName(), vapptemplate, desiredNetConfig, true)

	check.Assert(err, IsNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	vdc, err := vapp.getParentVDC()
	check.Assert(err, IsNil)

	vm, err := vdc.FindVMByName(vapp, check.TestName())
	check.Assert(err, IsNil)

	// Ensure network config was valid
	actualNetConfig, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)

	verifyNetworkConnectionSection(check, actualNetConfig, desiredNetConfig)

	// Cleanup
	err = vapp.RemoveVM(vm)
	check.Assert(err, IsNil)

	// Ensure network is detached from vApp to avoid conflicts in other tests
	task, err = vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

func verifyNetworkConnectionSection(check *C, actual, desired *types.NetworkConnectionSection) {

	check.Assert(len(actual.NetworkConnection), Equals, len(desired.NetworkConnection))
	check.Assert(actual.PrimaryNetworkConnectionIndex, Equals, desired.PrimaryNetworkConnectionIndex)

	// sort both objects by index before comparison
	sort.SliceStable(actual.NetworkConnection, func(i, j int) bool {
		return actual.NetworkConnection[i].NetworkConnectionIndex <
			actual.NetworkConnection[j].NetworkConnectionIndex
	})

	sort.SliceStable(desired.NetworkConnection, func(i, j int) bool {
		return desired.NetworkConnection[i].NetworkConnectionIndex <
			desired.NetworkConnection[j].NetworkConnectionIndex
	})

	for index := range actual.NetworkConnection {
		actualNic := actual.NetworkConnection[index]
		desiredNic := desired.NetworkConnection[index]

		check.Assert(actualNic.MACAddress, Not(Equals), "")
		check.Assert(actualNic.NetworkAdapterType, Not(Equals), "")
		check.Assert(actualNic.IPAddressAllocationMode, Equals, desiredNic.IPAddressAllocationMode)
		check.Assert(actualNic.Network, Equals, desiredNic.Network)
		check.Assert(actualNic.NetworkConnectionIndex, Equals, desiredNic.NetworkConnectionIndex)

		if actualNic.IPAddressAllocationMode != types.IPAllocationModeNone {
			check.Assert(actualNic.IPAddress, Not(Equals), "")
		}
	}
}

func (vcd *TestVCD) Test_RemoveAllNetworks(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	networkName := "AddAndRemoveNetworkTest"
	networkName2 := "AddAndRemoveNetworkTest2"
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
		Name:           networkName2,
		Gateway:        gateway,
		NetMask:        netmask,
		DNS1:           dns1,
		DNS2:           dns2,
		DNSSuffix:      dnsSuffix,
		StaticIPRanges: []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:   &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
	}

	// vCD 8.20 does not support sending guestVlanAllowed
	if vcd.client.APIVCDMaxVersionIs("> 27.0") {
		vappNetworkSettings.GuestVLANAllowed = &guestVlanAllowed
	} else {
		fmt.Printf("Skipping GuestVLANAllowed parameter as it is not supported on vCD 8.20")
	}

	task, err := vcd.vapp.AddIsolatedNetwork(vappNetworkSettings)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	task, err = vcd.vapp.AddIsolatedNetwork(vappNetworkSettings2)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	err = vcd.vapp.Refresh()
	check.Assert(err, IsNil)
	networkConfig, err := vcd.vapp.GetNetworkConfig()
	check.Assert(err, IsNil)

	check.Assert(len(networkConfig.NetworkConfig), Equals, 2)

	task, err = vcd.vapp.RemoveAllNetworks()
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
