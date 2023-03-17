//go:build functional || openapi || role || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_RoleTenantContext(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	clientRoles, err := adminOrg.client.GetAllRoles(nil)
	check.Assert(err, IsNil)
	check.Assert(clientRoles, NotNil)
	check.Assert(len(clientRoles), Not(Equals), 0)

	if testVerbose {
		fmt.Println("Client roles")
		for _, role := range clientRoles {
			fmt.Printf("%-40s %s\n", role.Role.Name, role.Role.Description)
			rights, err := role.GetRights(nil)
			check.Assert(err, IsNil)
			for i, right := range rights {
				fmt.Printf("\t%3d %s\n", i+1, right.Name)
			}
			fmt.Println()
		}
	}

	orgRoles, err := adminOrg.GetAllRoles(nil)
	check.Assert(err, IsNil)
	check.Assert(orgRoles, NotNil)
	check.Assert(len(orgRoles), Not(Equals), 0)

	if testVerbose {
		fmt.Println("ORG roles")
		for _, role := range orgRoles {
			fmt.Printf("%-40s %s\n", role.Role.Name, role.Role.Description)
			rights, err := role.GetRights(nil)
			check.Assert(err, IsNil)
			for i, right := range rights {
				fmt.Printf("\t%3d %s\n", i+1, right.Name)
			}
			fmt.Println()
		}
	}

}

func (vcd *TestVCD) Test_Rights(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	allOrgRights, err := adminOrg.GetAllRights(nil)
	check.Assert(err, IsNil)
	check.Assert(allOrgRights, NotNil)

	if testVerbose {
		fmt.Printf("(org) how many rights: %d\n", len(allOrgRights))
		for i, oneRight := range allOrgRights {
			fmt.Printf("%3d %-20s %-53s %s\n", i, oneRight.Name, oneRight.ID, oneRight.Category)
		}
	}
	allExistingRights, err := adminOrg.client.GetAllRights(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingRights, NotNil)

	if testVerbose {
		fmt.Printf("(global) how many rights: %d\n", len(allExistingRights))
		for i, oneRight := range allExistingRights {
			fmt.Printf("%3d %-20s %-53s %s\n", i, oneRight.Name, oneRight.ID, oneRight.Category)
		}
	}
	// Test a sample of rights
	maxItems := 10

	for i, right := range allExistingRights {
		searchRight(adminOrg.client, right.Name, right.ID, check)
		if i > maxItems {
			break
		}
	}
	var rigthNamesWithCommas = []string{
		"vApp: VM Migrate, Force Undeploy, Relocate, Consolidate",
		"Role: Create, Edit, Delete, or Copy",
		"Task: Resume, Abort, or Fail",
		"VCD Extension: Register, Unregister, Refresh, Associate or Disassociate",
		"UI Plugins: Define, Upload, Modify, Delete, Associate or Disassociate",
	}
	for _, name := range rigthNamesWithCommas {
		searchRight(adminOrg.client, name, "", check)
	}

	rightsCategories, err := adminOrg.client.GetAllRightsCategories(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rightsCategories) > 0, Equals, true)
}

func searchRight(client *Client, name, id string, check *C) {
	fullRightByName, err := client.GetRightByName(name)
	check.Assert(err, IsNil)
	check.Assert(fullRightByName, NotNil)
	if id != "" {
		fullRightById, err := client.GetRightById(id)
		check.Assert(err, IsNil)
		check.Assert(fullRightById, NotNil)
		category, err := client.GetRightsCategoryById(fullRightById.Category)
		check.Assert(err, IsNil)
		check.Assert(fullRightById.Category, Equals, category.Id)
	}
}
