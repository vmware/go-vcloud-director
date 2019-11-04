// +build org functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"math"
	"time"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Tests Refresh for Org by updating the org and then asserting if the
// variable is updated.
func (vcd *TestVCD) Test_RefreshOrg(check *C) {

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	adminOrg, err := vcd.client.GetAdminOrgByName(TestRefreshOrg)
	if adminOrg != nil {
		check.Assert(err, IsNil)
		err := adminOrg.Delete(true, true)
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
	org, err := vcd.client.GetOrgByName(TestRefreshOrg)
	check.Assert(err, IsNil)
	check.Assert(org.Org.Name, Equals, TestRefreshOrg)
	// fetch admin version of org for updating
	adminOrg, err = vcd.client.GetAdminOrgByName(TestRefreshOrg)
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
	org, _ := vcd.client.GetAdminOrgByName(TestDeleteOrg)
	if org != nil {
		err := org.Delete(true, true)
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

	org, err = vcd.client.GetAdminOrgByName(TestDeleteOrg)
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
	type updateSet struct {
		orgName            string
		enabled            bool
		canPublishCatalogs bool
	}

	// Tests a combination of enabled and canPublishCatalogs to see
	// whether they are updated correctly
	var updateOrgs = []updateSet{
		{TestUpdateOrg + "1", true, false},
		{TestUpdateOrg + "2", false, false},
		{TestUpdateOrg + "3", true, true},
		{TestUpdateOrg + "4", false, true},
	}

	for _, uo := range updateOrgs {

		fmt.Printf("Org %s - enabled %v - catalogs %v\n", uo.orgName, uo.enabled, uo.canPublishCatalogs)
		task, err := CreateOrg(vcd.client, uo.orgName, uo.orgName, uo.orgName, &types.OrgSettings{
			OrgGeneralSettings: &types.OrgGeneralSettings{CanPublishCatalogs: uo.canPublishCatalogs},
			OrgLdapSettings:    &types.OrgLdapSettingsType{OrgLdapMode: "NONE"},
		}, uo.enabled)
		check.Assert(err, IsNil)
		check.Assert(task, Not(Equals), Task{})
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		AddToCleanupList(uo.orgName, "org", "", "TestUpdateOrg")
		// fetch newly created org
		adminOrg, err := vcd.client.GetAdminOrgByName(uo.orgName)
		check.Assert(err, IsNil)
		check.Assert(adminOrg, NotNil)

		check.Assert(adminOrg.AdminOrg.Name, Equals, uo.orgName)
		check.Assert(adminOrg.AdminOrg.Description, Equals, uo.orgName)
		updatedDescription := "description_changed"
		updatedFullName := "full_name_changed"
		adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota = 100
		adminOrg.AdminOrg.Description = updatedDescription
		adminOrg.AdminOrg.FullName = updatedFullName
		adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs = !uo.canPublishCatalogs
		adminOrg.AdminOrg.IsEnabled = !uo.enabled

		task, err = adminOrg.Update()
		check.Assert(err, IsNil)
		check.Assert(task, Not(Equals), Task{})
		// Wait until update is complete
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)

		// Get the Org again
		updatedAdminOrg, err := vcd.client.GetAdminOrgByName(uo.orgName)
		check.Assert(err, IsNil)
		check.Assert(updatedAdminOrg, NotNil)

		check.Assert(updatedAdminOrg.AdminOrg.IsEnabled, Equals, !uo.enabled)
		check.Assert(updatedAdminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs, Equals, !uo.canPublishCatalogs)
		if testVerbose {
			fmt.Printf("[updated] Org %s - enabled %v (expected %v) - catalogs %v (expected %v)\n",
				updatedAdminOrg.AdminOrg.Name,
				updatedAdminOrg.AdminOrg.IsEnabled, !uo.enabled,
				adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs, !uo.canPublishCatalogs)
		}
		check.Assert(err, IsNil)
		check.Assert(updatedAdminOrg.AdminOrg.Description, Equals, updatedDescription)
		check.Assert(updatedAdminOrg.AdminOrg.FullName, Equals, updatedFullName)
		check.Assert(updatedAdminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota, Equals, 100)
		// Delete, with force and recursive true
		err = updatedAdminOrg.Delete(true, true)
		check.Assert(err, IsNil)
		doesOrgExist(check, vcd)
	}
}

func doesOrgExist(check *C, vcd *TestVCD) {
	var org *AdminOrg
	for i := 0; i < 30; i++ {
		org, _ = vcd.client.GetAdminOrgByName(TestDeleteOrg)
		if org == nil {
			break
		} else {
			time.Sleep(time.Second)
		}
	}
	check.Assert(org, IsNil)
}

// Tests org function GetVDCByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found vdc, or
// if the invalid vdc is found by the function.  Also tests an vdc
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_GetVdcByName(check *C) {
	vdc, err := vcd.org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
	// Try a vdc that doesn't exist
	vdc, err = vcd.org.GetVDCByName(INVALID_NAME, false)
	check.Assert(err, NotNil)
	check.Assert(vdc, IsNil)
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
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	vdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
	// Try a vdc that doesn't exist
	vdc, err = adminOrg.GetVDCByName(INVALID_NAME, false)
	check.Assert(vdc, IsNil)
	check.Assert(err, NotNil)
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
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

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

		vdc, _ := adminOrg.GetVDCByName(vdcConfiguration.Name, false)
		if vdc != nil {
			err = vdc.DeleteWait(true, true)
			check.Assert(err, IsNil)
		}

		task, err := adminOrg.CreateVdc(vdcConfiguration)
		check.Assert(err, Not(IsNil))
		check.Assert(task, Equals, Task{})
		check.Assert(err.Error(), Equals, "VdcConfiguration missing required field: ComputeCapacity[0].Memory.Units")
		vdcConfiguration.ComputeCapacity[0].Memory.Units = "MB"

		err = adminOrg.CreateVdcWait(vdcConfiguration)
		check.Assert(err, IsNil)

		AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, "Test_CreateVdc")

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

// Tests FindCatalog with Catalog in config file. Fails if the name and
// description don't match the catalog elements in the config file or if
// function returns an error.  Also tests an catalog
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_FindCatalog(check *C) {
	// Find Catalog
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
	// Check Invalid Catalog
	cat, err = vcd.org.GetCatalogByName(INVALID_NAME, false)
	check.Assert(err, NotNil)
	check.Assert(cat, IsNil)
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
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	// Find Catalog
	cat, err := adminOrg.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(cat, Not(Equals), Catalog{})
	check.Assert(err, IsNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
	// Check Invalid Catalog
	cat, err = adminOrg.GetCatalogByName(INVALID_NAME, false)
	check.Assert(err, NotNil)
	check.Assert(cat, IsNil)
}

// Tests CreateCatalog by creating a catalog using an AdminOrg and
// asserts that the catalog returned contains the right contents or if it fails.
// Then Deletes the catalog.
func (vcd *TestVCD) Test_AdminOrgCreateCatalog(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	oldAdminCatalog, _ := adminOrg.GetAdminCatalogByName(TestCreateCatalog, false)
	if oldAdminCatalog != nil {
		err = oldAdminCatalog.Delete(true, true)
		check.Assert(err, IsNil)
	}
	adminCatalog, err := adminOrg.CreateCatalog(TestCreateCatalog, TestCreateCatalogDesc)
	check.Assert(err, IsNil)
	AddToCleanupList(TestCreateCatalog, "catalog", vcd.org.Org.Name, "Test_CreateCatalog")
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, TestCreateCatalog)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, TestCreateCatalogDesc)
	task := NewTask(&vcd.client.Client)
	task.Task = adminCatalog.AdminCatalog.Tasks.Task[0]
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	adminOrg, err = vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	copyAdminCatalog, err := adminOrg.GetAdminCatalogByName(TestCreateCatalog, false)
	check.Assert(err, IsNil)
	check.Assert(copyAdminCatalog, NotNil)
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, copyAdminCatalog.AdminCatalog.Name)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, copyAdminCatalog.AdminCatalog.Description)
	check.Assert(adminCatalog.AdminCatalog.IsPublished, Equals, false)
	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
}

// Tests CreateCatalog by creating a catalog using an Org and
// asserts that the catalog returned contains the right contents or if it fails.
// Then Deletes the catalog.
func (vcd *TestVCD) Test_OrgCreateCatalog(check *C) {
	org, err := vcd.client.GetOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	oldCatalog, _ := org.GetCatalogByName(TestCreateCatalog, false)
	if oldCatalog != nil {
		err = oldCatalog.Delete(true, true)
		check.Assert(err, IsNil)
	}
	catalog, err := org.CreateCatalog(TestCreateCatalog, TestCreateCatalogDesc)
	check.Assert(err, IsNil)
	AddToCleanupList(TestCreateCatalog, "catalog", vcd.org.Org.Name, "Test_CreateCatalog")
	check.Assert(catalog.Catalog.Name, Equals, TestCreateCatalog)
	check.Assert(catalog.Catalog.Description, Equals, TestCreateCatalogDesc)
	task := NewTask(&vcd.client.Client)
	task.Task = catalog.Catalog.Tasks.Task[0]
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	org, err = vcd.client.GetOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	copyCatalog, err := org.GetCatalogByName(TestCreateCatalog, false)
	check.Assert(err, IsNil)
	check.Assert(copyCatalog, NotNil)
	check.Assert(catalog.Catalog.Name, Equals, copyCatalog.Catalog.Name)
	check.Assert(catalog.Catalog.Description, Equals, copyCatalog.Catalog.Description)
	check.Assert(catalog.Catalog.IsPublished, Equals, false)
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
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	// Find Catalog
	adminCatalog, err := adminOrg.GetAdminCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog, NotNil)
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(adminCatalog.AdminCatalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
}

// Tests Refresh for VDC by updating it and then asserting if the
// variable is updated.
func (vcd *TestVCD) Test_RefreshVdc(check *C) {

	adminOrg, vdcConfiguration, err := setupVDc(vcd, check)
	check.Assert(err, IsNil)

	// Refresh so the new VDC shows up in the org's list
	err = adminOrg.Refresh()
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vdcConfiguration.Name, false)

	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)
	check.Assert(adminVdc.AdminVdc.Name, Equals, vdcConfiguration.Name)
	check.Assert(adminVdc.AdminVdc.IsEnabled, Equals, vdcConfiguration.IsEnabled)
	check.Assert(adminVdc.AdminVdc.AllocationModel, Equals, vdcConfiguration.AllocationModel)

	adminVdc.AdminVdc.Name = TestRefreshOrgVdc
	_, err = adminVdc.Update()
	check.Assert(err, IsNil)
	AddToCleanupList(TestRefreshOrgVdc, "vdc", vcd.org.Org.Name, check.TestName())

	// Test Refresh on vdc
	err = adminVdc.Refresh()
	check.Assert(err, IsNil)
	check.Assert(adminVdc.AdminVdc.Name, Equals, TestRefreshOrgVdc)
}

func setupVDc(vcd *TestVCD, check *C) (AdminOrg, *types.VdcConfiguration, error) {
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
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
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
	vdcConfiguration := &types.VdcConfiguration{
		Name:            TestCreateOrgVdc + "ForRefresh",
		Xmlns:           types.XMLNamespaceVCloud,
		AllocationModel: "AllocationPool",
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
					Units:     "MB",
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
	vdc, _ := adminOrg.GetVDCByName(vdcConfiguration.Name, false)
	if vdc != nil {
		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)
	}
	err = adminOrg.CreateVdcWait(vdcConfiguration)
	check.Assert(err, IsNil)
	AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, check.TestName())
	return *adminOrg, vdcConfiguration, err
}

// Tests VDC by updating it and then asserting if the
// variable is updated.
func (vcd *TestVCD) Test_UpdateVdc(check *C) {
	adminOrg, vdcConfiguration, err := setupVDc(vcd, check)
	check.Assert(err, IsNil)

	// Refresh so the new VDC shows up in the org's list
	err = adminOrg.Refresh()
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vdcConfiguration.Name, false)

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
	adminVdc.AdminVdc.IsThinProvision = &falseRef
	adminVdc.AdminVdc.NetworkQuota = quota
	adminVdc.AdminVdc.VMQuota = quota
	adminVdc.AdminVdc.OverCommitAllowed = false
	adminVdc.AdminVdc.VCpuInMhz = vCpu
	adminVdc.AdminVdc.UsesFastProvisioning = &falseRef
	adminVdc.AdminVdc.ResourceGuaranteedCpu = &guaranteed
	adminVdc.AdminVdc.ResourceGuaranteedMemory = &guaranteed

	updatedVdc, err := adminVdc.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedVdc, Not(IsNil))
	check.Assert(updatedVdc.AdminVdc.Description, Equals, updateDescription)
	check.Assert(updatedVdc.AdminVdc.ComputeCapacity[0].CPU.Allocated, Equals, computeCapacity[0].CPU.Allocated)
	check.Assert(updatedVdc.AdminVdc.IsEnabled, Equals, false)
	check.Assert(*updatedVdc.AdminVdc.IsThinProvision, Equals, false)
	check.Assert(updatedVdc.AdminVdc.NetworkQuota, Equals, quota)
	check.Assert(updatedVdc.AdminVdc.VMQuota, Equals, quota)
	check.Assert(updatedVdc.AdminVdc.OverCommitAllowed, Equals, false)
	check.Assert(updatedVdc.AdminVdc.VCpuInMhz, Equals, vCpu)
	check.Assert(*updatedVdc.AdminVdc.UsesFastProvisioning, Equals, false)
	check.Assert(math.Abs(*updatedVdc.AdminVdc.ResourceGuaranteedCpu-guaranteed) < 0.001, Equals, true)
	check.Assert(math.Abs(*updatedVdc.AdminVdc.ResourceGuaranteedMemory-guaranteed) < 0.001, Equals, true)
}

// Tests org function GetAdminVdcByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found VDC, or
// if the invalid VDC is found by the function.  Also tests an VDC
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_GetAdminVdcByName(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)
	check.Assert(adminVdc.AdminVdc.Name, Equals, vcd.config.VCD.Vdc)
	// Try a vdc that doesn't exist
	adminVdc, err = adminOrg.GetAdminVDCByName(INVALID_NAME, false)
	check.Assert(err, NotNil)
	check.Assert(adminVdc, IsNil)
}

// Tests catalog retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_OrgGetCatalog(check *C) {

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return vcd.org.GetCatalogByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return vcd.org.GetCatalogById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return vcd.org.GetCatalogByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "Org",
		parentName:    vcd.config.VCD.Org,
		entityType:    "Catalog",
		entityName:    vcd.config.VCD.Catalog.Name,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}

	vcd.testFinderGetGenericEntity(def, check)
}

// Tests admin catalog retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_AdminOrgGetAdminCatalog(check *C) {
	catalogName := vcd.config.VCD.Catalog.Name
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_AdminOrgGetAdminCatalog: Org name not given.")
		return
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return adminOrg.GetAdminCatalogByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) {
		return adminOrg.GetAdminCatalogById(id, refresh)
	}
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return adminOrg.GetAdminCatalogByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "AdminOrg",
		parentName:    vcd.config.VCD.Org,
		entityType:    "AdminCatalog",
		entityName:    catalogName,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}

	vcd.testFinderGetGenericEntity(def, check)

}

// Tests catalog retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_AdminOrgGetCatalog(check *C) {
	catalogName := vcd.config.VCD.Catalog.Name

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_AdminOrgGetCatalog: Org name not given.")
		return
	}
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return adminOrg.GetCatalogByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return adminOrg.GetCatalogById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return adminOrg.GetCatalogByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "AdminOrg",
		parentName:    vcd.config.VCD.Org,
		entityType:    "Catalog",
		entityName:    catalogName,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}

	vcd.testFinderGetGenericEntity(def, check)

}

// Tests VDC retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_AdminOrgGetVdc(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_AdminOrgGetVdc: Org name not given.")
		return
	}
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) { return adminOrg.GetVDCByName(name, refresh) }
	getById := func(id string, refresh bool) (genericEntity, error) { return adminOrg.GetVDCById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) { return adminOrg.GetVDCByNameOrId(id, refresh) }

	var def = getterTestDefinition{
		parentType:    "AdminOrg",
		parentName:    vcd.config.VCD.Org,
		entityType:    "Vdc",
		getterPrefix:  "VDC",
		entityName:    vcd.config.VCD.Vdc,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

// Tests VDC retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_AdminOrgGetAdminVdc(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_AdminOrgGetAdminVdc: Org name not given.")
		return
	}
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return adminOrg.GetAdminVDCByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return adminOrg.GetAdminVDCById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return adminOrg.GetAdminVDCByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "AdminOrg",
		parentName:    vcd.config.VCD.Org,
		entityType:    "AdminVdc",
		getterPrefix:  "AdminVDC",
		entityName:    vcd.config.VCD.Vdc,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

// Tests VDC retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_OrgGetVdc(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_OrgGetVdc: Org name not given.")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) { return org.GetVDCByName(name, refresh) }
	getById := func(id string, refresh bool) (genericEntity, error) { return org.GetVDCById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) { return org.GetVDCByNameOrId(id, refresh) }

	var def = getterTestDefinition{
		parentType:    "Org",
		parentName:    vcd.config.VCD.Org,
		entityType:    "Vdc",
		getterPrefix:  "VDC",
		entityName:    vcd.config.VCD.Vdc,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}
