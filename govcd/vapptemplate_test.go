// +build vapp functional ALL

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
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	if err != nil {
		check.Skip("Test_GetVAppTemplate: Catalog not found. Test can't proceed")
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_GetVAppTemplate: Catalog Item not given. Test can't proceed")
	}

	catItem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.CatalogItem)
	check.Assert(err, IsNil)

	// Get VAppTemplate
	vAppTemplate, err := catItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	oldVAppTemplate := vAppTemplate

	err = vAppTemplate.Refresh()
	check.Assert(oldVAppTemplate.VAppTemplate, Equals, vAppTemplate.VAppTemplate)
}
