//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

// Test_ContentLibraryItem tests CRUD operations for a Content Library Item
func (vcd *TestVCD) Test_ContentLibraryItem(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	cl, err := vcd.client.GetContentLibraryByName("adam-lib-1")
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)

	cli, err := cl.CreateContentLibraryItem(&types.ContentLibraryItem{
		Name:        "adam-8",
		Description: "testing for Terraform provider",
	}, "../test-resources/test_vapp_template.ova")
	check.Assert(err, IsNil)
	check.Assert(cli, NotNil)
}
