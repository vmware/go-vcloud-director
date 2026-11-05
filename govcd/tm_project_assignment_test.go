//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmProjectAssignmentLifecycle(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	org, cleanupOrg := createOrg(vcd, check, false)
	defer cleanupOrg()
	options := TmOperationOptions{TenantContext: &TenantContext{OrgId: org.TmOrg.ID, OrgName: org.TmOrg.Name}}

	for _, async := range []bool{false, true} {
		func() {
			uuid, err := getPseudoUuid()
			check.Assert(err, IsNil)
			config := &types.TmProjectAssignment{ID: "urn:vcloud:projectAssignment:" + strings.ToLower(uuid), Name: check.TestName()}
			var project *TmProjectAssignment
			if async {
				task, err := vcd.client.CreateTmProjectAssignmentAsync(config, options)
				tmTestWaitTask(check, task, err)
				project, err = vcd.client.GetTmProjectAssignmentById(config.ID, options)
				check.Assert(err, IsNil)
			} else {
				project, err = vcd.client.CreateTmProjectAssignment(config, options)
				check.Assert(err, IsNil)
			}
			deleted := false
			defer func() {
				if !deleted {
					check.Check(project.Delete(), IsNil)
				}
			}()
			check.Assert(project.TmProjectAssignment.ID, Equals, config.ID)
			byName, err := vcd.client.GetTmProjectAssignmentByName(config.Name, options)
			check.Assert(err, IsNil)
			check.Assert(byName.TmProjectAssignment.ID, Equals, config.ID)
			byNameOrId, err := vcd.client.GetTmProjectAssignmentByNameOrId(config.ID, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmProjectAssignment.ID, Equals, config.ID)
			config.Description = "updated project description"
			updated, err := project.Update(config)
			check.Assert(err, IsNil)
			check.Assert(updated.TmProjectAssignment.Description, Equals, config.Description)
			config.Description = ""
			task, err := updated.UpdateAsync(config)
			tmTestWaitTask(check, task, err)
			refreshed, err := vcd.client.GetTmProjectAssignmentById(config.ID, options)
			check.Assert(err, IsNil)
			check.Assert(refreshed.TmProjectAssignment.Description, Equals, "")
			members := &types.TmProjectAssignmentMembers{Values: []types.TmProjectAssignmentMember{}}
			check.Assert(refreshed.UpdateMembers(members), IsNil)
			task, err = refreshed.UpdateMembersAsync(members)
			tmTestWaitTask(check, task, err)
			actualMembers, err := refreshed.GetAllMembers(url.Values{"pageSize": {"1"}})
			check.Assert(err, IsNil)
			check.Assert(len(actualMembers), Equals, 0)
			projects, err := vcd.client.GetAllTmProjectAssignments(url.Values{"pageSize": {"1"}}, options)
			check.Assert(err, IsNil)
			found := false
			for _, p := range projects {
				if p.TmProjectAssignment.ID == config.ID {
					found = true
				}
			}
			check.Assert(found, Equals, true)
			if async {
				task, err = refreshed.DeleteAsync()
				tmTestWaitTask(check, task, err)
			} else {
				check.Assert(refreshed.Delete(), IsNil)
			}
			deleted = true
			// A deleted project may answer with an error other than "not found" (e.g. access denied),
			// so verify its absence in the listing instead of checking by the project id
			projects, err = vcd.client.GetAllTmProjectAssignments(nil, options)
			check.Assert(err, IsNil)
			for _, p := range projects {
				check.Assert(p.TmProjectAssignment.ID, Not(Equals), config.ID)
			}
		}()
	}
}

func (vcd *TestVCD) Test_TmProjectAssignmentInventory(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	options := tmTestOptions(vcd, check, "40.0")
	all, err := vcd.client.GetAllTmProjectAssignments(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(all) == 0 {
		check.Skip("no ProjectAssignment resources available")
	}
	names := tmNameCounts(all, func(e *TmProjectAssignment) string { return e.TmProjectAssignment.Name })
	seen := make(map[string]bool)
	for _, item := range all {
		id := item.TmProjectAssignment.ID
		check.Assert(id, Not(Equals), "")
		check.Assert(seen[id], Equals, false)
		seen[id] = true
		byID, err := vcd.client.GetTmProjectAssignmentById(id, options)
		check.Assert(err, IsNil)
		check.Assert(byID.TmProjectAssignment.ID, Equals, id)
		if name := item.TmProjectAssignment.Name; name != "" && names[name] == 1 {
			byName, err := vcd.client.GetTmProjectAssignmentByName(name, options)
			check.Assert(err, IsNil)
			check.Assert(byName.TmProjectAssignment.ID, Equals, id)
			byNameOrId, err := vcd.client.GetTmProjectAssignmentByNameOrId(name, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmProjectAssignment.ID, Equals, id)
			byNameOrId, err = vcd.client.GetTmProjectAssignmentByNameOrId(id, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmProjectAssignment.ID, Equals, id)
		}
		members, err := item.GetAllMembers(url.Values{"pageSize": {"1"}})
		check.Assert(err, IsNil)
		for _, member := range members {
			check.Assert(member.Member, NotNil)
		}
	}
}

func (vcd *TestVCD) Test_TmProjectAccessLevels(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	options := tmTestOptions(vcd, check, "40.0")

	levels, err := vcd.client.GetAllTmProjectAccessLevels(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	check.Assert(len(levels) > 0, Equals, true)
	for _, level := range levels {
		check.Assert(level.ID, Not(Equals), "")
	}
}
