// +build search functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"os"
	"regexp"
	"time"

	. "gopkg.in/check.v1"
)

type vappTemplateData struct {
	name                     string
	itemCreationDate         string
	vappTemplateCreationDate string
	metadata                 stringMap
	created                  bool
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

	baseName := "catItemQuery"
	requestData := []vappTemplateData{
		{baseName + "1", "", "", stringMap{"one": "first", "two": "second"}, false},
		{baseName + "2", "", "", stringMap{"abc": "first", "def": "dummy"}, false},
		{baseName + "3", "", "", stringMap{"one": "first", "two": "second"}, false},
		{baseName + "4", "", "", stringMap{"abc": "first", "def": "second", "xyz": "final"}, false},
		{baseName + "5", "", "", stringMap{"ghj": "first", "klm": "second"}, false},
	}

	data, err := createMultipleCatalogItems(org.AdminOrg.Name, catalog, requestData)
	check.Assert(err, IsNil)
	check.Assert(len(data), Equals, len(requestData))

	queryType := QtVappTemplate

	if client.IsSysAdmin {
		queryType = QtAdminVappTemplate
	}

	var tests = []struct {
		key       string
		value     string
		wantItems int
		wantName  string
	}{
		{"one", "first", 2, ""},
		{"two", "second", 2, ""},
		{"def", "dummy", 1, baseName + "2"},
		{"xyz", "final", 1, baseName + "4"},
	}

	for n, tc := range tests {

		var criteria = NewFilterDef()
		err = criteria.AddMetadataFilter(tc.key, tc.value, "", false, false)
		check.Assert(err, IsNil)
		queryItems, _, err := client.SearchByFilter(queryType, criteria)
		check.Assert(err, IsNil)
		fmt.Printf("\n%d, %#v\n", n, tc)
		for i, item := range queryItems {
			fmt.Printf("%d %s\n", i, item.GetName())
		}

		lenBiggerThanZero := len(queryItems) > 0
		check.Assert(lenBiggerThanZero, Equals, true)
		lenAsExpected := len(queryItems) == tc.wantItems
		check.Assert(lenAsExpected, Equals, true)
		if tc.wantName != "" {
			check.Assert(queryItems[0].GetName(), Equals, tc.wantName)
		}
	}

	for _, item := range data {
		if !item.created {
			continue
		}
		catalogItem, err := catalog.GetCatalogItemByName(item.name, true)
		check.Assert(err, IsNil)
		err = catalogItem.Delete()
		check.Assert(err, IsNil)
		fmt.Printf("deleted %s\n", item.name)
	}
}

// makeFiltersFromEdgeGateways looks at the existing edge gateways and creates a set of criteria to retrieve each of them
func makeFiltersFromEdgeGateways(vdc *Vdc) ([]*FilterDef, error) {
	egwResult, err := vdc.GetEdgeGatewayRecordsType(false)
	if err != nil {
		return nil, err
	}

	if egwResult.EdgeGatewayRecord == nil || len(egwResult.EdgeGatewayRecord) == 0 {
		return []*FilterDef{}, nil
	}
	var filters = make([]*FilterDef, len(egwResult.EdgeGatewayRecord))
	for i, net := range egwResult.EdgeGatewayRecord {

		filter := NewFilterDef()
		err = filter.AddFilter(FilterNameRegex, net.Name)
		if err != nil {
			return nil, err
		}
		filters[i] = filter
	}
	return filters, nil
}

// guessMetadataType guesses the type of a metadata value from its contents
// We do this because the API doesn't return the metadata type
// If the value looks like a number, or a true/false value, the corresponding type is returned
// Otherwise, we assume it's a string.
func guessMetadataType(value string) string {
	fType := "STRING"
	reNumber := regexp.MustCompile(`^[0-9]+$`)
	reBool := regexp.MustCompile(`^(?:true|false)$`)
	if reNumber.MatchString(value) {
		fType = "NUMBER"
	}
	if reBool.MatchString(value) {
		fType = "BOOLEAN"
	}
	return fType
}

// makeFiltersFromNetworks looks at the existing networks and creates a set of criteria to retrieve each of them
func makeFiltersFromNetworks(vdc *Vdc) ([]*FilterDef, error) {
	netList, err := vdc.GetNetworkList()
	if err != nil {
		return nil, err
	}
	var filters = make([]*FilterDef, len(netList))
	for i, net := range netList {

		filter := NewFilterDef()
		err = filter.AddFilter(FilterNameRegex, net.Name)
		if err != nil {
			return nil, err
		}
		err = filter.AddFilter(FilterIp, net.DefaultGateway)
		if err != nil {
			return nil, err
		}
		metadata, err := getMetadata(vdc.client, net.HREF)
		if err == nil && metadata != nil && len(metadata.MetadataEntry) > 0 {
			for _, md := range metadata.MetadataEntry {
				fType := guessMetadataType(md.TypedValue.Value)
				err = filter.AddMetadataFilter(md.Key, md.TypedValue.Value, fType, false, false)
				if err != nil {
					return nil, err
				}
			}
		}
		filters[i] = filter
	}
	return filters, nil
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

	filters, err := makeFiltersFromNetworks(vdc)
	check.Assert(err, IsNil)
	check.Assert(filters, NotNil)

	for _, criteria := range filters {
		queryItems, explanation, err := client.Client.SearchByFilter(QtOrgVdcNetwork, criteria)
		check.Assert(err, IsNil)
		check.Assert(len(queryItems), Equals, 1)
		fmt.Printf("%s\n", explanation)
		for i, item := range queryItems {
			fmt.Printf("( I) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
		}
		if len(criteria.Metadata) > 0 {
			// Search with Metadata API
			criteria.UseMetadataApiFilter = true
			queryItems, explanation, err = client.Client.SearchByFilter(QtOrgVdcNetwork, criteria)
			check.Assert(err, IsNil)
			check.Assert(len(queryItems), Equals, 1)
			for i, item := range queryItems {
				fmt.Printf("(II) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
			}
		}
	}

	/*
		for _, value := range []string{"direct", "routed", "isolated"} {
			var criteria = NewFilterDef()
			err = criteria.AddMetadataFilter("search", value, "STRING", false, true)
			queryItems, explanation, err := client.Client.SearchByFilter(QtOrgVdcNetwork, criteria)
			check.Assert(err, IsNil)
			fmt.Printf("%s\n\n",explanation)
			for i, item := range queryItems {
				fmt.Printf("%2d %-10s %-20s %s\n",i, item.GetType(), item.GetName(), item.GetIp())
			}
			//fmt.Printf("%#v\n", queryItems)
		}
		//for _, value := range []string{"10.150.191.253", "192.168.2.1", "192.168.3.1"} {
		for _, value := range []string{`10.150.191.\d+`, `192.168.2.\d+`, `192.168.3.\d+`} {
			//for _, value := range []string{"10.150.191.250,10.150.191.255", "192.168.2.1,192.168.2.2", "192.168.3.1,192.168.3.5"} {
			var criteria = NewFilterDef()
			err = criteria.AddFilter("ip", value)
			check.Assert(err, IsNil)
			queryItems, _, err := client.Client.SearchByFilter(QtOrgVdcNetwork, criteria)
			check.Assert(err, IsNil)
			fmt.Printf("\nvalue  %s : found %d\n", value, len(queryItems))
			for _, item := range queryItems {
				fmt.Printf("%s %s\n", item.GetName(), item.GetIp())
			}
		}
		var criteria = NewFilterDef()
		err = criteria.AddMetadataFilter("codename", "straight", "STRING", false, true)
		check.Assert(err, IsNil)
		queryItems, _, err := client.Client.SearchByFilter(QtOrgVdcNetwork, criteria)
		check.Assert(err, IsNil)
		fmt.Printf("\nNETWORK  found %d\n", len(queryItems))
		for _, item := range queryItems {
			fmt.Printf("%s %s\n", item.GetName(), item.GetIp())
		}
	*/

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

	filters, err := makeFiltersFromEdgeGateways(vdc)
	check.Assert(err, IsNil)
	check.Assert(filters, NotNil)

	for _, criteria := range filters {
		queryItems, explanation, err := client.Client.SearchByFilter(QtEdgeGateway, criteria)
		check.Assert(err, IsNil)
		check.Assert(len(queryItems), Equals, 1)
		fmt.Printf("%s\n", explanation)
		for i, item := range queryItems {
			fmt.Printf("( I) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
		}
	}

}
func makeFiltersFromCatalogs(org *AdminOrg) ([]*FilterDef, error) {
	if org.AdminOrg.Catalogs == nil || len(org.AdminOrg.Catalogs.Catalog) == 0 {
		return []*FilterDef{}, nil
	}
	var filters []*FilterDef
	for _, ref := range org.AdminOrg.Catalogs.Catalog {
		cat, err := org.GetCatalogByName(ref.Name, false)
		if err != nil {
			return []*FilterDef{}, err
		}
		filter := NewFilterDef()
		_ = filter.AddFilter(FilterNameRegex, cat.Catalog.Name)
		_ = filter.AddFilter(FilterDate, "=="+cat.Catalog.DateCreated)

		// TODO: add metadata
		filters = append(filters, filter)
	}
	return filters, nil
}

func (vcd *TestVCD) Test_SearchCatalog(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("no org provided. Skipping test")
	}
	client := vcd.client
	// Fetching organization and VDC
	org, err := client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	filters, err := makeFiltersFromCatalogs(org)
	check.Assert(err, IsNil)
	check.Assert(filters, NotNil)

	for _, criteria := range filters {
		queryItems, explanation, err := client.Client.SearchByFilter(QtAdminCatalog, criteria)
		check.Assert(err, IsNil)
		check.Assert(len(queryItems), Equals, 1)
		fmt.Printf("%s\n", explanation)
		for i, item := range queryItems {
			fmt.Printf("( I) %2d %-10s %-20s %s\n\n", i, item.GetType(), item.GetName(), item.GetIp())
		}
	}

}

// TODO: make test function for media items search
/*
	criteria = NewFilterDef()
	//err = criteria.AddFilter("name_regex", "^test")
	err = criteria.AddFilter("date", "> 2020-03-08")
	//err = criteria.AddMetadataFilter("one", "abc", "", false, false )
	check.Assert(err, IsNil)
	queryType = QtMedia
	if catalog.client.IsSysAdmin {
		queryType = QtAdminMedia
	}
	queryItems, _, err = client.SearchByFilter(queryType, criteria)
	check.Assert(err, IsNil)
	fmt.Printf("\nmedia found %d\n", len(queryItems))
	for _, item := range queryItems {
		fmt.Printf("%s %s\n", item.GetName(), item.GetHref())
	}
	media, ok := queryItems[0].(QueryMedia)
	check.Assert(ok, Equals, true)
	check.Assert(media.Name, Not(Equals), "")
	// fmt.Printf("%# v\n", pretty.Formatter(media))

*/

// createMultipleCatalogItems deploys several catalog items, as defined in requestData
// Returns a set of vappTemplateData with what was created.
// If the requested objects exist already, returns updated information about the existing items.
func createMultipleCatalogItems(orgName string, catalog *Catalog, requestData []vappTemplateData) ([]vappTemplateData, error) {
	var data []vappTemplateData
	ova := "../test-resources/test_vapp_template.ova"
	_, err := os.Stat(ova)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("sample OVA %s not found", ova)
	}
	overallStart := time.Now()
	for _, requested := range requestData {
		name := requested.name

		var item *CatalogItem
		var vappTemplate VAppTemplate
		created := false
		item, err := catalog.GetCatalogItemByName(name, false)
		if err == nil {
			// If the item already exists, we skip the creation, and just retrieve the vapp template
			vappTemplate, err = item.GetVAppTemplate()
			if err != nil {
				return nil, err
			}
		} else {

			start := time.Now()
			fmt.Printf("%-55s %s ", start, name)
			task, err := catalog.UploadOvf(ova, name, "test "+name, 10)
			if err != nil {
				return nil, err
			}
			if os.Getenv("GOVCD_KEEP_TEST_OBJECTS") == "" {
				AddToCleanupList(name, "catalogItem", orgName+"|"+catalog.Catalog.Name, "createMultipleCatalogItems")
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return nil, err
			}
			item, err = catalog.GetCatalogItemByName(name, true)
			if err != nil {
				return nil, err
			}
			vappTemplate, err = item.GetVAppTemplate()
			if err != nil {
				return nil, err
			}

			for k, v := range requested.metadata {
				_, err := vappTemplate.AddMetadata(k, v)
				if err != nil {
					return nil, err
				}
			}
			duration := time.Since(start)
			fmt.Printf("- elapsed: %s\n", duration)

			created = true
		}
		data = append(data, vappTemplateData{
			name:                     name,
			itemCreationDate:         item.CatalogItem.DateCreated,
			vappTemplateCreationDate: vappTemplate.VAppTemplate.DateCreated,
			metadata:                 requested.metadata,
			created:                  created,
		})
	}
	overallDuration := time.Since(overallStart)
	fmt.Printf("total elapsed: %s\n", overallDuration)

	return data, nil
}
