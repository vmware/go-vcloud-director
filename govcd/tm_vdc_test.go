//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmVdc(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	vdc, err := vcd.client.GetTmVdcByName(vcd.config.Tm.Vdc)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Get by ID
	vdcById, err := vcd.client.GetTmVdcById(vdc.TmVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(vdcById, NotNil)

	check.Assert(vdc.TmVdc, DeepEquals, vdcById.TmVdc)

	allTmVdc, err := vcd.client.GetAllTmVdcs(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allTmVdc) > 0, Equals, true)
}
