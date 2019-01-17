/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"time"

	"github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

// Tests Refresh for Org by updating the org and then asserting if the
// variable is updated.
func (vcd *TestVCD) Test_RefreshOrg(check *C) {

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	adminOrg, err := GetAdminOrgByName(vcd.client, TestRefreshOrg)
	if adminOrg != (AdminOrg{}) {
		err = adminOrg.Delete(true, true)
		check.Assert(err, IsNil)
	}
	task, err := CreateOrg(vcd.client, TestRefreshOrg, TestRefreshOrg, TestRefreshOrg, &types.OrgSettings{
		OrgLdapSettings: &types.OrgLdapSettingsType{OrgLdapMode: "NONE"},
	}, true)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestRefreshOrg, "org", "", "Test_RefreshOrg")

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// fetch newly created org
	org, err := GetOrgByName(vcd.client, TestRefreshOrg)
	check.Assert(err, IsNil)
	check.Assert(org.Org.Name, Equals, TestRefreshOrg)
	// fetch admin version of org for updating
	adminOrg, err = GetAdminOrgByName(vcd.client, TestRefreshOrg)
	check.Assert(err, IsNil)
	check.Assert(adminOrg.AdminOrg.Name, Equals, TestRefreshOrg)
	adminOrg.AdminOrg.FullName = TestRefreshOrgFullName
	task, err = adminOrg.Update()
	check.Assert(err, IsNil)
	// Wait until update is complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	// Test Refresh on normal org
	err = org.Refresh()
	check.Assert(err, IsNil)
	check.Assert(org.Org.FullName, Equals, TestRefreshOrgFullName)
	// Test Refresh on admin org
	err = adminOrg.Refresh()
	check.Assert(err, IsNil)
	check.Assert(adminOrg.AdminOrg.FullName, Equals, TestRefreshOrgFullName)
	// Delete, with force and recursive true
	err = adminOrg.Delete(true, true)
	check.Assert(err, IsNil)
}

// Creates an org DELETEORG and then deletes it to test functionality of
// delete org. Fails if org still exists
func (vcd *TestVCD) Test_DeleteOrg(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	org, err := GetAdminOrgByName(vcd.client, TestDeleteOrg)
	if org != (AdminOrg{}) {
		err = org.Delete(true, true)
		check.Assert(err, IsNil)
	}
	task, err := CreateOrg(vcd.client, TestDeleteOrg, TestDeleteOrg, TestDeleteOrg, &types.OrgSettings{}, true)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestDeleteOrg, "org", "", "Test_DeleteOrg")
	// fetch newly created org
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	org, err = GetAdminOrgByName(vcd.client, TestDeleteOrg)
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, TestDeleteOrg)
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	doesOrgExist(check, vcd)
}

// Creates a org UPDATEORG, changes the deployed vm quota on the org,
// and tests the update functionality of the org. Then it deletes the org.
// Fails if the deployedvmquota variable is not changed when the org is
// refetched.
func (vcd *TestVCD) Test_UpdateOrg(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	org, err := GetAdminOrgByName(vcd.client, TestUpdateOrg)
	if org != (AdminOrg{}) {
		err = org.Delete(true, true)
		check.Assert(err, IsNil)
	}
	task, err := CreateOrg(vcd.client, TestUpdateOrg, TestUpdateOrg, TestUpdateOrg, &types.OrgSettings{
		OrgLdapSettings: &types.OrgLdapSettingsType{OrgLdapMode: "NONE"},
	}, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	AddToCleanupList(TestUpdateOrg, "org", "", "TestUpdateOrg")
	// fetch newly created org
	org, err = GetAdminOrgByName(vcd.client, TestUpdateOrg)
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, TestUpdateOrg)
	check.Assert(org.AdminOrg.Description, Equals, TestUpdateOrg)
	org.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota = 100
	task, err = org.Update()
	check.Assert(err, IsNil)
	// Wait until update is complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	// Refresh
	err = org.Refresh()
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota, Equals, 100)
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	doesOrgExist(check, vcd)
}

func doesOrgExist(check *C, vcd *TestVCD) {
	var org AdminOrg
	for i := 0; i < 30; i++ {
		org, _ = GetAdminOrgByName(vcd.client, TestDeleteOrg)
		if org == (AdminOrg{}) {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}
	check.Assert(org, Equals, AdminOrg{})
}

// Tests org function GetVDCByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found vdc, or
// if the invalid vdc is found by the function.  Also tests an vdc
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_GetVdcByName(check *C) {
	vdc, err := vcd.org.GetVdcByName(vcd.config.VCD.Vdc)
	check.Assert(vdc, Not(Equals), Vdc{})
	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
	// Try a vdc that doesn't exist
	vdc, err = vcd.org.GetVdcByName(INVALID_NAME)
	check.Assert(vdc, Equals, Vdc{})
	check.Assert(err, IsNil)
}

// Tests org function Admin version of GetVDCByName with the vdc
// specified in the config file. Fails if the names don't match
// or the function returns an error.  Also tests an vdc
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_Admin_GetVdcByName(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})
	vdc, err := adminOrg.GetVdcByName(vcd.config.VCD.Vdc)
	check.Assert(vdc, Not(Equals), Vdc{})
	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
	// Try a vdc that doesn't exist
	vdc, err = adminOrg.GetVdcByName(INVALID_NAME)
	check.Assert(vdc, Equals, Vdc{})
	check.Assert(err, IsNil)
}

// Tests org function GetVDCByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found vdc, or
// if the invalid vdc is found by the function.  Also tests an vdc
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_CreateVdc(check *C) {
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
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})

	results, err := vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "providerVdc",
		"filter": fmt.Sprintf("(name==%s)", vcd.config.VCD.ProviderVdc.Name),
	})
	check.Assert(err, IsNil)
	if len(results.Results.VMWProviderVdcRecord) == 0 {
		check.Skip(fmt.Sprintf("No Provider VDC found with name '%s'", vcd.config.VCD.ProviderVdc.Name))
	}
	providerVdcHref := results.Results.VMWProviderVdcRecord[0].HREF

	results, err = vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "providerVdcStorageProfile",
		"filter": fmt.Sprintf("(name==%s)", vcd.config.VCD.ProviderVdc.StorageProfile),
	})
	check.Assert(err, IsNil)
	if len(results.Results.ProviderVdcStorageProfileRecord) == 0 {
		check.Skip(fmt.Sprintf("No storage profile found with name '%s'", vcd.config.VCD.ProviderVdc.StorageProfile))
	}
	providerVdcStorageProfileHref := results.Results.ProviderVdcStorageProfileRecord[0].HREF

	results, err = vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "networkPool",
		"filter": fmt.Sprintf("(name==%s)", vcd.config.VCD.ProviderVdc.NetworkPool),
	})
	check.Assert(err, IsNil)
	if len(results.Results.NetworkPoolRecord) == 0 {
		check.Skip(fmt.Sprintf("No network pool found with name '%s'", vcd.config.VCD.ProviderVdc.NetworkPool))
	}
	networkPoolHref := results.Results.NetworkPoolRecord[0].HREF

	allocationModels := []string{"AllocationVApp", "AllocationPool", "ReservationPool"}
	for i, allocationModel := range allocationModels {
		vdcConfiguration := &types.VdcConfiguration{
			Name:            fmt.Sprintf("%s%d", TestCreateOrgVdc, i),
			Xmlns:           "http://www.vmware.com/vcloud/v1.5",
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
			VdcStorageProfile: &types.VdcStorageProfile{
				Enabled: true,
				Units:   "MB",
				Limit:   1024,
				Default: true,
				ProviderVdcStorageProfile: &types.Reference{
					HREF: providerVdcStorageProfileHref,
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

		vdc, err := adminOrg.GetVdcByName(vdcConfiguration.Name)
		check.Assert(err, IsNil)
		if vdc != (Vdc{}) {
			err = vdc.DeleteWait(true, true)
			check.Assert(err, IsNil)
		}

		task, err := adminOrg.CreateVdc(vdcConfiguration)
		check.Assert(task, Equals, Task{})
		check.Assert(err, Not(IsNil))
		check.Assert(err.Error(), Equals, "VdcConfiguration missing required field: ComputeCapacity[0].Memory.Units")
		vdcConfiguration.ComputeCapacity[0].Memory.Units = "MB"

		err = adminOrg.CreateVdcWait(vdcConfiguration)
		check.Assert(err, IsNil)

		AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, "Test_CreateVdc")

		// Refresh so the new VDC shows up in the org's list
		err = adminOrg.Refresh()
		check.Assert(err, IsNil)

		vdc, err = adminOrg.GetVdcByName(vdcConfiguration.Name)
		check.Assert(err, IsNil)
		check.Assert(vdc, Not(Equals), Vdc{})
		check.Assert(vdc.Vdc.Name, Equals, vdcConfiguration.Name)
		check.Assert(vdc.Vdc.IsEnabled, Equals, vdcConfiguration.IsEnabled)
		check.Assert(vdc.Vdc.AllocationModel, Equals, vdcConfiguration.AllocationModel)

		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)

		err = adminOrg.Refresh()
		check.Assert(err, IsNil)
		vdc, err = adminOrg.GetVdcByName(vdcConfiguration.Name)
		check.Assert(err, IsNil)
		check.Assert(vdc, Equals, Vdc{})
	}
}

// Tests FindCatalog with Catalog in config file. Fails if the name and
// description don't match the catalog elements in the config file or if
// function returns an error.  Also tests an catalog
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_FindCatalog(check *C) {
	// Find Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(cat, Not(Equals), Catalog{})
	check.Assert(err, IsNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
	// Check Invalid Catalog
	cat, err = vcd.org.FindCatalog(INVALID_NAME)
	check.Assert(cat, Equals, Catalog{})
	check.Assert(err, IsNil)
}

// Tests Admin version of FindCatalog with Catalog in config file. Asserts an
// error if the name and description don't match the catalog elements in
// the config file or if function returns an error.  Also tests an catalog
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_Admin_FindCatalog(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	// Fetch admin org version of current test org
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	// Find Catalog
	cat, err := adminOrg.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(cat, Not(Equals), Catalog{})
	check.Assert(err, IsNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
	// Check Invalid Catalog
	cat, err = adminOrg.FindCatalog(INVALID_NAME)
	check.Assert(cat, Equals, Catalog{})
	check.Assert(err, IsNil)
}

// Tests CreateCatalog by creating a catalog named CatalogCreationTest and
// asserts that the catalog returned contains the right contents or if it fails.
// Then Deletes the catalog.
func (vcd *TestVCD) Test_CreateCatalog(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	catalog, err := org.FindAdminCatalog(TestCreateCatalog)
	if catalog != (AdminCatalog{}) {
		err = catalog.Delete(true, true)
		check.Assert(err, IsNil)
	}
	catalog, err = org.CreateCatalog(TestCreateCatalog, TestCreateCatalogDesc)
	check.Assert(err, IsNil)
	AddToCleanupList(TestCreateCatalog, "catalog", vcd.org.Org.Name, "Test_CreateCatalog")
	check.Assert(catalog.AdminCatalog.Name, Equals, TestCreateCatalog)
	check.Assert(catalog.AdminCatalog.Description, Equals, TestCreateCatalogDesc)
	task := NewTask(&vcd.client.Client)
	task.Task = catalog.AdminCatalog.Tasks.Task[0]
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	org, err = GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	copyCatalog, err := org.FindAdminCatalog(TestCreateCatalog)
	check.Assert(copyCatalog, Not(Equals), AdminCatalog{})
	check.Assert(err, IsNil)
	check.Assert(catalog.AdminCatalog.Name, Equals, copyCatalog.AdminCatalog.Name)
	check.Assert(catalog.AdminCatalog.Description, Equals, copyCatalog.AdminCatalog.Description)
	check.Assert(catalog.AdminCatalog.IsPublished, Equals, false)
	err = catalog.Delete(true, true)
	check.Assert(err, IsNil)
}

// Test for GetAdminCatalog. Gets admin version of Catalog, and asserts that
// the names and description match, and that no error is returned
func (vcd *TestVCD) Test_GetAdminCatalog(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	// Fetch admin org version of current test org
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	// Find Catalog
	cat, err := adminOrg.FindAdminCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)
	check.Assert(cat.AdminCatalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.AdminCatalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
}
