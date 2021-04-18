// +build vapp functional ALL

/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"

	. "gopkg.in/check.v1"
)

// TODO: Write test for InstantiateVAppTemplate

func (vcd *TestVCD) Test_RefreshVAppTemplate(check *C) {
	ctx := context.Background()
	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_GetVAppTemplate: Catalog not found. Test can't proceed")
		return
	}
	check.Assert(cat, NotNil)

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_GetVAppTemplate: Catalog Item not given. Test can't proceed")
	}

	catItem, err := cat.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	// Get VAppTemplate
	vAppTemplate, err := catItem.GetVAppTemplate(ctx)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	oldVAppTemplate := vAppTemplate

	err = vAppTemplate.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(oldVAppTemplate.VAppTemplate.ID, Equals, vAppTemplate.VAppTemplate.ID)
	check.Assert(oldVAppTemplate.VAppTemplate.Name, Equals, vAppTemplate.VAppTemplate.Name)
	check.Assert(oldVAppTemplate.VAppTemplate.HREF, Equals, vAppTemplate.VAppTemplate.HREF)
}
