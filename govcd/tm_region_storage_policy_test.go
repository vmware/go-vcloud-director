//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_RegionStoragePolicy(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	allTmVdc, err := vcd.client.GetAllRegionStoragePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allTmVdc) > 0, Equals, true)
}
