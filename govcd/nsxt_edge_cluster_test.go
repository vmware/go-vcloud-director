//go:build network || nsxt || functional || openapi || ALL
// +build network nsxt functional openapi ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAllNsxtEdgeClusters(check *C) {
	skipNoNsxtConfiguration(vcd, check)

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	nsxtVdc, err := vcd.org.GetVDCByNameOrId(vcd.config.VCD.Nsxt.Vdc, true)
	check.Assert(err, IsNil)

	tier0Router, err := nsxtVdc.GetAllNsxtEdgeClusters(nil)
	check.Assert(err, IsNil)
	check.Assert(tier0Router, NotNil)
	check.Assert(len(tier0Router) > 0, Equals, true)
}

func (vcd *TestVCD) Test_GetNsxtEdgeClusterByName(check *C) {
	skipNoNsxtConfiguration(vcd, check)

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	nsxtVdc, err := vcd.org.GetVDCByNameOrId(vcd.config.VCD.Nsxt.Vdc, true)
	check.Assert(err, IsNil)

	allEdgeClusters, err := nsxtVdc.GetAllNsxtEdgeClusters(nil)
	check.Assert(err, IsNil)
	check.Assert(allEdgeClusters, NotNil)

	edgeCluster, err := nsxtVdc.GetNsxtEdgeClusterByName(allEdgeClusters[0].NsxtEdgeCluster.Name)
	check.Assert(err, IsNil)
	check.Assert(allEdgeClusters, NotNil)
	check.Assert(edgeCluster, DeepEquals, allEdgeClusters[0])

}
