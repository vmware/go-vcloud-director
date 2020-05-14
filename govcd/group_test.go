// +build user functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GroupCRUD(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	// type groupTestData struct {
	// 	name     string
	// 	roleName string
	// }

	type groupTestData struct {
		name       string // name of the user. Note: only lowercase letters allowed
		roleName   string // the role this user is created with
		secondRole string // The role to which we change using Update()
	}
	groupData := []groupTestData{
		{
			name:       "test_group_admin",
			roleName:   OrgUserRoleOrganizationAdministrator,
			secondRole: OrgUserRoleVappAuthor,
		},
		{
			name:       "test_group_vapp_author",
			roleName:   OrgUserRoleVappAuthor,
			secondRole: OrgUserRoleVappUser,
		},
		{
			name:       "test_group_vapp_user",
			roleName:   OrgUserRoleVappUser,
			secondRole: OrgUserRoleConsoleAccessOnly,
		},
		{
			name:       "test_group_console_access",
			roleName:   OrgUserRoleConsoleAccessOnly,
			secondRole: OrgUserRoleCatalogAuthor,
		},
		{
			name:       "test_group_catalog_author",
			roleName:   OrgUserRoleCatalogAuthor,
			secondRole: OrgUserRoleOrganizationAdministrator,
		},
	}

	for _, gd := range groupData {

		role, err := adminOrg.GetRoleReference(gd.roleName)
		check.Assert(err, IsNil)

		g := types.Group{
			Name:         gd.name,
			Role:         role,
			ProviderType: "SAML",
		}

		newGroup := NewGroup(adminOrg.client, adminOrg)
		newGroup.Group = &g
		grp, err := adminOrg.CreateGroup(newGroup.Group)
		check.Assert(err, IsNil)

		secondRole, err := adminOrg.GetRoleReference(gd.secondRole)
		check.Assert(err, IsNil)
		grp.Group.Role = secondRole
		grp.Group.Description = "adding description"

		err = grp.Update()
		check.Assert(err, IsNil)

		err = grp.Delete()
		check.Assert(err, IsNil)

	}

	// quotaDeployed := 10
	// quotaStored := 10
	// for _, ud := range userData {
	// 	quotaDeployed += 2
	// 	quotaStored += 2
	// 	fmt.Printf("# Creating user %s with role %s\n", ud.name, ud.roleName)
	// 	// Uncomment the following lines to see creation request and response
	// 	// enableDebugShowRequest()
	// 	// enableDebugShowResponse()
	// 	var userDefinition = OrgUserConfiguration{
	// 		Name:            ud.name,
	// 		Password:        "user_pass",
	// 		RoleName:        ud.roleName,
	// 		ProviderType:    OrgUserProviderIntegrated,
	// 		DeployedVmQuota: quotaDeployed,
	// 		StoredVmQuota:   quotaStored,
	// 		FullName:        strings.ReplaceAll(ud.name, "_", " "),
	// 		Description:     "user " + strings.ReplaceAll(ud.name, "_", " "),
	// 		IsEnabled:       true,
	// 		IM:              "TextIM",
	// 		EmailAddress:    "somename@somedomain.com",
	// 		Telephone:       "999 888-7777",
	// 	}

	// 	user, err := adminOrg.CreateUserSimple(userDefinition)
	// 	// disableDebugShowRequest()
	// 	// disableDebugShowResponse()
	// 	check.Assert(err, IsNil)

	// 	AddToCleanupList(ud.name, "user", user.AdminOrg.AdminOrg.Name, check.TestName())
	// 	check.Assert(user.User, NotNil)
	// 	check.Assert(user.User.Name, Equals, ud.name)
	// 	check.Assert(user.GetRoleName(), Equals, ud.roleName)
	// 	check.Assert(user.User.IsEnabled, Equals, true)
	// 	check.Assert(user.User.FullName, Equals, userDefinition.FullName)
	// 	check.Assert(user.User.EmailAddress, Equals, userDefinition.EmailAddress)
	// 	check.Assert(user.User.IM, Equals, userDefinition.IM)
	// 	check.Assert(user.User.Telephone, Equals, userDefinition.Telephone)
	// 	check.Assert(user.User.StoredVmQuota, Equals, userDefinition.StoredVmQuota)
	// 	check.Assert(user.User.DeployedVmQuota, Equals, userDefinition.DeployedVmQuota)

	// 	err = user.Disable()
	// 	check.Assert(err, IsNil)
	// 	check.Assert(user.User.IsEnabled, Equals, false)

	// 	fmt.Printf("# Updating user %s with role %s\n", ud.name, ud.secondRole)
	// 	err = user.ChangeRole(ud.secondRole)
	// 	check.Assert(err, IsNil)
	// 	check.Assert(user.GetRoleName(), Equals, ud.secondRole)

	// 	err = user.Enable()
	// 	check.Assert(err, IsNil)
	// 	check.Assert(user.User.IsEnabled, Equals, true)
	// 	err = user.ChangePassword("new_pass")
	// 	check.Assert(err, IsNil)
	// }

	// var enableMap = map[bool]string{
	// 	true:  "enabled",
	// 	false: "disabled",
	// }
	// for _, ud := range userData {
	// 	user, err := adminOrg.GetUserByNameOrId(ud.name, true)
	// 	check.Assert(err, IsNil)

	// 	fmt.Printf("# deleting user %s (%s - %s)\n", ud.name, user.GetRoleName(), enableMap[user.User.IsEnabled])
	// 	// uncomment the following two lines to see the deletion request and response
	// 	// enableDebugShowRequest()
	// 	// enableDebugShowResponse()
	// 	err = user.Delete(true)
	// 	// disableDebugShowRequest()
	// 	// disableDebugShowResponse()
	// 	check.Assert(err, IsNil)
	// 	user, err = adminOrg.GetUserByNameOrId(user.User.ID, true)
	// 	check.Assert(err, NotNil)
	// 	// Tests both the error directly and the function IsNotFound
	// 	check.Assert(err, Equals, ErrorEntityNotFound)
	// 	check.Assert(IsNotFound(err), Equals, true)
	// 	// Expect a null pointer when user is not found
	// 	check.Assert(user, IsNil)
	// }
}
