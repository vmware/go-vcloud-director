/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NetRefresh(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())

	network, err := vcd.vdc.FindVDCNetwork(vcd.config.VCD.Network)

	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network)
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

// Tests the creation of an Org VDC network connected to an Edge Gateway
func (vcd *TestVCD) Test_CreateOrgVdcNetworkEGW(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	networkName := TestCreateOrgVdcNetworkEGW

	err := RemoveOrgVdcNetworkIfExists(vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	edgeGWName := vcd.config.VCD.EdgeGateway
	if edgeGWName == "" {
		check.Skip("Edge Gateway not provided")
	}
	edgeGateway, err := vcd.vdc.FindEdgeGateway(edgeGWName)
	if err != nil {
		check.Skip(fmt.Sprintf("Edge Gateway %s not found", edgeGWName))
	}

	var networkConfig = types.OrgVDCNetwork{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Name:  networkName,
		// Description: "Created by govcd tests",
		Configuration: &types.NetworkConfiguration{
			FenceMode: "natRouted",
			IPScopes: &types.IPScopes{
				IPScope: types.IPScope{
					IsInherited: false,
					Gateway:     "10.10.102.1",
					Netmask:     "255.255.255.0",
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: "10.10.102.2",
								EndAddress:   "10.10.102.100"},
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

	LogNetwork(networkConfig)
	err = vcd.vdc.CreateOrgVDCNetworkWait(&networkConfig)
	if err != nil {
		fmt.Printf("error creating Network <%s>: %s\n", networkName, err)
	}
	check.Assert(err, IsNil)
	AddToCleanupList(TestCreateOrgVdcNetworkEGW,
		"network",
		vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name,
		"Test_CreateOrgVdcNetworkEGW")
}

// Tests the creation of an isolated Org VDC network
func (vcd *TestVCD) Test_CreateOrgVdcNetworkIso(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	networkName := TestCreateOrgVdcNetworkIso

	err := RemoveOrgVdcNetworkIfExists(vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	var networkConfig = types.OrgVDCNetwork{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Name:  networkName,
		// Description: "Created by govcd tests",
		Configuration: &types.NetworkConfiguration{
			FenceMode: "isolated",
			/*One of:
				bridged (connected directly to the ParentNetwork),
			  isolated (not connected to any other network),
			  natRouted (connected to the ParentNetwork via a NAT service)
			  https://code.vmware.com/apis/287/vcloud#/doc/doc/types/OrgVdcNetworkType.html
			*/
			IPScopes: &types.IPScopes{
				IPScope: types.IPScope{
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
			},
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
	err := RemoveOrgVdcNetworkIfExists(vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("[Test_CreateOrgVdcNetworkDirect] external network not provided")
	}
	externalNetwork, err := GetExternalNetworkByName(vcd.client, vcd.config.VCD.ExternalNetwork)
	if err != nil {
		check.Skip("[Test_CreateOrgVdcNetworkDirect] parent network not found")
	}
	// Note that there is no IPScope for this type of network
	var networkConfig = types.OrgVDCNetwork{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Name:  networkName,
		Configuration: &types.NetworkConfiguration{
			FenceMode: "bridged",
			ParentNetwork: &types.Reference{
				HREF: externalNetwork.HREF,
				Name: externalNetwork.Name,
				Type: externalNetwork.Type,
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
		fmt.Printf("error performing task: %#v", err)
	}
	check.Assert(err, IsNil)
}
