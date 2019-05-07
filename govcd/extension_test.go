// +build extnetwork network functional ALL

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
func (vcd *TestVCD) Test_GetExternalNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	networkName := vcd.config.VCD.ExternalNetwork
	if networkName == "" {
		check.Skip("No external network provided")
	}
	externalNetwork, err := GetExternalNetworkByName(vcd.client, networkName)
	check.Assert(err, IsNil)
	LogExternalNetwork(*externalNetwork)
	check.Assert(externalNetwork.HREF, Not(Equals), "")
	check.Assert(externalNetwork.Name, Equals, networkName)
	check.Assert(externalNetwork.Type, Equals, types.MimeExtensionNetwork)
}

func (vcd *TestVCD) Test_CreateExternalNetwork(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("Test_CreateExternalNetwork: External network isn't configured. Test can't proceed")
	}

	if vcd.config.VCD.VimServer == "" {
		check.Skip("Test_CreateExternalNetwork: Vim server isn't configured. Test can't proceed")
	}

	if vcd.config.VCD.ExternalNetworkPortGroup == "" {
		check.Skip("Test_CreateExternalNetwork: Port group isn't configured. Test can't proceed")
	}

	if vcd.config.VCD.ExternalNetworkPortGroupType == "" {
		check.Skip("Test_CreateExternalNetwork: Port group type isn't configured. Test can't proceed")
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
		check.Skip(fmt.Sprintf("More then one found with name '%s'", vcd.config.VCD.ExternalNetworkPortGroup))
	}

	externalNetwork := &types.ExternalNetwork{
		Name:        TestCreateExternalNetwork,
		Description: "Test Create External Network",
		Xmlns:       types.XMLNamespaceExtension,
		XmlnsVCloud: types.XMLNamespaceVCloud,
		Configuration: &types.NetworkConfiguration{
			Xmlns: types.XMLNamespaceVCloud,
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					Gateway:   "192.168.201.1",
					Netmask:   "255.255.255.0",
					DNS1:      "192.168.202.253",
					DNS2:      "192.168.202.254",
					DNSSuffix: "some.net",
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
	check.Assert(task.Task, Not(Equals), types.Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	newExternalNetwork, err := GetExternalNetwork(vcd.client, TestCreateExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(newExternalNetwork.ExternalNetwork.Name, Equals, TestCreateExternalNetwork)

	ipScope := newExternalNetwork.ExternalNetwork.Configuration.IPScopes.IPScope
	check.Assert(ipScope[0].Gateway, Equals, "192.168.201.1")
	check.Assert(ipScope[0].Netmask, Equals, "255.255.255.0")
	check.Assert(ipScope[0].DNS1, Equals, "192.168.202.253")
	check.Assert(ipScope[0].DNS2, Equals, "192.168.202.254")
	check.Assert(ipScope[0].DNSSuffix, Equals, "some.net")

	check.Assert(len(ipScope[0].IPRanges.IPRange), Equals, 1)
	ipRange := ipScope[0].IPRanges.IPRange[0]
	check.Assert(ipRange.StartAddress, Equals, "192.168.201.3")
	check.Assert(ipRange.EndAddress, Equals, "192.168.201.250")

	check.Assert(newExternalNetwork.ExternalNetwork.Configuration.FenceMode, Equals, "isolated")

	// Needs to be deleted as port group used by other tests
	// Workaround to refresh until task is fully completed - as task wait isn't enough
	// Task still exists and creates NETWORK_DELETE error, so we wait until disappears
	for i := 0; i < 30; i++ {
		err = newExternalNetwork.Refresh()
		check.Assert(err, IsNil)
		if newExternalNetwork.ExternalNetwork.Tasks != nil && len(newExternalNetwork.ExternalNetwork.Tasks.Task) == 0 {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	err = newExternalNetwork.DeleteWait()
	if err != nil {
		AddToCleanupList(externalNetwork.Name, "externalNetwork", "", "Test_CreateExternalNetwork")
	}
	check.Assert(err, IsNil)
}

func init() {
	testingTags["extnetwork"] = "extension_test.go"
}
