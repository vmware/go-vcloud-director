//go:build vsphere || functional || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
	"strings"
)

func (vcd *TestVCD) Test_GetStorageProfiles(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("this test requires system administrator privileges")
	}
	vcenters, err := vcd.client.GetAllVCenters(nil)
	check.Assert(err, IsNil)

	check.Assert(len(vcenters) > 0, Equals, true)

	vc := vcenters[0]

	if vcd.config.Vsphere.ResourcePoolForVcd1 == "" {
		check.Skip("no resource pool found for this VCD")
	}

	resourcePool, err := vc.GetResourcePoolByName(vcd.config.Vsphere.ResourcePoolForVcd1)
	check.Assert(err, IsNil)

	allStorageProfiles, err := vc.GetAllStorageProfiles(resourcePool.ResourcePool.Moref, nil)
	check.Assert(err, IsNil)

	for i, sp := range allStorageProfiles {
		spById, err := vc.GetStorageProfileById(resourcePool.ResourcePool.Moref, sp.StorageProfile.Moref)
		check.Assert(err, IsNil)
		check.Assert(spById.StorageProfile.Moref, Equals, sp.StorageProfile.Moref)
		check.Assert(spById.StorageProfile.Name, Equals, sp.StorageProfile.Name)
		spByName, err := vc.GetStorageProfileByName(resourcePool.ResourcePool.Moref, sp.StorageProfile.Name)
		if err != nil && strings.Contains(err.Error(), "more than one") {
			fmt.Printf("%s\n", err)
			continue
		}
		check.Assert(err, IsNil)
		check.Assert(spByName.StorageProfile.Moref, Equals, sp.StorageProfile.Moref)
		check.Assert(spByName.StorageProfile.Name, Equals, sp.StorageProfile.Name)
		if testVerbose {
			fmt.Printf("%2d %# v\n", i, pretty.Formatter(sp.StorageProfile))
		}
	}
}
