// +build functional openapi role ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TenantContext(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	clientRoles, err := adminOrg.client.GetAllRoles(nil)
	check.Assert(err, IsNil)
	check.Assert(clientRoles, NotNil)
	check.Assert(len(clientRoles), Not(Equals), 0)

	fmt.Println("Client roles")
	for _, role := range clientRoles {
		fmt.Printf("%-40s %s\n", role.Role.Name, role.Role.Description)
		rights, err := role.GetRoleRights(nil)
		check.Assert(err, IsNil)
		for i, right := range rights {
			fmt.Printf("\t%3d %s\n", i+1, right.Name)
		}
		fmt.Println()
	}

	orgRoles, err := adminOrg.GetAllRoles(nil)
	check.Assert(err, IsNil)
	check.Assert(orgRoles, NotNil)
	check.Assert(len(orgRoles), Not(Equals), 0)

	fmt.Println("ORG roles")
	for _, role := range orgRoles {
		fmt.Printf("%-40s %s\n", role.Role.Name, role.Role.Description)
		rights, err := role.GetRoleRights(nil)
		check.Assert(err, IsNil)
		for i, right := range rights {
			fmt.Printf("\t%3d %s\n", i+1, right.Name)
		}
		fmt.Println()
	}

}

func (vcd *TestVCD) Test_Rights(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	allExistingRights, err := adminOrg.GetAllRights(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingRights, NotNil)

	fmt.Printf("(org) how many rights: %d\n", len(allExistingRights))

	for i, oneRight := range allExistingRights {
		fmt.Printf("%3d %-20s %-53s %s\n", i, oneRight.Name, oneRight.ID, oneRight.Category)
	}
	allExistingRights, err = adminOrg.client.GetAllRights(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingRights, NotNil)

	fmt.Printf("(global) how many rights: %d\n", len(allExistingRights))

	for i, oneRight := range allExistingRights {
		fmt.Printf("%3d %-20s %-53s %s\n", i, oneRight.Name, oneRight.ID, oneRight.Category)
	}

}
