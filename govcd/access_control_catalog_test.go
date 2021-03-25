// +build functional catalog ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"os"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// catalogTenantContext defines whether we use tenant context during catalog tests.
// By default is off, unless variable VCD_CATALOG_TENANT_CONTEXT is set
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
	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	adminorg, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)

	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalogName := "ac-admin-catalog"
	// Create a new catalog
	adminCatalog, err := adminorg.CreateCatalog(ctx, catalogName, catalogName)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog, NotNil)
	AddToCleanupList(catalogName, "catalog", vcd.config.VCD.Org, check.TestName())
	vcd.testCatalogAccessControl(adminorg, adminCatalog, check.TestName(), catalogName, check)

	orgInfo, err := adminCatalog.getOrgInfo(ctx)
	check.Assert(err, IsNil)
	check.Assert(orgInfo.id, Equals, extractUuid(adminorg.AdminOrg.ID))
	check.Assert(orgInfo.name, Equals, adminorg.AdminOrg.Name)

	err = adminCatalog.Delete(ctx, true, true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_CatalogAccessControl(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_CatalogAccessControl: Org name not given.")
		return
	}
	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	adminorg, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)

	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalogName := "ac-catalog"
	// Create a new catalog
	adminCatalog, err := adminorg.CreateCatalog(ctx, catalogName, catalogName)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog, NotNil)
	AddToCleanupList(catalogName, "catalog", vcd.config.VCD.Org, check.TestName())
	catalog, err := org.GetCatalogByName(ctx, catalogName, true)
	check.Assert(err, IsNil)
	vcd.testCatalogAccessControl(adminorg, catalog, check.TestName(), catalogName, check)

	orgInfo, err := catalog.getOrgInfo(ctx)
	check.Assert(err, IsNil)
	check.Assert(orgInfo.id, Equals, extractUuid(adminorg.AdminOrg.ID))
	check.Assert(orgInfo.name, Equals, adminorg.AdminOrg.Name)

	err = catalog.Delete(ctx, true, true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) testCatalogAccessControl(adminOrg *AdminOrg, catalog accessControlType, testName, catalogName string, check *C) {

	var users = []struct {
		name string
		role string
		user *OrgUser
	}{
		{"ac-user1", OrgUserRoleVappAuthor, nil},
		{"ac-user2", OrgUserRoleOrganizationAdministrator, nil},
		{"ac-user3", OrgUserRoleCatalogAuthor, nil},
	}

	orgName := "ac-testorg"
	var newOrg *AdminOrg
	var err error
	if vcd.client.Client.IsSysAdmin {

		// Create a new Org
		task, err := CreateOrg(ctx, vcd.client, orgName, orgName, orgName, &types.OrgSettings{}, true)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion(ctx)
		check.Assert(err, IsNil)
		newOrg, err = vcd.client.GetAdminOrgByName(ctx, orgName)
		check.Assert(err, IsNil)
		AddToCleanupList(orgName, "org", "", testName)
		defer func() {
			if testVerbose {
				fmt.Printf("deleting %s\n", orgName)
			}
			err = newOrg.Disable(ctx)
			check.Assert(err, IsNil)
			err = newOrg.Delete(ctx, true, true)
			check.Assert(err, IsNil)
		}()
	}

	checkEmpty := func() {
		settings, err := catalog.GetAccessControl(ctx, catalogTenantContext)
		check.Assert(err, IsNil)
		check.Assert(settings.IsSharedToEveryone, Equals, false) // There should not be a global sharing
		check.Assert(settings.AccessSettings, IsNil)             // There should not be any explicit sharing
	}

	// Create three users
	for i := 0; i < len(users); i++ {
		users[i].user, err = adminOrg.CreateUserSimple(ctx, OrgUserConfiguration{
			Name: users[i].name, Password: users[i].name, RoleName: users[i].role, IsEnabled: true,
		})
		check.Assert(err, IsNil)
		check.Assert(users[i].user, NotNil)
		AddToCleanupList(users[i].name, "user", vcd.config.VCD.Org, testName)
	}
	defer func() {
		for i := 0; i < len(users); i++ {
			if testVerbose {
				fmt.Printf("deleting %s\n", users[i].name)
			}
			err = users[i].user.Delete(ctx, false)
			check.Assert(err, IsNil)
		}
	}()
	checkEmpty()
	//globalSettings := types.ControlAccessParams{
	//	IsSharedToEveryone:  true,
	//	EveryoneAccessLevel: takeStringPointer(types.ControlAccessReadWrite),
	//	AccessSettings: nil,
	//}
	//err = testAccessControl(catalogName+" catalog global", catalog, globalSettings, globalSettings, true, check)
	//check.Assert(err, IsNil)

	// Set control access to one user
	oneUserSettings := types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: nil,
		AccessSettings: &types.AccessSettingList{
			[]*types.AccessSetting{
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: users[0].user.User.Href,
						Name: users[0].user.User.Name,
						Type: users[0].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadWrite,
				},
			},
		},
	}
	err = testAccessControl(ctx, catalogName+" catalog one user", catalog, oneUserSettings, oneUserSettings, true, catalogTenantContext, check)
	check.Assert(err, IsNil)

	// Check that vapp.GetAccessControl and vdc.GetVappAccessControl return the same data
	controlAccess, err := catalog.GetAccessControl(ctx, catalogTenantContext)
	check.Assert(err, IsNil)
	orgControlAccessByName, err := adminOrg.GetCatalogAccessControl(ctx, catalogName, catalogTenantContext)
	check.Assert(err, IsNil)
	check.Assert(controlAccess, DeepEquals, orgControlAccessByName)

	orgControlAccessById, err := adminOrg.GetCatalogAccessControl(ctx, catalog.GetId(), catalogTenantContext)
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
						HREF: users[0].user.User.Href,
						//Name: users[0].user.User.Name, // Pass info without name for one of the subjects
						Type: users[0].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadOnly,
				},
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: users[1].user.User.Href,
						Name: users[1].user.User.Name,
						Type: users[1].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessFullControl,
				},
			},
		},
	}
	err = testAccessControl(ctx, catalogName+" catalog two users", catalog, twoUserSettings, twoUserSettings, true, catalogTenantContext, check)
	check.Assert(err, IsNil)

	// Check removal of sharing setting
	err = catalog.RemoveAccessControl(ctx, catalogTenantContext)
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
						HREF: users[0].user.User.Href,
						Name: users[0].user.User.Name,
						Type: users[0].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadOnly,
				},
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: users[1].user.User.Href,
						//Name: users[1].user.User.Name,// Pass info without name for one of the subjects
						Type: users[1].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessFullControl,
				},
				&types.AccessSetting{
					Subject: &types.LocalSubject{
						HREF: users[2].user.User.Href,
						Name: users[2].user.User.Name,
						Type: users[2].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadWrite,
				},
			},
		},
	}
	err = testAccessControl(ctx, catalogName+" catalog three users", catalog, threeUserSettings, threeUserSettings, true, catalogTenantContext, check)
	check.Assert(err, IsNil)

	if vcd.client.Client.IsSysAdmin && newOrg != nil {

		// Set control access to three users and an org
		threeUserOrgSettings := types.ControlAccessParams{
			IsSharedToEveryone:  false,
			EveryoneAccessLevel: nil,
			AccessSettings: &types.AccessSettingList{
				[]*types.AccessSetting{
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: users[0].user.User.Href,
							Name: users[0].user.User.Name,
							Type: users[0].user.User.Type,
						},
						ExternalSubject: nil,
						AccessLevel:     types.ControlAccessReadOnly,
					},
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: users[1].user.User.Href,
							//Name: users[1].user.User.Name,// Pass info without name for one of the subjects
							Type: users[1].user.User.Type,
						},
						ExternalSubject: nil,
						AccessLevel:     types.ControlAccessFullControl,
					},
					&types.AccessSetting{
						Subject: &types.LocalSubject{
							HREF: users[2].user.User.Href,
							Name: users[2].user.User.Name,
							Type: users[2].user.User.Type,
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
		err = testAccessControl(ctx, catalogName+" catalog three users and org", catalog, threeUserOrgSettings, threeUserOrgSettings, true, catalogTenantContext, check)
		check.Assert(err, IsNil)

		err = catalog.RemoveAccessControl(ctx, catalogTenantContext)
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
		err = testAccessControl(ctx, catalogName+" catalog two org", catalog, twoOrgsSettings, twoOrgsSettings, true, catalogTenantContext, check)
		check.Assert(err, IsNil)
	}

	// Set empty settings explicitly
	emptySettings := types.ControlAccessParams{
		IsSharedToEveryone: false,
	}
	err = testAccessControl(ctx, catalogName+" catalog empty", catalog, emptySettings, emptySettings, false, catalogTenantContext, check)
	check.Assert(err, IsNil)

	checkEmpty()

}
