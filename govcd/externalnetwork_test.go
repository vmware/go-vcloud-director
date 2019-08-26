// +build extnetwork network functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Retrieves an external network and checks that its contents are filled as expected
func (vcd *TestVCD) Test_ExternalNetworkGetByName(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	externalNetwork, err := vcd.client.GetExternalNetworkByName(vcd.config.VCD.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(externalNetwork, NotNil)

	check.Assert(externalNetwork.ExternalNetwork.Name, Equals, vcd.config.VCD.ExternalNetwork)
}

// Helper function that creates an external network to be used in other tests
func (vcd *TestVCD) testCreateExternalNetwork(testName, networkName, dnsSuffix string) (skippingReason string, externalNetwork *types.ExternalNetwork, task Task, err error) {

	if vcd.skipAdminTests {
		return fmt.Sprintf(TestRequiresSysAdminPrivileges, testName), externalNetwork, Task{}, nil
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		return fmt.Sprintf("%s: External network isn't configured. Test can't proceed", testName), externalNetwork, Task{}, nil
	}

	if vcd.config.VCD.VimServer == "" {
		return fmt.Sprintf("%s: Vim server isn't configured. Test can't proceed", testName), externalNetwork, Task{}, nil
	}

	if vcd.config.VCD.ExternalNetworkPortGroup == "" {
		return fmt.Sprintf("%s: Port group isn't configured. Test can't proceed", testName), externalNetwork, Task{}, nil
	}

	if vcd.config.VCD.ExternalNetworkPortGroupType == "" {
		return fmt.Sprintf("%s: Port group type isn't configured. Test can't proceed", testName), externalNetwork, Task{}, nil
	}

	virtualCenters, err := QueryVirtualCenters(vcd.client, fmt.Sprintf("(name==%s)", vcd.config.VCD.VimServer))
	if err != nil {
		return "", externalNetwork, Task{}, err
	}
	if len(virtualCenters) == 0 {
		return fmt.Sprintf("No vSphere server found with name '%s'", vcd.config.VCD.VimServer), externalNetwork, Task{}, nil
	}
	vimServerHref := virtualCenters[0].HREF

	// Resolve port group info
	portGroups, err := QueryPortGroups(vcd.client, fmt.Sprintf("(name==%s;portgroupType==%s)", url.QueryEscape(vcd.config.VCD.ExternalNetworkPortGroup), vcd.config.VCD.ExternalNetworkPortGroupType))
	if err != nil {
		return "", externalNetwork, Task{}, err
	}
	if len(portGroups) == 0 {
		return fmt.Sprintf("No port group found with name '%s'", vcd.config.VCD.ExternalNetworkPortGroup), externalNetwork, Task{}, nil
	}
	if len(portGroups) > 1 {
		return fmt.Sprintf("More than one port group found with name '%s'", vcd.config.VCD.ExternalNetworkPortGroup), externalNetwork, Task{}, nil
	}

	externalNetwork = &types.ExternalNetwork{
		Name:        networkName,
		Description: "Test Create External Network",
		Configuration: &types.NetworkConfiguration{
			Xmlns: types.XMLNamespaceVCloud,
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					Gateway:   "192.168.201.1",
					Netmask:   "255.255.255.0",
					DNS1:      "192.168.202.253",
					DNS2:      "192.168.202.254",
					DNSSuffix: dnsSuffix,
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: "192.168.201.3",
								EndAddress:   "192.168.201.250",
							},
						},
					},
				},
				}},
			FenceMode: "isolated",
		},
		VimPortGroupRefs: &types.VimObjectRefs{
			VimObjectRef: []*types.VimObjectRef{
				&types.VimObjectRef{
					VimServerRef: &types.Reference{
						HREF: vimServerHref,
					},
					MoRef:         portGroups[0].MoRef,
					VimObjectType: vcd.config.VCD.ExternalNetworkPortGroupType,
				},
			},
		},
	}
	task, err = CreateExternalNetwork(vcd.client, externalNetwork)
	return skippingReason, externalNetwork, task, err
}

// Tests System function Delete by creating external network and
// deleting it after.
func (vcd *TestVCD) Test_ExternalNetworkDelete(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skippingReason, externalNetwork, task, err := vcd.testCreateExternalNetwork(check.TestName(), TestDeleteExternalNetwork, "")
	if skippingReason != "" {
		check.Skip(skippingReason)
	}

	check.Assert(err, IsNil)
	AddToCleanupList(externalNetwork.Name, "externalNetwork", "", "Test_ExternalNetworkDelete")
	check.Assert(task.Task, Not(Equals), types.Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	createdExternalNetwork, err := vcd.client.GetExternalNetworkByName(externalNetwork.Name)
	check.Assert(err, IsNil)
	check.Assert(createdExternalNetwork, NotNil)

	err = createdExternalNetwork.DeleteWait()
	check.Assert(err, IsNil)

	// check through existing external networks
	_, err = vcd.client.GetExternalNetworkByName(externalNetwork.Name)
	check.Assert(err, NotNil)
}

// Retrieves an external network and checks that its contents are filled as expected
func (vcd *TestVCD) Test_GetExternalNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	networkName := vcd.config.VCD.ExternalNetwork
	if networkName == "" {
		check.Skip("No external network provided")
	}
	externalNetwork, err := vcd.client.GetExternalNetworkByName(networkName)
	check.Assert(err, IsNil)
	check.Assert(externalNetwork, NotNil)
	LogExternalNetwork(*externalNetwork.ExternalNetwork)
	check.Assert(externalNetwork.ExternalNetwork.HREF, Not(Equals), "")
	check.Assert(externalNetwork.ExternalNetwork.Name, Equals, networkName)
	check.Assert(externalNetwork.ExternalNetwork.Type, Equals, types.MimeExternalNetwork)
}

func (vcd *TestVCD) Test_CreateExternalNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	dnsSuffix := "some.net"
	skippingReason, externalNetwork, task, err := vcd.testCreateExternalNetwork(check.TestName(), TestCreateExternalNetwork, dnsSuffix)
	if skippingReason != "" {
		check.Skip(skippingReason)
	}

	check.Assert(err, IsNil)
	check.Assert(task.Task, Not(Equals), types.Task{})

	AddToCleanupList(externalNetwork.Name, "externalNetwork", "", "Test_CreateExternalNetwork")
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	newExternalNetwork, err := vcd.client.GetExternalNetworkByName(TestCreateExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(newExternalNetwork, NotNil)
	check.Assert(newExternalNetwork.ExternalNetwork.Name, Equals, TestCreateExternalNetwork)

	ipScope := newExternalNetwork.ExternalNetwork.Configuration.IPScopes.IPScope
	check.Assert(ipScope[0].Gateway, Equals, "192.168.201.1")
	check.Assert(ipScope[0].Netmask, Equals, "255.255.255.0")
	check.Assert(ipScope[0].DNS1, Equals, "192.168.202.253")
	check.Assert(ipScope[0].DNS2, Equals, "192.168.202.254")
	check.Assert(ipScope[0].DNSSuffix, Equals, dnsSuffix)

	check.Assert(len(ipScope[0].IPRanges.IPRange), Equals, 1)
	ipRange := ipScope[0].IPRanges.IPRange[0]
	check.Assert(ipRange.StartAddress, Equals, "192.168.201.3")
	check.Assert(ipRange.EndAddress, Equals, "192.168.201.250")

	check.Assert(newExternalNetwork.ExternalNetwork.Configuration.FenceMode, Equals, "isolated")
	check.Assert(newExternalNetwork.ExternalNetwork.Description, Equals, "Test Create External Network")
	check.Assert(newExternalNetwork.ExternalNetwork.VimPortGroupRef, NotNil)
	check.Assert(newExternalNetwork.ExternalNetwork.VimPortGroupRef.VimObjectType, Equals, externalNetwork.VimPortGroupRefs.VimObjectRef[0].VimObjectType)
	check.Assert(newExternalNetwork.ExternalNetwork.VimPortGroupRef.MoRef, Equals, externalNetwork.VimPortGroupRefs.VimObjectRef[0].MoRef)
	check.Assert(newExternalNetwork.ExternalNetwork.VimPortGroupRef.VimServerRef.HREF, Equals, externalNetwork.VimPortGroupRefs.VimObjectRef[0].VimServerRef.HREF)

	err = newExternalNetwork.DeleteWait()
	check.Assert(err, IsNil)
}

func init() {
	testingTags["extnetwork"] = "externalnetwork_test.go"
}

// Tests ExternalNetwork retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_SystemGetExternalNetwork(check *C) {

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return vcd.client.GetExternalNetworkByName(name)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return vcd.client.GetExternalNetworkById(id) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return vcd.client.GetExternalNetworkByNameOrId(id)
	}

	var def = getterTestDefinition{
		parentType:    "VDCClient",
		parentName:    "System",
		entityType:    "ExternalNetwork",
		entityName:    vcd.config.VCD.ExternalNetwork,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}
