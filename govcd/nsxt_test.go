// +build ALL openapi functional nsxt

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_QueryNsxtManagerByName(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(ctx, vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers), Equals, 1)
}

func (vcd *TestVCD) Test_GetAllNsxtTier0Routers(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)

	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(ctx, vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers), Equals, 1)

	uuid, err := GetUuidFromHref(nsxtManagers[0].HREF, true)
	check.Assert(err, IsNil)
	urn, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", uuid)
	check.Assert(err, IsNil)

	tier0Router, err := vcd.client.GetImportableNsxtTier0RouterByName(ctx, vcd.config.VCD.Nsxt.Tier0router, urn)
	check.Assert(err, IsNil)
	check.Assert(tier0Router, NotNil)
}
