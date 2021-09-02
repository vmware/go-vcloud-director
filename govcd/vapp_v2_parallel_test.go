//go:build vapp || functional || ALL
// +build vapp functional ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"fmt"
	"sync"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// TestRecomposeParallelVappV2 tests the deployment of several VMs in parallel
func (vcd *TestVCD) TestRecomposeParallelVappV2(check *C) {
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	name := check.TestName()
	description := "test compose raw vAppV2"

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	catalogItem, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	vappTemplate, err := catalogItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vappTemplate.VAppTemplate.Children, NotNil)
	check.Assert(vappTemplate.VAppTemplate.Children.VM, NotNil)

	computePolicies, err := org.GetAllVdcComputePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(len(computePolicies), Not(Equals), 0)
	vmTemplate := vappTemplate.VAppTemplate.Children.VM[0]

	vapp, err := vdc.ComposeVAppV2(&types.ComposeVAppParamsV2{
		Name:        name,
		Deploy:      false,
		PowerOn:     false,
		LinkedClone: false,
		Description: description,
	})

	check.Assert(err, IsNil)
	AddToCleanupList(name, "vapp", vdc.Vdc.Name, name)

	type vmDef struct {
		name       string
		definition interface{}
	}

	numOfVms := 10

	// The first VM is built from VCD internal catalog
	var vms = []vmDef{
		{"vm1", &types.VmType{
			Name:        "vm1",
			Description: "VM 1 descr",
			GuestCustomizationSection: &types.GuestCustomizationSection{
				Info:         "Specifies Guest OS Customization Settings",
				ComputerName: "vm1",
			},
			NetworkConnectionSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          takeBoolPointer(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           takeIntAddress(2),
				NumCoresPerSocket: takeIntAddress(1),
				CpuResourceMhz: &types.CpuResourceMhz{
					Configured: 0,
				},
				MemoryResourceMb: &types.MemoryResourceMb{Configured: 1024},
				MediaSection:     nil,
				DiskSection: &types.DiskSection{
					DiskSettings: []*types.DiskSettings{
						&types.DiskSettings{
							SizeMb:            1024,
							UnitNumber:        0,
							BusNumber:         0,
							AdapterType:       "5",
							ThinProvisioned:   takeBoolPointer(true),
							OverrideVmDefault: true,
						},
					},
				},
				HardwareVersion: &types.HardwareVersion{Value: "vmx-14"},
				VirtualCpuType:  "VM32",
			},
		}},
	}

	// All the other VM definitions are built from the same VM template
	for i := 2; i <= numOfVms; i++ {
		name := fmt.Sprintf("vm%d", i)
		vms = append(vms, vmDef{name, &types.Reference{
			HREF: vmTemplate.HREF,
			ID:   vmTemplate.ID,
			Type: vmTemplate.Type,
			Name: name,
		}})
	}

	wg := sync.WaitGroup{}
	wg.Add(len(vms))

	// The VM deployments are called simultaneously using goroutines
	for _, vm := range vms {
		go func(name string, creation interface{}) {
			defer wg.Done()
			reconfiguredVapp, err := CreateParallelVm(&vcd.client.Client, vapp.VAppV2.HREF, name, creation, len(vms))
			check.Assert(err, IsNil)
			check.Assert(reconfiguredVapp, NotNil)
			vapp = reconfiguredVapp
		}(vm.name, vm.definition)
	}

	wg.Wait()

	err = vapp.Refresh()
	check.Assert(err, IsNil)

	readVapp, err := vdc.GetVAppByName(name, true)
	check.Assert(err, IsNil)
	check.Assert(readVapp.VApp.Name, Equals, name)
	check.Assert(readVapp.VApp.Description, Equals, description)

	check.Assert(readVapp.VApp.Children, NotNil)
	check.Assert(readVapp.VApp.Children.VM, NotNil)
	check.Assert(len(readVapp.VApp.Children.VM), Equals, len(vms))

	check.Assert(vapp.VAppV2.Children, NotNil)
	check.Assert(vapp.VAppV2.Children.VM, NotNil)
	check.Assert(len(vapp.VAppV2.Children.VM), Equals, len(vms))

	var task Task
	task, err = vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}
