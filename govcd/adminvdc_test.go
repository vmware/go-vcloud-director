// +build org functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// Tests org function GetVDCByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found vdc, or
// if the invalid vdc is found by the function.  Also tests an vdc
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_CreateOrgVdcWithFlex(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	// Flex vCD supported from 9.7 vCD
	if vcd.client.Client.APIVCDMaxVersionIs("< 32.0") {
		check.Skip(fmt.Sprintf("Test %s requires vCD 9.7 or higher", check.TestName()))
	}

	if vcd.config.VCD.ProviderVdc.Name == "" {
		check.Skip("No Provider VDC name given for VDC tests")
	}
	if vcd.config.VCD.ProviderVdc.StorageProfile == "" {
		check.Skip("No Storage Profile given for VDC tests")
	}
	if vcd.config.VCD.ProviderVdc.NetworkPool == "" {
		check.Skip("No Network Pool given for VDC tests")
	}
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	results, err := vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "providerVdc",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.Name),
	})
	check.Assert(err, IsNil)
	if len(results.Results.VMWProviderVdcRecord) == 0 {
		check.Skip(fmt.Sprintf("No Provider VDC found with name '%s'", vcd.config.VCD.ProviderVdc.Name))
	}
	providerVdcHref := results.Results.VMWProviderVdcRecord[0].HREF

	results, err = vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "providerVdcStorageProfile",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.StorageProfile),
	})
	check.Assert(err, IsNil)
	if len(results.Results.ProviderVdcStorageProfileRecord) == 0 {
		check.Skip(fmt.Sprintf("No storage profile found with name '%s'", vcd.config.VCD.ProviderVdc.StorageProfile))
	}
	providerVdcStorageProfileHref := results.Results.ProviderVdcStorageProfileRecord[0].HREF

	results, err = vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "networkPool",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.NetworkPool),
	})
	check.Assert(err, IsNil)
	if len(results.Results.NetworkPoolRecord) == 0 {
		check.Skip(fmt.Sprintf("No network pool found with name '%s'", vcd.config.VCD.ProviderVdc.NetworkPool))
	}
	networkPoolHref := results.Results.NetworkPoolRecord[0].HREF

	allocationModels := []string{"AllocationVApp", "AllocationPool", "ReservationPool", "Flex"}
	trueValue := true
	for i, allocationModel := range allocationModels {
		vdcConfiguration := &types.VdcConfiguration{
			Name:            fmt.Sprintf("%s%d", TestCreateOrgVdc, i),
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
			VdcStorageProfile: []*types.VdcStorageProfile{&types.VdcStorageProfile{
				Enabled: true,
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
		}

		vdc, _ := adminOrg.GetVDCByName(vdcConfiguration.Name, false)
		if vdc != nil {
			err = vdc.DeleteWait(true, true)
			check.Assert(err, IsNil)
		}

		task, err := adminOrg.CreateOrgVdcAsync(vdcConfiguration)
		check.Assert(err, Not(IsNil))
		check.Assert(task, Equals, Task{})
		check.Assert(err.Error(), Equals, "VdcConfiguration missing required field: ComputeCapacity[0].Memory.Units")
		vdcConfiguration.ComputeCapacity[0].Memory.Units = "MB"

		vdc, err = adminOrg.CreateOrgVdc(vdcConfiguration)
		check.Assert(vdc, NotNil)
		check.Assert(err, IsNil)

		AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, "Test_CreateVdcWithFlex")

		vdc, err = adminOrg.GetVDCByName(vdcConfiguration.Name, true)
		check.Assert(err, IsNil)
		check.Assert(vdc, NotNil)
		check.Assert(vdc.Vdc.Name, Equals, vdcConfiguration.Name)
		check.Assert(vdc.Vdc.IsEnabled, Equals, vdcConfiguration.IsEnabled)
		check.Assert(vdc.Vdc.AllocationModel, Equals, vdcConfiguration.AllocationModel)

		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)

		vdc, err = adminOrg.GetVDCByName(vdcConfiguration.Name, true)
		check.Assert(err, NotNil)
		check.Assert(vdc, IsNil)
	}
}
