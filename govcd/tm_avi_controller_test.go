//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmAviControllerInventory(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := tmTestOptions(vcd, check, "40.1")
	all, err := vcd.client.GetAllTmAviControllers(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(all) == 0 {
		check.Skip("no AviController resources available")
	}
	names := tmNameCounts(all, func(e *TmAviController) string { return e.TmAviController.Name })
	seen := make(map[string]bool)
	for _, item := range all {
		id := item.TmAviController.ID
		check.Assert(id, Not(Equals), "")
		check.Assert(seen[id], Equals, false)
		seen[id] = true
		byID, err := vcd.client.GetTmAviControllerById(id, options)
		check.Assert(err, IsNil)
		check.Assert(byID.TmAviController.ID, Equals, id)
		if name := item.TmAviController.Name; name != "" && names[name] == 1 {
			byName, err := vcd.client.GetTmAviControllerByName(name, options)
			check.Assert(err, IsNil)
			check.Assert(byName.TmAviController.ID, Equals, id)
			byNameOrId, err := vcd.client.GetTmAviControllerByNameOrId(name, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmAviController.ID, Equals, id)
			byNameOrId, err = vcd.client.GetTmAviControllerByNameOrId(id, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmAviController.ID, Equals, id)
		}
	}
}
