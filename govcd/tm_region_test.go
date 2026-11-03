//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"encoding/json"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmRegion(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	// This test creates these entities itself, so the ones shared with other tests must be gone
	releaseTmSharedFixture(check)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()

	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	r := &types.Region{
		Name: "testtmregion",
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
	err = createdRegion.DeleteWithForce(false)
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetRegionByName(createdRegion.Region.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)

	// Create async
	task, err := vcd.client.CreateRegionAsync(r)
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	byIdAsync, err := vcd.client.GetRegionById(task.Task.Owner.ID)
	check.Assert(err, IsNil)
	check.Assert(byIdAsync, NotNil)
	AddToCleanupListOpenApi(createdRegion.Region.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointRegions+createdRegion.Region.ID)

	err = byIdAsync.DeleteWithForce(true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_TmRegionCapabilitiesAndCompatibleZones(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	region, cleanup := tmTestRegion(vcd, check)
	defer cleanup()
	zones, err := region.GetAllZones(nil)
	check.Assert(err, IsNil)
	zoneNames := map[string]bool{}
	for _, zone := range zones {
		zoneNames[zone.Zone.Name] = true
	}
	groups, err := region.GetAllNamespaceCompatibleZoneGroups(url.Values{"pageSize": {"1"}})
	check.Assert(err, IsNil)
	for _, group := range groups {
		check.Assert(len(group.Zones) > 0, Equals, true)
		for _, zone := range group.Zones {
			check.Assert(zoneNames[zone], Equals, true, Commentf("compatible zone %q is not in the region's zone inventory", zone))
		}
	}
	capabilities, err := region.GetAllCapabilities(url.Values{"pageSize": {"1"}})
	check.Assert(err, IsNil)
	for _, capability := range capabilities {
		check.Assert(capability.Name, Not(Equals), "")
		if len(capability.Value) > 0 {
			check.Assert(json.Valid(capability.Value), Equals, true)
		}
		if len(capability.Settings) > 0 {
			check.Assert(json.Valid(capability.Settings), Equals, true)
		}
	}
}
