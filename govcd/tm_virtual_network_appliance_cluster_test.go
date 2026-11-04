//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmVirtualNetworkApplianceClusterInventory(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := tmTestOptions(vcd, check, "40.1")
	all, err := vcd.client.GetAllTmVirtualNetworkApplianceClusters(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(all) == 0 {
		check.Skip("no VirtualNetworkApplianceCluster resources available")
	}
	names := tmNameCounts(all, func(e *TmVirtualNetworkApplianceCluster) string { return e.TmVirtualNetworkApplianceCluster.Name })
	seen := make(map[string]bool)
	for _, item := range all {
		id := item.TmVirtualNetworkApplianceCluster.ID
		check.Assert(id, Not(Equals), "")
		check.Assert(seen[id], Equals, false)
		seen[id] = true
		byID, err := vcd.client.GetTmVirtualNetworkApplianceClusterById(id, options)
		check.Assert(err, IsNil)
		check.Assert(byID.TmVirtualNetworkApplianceCluster.ID, Equals, id)
		if name := item.TmVirtualNetworkApplianceCluster.Name; name != "" && names[name] == 1 {
			byName, err := vcd.client.GetTmVirtualNetworkApplianceClusterByName(name, options)
			check.Assert(err, IsNil)
			check.Assert(byName.TmVirtualNetworkApplianceCluster.ID, Equals, id)
			byNameOrId, err := vcd.client.GetTmVirtualNetworkApplianceClusterByNameOrId(name, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmVirtualNetworkApplianceCluster.ID, Equals, id)
			byNameOrId, err = vcd.client.GetTmVirtualNetworkApplianceClusterByNameOrId(id, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmVirtualNetworkApplianceCluster.ID, Equals, id)
		}
		statuses, err := item.GetAllTransportNodeStatuses()
		check.Assert(err, IsNil)
		check.Assert(statuses, NotNil)
	}
}
