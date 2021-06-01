// +build functional openapi role ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GlobalRoles(check *C) {
	client := vcd.client.Client
	if !client.IsSysAdmin {
		check.Skip("test Test_GlobalRoles requires system administrator privileges")
	}

	// Step 1 - Get all global roles
	allExistingGlobalRoles, err := client.GetAllGlobalRoles(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingGlobalRoles, NotNil)

	// Step 2 - Get all roles using query filters
	for _, oneGlobalRole := range allExistingGlobalRoles {

		// Step 2.1 - retrieve specific global role by using FIQL filter
		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+oneGlobalRole.GlobalRole.Id)

		expectOneGlobalRoleResultById, err := client.GetAllGlobalRoles(queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOneGlobalRoleResultById) == 1, Equals, true)

		// Step 2.2 - retrieve specific global role by using endpoint
		exactItem, err := client.GetGlobalRoleById(oneGlobalRole.GlobalRole.Id)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		check.Assert(exactItem, NotNil)

		// Step 2.3 - compare struct retrieved by using filter and the one retrieved by exact endpoint ID
		check.Assert(oneGlobalRole, DeepEquals, expectOneGlobalRoleResultById[0])

	}

	// Step 3 - Create a new global role and ensure it is created as specified by doing deep comparison

	newGR := &types.GlobalRole{
		Name:        check.TestName(),
		Description: "Global Role created by test",
		// This BundleKey is being set by VCD even if it is not sent
		BundleKey: "com.vmware.vcloud.undefined.key",
		ReadOnly:  false,
	}

	createdGlobalRole, err := client.CreateGlobalRole(newGR)
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(createdGlobalRole.GlobalRole.Name, check.TestName(),
		types.OpenApiPathVersion1_0_0+types.OpenApiEndpointGlobalRoles+createdGlobalRole.GlobalRole.Id)

	// Ensure supplied and created structs differ only by ID
	newGR.Id = createdGlobalRole.GlobalRole.Id
	check.Assert(createdGlobalRole.GlobalRole, DeepEquals, newGR)

	// Step 4 - updated created global role
	createdGlobalRole.GlobalRole.Description = "Updated description"
	updatedGlobalRole, err := createdGlobalRole.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedGlobalRole.GlobalRole, DeepEquals, createdGlobalRole.GlobalRole)

	// Step 5 - add rights to global role

	// These rights include 5 implied rights, which will be added by globalRole.AddRights
	rightName1 := "Catalog: Add vApp from My Cloud"
	rightName2 := "Catalog: Edit Properties"

	right1, err := client.GetRightByName(rightName1)
	check.Assert(err, IsNil)
	right2, err := client.GetRightByName(rightName2)
	check.Assert(err, IsNil)
	err = updatedGlobalRole.AddRights([]types.OpenApiReference{
		{Name: rightName1, ID: right1.ID},
		{Name: rightName2, ID: right2.ID},
	})
	check.Assert(err, IsNil)

	// Calculate the total amount of rights we should expect to be added to the global role
	var unique = make(map[string]bool)
	for _, r := range []*types.Right{right1, right2} {
		_, seen := unique[r.ID]
		if !seen {
			unique[r.ID] = true
		}
		for _, implied := range r.ImpliedRights {
			_, seen := unique[implied.ID]
			if !seen {
				unique[implied.ID] = true
			}
		}
	}
	rights, err := updatedGlobalRole.GetGlobalRoleRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, len(unique))

	// Step 6 - remove 1 right from global role

	err = updatedGlobalRole.RemoveRights([]types.OpenApiReference{{Name: right1.Name, ID: right1.ID}})
	check.Assert(err, IsNil)
	rights, err = updatedGlobalRole.GetGlobalRoleRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, len(unique)-1)

	// Step 7 - remove all rights from global role
	err = updatedGlobalRole.RemoveAllRights()
	check.Assert(err, IsNil)

	rights, err = updatedGlobalRole.GetGlobalRoleRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, 0)

	// Step 8 - delete created global role
	err = updatedGlobalRole.Delete()
	check.Assert(err, IsNil)

	// Step 9 - try to read deleted global role and expect error to contain 'ErrorEntityNotFound'
	// Read is tricky - it throws an error ACCESS_TO_RESOURCE_IS_FORBIDDEN when the resource with ID does not
	// exist therefore one cannot know what kind of error occurred.
	deletedGlobalRole, err := client.GetGlobalRoleById(createdGlobalRole.GlobalRole.Id)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedGlobalRole, IsNil)
}
