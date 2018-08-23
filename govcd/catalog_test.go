/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_FindCatalogItem(check *C) {
	// Fetch Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	// Find Catalog Item
	catitem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.Catalogitem)
	check.Assert(err, IsNil)
	check.Assert(catitem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.Catalogitem)
	// If given a description in config file then it checks if the descriptions match
	// Otherwise it skips the assert
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(catitem.CatalogItem.Description, Equals, vcd.config.VCD.Catalog.CatalogItemDescription)
	}
	// Test non-existant catalog item
	catitem, err = cat.FindCatalogItem("INVALID")
	check.Assert(err, NotNil)
}
