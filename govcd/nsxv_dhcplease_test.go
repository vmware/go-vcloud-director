// +build nsxv vm functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_VMGetDhcpIp proves that it is possible to wait until DHCP lease is acquired by VM and
// report the IP even if VM does not have guest tools installed.
// The test does below actions:
// 1. Ensures vApp and VM exists or exits early
// 2. Ensures there is a DHCP configuration for network
// 3. Backs up network configuration for VM
// 4. Powers off VM
// 5. Sets VM network adapter to use DHCP
// 6. Powers on VM and checks for a DHCP lease assigned to VM (by MAC address)
// 7. If a DHCP lease is found VM network settings are restored and
func (vcd *TestVCD) Test_VMGetDhcpAddress(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}

	vapp := vcd.findFirstVapp()
	existingVm, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm, err := vcd.client.Client.GetVMByHref(existingVm.HREF)
	check.Assert(err, IsNil)

	edgeGateway, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	if err != nil {
		check.Skip(fmt.Sprintf("Edge Gateway %s not found", vcd.config.VCD.EdgeGateway))
	}

	// Setup Org network with a single IP in DHCP pool
	network, _ := getOrgVdcNetworkWithDhcp(vcd, check, edgeGateway)

	// Attach Org network to vApp
	task, err := vapp.AddRAWNetworkConfig([]*types.OrgVDCNetwork{network})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Backup VM network configuration
	netCfgBackup, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)

	// Set Network to use DHCP
	vmStatus, err := vm.GetStatus()
	check.Assert(err, IsNil)
	if vmStatus != "POWERED_OFF" {
		task, err := vm.PowerOff()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		check.Assert(task.Task.Status, Equals, "success")
	}

	// Get network config and update it to use DHCP
	netCfg, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)

	netCfg.NetworkConnection[0].Network = network.Name
	netCfg.NetworkConnection[0].IPAddressAllocationMode = types.IPAllocationModeDHCP

	// Update network configuration to use DHCP
	err = vm.UpdateNetworkConnectionSection(netCfg)
	check.Assert(err, IsNil)

	// Get network configuration to have network adapter MAC address
	// netConfigWithMac, err := vm.GetNetworkConnectionSection()
	// check.Assert(err, IsNil)

	// nicMacAddress := netConfigWithMac.NetworkConnection[0].MACAddress

	// Power on VM
	task, err = vm.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Wait and check DHCP lease acquired
	// waitForDhcpLease(check, vm, edgeGateway, nicMacAddress, dhcpSubnet)
	ip, err := vm.WaitForDhcpIpByNicIndex(0, 200)
	check.Assert(err, IsNil)
	fmt.Println("Got IP: " + ip)

	// Restore network configuration
	err = vm.UpdateNetworkConnectionSection(netCfgBackup)
	check.Assert(err, IsNil)

	networkAfter, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)

	// Filter out always differing fields and do deep comparison of objects
	netCfgBackup.Link = &types.Link{}
	networkAfter.Link = &types.Link{}
	check.Assert(networkAfter, DeepEquals, netCfgBackup)

}

// getOrgVdcNetworkWithDhcp is a helper that creates a routed Org network and a DHCP pool with
// single IP address to be assigned. Org vDC network and IP address assigned to DHCP pool are
// returned
func getOrgVdcNetworkWithDhcp(vcd *TestVCD, check *C, edgeGateway *EdgeGateway) (*types.OrgVDCNetwork, string) {
	// dhcpPoolIpAddress := "32.32.32.22"
	var networkConfig = types.OrgVDCNetwork{
		Xmlns:       types.XMLNamespaceVCloud,
		Name:        TestCreateOrgVdcNetworkDhcp,
		Description: TestCreateOrgVdcNetworkDhcp,
		Configuration: &types.NetworkConfiguration{
			FenceMode: types.FenceModeNAT,
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					IsInherited: false,
					Gateway:     "32.32.32.1",
					Netmask:     "255.255.255.0",
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: "32.32.32.10",
								EndAddress:   "32.32.32.20",
							},
						},
					},
				},
				},
			},
			BackwardCompatibilityMode: true,
		},
		EdgeGateway: &types.Reference{
			HREF: edgeGateway.EdgeGateway.HREF,
			ID:   edgeGateway.EdgeGateway.ID,
			Name: edgeGateway.EdgeGateway.Name,
			Type: edgeGateway.EdgeGateway.Type,
		},
		IsShared: false,
	}

	// Create network
	err := vcd.vdc.CreateOrgVDCNetworkWait(&networkConfig)
	if err != nil {
		fmt.Printf("error creating Network <%s>: %s\n", TestCreateOrgVdcNetworkDhcp, err)
	}
	check.Assert(err, IsNil)
	AddToCleanupList(TestCreateOrgVdcNetworkDhcp, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name,
		"TestCreateOrgVdcNetworkDhcp")
	network, err := vcd.vdc.GetOrgVdcNetworkByName(TestCreateOrgVdcNetworkDhcp, true)
	check.Assert(err, IsNil)

	// Add DHCP pool
	dhcpPoolConfig := make([]interface{}, 1)
	dhcpPool := make(map[string]interface{})
	dhcpPool["start_address"] = "32.32.32.21"
	dhcpPool["end_address"] = "32.32.32.250"
	dhcpPool["default_lease_time"] = 3600
	dhcpPool["max_lease_time"] = 7200
	dhcpPoolConfig[0] = dhcpPool
	task, err := edgeGateway.AddDhcpPool(network.OrgVDCNetwork, dhcpPoolConfig)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	return network.OrgVDCNetwork, "32.32.32.0/24"
}

// waitForDhcpLease runs a timer for at least 3 minutes and looks for DHCP lease for an active DHCP
// lease for a provided nicMacAddress and validates that the leased IP belongs to dhcpSubnet subnet
func waitForDhcpLease(check *C, vm *VM, edgeGateway *EdgeGateway, nicMacAddress, dhcpSubnet string) {
	// Start operation timer
	startTime := time.Now()
	// Loop and wait for the lease to be acquired
	if testVerbose {
		fmt.Printf("Waiting for VM \"%s\" to get DHCP lease for NIC 0 with MAC %s (\".\" = 5s): ",
			vm.VM.Name, nicMacAddress)
	}

	var dhcpLease *types.EdgeDhcpLeaseInfo

	var stopLoop bool
	var err error
	retryTimeout := vm.client.MaxRetryTimeout
	// due to the VMs taking long time to boot it needs to be at least 3 minutes
	// may be even more in slower environments
	if retryTimeout < 3*60 { // 3 minutes
		retryTimeout = 3 * 60 // 3 minutes
	}
	timeOutAfterInterval := time.Duration(retryTimeout) * time.Second
	timeoutAfter := time.After(timeOutAfterInterval)
	tick := time.NewTicker(time.Duration(5) * time.Second)

	for !stopLoop {
		select {
		case <-timeoutAfter:
			stopLoop = true
			check.Errorf("timed out waiting for VM to acquire DHCP lease")
		case <-tick.C:
			dhcpLease, err = edgeGateway.GetNsxvActiveDhcpLeaseByMac(nicMacAddress)

			if err == nil {
				if testVerbose {
					fmt.Printf(" OK (IP=%s)\n", dhcpLease.IpAddress)
				}
				stopLoop = true
				break
			}

			if !IsNotFound(err) {
				stopLoop = true
				check.Assert(err, IsNil)
			}

			if IsNotFound(err) {
				if testVerbose {
					fmt.Printf(".")
				}
				continue
			}
		}
	}
	if testVerbose {
		fmt.Println("Time taken to acquire DHCP lease after boot: ", time.Since(startTime))
	}
	// Ensure the IP address is in DHCP pool subnet
	_, ipNet, err := net.ParseCIDR(dhcpSubnet)
	check.Assert(err, IsNil)
	check.Assert(ipNet.Contains(net.ParseIP(dhcpLease.IpAddress)), Equals, true)

	// Ensure vCD API itself reports correct IP for
	time.Sleep(1 * time.Minute)
	netWorkCfg, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)
	check.Assert(netWorkCfg.NetworkConnection[0].IPAddress, Equals, dhcpLease.IpAddress)
}
