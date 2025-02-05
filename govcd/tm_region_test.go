//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmRegion(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()

	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	r := &types.Region{
		Name: check.TestName(),
		NsxManager: &types.OpenApiReference{
			ID: nsxtManager.NsxtManagerOpenApi.ID,
		},
		Supervisors: []types.OpenApiReference{
			{
				ID:   supervisor.Supervisor.SupervisorID,
				Name: supervisor.Supervisor.Name,
			},
		},
		StoragePolicies: []string{vcd.config.Tm.VcenterStorageProfile},
	}

	createdRegion, err := vcd.client.CreateRegion(r)
	check.Assert(err, IsNil)
	check.Assert(createdRegion.Region, NotNil)
	AddToCleanupListOpenApi(createdRegion.Region.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointRegions+createdRegion.Region.ID)

	check.Assert(createdRegion.Region.Status, Equals, "READY") // Region is operational

	// Get By Name
	byName, err := vcd.client.GetRegionByName(r.Name)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := vcd.client.GetRegionById(createdRegion.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	check.Assert(byName.Region, DeepEquals, byId.Region)

	// Get All
	allRegions, err := vcd.client.GetAllRegions(nil)
	check.Assert(err, IsNil)
	check.Assert(allRegions, NotNil)
	check.Assert(len(allRegions) > 0, Equals, true)

	// Update
	createdRegion.Region.Description = "new-description"
	updated, err := createdRegion.Update(createdRegion.Region)
	check.Assert(err, IsNil)
	check.Assert(updated, NotNil)
	check.Assert(updated.Region.Description, Equals, "new-description")

	// Delete
	err = createdRegion.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetRegionByName(createdRegion.Region.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)
}
