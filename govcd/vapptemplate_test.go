//go:build vapp || functional || ALL

/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

// TODO: Write test for InstantiateVAppTemplate

func (vcd *TestVCD) Test_RefreshVAppTemplate(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_GetVAppTemplate: Catalog not found. Test can't proceed")
		return
	}
	check.Assert(cat, NotNil)

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_GetVAppTemplate: Catalog Item not given. Test can't proceed")
	}

	catItem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	// Get VAppTemplate
	vAppTemplate, err := catItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	oldVAppTemplate := vAppTemplate

	err = vAppTemplate.Refresh()
	check.Assert(err, IsNil)
	check.Assert(oldVAppTemplate.VAppTemplate.ID, Equals, vAppTemplate.VAppTemplate.ID)
	check.Assert(oldVAppTemplate.VAppTemplate.Name, Equals, vAppTemplate.VAppTemplate.Name)
	check.Assert(oldVAppTemplate.VAppTemplate.HREF, Equals, vAppTemplate.VAppTemplate.HREF)

	catalogItemHref, err := vAppTemplate.GetCatalogItemHref()
	check.Assert(err, IsNil)
	check.Assert(catalogItemHref, Not(Equals), "")

	catalogItemId, err := vAppTemplate.GetCatalogItemId()
	check.Assert(err, IsNil)
	check.Assert(catalogItemId, Not(Equals), "")
}

func (vcd *TestVCD) Test_UpdateAndDeleteVAppTemplateFromOvaFile(check *C) {
	testUploadAndDeleteVAppTemplate(vcd, check, false)
}

func (vcd *TestVCD) Test_UpdateAndDeleteVAppTemplateFromUrl(check *C) {
	testUploadAndDeleteVAppTemplate(vcd, check, true)
}

func (vcd *TestVCD) Test_GetInformationFromVAppTemplate(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip(check.TestName() + ": Catalog not given in testing configuration. Test can't proceed")
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip(check.TestName() + ": Catalog Item not given in testing configuration. Test can't proceed")
	}

	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	vAppTemplate, err := catalog.GetVAppTemplateByName(vcd.config.VCD.Catalog.CatalogItem)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)

	catalogName, err := vAppTemplate.GetCatalogName()
	check.Assert(err, IsNil)
	check.Assert(catalogName, Equals, catalog.Catalog.Name)

	vdcId, err := vAppTemplate.GetVdcName()
	check.Assert(err, IsNil)
	check.Assert(vdcId, Equals, vcd.vdc.Vdc.Name)
}

func testUploadAndDeleteVAppTemplate(vcd *TestVCD, check *C, isOvfLink bool) {
	fmt.Printf("Running: %s\n", check.TestName())
	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip(check.TestName() + ": Catalog not found. Test can't proceed")
		return
	}
	check.Assert(catalog, NotNil)

	itemName := check.TestName()

	description := "upload from test"

	if isOvfLink {
		uploadTask, err := catalog.UploadOvfByLink(vcd.config.OVA.OvfUrl, itemName, description)
		check.Assert(err, IsNil)
		err = uploadTask.WaitTaskCompletion()
		check.Assert(err, IsNil)
	} else {
		task, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, itemName, description, 1024)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, check.TestName())

	vAppTemplate, err := catalog.GetVAppTemplateByName(itemName)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, itemName)

	// FIXME: Due to bug in OVF Link upload in VCD, this assert is skipped
	if !isOvfLink {
		check.Assert(vAppTemplate.VAppTemplate.Description, Equals, description)
	}

	nameForUpdate := itemName + "updated"
	descriptionForUpdate := description + "updated"

	AddToCleanupList(nameForUpdate, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, check.TestName())

	vAppTemplate.VAppTemplate.Name = nameForUpdate
	vAppTemplate.VAppTemplate.Description = descriptionForUpdate
	vAppTemplate.VAppTemplate.GoldMaster = true

	_, err = vAppTemplate.Update()
	check.Assert(err, IsNil)
	err = vAppTemplate.Refresh()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, nameForUpdate)
	check.Assert(vAppTemplate.VAppTemplate.Description, Equals, descriptionForUpdate)
	check.Assert(vAppTemplate.VAppTemplate.GoldMaster, Equals, true)

	err = vAppTemplate.Delete()
	check.Assert(err, IsNil)
	vAppTemplate, err = catalog.GetVAppTemplateByName(itemName)
	check.Assert(err, NotNil)
	check.Assert(vAppTemplate, IsNil)
}

func (vcd *TestVCD) Test_VappTemplateLeaseUpdate(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Org == "" {
		check.Skip("Organization not set in configuration")
	}
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	orgVappTemplateLease := org.AdminOrg.OrgSettings.OrgVAppTemplateSettings

	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.NsxtBackedCatalogName, false)
	if err != nil {
		check.Skip("Test_GetVAppTemplate: Catalog not found. Test can't proceed")
		return
	}
	check.Assert(catalog, NotNil)

	itemName := check.TestName()
	description := "upload from test"

	task, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, itemName, description, 1024)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.NsxtBackedCatalogName, check.TestName())

	vAppTemplate, err := catalog.GetVAppTemplateByName(itemName)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, itemName)

	lease, err := vAppTemplate.GetLease()
	check.Assert(err, IsNil)
	check.Assert(lease, NotNil)

	// Check that lease in vAppTemplate is the same as the default lease in the organization
	check.Assert(lease.StorageLeaseInSeconds, Equals, *orgVappTemplateLease.StorageLeaseSeconds)
	printVerbose("lease storage at Org level: %6d\n", *orgVappTemplateLease.StorageLeaseSeconds)
	printVerbose("lease storage in vApp Template before: %6d\n", lease.StorageLeaseInSeconds)
	secondsInDay := 60 * 60 * 24

	// Set lease to 7 days storage
	err = vAppTemplate.RenewLease(secondsInDay * 7)
	check.Assert(err, IsNil)

	// Make sure the vAppTemplate internal values were updated
	check.Assert(vAppTemplate.VAppTemplate.LeaseSettingsSection.StorageLeaseInSeconds, Equals, secondsInDay*7)

	newLease, err := vAppTemplate.GetLease()
	check.Assert(err, IsNil)
	check.Assert(newLease, NotNil)
	check.Assert(newLease.StorageLeaseInSeconds, Equals, secondsInDay*7)

	printVerbose("lease storage in vAppTemplate after: %6d\n", newLease.StorageLeaseInSeconds)

	// Set lease to "never expires", which defaults to the Org maximum lease if the Org itself has lower limits
	err = vAppTemplate.RenewLease(0)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate.VAppTemplate.LeaseSettingsSection.StorageLeaseInSeconds, Equals, *orgVappTemplateLease.StorageLeaseSeconds)

	newLease, err = vAppTemplate.GetLease()
	check.Assert(err, IsNil)
	check.Assert(newLease, NotNil)
	printVerbose("lease storage in vAppTemplate (reset): %d\n", newLease.StorageLeaseInSeconds)

	if *orgVappTemplateLease.StorageLeaseSeconds != 0 {
		// Check that setting a lease higher than allowed by the Org settings results in an error
		err = vAppTemplate.RenewLease(*orgVappTemplateLease.StorageLeaseSeconds + 3600)
		check.Assert(err, NotNil)
		// Note: the same operation in a vApp results with the lease settings silently going back to the Organization defaults
	}

	err = vAppTemplate.Delete()
	check.Assert(err, IsNil)
}
