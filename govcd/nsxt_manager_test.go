//go:build network || nsxt || functional || openapi || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"strings"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtManager(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	vcd.skipIfNotSysAdmin(check)

	nsxtManager, err := vcd.client.GetNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(nsxtManager, NotNil)

	urn, err := nsxtManager.Urn()
	check.Assert(err, IsNil)
	check.Assert(strings.HasPrefix(urn, "urn:vcloud:nsxtmanager:"), Equals, true)

}
