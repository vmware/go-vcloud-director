/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/types/v56"
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
	expectedType := "application/vnd.vmware.admin.vmwexternalnet+xml"
	check.Assert(externalNetwork.Name, Equals, networkName)
	check.Assert(externalNetwork.Type, Equals, expectedType)
}

func (vcd *TestVCD) Test_CreateExternalNetwork(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'System'")
	}

	virtualCenters, err := queryVirtualCenters(vcd.client, fmt.Sprintf("(name==%s)", vcd.config.VCD.VimServer))
	check.Assert(err, IsNil)
	if len(virtualCenters) == 0 {
		check.Skip(fmt.Sprintf("No vSphere server found with name '%s'", vcd.config.VCD.VimServer))
	}
	vimServerHref := virtualCenters[0].HREF

	externalNetwork := &types.ExternalNetwork{
		Name:        TestCreateExternalNetwork,
		Description: "Test Create External Network",
		Xmlns:       "http://www.vmware.com/vcloud/extension/v1.5",
		XmlnsVCloud: "http://www.vmware.com/vcloud/v1.5",
		Configuration: &types.NetworkConfiguration{
			Xmlns: "http://www.vmware.com/vcloud/v1.5",
			IPScopes: &types.IPScopes{
				IPScope: types.IPScope{
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
			},
			FenceMode: "isolated",
		},
		VimPortGroupRefs: &types.VimObjectRefs{
			VimObjectRef: []*types.VimObjectRef{
				&types.VimObjectRef{
					VimServerRef: &types.Reference{
						HREF: vimServerHref,
					},
					MoRef:         vcd.config.VCD.ExternalNetworkPortGroup,
					VimObjectType: "DV_PORTGROUP",
				},
			},
		},
	}
	task, err := CreateExternalNetwork(vcd.client, externalNetwork)
	check.Assert(err, IsNil)
	AddToCleanupList(externalNetwork.Name, "externalNetwork", "", "Test_CreateExternalNetwork")
	check.Assert(task, Not(Equals), Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	externalNetworkRef, err := GetExternalNetworkByName(vcd.client, TestCreateExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(externalNetworkRef.Name, Equals, TestCreateExternalNetwork)
}
