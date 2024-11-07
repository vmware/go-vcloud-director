//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
	"strings"
)

// TODO: TM: Tests missing: Tenant, subscribed catalog, shared catalog

// Test_ContentLibraryProvider tests CRUD operations for a Content Library with the Provider user
func (vcd *TestVCD) Test_ContentLibraryProvider(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	cls, err := vcd.client.GetAllContentLibraries(nil)
	check.Assert(err, IsNil)
	existingContentLibraryCount := len(cls)

	rsp, err := region.GetStoragePolicyByName(vcd.config.Tm.RegionStoragePolicy)
	check.Assert(err, IsNil)
	check.Assert(rsp, NotNil)

	clDefinition := &types.ContentLibrary{
		Name:           check.TestName(),
		StorageClasses: []types.OpenApiReference{{ID: rsp.RegionStoragePolicy.ID}},
		AutoAttach:     true, // TODO: TM: Test with false, still does not work
		Description:    check.TestName(),
	}

	createdCl, err := vcd.client.CreateContentLibrary(clDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdCl, NotNil)
	AddToCleanupListOpenApi(createdCl.ContentLibrary.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraries+createdCl.ContentLibrary.ID)

	// Defer deletion for a correct cleanup
	defer func() {
		err = createdCl.Delete()
		check.Assert(err, IsNil)
	}()
	check.Assert(isUrn(createdCl.ContentLibrary.ID), Equals, true)
	check.Assert(createdCl.ContentLibrary.Name, Equals, clDefinition.Name)
	check.Assert(createdCl.ContentLibrary.Description, Equals, clDefinition.Description)
	check.Assert(len(createdCl.ContentLibrary.StorageClasses), Equals, 1)
	check.Assert(createdCl.ContentLibrary.StorageClasses[0].ID, Equals, strings.ReplaceAll(rsp.RegionStoragePolicy.ID, "regionStoragePolicy", "storageClass")) // TODO: TM: Revisit this at some point
	check.Assert(createdCl.ContentLibrary.AutoAttach, Equals, clDefinition.AutoAttach)
	// "Computed" values
	check.Assert(createdCl.ContentLibrary.IsShared, Equals, true) // TODO: TM: Still not used in UI
	check.Assert(createdCl.ContentLibrary.IsSubscribed, Equals, false)
	check.Assert(createdCl.ContentLibrary.LibraryType, Equals, "PROVIDER") // TODO: TM: Test with Tenant once implemented
	check.Assert(createdCl.ContentLibrary.VersionNumber, Equals, int64(1))
	check.Assert(createdCl.ContentLibrary.Org, NotNil)
	check.Assert(createdCl.ContentLibrary.Org.Name, Equals, "System")
	check.Assert(createdCl.ContentLibrary.SubscriptionConfig, IsNil)
	check.Assert(createdCl.ContentLibrary.CreationDate, Not(Equals), "")

	cls, err = vcd.client.GetAllContentLibraries(nil)
	check.Assert(err, IsNil)
	check.Assert(len(cls), Equals, existingContentLibraryCount+1)
	for _, l := range cls {
		if l.ContentLibrary.ID == createdCl.ContentLibrary.ID {
			check.Assert(*l.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)
			break
		}
	}

	cl, err := vcd.client.GetContentLibraryByName(check.TestName())
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)
	check.Assert(*cl.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)

	cl, err = vcd.client.GetContentLibraryById(cl.ContentLibrary.ID)
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)
	check.Assert(*cl.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)
}

// getOrCreateRegion will check configuration file and create a Region if
// stated in the 'createRegion' testing property not present in TM.
// Otherwise, it just retrieves it
func getOrCreateRegion(vcd *TestVCD, nsxtManager *NsxtManagerOpenApi, supervisor *Supervisor, check *C) (*Region, func()) {
	region, err := vcd.client.GetRegionByName(vcd.config.Tm.Region)
	if err == nil {
		return region, func() {}
	}
	if !ContainsNotFound(err) {
		check.Fatal(err)
		return nil, nil
	}
	if !vcd.config.Tm.CreateRegion {
		check.Skip("Region is not configured and configuration is not allowed in config file")
		return nil, nil
	}
	if nsxtManager == nil || supervisor == nil {
		check.Fatalf("getOrCreateRegion requires a not nil NSX-T Manager and Supervisor")
	}

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
		IsEnabled:       true,
	}

	region, err = vcd.client.CreateRegion(r)
	check.Assert(err, IsNil)
	check.Assert(region, NotNil)
	regionCreated := true
	AddToCleanupListOpenApi(region.Region.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointRegions+region.Region.ID)

	return region, func() {
		if !regionCreated {
			return
		}
		err = region.Delete()
		check.Assert(err, IsNil)
	}

}
