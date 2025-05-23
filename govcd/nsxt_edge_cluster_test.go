//go:build network || nsxt || functional || openapi || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAllNsxtEdgeClusters(check *C) {
	skipNoNsxtConfiguration(vcd, check)

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	nsxtVdc, err := vcd.org.GetVDCByNameOrId(vcd.config.VCD.Nsxt.Vdc, true)
	check.Assert(err, IsNil)

	edgeClusters, err := nsxtVdc.GetAllNsxtEdgeClusters(nil)
	check.Assert(err, IsNil)
	check.Assert(edgeClusters, NotNil)
	check.Assert(len(edgeClusters) > 0, Equals, true)

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("orgVdcId==%s", nsxtVdc.Vdc.ID))
	allEdgeClusters, err := vcd.client.GetAllNsxtEdgeClusters(queryParams)
	check.Assert(err, IsNil)
	check.Assert(allEdgeClusters, NotNil)
	check.Assert(len(allEdgeClusters) > 0, Equals, true)
}

func (vcd *TestVCD) Test_GetNsxtEdgeClusterByName(check *C) {
	skipNoNsxtConfiguration(vcd, check)

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	nsxtVdc, err := vcd.org.GetVDCByNameOrId(vcd.config.VCD.Nsxt.Vdc, true)
	check.Assert(err, IsNil)

	edgeCluster, err := nsxtVdc.GetNsxtEdgeClusterByName(vcd.config.VCD.Nsxt.NsxtEdgeCluster)
	check.Assert(err, IsNil)
	check.Assert(edgeCluster, NotNil)
	check.Assert(edgeCluster.NsxtEdgeCluster.Name, Equals, vcd.config.VCD.Nsxt.NsxtEdgeCluster)

}
