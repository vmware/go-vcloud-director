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

func (vcd *TestVCD) Test_RightsBundle(check *C) {
	client := vcd.client.Client
	if !client.IsSysAdmin {
		check.Skip("test Test_RightsBundle requires system administrator privileges")
	}

	// Step 1 - Get all rights bundles
	allExistingRightsBundle, err := client.GetAllRightsBundles(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingRightsBundle, NotNil)

	// Step 2 - Get all roles using query filters
	for _, oneRightsBundle := range allExistingRightsBundle {

		// Step 2.1 - retrieve specific rights bundle by using FIQL filter
		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+oneRightsBundle.RightsBundle.Id)

		expectOneRightsBundleResultById, err := client.GetAllRightsBundles(queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOneRightsBundleResultById) == 1, Equals, true)

		// Step 2.2 - retrieve specific rights bundle by using endpoint
		exactItem, err := client.GetRightsBundleById(oneRightsBundle.RightsBundle.Id)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		check.Assert(exactItem, NotNil)

		// Step 2.3 - compare struct retrieved by using filter and the one retrieved by exact endpoint ID
		check.Assert(oneRightsBundle, DeepEquals, expectOneRightsBundleResultById[0])

	}

	// Step 3 - Create a new rights bundle and ensure it is created as specified by doing deep comparison

	newGR := &types.RightsBundle{
		Name:        check.TestName(),
		Description: "Global Role created by test",
		// This BundleKey is being set by VCD even if it is not sent
		BundleKey: "com.vmware.vcloud.undefined.key",
		ReadOnly:  false,
	}

	createdRightsBundle, err := client.CreateRightsBundle(newGR)
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(createdRightsBundle.RightsBundle.Name, check.TestName(),
		types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRightsBundles+createdRightsBundle.RightsBundle.Id)

	// Ensure supplied and created structs differ only by ID
	newGR.Id = createdRightsBundle.RightsBundle.Id
	check.Assert(createdRightsBundle.RightsBundle, DeepEquals, newGR)

	// Step 4 - updated created rights bundle
	createdRightsBundle.RightsBundle.Description = "Updated description"
	updatedRightsBundle, err := createdRightsBundle.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedRightsBundle.RightsBundle, DeepEquals, createdRightsBundle.RightsBundle)

	// Step 5 - add rights to rights bundle

	// These rights include 5 implied rights, which will be added by globalRole.AddRights
	rightName1 := "Catalog: Add vApp from My Cloud"
	rightName2 := "Catalog: Edit Properties"

	right1, err := client.GetRightByName(rightName1)
	check.Assert(err, IsNil)
	right2, err := client.GetRightByName(rightName2)
	check.Assert(err, IsNil)
	err = updatedRightsBundle.AddRights([]types.OpenApiReference{
		{Name: rightName1, ID: right1.ID},
		{Name: rightName2, ID: right2.ID},
	})
	check.Assert(err, IsNil)

	// Calculate the total amount of rights we should expect to be added to the rights bundle
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
	rights, err := updatedRightsBundle.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, len(unique))

	// Step 6 - remove 1 right from rights bundle

	err = updatedRightsBundle.RemoveRights([]types.OpenApiReference{{Name: right1.Name, ID: right1.ID}})
	check.Assert(err, IsNil)
	rights, err = updatedRightsBundle.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, len(unique)-1)

	testRightsContainerTenants(vcd, check, updatedRightsBundle)

	// Step 7 - remove all rights from rights bundle
	err = updatedRightsBundle.RemoveAllRights()
	check.Assert(err, IsNil)

	rights, err = updatedRightsBundle.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, 0)

	// Step 8 - delete created rights bundle
	err = updatedRightsBundle.Delete()
	check.Assert(err, IsNil)

	// Step 9 - try to read deleted rights bundle and expect error to contain 'ErrorEntityNotFound'
	// Read is tricky - it throws an error ACCESS_TO_RESOURCE_IS_FORBIDDEN when the resource with ID does not
	// exist therefore one cannot know what kind of error occurred.
	deletedRightsBundle, err := client.GetRightsBundleById(createdRightsBundle.RightsBundle.Id)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedRightsBundle, IsNil)
}
