//go:build user || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

/*
  TODO: Add test for takeOwnership.

This is more complicated than it looks, because it requires the following:

Either:
1a. Separate connection with a newly created user [requires test enhancement]
2a. Creation of entities with new user (vapp/catalog/catalog items)

OR
1b. create entities with the user that runs the tests
2b. change ownership of such entities to the new user [requires new feature]

3. Check that the user is the intended one (this is currently doable, because we can
  inspect the Owner structure of the entity being created)

4. Try deleting the user that owns the new entities
5. get an error
6. take ownership from the user
7. delete the user and see the operation succeed
8. Check that the new entities belong to the current user
9. Delete the new entities
*/

// Checks that the default roles are available from the organization
func (vcd *TestVCD) Test_GetRoleReference(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	Roles := []string{
		OrgUserRoleOrganizationAdministrator,
		OrgUserRoleVappUser,
		OrgUserRoleCatalogAuthor,
		OrgUserRoleConsoleAccessOnly,
	}
	for _, roleName := range Roles {
		roleReference, err := adminOrg.GetRoleReference(roleName)
		check.Assert(err, IsNil)
		check.Assert(roleReference, NotNil)
		check.Assert(roleReference.Name, Equals, roleName)
		check.Assert(roleReference.HREF, Not(Equals), "")
	}
}

// Checks that we can retrieve a user by name or ID
func (vcd *TestVCD) Test_GetUserByNameOrId(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	// We get the list of users from the organization
	var userRefs []types.Reference
	for _, userRef := range adminOrg.AdminOrg.Users.User {
		userRefs = append(userRefs, *userRef)
	}

	// Using the list above, we first try to get each user by name
	for _, userRef := range userRefs {
		user, err := adminOrg.GetUserByName(userRef.Name, false)
		check.Assert(err, IsNil)
		check.Assert(user, NotNil)
		check.Assert(user.User.Name, Equals, userRef.Name)

		// Then we try to get the same user by ID
		user, err = adminOrg.GetUserById(userRef.ID, false)
		check.Assert(err, IsNil)
		check.Assert(user, NotNil)
		check.Assert(user.User.Name, Equals, userRef.Name)

		// Then we try to get the same user by Name or ID combined
		user, err = adminOrg.GetUserByNameOrId(userRef.ID, true)
		check.Assert(err, IsNil)
		check.Assert(user, NotNil)
		check.Assert(user.User.Name, Equals, userRef.Name)

		user, err = adminOrg.GetUserByNameOrId(userRef.Name, false)
		check.Assert(err, IsNil)
		check.Assert(user, NotNil)
		check.Assert(user.User.Name, Equals, userRef.Name)
	}
}

// This test creates 5 users using 5 available roles,
// Then updates each of them with a different role,
// Furthermore, disables, and then enables the users again
// and finally deletes all of them
func (vcd *TestVCD) Test_UserCRUD(check *C) {
	vcd.checkSkipWhenApiToken(check)
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	type userTestData struct {
		name       string // name of the user. Note: only lowercase letters allowed
		roleName   string // the role this user is created with
		secondRole string // The role to which we change using Update()
	}
	userData := []userTestData{
		{
			name:       "test_user_admin",
			roleName:   OrgUserRoleOrganizationAdministrator,
			secondRole: OrgUserRoleVappAuthor,
		},
		{
			name:       "test_user_vapp_author",
			roleName:   OrgUserRoleVappAuthor,
			secondRole: OrgUserRoleVappUser,
		},
		{
			name:       "test_user_vapp_user",
			roleName:   OrgUserRoleVappUser,
			secondRole: OrgUserRoleConsoleAccessOnly,
		},
		{
			name:       "test_user_console_access",
			roleName:   OrgUserRoleConsoleAccessOnly,
			secondRole: OrgUserRoleCatalogAuthor,
		},
		{
			name:       "test_user_catalog_author",
			roleName:   OrgUserRoleCatalogAuthor,
			secondRole: OrgUserRoleOrganizationAdministrator,
		},
	}

	quotaDeployed := 10
	quotaStored := 10
	for _, ud := range userData {
		quotaDeployed += 2
		quotaStored += 2
		fmt.Printf("# Creating user %s with role %s\n", ud.name, ud.roleName)
		// Uncomment the following lines to see creation request and response
		// enableDebugShowRequest()
		// enableDebugShowResponse()
		var userDefinition = OrgUserConfiguration{
			Name:            ud.name,
			Password:        "user_pass",
			RoleName:        ud.roleName,
			ProviderType:    OrgUserProviderIntegrated,
			DeployedVmQuota: quotaDeployed,
			StoredVmQuota:   quotaStored,
			FullName:        strings.ReplaceAll(ud.name, "_", " "),
			Description:     "user " + strings.ReplaceAll(ud.name, "_", " "),
			IsEnabled:       true,
			IsExternal:      false,
			IM:              "TextIM",
			EmailAddress:    "somename@somedomain.com",
			Telephone:       "999 888-7777",
		}

		user, err := adminOrg.CreateUserSimple(userDefinition)
		// disableDebugShowRequest()
		// disableDebugShowResponse()
		check.Assert(err, IsNil)

		AddToCleanupList(ud.name, "user", user.AdminOrg.AdminOrg.Name, check.TestName())
		check.Assert(user.User, NotNil)
		check.Assert(user.User.Name, Equals, ud.name)
		check.Assert(user.GetRoleName(), Equals, ud.roleName)
		check.Assert(user.User.IsEnabled, Equals, true)
		check.Assert(user.User.FullName, Equals, userDefinition.FullName)
		check.Assert(user.User.EmailAddress, Equals, userDefinition.EmailAddress)
		check.Assert(user.User.IM, Equals, userDefinition.IM)
		check.Assert(user.User.Telephone, Equals, userDefinition.Telephone)
		check.Assert(user.User.StoredVmQuota, Equals, userDefinition.StoredVmQuota)
		check.Assert(user.User.DeployedVmQuota, Equals, userDefinition.DeployedVmQuota)
		check.Assert(user.User.IsExternal, Equals, userDefinition.IsExternal)

		// change DeployedVmQuota and StoredVmQuota to 0 and assert
		// this will make DeployedVmQuota and StoredVmQuota unlimited
		user.User.DeployedVmQuota = 0
		user.User.StoredVmQuota = 0
		err = user.Update()
		check.Assert(err, IsNil)

		// Get the user from API again
		user, err = adminOrg.GetUserByHref(user.User.Href)
		check.Assert(err, IsNil)
		check.Assert(user.User.DeployedVmQuota, Equals, 0)
		check.Assert(user.User.StoredVmQuota, Equals, 0)

		err = user.Disable()
		check.Assert(err, IsNil)
		check.Assert(user.User.IsEnabled, Equals, false)

		fmt.Printf("# Updating user %s with role %s\n", ud.name, ud.secondRole)
		err = user.ChangeRole(ud.secondRole)
		check.Assert(err, IsNil)
		check.Assert(user.GetRoleName(), Equals, ud.secondRole)

		err = user.Enable()
		check.Assert(err, IsNil)
		check.Assert(user.User.IsEnabled, Equals, true)
		err = user.ChangePassword("new_pass")
		check.Assert(err, IsNil)
	}

	var enableMap = map[bool]string{
		true:  "enabled",
		false: "disabled",
	}
	for _, ud := range userData {
		user, err := adminOrg.GetUserByNameOrId(ud.name, true)
		check.Assert(err, IsNil)

		fmt.Printf("# deleting user %s (%s - %s)\n", ud.name, user.GetRoleName(), enableMap[user.User.IsEnabled])
		// uncomment the following two lines to see the deletion request and response
		// enableDebugShowRequest()
		// enableDebugShowResponse()
		err = user.Delete(true)
		// disableDebugShowRequest()
		// disableDebugShowResponse()
		check.Assert(err, IsNil)
		user, err = adminOrg.GetUserByNameOrId(user.User.ID, true)
		check.Assert(err, NotNil)
		// Tests both the error directly and the function IsNotFound
		check.Assert(err, Equals, ErrorEntityNotFound)
		check.Assert(IsNotFound(err), Equals, true)
		// Expect a null pointer when user is not found
		check.Assert(user, IsNil)
	}
}

func init() {
	testingTags["user"] = "user_test.go"
}
