//go:build functional || openapi || role || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (vcd *TestVCD) Test_Roles(check *C) {

	vcd.checkSkipWhenApiToken(check)
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	// Step 1 - Get all roles
	allExistingRoles, err := adminOrg.GetAllRoles(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingRoles, NotNil)

	// Step 2 - Get all roles using query filters
	for _, oneRole := range allExistingRoles {

		// Step 2.1 - retrieve specific role by using FIQL filter
		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+oneRole.Role.ID)

		expectOneRoleResultById, err := adminOrg.GetAllRoles(queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOneRoleResultById) == 1, Equals, true)

		// Step 2.2 - retrieve specific role by using endpoint
		exactItem, err := adminOrg.GetRoleById(oneRole.Role.ID)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		check.Assert(exactItem, NotNil)

		// Step 2.3 - compare struct retrieved by using filter and the one retrieved by exact endpoint ID
		check.Assert(oneRole, DeepEquals, expectOneRoleResultById[0])

	}

	// Step 3 - Create a new role and ensure it is created as specified by doing deep comparison

	newR := &types.Role{
		Name:        check.TestName(),
		Description: "Role created by test",
		// This BundleKey is being set by VCD even if it is not sent
		BundleKey: types.VcloudUndefinedKey,
		ReadOnly:  false,
	}

	createdRole, err := adminOrg.CreateRole(newR)
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(createdRole.Role.Name, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRoles+createdRole.Role.ID)

	// Ensure supplied and created structs differ only by ID
	newR.ID = createdRole.Role.ID
	check.Assert(createdRole.Role, DeepEquals, newR)

	// Check that the new role is found in the Organization structure
	roleRef, err := adminOrg.GetRoleReference(createdRole.Role.Name)
	check.Assert(err, IsNil)
	check.Assert(roleRef, NotNil)

	// Step 4 - updated created role
	createdRole.Role.Description = "Updated description"
	updatedRole, err := createdRole.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedRole.Role, DeepEquals, createdRole.Role)

	// Step 5 - add rights to role

	// These rights include 5 implied rights, which will be added by role.AddRights
	rightNames := []string{"Catalog: Add vApp from My Cloud", "Catalog: Edit Properties"}

	rightSet, err := getRightsSet(adminOrg.client, rightNames)
	check.Assert(err, IsNil)

	err = updatedRole.AddRights(rightSet)
	check.Assert(err, IsNil)

	rights, err := updatedRole.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, len(rightSet))

	// Step 6 - remove 1 right from role

	err = updatedRole.RemoveRights([]types.OpenApiReference{rightSet[0]})
	check.Assert(err, IsNil)
	rights, err = updatedRole.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, len(rightSet)-1)

	// Step 7 - remove all rights from role
	err = updatedRole.RemoveAllRights()
	check.Assert(err, IsNil)

	rights, err = updatedRole.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, 0)

	// Step 8 - delete created role
	err = updatedRole.Delete()
	check.Assert(err, IsNil)

	// Step 9 - try to read deleted role and expect error to contain 'ErrorEntityNotFound'
	// Read is tricky - it throws an error ACCESS_TO_RESOURCE_IS_FORBIDDEN when the resource with ID does not
	// exist therefore one cannot know what kind of error occurred.
	deletedRole, err := adminOrg.GetRoleById(createdRole.Role.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedRole, IsNil)
}
