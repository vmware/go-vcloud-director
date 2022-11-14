//go:build task || functional || ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"time"

	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_QueryTaskList(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	catalog, err := adminOrg.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	adminCatalog, err := adminOrg.GetAdminCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	startQuery := time.Now()
	allTasks, err := vcd.client.Client.QueryTaskList(map[string]string{
		"status": "running,preRunning,queued",
	})
	check.Assert(err, IsNil)
	if testVerbose {
		fmt.Printf("%# v\n%s\n", pretty.Formatter(allTasks), time.Since(startQuery))
	}
	// search using a client, giving the org and catalog names
	resultByClient, err := vcd.client.Client.QueryTaskList(map[string]string{
		"orgName":    vcd.config.VCD.Org,
		"objectName": adminCatalog.AdminCatalog.Name,
		"name":       "catalogCreateCatalog"})
	check.Assert(err, IsNil)

	// search using an admin catalog, which will search by its HREF
	resultByAdminCatalog, err := adminCatalog.QueryTaskList(map[string]string{
		"name": "catalogCreateCatalog",
	})
	check.Assert(err, IsNil)
	// search using a catalog, which will search by its HREF
	resultByCatalog, err := catalog.QueryTaskList(map[string]string{
		"name": "catalogCreateCatalog",
	})
	check.Assert(err, IsNil)

	check.Assert(len(resultByClient), Equals, len(resultByAdminCatalog))
	check.Assert(len(resultByClient), Equals, len(resultByCatalog))
	if len(resultByAdminCatalog) > 0 {
		// there should be only one task for catalog creation
		check.Assert(len(resultByClient), Equals, 1)
		// check correspondence between task and its related object
		// and also that all sets have returned the same result
		catalogHref, err := adminCatalog.GetCatalogHref()
		check.Assert(err, IsNil)
		check.Assert(resultByAdminCatalog[0].HREF, Equals, resultByClient[0].HREF)
		check.Assert(resultByCatalog[0].HREF, Equals, resultByClient[0].HREF)
		check.Assert(resultByClient[0].Object, Equals, catalogHref)
		check.Assert(resultByAdminCatalog[0].ObjectName, Equals, adminCatalog.AdminCatalog.Name)
		check.Assert(resultByCatalog[0].ObjectName, Equals, adminCatalog.AdminCatalog.Name)
		check.Assert(resultByAdminCatalog[0].ObjectType, Equals, "catalog")
		check.Assert(resultByCatalog[0].ObjectType, Equals, "catalog")
	}
}

func init() {
	testingTags["task"] = "task_test.go"
}
