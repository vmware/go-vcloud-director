//go:build functional || vapp || ALL

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

// vappTenantContext defines whether we use tenant context during vApp tests.
// By default is ON. It is disabled if VCD_VAPP_SYSTEM_CONTEXT is set
var vappTenantContext = os.Getenv("VCD_VAPP_SYSTEM_CONTEXT") == ""

// GetId completes the implementation of interface accessControlType
func (vapp VApp) GetId() string {
	return vapp.VApp.ID
}

func (vcd *TestVCD) Test_VappAccessControl(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_VappAccessControl: Org name not given.")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_VappAccessControl: VDC name not given.")
		return
	}
	vcd.checkSkipWhenApiToken(check)
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	vappName := "ac-vapp"
	var users = []struct {
		name string
		role string
		user *OrgUser
	}{
		{"ac-user1", OrgUserRoleVappAuthor, nil},
		{"ac-user2", OrgUserRoleOrganizationAdministrator, nil},
		{"ac-user3", OrgUserRoleCatalogAuthor, nil},
	}

	// Create a new vApp
	vapp, err := makeEmptyVapp(vdc, vappName, "")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)
	AddToCleanupList(vappName, "vapp", vcd.config.VCD.Vdc, "Test_VappAccessControl")

	checkEmpty := func() {
		settings, err := vapp.GetAccessControl(vappTenantContext)
		check.Assert(err, IsNil)
		check.Assert(settings.IsSharedToEveryone, Equals, false) // There should not be a global sharing
		check.Assert(settings.AccessSettings, IsNil)             // There should not be any explicit sharing
	}

	// Create three users
	for i := 0; i < len(users); i++ {
		users[i].user, err = org.CreateUserSimple(OrgUserConfiguration{
			Name: users[i].name, Password: users[i].name, RoleName: users[i].role, IsEnabled: true,
		})
		check.Assert(err, IsNil)
		check.Assert(users[i].user, NotNil)
		AddToCleanupList(users[i].name, "user", vcd.config.VCD.Org, "Test_VappAccessControl")
	}

	// Clean up environment
	defer func() {
		if testVerbose {
			fmt.Printf("deleting %s\n", vappName)
		}
		task, err := vapp.Delete()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		for i := 0; i < len(users); i++ {
			if testVerbose {
				fmt.Printf("deleting %s\n", users[i].name)
			}
			err = users[i].user.Delete(false)
			check.Assert(err, IsNil)
		}
	}()
	checkEmpty()

	// Set access control to every user and group
	allUsersSettings := types.ControlAccessParams{
		EveryoneAccessLevel: addrOf(types.ControlAccessReadOnly),
		IsSharedToEveryone:  true,
	}

	// Use generic testAccessControl. Here vapp is passed as accessControlType interface
	err = testAccessControl("vapp all users RO", vapp, allUsersSettings, allUsersSettings, true, vappTenantContext, check)
	check.Assert(err, IsNil)

	allUsersSettings = types.ControlAccessParams{
		EveryoneAccessLevel: addrOf(types.ControlAccessReadWrite),
		IsSharedToEveryone:  true,
	}
	err = testAccessControl("vapp all users R/W", vapp, allUsersSettings, allUsersSettings, true, vappTenantContext, check)
	check.Assert(err, IsNil)

	// Set access control to one user
	oneUserSettings := types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: nil,
		AccessSettings: &types.AccessSettingList{
			AccessSetting: []*types.AccessSetting{
				{
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
	err = testAccessControl("vapp one user", vapp, oneUserSettings, oneUserSettings, true, vappTenantContext, check)
	check.Assert(err, IsNil)

	// Check that vapp.GetAccessControl and vdc.GetVappAccessControl return the same data
	controlAccess, err := vapp.GetAccessControl(vappTenantContext)
	check.Assert(err, IsNil)
	vdcControlAccessName, err := vdc.GetVappAccessControl(vappName, vappTenantContext)
	check.Assert(err, IsNil)
	check.Assert(controlAccess, DeepEquals, vdcControlAccessName)

	vdcControlAccessId, err := vdc.GetVappAccessControl(vapp.VApp.ID, vappTenantContext)
	check.Assert(err, IsNil)
	check.Assert(controlAccess, DeepEquals, vdcControlAccessId)

	// Set access control to two users
	twoUserSettings := types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: nil,
		AccessSettings: &types.AccessSettingList{
			AccessSetting: []*types.AccessSetting{
				{
					Subject: &types.LocalSubject{
						HREF: users[0].user.User.Href,
						//Name: users[0].user.User.Name, // Pass info without name for one of the subjects
						Type: users[0].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadOnly,
				},
				{
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
	err = testAccessControl("vapp two users", vapp, twoUserSettings, twoUserSettings, true, vappTenantContext, check)
	check.Assert(err, IsNil)

	// Check removal of sharing setting
	err = vapp.RemoveAccessControl(vappTenantContext)
	check.Assert(err, IsNil)
	checkEmpty()

	// Set access control to three users
	threeUserSettings := types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: nil,
		AccessSettings: &types.AccessSettingList{
			AccessSetting: []*types.AccessSetting{
				{
					Subject: &types.LocalSubject{
						HREF: users[0].user.User.Href,
						Name: users[0].user.User.Name,
						Type: users[0].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessReadOnly,
				},
				{
					Subject: &types.LocalSubject{
						HREF: users[1].user.User.Href,
						//Name: users[1].user.User.Name,// Pass info without name for one of the subjects
						Type: users[1].user.User.Type,
					},
					ExternalSubject: nil,
					AccessLevel:     types.ControlAccessFullControl,
				},
				{
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
	err = testAccessControl("vapp three users", vapp, threeUserSettings, threeUserSettings, true, vappTenantContext, check)
	check.Assert(err, IsNil)

	// Set empty settings explicitly
	emptySettings := types.ControlAccessParams{
		IsSharedToEveryone: false,
	}
	err = testAccessControl("vapp empty", vapp, emptySettings, emptySettings, false, vappTenantContext, check)
	check.Assert(err, IsNil)

	checkEmpty()

	orgInfo, err := vapp.getOrgInfo()
	check.Assert(err, IsNil)
	check.Assert(orgInfo.OrgId, Equals, extractUuid(org.AdminOrg.ID))
	check.Assert(orgInfo.OrgName, Equals, org.AdminOrg.Name)
}
