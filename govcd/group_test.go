//go:build user || functional || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// test_GroupCRUD tests out CRUD capabilities for Org Groups.
// Note: Because it requires LDAP to be functional - this test is run from main test "Test_LDAP"
// which sets up LDAP configuration.
func (vcd *TestVCD) test_GroupCRUD(check *C) {
	fmt.Printf("Running: %s\n", "test_GroupCRUD")
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	type groupTestData struct {
		name         string
		secondName   string
		roleName     string
		secondRole   string
		providerType string
	}
	groupData := []groupTestData{
		{
			name:         "ship_crew",
			secondName:   "admin_staff",
			roleName:     OrgUserRoleOrganizationAdministrator,
			secondRole:   OrgUserRoleVappAuthor,
			providerType: OrgUserProviderIntegrated,
		},
		{
			name:         "admin_staff",
			secondName:   "ship_crew",
			roleName:     OrgUserRoleVappAuthor,
			secondRole:   OrgUserRoleVappUser,
			providerType: OrgUserProviderIntegrated,
		},
		// SAML must be configured in vCD Org to make providerType=OrgUserProviderSAML tests work
		// {
		// 	name:         "test_group_vapp_user",
		// 	roleName:     OrgUserRoleVappUser,
		// 	secondRole:   OrgUserRoleConsoleAccessOnly,
		// 	providerType: OrgUserProviderSAML,
		// },
		// {
		// 	name:         "test_group_console_access",
		// 	roleName:     OrgUserRoleConsoleAccessOnly,
		// 	secondRole:   OrgUserRoleCatalogAuthor,
		// 	providerType: OrgUserProviderSAML,
		// },
		// {
		// 	name:         "test_group_catalog_author",
		// 	roleName:     OrgUserRoleCatalogAuthor,
		// 	secondRole:   OrgUserRoleOrganizationAdministrator,
		// 	providerType: OrgUserProviderSAML,
		// },
		// {
		// 	name:         "test_group_defered_to_identity_provider",
		// 	roleName:     OrgUserRoleDeferToIdentityProvider,
		// 	secondRole:   OrgUserRoleOrganizationAdministrator,
		// 	providerType: OrgUserProviderSAML,
		// },
	}

	for _, gd := range groupData {

		role, err := adminOrg.GetRoleReference(gd.roleName)
		check.Assert(err, IsNil)

		groupDefinition := types.Group{
			Name:         gd.name,
			Role:         role,
			ProviderType: gd.providerType,
		}

		newGroup := NewGroup(adminOrg.client, adminOrg)
		newGroup.Group = &groupDefinition
		if testVerbose {
			fmt.Printf("# creating '%s' group '%s' with role '%s'\n", gd.providerType, gd.name, gd.roleName)
		}
		createdGroup, err := adminOrg.CreateGroup(newGroup.Group)
		check.Assert(err, IsNil)
		AddToCleanupList(gd.name, "group", newGroup.AdminOrg.AdminOrg.Name, "test_GroupCRUD")

		foundGroup, err := adminOrg.GetGroupByName(gd.name, true)
		check.Assert(err, IsNil)

		check.Assert(foundGroup.Group.Href, Equals, createdGroup.Group.Href)
		check.Assert(foundGroup.Group.Name, Equals, createdGroup.Group.Name)

		// Setup for update
		secondRole, err := adminOrg.GetRoleReference(gd.secondRole)
		check.Assert(err, IsNil)
		createdGroup.Group.Role = secondRole

		if testVerbose {
			fmt.Printf("# updating '%s' group '%s' to role '%s'\n", gd.providerType, gd.name, gd.secondRole)
		}
		err = createdGroup.Update()
		check.Assert(err, IsNil)

		foundGroup2, err := adminOrg.GetGroupByName(gd.name, true)
		check.Assert(err, IsNil)
		check.Assert(foundGroup2.Group.Href, Equals, createdGroup.Group.Href)
		check.Assert(foundGroup2.Group.Name, Equals, createdGroup.Group.Name)

		if testVerbose {
			fmt.Printf("# removing '%s' group '%s'\n", gd.providerType, gd.name)
		}
		err = createdGroup.Delete()
		check.Assert(err, IsNil)
	}
}

// test_GroupFinderGetGenericEntity uses testFinderGetGenericEntity to validate that ByName, ById
// ByNameOrId method work properly.
// Note: Because it requires LDAP to be functional - this test is run from main test "Test_LDAP"
// which sets up LDAP configuration.
func (vcd *TestVCD) test_GroupFinderGetGenericEntity(check *C) {
	fmt.Printf("Running: %s\n", "test_GroupFinderGetGenericEntity")

	const groupName = "ship_crew"
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	role, err := adminOrg.GetRoleReference(OrgUserRoleOrganizationAdministrator)
	check.Assert(err, IsNil)

	group := NewGroup(adminOrg.client, adminOrg)
	group.Group = &types.Group{
		Name:         groupName,
		Role:         role,
		ProviderType: OrgUserProviderIntegrated,
	}

	_, err = adminOrg.CreateGroup(group.Group)
	check.Assert(err, IsNil)
	AddToCleanupList(groupName, "group", group.AdminOrg.AdminOrg.Name, check.TestName())

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return adminOrg.GetGroupByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return adminOrg.GetGroupById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return adminOrg.GetGroupByNameOrId(id, refresh)
	}

	// Refresh adminOrg so that user data is present
	err = adminOrg.Refresh()
	check.Assert(err, IsNil)

	var def = getterTestDefinition{
		parentType:    "AdminOrg",
		parentName:    vcd.config.VCD.Org,
		entityType:    "OrgGroup",
		entityName:    groupName,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)

	// Remove group because LDAP cleanup will fail.
	grp, err := adminOrg.GetGroupByName(group.Group.Name, true)
	check.Assert(err, IsNil)
	err = grp.Delete()
	check.Assert(err, IsNil)
}

// test_GroupUserListIsPopulated checks that when retrieving the existing groups from the testing LDAP,
// the user list reference is populated.
// Note: Because it requires LDAP to be functional - this test is run from main test "Test_LDAP"
// which sets up LDAP configuration.
func (vcd *TestVCD) test_GroupUserListIsPopulated(check *C) {
	fmt.Printf("Running: %s\n", "test_GroupUserListIsPopulated")
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	roleRef, err := adminOrg.GetRoleReference(OrgUserRoleOrganizationAdministrator)
	check.Assert(err, IsNil)

	group := NewGroup(adminOrg.client, adminOrg)
	const groupName = "ship_crew"
	group.Group = &types.Group{
		Name:         groupName,
		Role:         roleRef,
		ProviderType: OrgUserProviderIntegrated,
	}

	_, err = adminOrg.CreateGroup(group.Group)
	check.Assert(err, IsNil)
	AddToCleanupList(groupName, "group", group.AdminOrg.AdminOrg.Name, check.TestName())

	user := NewUser(adminOrg.client, adminOrg)
	const userName = "fry"
	user.User = &types.User{
		Name:         userName,
		Password:     userName,
		Role:         roleRef,
		IsExternal:   true,
		IsEnabled:    true,
		ProviderType: OrgUserProviderIntegrated,
	}
	_, err = adminOrg.CreateUser(user.User)
	check.Assert(err, IsNil)
	AddToCleanupList(userName, "user", group.AdminOrg.AdminOrg.Name, check.TestName())

	grp, err := adminOrg.GetGroupByName(group.Group.Name, true)
	check.Assert(err, IsNil)
	check.Assert(grp.Group.UsersList, NotNil)
	check.Assert(grp.Group.UsersList.UserReference[0], NotNil)

	// We check here that usersList doesn't make VCD fail, they should be sent as nil
	err = grp.Update()
	check.Assert(err, IsNil)

	user, err = adminOrg.GetUserByHref(grp.Group.UsersList.UserReference[0].HREF)
	check.Assert(err, IsNil)
	check.Assert(user.User.Name, Equals, userName)
	check.Assert(len(user.User.GroupReferences.GroupReference), Equals, 1)

	// We check here that the user used for update is the same as we had originally, except the user list
	grp.Group.UsersList = nil
	check.Assert(copyWithoutUserList(grp.Group), DeepEquals, grp.Group)

	// We check here that groupReferences doesn't make VCD fail, they should be sent as nil
	err = user.Update()
	check.Assert(err, IsNil)

	// Cleanup
	err = user.Delete(false)
	check.Assert(err, IsNil)
	err = grp.Delete()
	check.Assert(err, IsNil)
}
