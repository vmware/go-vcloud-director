//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmRegion(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	region, err := vcd.client.GetRegionByName(vcd.config.Tm.Region)
	check.Assert(err, IsNil)
	check.Assert(region, NotNil)

	// Get by ID
	regionById, err := vcd.client.GetRegionById(region.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(regionById, NotNil)

	check.Assert(region.Region, DeepEquals, regionById.Region)

	allTmVdc, err := vcd.client.GetAllRegions(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allTmVdc) > 0, Equals, true)
}
