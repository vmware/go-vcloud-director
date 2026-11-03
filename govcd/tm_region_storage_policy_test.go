//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
	"net/http"
	"net/url"
	"sort"
)

func (vcd *TestVCD) Test_RegionStoragePolicy(check *C) {
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

	allRegionStoragePolicies, err := vcd.client.GetAllRegionStoragePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allRegionStoragePolicies), Not(Equals), 0)

	rspById, err := vcd.client.GetRegionStoragePolicyById(allRegionStoragePolicies[0].RegionStoragePolicy.ID)
	check.Assert(err, IsNil)
	check.Assert(rspById, NotNil)
	check.Assert(*rspById.RegionStoragePolicy, DeepEquals, *allRegionStoragePolicies[0].RegionStoragePolicy)

	rspByName, err := vcd.client.GetRegionStoragePolicyByName(allRegionStoragePolicies[0].RegionStoragePolicy.Name)
	check.Assert(err, IsNil)
	check.Assert(rspByName, NotNil)
	check.Assert(*rspByName.RegionStoragePolicy, DeepEquals, *rspById.RegionStoragePolicy)

	rspByName2, err := region.GetStoragePolicyByName(allRegionStoragePolicies[0].RegionStoragePolicy.Name)
	check.Assert(err, IsNil)
	check.Assert(rspByName2, NotNil)
	check.Assert(*rspByName2.RegionStoragePolicy, DeepEquals, *allRegionStoragePolicies[0].RegionStoragePolicy)

	// Check ENF errors
	_, err = vcd.client.GetRegionStoragePolicyById("urn:vcloud:regionStoragePolicy:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(err, NotNil)
	// TODO:TM: Right now throws a 500 NPE...
	// check.Assert(ContainsNotFound(err), Equals, true)

	_, err = region.GetStoragePolicyByName("NotExists")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetRegionStoragePolicyByName("NotExists")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
}

func (vcd *TestVCD) Test_RegionStoragePolicyLifecycle(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	region, cleanup := tmTestRegion(vcd, check)
	defer cleanup()
	name := tmTestUnassignedStoragePolicy(vcd, region, check)
	// Compare names locally so punctuation is not interpreted as a FIQL expression.
	query := url.Values{"filter": {"region.id==" + region.Region.ID}, "pageSize": {"1"}}
	policies, err := vcd.client.GetAllRegionStoragePolicies(query)
	check.Assert(err, IsNil)
	config := &types.RegionStoragePolicy{Name: name, Region: &types.OpenApiReference{ID: region.Region.ID}}
	for _, async := range []bool{false, true} {
		func() {
			var policy *RegionStoragePolicy
			if async {
				task, err := vcd.client.CreateRegionStoragePolicyAsync(config)
				tmTestWaitTask(check, task, err)
				check.Assert(task.Task.Owner, NotNil)
				check.Assert(task.Task.Owner.ID, Equals, region.Region.ID)
				policies, err = vcd.client.GetAllRegionStoragePolicies(query)
				check.Assert(err, IsNil)
				for _, candidate := range policies {
					if candidate.RegionStoragePolicy.Name == config.Name {
						check.Assert(policy, IsNil)
						policy = candidate
					}
				}
			} else {
				policy, err = vcd.client.CreateRegionStoragePolicy(config)
				check.Assert(err, IsNil)
			}
			check.Assert(policy, NotNil)
			deleted := false
			defer func() {
				if !deleted {
					check.Check(policy.Delete(), IsNil)
				}
			}()
			AddToCleanupListOpenApi(policy.RegionStoragePolicy.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointRegionStoragePolicies+policy.RegionStoragePolicy.ID)
			check.Assert(policy.RegionStoragePolicy.Name, Equals, config.Name)
			check.Assert(policy.RegionStoragePolicy.Region.ID, Equals, region.Region.ID)
			byID, err := vcd.client.GetRegionStoragePolicyById(policy.RegionStoragePolicy.ID)
			check.Assert(err, IsNil)
			check.Assert(byID.RegionStoragePolicy.Name, Equals, config.Name)
			check.Assert(policy.Delete(), IsNil)
			deleted = true
			policies, err = vcd.client.GetAllRegionStoragePolicies(query)
			check.Assert(err, IsNil)
			for _, candidate := range policies {
				check.Assert(candidate.RegionStoragePolicy.ID, Not(Equals), policy.RegionStoragePolicy.ID)
			}
		}()
	}
}

func (vcd *TestVCD) Test_RegionStoragePolicyRestriction(check *C) {
	requireTmTestVersion(vcd, check, "9.1.2")
	region, cleanup := tmTestRegion(vcd, check)
	defer cleanup()
	name := tmTestUnassignedStoragePolicy(vcd, region, check)
	policy, err := vcd.client.CreateRegionStoragePolicy(&types.RegionStoragePolicy{Name: name, Region: &types.OpenApiReference{ID: region.Region.ID}})
	check.Assert(err, IsNil)
	defer func() { check.Check(policy.Delete(), IsNil) }()
	for _, restricted := range []bool{true, false} {
		config := *policy.RegionStoragePolicy
		config.IsRestricted = &restricted
		updated, err := policy.Update(&config)
		check.Assert(err, IsNil)
		check.Assert(config.ID, Equals, policy.RegionStoragePolicy.ID)
		check.Assert(updated.RegionStoragePolicy.ID, Equals, policy.RegionStoragePolicy.ID)
		check.Assert(updated.RegionStoragePolicy.IsRestricted, NotNil)
		check.Assert(*updated.RegionStoragePolicy.IsRestricted, Equals, restricted)
		// The new request helper negotiates 9.1.2 for the write and its response.
		// Read at the same version to verify persistence of the new property.
		c := crudConfig{
			entityLabel:    types.OpenApiEndpointRegionStoragePolicies,
			endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies,
			requiresTm:     true,
			minimumVersion: "9.1.2",
			expectedStatus: http.StatusOK,
		}
		actual, err := getInnerEntity[types.RegionStoragePolicy](&vcd.client.Client, c.item(policy.RegionStoragePolicy.ID).withTmOptions(nil))
		check.Assert(err, IsNil)
		check.Assert(actual.IsRestricted, NotNil)
		check.Assert(*actual.IsRestricted, Equals, restricted)
	}
}

// Discover a policy available to every supervisor cluster and not already assigned
// to the region. Existing assignments and their restriction flags are never changed.
func tmTestUnassignedStoragePolicy(vcd *TestVCD, region *Region, check *C) string {
	assigned, err := vcd.client.GetAllRegionStoragePolicies(url.Values{"filter": {"region.id==" + region.Region.ID}})
	check.Assert(err, IsNil)
	assignedNames := map[string]bool{}
	for _, policy := range assigned {
		assignedNames[policy.RegionStoragePolicy.Name] = true
	}
	clusters, err := vcd.client.GetAllTmComputeClusters(nil)
	check.Assert(err, IsNil)
	var commonNames map[string]bool
	for _, ref := range region.Region.Supervisors {
		supervisor, err := vcd.client.GetSupervisorById(ref.ID)
		check.Assert(err, IsNil)
		check.Assert(supervisor.Supervisor.VirtualCenter, NotNil)
		vc, err := vcd.client.GetVCenterById(supervisor.Supervisor.VirtualCenter.ID)
		check.Assert(err, IsNil)
		pools, err := vc.GetAllResourcePools(nil)
		check.Assert(err, IsNil)
		foundCluster := false
		for _, cluster := range clusters {
			if cluster.Supervisor == nil || cluster.Supervisor.ID != ref.ID {
				continue
			}
			foundCluster = true
			poolID := ""
			for _, pool := range pools {
				if pool.ResourcePool.ClusterMoref == cluster.Moref {
					poolID = pool.ResourcePool.Moref
					break
				}
			}
			if poolID == "" {
				check.Skip("no resource pool discoverable for a supervisor cluster")
			}
			profiles, err := vc.GetAllStorageProfiles(poolID, nil)
			check.Assert(err, IsNil)
			names := map[string]bool{}
			for _, profile := range profiles {
				if profile.StorageProfile.Name != "" {
					names[profile.StorageProfile.Name] = true
				}
			}
			if commonNames == nil {
				commonNames = names
			} else {
				for name := range commonNames {
					if !names[name] {
						delete(commonNames, name)
					}
				}
			}
		}
		if !foundCluster {
			check.Skip("no compute cluster discoverable for a region supervisor")
		}
	}
	var names []string
	for name := range commonNames {
		if !assignedNames[name] {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		check.Skip("no unassigned storage policy available to all region supervisor clusters")
	}
	sort.Strings(names)
	return names[0]
}
