// +build functional vapp ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

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
	vapp, err := makeEmptyVapp(vdc, vappName)
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)
	AddToCleanupList(vappName, "vapp", vcd.config.VCD.Org+"|"+vcd.config.VCD.Vdc, "Test_VappAccessControl")

	checkEmpty := func() {
		settings, err := vapp.GetAccessControl()
		check.Assert(err, IsNil)
		check.Assert(settings.IsSharedToEveryone, Equals, false) // There should not be a global sharing
		check.Assert(settings.AccessSettings, IsNil)             // There should not be any explicit sharing
	}

	// Create three users
	for i := 0; i < len(users); i++ {
		//user, err := org.GetUserByName(users[i].name, true)
		//if err == nil {
		//	users[i].user = user
		//	continue
		//}

		users[i].user, err = org.CreateUserSimple(OrgUserConfiguration{
			Name: users[i].name, Password: users[i].name, RoleName: users[i].role, IsEnabled: true,
		})
		check.Assert(err, IsNil)
		check.Assert(users[i].user, NotNil)
		AddToCleanupList(users[i].name, "user", vcd.config.VCD.Org, "Test_GetVappControlAccess")
	}

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

	// Set control access to every user and group
	allUsersSettings := types.ControlAccessParams{
		EveryoneAccessLevel: takeStringPointer(types.ControlAccessReadOnly),
		IsSharedToEveryone:  true,
	}
	err = testAccessControl("vapp all users RO", vapp, allUsersSettings, allUsersSettings, true, check)
	check.Assert(err, IsNil)

	allUsersSettings = types.ControlAccessParams{
		EveryoneAccessLevel: takeStringPointer(types.ControlAccessReadWrite),
		IsSharedToEveryone:  true,
	}
	err = testAccessControl("vapp all users R/W", vapp, allUsersSettings, allUsersSettings, true, check)
	check.Assert(err, IsNil)

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
	err = testAccessControl("vapp one user", vapp, oneUserSettings, oneUserSettings, true, check)
	check.Assert(err, IsNil)

	// Check that vapp.GetAccessControl and vdc.GetVappControlAccess return the same data
	controlAccess, err := vapp.GetAccessControl()
	check.Assert(err, IsNil)
	vdcControlAccessName, err := vdc.GetVappControlAccess(vappName)
	check.Assert(err, IsNil)
	check.Assert(controlAccess, DeepEquals, vdcControlAccessName)

	vdcControlAccessId, err := vdc.GetVappControlAccess(vapp.VApp.ID)
	check.Assert(err, IsNil)
	check.Assert(controlAccess, DeepEquals, vdcControlAccessId)

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
	err = testAccessControl("vapp two users", vapp, twoUserSettings, twoUserSettings, true, check)
	check.Assert(err, IsNil)

	// Check removal of sharing setting
	err = vapp.RemoveAccessControl()
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
	err = testAccessControl("vapp three users", vapp, threeUserSettings, threeUserSettings, true, check)
	check.Assert(err, IsNil)

	// Set empty settings explicitly
	emptySettings := types.ControlAccessParams{
		IsSharedToEveryone: false,
	}
	err = testAccessControl("vapp empty", vapp, emptySettings, emptySettings, false, check)
	check.Assert(err, IsNil)

	checkEmpty()

}
