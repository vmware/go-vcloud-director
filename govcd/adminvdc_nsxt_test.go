//go:build org || functional || nsxt || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

// This file tests out NSX-T related Org VDC capabilities

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_CreateNsxtOrgVdc(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	skipNoNsxtConfiguration(vcd, check)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	pVdcs, err := QueryProviderVdcByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	if len(pVdcs) == 0 {
		check.Skip(fmt.Sprintf("No NSX-T Provider VDC found with name '%s'", vcd.config.VCD.NsxtProviderVdc.Name))
	}
	providerVdcHref := pVdcs[0].HREF

	pvdcStorageProfile, err := vcd.client.QueryProviderVdcStorageProfileByName(vcd.config.VCD.NsxtProviderVdc.StorageProfile, providerVdcHref)

	check.Assert(err, IsNil)
	providerVdcStorageProfileHref := pvdcStorageProfile.HREF

	networkPools, err := QueryNetworkPoolByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.NetworkPool)
	check.Assert(err, IsNil)
	if len(networkPools) == 0 {
		check.Skip(fmt.Sprintf("No network pool found with name '%s'", vcd.config.VCD.NsxtProviderVdc.NetworkPool))
	}

	networkPoolHref := networkPools[0].HREF

	allocationModels := []string{"AllocationVApp", "AllocationPool", "ReservationPool", "Flex"}
	trueValue := true
	for i, allocationModel := range allocationModels {
		vdcConfiguration := &types.VdcConfiguration{
			Name:            fmt.Sprintf("%s%d", "TestNsxtVdc", i),
			Xmlns:           types.XMLNamespaceVCloud,
			AllocationModel: allocationModel,
			ComputeCapacity: []*types.ComputeCapacity{
				&types.ComputeCapacity{
					CPU: &types.CapacityWithUsage{
						Units:     "MHz",
						Allocated: 1024,
						Limit:     1024,
					},
					Memory: &types.CapacityWithUsage{
						Allocated: 1024,
						Limit:     1024,
					},
				},
			},
			VdcStorageProfile: []*types.VdcStorageProfileConfiguration{&types.VdcStorageProfileConfiguration{
				Enabled: addrOf(true),
				Units:   "MB",
				Limit:   1024,
				Default: true,
				ProviderVdcStorageProfile: &types.Reference{
					HREF: providerVdcStorageProfileHref,
				},
			},
			},
			NetworkPoolReference: &types.Reference{
				HREF: networkPoolHref,
			},
			ProviderVdcReference: &types.Reference{
				HREF: providerVdcHref,
			},
			IsEnabled:            true,
			IsThinProvision:      true,
			UsesFastProvisioning: true,
		}

		if allocationModel == "Flex" {
			vdcConfiguration.IsElastic = &trueValue
			vdcConfiguration.IncludeMemoryOverhead = &trueValue
			// Memory guaranteed percentage is required when IncludeMemoryOverhead is true in VCD 10.6+
			vdcConfiguration.ResourceGuaranteedMemory = addrOf(1.00)
		}

		vdc, _ := adminOrg.GetVDCByName(vdcConfiguration.Name, false)
		if vdc != nil {
			err = vdc.DeleteWait(true, true)
			check.Assert(err, IsNil)
		}

		// expected to fail due to missing value
		task, err := adminOrg.CreateOrgVdcAsync(vdcConfiguration)
		check.Assert(err, Not(IsNil))
		check.Assert(task, Equals, Task{})
		// checks function validation
		check.Assert(err.Error(), Equals, "VdcConfiguration missing required field: ComputeCapacity[0].Memory.Units")

		vdcConfiguration.ComputeCapacity[0].Memory.Units = "MB"

		vdc, err = adminOrg.CreateOrgVdc(vdcConfiguration)
		check.Assert(vdc, NotNil)
		check.Assert(err, IsNil)

		AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, check.TestName())

		adminVdc, err := adminOrg.GetAdminVDCByName(vdcConfiguration.Name, true)
		check.Assert(err, IsNil)
		check.Assert(vdc, NotNil)
		check.Assert(vdc.Vdc.Name, Equals, vdcConfiguration.Name)
		check.Assert(vdc.Vdc.IsEnabled, Equals, vdcConfiguration.IsEnabled)
		check.Assert(vdc.Vdc.AllocationModel, Equals, vdcConfiguration.AllocationModel)
		check.Assert(len(adminVdc.AdminVdc.ResourcePoolRefs.VimObjectRef) > 0, Equals, true)

		// Test  update
		adminVdc.AdminVdc.Description = "updated-description" + check.TestName()
		updatedAdminVdc, err := adminVdc.Update()
		check.Assert(err, IsNil)
		check.Assert(updatedAdminVdc.AdminVdc, Equals, adminVdc.AdminVdc)

		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)

		vdc, err = adminOrg.GetVDCByName(vdcConfiguration.Name, true)
		check.Assert(err, NotNil)
		check.Assert(vdc, IsNil)
	}
}
