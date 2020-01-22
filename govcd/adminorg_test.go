// +build org functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"fmt"
	typev32 "github.com/vmware/go-vcloud-director/v2/types/v32"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// Creates a Catalog and then verify that finds it
func (vcd *TestVCD) Test_FindAdminCatalogRecords(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	catalogName := "catalogForQuery"
	adminCatalog, err := adminOrg.CreateCatalog(catalogName, "catalogForQueryDescription")
	check.Assert(err, IsNil)
	AddToCleanupList(catalogName, "catalog", vcd.config.VCD.Org, check.TestName())
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, catalogName)

	// just imitate wait
	err = adminOrg.Refresh()
	check.Assert(err, IsNil)

	findRecords, err := adminOrg.FindAdminCatalogRecords(catalogName)
	check.Assert(err, IsNil)
	check.Assert(findRecords, NotNil)
	check.Assert(len(findRecords), Equals, 1)
	check.Assert(findRecords[0].Name, Equals, catalogName)
	check.Assert(findRecords[0].OrgName, Equals, adminOrg.AdminOrg.Name)
}

// Tests AdminOrg lease settings for vApp and vApp template
func (vcd *TestVCD) TestAdminOrg_SetLease(check *C) {
	type leaseParams struct {
		deploymentLeaseSeconds                     int
		vappStorageLease                           int
		vappTemplateStorageLease                   int
		powerOffOnRuntimeLeaseExpiration           bool
		vappDeleteOnStorageLeaseExpiration         bool
		vappTemplateDeleteOnStorageLeaseExpiration bool
	}

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	// Save vApp and vApp template lease parameters
	var saveParams = leaseParams{
		deploymentLeaseSeconds:                     *adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeploymentLeaseSeconds,
		vappStorageLease:                           *adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.StorageLeaseSeconds,
		vappTemplateStorageLease:                   *adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.StorageLeaseSeconds,
		powerOffOnRuntimeLeaseExpiration:           *adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.PowerOffOnRuntimeLeaseExpiration,
		vappDeleteOnStorageLeaseExpiration:         *adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeleteOnStorageLeaseExpiration,
		vappTemplateDeleteOnStorageLeaseExpiration: *adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.DeleteOnStorageLeaseExpiration,
	}

	var leaseData = []leaseParams{
		{
			deploymentLeaseSeconds:                     0, // never expires
			vappStorageLease:                           0, // never expires
			vappTemplateStorageLease:                   0, // never expires
			powerOffOnRuntimeLeaseExpiration:           true,
			vappDeleteOnStorageLeaseExpiration:         true,
			vappTemplateDeleteOnStorageLeaseExpiration: true,
		},
		{
			deploymentLeaseSeconds:                     0, // never expires
			vappStorageLease:                           0, // never expires
			vappTemplateStorageLease:                   0, // never expires
			powerOffOnRuntimeLeaseExpiration:           false,
			vappDeleteOnStorageLeaseExpiration:         false,
			vappTemplateDeleteOnStorageLeaseExpiration: false,
		},
		{
			deploymentLeaseSeconds:                     3600,          // 1 hour
			vappStorageLease:                           3600 * 24,     // 1 day
			vappTemplateStorageLease:                   3600 * 24 * 7, // 1 week
			powerOffOnRuntimeLeaseExpiration:           true,
			vappDeleteOnStorageLeaseExpiration:         true,
			vappTemplateDeleteOnStorageLeaseExpiration: true,
		},
		{
			deploymentLeaseSeconds:                     3600,          // 1 hour
			vappStorageLease:                           3600 * 24,     // 1 day
			vappTemplateStorageLease:                   3600 * 24 * 7, // 1 week
			powerOffOnRuntimeLeaseExpiration:           false,
			vappDeleteOnStorageLeaseExpiration:         false,
			vappTemplateDeleteOnStorageLeaseExpiration: false,
		},
		{
			deploymentLeaseSeconds:                     3600 * 24 * 30,  // 1 month
			vappStorageLease:                           3600 * 24 * 90,  // 1 quarter
			vappTemplateStorageLease:                   3600 * 24 * 365, // 1 year
			powerOffOnRuntimeLeaseExpiration:           true,
			vappDeleteOnStorageLeaseExpiration:         true,
			vappTemplateDeleteOnStorageLeaseExpiration: true,
		},
		{
			deploymentLeaseSeconds:                     3600 * 24 * 30,  // 1 month
			vappStorageLease:                           3600 * 24 * 90,  // 1 quarter
			vappTemplateStorageLease:                   3600 * 24 * 365, // 1 year
			powerOffOnRuntimeLeaseExpiration:           false,
			vappDeleteOnStorageLeaseExpiration:         false,
			vappTemplateDeleteOnStorageLeaseExpiration: false,
		},
	}

	for _, info := range leaseData {

		fmt.Printf("update lease params %v\n", info)
		// Change the lease parameters for both vapp and vApp template
		adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.StorageLeaseSeconds = &info.vappStorageLease
		adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeploymentLeaseSeconds = &info.deploymentLeaseSeconds
		adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.PowerOffOnRuntimeLeaseExpiration = &info.powerOffOnRuntimeLeaseExpiration
		adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeleteOnStorageLeaseExpiration = &info.vappDeleteOnStorageLeaseExpiration

		adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.StorageLeaseSeconds = &info.vappTemplateStorageLease
		adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.DeleteOnStorageLeaseExpiration = &info.vappTemplateDeleteOnStorageLeaseExpiration

		task, err := adminOrg.Update()
		check.Assert(err, IsNil)
		check.Assert(task, NotNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)

		// Check the results
		check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeploymentLeaseSeconds, Equals, info.deploymentLeaseSeconds)
		check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.StorageLeaseSeconds, Equals, info.vappStorageLease)
		check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.PowerOffOnRuntimeLeaseExpiration, Equals, info.powerOffOnRuntimeLeaseExpiration)
		check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeleteOnStorageLeaseExpiration, Equals, info.vappDeleteOnStorageLeaseExpiration)
		check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.DeleteOnStorageLeaseExpiration, Equals, info.vappTemplateDeleteOnStorageLeaseExpiration)
		check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.StorageLeaseSeconds, Equals, info.vappTemplateStorageLease)

	}
	// Restore the initial parameters
	adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.StorageLeaseSeconds = &saveParams.vappStorageLease
	adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeploymentLeaseSeconds = &saveParams.deploymentLeaseSeconds
	adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.PowerOffOnRuntimeLeaseExpiration = &saveParams.powerOffOnRuntimeLeaseExpiration
	adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeleteOnStorageLeaseExpiration = &saveParams.vappDeleteOnStorageLeaseExpiration

	adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.StorageLeaseSeconds = &saveParams.vappTemplateStorageLease
	adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.DeleteOnStorageLeaseExpiration = &saveParams.vappTemplateDeleteOnStorageLeaseExpiration

	fmt.Printf("restore lease params %v\n", saveParams)
	task, err := adminOrg.Update()
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Check that the initial parameters were restored
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeploymentLeaseSeconds, Equals, saveParams.deploymentLeaseSeconds)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.StorageLeaseSeconds, Equals, saveParams.vappStorageLease)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.PowerOffOnRuntimeLeaseExpiration, Equals, saveParams.powerOffOnRuntimeLeaseExpiration)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeleteOnStorageLeaseExpiration, Equals, saveParams.vappDeleteOnStorageLeaseExpiration)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.DeleteOnStorageLeaseExpiration, Equals, saveParams.vappTemplateDeleteOnStorageLeaseExpiration)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.StorageLeaseSeconds, Equals, saveParams.vappTemplateStorageLease)

}

// Tests org function GetVDCByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found vdc, or
// if the invalid vdc is found by the function.  Also tests an vdc
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_CreateVdcWithFlex(check *C) {
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
		vdcConfiguration := &typev32.VdcCreateConfiguration{VdcConfiguration: types.VdcConfiguration{
			Name:            fmt.Sprintf("%s%d", TestCreateOrgVdc, i),
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
		},
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

		task, err := adminOrg.CreateVdc_v32(vdcConfiguration)
		check.Assert(err, Not(IsNil))
		check.Assert(task, Equals, Task{})
		check.Assert(err.Error(), Equals, "VdcConfiguration missing required field: ComputeCapacity[0].Memory.Units")
		vdcConfiguration.ComputeCapacity[0].Memory.Units = "MB"

		err = adminOrg.CreateVdcWait_v32(vdcConfiguration)
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
