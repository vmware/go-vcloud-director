//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmInfraPolicyInventory(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := tmTestOptions(vcd, check, "40.1")
	all, err := vcd.client.GetAllTmInfraPolicies(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(all) == 0 {
		check.Skip("no InfraPolicy resources available")
	}
	seen := make(map[string]bool)
	for _, item := range all {
		id := item.TmInfraPolicy.ID
		check.Assert(id, Not(Equals), "")
		check.Assert(seen[id], Equals, false)
		seen[id] = true
		byID, err := vcd.client.GetTmInfraPolicyById(id, options)
		check.Assert(err, IsNil)
		check.Assert(byID.TmInfraPolicy.ID, Equals, id)
		statuses, err := item.GetAllVCenterPolicies(url.Values{"pageSize": {"1"}})
		check.Assert(err, IsNil)
		for _, status := range statuses {
			check.Assert(status.VirtualCenterRef, NotNil)
		}
	}
}
