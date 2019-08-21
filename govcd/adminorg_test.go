// +build org functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
)

// Creates a Catalog and then verify that finds it
func (vcd *TestVCD) Test_FindCatalogRecordTypes(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	catalogName := "catalogForQuery"
	adminCatalog, err := adminOrg.CreateCatalog(catalogName, "catalogForQueryDescription")
	check.Assert(err, IsNil)
	AddToCleanupList(catalogName, "catalog", vcd.config.VCD.Org, check.TestName())
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, catalogName)

	// just imitate wait
	err = adminOrg.Refresh()
	check.Assert(err, IsNil)

	findRecords, err := adminOrg.FindCatalogRecordTypes(catalogName)
	check.Assert(err, IsNil)
	check.Assert(findRecords, NotNil)
	check.Assert(len(findRecords), Equals, 1)
	check.Assert(findRecords[0].Name, Equals, catalogName)
	check.Assert(findRecords[0].OrgName, Equals, adminOrg.AdminOrg.Name)
}
