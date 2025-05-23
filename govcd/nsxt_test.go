//go:build ALL || openapi || functional || nsxt

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_QueryNsxtManagerByName(check *C) {
	vcd.skipIfNotSysAdmin(check)
	skipNoNsxtConfiguration(vcd, check)
	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers), Equals, 1)
}

func (vcd *TestVCD) Test_GetAllNsxtTier0Routers(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)

	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers), Equals, 1)

	uuid, err := GetUuidFromHref(nsxtManagers[0].HREF, true)
	check.Assert(err, IsNil)
	urn, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", uuid)
	check.Assert(err, IsNil)

	tier0Router, err := vcd.client.GetImportableNsxtTier0RouterByName(vcd.config.VCD.Nsxt.Tier0router, urn)
	check.Assert(err, IsNil)
	check.Assert(tier0Router, NotNil)
}
