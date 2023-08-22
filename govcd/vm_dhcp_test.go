//go:build nsxv || vm || functional || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_VMGetDhcpAddress proves that it is possible to wait until DHCP lease is acquired by VM and
// report the IP even if VM does not have guest tools installed.
// The test does below actions:
// 1. Creates VM
// 2. Ensures there is a DHCP configuration for network
// 3. Powers off VM
// 4. Sets VM network adapter to use DHCP
// 5. Powers on VM and checks for a DHCP lease assigned to VM
// 6. Cleans up
func (vcd *TestVCD) Test_VMGetDhcpAddress(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}

	// Construct new VM for test
	vapp, err := deployVappForTest(vcd, "GetDhcpAddress")
	check.Assert(err, IsNil)
	vmType, _ := vcd.findFirstVm(*vapp)
	vm := &VM{
		VM:     &vmType,
		client: vapp.client,
	}

	edgeGateway, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	if err != nil {
		check.Skip(fmt.Sprintf("Edge Gateway %s not found", vcd.config.VCD.EdgeGateway))
	}

	// Setup Org network with a single IP in DHCP pool
	network := makeOrgVdcNetworkWithDhcp(vcd, check, edgeGateway)

	// Attach Org network to vApp
	_, err = vapp.AddOrgNetwork(&VappNetworkSettings{}, network.OrgVDCNetwork, false)
	check.Assert(err, IsNil)

	// Get network config and update it to use DHCP
	netCfg, err := vm.GetNetworkConnectionSection()
	check.Assert(err, IsNil)
	check.Assert(netCfg, NotNil)

	netCfg.NetworkConnection[0].Network = network.OrgVDCNetwork.Name
	netCfg.NetworkConnection[0].IPAddressAllocationMode = types.IPAllocationModeDHCP
	netCfg.NetworkConnection[0].IsConnected = true

	secondNic := &types.NetworkConnection{
		Network:                 network.OrgVDCNetwork.Name,
		IPAddressAllocationMode: types.IPAllocationModeDHCP,
		NetworkConnectionIndex:  1,
		IsConnected:             true,
	}
	netCfg.NetworkConnection = append(netCfg.NetworkConnection, secondNic)

	// Update network configuration to use DHCP
	err = vm.UpdateNetworkConnectionSection(netCfg)
	check.Assert(err, IsNil)

	if testVerbose {
		fmt.Printf("# Time out waiting for DHCP IPs on powered off VMs: ")
	}
	// Pretend we are waiting for DHCP addresses when VM is powered off - it must timeout
	ips, hasTimedOut, err := vm.WaitForDhcpIpByNicIndexes([]int{0, 1}, 10, true)
	check.Assert(err, IsNil)
	check.Assert(hasTimedOut, Equals, true)
	check.Assert(ips, HasLen, 2)
	check.Assert(ips[0], Equals, "")
	check.Assert(ips[1], Equals, "")

	if testVerbose {
		fmt.Println("OK")
	}

	// err = vm.PowerOnAndForceCustomization()
	task, err := vapp.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	if testVerbose {
		fmt.Printf("# Get IPs for NICs 0 and 1: ")
	}
	// Wait and check DHCP lease acquired
	ips, hasTimedOut, err = vm.WaitForDhcpIpByNicIndexes([]int{0, 1}, 300, true)
	check.Assert(err, IsNil)
	check.Assert(hasTimedOut, Equals, false)
	check.Assert(ips, HasLen, 2)
	check.Assert(ips[0], Matches, `^32.32.32.\d{1,3}$`)
	check.Assert(ips[1], Matches, `^32.32.32.\d{1,3}$`)

	if testVerbose {
		fmt.Printf("OK:(NICs 0 and 1): %s, %s\n", ips[0], ips[1])
	}

	// DHCP lease was received so VMs MAC address should have an active lease
	if testVerbose {
		fmt.Printf("# Get active lease for NICs with MAC 0: ")
	}
	lease, err := edgeGateway.GetNsxvActiveDhcpLeaseByMac(netCfg.NetworkConnection[0].MACAddress)
	check.Assert(err, IsNil)
	check.Assert(lease, NotNil)
	// This check fails for a known bug in vCD
	//check.Assert(lease.IpAddress, Matches, `^32.32.32.\d{1,3}$`)
	if testVerbose {
		fmt.Printf("Ok. (Got active lease for MAC 0: %s)\n", lease.IpAddress)
	}

	if testVerbose {
		fmt.Printf("# Check number of leases on Edge Gateway: ")
	}
	allLeases, err := edgeGateway.GetAllNsxvDhcpLeases()
	check.Assert(err, IsNil)
	check.Assert(allLeases, NotNil)
	check.Assert(len(allLeases) > 0, Equals, true)
	if testVerbose {
		fmt.Printf("OK: (%d leases found)\n", len(allLeases))
	}

	// Check for a single NIC
	if testVerbose {
		fmt.Printf("# Get IP for single NIC 0: ")
	}
	ips, hasTimedOut, err = vm.WaitForDhcpIpByNicIndexes([]int{0}, 300, true)
	check.Assert(err, IsNil)
	check.Assert(hasTimedOut, Equals, false)
	check.Assert(ips, HasLen, 1)

	// This check fails for a known bug in vCD
	// TODO: re-enable when the bug is fixed
	//check.Assert(ips[0], Matches, `^32.32.32.\d{1,3}$`)
	if testVerbose {
		fmt.Printf("OK: Got IP for NICs 0: %s\n", ips[0])
	}

	// Check if IPs are reported by only using VMware tools
	if testVerbose {
		fmt.Printf("# Get IPs for NICs 0 and 1 (only using guest tools): ")
	}
	ips, hasTimedOut, err = vm.WaitForDhcpIpByNicIndexes([]int{0, 1}, 300, false)
	check.Assert(err, IsNil)
	check.Assert(hasTimedOut, Equals, false)
	check.Assert(ips, HasLen, 2)
	// This check fails for a known bug in vCD
	//check.Assert(ips[0], Matches, `^32.32.32.\d{1,3}$`)
	//check.Assert(ips[1], Matches, `^32.32.32.\d{1,3}$`)
	if testVerbose {
		fmt.Printf("OK: IPs for NICs 0 and 1 (via guest tools): %s, %s\n", ips[0], ips[1])
	}

	// Cleanup vApp
	err = deleteVapp(vcd, vapp.VApp.Name)
	check.Assert(err, IsNil)
	task, err = network.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// makeOrgVdcNetworkWithDhcp is a helper that creates a routed Org network and a DHCP pool with
// single IP address to be assigned. Org vDC network and IP address assigned to DHCP pool are
// returned
func makeOrgVdcNetworkWithDhcp(vcd *TestVCD, check *C, edgeGateway *EdgeGateway) *OrgVDCNetwork {
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
	AddToCleanupList(TestCreateOrgVdcNetworkDhcp, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "TestCreateOrgVdcNetworkDhcp")
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

	return network
}
