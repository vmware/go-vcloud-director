// +build concurrent

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"sync"

	. "gopkg.in/check.v1"
)

// Test_VAPPRefreshConcurrent is meant to prove and show that structures across
// go-vcloud-director are not thread-safe. Go tests must be run with -race flag
// to capture race condition. It is also guarded by `concurrent` build tag and
// is not run by default.
//
// Run with go test -race -tags "concurrent" -check.vv -check.f Test_VAPPRefreshConcurrent .
func (vcd *TestVCD) Test_VAPPRefreshConcurrent(check *C) {
	var waitgroup sync.WaitGroup

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}
	vapp, err := vcd.vdc.GetVAppByName(ctx, vcd.vapp.VApp.Name, false)
	check.Assert(err, IsNil)

	for counter := 0; counter < 5; counter++ {
		waitgroup.Add(1)
		go func() {
			_ = vapp.Refresh(ctx)
			waitgroup.Done()
			// check.Assert(err, IsNil)
		}()
	}
	waitgroup.Wait()
}
