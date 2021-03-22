// +build org functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"
	"math"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Tests org function GetVDCByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found VDC, or
// if the invalid vdc is found by the function.  Also tests a VDC
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_CreateOrgVdcWithFlex(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
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
	ctx := context.Background()

	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	results, err := vcd.client.QueryWithNotEncodedParams(ctx, nil, map[string]string{
		"type":   "providerVdc",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.Name),
	})
	check.Assert(err, IsNil)
	if len(results.Results.VMWProviderVdcRecord) == 0 {
		check.Skip(fmt.Sprintf("No Provider VDC found with name '%s'", vcd.config.VCD.ProviderVdc.Name))
	}
	providerVdcHref := results.Results.VMWProviderVdcRecord[0].HREF

	results, err = vcd.client.QueryWithNotEncodedParams(ctx, nil, map[string]string{
		"type":   "providerVdcStorageProfile",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.StorageProfile),
	})
	check.Assert(err, IsNil)
	if len(results.Results.ProviderVdcStorageProfileRecord) == 0 {
		check.Skip(fmt.Sprintf("No storage profile found with name '%s'", vcd.config.VCD.ProviderVdc.StorageProfile))
	}
	providerVdcStorageProfileHref := results.Results.ProviderVdcStorageProfileRecord[0].HREF

	results, err = vcd.client.QueryWithNotEncodedParams(ctx, nil, map[string]string{
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
			VdcStorageProfile: []*types.VdcStorageProfileConfiguration{&types.VdcStorageProfileConfiguration{
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

		vdc, _ := adminOrg.GetVDCByName(ctx, vdcConfiguration.Name, false)
		if vdc != nil {
			err = vdc.DeleteWait(ctx, true, true)
			check.Assert(err, IsNil)
		}

		// expected to fail due to missing value
		task, err := adminOrg.CreateOrgVdcAsync(ctx, vdcConfiguration)
		check.Assert(err, Not(IsNil))
		check.Assert(task, Equals, Task{})
		// checks function validation
		check.Assert(err.Error(), Equals, "VdcConfiguration missing required field: ComputeCapacity[0].Memory.Units")

		vdcConfiguration.ComputeCapacity[0].Memory.Units = "MB"

		vdc, err = adminOrg.CreateOrgVdc(ctx, vdcConfiguration)
		check.Assert(vdc, NotNil)
		check.Assert(err, IsNil)

		AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, "Test_CreateVdcWithFlex")

		vdc, err = adminOrg.GetVDCByName(ctx, vdcConfiguration.Name, true)
		check.Assert(err, IsNil)
		check.Assert(vdc, NotNil)
		check.Assert(vdc.Vdc.Name, Equals, vdcConfiguration.Name)
		check.Assert(vdc.Vdc.IsEnabled, Equals, vdcConfiguration.IsEnabled)
		check.Assert(vdc.Vdc.AllocationModel, Equals, vdcConfiguration.AllocationModel)

		err = vdc.DeleteWait(ctx, true, true)
		check.Assert(err, IsNil)

		vdc, err = adminOrg.GetVDCByName(ctx, vdcConfiguration.Name, true)
		check.Assert(err, NotNil)
		check.Assert(vdc, IsNil)
	}
}

// Tests VDC by updating it and then asserting if the
// variable is updated.
func (vcd *TestVCD) Test_UpdateVdcFlex(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	ctx := context.Background()

	adminOrg, vdcConfiguration, err := setupVdc(vcd, check, "Flex")
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(ctx, vdcConfiguration.Name, true)

	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)
	check.Assert(adminVdc.AdminVdc.Name, Equals, vdcConfiguration.Name)
	check.Assert(adminVdc.AdminVdc.IsEnabled, Equals, vdcConfiguration.IsEnabled)
	check.Assert(adminVdc.AdminVdc.AllocationModel, Equals, vdcConfiguration.AllocationModel)

	updateDescription := "updateDescription"
	computeCapacity := []*types.ComputeCapacity{
		&types.ComputeCapacity{
			CPU: &types.CapacityWithUsage{
				Units:     "MHz",
				Allocated: 2024,
				Limit:     2024,
			},
			Memory: &types.CapacityWithUsage{
				Allocated: 2024,
				Limit:     2024,
				Units:     "MB",
			},
		},
	}
	quota := 111
	vCpu := int64(1000)
	guaranteed := float64(0.6)
	adminVdc.AdminVdc.Description = updateDescription
	adminVdc.AdminVdc.ComputeCapacity = computeCapacity
	adminVdc.AdminVdc.IsEnabled = false
	falseRef := false
	trueRef := true
	adminVdc.AdminVdc.IsThinProvision = &falseRef
	adminVdc.AdminVdc.NetworkQuota = quota
	adminVdc.AdminVdc.VMQuota = quota
	adminVdc.AdminVdc.OverCommitAllowed = false
	adminVdc.AdminVdc.VCpuInMhz = &vCpu
	adminVdc.AdminVdc.UsesFastProvisioning = &falseRef
	adminVdc.AdminVdc.ResourceGuaranteedCpu = &guaranteed
	adminVdc.AdminVdc.ResourceGuaranteedMemory = &guaranteed
	adminVdc.AdminVdc.IsElastic = &trueRef
	adminVdc.AdminVdc.IncludeMemoryOverhead = &falseRef

	updatedVdc, err := adminVdc.Update(ctx)
	check.Assert(err, IsNil)
	check.Assert(updatedVdc, Not(IsNil))
	check.Assert(updatedVdc.AdminVdc.Description, Equals, updateDescription)
	check.Assert(updatedVdc.AdminVdc.ComputeCapacity[0].CPU.Allocated, Equals, computeCapacity[0].CPU.Allocated)
	check.Assert(updatedVdc.AdminVdc.IsEnabled, Equals, false)
	check.Assert(*updatedVdc.AdminVdc.IsThinProvision, Equals, false)
	check.Assert(updatedVdc.AdminVdc.NetworkQuota, Equals, quota)
	check.Assert(updatedVdc.AdminVdc.VMQuota, Equals, quota)
	check.Assert(updatedVdc.AdminVdc.OverCommitAllowed, Equals, false)
	check.Assert(*updatedVdc.AdminVdc.VCpuInMhz, Equals, vCpu)
	check.Assert(*updatedVdc.AdminVdc.UsesFastProvisioning, Equals, false)
	check.Assert(math.Abs(*updatedVdc.AdminVdc.ResourceGuaranteedCpu-guaranteed) < 0.001, Equals, true)
	check.Assert(math.Abs(*updatedVdc.AdminVdc.ResourceGuaranteedMemory-guaranteed) < 0.001, Equals, true)
	check.Assert(*updatedVdc.AdminVdc.IsElastic, Equals, true)
	check.Assert(*updatedVdc.AdminVdc.IncludeMemoryOverhead, Equals, false)
}

// Tests VDC storage profile update
func (vcd *TestVCD) Test_VdcUpdateStorageProfile(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	ctx := context.Background()

	adminOrg, vdcConfiguration, err := setupVdc(vcd, check, "Flex")
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(ctx, vdcConfiguration.Name, true)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)

	foundStorageProfile, err := GetStorageProfileByHref(ctx, vcd.client, adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile[0].HREF)
	check.Assert(err, IsNil)
	check.Assert(foundStorageProfile, Not(Equals), types.VdcStorageProfile{})
	check.Assert(foundStorageProfile, NotNil)

	storageProfileId, err := GetUuidFromHref(adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile[0].HREF, true)
	check.Assert(err, IsNil)
	check.Assert(storageProfileId, NotNil)

	updatedVdc, err := adminVdc.UpdateStorageProfile(ctx, storageProfileId, &types.AdminVdcStorageProfile{
		Name:                      foundStorageProfile.ProviderVdcStorageProfile.Name,
		Default:                   true,
		Limit:                     9081,
		Enabled:                   takeBoolPointer(true),
		Units:                     "MB",
		IopsSettings:              nil,
		ProviderVdcStorageProfile: &types.Reference{HREF: foundStorageProfile.ProviderVdcStorageProfile.HREF},
	})
	check.Assert(err, IsNil)
	check.Assert(updatedVdc, Not(IsNil))

	updatedStorageProfile, err := GetStorageProfileByHref(ctx, vcd.client, adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile[0].HREF)
	check.Assert(err, IsNil)
	check.Assert(updatedStorageProfile, Not(Equals), types.VdcStorageProfile{})
	check.Assert(updatedStorageProfile, NotNil)

	check.Assert(updatedStorageProfile.Enabled, Equals, true)
	check.Assert(updatedStorageProfile.Limit, Equals, int64(9081))
	check.Assert(updatedStorageProfile.Default, Equals, true)
	check.Assert(updatedStorageProfile.Units, Equals, "MB")
}
