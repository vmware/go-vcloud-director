// +build concurrent

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"sync"

	. "gopkg.in/check.v1"
)

// Test_VMRefreshConcurrent is meant to prove and show that structures across
// go-vcloud-director are not thread-safe. Go tests must be run with -race flag
// to capture race condition. It is also guarded by `concurrent` build tag and
// is not run by default.
//
// Run with go test -race -tags "concurrent" -check.vv -check.f Test_VMRefreshConcurrent .
func (vcd *TestVCD) Test_VMRefreshConcurrent(check *C) {
	var waitgroup sync.WaitGroup

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	for counter := 0; counter < 5; counter++ {
		waitgroup.Add(1)
		go func() {
			_ = vm.Refresh(ctx)
			waitgroup.Done()
			// check.Assert(err, IsNil)
		}()
	}
	waitgroup.Wait()

}
