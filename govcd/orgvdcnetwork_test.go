// +build network functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NetRefresh(check *C) {
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	network, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)

	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network.Net1)
	save_network := network

	err = network.Refresh()

	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, save_network.OrgVDCNetwork.Name)
	check.Assert(network.OrgVDCNetwork.HREF, Equals, save_network.OrgVDCNetwork.HREF)
	check.Assert(network.OrgVDCNetwork.Type, Equals, save_network.OrgVDCNetwork.Type)
	check.Assert(network.OrgVDCNetwork.ID, Equals, save_network.OrgVDCNetwork.ID)
	check.Assert(network.OrgVDCNetwork.Description, Equals, save_network.OrgVDCNetwork.Description)
	check.Assert(network.OrgVDCNetwork.EdgeGateway, DeepEquals, save_network.OrgVDCNetwork.EdgeGateway)
	check.Assert(network.OrgVDCNetwork.Status, Equals, save_network.OrgVDCNetwork.Status)

}

func (vcd *TestVCD) Test_CreateOrgVdcNetworkRouted(check *C) {
	vcd.testCreateOrgVdcNetworkRouted(check, "10.10.101", false, false)
}

func (vcd *TestVCD) Test_CreateOrgVdcNetworkRoutedSubInterface(check *C) {
	vcd.testCreateOrgVdcNetworkRouted(check, "10.10.102", true, false)
}

func (vcd *TestVCD) Test_CreateOrgVdcNetworkRoutedDistributed(check *C) {
	vcd.testCreateOrgVdcNetworkRouted(check, "10.10.103", false, true)
}

// Tests the creation of an Org VDC network connected to an Edge Gateway
func (vcd *TestVCD) testCreateOrgVdcNetworkRouted(check *C, ipSubnet string, subInterface, distributed bool) {
	fmt.Printf("Running: %s\n", check.TestName())
	networkName := TestCreateOrgVdcNetworkRouted

	gateway := ipSubnet + ".1"
	startAddress := ipSubnet + ".2"
	endAddress := ipSubnet + ".50"

	if subInterface {
		networkName += "-sub"
	}
	err := RemoveOrgVdcNetworkIfExists(*vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	edgeGWName := vcd.config.VCD.EdgeGateway
	if edgeGWName == "" {
		check.Skip("Edge Gateway not provided")
	}
	edgeGateway, err := vcd.vdc.GetEdgeGatewayByName(edgeGWName, false)
	if err != nil {
		check.Skip(fmt.Sprintf("Edge Gateway %s not found", edgeGWName))
	}

	networkDescription := "Created by govcd tests"
	var networkConfig = types.OrgVDCNetwork{
		Xmlns:       types.XMLNamespaceVCloud,
		Name:        networkName,
		Description: networkDescription,
		Configuration: &types.NetworkConfiguration{
			FenceMode: types.FenceModeNAT,
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					IsInherited: false,
					Gateway:     gateway,
					Netmask:     "255.255.255.0",
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: startAddress,
								EndAddress:   endAddress,
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
	if subInterface && distributed {
		check.Skip("A network can't be at the same time distributed and subInterface")
	}
	if subInterface {
		networkConfig.Configuration.SubInterface = &subInterface
	}

	if distributed {
		distributedRoutingEnabled := edgeGateway.EdgeGateway.Configuration.DistributedRoutingEnabled
		if distributedRoutingEnabled != nil && *distributedRoutingEnabled {
			networkConfig.Configuration.DistributedInterface = &distributed
		} else {
			check.Skip(fmt.Sprintf("edge gateway %s doesn't have distributed routing enabled", edgeGWName))
		}
	}

	LogNetwork(networkConfig)
	err = vcd.vdc.CreateOrgVDCNetworkWait(&networkConfig)
	if err != nil {
		fmt.Printf("error creating Network <%s>: %s\n", networkName, err)
	}
	check.Assert(err, IsNil)
	AddToCleanupList(networkName,
		"network",
		vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name,
		"Test_CreateOrgVdcNetworkRouted")
	network, err := vcd.vdc.GetOrgVdcNetworkByName(networkName, true)
	check.Assert(err, IsNil)
	check.Assert(network, NotNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, networkName)
	check.Assert(network.OrgVDCNetwork.Description, Equals, networkDescription)
	if subInterface {
		check.Assert(network.OrgVDCNetwork.Configuration.SubInterface, NotNil)
		check.Assert(*network.OrgVDCNetwork.Configuration.SubInterface, Equals, true)
	}

	// Tests FindEdgeGatewayNameByNetwork
	// Note: is should work without refreshing either VDC or edge gateway
	connectedGw, err := vcd.vdc.FindEdgeGatewayNameByNetwork(networkName)
	check.Assert(err, IsNil)
	check.Assert(connectedGw, Equals, edgeGWName)

	task, err := network.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// Tests that we can get a network list, and a known network is retrieved from that list
func (vcd *TestVCD) Test_GetNetworkList(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	networkName := vcd.config.VCD.Network.Net1
	if networkName == "" {
		check.Skip("no network name provided")
	}
	networks, err := vcd.vdc.GetNetworkList()
	check.Assert(err, IsNil)
	found := false
	for _, net := range networks {
		// Check that we don't get invalid fields
		knownType := net.LinkType == 0 || net.LinkType == 1 || net.LinkType == 2
		check.Assert(knownType, Equals, true)
		// Check that the `ConnectTo` field is not empty
		check.Assert(net.ConnectedTo, Not(Equals), "")
		if net.Name == networkName {
			found = true
		}
	}
	check.Assert(found, Equals, true)
}

// Tests the creation of an isolated Org VDC network
func (vcd *TestVCD) Test_CreateOrgVdcNetworkIso(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	networkName := TestCreateOrgVdcNetworkIso

	err := RemoveOrgVdcNetworkIfExists(*vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	var networkConfig = types.OrgVDCNetwork{
		Xmlns: types.XMLNamespaceVCloud,
		Name:  networkName,
		// Description: "Created by govcd tests",
		Configuration: &types.NetworkConfiguration{
			FenceMode: types.FenceModeIsolated,
			/*One of:
				bridged (connected directly to the ParentNetwork),
			  isolated (not connected to any other network),
			  natRouted (connected to the ParentNetwork via a NAT service)
			  https://code.vmware.com/apis/287/vcloud#/doc/doc/types/OrgVdcNetworkType.html
			*/
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					IsInherited: false,
					Gateway:     "192.168.2.1",
					Netmask:     "255.255.255.0",
					DNS1:        "192.168.2.1",
					DNS2:        "",
					DNSSuffix:   "",
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: "192.168.2.2",
								EndAddress:   "192.168.2.100"},
						},
					},
				},
				}},
			BackwardCompatibilityMode: true,
		},
		IsShared: false,
	}
	LogNetwork(networkConfig)
	err = vcd.vdc.CreateOrgVDCNetworkWait(&networkConfig)
	if err != nil {
		fmt.Printf("error creating Network <%s>: %s\n", networkName, err)
	}
	check.Assert(err, IsNil)
	AddToCleanupList(TestCreateOrgVdcNetworkIso,
		"network",
		vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name,
		"Test_CreateOrgVdcNetworkIso")
}

// Tests the creation of a Org VDC network connected to an external network
func (vcd *TestVCD) Test_CreateOrgVdcNetworkDirect(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	networkName := TestCreateOrgVdcNetworkDirect

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	err := RemoveOrgVdcNetworkIfExists(*vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("[Test_CreateOrgVdcNetworkDirect] external network not provided")
	}
	externalNetwork, err := vcd.client.GetExternalNetworkByName(vcd.config.VCD.ExternalNetwork)
	if err != nil {
		check.Skip("[Test_CreateOrgVdcNetworkDirect] parent network not found")
		return
	}
	// Note that there is no IPScope for this type of network
	var networkConfig = types.OrgVDCNetwork{
		Xmlns: types.XMLNamespaceVCloud,
		Name:  networkName,
		Configuration: &types.NetworkConfiguration{
			FenceMode: types.FenceModeBridged,
			ParentNetwork: &types.Reference{
				HREF: externalNetwork.ExternalNetwork.HREF,
				Name: externalNetwork.ExternalNetwork.Name,
				Type: externalNetwork.ExternalNetwork.Type,
			},
			BackwardCompatibilityMode: true,
		},
		IsShared: false,
	}
	LogNetwork(networkConfig)

	task, err := vcd.vdc.CreateOrgVDCNetwork(&networkConfig)
	if err != nil {
		fmt.Printf("error creating the network: %s", err)
	}
	check.Assert(err, IsNil)
	if task == (Task{}) {
		fmt.Printf("NULL task retrieved after network creation")
	}
	check.Assert(task.Task.HREF, Not(Equals), "")

	AddToCleanupList(TestCreateOrgVdcNetworkDirect,
		"network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name,
		"Test_CreateOrgVdcNetworkDirect")

	// err = task.WaitTaskCompletion()
	err = task.WaitInspectTaskCompletion(LogTask, 10)
	if err != nil {
		fmt.Printf("error performing task: %s", err)
	}
	check.Assert(err, IsNil)

	// Testing RemoveOrgVdcNetworkIfExists:
	// (1) Make sure the network exists
	newNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(TestCreateOrgVdcNetworkDirect, true)
	check.Assert(err, IsNil)
	check.Assert(newNetwork, NotNil)

	// (2) Removing the network. It should return nil, as a successful deletion
	err = RemoveOrgVdcNetworkIfExists(*vcd.vdc, TestCreateOrgVdcNetworkDirect)
	check.Assert(err, IsNil)

	// (3) Look for the network again. It should be deleted
	_, err = vcd.vdc.GetOrgVdcNetworkByName(TestCreateOrgVdcNetworkDirect, true)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)

	// (4) Attempting a second conditional deletion. It should also return nil, as the network was not found
	err = RemoveOrgVdcNetworkIfExists(*vcd.vdc, TestCreateOrgVdcNetworkDirect)
	check.Assert(err, IsNil)
}
