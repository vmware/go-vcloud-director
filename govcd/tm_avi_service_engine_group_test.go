//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmAviServiceEngineGroupInventory(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := tmTestOptions(vcd, check, "40.1")
	all, err := vcd.client.GetAllTmAviServiceEngineGroups(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(all) == 0 {
		check.Skip("no AviServiceEngineGroup resources available")
	}
	names := tmNameCounts(all, func(e *TmAviServiceEngineGroup) string { return e.TmAviServiceEngineGroup.Name })
	seen := make(map[string]bool)
	for _, item := range all {
		id := item.TmAviServiceEngineGroup.ID
		check.Assert(id, Not(Equals), "")
		check.Assert(seen[id], Equals, false)
		seen[id] = true
		byID, err := vcd.client.GetTmAviServiceEngineGroupById(id, options)
		check.Assert(err, IsNil)
		check.Assert(byID.TmAviServiceEngineGroup.ID, Equals, id)
		if name := item.TmAviServiceEngineGroup.Name; name != "" && names[name] == 1 {
			byName, err := vcd.client.GetTmAviServiceEngineGroupByName(name, options)
			check.Assert(err, IsNil)
			check.Assert(byName.TmAviServiceEngineGroup.ID, Equals, id)
			byNameOrId, err := vcd.client.GetTmAviServiceEngineGroupByNameOrId(name, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmAviServiceEngineGroup.ID, Equals, id)
			byNameOrId, err = vcd.client.GetTmAviServiceEngineGroupByNameOrId(id, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmAviServiceEngineGroup.ID, Equals, id)
		}
	}
}

func (vcd *TestVCD) Test_TmAviServiceEngineGroupAssignmentInventory(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := tmTestOptions(vcd, check, "40.1")
	all, err := vcd.client.GetAllTmAviServiceEngineGroupAssignments(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(all) == 0 {
		check.Skip("no AviServiceEngineGroupAssignment resources available")
	}
	seen := make(map[string]bool)
	for _, item := range all {
		id := item.TmAviServiceEngineGroupAssignment.ID
		check.Assert(id, Not(Equals), "")
		check.Assert(seen[id], Equals, false)
		seen[id] = true
		byID, err := vcd.client.GetTmAviServiceEngineGroupAssignmentById(id, options)
		check.Assert(err, IsNil)
		check.Assert(byID.TmAviServiceEngineGroupAssignment.ID, Equals, id)
	}
}
