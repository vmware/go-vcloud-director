//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmVCenterComputePolicyInventory(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := tmTestOptions(vcd, check, "40.1")
	all, err := vcd.client.GetAllTmVCenterComputePolicies(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(all) == 0 {
		check.Skip("no VCenterComputePolicy resources available")
	}
	seen := make(map[string]bool)
	for _, item := range all {
		id := item.TmVCenterComputePolicy.ID
		check.Assert(id, Not(Equals), "")
		check.Assert(seen[id], Equals, false)
		seen[id] = true
		byID, err := vcd.client.GetTmVCenterComputePolicyById(id, options)
		check.Assert(err, IsNil)
		check.Assert(byID.TmVCenterComputePolicy.ID, Equals, id)
		tags, err := item.GetAllTags(url.Values{"pageSize": {"1"}})
		check.Assert(err, IsNil)
		for _, tag := range tags {
			check.Assert(tag.ID, Not(Equals), "")
		}
	}
}

func (vcd *TestVCD) Test_TmVCenterComputePolicyNames(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := tmTestOptions(vcd, check, "40.1")

	names, err := vcd.client.GetAllTmVCenterComputePolicyNames(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(names) == 0 {
		check.Skip("no vCenter compute policy names available")
	}
	for _, name := range names {
		check.Assert(name.Name, Not(Equals), "")
	}
}
