//go:build functional || catalog || ALL
// +build functional catalog ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"os"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// catalogTenantContext defines whether we use tenant context during catalog tests.
// By default, is off, unless variable VCD_CATALOG_TENANT_CONTEXT is set
var catalogTenantContext = os.Getenv("VCD_CATALOG_TENANT_CONTEXT") != ""

// GetId completes the implementation of interface accessControlType
func (catalog Catalog) GetId() string {
	return catalog.Catalog.ID
}

// GetId completes the implementation of interface accessControlType
func (catalog AdminCatalog) GetId() string {
	return catalog.AdminCatalog.ID
}

func (vcd *TestVCD) Test_AdminCatalogAccessControl(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_AdminCatalogAccessControl: Org name not given.")
		return
	}
	vcd.checkSkipWhenApiToken(check)
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	adminorg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)

	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalogName := "ac-admin-catalog"
	// Create a new catalog
	adminCatalog, err := adminorg.CreateCatalog(catalogName, catalogName)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog, NotNil)
	AddToCleanupList(catalogName, "catalog", vcd.config.VCD.Org, check.TestName())
	vcd.testCatalogAccessControl(adminorg, adminCatalog, check.TestName(), catalogName, check)

	orgInfo, err := adminCatalog.getOrgInfo()
	check.Assert(err, IsNil)
	check.Assert(orgInfo.OrgId, Equals, extractUuid(adminorg.AdminOrg.ID))
	check.Assert(orgInfo.OrgName, Equals, adminorg.AdminOrg.Name)

	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_CatalogAccessControl(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_CatalogAccessControl: Org name not given.")
		return
	}
	vcd.checkSkipWhenApiToken(check)
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	adminorg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)

	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalogName := "ac-catalog"
	// Create a new catalog
	adminCatalog, err := adminorg.CreateCatalog(catalogName, catalogName)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog, NotNil)
	AddToCleanupList(catalogName, "catalog", vcd.config.VCD.Org, check.TestName())
	catalog, err := org.GetCatalogByName(catalogName, true)
	check.Assert(err, IsNil)
	vcd.testCatalogAccessControl(adminorg, catalog, check.TestName(), catalogName, check)

	orgInfo, err := catalog.getOrgInfo()
	check.Assert(err, IsNil)
	check.Assert(orgInfo.OrgId, Equals, extractUuid(adminorg.AdminOrg.ID))
	check.Assert(orgInfo.OrgName, Equals, adminorg.AdminOrg.Name)

	err = catalog.Delete(true, true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) testCatalogAccessControl(adminOrg *AdminOrg, catalog accessControlType, testName, catalogName string, check *C) {

	type testUser struct {
		name string
		role string
		user *OrgUser
	}
	var sameOrgUsers = []testUser{
		{"ac-vapp-author", OrgUserRoleVappAuthor, nil},
		{"ac-org-admin", OrgUserRoleOrganizationAdministrator, nil},
		{"ac-cat-author", OrgUserRoleCatalogAuthor, nil},
	}
	var newOrgUsers = []testUser{
		{"ac-new-vapp-author", OrgUserRoleVappAuthor, nil},
		{"ac-new-org-admin", OrgUserRoleOrganizationAdministrator, nil},
		{"ac-new-cat-author", OrgUserRoleCatalogAuthor, nil},
	}

	orgName := "ac-testorg"
	var newOrg *AdminOrg
	var err error
	if vcd.client.Client.IsSysAdmin {

		// Create a new Org
		task, err := CreateOrg(vcd.client, orgName, orgName, orgName, &types.OrgSettings{}, true)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		newOrg, err = vcd.client.GetAdminOrgByName(orgName)
		check.Assert(err, IsNil)
		AddToCleanupList(orgName, "org", "", testName)
		defer func() {
			if testVerbose {
				fmt.Printf("deleting %s\n", orgName)
			}
			err = newOrg.Disable()
			check.Assert(err, IsNil)
			err = newOrg.Delete(true, true)
			check.Assert(err, IsNil)
		}()
		// Create three users in the new org
		for i := 0; i < len(newOrgUsers); i++ {
			newOrgUsers[i].user, err = newOrg.CreateUserSimple(OrgUserConfiguration{
				Name: newOrgUsers[i].name, Password: newOrgUsers[i].name, RoleName: newOrgUsers[i].role, IsEnabled: true,
			})
			check.Assert(err, IsNil)
			check.Assert(newOrgUsers[i].user, NotNil)
			// No need to add this user to clean-up list. It will be removed when its org is deleted
		}
	}

	checkAllUsersAccess := func(label string) {
		if !testVerbose {
			return
		}
		fmt.Printf("[checkAllUsersAccess] %s\n", label)
		for _, user := range sameOrgUsers {
			err := testAccessToSharedCatalog(adminOrg, catalogName, user.name, user.name)
			check.Assert(err, IsNil)
		}
		for _, user := range newOrgUsers {
			err := testAccessToSharedCatalog(newOrg, catalogName, user.name, user.name)
			check.Assert(err, IsNil)
		}
	}
	checkEmpty := func() {
		settings, err := catalog.GetAccessControl(catalogTenantContext)
		check.Assert(err, IsNil)
		check.Assert(settings.IsSharedToEveryone, Equals, false) // There should not be a global sharing
		check.Assert(settings.AccessSettings, IsNil)             // There should not be any explicit sharing
	}

	// Create three users in the same org
	for i := 0; i < len(sameOrgUsers); i++ {
		sameOrgUsers[i].user, err = adminOrg.CreateUserSimple(OrgUserConfiguration{
			Name: sameOrgUsers[i].name, Password: sameOrgUsers[i].name, RoleName: sameOrgUsers[i].role, IsEnabled: true,
		})
		check.Assert(err, IsNil)
		check.Assert(sameOrgUsers[i].user, NotNil)
		AddToCleanupList(sameOrgUsers[i].name, "user", vcd.config.VCD.Org, testName)
	}

	defer func() {
		for i := 0; i < len(sameOrgUsers); i++ {
			if testVerbose {
				fmt.Printf("deleting %s\n", sameOrgUsers[i].name)
			}
			err = sameOrgUsers[i].user.Delete(false)
			check.Assert(err, IsNil)
		}
	}()
	checkEmpty()
	checkAllUsersAccess("empty")
	globalSettings := types.ControlAccessParams{
		IsSharedToEveryone:  true,
		EveryoneAccessLevel: takeStringPointer(types.ControlAccessReadOnly),
		AccessSettings:      nil,
	}
	err = testAccessControl(catalogName+" catalog global", catalog, globalSettings, globalSettings, true, catalogTenantContext, check)
	check.Assert(err, IsNil)

	// All users on all orgs should be able to see the catalog
	checkAllUsersAccess("shared to everyone")

	// Set control access to one user
	oneUserSettings := types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: nil,
		AccessSettings: &types.AccessSettingList{
			[]*types.AccessSetting{
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: sameOrgUsers[0].user.User.Href,
						Name: sameOrgUsers[0].user.User.Name,
						Type: sameOrgUsers[0].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadWrite,
				},
			},
		},
	}
	err = testAccessControl(catalogName+" catalog one user", catalog, oneUserSettings, oneUserSettings, true, catalogTenantContext, check)
	check.Assert(err, IsNil)

	checkAllUsersAccess("one user - same org")

	// Set control access to one user on a different org
	oneUserOtherOrgSettings := types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: nil,
		AccessSettings: &types.AccessSettingList{
			[]*types.AccessSetting{
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: newOrgUsers[1].user.User.Href,
						Name: newOrgUsers[1].user.User.Name,
						Type: newOrgUsers[1].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadWrite,
				},
			},
		},
	}
	// sharing with a user of a different Org is not allowed: we expect a failure
	err = testAccessControl(catalogName+" catalog one user - other org", catalog, oneUserOtherOrgSettings, oneUserOtherOrgSettings, true, catalogTenantContext, check)
	check.Assert(err, NotNil)
	check.Assert(err.Error(), Matches, `.*Sharing is only supported in an organization.*`)
	checkAllUsersAccess("one user - other org")

	// Check that catalog.GetAccessControl and vdc.GetCatalogAccessControl return the same data
	controlAccess, err := catalog.GetAccessControl(catalogTenantContext)
	check.Assert(err, IsNil)
	orgControlAccessByName, err := adminOrg.GetCatalogAccessControl(catalogName, catalogTenantContext)
	check.Assert(err, IsNil)
	check.Assert(controlAccess, DeepEquals, orgControlAccessByName)

	orgControlAccessById, err := adminOrg.GetCatalogAccessControl(catalog.GetId(), catalogTenantContext)
	check.Assert(err, IsNil)
	check.Assert(controlAccess, DeepEquals, orgControlAccessById)

	// Set control access to two users
	twoUserSettings := types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: nil,
		AccessSettings: &types.AccessSettingList{
			[]*types.AccessSetting{
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: sameOrgUsers[0].user.User.Href,
						//Name: users[0].user.User.Name, // Pass info without name for one of the subjects
						Type: sameOrgUsers[0].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadOnly,
				},
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: sameOrgUsers[1].user.User.Href,
						Name: sameOrgUsers[1].user.User.Name,
						Type: sameOrgUsers[1].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessFullControl,
				},
			},
		},
	}
	err = testAccessControl(catalogName+" catalog two users", catalog, twoUserSettings, twoUserSettings, true, catalogTenantContext, check)
	check.Assert(err, IsNil)

	checkAllUsersAccess("two users - same org")

	// Check removal of sharing setting
	err = catalog.RemoveAccessControl(catalogTenantContext)
	check.Assert(err, IsNil)
	checkEmpty()

	// Set control access to three users
	threeUserSettings := types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: nil,
		AccessSettings: &types.AccessSettingList{
			[]*types.AccessSetting{
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: sameOrgUsers[0].user.User.Href,
						Name: sameOrgUsers[0].user.User.Name,
						Type: sameOrgUsers[0].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadOnly,
				},
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: sameOrgUsers[1].user.User.Href,
						//Name: users[1].user.User.Name,// Pass info without name for one of the subjects
						Type: sameOrgUsers[1].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessFullControl,
				},
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: sameOrgUsers[2].user.User.Href,
						Name: sameOrgUsers[2].user.User.Name,
						Type: sameOrgUsers[2].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadWrite,
				},
			},
		},
	}
	err = testAccessControl(catalogName+" catalog three users", catalog, threeUserSettings, threeUserSettings, true, catalogTenantContext, check)
	check.Assert(err, IsNil)
	checkAllUsersAccess("three users - same org")

	if vcd.client.Client.IsSysAdmin && newOrg != nil {

		// Set control access to three users and an org
		threeUserOrgSettings := types.ControlAccessParams{
			IsSharedToEveryone:  false,
			EveryoneAccessLevel: nil,
			AccessSettings: &types.AccessSettingList{
				[]*types.AccessSetting{
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: sameOrgUsers[0].user.User.Href,
							Name: sameOrgUsers[0].user.User.Name,
							Type: sameOrgUsers[0].user.User.Type,
						},
						ExternalSubject: nil,
						AccessLevel:     types.ControlAccessReadOnly,
					},
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: sameOrgUsers[1].user.User.Href,
							//Name: users[1].user.User.Name,// Pass info without name for one of the subjects
							Type: sameOrgUsers[1].user.User.Type,
						},
						ExternalSubject: nil,
						AccessLevel:     types.ControlAccessFullControl,
					},
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: sameOrgUsers[2].user.User.Href,
							Name: sameOrgUsers[2].user.User.Name,
							Type: sameOrgUsers[2].user.User.Type,
						},
						ExternalSubject: nil,
						AccessLevel:     types.ControlAccessReadWrite,
					},
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: newOrg.AdminOrg.HREF,
							Name: newOrg.AdminOrg.Name,
							Type: newOrg.AdminOrg.Type,
						},
						ExternalSubject: nil,
						AccessLevel:     types.ControlAccessReadOnly,
					},
				},
			},
		}
		err = testAccessControl(catalogName+" catalog three users and org", catalog, threeUserOrgSettings, threeUserOrgSettings, true, catalogTenantContext, check)
		check.Assert(err, IsNil)
		checkAllUsersAccess("three users + one org ")

		err = catalog.RemoveAccessControl(catalogTenantContext)
		check.Assert(err, IsNil)
		checkEmpty()

		// Set control access to two org
		twoOrgsSettings := types.ControlAccessParams{
			IsSharedToEveryone:  false,
			EveryoneAccessLevel: nil,
			AccessSettings: &types.AccessSettingList{
				[]*types.AccessSetting{
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: adminOrg.AdminOrg.HREF,
							Name: adminOrg.AdminOrg.Name,
							Type: adminOrg.AdminOrg.Type,
						},
						ExternalSubject: nil,
						AccessLevel:     types.ControlAccessFullControl,
					},
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: newOrg.AdminOrg.HREF,
							Name: newOrg.AdminOrg.Name,
							Type: newOrg.AdminOrg.Type,
						},
						ExternalSubject: nil,
						AccessLevel:     types.ControlAccessReadOnly,
					},
				},
			},
		}
		err = testAccessControl(catalogName+" catalog two org", catalog, twoOrgsSettings, twoOrgsSettings, true, catalogTenantContext, check)
		check.Assert(err, IsNil)
		checkAllUsersAccess("two orgs ")
	}

	// Set empty settings explicitly
	emptySettings := types.ControlAccessParams{
		IsSharedToEveryone: false,
	}
	err = testAccessControl(catalogName+" catalog empty", catalog, emptySettings, emptySettings, false, catalogTenantContext, check)
	check.Assert(err, IsNil)

	checkEmpty()

}

// testAccessToSharedCatalog is an informational tool to show whether a given user can access a shared catalog
// Its output is visible when -vcd-verbose is used on the command line
func testAccessToSharedCatalog(org *AdminOrg, catalogName, userName, password string) error {
	vcdClient := NewVCDClient(org.client.VCDHREF, true)
	err := vcdClient.Authenticate(userName, password, org.AdminOrg.Name)
	if err != nil {
		return err
	}
	catalogs, err := queryCatalogList(&vcdClient.Client, nil)
	if err != nil {
		return fmt.Errorf("error retrieving catalogs as user '%s': %s", userName, err)
	}

	found := false
	for _, cat := range catalogs {
		if catalogName == cat.Name {
			found = true
		}
	}

	if testVerbose {
		if found {
			fmt.Printf("++ user %s from org %s can access catalog %s\n", userName, org.AdminOrg.Name, catalogName)
		} else {
			fmt.Printf("## user %s from org %s CANNOT access catalog %s\n", userName, org.AdminOrg.Name, catalogName)
		}
	}
	return nil
}
