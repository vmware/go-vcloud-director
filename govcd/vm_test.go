/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 * Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) find_first_vm(vapp VApp) (types.VM, string) {
	for _, vm := range vapp.VApp.Children.VM {
		if vm.Name != "" {
			return *vm, vm.Name
		}
	}
	return types.VM{}, ""
}

func (vcd *TestVCD) find_first_vapp() VApp {
	client := vcd.client
	config := vcd.config
	org, err := GetOrgByName(client, config.VCD.Org)
	if err != nil {
		fmt.Println(err)
		return VApp{}
	}
	vdc, err := org.GetVdcByName(config.VCD.Vdc)
	if err != nil {
		fmt.Println(err)
		return VApp{}
	}
	wanted_vapp := vcd.vapp.VApp.Name
	vapp_name := ""
	for _, res := range vdc.Vdc.ResourceEntities {
		for _, item := range res.ResourceEntity {
			// Finding a named vApp, if it was defined in config
			if wanted_vapp != "" {
				if item.Name == wanted_vapp {
					vapp_name = item.Name
					break
				}
			} else {
				// Otherwise, we get the first vApp from the vDC list
				if item.Type == "application/vnd.vmware.vcloud.vApp+xml" {
					vapp_name = item.Name
					break
				}
			}
		}
	}
	if wanted_vapp == "" {
		return VApp{}
	}
	vapp, _ := vdc.FindVAppByName(vapp_name)
	return vapp
}

func (vcd *TestVCD) Test_FindVMByHREF(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}

	fmt.Printf("Running: %s\n", check.TestName())
	vapp := vcd.find_first_vapp()
	if vapp.VApp.Name == "" {
		check.Skip("Disabled: No suitable vApp found in vDC")
	}
	vm, vm_name := vcd.find_first_vm(vapp)
	if vm.Name == "" {
		check.Skip("Disabled: No suitable VM found in vDC")
	}

	vm_href := vm.HREF
	new_vm, err := vcd.client.FindVMByHREF(vm_href)

	check.Assert(err, IsNil)
	check.Assert(new_vm.VM.Name, Equals, vm_name)
	check.Assert(new_vm.VM.VirtualHardwareSection.Item, NotNil)
}
