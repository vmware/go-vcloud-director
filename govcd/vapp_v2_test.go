// +build vapp functional ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) TestComposeVappV2(check *C) {
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
	var def = types.ComposeVAppParamsV2{
		Name:             name,
		Description:      description,
		PowerOn:          true,
		AllEULAsAccepted: true,
		InstantiationParams: &types.InstantiationParams{
			CustomizationSection:         nil,
			DefaultStorageProfileSection: nil,
			GuestCustomizationSection:    nil,
			LeaseSettingsSection:         nil,
			NetworkConfigSection:         nil,
			NetworkConnectionSection:     nil,
			ProductSection:               nil,
		},
		SourcedItem: []*types.SourcedCompositionItemParamV2{
			{
				Source: &types.Reference{
					HREF: vmTemplate.HREF,
					ID:   vmTemplate.ID,
					Type: vmTemplate.Type,
					Name: "vm1",
				},
			},
			{
				Source: &types.Reference{
					HREF: vmTemplate.HREF,
					ID:   vmTemplate.ID,
					Type: vmTemplate.Type,
					Name: "vm2",
				},
			},
			{
				Source: &types.Reference{
					HREF: vmTemplate.HREF,
					ID:   vmTemplate.ID,
					Type: vmTemplate.Type,
					Name: "vm3",
				},
			},
		},
		CreateItem: []*types.CreateItem{
			{
				Name:        "vm4",
				Description: "VM 4 descr",
				GuestCustomizationSection: &types.GuestCustomizationSection{
					Info:         "Specifies Guest OS Customization Settings",
					ComputerName: "vm4",
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
					MemoryResourceMb: &types.MemoryResourceMb{
						Configured: 1024,
					},
					MediaSection: nil,
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
					HardwareVersion: &types.HardwareVersion{
						Value: "vmx-14",
					},
					VirtualCpuType: "VM32",
				},
			},
		},
	}

	vapp, err := vdc.ComposeVAppV2(&def)
	check.Assert(err, IsNil)

	AddToCleanupList(name, "vapp", vdc.Vdc.Name, name)

	var task Task

	check.Assert(vapp.VAppV2.Name, Equals, name)
	check.Assert(vapp.VAppV2.Description, Equals, description)

	check.Assert(len(vapp.VAppV2.Children.VM), Equals, len(def.SourcedItem) +1 )

	task, err = vapp.Undeploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	task, err = vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}
