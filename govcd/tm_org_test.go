//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func (vcd *TestVCD) Test_TmOrg(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	cfg := &types.TmOrg{
		Name:          check.TestName(),
		DisplayName:   check.TestName(),
		CanManageOrgs: false,
	}

	tmOrg, err := vcd.client.CreateTmOrg(cfg)
	check.Assert(err, IsNil)
	check.Assert(tmOrg, NotNil)

	// Add to cleanup list
	PrependToCleanupListOpenApi(tmOrg.TmOrg.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgs+tmOrg.TmOrg.ID)

	// Get By Name
	byName, err := vcd.client.GetTmOrgByName(cfg.Name)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := vcd.client.GetTmOrgById(tmOrg.TmOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	// Get All
	allTmOrgs, err := vcd.client.GetAllTmOrgs(nil)
	check.Assert(err, IsNil)
	check.Assert(allTmOrgs, NotNil)
	check.Assert(len(allTmOrgs) > 0, Equals, true)

	// Update
	tmOrg.TmOrg.IsEnabled = false
	updated, err := tmOrg.Update(tmOrg.TmOrg)
	check.Assert(err, IsNil)
	check.Assert(updated, NotNil)

	// Delete
	err = tmOrg.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetTmOrgByName(cfg.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)
}
