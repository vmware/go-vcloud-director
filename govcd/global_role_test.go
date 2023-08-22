//go:build functional || openapi || role || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

type rightsProviderCollection interface {
	PublishAllTenants() error
	UnpublishAllTenants() error
	PublishTenants([]types.OpenApiReference) error
	UnpublishTenants([]types.OpenApiReference) error
	GetTenants(queryParameters url.Values) ([]types.OpenApiReference, error)
	ReplacePublishedTenants([]types.OpenApiReference) error
}

func (vcd *TestVCD) Test_GlobalRoles(check *C) {
	client := vcd.client.Client
	if !client.IsSysAdmin {
		check.Skip("test Test_GlobalRoles requires system administrator privileges")
	}
	vcd.checkSkipWhenApiToken(check)

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
		BundleKey: types.VcloudUndefinedKey,
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

	// These rights include 5 implied rights
	rightNames := []string{
		"Catalog: Add vApp from My Cloud",
		"Catalog: Edit Properties",
	}
	// Add an intentional duplicate to test the validity of getRightsSet and FindMissingImpliedRights
	rightNames = append(rightNames, rightNames[1])

	rightSet, err := getRightsSet(&client, rightNames)
	check.Assert(err, IsNil)

	err = updatedGlobalRole.AddRights(rightSet)
	check.Assert(err, IsNil)

	// Calculate the total amount of rights we should expect to be added to the global role
	rights, err := updatedGlobalRole.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, len(rightSet))

	// Step 6 - remove 1 right from global role

	err = updatedGlobalRole.RemoveRights([]types.OpenApiReference{rightSet[0]})
	check.Assert(err, IsNil)
	rights, err = updatedGlobalRole.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Equals, len(rightSet)-1)

	testRightsContainerTenants(vcd, check, updatedGlobalRole)

	// Step 7 - remove all rights from global role
	err = updatedGlobalRole.RemoveAllRights()
	check.Assert(err, IsNil)

	rights, err = updatedGlobalRole.GetRights(nil)
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

func foundOrg(name, id string, items []types.OpenApiReference) bool {
	for _, item := range items {
		if item.ID == id && item.Name == name {
			return true
		}
	}
	return false
}

// testRightsContainerTenants is a sub-test that checks the validity of the tenants
// registered to the container
func testRightsContainerTenants(vcd *TestVCD, check *C, rpc rightsProviderCollection) {

	newOrgName := check.TestName() + "-org"
	task, err := CreateOrg(vcd.client, newOrgName, newOrgName, newOrgName, &types.OrgSettings{}, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	newOrg, err := vcd.client.GetAdminOrgByName(newOrgName)
	check.Assert(err, IsNil)
	AddToCleanupList(newOrgName, "org", "", "testRightsContainerTenants")

	err = rpc.PublishTenants([]types.OpenApiReference{
		{ID: vcd.org.Org.ID, Name: vcd.org.Org.Name},
		{ID: newOrg.AdminOrg.ID, Name: newOrg.AdminOrg.Name},
	})
	check.Assert(err, IsNil)

	tenants, err := rpc.GetTenants(nil)
	check.Assert(err, IsNil)
	check.Assert(len(tenants), Equals, 2)

	check.Assert(foundOrg(vcd.org.Org.Name, vcd.org.Org.ID, tenants), Equals, true)
	check.Assert(foundOrg(newOrg.AdminOrg.Name, newOrg.AdminOrg.ID, tenants), Equals, true)

	err = rpc.UnpublishTenants(tenants)
	check.Assert(err, IsNil)
	tenants, err = rpc.GetTenants(nil)
	check.Assert(err, IsNil)
	check.Assert(len(tenants), Equals, 0)

	err = rpc.PublishTenants([]types.OpenApiReference{
		{ID: vcd.org.Org.ID, Name: vcd.org.Org.Name},
	})
	check.Assert(err, IsNil)

	tenants, err = rpc.GetTenants(nil)
	check.Assert(err, IsNil)
	check.Assert(len(tenants), Equals, 1)

	check.Assert(foundOrg(vcd.org.Org.Name, vcd.org.Org.ID, tenants), Equals, true)

	err = rpc.ReplacePublishedTenants([]types.OpenApiReference{
		{ID: vcd.org.Org.ID, Name: vcd.org.Org.Name},
		{ID: newOrg.AdminOrg.ID, Name: newOrg.AdminOrg.Name},
	})
	check.Assert(err, IsNil)
	tenants, err = rpc.GetTenants(nil)
	check.Assert(err, IsNil)
	check.Assert(len(tenants), Equals, 2)

	check.Assert(foundOrg(vcd.org.Org.Name, vcd.org.Org.ID, tenants), Equals, true)
	check.Assert(foundOrg(newOrg.AdminOrg.Name, newOrg.AdminOrg.ID, tenants), Equals, true)

	err = rpc.UnpublishTenants(tenants)
	check.Assert(err, IsNil)
	tenants, err = rpc.GetTenants(nil)
	check.Assert(err, IsNil)
	check.Assert(len(tenants), Equals, 0)

	err = rpc.PublishAllTenants()
	check.Assert(err, IsNil)

	tenants, err = rpc.GetTenants(nil)
	check.Assert(err, IsNil)
	check.Assert(len(tenants), Not(Equals), 0)

	check.Assert(foundOrg(vcd.org.Org.Name, vcd.org.Org.ID, tenants), Equals, true)
	check.Assert(foundOrg(newOrg.AdminOrg.Name, newOrg.AdminOrg.ID, tenants), Equals, true)

	err = rpc.UnpublishAllTenants()
	check.Assert(err, IsNil)
	tenants, err = rpc.GetTenants(nil)
	check.Assert(err, IsNil)
	check.Assert(len(tenants), Equals, 0)
	err = newOrg.Delete(true, true)
	check.Assert(err, IsNil)
}

// getRightsSet is a convenience function that retrieves a list of rights
// from a list of right names, and adds the implied rights
func getRightsSet(client *Client, rightNames []string) ([]types.OpenApiReference, error) {
	var rightList []types.OpenApiReference
	var uniqueNames = make(map[string]bool)

	for _, name := range rightNames {
		_, seen := uniqueNames[name]
		if seen {
			continue
		}
		right, err := client.GetRightByName(name)
		if err != nil {
			return nil, err
		}
		rightList = append(rightList, types.OpenApiReference{
			Name: right.Name,
			ID:   right.ID,
		})
		uniqueNames[name] = true
	}
	implied, err := FindMissingImpliedRights(client, rightList)
	if err != nil {
		return nil, err
	}
	for _, ir := range implied {
		_, seen := uniqueNames[ir.Name]
		if seen {
			continue
		}
		rightList = append(rightList, ir)
	}
	return rightList, nil
}
