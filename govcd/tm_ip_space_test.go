//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmIpSpace(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()

	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	ipSpaceType := &types.TmIpSpace{
		Name:        check.TestName(),
		RegionRef:   types.OpenApiReference{ID: region.Region.ID},
		Description: check.TestName(),
		DefaultQuota: types.TmIpSpaceDefaultQuota{
			MaxCidrCount:  3,
			MaxIPCount:    -1,
			MaxSubnetSize: 24,
		},
		ExternalScopeCidr: "12.12.0.0/30",
		InternalScopeCidrBlocks: []types.TmIpSpaceInternalScopeCidrBlocks{
			{
				Cidr: "10.0.0.0/24",
			},
		},
	}

	createdIpSpace, err := vcd.client.CreateTmIpSpace(ipSpaceType)
	check.Assert(err, IsNil)
	check.Assert(createdIpSpace, NotNil)
	// Add to cleanup list
	PrependToCleanupListOpenApi(createdIpSpace.TmIpSpace.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmIpSpaces+createdIpSpace.TmIpSpace.ID)

	// Get TM VDC By Name
	byName, err := vcd.client.GetTmIpSpaceByName(ipSpaceType.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.TmIpSpace, DeepEquals, createdIpSpace.TmIpSpace)

	// Get TM VDC By Id
	byId, err := vcd.client.GetTmIpSpaceById(createdIpSpace.TmIpSpace.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmIpSpace, DeepEquals, createdIpSpace.TmIpSpace)

	// Get By Name and Org ID
	byNameAndOrgId, err := vcd.client.GetTmIpSpaceByNameAndRegionId(createdIpSpace.TmIpSpace.Name, region.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(byNameAndOrgId.TmIpSpace, DeepEquals, createdIpSpace.TmIpSpace)

	// Get By Name and Org ID in non existent Region
	byNameAndInvalidOrgId, err := vcd.client.GetTmIpSpaceByNameAndRegionId(createdIpSpace.TmIpSpace.Name, "urn:vcloud:region:a93c9db9-0000-0000-0000-a8f7eeda85f9")
	check.Assert(err, NotNil)
	check.Assert(byNameAndInvalidOrgId, IsNil)

	// Not Found tests
	byNameInvalid, err := vcd.client.GetTmIpSpaceByName("fake-name")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byNameInvalid, IsNil)

	byIdInvalid, err := vcd.client.GetTmIpSpaceById("urn:vcloud:ipSpace:5344b964-0000-0000-0000-d554913db643")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdInvalid, IsNil)

	// Update
	createdIpSpace.TmIpSpace.Name = check.TestName() + "-update"
	updatedVdc, err := createdIpSpace.Update(createdIpSpace.TmIpSpace)
	check.Assert(err, IsNil)
	check.Assert(updatedVdc.TmIpSpace, DeepEquals, createdIpSpace.TmIpSpace)

	// Delete
	err = createdIpSpace.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetTmIpSpaceByName(createdIpSpace.TmIpSpace.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)

	// Create async
	task, err := vcd.client.CreateTmIpSpaceAsync(ipSpaceType)
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	byIdAsync, err := vcd.client.GetTmIpSpaceById(task.Task.Owner.ID)
	check.Assert(err, IsNil)
	check.Assert(byIdAsync, NotNil)
	PrependToCleanupListOpenApi(byIdAsync.TmIpSpace.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmIpSpaces+createdIpSpace.TmIpSpace.ID)

	err = byIdAsync.Delete()
	check.Assert(err, IsNil)
}
