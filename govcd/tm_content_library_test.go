//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_ContentLibrary(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	rsp, err := vcd.client.GetRegionStoragePolicyByName("vSAN Default Storage Policy")
	check.Assert(err, IsNil)
	check.Assert(rsp, NotNil)

	cls, err := vcd.client.GetAllContentLibraries(nil)
	check.Assert(err, IsNil)

	currentContentLibraries := len(cls)

	cl, err := vcd.client.CreateContentLibrary(&types.ContentLibrary{
		Name:            "abarreiro",
		StoragePolicies: []*types.OpenApiReference{{ID: rsp.RegionStoragePolicy.ID}},
		AutoAttach:      false,
		Description:     "Terraform test",
		IsShared:        false,
		IsSubscribed:    false,
		Org:             nil,
	})
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)

	cls, err = vcd.client.GetAllContentLibraries(nil)
	check.Assert(err, IsNil)
	check.Assert(len(cls), Equals, currentContentLibraries+1)

	err = cl.Delete()
	check.Assert(err, IsNil)
}
