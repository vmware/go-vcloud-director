/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"net/url"
	"time"
)

// Retrieves an external network and checks that its contents are filled as expected
func (vcd *TestVCD) Test_ExternalNetworkGetByName(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	externalNetwork := NewExternalNetwork(&vcd.client.Client)
	err := externalNetwork.GetByName(vcd.config.VCD.ExternalNetwork)
	check.Assert(err, IsNil)

	check.Assert(externalNetwork.ExternalNetwork.Name, Equals, vcd.config.VCD.ExternalNetwork)
}

// Tests System function Delete by creating external network and
// deleting it after.
func (vcd *TestVCD) Test_ExternalNetworkDelete(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("Test_GetByName: External network isn't configured. Test can't proceed")
	}

	if vcd.config.VCD.VimServer == "" {
		check.Skip("Test_GetByName: Vim server isn't configured. Test can't proceed")
	}

	if vcd.config.VCD.ExternalNetworkPortGroup == "" {
		check.Skip("Test_GetByName: Port group isn't configured. Test can't proceed")
	}

	if vcd.config.VCD.ExternalNetworkPortGroupType == "" {
		check.Skip("Test_GetByName: Port group type isn't configured. Test can't proceed")
	}

	virtualCenters, err := QueryVirtualCenters(vcd.client, fmt.Sprintf("(name==%s)", vcd.config.VCD.VimServer))
	check.Assert(err, IsNil)
	if len(virtualCenters) == 0 {
		check.Skip(fmt.Sprintf("No vSphere server found with name '%s'", vcd.config.VCD.VimServer))
	}
	vimServerHref := virtualCenters[0].HREF

	// Resolve port group info
	portGroups, err := QueryPortGroups(vcd.client, fmt.Sprintf("(name==%s;portgroupType==%s)", url.QueryEscape(vcd.config.VCD.ExternalNetworkPortGroup), vcd.config.VCD.ExternalNetworkPortGroupType))
	check.Assert(err, IsNil)
	if len(portGroups) == 0 {
		check.Skip(fmt.Sprintf("No port group found with name '%s'", vcd.config.VCD.ExternalNetworkPortGroup))
	}
	if len(portGroups) > 1 {
		check.Skip(fmt.Sprintf("More than one port group found with name '%s'", vcd.config.VCD.ExternalNetworkPortGroup))
	}

	externalNetwork := &types.ExternalNetwork{
		Name:        TestDeleteExternalNetwork,
		Description: "Test Create External Network",
		Xmlns:       types.XMLNamespaceExtension,
		XmlnsVCloud: types.XMLNamespaceVCloud,
		Configuration: &types.NetworkConfiguration{
			Xmlns: types.XMLNamespaceVCloud,
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					Gateway: "192.168.201.1",
					Netmask: "255.255.255.0",
					DNS1:    "192.168.202.253",
					DNS2:    "192.168.202.254",
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
	task, err := CreateExternalNetwork(vcd.client, externalNetwork)
	check.Assert(err, IsNil)
	check.Assert(task, Not(Equals), Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	createdExternalNetwork, err := GetExternalNetwork(vcd.client, externalNetwork.Name)
	check.Assert(err, IsNil)

	// Workaround to refresh until task is fully completed - as task wait isn't enough
	// Task still exists and creates NETWORK_DELETE error, so we wait until disappears
	for i := 0; i < 30; i++ {
		err = createdExternalNetwork.Refresh()
		check.Assert(err, IsNil)
		if createdExternalNetwork.ExternalNetwork.Tasks != nil && len(createdExternalNetwork.ExternalNetwork.Tasks.Task) == 0 {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	err = createdExternalNetwork.DeleteWait()
	if err != nil {
		AddToCleanupList(externalNetwork.Name, "externalNetwork", "", "Test_ExternalNetworkDelete")
	}
	check.Assert(err, IsNil)

	// check through existing catalogItems
	_, err = GetExternalNetwork(vcd.client, externalNetwork.Name)
	check.Assert(err, ErrorMatches, "external network.*not found")
}
