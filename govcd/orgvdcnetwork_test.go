//go:build network || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
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

func (vcd *TestVCD) Test_CreateUpdateOrgVdcNetworkRouted(check *C) {
	vcd.testCreateUpdateOrgVdcNetworkRouted(check, "10.10.101", false, false)
}

func (vcd *TestVCD) Test_CreateUpdateOrgVdcNetworkRoutedSubInterface(check *C) {
	vcd.testCreateUpdateOrgVdcNetworkRouted(check, "10.10.102", true, false)
}

func (vcd *TestVCD) Test_CreateUpdateOrgVdcNetworkRoutedDistributed(check *C) {
	vcd.testCreateUpdateOrgVdcNetworkRouted(check, "10.10.103", false, true)
}

// Tests the creation and update of an Org VDC network connected to an Edge Gateway
func (vcd *TestVCD) testCreateUpdateOrgVdcNetworkRouted(check *C, ipSubnet string, subInterface, distributed bool) {
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
	AddToCleanupList(networkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_CreateOrgVdcNetworkRouted")
	network, err := vcd.vdc.GetOrgVdcNetworkByName(networkName, true)
	check.Assert(err, IsNil)
	check.Assert(network, NotNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, networkName)
	check.Assert(network.OrgVDCNetwork.Description, Equals, networkDescription)
	check.Assert(network.OrgVDCNetwork.Configuration, NotNil)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes, NotNil)
	check.Assert(len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope), Not(Equals), 0)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway, Equals, gateway)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, startAddress)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, endAddress)

	if subInterface {
		check.Assert(network.OrgVDCNetwork.Configuration.SubInterface, NotNil)
		check.Assert(*network.OrgVDCNetwork.Configuration.SubInterface, Equals, true)
	}

	// Tests FindEdgeGatewayNameByNetwork
	// Note: is should work without refreshing either VDC or edge gateway
	connectedGw, err := vcd.vdc.FindEdgeGatewayNameByNetwork(networkName)
	check.Assert(err, IsNil)
	check.Assert(connectedGw, Equals, edgeGWName)

	networkId := network.OrgVDCNetwork.ID
	updatedNetworkName := networkName + "-updated"
	updatedNetworkDescription := "Updated by govcd tests"
	updatedStartAddress := ipSubnet + ".5"
	updatedEndAddress := ipSubnet + ".30"
	network.OrgVDCNetwork.Name = updatedNetworkName
	network.OrgVDCNetwork.Description = updatedNetworkDescription
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress = updatedStartAddress
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress = updatedEndAddress

	err = network.Update()
	check.Assert(err, IsNil)
	AddToCleanupList(updatedNetworkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_CreateOrgVdcNetworkRouted")

	network, err = vcd.vdc.GetOrgVdcNetworkById(networkId, true)
	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, updatedNetworkName)
	check.Assert(network.OrgVDCNetwork.Description, Equals, updatedNetworkDescription)
	check.Assert(network.OrgVDCNetwork.Configuration, NotNil)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes, NotNil)
	check.Assert(len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope), Not(Equals), 0)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway, Equals, gateway)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, updatedStartAddress)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, updatedEndAddress)

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

// Test_GetNetworkListLarge makes sure we can query a number of networks larger than the default query page length
func (vcd *TestVCD) Test_GetNetworkListLarge(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	defaultPagelength := 25
	externalNetwork, err := vcd.client.GetExternalNetworkByName(vcd.config.VCD.ExternalNetwork)
	if err != nil {
		check.Skip("[Test_GetNetworkListLarge] parent network not found")
		return
	}
	numOfNetworks := defaultPagelength + 1
	baseName := vcd.config.VCD.Org
	for i := 1; i <= numOfNetworks; i++ {
		networkName := fmt.Sprintf("net-%s-d-%d", baseName, i)
		if testVerbose {
			fmt.Printf("creating network %s\n", networkName)
		}
		description := fmt.Sprintf("Created by govcd test - network n. %d", i)
		var networkConfig = types.OrgVDCNetwork{
			Xmlns:       types.XMLNamespaceVCloud,
			Name:        networkName,
			Description: description,
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
		task, err := vcd.vdc.CreateOrgVDCNetwork(&networkConfig)
		check.Assert(err, IsNil)

		AddToCleanupList(networkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_GetNetworkListLarge")
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	err = vcd.vdc.Refresh()
	check.Assert(err, IsNil)

	knownNetworkName1 := fmt.Sprintf("net-%s-d", baseName)
	knownNetworkName2 := fmt.Sprintf("net-%s-d-%d", baseName, numOfNetworks)
	networks, err := vcd.vdc.GetNetworkList()
	check.Assert(err, IsNil)
	if testVerbose {
		fmt.Printf("Number of networks: %d\n", len(networks))
	}
	check.Assert(len(networks) > defaultPagelength, Equals, true)
	found1 := false
	found2 := false
	for _, net := range networks {
		if net.Name == knownNetworkName1 {
			found1 = true
		}
		if net.Name == knownNetworkName2 {
			found2 = true
		}
	}
	check.Assert(found1, Equals, true)
	check.Assert(found2, Equals, true)

	for i := 1; i <= numOfNetworks; i++ {
		networkName := fmt.Sprintf("net-%s-d-%d", baseName, i)
		if testVerbose {
			fmt.Printf("Removing network %s\n", networkName)
		}
		network, err := vcd.vdc.GetOrgVdcNetworkByName(networkName, false)
		check.Assert(err, IsNil)
		_, err = network.Delete()
		check.Assert(err, IsNil)
	}
	err = vcd.vdc.Refresh()
	check.Assert(err, IsNil)
}

// Tests the creation and update of an isolated Org VDC network
func (vcd *TestVCD) Test_CreateUpdateOrgVdcNetworkIso(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	networkName := TestCreateOrgVdcNetworkIso

	err := RemoveOrgVdcNetworkIfExists(*vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	var (
		gateway             = "192.168.2.1"
		updatedNetworkName  = TestCreateOrgVdcNetworkIso + "updated"
		startAddress        = "192.168.2.2"
		updatedStartAddress = "192.168.2.5"
		endAddress          = "192.168.2.100"
		updatedEndAddress   = "192.168.2.50"
		netmask             = "255.255.255.0"
		dns1                = gateway
		dns2                = "8.8.8.8"
		dnsSuffix           = "vcloud.org"
		description         = "Created by govcd tests"
		updatedDescription  = "Updated by govcd tests"
		networkConfig       = types.OrgVDCNetwork{
			Xmlns:       types.XMLNamespaceVCloud,
			Name:        networkName,
			Description: description,
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
						Gateway:     gateway,
						Netmask:     netmask,
						DNS1:        dns1,
						DNS2:        dns2,
						DNSSuffix:   dnsSuffix,
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
			IsShared: false,
		}
	)

	LogNetwork(networkConfig)
	err = vcd.vdc.CreateOrgVDCNetworkWait(&networkConfig)
	if err != nil {
		fmt.Printf("error creating Network <%s>: %s\n", networkName, err)
	}
	check.Assert(err, IsNil)
	AddToCleanupList(networkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_CreateOrgVdcNetworkIso")

	network, err := vcd.vdc.GetOrgVdcNetworkByName(networkName, true)
	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Description, Equals, description)
	check.Assert(network.OrgVDCNetwork.Configuration.FenceMode, Equals, types.FenceModeIsolated)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes, NotNil)
	check.Assert(len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope), Not(Equals), 0)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway, Equals, gateway)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask, Equals, netmask)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS1, Equals, dns1)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS2, Equals, dns2)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNSSuffix, Equals, dnsSuffix)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges, NotNil)
	check.Assert(len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange), Not(Equals), 0)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, startAddress)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, endAddress)

	networkId := network.OrgVDCNetwork.ID
	network.OrgVDCNetwork.Description = updatedDescription
	network.OrgVDCNetwork.Name = updatedNetworkName
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress = updatedStartAddress
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress = updatedEndAddress
	err = network.Update()
	check.Assert(err, IsNil)
	AddToCleanupList(updatedNetworkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_CreateOrgVdcNetworkIso")

	network, err = vcd.vdc.GetOrgVdcNetworkById(networkId, true)
	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, updatedNetworkName)
	check.Assert(network.OrgVDCNetwork.Description, Equals, updatedDescription)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes, NotNil)
	check.Assert(len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope), Not(Equals), 0)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask, Equals, netmask)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS2, Equals, dns2)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNSSuffix, Equals, dnsSuffix)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges, NotNil)
	check.Assert(len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange), Not(Equals), 0)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress, Equals, updatedStartAddress)
	check.Assert(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress, Equals, updatedEndAddress)
	task, err := network.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

}

// Tests the creation and update of a Org VDC network connected to an external network
func (vcd *TestVCD) Test_CreateUpdateOrgVdcNetworkDirect(check *C) {
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
	description := "Created by govcd test"
	var networkConfig = types.OrgVDCNetwork{
		Xmlns:       types.XMLNamespaceVCloud,
		Name:        networkName,
		Description: description,
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

	AddToCleanupList(networkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, check.TestName())

	// err = task.WaitTaskCompletion()
	err = task.WaitInspectTaskCompletion(LogTask, 10)
	if err != nil {
		fmt.Printf("error performing task: %s", err)
	}
	check.Assert(err, IsNil)

	// Retrieving the network
	newNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(networkName, true)
	check.Assert(err, IsNil)
	check.Assert(newNetwork, NotNil)
	check.Assert(newNetwork.OrgVDCNetwork.Name, Equals, networkName)
	check.Assert(newNetwork.OrgVDCNetwork.Description, Equals, description)
	check.Assert(newNetwork.OrgVDCNetwork.Configuration, NotNil)
	check.Assert(newNetwork.OrgVDCNetwork.Configuration.ParentNetwork, NotNil)
	check.Assert(newNetwork.OrgVDCNetwork.Configuration.ParentNetwork.ID, Equals, externalNetwork.ExternalNetwork.ID)
	check.Assert(newNetwork.OrgVDCNetwork.Configuration.ParentNetwork.Name, Equals, externalNetwork.ExternalNetwork.Name)

	// Updating
	updatedNetworkName := networkName + "-updated"
	updatedDescription := "Updated by govcd tests"
	newNetwork.OrgVDCNetwork.Description = updatedDescription
	newNetwork.OrgVDCNetwork.Name = updatedNetworkName
	err = newNetwork.Update()
	check.Assert(err, IsNil)

	AddToCleanupList(updatedNetworkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_CreateOrgVdcNetworkDirect")

	// Check the values of the updated entities. The new values should be available
	// immediately.
	check.Assert(newNetwork.OrgVDCNetwork.Name, Equals, updatedNetworkName)
	check.Assert(newNetwork.OrgVDCNetwork.Description, Equals, updatedDescription)

	// Gets a new copy of the network and check the values.
	newNetwork, err = vcd.vdc.GetOrgVdcNetworkByName(updatedNetworkName, true)
	check.Assert(err, IsNil)
	check.Assert(newNetwork, NotNil)
	check.Assert(newNetwork.OrgVDCNetwork.Name, Equals, updatedNetworkName)
	check.Assert(newNetwork.OrgVDCNetwork.Description, Equals, updatedDescription)

	// restoring original name
	newNetwork.OrgVDCNetwork.Name = networkName
	err = newNetwork.Update()
	check.Assert(err, IsNil)
	check.Assert(newNetwork.OrgVDCNetwork.Name, Equals, networkName)

	// Testing RemoveOrgVdcNetworkIfExists:
	// (1) Make sure the network exists
	newNetwork, err = vcd.vdc.GetOrgVdcNetworkByName(networkName, true)
	check.Assert(err, IsNil)
	check.Assert(newNetwork, NotNil)

	// (2) Removing the network. It should return nil, as a successful deletion
	err = RemoveOrgVdcNetworkIfExists(*vcd.vdc, networkName)
	check.Assert(err, IsNil)

	// (3) Look for the network again. It should be deleted
	_, err = vcd.vdc.GetOrgVdcNetworkByName(networkName, true)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)

	// (4) Attempting a second conditional deletion. It should also return nil, as the network was not found
	err = RemoveOrgVdcNetworkIfExists(*vcd.vdc, networkName)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_NetworkUpdateRename(check *C) {
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	network, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)

	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network.Net1)
	var saveNetwork = types.OrgVDCNetwork{
		HREF:        network.OrgVDCNetwork.HREF,
		Type:        network.OrgVDCNetwork.Type,
		ID:          network.OrgVDCNetwork.ID,
		Name:        network.OrgVDCNetwork.Name,
		Status:      network.OrgVDCNetwork.Status,
		Description: network.OrgVDCNetwork.Description,
	}

	// Change name and description
	updatedNetworkName := "UpdatedNetwork"
	updatedDescription := "Updated description"
	network.OrgVDCNetwork.Description = updatedDescription
	network.OrgVDCNetwork.Name = updatedNetworkName
	err = network.Update()
	check.Assert(err, IsNil)

	// Retrieve the network again
	network, err = vcd.vdc.GetOrgVdcNetworkById(saveNetwork.ID, false)
	check.Assert(err, IsNil)

	check.Assert(network.OrgVDCNetwork.Name, Equals, updatedNetworkName)
	check.Assert(network.OrgVDCNetwork.Description, Equals, updatedDescription)
	check.Assert(network.OrgVDCNetwork.HREF, Equals, saveNetwork.HREF)
	check.Assert(network.OrgVDCNetwork.Type, Equals, saveNetwork.Type)
	check.Assert(network.OrgVDCNetwork.ID, Equals, saveNetwork.ID)
	check.Assert(network.OrgVDCNetwork.EdgeGateway, DeepEquals, saveNetwork.EdgeGateway)
	check.Assert(network.OrgVDCNetwork.Status, Equals, saveNetwork.Status)

	// Change back name and description to their original values
	network.OrgVDCNetwork.Description = saveNetwork.Description
	network.OrgVDCNetwork.Name = saveNetwork.Name

	err = network.Update()
	check.Assert(err, IsNil)
	network, err = vcd.vdc.GetOrgVdcNetworkById(saveNetwork.ID, false)
	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Description, Equals, saveNetwork.Description)
	check.Assert(network.OrgVDCNetwork.Name, Equals, saveNetwork.Name)

	// An update without any changes should run without producing side effects
	err = network.Update()
	check.Assert(err, IsNil)
	network, err = vcd.vdc.GetOrgVdcNetworkById(saveNetwork.ID, false)
	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Description, Equals, saveNetwork.Description)
	check.Assert(network.OrgVDCNetwork.Name, Equals, saveNetwork.Name)

	// We run some of the above operations using Rename instead of Update
	err = network.Rename(updatedNetworkName)
	check.Assert(err, IsNil)
	network, err = vcd.vdc.GetOrgVdcNetworkById(saveNetwork.ID, false)
	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, updatedNetworkName)

	// Trying to rename with the same name should give an error
	err = network.Rename(updatedNetworkName)
	check.Assert(err, NotNil)

	// Setting the original name again
	err = network.Rename(saveNetwork.Name)
	check.Assert(err, IsNil)
	network, err = vcd.vdc.GetOrgVdcNetworkById(saveNetwork.ID, false)
	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, saveNetwork.Name)

}
