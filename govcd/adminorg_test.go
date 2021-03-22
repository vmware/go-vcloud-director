// +build org functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"context"
	"fmt"

	. "gopkg.in/check.v1"
)

// Creates a Catalog and then verify that finds it
func (vcd *TestVCD) Test_FindAdminCatalogRecords(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	ctx := context.Background()

	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	catalogName := "catalogForQuery"
	adminCatalog, err := adminOrg.CreateCatalog(ctx, catalogName, "catalogForQueryDescription")
	check.Assert(err, IsNil)
	AddToCleanupList(catalogName, "catalog", vcd.config.VCD.Org, check.TestName())
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, catalogName)

	// just imitate wait
	err = adminOrg.Refresh(ctx)
	check.Assert(err, IsNil)

	findRecords, err := adminOrg.FindAdminCatalogRecords(ctx, catalogName)
	check.Assert(err, IsNil)
	check.Assert(findRecords, NotNil)
	check.Assert(len(findRecords), Equals, 1)
	check.Assert(findRecords[0].Name, Equals, catalogName)
	check.Assert(findRecords[0].OrgName, Equals, adminOrg.AdminOrg.Name)
}

// Tests AdminOrg lease settings for vApp and vApp template
func (vcd *TestVCD) TestAdminOrg_SetLease(ctx context.Context, check *C) {
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
	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
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

		task, err := adminOrg.Update(ctx)
		check.Assert(err, IsNil)
		check.Assert(task, NotNil)
		err = task.WaitTaskCompletion(ctx)
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
	task, err := adminOrg.Update(ctx)
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Check that the initial parameters were restored
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeploymentLeaseSeconds, Equals, saveParams.deploymentLeaseSeconds)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.StorageLeaseSeconds, Equals, saveParams.vappStorageLease)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.PowerOffOnRuntimeLeaseExpiration, Equals, saveParams.powerOffOnRuntimeLeaseExpiration)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeleteOnStorageLeaseExpiration, Equals, saveParams.vappDeleteOnStorageLeaseExpiration)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.DeleteOnStorageLeaseExpiration, Equals, saveParams.vappTemplateDeleteOnStorageLeaseExpiration)
	check.Assert(*adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings.StorageLeaseSeconds, Equals, saveParams.vappTemplateStorageLease)

}

func (vcd *TestVCD) TestOrg_AdminOrg_QueryCatalogList(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("no org name provided. test skipped")
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("no catalog name provided. test skipped")
	}
	ctx := context.Background()

	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	// gets the catalog list as an adminOrg
	catalogsInAdminOrg, err := adminOrg.QueryCatalogList(ctx)
	check.Assert(err, IsNil)

	// gets the catalog list as an Org
	catalogsInOrg, err := org.QueryCatalogList(ctx)
	check.Assert(err, IsNil)

	foundInOrg := false
	// Searches the org catalogs list for a known catalog
	for _, catOrg := range catalogsInOrg {
		if catOrg.Name == vcd.config.VCD.Catalog.Name {
			foundInOrg = true
		}
	}
	check.Assert(foundInOrg, Equals, true)

	foundInAdminOrg := false
	// Searches the admin org catalogs list for a known catalog
	for _, catOrg := range catalogsInAdminOrg {
		if catOrg.Name == vcd.config.VCD.Catalog.Name {
			foundInAdminOrg = true
		}
	}
	check.Assert(foundInAdminOrg, Equals, true)

	// both lists should have the same number of items
	check.Assert(len(catalogsInAdminOrg), Equals, len(catalogsInOrg))

	// Check that every item in one list is also in the other list
	for _, catA := range catalogsInAdminOrg {
		foundInBoth := false
		for _, catO := range catalogsInOrg {
			if catA.Name == catO.Name {
				foundInBoth = true
			}
		}
		check.Assert(foundInBoth, Equals, true)
	}
}

// Test_GetAllVDCs checks that adminOrg.GetAllVDCs returns at least one VDC
func (vcd *TestVCD) Test_GetAllVDCs(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	ctx := context.Background()

	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	vdcs, err := adminOrg.GetAllVDCs(ctx, true)
	check.Assert(err, IsNil)
	check.Assert(len(vdcs) > 0, Equals, true)
}

// Test_GetAllStorageProfileReferences checks that adminOrg.GetAllStorageProfileReferences returns at least one storage
// profile reference
func (vcd *TestVCD) Test_GetAllStorageProfileReferences(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	ctx := context.Background()

	adminOrg, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	storageProfileReferences, err := adminOrg.GetAllStorageProfileReferences(ctx, true)
	check.Assert(err, IsNil)
	check.Assert(len(storageProfileReferences) > 0, Equals, true)
}
