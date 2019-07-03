// +build user functional ALL

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

// Checks that the default roles are available from the organization
func (vcd *TestVCD) Test_GetRole(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})
	Roles := []string{
		OrgUserRoleOrganizationAdministrator,
		OrgUserRoleVappUser,
		OrgUserRoleCatalogAuthor,
		OrgUserRoleConsoleAccessOnly,
	}
	for _, roleName := range Roles {
		// fmt.Printf("# retrieving role %s\n", roleName)
		roleReference, err := adminOrg.GetRole(roleName)
		check.Assert(err, IsNil)
		check.Assert(roleReference.Name, Equals, roleName)
		check.Assert(roleReference.HREF, Not(Equals), "")
	}
}

// Checks that we can retrieve an user by name or ID
func (vcd *TestVCD) Test_GetUserByName(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})

	// We get the list of users from the organization
	var userRefs []types.Reference
	for _, userRef := range adminOrg.AdminOrg.Users.User {
		userRefs = append(userRefs, *userRef)
	}

	// Using the list above, we first try to get each user by name
	for _, userRef := range userRefs {
		user, err := adminOrg.GetUserByNameOrId(userRef.Name, false)
		check.Assert(user, Not(Equals), OrgUser{})
		check.Assert(err, IsNil)
		check.Assert(user.User.Name, Equals, userRef.Name)

		// Then we try to get the same user by ID
		user, err = adminOrg.GetUserByNameOrId(userRef.ID, false)
		check.Assert(user, Not(Equals), OrgUser{})
		check.Assert(err, IsNil)
		check.Assert(user.User.Name, Equals, userRef.Name)
		// Uncomment this line to see the full user structure
		// ShowUser(*user.OrgUser)
	}
}

// This test creates 5 users using 5 available roles,
// Then updates each of them with a different role,
// Furthermore, disables, and then enables the users again
// and finally deletes all of them
func (vcd *TestVCD) Test_UserCRUD(check *C) {

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})

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

	for _, ud := range userData {
		fmt.Printf("# Creating user %s with role %s\n", ud.name, ud.roleName)
		// Uncomment the following lines to see creation request and response
		// enableDebugShowRequest()
		// enableDebugShowResponse()
		user, err := adminOrg.SimpleCreateUser(OrgUserConfiguration{
			Name:            ud.name,
			Password:        "user_pass",
			RoleName:        ud.roleName,
			ProviderType:    OrgUserProviderIntegrated,
			DeployedVmQuota: 10,
			StoredVmQuota:   10,
			FullName:        strings.ReplaceAll(ud.name, "_", " "),
			Description:     "user " + strings.ReplaceAll(ud.name, "_", " "),
			IsEnabled:       true,
		})
		// disableDebugShowRequest()
		// disableDebugShowResponse()
		check.Assert(err, IsNil)

		AddToCleanupList(ud.name, "user", user.adminOrg.AdminOrg.Name, check.TestName())
		check.Assert(user.User, NotNil)
		check.Assert(user.User.Name, Equals, ud.name)
		check.Assert(user.GetRoleName(), Equals, ud.roleName)
		check.Assert(user.User.IsEnabled, Equals, true)

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
		err = user.SafeDelete()
		// disableDebugShowRequest()
		// disableDebugShowResponse()
		check.Assert(err, IsNil)
		user, err = adminOrg.GetUserByNameOrId(user.User.ID, true)
		check.Assert(err, NotNil)
		// Tests both the error directly and the function IsNotFound
		check.Assert(err, Equals, ErrorEntityNotFound)
		check.Assert(IsNotFound(err), Equals, true)
		check.Assert(user.User, IsNil)
	}
}
