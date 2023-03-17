//go:build extnetwork || network || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"sort"

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

	// Sort returned IP scopes by gateway because API is not guaranteed to return it in the same
	// order
	sort.SliceStable(ipScope, func(i, j int) bool {
		return ipScope[i].Gateway < ipScope[j].Gateway
	})

	check.Assert(len(ipScope), Equals, 2)
	// Check IPScope 1
	check.Assert(ipScope[0].Gateway, Equals, "192.168.201.1")
	check.Assert(ipScope[0].Netmask, Equals, "255.255.255.0")
	check.Assert(ipScope[0].DNS1, Equals, "192.168.202.253")
	check.Assert(ipScope[0].DNS2, Equals, "192.168.202.254")
	check.Assert(ipScope[0].DNSSuffix, Equals, dnsSuffix)
	// Check IPScope 2
	check.Assert(ipScope[1].Gateway, Equals, "192.168.231.1")
	check.Assert(ipScope[1].Netmask, Equals, "255.255.255.0")
	check.Assert(ipScope[1].DNS1, Equals, "192.168.232.253")
	check.Assert(ipScope[1].DNS2, Equals, "192.168.232.254")
	check.Assert(ipScope[1].DNSSuffix, Equals, dnsSuffix)

	// Sort IP ranges and check them on IPScope 0
	sort.SliceStable(ipScope[0].IPRanges.IPRange, func(i, j int) bool {
		return ipScope[0].IPRanges.IPRange[i].StartAddress < ipScope[0].IPRanges.IPRange[j].StartAddress
	})

	check.Assert(len(ipScope[0].IPRanges.IPRange), Equals, 2)
	ipRange1 := ipScope[0].IPRanges.IPRange[1]
	check.Assert(ipRange1.StartAddress, Equals, "192.168.201.3")
	check.Assert(ipRange1.EndAddress, Equals, "192.168.201.100")

	ipRange2 := ipScope[0].IPRanges.IPRange[0]
	check.Assert(ipRange2.StartAddress, Equals, "192.168.201.105")
	check.Assert(ipRange2.EndAddress, Equals, "192.168.201.140")

	// Sort IP ranges and check them on IPScope 1
	sort.SliceStable(ipScope[1].IPRanges.IPRange, func(i, j int) bool {
		return ipScope[1].IPRanges.IPRange[i].StartAddress < ipScope[1].IPRanges.IPRange[j].StartAddress
	})
	ipRange1 = ipScope[1].IPRanges.IPRange[2]
	check.Assert(ipRange1.StartAddress, Equals, "192.168.231.3")
	check.Assert(ipRange1.EndAddress, Equals, "192.168.231.100")

	ipRange2 = ipScope[1].IPRanges.IPRange[0]
	check.Assert(ipRange2.StartAddress, Equals, "192.168.231.105")
	check.Assert(ipRange2.EndAddress, Equals, "192.168.231.140")

	ipRange3 := ipScope[1].IPRanges.IPRange[1]
	check.Assert(ipRange3.StartAddress, Equals, "192.168.231.145")
	check.Assert(ipRange3.EndAddress, Equals, "192.168.231.150")

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
