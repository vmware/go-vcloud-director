//go:build vapp || functional || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (vcd *TestVCD) Test_UpdateNetworkFirewallRules(check *C) {
	vapp, networkName, vappNetworkConfig, err := vcd.prepareVappWithVappNetwork(check, "Test_UpdateNetworkFirewallRulesVapp", vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}
	check.Assert(networkFound, Not(Equals), types.VAppNetworkConfiguration{})

	uuid, err := GetUuidFromHref(networkFound.Link.HREF, false)
	check.Assert(err, IsNil)

	result, err := vapp.UpdateNetworkFirewallRules(uuid, []*types.FirewallRule{&types.FirewallRule{Description: "myFirstRule1", IsEnabled: true, Policy: "allow",
		DestinationPortRange: "Any", DestinationIP: "Any", SourcePortRange: "Any", SourceIP: "Any", Protocols: &types.FirewallRuleProtocols{TCP: true}},
		&types.FirewallRule{Description: "myFirstRule2", IsEnabled: false, Policy: "drop", DestinationPortRange: "Any",
			DestinationIP: "Any", SourcePortRange: "Any", SourceIP: "Any", Protocols: &types.FirewallRuleProtocols{Any: true}}}, true, "drop", true)
	check.Assert(err, IsNil)
	check.Assert(result, NotNil)
	check.Assert(len(result.Configuration.Features.FirewallService.FirewallRule), Equals, 2)

	// verify
	check.Assert(result.Configuration.Features.FirewallService.IsEnabled, Equals, true)
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].Description, Equals, "myFirstRule1")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].IsEnabled, Equals, true)
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].Policy, Equals, "allow")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].DestinationPortRange, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].DestinationIP, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].SourcePortRange, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].SourceIP, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].Protocols.TCP, Equals, true)
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].Protocols.TCP, Equals, true)

	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].Description, Equals, "myFirstRule2")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].IsEnabled, Equals, false)
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].Policy, Equals, "drop")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].DestinationPortRange, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].DestinationIP, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].SourcePortRange, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].SourceIP, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].Protocols.Any, Equals, true)

	check.Assert(result.Configuration.Features.FirewallService.DefaultAction, Equals, "drop")
	check.Assert(result.Configuration.Features.FirewallService.LogDefaultAction, Equals, true)

	err = vapp.RemoveAllNetworkFirewallRules(uuid)
	check.Assert(err, IsNil)

	vappNetwork, err := vapp.GetVappNetworkById(uuid, true)
	check.Assert(err, IsNil)
	check.Assert(len(vappNetwork.Configuration.Features.FirewallService.FirewallRule), Equals, 0)

	//cleanup
	task, err := vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) prepareVappWithVappNetwork(check *C, vappName, orgVdcNetworkName string) (*VApp, string, *types.NetworkConfigSection, error) {
	fmt.Printf("Running: %s\n", check.TestName())

	vapp, err := deployVappForTest(vcd, vappName)
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	vappNetworkName := vappName + "_network"
	description := "Created in test"
	var guestVlanAllowed = true
	var retainIpMacEnabled = true

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(orgVdcNetworkName, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	vappNetworkSettings := &VappNetworkSettings{
		Name:               vappNetworkName,
		Gateway:            "192.168.0.1",
		NetMask:            "255.255.255.0",
		DNS1:               "8.8.8.8",
		DNS2:               "1.1.1.1",
		DNSSuffix:          "biz.biz",
		StaticIPRanges:     []*types.IPRange{{StartAddress: "192.168.0.10", EndAddress: "192.168.0.20"}},
		DhcpSettings:       &DhcpSettings{IsEnabled: true, MaxLeaseTime: 3500, DefaultLeaseTime: 2400, IPRange: &types.IPRange{StartAddress: "192.168.0.30", EndAddress: "192.168.0.40"}},
		GuestVLANAllowed:   &guestVlanAllowed,
		Description:        description,
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	vappNetworkConfig, err := vapp.CreateVappNetwork(vappNetworkSettings, orgVdcNetwork.OrgVDCNetwork)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)
	return vapp, vappNetworkName, vappNetworkConfig, err
}

func (vcd *TestVCD) Test_GetVappNetworkByNameOrId(check *C) {
	vapp, networkName, vappNetworkConfig, err := vcd.prepareVappWithVappNetwork(check, "Test_GetVappNetworkByNameOrId", vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}
	check.Assert(networkFound, Not(Equals), types.VAppNetworkConfiguration{})

	uuid, err := GetUuidFromHref(networkFound.Link.HREF, false)
	check.Assert(err, IsNil)

	vappNetwork, err := vapp.GetVappNetworkById(uuid, true)
	check.Assert(err, IsNil)
	check.Assert(vappNetwork, NotNil)

	vappNetwork, err = vapp.GetVappNetworkByName(networkName, false)
	check.Assert(err, IsNil)
	check.Assert(vappNetwork, NotNil)

	//cleanup
	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) Test_UpdateNetworkNatRules(check *C) {
	vapp, networkName, vappNetworkConfig, err := vcd.prepareVappWithVappNetwork(check, "Test_UpdateNetworkNatRules", vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}
	check.Assert(networkFound, Not(Equals), types.VAppNetworkConfiguration{})

	desiredNetConfig := types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 0
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModePool,
			Network:                 "Test_UpdateNetworkNatRules_network",
			NetworkConnectionIndex:  0,
		})

	// Get org and vdc
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Find catalog and catalog item
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	catalogItem, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	vappTemplate, err := catalogItem.GetVAppTemplate()
	check.Assert(err, IsNil)

	vm, err := spawnVM("FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", false)
	check.Assert(err, IsNil)

	vm2, err := spawnVM("SecondNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, "", false)
	check.Assert(err, IsNil)

	uuid, err := GetUuidFromHref(networkFound.Link.HREF, false)
	check.Assert(err, IsNil)

	result, err := vapp.UpdateNetworkNatRules(uuid, []*types.NatRule{&types.NatRule{VMRule: &types.NatVMRule{
		ExternalPort: -1, InternalPort: 22, VMNicID: 0,
		VAppScopedVMID: vm.VM.VAppScopedLocalID, Protocol: "TCP"}},
		&types.NatRule{VMRule: &types.NatVMRule{
			ExternalPort: 80, InternalPort: 22, VMNicID: 0,
			VAppScopedVMID: vm2.VM.VAppScopedLocalID, Protocol: "UDP"}}},
		true, "portForwarding", "allowTraffic")
	check.Assert(err, IsNil)
	check.Assert(result, NotNil)
	check.Assert(len(result.Configuration.Features.NatService.NatRule), Equals, 2)

	// verify
	check.Assert(result.Configuration.Features.NatService.IsEnabled, Equals, true)
	check.Assert(result.Configuration.Features.NatService.Policy, Equals, "allowTraffic")
	check.Assert(result.Configuration.Features.NatService.NatType, Equals, "portForwarding")

	check.Assert(result.Configuration.Features.NatService.NatRule[0].VMRule.InternalPort, Equals, 22)
	check.Assert(result.Configuration.Features.NatService.NatRule[0].VMRule.ExternalPort, Equals, -1)
	check.Assert(result.Configuration.Features.NatService.NatRule[0].VMRule.VAppScopedVMID, Equals, vm.VM.VAppScopedLocalID)
	check.Assert(result.Configuration.Features.NatService.NatRule[0].VMRule.VMNicID, Equals, 0)

	check.Assert(result.Configuration.Features.NatService.NatRule[1].VMRule.InternalPort, Equals, 22)
	check.Assert(result.Configuration.Features.NatService.NatRule[1].VMRule.ExternalPort, Equals, 80)
	check.Assert(result.Configuration.Features.NatService.NatRule[1].VMRule.VAppScopedVMID, Equals, vm2.VM.VAppScopedLocalID)
	check.Assert(result.Configuration.Features.NatService.NatRule[1].VMRule.VMNicID, Equals, 0)

	result, err = vapp.UpdateNetworkNatRules(uuid, []*types.NatRule{&types.NatRule{OneToOneVMRule: &types.NatOneToOneVMRule{
		MappingMode: "automatic", VMNicID: 0,
		VAppScopedVMID: vm.VM.VAppScopedLocalID}},
		&types.NatRule{OneToOneVMRule: &types.NatOneToOneVMRule{
			MappingMode: "manual", VMNicID: 0,
			VAppScopedVMID:    vm2.VM.VAppScopedLocalID,
			ExternalIPAddress: addrOf("192.168.100.1")}}},
		false, "ipTranslation", "allowTrafficIn")
	check.Assert(err, IsNil)
	check.Assert(result, NotNil)
	check.Assert(len(result.Configuration.Features.NatService.NatRule), Equals, 2)

	// verify
	check.Assert(result.Configuration.Features.NatService.IsEnabled, Equals, false)
	check.Assert(result.Configuration.Features.NatService.Policy, Equals, "allowTrafficIn")
	check.Assert(result.Configuration.Features.NatService.NatType, Equals, "ipTranslation")

	check.Assert(result.Configuration.Features.NatService.NatRule[0].OneToOneVMRule.MappingMode, Equals, "automatic")
	check.Assert(result.Configuration.Features.NatService.NatRule[0].OneToOneVMRule.VAppScopedVMID, Equals, vm.VM.VAppScopedLocalID)
	check.Assert(result.Configuration.Features.NatService.NatRule[0].OneToOneVMRule.VMNicID, Equals, 0)

	check.Assert(result.Configuration.Features.NatService.NatRule[1].OneToOneVMRule.MappingMode, Equals, "manual")
	check.Assert(result.Configuration.Features.NatService.NatRule[1].OneToOneVMRule.VAppScopedVMID, Equals, vm2.VM.VAppScopedLocalID)
	check.Assert(result.Configuration.Features.NatService.NatRule[1].OneToOneVMRule.VMNicID, Equals, 0)
	check.Assert(result.Configuration.Features.NatService.NatRule[1].OneToOneVMRule.ExternalIPAddress, NotNil)
	check.Assert(*result.Configuration.Features.NatService.NatRule[1].OneToOneVMRule.ExternalIPAddress, Equals, "192.168.100.1")

	err = vapp.RemoveAllNetworkNatRules(uuid)
	check.Assert(err, IsNil)

	vappNetwork, err := vapp.GetVappNetworkById(uuid, true)
	check.Assert(err, IsNil)
	check.Assert(len(vappNetwork.Configuration.Features.NatService.NatRule), Equals, 0)

	//cleanup
	task, err := vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func createRoutedNetwork(vcd *TestVCD, check *C, networkName string) {
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
					Gateway:     "192.168.100.1",
					Netmask:     "255.255.255.0",
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: "192.168.100.2",
								EndAddress:   "192.168.100.50",
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
	err = vcd.vdc.CreateOrgVDCNetworkWait(&networkConfig)
	check.Assert(err, IsNil)
	AddToCleanupList(networkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, networkName)
}

func (vcd *TestVCD) Test_UpdateNetworkStaticRoutes(check *C) {
	testName := check.TestName()
	createRoutedNetwork(vcd, check, testName)
	vapp, vappNetworkName, vappNetworkConfig, err := vcd.prepareVappWithVappNetwork(check, testName, testName)
	check.Assert(err, IsNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vappNetworkName {
			networkFound = networkConfig
		}
	}
	check.Assert(networkFound, Not(Equals), types.VAppNetworkConfiguration{})

	desiredNetConfig := types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 0
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModePool,
			Network:                 vappNetworkName,
			NetworkConnectionIndex:  0,
		})

	uuid, err := GetUuidFromHref(networkFound.Link.HREF, false)
	check.Assert(err, IsNil)

	result, err := vapp.UpdateNetworkStaticRouting(uuid, []*types.StaticRoute{&types.StaticRoute{Name: "test1",
		Network: "192.168.2.0/24", NextHopIP: "192.168.100.15"}, &types.StaticRoute{Name: "test2",
		Network: "192.168.3.0/24", NextHopIP: "192.168.100.16"}}, true)
	check.Assert(err, IsNil)
	check.Assert(result, NotNil)
	check.Assert(len(result.Configuration.Features.StaticRoutingService.StaticRoute), Equals, 2)

	// verify
	check.Assert(result.Configuration.Features.StaticRoutingService.IsEnabled, Equals, true)

	check.Assert(result.Configuration.Features.StaticRoutingService.StaticRoute[0].Name, Equals, "test1")
	check.Assert(result.Configuration.Features.StaticRoutingService.StaticRoute[0].Network, Equals, "192.168.2.0/24")
	check.Assert(result.Configuration.Features.StaticRoutingService.StaticRoute[0].NextHopIP, Equals, "192.168.100.15")

	check.Assert(result.Configuration.Features.StaticRoutingService.StaticRoute[1].Name, Equals, "test2")
	check.Assert(result.Configuration.Features.StaticRoutingService.StaticRoute[1].Network, Equals, "192.168.3.0/24")
	check.Assert(result.Configuration.Features.StaticRoutingService.StaticRoute[1].NextHopIP, Equals, "192.168.100.16")

	err = vapp.RemoveAllNetworkStaticRoutes(uuid)
	check.Assert(err, IsNil)

	vappNetwork, err := vapp.GetVappNetworkById(uuid, true)
	check.Assert(err, IsNil)
	check.Assert(len(vappNetwork.Configuration.Features.StaticRoutingService.StaticRoute), Equals, 0)

	//cleanup
	task, err := vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	network, err := vcd.vdc.GetOrgVdcNetworkByName(testName, true)
	check.Assert(err, IsNil)
	task, err = network.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}
