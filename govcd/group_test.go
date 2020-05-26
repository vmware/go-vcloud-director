// +build user functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GroupCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	// LDAP must be configured for this test to work
	networkName, vappName, vmName := vcd.configureLdap(check)
	defer func() {
		// Immediately release resources for further tests
		vcd.unconfigureLdap(check, networkName, vappName, vmName)
	}()

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
		// SAML must be configured on the system to make it work
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
			ProviderType: OrgUserProviderIntegrated, // Integrated covers LDAP
		}

		newGroup := NewGroup(adminOrg.client, adminOrg)
		newGroup.Group = &groupDefinition
		if testVerbose {
			fmt.Printf("# creating '%s' group '%s' with role '%s'\n", gd.providerType, gd.name, gd.roleName)
		}
		createdGroup, err := adminOrg.CreateGroup(newGroup.Group)
		check.Assert(err, IsNil)
		AddToCleanupList(gd.name, "group", newGroup.AdminOrg.AdminOrg.Name, check.TestName())

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

// Test_GroupFinderGetGenericEntity uses testFinderGetGenericEntity to validate that ByName, ById
// ByNameOrId method work properly.
func (vcd *TestVCD) Test_GroupFinderGetGenericEntity(check *C) {
	const groupName = "group_generic_entity"
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	role, err := adminOrg.GetRoleReference(OrgUserRoleOrganizationAdministrator)
	check.Assert(err, IsNil)

	group := NewGroup(adminOrg.client, adminOrg)
	group.Group = &types.Group{
		Name:         groupName,
		Role:         role,
		ProviderType: OrgUserProviderSAML,
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
}
