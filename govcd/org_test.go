//go:build org || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"math"

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
		orgName              string
		enabled              bool
		canPublishCatalogs   bool
		canPublishExternally bool
		canSubscribe         bool
	}

	// Tests a combination of enabled and canPublishCatalogs to see
	// whether they are updated correctly
	var updateOrgs = []updateSet{
		{TestUpdateOrg + "1", true, false, false, false},
		{TestUpdateOrg + "2", false, false, false, false},
		{TestUpdateOrg + "3", true, true, true, false},
		{TestUpdateOrg + "4", false, true, false, true},
	}

	for _, uo := range updateOrgs {
		if vcd.client.Client.APIVCDMaxVersionIs("= 37.2") && !uo.enabled {
			// TODO revisit once bug is fixed in VCD
			fmt.Println("[INFO] VCD 10.4.2 has a bug that prevents creating a disabled Org - Changing 'enabled' parameter to 'true'")
			uo.enabled = true
		}
		fmt.Printf("Org %s - enabled %v - catalogs %v\n", uo.orgName, uo.enabled, uo.canPublishCatalogs)
		task, err := CreateOrg(vcd.client, uo.orgName, uo.orgName, uo.orgName, &types.OrgSettings{
			OrgGeneralSettings: &types.OrgGeneralSettings{
				CanPublishCatalogs:   uo.canPublishCatalogs,
				CanPublishExternally: uo.canPublishExternally,
				CanSubscribe:         uo.canSubscribe,
			},
			OrgLdapSettings: &types.OrgLdapSettingsType{OrgLdapMode: "NONE"},
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
		adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally = !uo.canPublishExternally
		adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanSubscribe = !uo.canSubscribe

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
		check.Assert(updatedAdminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally, Equals, !uo.canPublishExternally)
		check.Assert(updatedAdminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanSubscribe, Equals, !uo.canSubscribe)
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

// Tests org function GetVDCByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found VDC, or
// if the invalid vdc is found by the function.  Also tests a VDC
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
// or the function returns an error.  Also tests a vdc
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
// Fails if the config file name doesn't match with the found VDC, or
// if the invalid vdc is found by the function.  Also tests a VDC
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
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.Name),
	})
	check.Assert(err, IsNil)
	if len(results.Results.VMWProviderVdcRecord) == 0 {
		check.Skip(fmt.Sprintf("No Provider VDC found with name '%s'", vcd.config.VCD.ProviderVdc.Name))
	}
	providerVdcHref := results.Results.VMWProviderVdcRecord[0].HREF

	storageProfile, err := vcd.client.QueryProviderVdcStorageProfileByName(vcd.config.VCD.ProviderVdc.StorageProfile, providerVdcHref)
	check.Assert(err, IsNil)

	results, err = vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "networkPool",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.NetworkPool),
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
					HREF: storageProfile.HREF,
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
		check.Assert(err, NotNil)
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
	// Immediately after the catalog creation, the creation task should be already complete
	check.Assert(ResourceComplete(adminCatalog.AdminCatalog.Tasks), Equals, true)

	adminOrg, err = vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	copyAdminCatalog, err := adminOrg.GetAdminCatalogByName(TestCreateCatalog, false)
	check.Assert(err, IsNil)
	check.Assert(copyAdminCatalog, NotNil)
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, copyAdminCatalog.AdminCatalog.Name)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, copyAdminCatalog.AdminCatalog.Description)
	check.Assert(adminCatalog.AdminCatalog.IsPublished, Equals, false)
	check.Assert(adminCatalog.AdminCatalog.CatalogStorageProfiles, NotNil)
	check.Assert(adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile, IsNil)
	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_AdminOrgCreateCatalogWithStorageProfile(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	oldAdminCatalog, _ := adminOrg.GetAdminCatalogByName(check.TestName(), false)
	if oldAdminCatalog != nil {
		err = oldAdminCatalog.Delete(true, true)
		check.Assert(err, IsNil)
	}

	// Lookup storage profile to use in catalog
	storageProfile, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)
	createStorageProfiles := &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{&storageProfile}}

	adminCatalog, err := adminOrg.CreateCatalogWithStorageProfile(check.TestName(), TestCreateCatalogDesc, createStorageProfiles)
	check.Assert(err, IsNil)
	AddToCleanupList(check.TestName(), "catalog", vcd.org.Org.Name, check.TestName())
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, check.TestName())
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, TestCreateCatalogDesc)
	// Accessing the task directly with `adminCatalog.AdminCatalog.Tasks.Task[0]` is not safe for Org user
	err = adminCatalog.WaitForTasks()
	check.Assert(err, IsNil)
	adminOrg, err = vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	copyAdminCatalog, err := adminOrg.GetAdminCatalogByName(check.TestName(), false)
	check.Assert(err, IsNil)
	check.Assert(copyAdminCatalog, NotNil)
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, copyAdminCatalog.AdminCatalog.Name)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, copyAdminCatalog.AdminCatalog.Description)
	check.Assert(adminCatalog.AdminCatalog.IsPublished, Equals, false)

	// Try to update storage profile for catalog if secondary profile is defined
	if vcd.config.VCD.StorageProfile.SP2 != "" {
		updateStorageProfile, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP2)
		check.Assert(err, IsNil)
		updateStorageProfiles := &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{&updateStorageProfile}}
		adminCatalog.AdminCatalog.CatalogStorageProfiles = updateStorageProfiles
		err = adminCatalog.Update()
		check.Assert(err, IsNil)
	} else {
		fmt.Printf("# Skipping storage profile update for %s because secondary storage profile is not provided",
			adminCatalog.AdminCatalog.Name)
	}

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
	// Immediately after the catalog creation, the creation task should be already complete
	check.Assert(ResourceComplete(catalog.Catalog.Tasks), Equals, true)
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

func (vcd *TestVCD) Test_OrgCreateCatalogWithStorageProfile(check *C) {
	org, err := vcd.client.GetOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	oldCatalog, _ := org.GetCatalogByName(check.TestName(), false)
	if oldCatalog != nil {
		err = oldCatalog.Delete(true, true)
		check.Assert(err, IsNil)
	}

	// Lookup storage profile to use in catalog
	storageProfile, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)
	storageProfiles := &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{&storageProfile}}

	catalog, err := org.CreateCatalogWithStorageProfile(check.TestName(), TestCreateCatalogDesc, storageProfiles)
	check.Assert(err, IsNil)
	AddToCleanupList(check.TestName(), "catalog", vcd.org.Org.Name, check.TestName())
	check.Assert(catalog.Catalog.Name, Equals, check.TestName())
	check.Assert(catalog.Catalog.Description, Equals, TestCreateCatalogDesc)
	err = catalog.WaitForTasks()
	check.Assert(err, IsNil)
	org, err = vcd.client.GetOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	copyCatalog, err := org.GetCatalogByName(check.TestName(), false)
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

	adminOrg, vdcConfiguration, err := setupVdc(vcd, check, "AllocationPool")
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

	//cleanup
	vdc, err := adminOrg.GetVDCByName(vdcConfiguration.Name, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	err = vdc.DeleteWait(true, true)
	check.Assert(err, IsNil)
}

func setupVdc(vcd *TestVCD, check *C, allocationModel string) (AdminOrg, *types.VdcConfiguration, error) {
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
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.Name),
	})
	check.Assert(err, IsNil)
	if len(results.Results.VMWProviderVdcRecord) == 0 {
		check.Skip(fmt.Sprintf("No Provider VDC found with name '%s'", vcd.config.VCD.ProviderVdc.Name))
	}
	providerVdcHref := results.Results.VMWProviderVdcRecord[0].HREF
	storageProfile, err := vcd.client.QueryProviderVdcStorageProfileByName(vcd.config.VCD.ProviderVdc.StorageProfile, providerVdcHref)
	check.Assert(err, IsNil)

	check.Assert(storageProfile.HREF, Not(Equals), "")
	results, err = vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "networkPool",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.NetworkPool),
	})
	check.Assert(err, IsNil)
	if len(results.Results.NetworkPoolRecord) == 0 {
		check.Skip(fmt.Sprintf("No network pool found with name '%s'", vcd.config.VCD.ProviderVdc.NetworkPool))
	}
	networkPoolHref := results.Results.NetworkPoolRecord[0].HREF
	vdcConfiguration := &types.VdcConfiguration{
		Name:            check.TestName(),
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
					Units:     "MB",
				},
			},
		},
		VdcStorageProfile: []*types.VdcStorageProfileConfiguration{&types.VdcStorageProfileConfiguration{
			Enabled: addrOf(true),
			Units:   "MB",
			Limit:   1024,
			Default: true,
			ProviderVdcStorageProfile: &types.Reference{
				HREF: storageProfile.HREF,
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
	trueValue := true
	falseValue := true
	if allocationModel == "Flex" {
		vdcConfiguration.IsElastic = &falseValue
		vdcConfiguration.IncludeMemoryOverhead = &trueValue
		vdcConfiguration.ResourceGuaranteedMemory = addrOf(1.00)
	}

	vdc, _ := adminOrg.GetVDCByName(vdcConfiguration.Name, false)
	if vdc != nil {
		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)
	}
	_, err = adminOrg.CreateOrgVdc(vdcConfiguration)
	check.Assert(err, IsNil)
	AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, check.TestName())
	return *adminOrg, vdcConfiguration, err
}

func (vcd *TestVCD) Test_QueryStorageProfiles(check *C) {

	// retrieve Org and VDC
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)

	if adminVdc.AdminVdc.ProviderVdcReference == nil {
		check.Skip(fmt.Sprintf("test %s requires system administrator privileges", check.TestName()))
	}
	// Gets the Provider VDC from the AdminVdc structure
	providerVdcName := adminVdc.AdminVdc.ProviderVdcReference.Name
	check.Assert(providerVdcName, Not(Equals), "")
	providerVdcHref := adminVdc.AdminVdc.ProviderVdcReference.HREF
	check.Assert(providerVdcHref, Not(Equals), "")

	// Gets the full list of storage profilers
	rawSpList, err := vcd.client.Client.QueryAllProviderVdcStorageProfiles()
	check.Assert(err, IsNil)

	// Manually select the storage profiles that belong to the current provider VDC
	var spList []*types.QueryResultProviderVdcStorageProfileRecordType
	var duplicateNames = make(map[string]bool)
	var notLocalStorageProfile string
	var used = make(map[string]bool)
	for _, sp := range rawSpList {
		if sp.ProviderVdcHREF == providerVdcHref {
			spList = append(spList, sp)
		}
		_, seen := used[sp.Name]
		if seen {
			duplicateNames[sp.Name] = true
		}
		used[sp.Name] = true
	}
	// Find a storage profile from a different provider VDC
	for _, sp := range rawSpList {
		if sp.ProviderVdcHREF != providerVdcHref {
			_, isDuplicate := duplicateNames[sp.Name]
			if !isDuplicate {
				notLocalStorageProfile = sp.Name
			}
		}
	}

	// Get the list of local storage profiles (belonging to the Provider VDC that the adminVdc depends on)
	localSpList, err := vcd.client.Client.QueryProviderVdcStorageProfiles(providerVdcHref)
	check.Assert(err, IsNil)
	// Make sure the automated list and the manual list match
	check.Assert(spList, DeepEquals, localSpList)

	// Get the same list using the AdminVdc method and check that the result matches
	compatibleSpList, err := adminVdc.QueryCompatibleStorageProfiles()
	check.Assert(err, IsNil)
	check.Assert(compatibleSpList, DeepEquals, localSpList)

	for _, sp := range compatibleSpList {
		fullSp, err := vcd.client.QueryProviderVdcStorageProfileByName(sp.Name, providerVdcHref)
		check.Assert(err, IsNil)
		check.Assert(sp.HREF, Equals, fullSp.HREF)
		check.Assert(fullSp.ProviderVdcHREF, Equals, providerVdcHref)
	}

	// When we have duplicate names, we also check the effectiveness of the retrieval function with Provider VDC filter
	for name := range duplicateNames {
		// Duplicate name with specific provider VDC HREF will succeed
		fullSp, err := vcd.client.QueryProviderVdcStorageProfileByName(name, providerVdcHref)
		check.Assert(err, IsNil)
		check.Assert(fullSp.ProviderVdcHREF, Equals, providerVdcHref)
		// Duplicate name with empty provider VDC HREF will fail
		faultySp, err := vcd.client.QueryProviderVdcStorageProfileByName(name, "")
		check.Assert(err, NotNil)
		check.Assert(faultySp, IsNil)
	}

	// Search explicitly for a storage profile not present in current provider VDC
	if notLocalStorageProfile != "" {
		fullSp, err := vcd.client.QueryProviderVdcStorageProfileByName(notLocalStorageProfile, providerVdcHref)
		check.Assert(err, NotNil)
		check.Assert(fullSp, IsNil)
	}
}

func (vcd *TestVCD) Test_AddRemoveVdcStorageProfiles(check *C) {
	vcd.skipIfNotSysAdmin(check)
	if vcd.config.VCD.ProviderVdc.Name == "" {
		check.Skip("No provider VDC found in configuration")
	}
	providerVDCs, err := QueryProviderVdcByName(vcd.client, vcd.config.VCD.ProviderVdc.Name)
	check.Assert(err, IsNil)
	check.Assert(len(providerVDCs), Equals, 1)

	rawSpList, err := vcd.client.Client.QueryAllProviderVdcStorageProfiles()
	check.Assert(err, IsNil)
	var spList []*types.QueryResultProviderVdcStorageProfileRecordType
	for _, sp := range rawSpList {
		if sp.ProviderVdcHREF == providerVDCs[0].HREF {
			spList = append(spList, sp)
		}
	}

	localSpList, err := vcd.client.Client.QueryProviderVdcStorageProfiles(providerVDCs[0].HREF)
	check.Assert(err, IsNil)
	check.Assert(spList, DeepEquals, localSpList)

	const minSp = 2
	if len(spList) < minSp {
		check.Skip(fmt.Sprintf("At least %d  storage profiles are needed for this test", minSp))
	}
	var defaultSp *types.QueryResultProviderVdcStorageProfileRecordType
	var sp2 *types.QueryResultProviderVdcStorageProfileRecordType

	for i := 0; i < minSp; i++ {
		if spList[i].Name == vcd.config.VCD.ProviderVdc.StorageProfile {
			if defaultSp == nil {
				defaultSp = spList[i]
			}
		} else {
			if sp2 == nil {
				sp2 = spList[i]
			}
		}
	}

	check.Assert(defaultSp, NotNil)
	check.Assert(sp2, NotNil)

	// Create the VDC
	adminOrg, vdcConfiguration, err := setupVdc(vcd, check, "AllocationPool")
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vdcConfiguration.Name, true)
	check.Assert(err, IsNil)

	// Add another storage profile
	err = adminVdc.AddStorageProfileWait(&types.VdcStorageProfileConfiguration{
		Enabled: addrOf(true),
		Units:   "MB",
		Limit:   1024,
		Default: false,
		ProviderVdcStorageProfile: &types.Reference{
			HREF: sp2.HREF,
			Name: sp2.Name,
		},
	}, "new sp 2")
	check.Assert(err, IsNil)
	check.Assert(len(adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile), Equals, 2)

	// Find the default storage profile and makes sure it matches with the one we know to be the default
	defaultSpRef, err := adminVdc.GetDefaultStorageProfileReference()
	check.Assert(err, IsNil)
	check.Assert(defaultSp.Name, Equals, defaultSpRef.Name)

	// Remove the second storage profile
	err = adminVdc.RemoveStorageProfileWait(sp2.Name)
	check.Assert(err, IsNil)
	check.Assert(len(adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile), Equals, 1)

	// Add the second storage profile again
	err = adminVdc.AddStorageProfileWait(&types.VdcStorageProfileConfiguration{
		Enabled: addrOf(true),
		Units:   "MB",
		Limit:   1024,
		Default: false,
		ProviderVdcStorageProfile: &types.Reference{
			HREF: sp2.HREF,
			Name: sp2.Name,
		},
	}, "new sp 2")

	check.Assert(err, IsNil)
	check.Assert(len(adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile), Equals, 2)

	// Change default storage profile from the original one to the second one
	err = adminVdc.SetDefaultStorageProfile(sp2.Name)
	check.Assert(err, IsNil)

	// Check that the default storage profile was changed
	defaultSpRef, err = adminVdc.GetDefaultStorageProfileReference()
	check.Assert(err, IsNil)
	check.Assert(defaultSpRef.Name, Equals, sp2.Name)

	// Set the default storage profile again to the same item.
	// This proves that SetDefaultStorageProfile is idempotent
	err = adminVdc.SetDefaultStorageProfile(sp2.Name)
	check.Assert(err, IsNil)
	defaultSpRef, err = adminVdc.GetDefaultStorageProfileReference()
	check.Assert(err, IsNil)
	check.Assert(defaultSpRef.Name, Equals, sp2.Name)

	// Remove the former default storage profile
	err = adminVdc.RemoveStorageProfileWait(defaultSp.Name)
	check.Assert(err, IsNil)
	check.Assert(len(adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile), Equals, 1)

	// Delete the VDC
	vdc, err := adminOrg.GetVDCByName(adminVdc.AdminVdc.Name, false)
	check.Assert(err, IsNil)
	err = vdc.DeleteWait(true, true)
	check.Assert(err, IsNil)
}

// Tests VDC by updating it and then asserting if the
// variable is updated.
func (vcd *TestVCD) Test_UpdateVdc(check *C) {
	adminOrg, vdcConfiguration, err := setupVdc(vcd, check, "AllocationPool")
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
	adminVdc.AdminVdc.VCpuInMhz = &vCpu
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
	check.Assert(*updatedVdc.AdminVdc.VCpuInMhz, Equals, vCpu)
	check.Assert(*updatedVdc.AdminVdc.UsesFastProvisioning, Equals, false)
	check.Assert(math.Abs(*updatedVdc.AdminVdc.ResourceGuaranteedCpu-guaranteed) < 0.001, Equals, true)
	check.Assert(math.Abs(*updatedVdc.AdminVdc.ResourceGuaranteedMemory-guaranteed) < 0.001, Equals, true)
	vdc, err := adminOrg.GetVDCByName(updatedVdc.AdminVdc.Name, true)
	check.Assert(err, IsNil)
	task, err := vdc.Delete(true, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// Tests org function GetAdminVdcByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found VDC, or
// if the invalid VDC is found by the function.  Also tests a VDC
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

// Tests VDC retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_GetTaskList(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_GetTaskList: Org name not given.")
		return
	}
	// we need to have Tasks
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp wasn't properly created")
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	taskList, err := org.GetTaskList()
	check.Assert(err, IsNil)
	check.Assert(len(taskList.Task), Not(Equals), 0)
	check.Assert(taskList.Task[0], NotNil)
	check.Assert(taskList.Task[0].ID, Not(Equals), "")
	check.Assert(taskList.Task[0].Type, Not(Equals), "")
	check.Assert(taskList.Task[0].Owner, NotNil)
	check.Assert(taskList.Task[0].Owner.HREF, Not(Equals), "")
	check.Assert(taskList.Task[0].Status, Not(Equals), "")
	check.Assert(taskList.Task[0].Progress, FitsTypeOf, 0)
}

func (vcd *TestVCD) TestQueryOrgVdcList(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("TestQueryOrgVdcList: requires admin user")
		return
	}
	if vcd.config.VCD.Org == "" {
		check.Skip("TestQueryOrgVdcList: Org name not given.")
		return
	}

	if testVerbose {
		fmt.Println("# Setting up 2 additional Orgs and 1 additional VDC")
	}

	// Pre-create two more Orgs and one VDC to test that filtering behaves correctly
	newOrgName1 := spawnTestOrg(vcd, check, "org1")
	newOrgName2 := spawnTestOrg(vcd, check, "org2")
	vdc := spawnTestVdc(vcd, check, newOrgName1)

	// Dump structure
	if testVerbose {
		fmt.Println("# Org and VDC structure layout")
		queryOrgList := []string{"System", vcd.config.VCD.Org, newOrgName1, newOrgName2}
		for _, orgName := range queryOrgList {
			org, err := vcd.client.GetOrgByName(orgName)
			check.Assert(err, IsNil)
			check.Assert(org, NotNil)

			vdcs, err := org.QueryOrgVdcList()
			check.Assert(err, IsNil)
			if testVerbose {
				fmt.Printf("VDCs for Org '%s'\n", orgName)
				for i, vdc := range vdcs {
					fmt.Printf("%d %s -> %s\n", i+1, vdc.OrgName, vdc.Name)
				}
				fmt.Println()
			}
		}
		fmt.Println("")
	}

	// expectedVdcCountInSystem = 1 NSX-V VDC
	expectedVdcCountInSystem := 1
	// If an NSX-T VDC exists - then expected count of VDCs is at least 2
	if vcd.config.VCD.Nsxt.Vdc != "" {
		expectedVdcCountInSystem++
	}

	// System Org does not directly report any child VDCs
	validateQueryOrgVdcResults(vcd, check, "Org should have no VDCs", "System", addrOf(0), nil)
	validateQueryOrgVdcResults(vcd, check, fmt.Sprintf("Should have 1 VDC %s", vdc.Vdc.Name), newOrgName1, addrOf(1), nil)
	validateQueryOrgVdcResults(vcd, check, "Should have 0 VDCs", newOrgName2, addrOf(0), nil)
	// Main Org 'vcd.config.VCD.Org' is expected to have at least (expectedVdcCountInSystem). Might be more if there are
	// more VDCs created manually
	validateQueryOrgVdcResults(vcd, check, fmt.Sprintf("Should have %d VDCs or more", expectedVdcCountInSystem), vcd.config.VCD.Org, nil, &expectedVdcCountInSystem)

	task, err := vdc.Delete(true, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	org1, err := vcd.client.GetAdminOrgByName(newOrgName1)
	check.Assert(err, IsNil)
	err = org1.Delete(true, true)
	check.Assert(err, IsNil)
	org2, err := vcd.client.GetAdminOrgByName(newOrgName2)
	check.Assert(err, IsNil)
	err = org2.Delete(true, true)
	check.Assert(err, IsNil)
}

func validateQueryOrgVdcResults(vcd *TestVCD, check *C, name, orgName string, expectedVdcCount, expectedVdcCountOrMore *int) {
	if testVerbose {
		fmt.Printf("# Checking VDCs in Org '%s' (%s):\n", orgName, name)
	}

	org, err := vcd.client.GetOrgByName(orgName)
	check.Assert(err, IsNil)
	orgList, err := org.QueryOrgVdcList()
	check.Assert(err, IsNil)

	// Number of components should be equal to the one returned by 'adminOrg.GetAllVDCs' which looks up VDCs in
	// <AdminOrg> structure
	adminOrg, err := vcd.client.GetAdminOrgByName(orgName)
	check.Assert(err, IsNil)
	allVdcs, err := adminOrg.GetAllVDCs(true)
	check.Assert(err, IsNil)
	check.Assert(len(orgList), Equals, len(allVdcs))

	// Ensure the expected count of VDCs is found
	if expectedVdcCount != nil {
		check.Assert(len(orgList), Equals, *expectedVdcCount)
	}
	// Ensure that no less than 'expectedVdcCountOrMore' VDCs found in object. This validation allows to have more than
	if expectedVdcCountOrMore != nil {
		check.Assert(len(orgList) >= *expectedVdcCountOrMore, Equals, true)
	}

	if testVerbose {
		if expectedVdcCount != nil {
			fmt.Printf("Got %d VDCs in Org '%s'. Expected (%d)\n", len(orgList), orgName, *expectedVdcCount)
		}

		if expectedVdcCountOrMore != nil {
			fmt.Printf("Got %d VDCs in Org '%s'. Expected (%d) or more\n", len(orgList), orgName, *expectedVdcCountOrMore)
		}
	}

	// Ensure that all VDCs have the same parent (or different if a query was performed for 'System')
	for index := range orgList {
		if orgName == "System" {
			check.Assert(orgList[index].OrgName, Not(Equals), orgName)
		} else {
			check.Assert(orgList[index].OrgName, Equals, orgName)

		}

	}
	if testVerbose {
		fmt.Printf("%d VDC(s) in Org '%s' have correct parent set\n", len(orgList), orgName)
		fmt.Println()
	}
}
