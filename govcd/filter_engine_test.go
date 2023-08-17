//go:build search || functional || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"os"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (vcd *TestVCD) Test_SearchSpecificVappTemplate(check *C) {
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("no catalog provided. Skipping test")
	}
	// Fetching organization and catalog
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	client := catalog.client

	baseName := "catItemQuery"
	requestData := []VappTemplateData{
		{baseName + "1", "", "", StringMap{"one": "first", "two": "second"}, false},
		{baseName + "2", "", "", StringMap{"abc": "first", "def": "dummy"}, false},
		{baseName + "3", "", "", StringMap{"one": "first", "two": "second"}, false},
		{baseName + "4", "", "", StringMap{"abc": "first", "def": "second", "xyz": "final"}, false},
		{baseName + "5", "", "", StringMap{"ghj": "first", "klm": "second"}, false},
	}

	// Upload several vApp templates (will skip if they already exist)
	data, err := HelperCreateMultipleCatalogItems(catalog, requestData, testVerbose)
	check.Assert(err, IsNil)
	check.Assert(len(data), Equals, len(requestData))
	for _, item := range data {
		if item.Created && os.Getenv("GOVCD_KEEP_TEST_OBJECTS") == "" {
			AddToCleanupList(item.Name, "catalogItem", org.AdminOrg.Name+"|"+catalog.Catalog.Name, "Test_SearchVappTemplate")
		}
	}

	queryType := client.GetQueryType(types.QtVappTemplate)

	// metadata filters
	var tests = []struct {
		key       string
		value     string
		wantItems int
		wantName  string
	}{
		{"one", "first", 2, ""},
		{"two", "^s\\w+", 2, ""},
		{"def", "dummy", 1, baseName + "2"},
		{"xyz", "final", 1, baseName + "4"},
	}

	// Test with the explicit searches defined in the sample data
	for n, tc := range tests {
		var criteria = NewFilterDef()
		err = criteria.AddMetadataFilter(tc.key, tc.value, "", false, false)
		check.Assert(err, IsNil)
		queryItems, _, err := catalog.SearchByFilter(queryType, "catalogName", criteria)
		check.Assert(err, IsNil)
		printVerbose("\n%d, %#v\n", n, tc)
		for i, item := range queryItems {
			printVerbose("%d %s\n", i, item.GetName())
		}

		lenBiggerThanZero := len(queryItems) > 0
		check.Assert(lenBiggerThanZero, Equals, true)
		lenAsExpected := len(queryItems) == tc.wantItems
		check.Assert(lenAsExpected, Equals, true)
		if tc.wantName != "" {
			check.Assert(queryItems[0].GetName(), Equals, tc.wantName)
		}
	}

	// Remove items
	for _, item := range data {
		// If the item was already found in the server (item.Created = false)
		// we skip the deletion.
		// We also skip deletion if the variable GOVCD_KEEP_TEST_OBJECTS is set
		if !item.Created || os.Getenv("GOVCD_KEEP_TEST_OBJECTS") != "" {
			continue
		}

		catalogItem, err := catalog.GetCatalogItemByName(item.Name, true)
		check.Assert(err, IsNil)
		err = catalogItem.Delete()
		check.Assert(err, IsNil)
		printVerbose("deleted %s\n", item.Name)
	}
}

func (vcd *TestVCD) Test_SearchVappTemplate(check *C) {
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("no catalog provided. Skipping test")
	}
	// Fetching organization and catalog
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	client := catalog.client

	queryType := client.GetQueryType(types.QtVappTemplate)
	// Test with any vApp templates, using mass produced filters
	filters, err := HelperMakeFiltersFromVappTemplate(catalog)
	check.Assert(err, IsNil)
	for _, fm := range filters {
		// Tests both the search by metadata through the query and the search offline after fetching the items
		for _, useApiSearch := range []bool{false, true} {

			if len(fm.Criteria.Metadata) == 0 {
				continue
			}
			fm.Criteria.UseMetadataApiFilter = useApiSearch
			printVerbose("Use metadata API filter: %v\n", useApiSearch)
			queryItems, explanation, err := catalog.SearchByFilter(queryType, "catalogName", fm.Criteria)
			check.Assert(err, IsNil)

			convertedMatch, okMatch := fm.Entity.(QueryVAppTemplate)
			check.Assert(okMatch, Equals, true)
			check.Assert(convertedMatch.Name, Equals, fm.ExpectedName)

			printVerbose("%s\n", explanation)
			check.Assert(len(queryItems), Equals, 1)
			check.Assert(queryItems[0].GetName(), Equals, fm.ExpectedName)

			converted, ok := queryItems[0].(QueryVAppTemplate)
			check.Assert(ok, Equals, true)
			check.Assert(converted.Name, Equals, fm.ExpectedName)
		}
	}
}

func (vcd *TestVCD) Test_SearchCatalogItem(check *C) {
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("no catalog provided. Skipping test")
	}
	// Fetching organization and catalog
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	client := catalog.client

	queryType := client.GetQueryType(types.QtCatalogItem)
	// Test with any catalog items, using mass produced filters
	filters, err := HelperMakeFiltersFromCatalogItem(catalog)
	check.Assert(err, IsNil)
	for _, fm := range filters {
		queryItems, explanation, err := catalog.SearchByFilter(queryType, "catalog", fm.Criteria)
		check.Assert(err, IsNil)

		convertedMatch, okMatch := fm.Entity.(QueryCatalogItem)
		check.Assert(okMatch, Equals, true)
		check.Assert(convertedMatch.Name, Equals, fm.ExpectedName)
		check.Assert(len(queryItems), Equals, 1)
		printVerbose("%s\n", explanation)
		check.Assert(queryItems[0].GetName(), Equals, fm.ExpectedName)

		converted, ok := queryItems[0].(QueryCatalogItem)
		check.Assert(ok, Equals, true)
		check.Assert(converted.Name, Equals, fm.ExpectedName)
	}
}

func (vcd *TestVCD) Test_SearchNetwork(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("no org provided. Skipping test")
	}
	client := vcd.client
	// Fetching organization and VDC
	org, err := client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Get existing networks, and create sample filters to retrieve them
	filters, err := HelperMakeFiltersFromNetworks(vdc)
	check.Assert(err, IsNil)
	check.Assert(filters, NotNil)

	for _, fm := range filters {
		queryItems, explanation, err := vdc.SearchByFilter(types.QtOrgVdcNetwork, "vdc", fm.Criteria)
		check.Assert(err, IsNil)
		printVerbose("%s\n", explanation)
		check.Assert(len(queryItems), Equals, 1)
		for i, item := range queryItems {
			printVerbose("( I) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
		}
		if len(fm.Criteria.Metadata) > 0 {
			// Search with Metadata API
			fm.Criteria.UseMetadataApiFilter = true
			queryItems, explanation, err = vdc.SearchByFilter(types.QtOrgVdcNetwork, "vdc", fm.Criteria)
			check.Assert(err, IsNil)
			check.Assert(len(queryItems), Equals, 1)
			check.Assert(queryItems[0].GetName(), Equals, fm.ExpectedName)
			printVerbose("%s\n", explanation)
			for i, item := range queryItems {
				printVerbose("(II) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
			}
		}
	}

}

func (vcd *TestVCD) Test_SearchEdgeGateway(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("no org provided. Skipping test")
	}
	client := vcd.client
	// Fetching organization and VDC
	org, err := client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Get existing edge gateways, and create sample filters to retrieve them
	filters, err := HelperMakeFiltersFromEdgeGateways(vdc)
	check.Assert(err, IsNil)
	check.Assert(filters, NotNil)

	for _, fm := range filters {
		queryItems, explanation, err := vdc.SearchByFilter(types.QtEdgeGateway, "vdc", fm.Criteria)
		check.Assert(err, IsNil)
		printVerbose("%s\n", explanation)
		check.Assert(len(queryItems), Equals, 1)
		check.Assert(queryItems[0].GetName(), Equals, fm.ExpectedName)
		for i, item := range queryItems {
			printVerbose("( I) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
		}
	}
}

func (vcd *TestVCD) Test_SearchCatalog(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("no org provided. Skipping test")
	}
	client := vcd.client
	// Fetching organization
	org, err := client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	// Get existing catalogs, and create sample filters to retrieve them
	filters, err := HelperMakeFiltersFromCatalogs(org)
	check.Assert(err, IsNil)
	check.Assert(filters, NotNil)

	queryType := client.Client.GetQueryType(types.QtCatalog)
	for _, fm := range filters {
		queryItems, explanation, err := org.SearchByFilter(queryType, fm.Criteria)
		check.Assert(err, IsNil)
		check.Assert(len(queryItems), Equals, 1)
		printVerbose("%s\n", explanation)
		check.Assert(queryItems[0].GetName(), Equals, fm.ExpectedName)
		for i, item := range queryItems {
			printVerbose("( I) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
		}
	}
}

func (vcd *TestVCD) Test_SearchMediaItem(check *C) {
	if vcd.config.VCD.Vdc == "" {
		check.Skip("no VDC provided. Skipping test")
	}
	client := vcd.client
	// Fetching organization and VDC
	org, err := client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	// Get existing media, and create sample filters to retrieve them
	filters, err := HelperMakeFiltersFromMedia(vdc, catalog.Catalog.Name)
	check.Assert(err, IsNil)
	check.Assert(filters, NotNil)

	queryType := client.Client.GetQueryType(types.QtMedia)

	for _, fm := range filters {
		queryItems, explanation, err := catalog.SearchByFilter(queryType, "catalog", fm.Criteria)
		check.Assert(err, IsNil)
		printVerbose("%s\n", explanation)
		check.Assert(len(queryItems), Equals, 1)
		check.Assert(queryItems[0].GetName(), Equals, fm.ExpectedName)
		for i, item := range queryItems {
			printVerbose("( I) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
		}
	}
}

func (vcd *TestVCD) Test_SearchOrgVdc(check *C) {
	vcd.skipIfNotSysAdmin(check) // this test creates another VDC
	if vcd.config.VCD.Vdc == "" {
		check.Skip("no VDC provided. Skipping test")
	}
	client := vcd.client
	// Fetching organization and VDC
	org, err := client.GetOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	anotherVdc := spawnTestVdc(vcd, check, vcd.org.Org.Name)
	// Add some metadata to the newly created VDC
	_, err = anotherVdc.AddMetadata("key1", "value1")
	check.Assert(err, IsNil)
	_, err = anotherVdc.AddMetadata("key2", "value2")
	check.Assert(err, IsNil)
	_, err = anotherVdc.AddMetadata("key3", "value3")
	check.Assert(err, IsNil)

	// Get existing vdc, and create sample filters to retrieve them
	filters, err := HelperMakeFiltersFromOrgVdc(org)
	check.Assert(err, IsNil)
	check.Assert(filters, NotNil)

	queryType := client.Client.GetQueryType(types.QtOrgVdc)

	for _, fm := range filters {
		queryItems, explanation, err := org.SearchByFilter(queryType, fm.Criteria)
		check.Assert(err, IsNil)
		printVerbose("%s\n", explanation)
		check.Assert(len(queryItems), Equals, 1)
		check.Assert(queryItems[0].GetName(), Equals, fm.ExpectedName)
		for i, item := range queryItems {
			printVerbose("( I) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetParentName(), item.GetName())
		}
	}
	task, err := anotherVdc.Delete(true, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}
