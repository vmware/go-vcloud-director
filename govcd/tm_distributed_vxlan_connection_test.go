//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmDistributedVxlanConnectionInventory(check *C) {
	requireTmTestVersion(vcd, check, "9.1.1")
	options := tmTestOptions(vcd, check, "9.1.1")
	all, err := vcd.client.GetAllTmDistributedVxlanConnections(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(all) == 0 {
		check.Skip("no DistributedVxlanConnection resources available")
	}
	seen := make(map[string]bool)
	for _, item := range all {
		id := item.TmDistributedVxlanConnection.ID
		check.Assert(id, Not(Equals), "")
		check.Assert(seen[id], Equals, false)
		seen[id] = true
		byID, err := vcd.client.GetTmDistributedVxlanConnectionById(id, options)
		check.Assert(err, IsNil)
		check.Assert(byID.TmDistributedVxlanConnection.ID, Equals, id)
	}
}
