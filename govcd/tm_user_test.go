//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmUserLocal(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()

	// TODO: TM: Change to vcdClient.GetTmOrgById(orgId), requires implementing Role support for that type
	adminOrg, err := vcd.client.GetAdminOrgById(org.TmOrg.ID)

	roleOrgAdmin, err := adminOrg.GetRoleByName("Organization Administrator")
	check.Assert(err, IsNil)
	check.Assert(roleOrgAdmin, NotNil)

	roleOrgUser, err := adminOrg.GetRoleByName("Organization User")
	check.Assert(err, IsNil)
	check.Assert(roleOrgUser, NotNil)

	userConfig := &types.TmUser{
		Username:       "test-user",
		Password:       "CHANGE-ME",
		RoleEntityRefs: []*types.OpenApiReference{&types.OpenApiReference{ID: roleOrgAdmin.Role.ID}},
		ProviderType:   "LOCAL",
		OrgEntityRef:   &types.OpenApiReference{ID: org.TmOrg.ID, Name: org.TmOrg.Name},
	}

	tenantContext := &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	}

	// Create
	createdUser, err := vcd.client.CreateUser(userConfig, tenantContext)
	check.Assert(err, IsNil)
	check.Assert(createdUser, NotNil)
	check.Assert(createdUser.User, NotNil)

	// GetAll
	allUsers, err := vcd.client.GetAllUsers(nil, tenantContext)
	check.Assert(err, IsNil)
	check.Assert(len(allUsers), Equals, 1)
	check.Assert(allUsers[0].User.ID, Equals, createdUser.User.ID)

	// Get by Username
	byUsername, err := vcd.client.GetUserByName(userConfig.Username, tenantContext)
	check.Assert(err, IsNil)
	check.Assert(byUsername.User.ID, Equals, createdUser.User.ID)

	// Get by ID
	byId, err := vcd.client.GetUserById(createdUser.User.ID, tenantContext)
	check.Assert(err, IsNil)
	check.Assert(byId.User.ID, Equals, createdUser.User.ID)

	// Update
	updateConfig := &types.TmUser{
		ID:             byId.User.ID,
		Username:       "test-user-updated",
		Password:       "CHANGE-ME-UPDATED",
		RoleEntityRefs: []*types.OpenApiReference{&types.OpenApiReference{ID: roleOrgUser.Role.ID}},
		ProviderType:   "LOCAL",
		NameInSource:   userConfig.Username, // previous username must be provided
		OrgEntityRef:   &types.OpenApiReference{ID: org.TmOrg.ID, Name: org.TmOrg.Name},
	}

	updatedUser, err := byId.Update(updateConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedUser, NotNil)
	check.Assert(updatedUser.User.ID, Equals, createdUser.User.ID)

	// Delete
	err = updatedUser.Delete()
	check.Assert(err, IsNil)

	// Get by ID and fail
	byIdNotFound, err := vcd.client.GetUserById(updatedUser.User.ID, tenantContext)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdNotFound, IsNil)
}
