/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetVAppTemplate(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	if err != nil {
		check.Skip("Catalog not found. Test can't proceed")
	}

	catitem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.Catalogitem)
	check.Assert(err, IsNil)

	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()

	check.Assert(err, IsNil)
	check.Assert(vapptemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.Catalogitem)
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(vapptemplate.VAppTemplate.Description, Equals, vcd.config.VCD.Catalog.CatalogItemDescription)
	}
}
