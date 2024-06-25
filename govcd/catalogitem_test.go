//go:build catalog || functional || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	//"strings"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (vcd *TestVCD) Test_GetVAppTemplate(check *C) {

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

	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)

	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()

	check.Assert(err, IsNil)
	check.Assert(vapptemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(vapptemplate.VAppTemplate.Description, Equals, vcd.config.VCD.Catalog.CatalogItemDescription)
	}
}

// Tests System function Delete by creating catalog item and
// deleting it after.
func (vcd *TestVCD) Test_Delete(check *C) {
	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)
	AddToCleanupList(TestDeleteCatalogItem, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_Delete")

	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	// add catalogItem
	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, TestDeleteCatalogItem, "upload from delete catalog item test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	catalogItem, err := catalog.GetCatalogItemByName(TestDeleteCatalogItem, false)
	check.Assert(err, IsNil)

	err = catalogItem.Delete()
	check.Assert(err, IsNil)

	// check through existing catalogItems
	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	entityFound := false
	for _, catalogItems := range catalog.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if catalogItem.Name == TestDeleteCatalogItem {
				entityFound = true
			}
		}
	}
	check.Assert(entityFound, Equals, false)
}

func (vcd *TestVCD) TestQueryCatalogItemAndVAppTemplateList(check *C) {
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("no catalog provided. Skipping test")
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("no VDC provided. Skipping test")
	}
	// Fetching organization and catalog
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)

	// Fetching VDC
	vdc, err := org.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Get the list of catalog items using a query from catalog
	queryCatalogItemsByCatalog, err := catalog.QueryCatalogItemList()
	check.Assert(err, IsNil)
	check.Assert(queryCatalogItemsByCatalog, NotNil)

	// Get the list of catalog items using a query from VDC
	queryCatalogItemsByVdc, err := vdc.QueryCatalogItemList()
	check.Assert(err, IsNil)
	check.Assert(queryCatalogItemsByVdc, NotNil)

	// Make sure the lists have at least one item
	hasItemsFromCatalog := len(queryCatalogItemsByCatalog) > 0
	check.Assert(hasItemsFromCatalog, Equals, true)
	hasItemsFromVdc := len(queryCatalogItemsByVdc) > 0
	check.Assert(hasItemsFromVdc, Equals, true)

	// Building up a map of catalog items as they are recorded in the catalog
	var itemMapInCatalog = make(map[string]string)
	for _, item := range catalog.Catalog.CatalogItems {
		for _, catalogItem := range item.CatalogItem {
			itemMapInCatalog[catalogItem.Name] = catalogItem.HREF
		}
	}

	var itemMapInVdc = make(map[string]string)
	for _, resource := range vdc.AdminVdc.ResourceEntities {
		for _, item := range resource.ResourceEntity {
			if item.Type == types.MimeVAppTemplate {
				itemMapInVdc[item.Name] = item.HREF
			}
		}
	}

	// Compare the items in the query with the catalog list
	for _, qCatalogItem := range queryCatalogItemsByCatalog {
		itemHref, foundItem := itemMapInCatalog[qCatalogItem.Name]
		check.Assert(foundItem, Equals, true)
		if qCatalogItem.EntityType == types.QtVappTemplate {
			// If the item is not "media", compare the HREF
			check.Assert(itemHref, Equals, qCatalogItem.HREF)
		}
	}

	// Get the list of vApp templates using a query from catalog
	queryVappTemplatesByCatalog, err := catalog.QueryVappTemplateList()
	check.Assert(err, IsNil)
	check.Assert(queryVappTemplatesByCatalog, NotNil)

	// Get the list of vApp templates using a query from VDC
	queryVappTemplatesByVdc, err := vdc.QueryVappTemplateList()
	check.Assert(err, IsNil)
	check.Assert(queryVappTemplatesByVdc, NotNil)

	// Make sure the lists have at least one item
	hasItemsFromCatalog = len(queryVappTemplatesByCatalog) > 0
	check.Assert(hasItemsFromCatalog, Equals, true)
	hasItemsFromVdc = len(queryVappTemplatesByVdc) > 0
	check.Assert(hasItemsFromVdc, Equals, true)

	// Compare vApp templates with the list of  catalog items
	for _, qvAppTemplate := range queryVappTemplatesByCatalog {
		// Check that catalog item name and vApp template names match
		itemHref, foundItem := itemMapInCatalog[qvAppTemplate.Name]
		check.Assert(foundItem, Equals, true)

		// Retrieve the catalog item, and check the internal Entity HREF
		// against the vApp template HREF
		catalogItem, err := catalog.GetCatalogItemByHref(itemHref)
		check.Assert(err, IsNil)
		check.Assert(catalogItem, NotNil)

		check.Assert(catalogItem.CatalogItem.Entity, NotNil)
		check.Assert(catalogItem.CatalogItem.Entity.HREF, Equals, qvAppTemplate.HREF)
	}

	// Compare vApp templates from query with the list of vappTemplates from VDC
	for _, qvAppTemplate := range queryVappTemplatesByVdc {
		// Check that catalog item name and vApp template names match
		itemHref, foundItem := itemMapInVdc[qvAppTemplate.Name]
		check.Assert(foundItem, Equals, true)

		// Retrieve the vApp template
		vappTemplate, err := catalog.GetVappTemplateByHref(itemHref)
		check.Assert(err, IsNil)
		check.Assert(vappTemplate, NotNil)
	}

	// Compare vApp templates from query with one retrieved by name
	vAppTemplateQueryResult, err := catalog.QueryVappTemplateWithName(queryVappTemplatesByCatalog[0].Name)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplateQueryResult, NotNil)
	check.Assert(vAppTemplateQueryResult, DeepEquals, queryVappTemplatesByCatalog[0])
}

func (vcd *TestVCD) Test_DeleteNonEmptyCatalog(check *C) {
	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)

	catalogName := check.TestName()
	catalogItemName := check.TestName() + "_item"
	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalog, err := org.CreateCatalog(catalogName, catalogName)
	check.Assert(err, IsNil)
	AddToCleanupList(catalogName, "catalog", vcd.org.Org.Name, check.TestName())

	check.Assert(catalog, NotNil)

	// add catalogItem
	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, catalogItemName, "upload from delete catalog item test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
	AddToCleanupList(catalogItemName, "catalogItem", vcd.org.Org.Name+"|"+catalogName, check.TestName())

	retrievedCatalog, err := org.GetCatalogByName(catalogName, true)
	check.Assert(err, IsNil)
	catalogItem, err := retrievedCatalog.GetCatalogItemByName(catalogItemName, true)
	check.Assert(err, IsNil)
	check.Assert(catalogItem, NotNil)

	err = retrievedCatalog.Delete(true, true)
	check.Assert(err, IsNil)

	retrievedCatalog, err = org.GetCatalogByName(catalogName, true)
	check.Assert(err, NotNil)
	check.Assert(retrievedCatalog, IsNil)
}

func (vcd *TestVCD) Test_QueryVappTemplateList(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	catalogName := vcd.config.VCD.Catalog.Name
	if catalogName == "" {
		check.Skip("Test_QueryVappTemplateList: Catalog name not given")
		return
	}

	cat, err := vcd.org.GetCatalogByName(catalogName, false)
	if err != nil {
		check.Skip("Test_QueryVappTemplateList: Catalog not found")
		return
	}

	vAppTemplates, err := cat.QueryVappTemplateList()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplates, NotNil)

	// Check the number of vApp templates is one
	// Dump all vApp template structures to easily identify leftover objects if number is not 1
	if len(vAppTemplates) > 1 {
		fmt.Printf("%#v", vAppTemplates)
	}
	check.Assert(len(vAppTemplates), Equals, 1)

	// Check the name of the vApp template is what it should be
	check.Assert(vAppTemplates[0].Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	// Check the vApp Template retrieved before is the same as the one retrieved by name
	vAppTemplate, err := cat.QueryVappTemplateWithName(vAppTemplates[0].Name)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplates, NotNil)
	check.Assert(vAppTemplate, DeepEquals, vAppTemplates[0])
}
