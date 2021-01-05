// +build functional openapi ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/kr/pretty"
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (vcd *TestVCD) Test_Roles(check *C) {
	//adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	fmt.Printf("%s\n",adminOrg.AdminOrg.Name)
	//orgId, err := GetUuidFromHref(adminOrg.AdminOrg.HREF, true)
	//check.Assert(err, IsNil)
	//adminOrg.client.SetCustomHeader(map[string]string{types.HeaderTenantContext: orgId, types.HeaderAuthContext: adminOrg.AdminOrg.Name})

	// Step 1 - Get all roles
	allExistingRoles, err := adminOrg.GetAllOpenApiRoles(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingRoles, NotNil)

	// Step 2 - Get all roles using query filters
	for _, oneRole := range allExistingRoles {
		fmt.Printf("%# v\n",pretty.Formatter(oneRole.Role))

		// Step 2.1 - retrieve specific role by using FIQL filter
		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+oneRole.Role.ID)

		expectOneRoleResultById, err := adminOrg.GetAllOpenApiRoles(queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOneRoleResultById) == 1, Equals, true)

		// Step 2.2 - retrieve specific role by using endpoint
		exactItem, err := adminOrg.GetOpenApiRoleById(oneRole.Role.ID)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		check.Assert(exactItem, NotNil)

		// Step 2.3 - compare struct retrieved by using filter and the one retrieved by exact endpoint ID
		check.Assert(oneRole, DeepEquals, expectOneRoleResultById[0])

	}

	// Step 3 - CreateRole a new role and ensure it is created as specified by doing deep comparison

	newR := &types.Role{
		Name:        check.TestName(),
		Description: "Role created by test",
		// This BundleKey is being set by VCD even if it is not sent
		BundleKey: "com.vmware.vcloud.undefined.key",
		ReadOnly:  false,
	}

	createdRole, err := adminOrg.CreateRole(newR)
	check.Assert(err, IsNil)

	// Ensure supplied and created structs differ only by ID
	newR.ID = createdRole.Role.ID
	check.Assert(createdRole.Role, DeepEquals, newR)

	// Step 4 - updated created role
	createdRole.Role.Description = "Updated description"
	updatedRole, err := createdRole.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedRole.Role, DeepEquals, createdRole.Role)

	// Step 5 - delete created role
	err = updatedRole.Delete()
	check.Assert(err, IsNil)
	// Step 5 - try to read deleted role and expect error to contain 'ErrorEntityNotFound'
	// Read is tricky - it throws an error ACCESS_TO_RESOURCE_IS_FORBIDDEN when the resource with ID does not
	// exist therefore one cannot know what kind of error occurred.
	deletedRole, err := adminOrg.GetOpenApiRoleById(createdRole.Role.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedRole, IsNil)
}
