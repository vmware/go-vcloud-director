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

	// Pre-requisites
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	sp, err := region.GetStoragePolicyByName(vcd.config.Tm.RegionStoragePolicy)
	check.Assert(err, IsNil)
	check.Assert(sp, NotNil)

	cl, clCleanup := getOrCreateContentLibrary(vcd, sp, check)
	check.Assert(err, IsNil)
	defer clCleanup()

	// Test begins
	cli, err := cl.CreateContentLibraryItem(&types.ContentLibraryItem{
		Name:        check.TestName(),
		Description: check.TestName(),
	}, "../test-resources/test_vapp_template.ova")
	check.Assert(err, IsNil)
	check.Assert(cli, NotNil)
	AddToCleanupListOpenApi(cli.ContentLibraryItem.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraryItems+cli.ContentLibraryItem.ID)

	// Defer deletion for a correct cleanup
	defer func() {
		err = cli.Delete()
		check.Assert(err, IsNil)
	}()

	obtainedCliByName, err := cl.GetContentLibraryItemByName(check.TestName())
	check.Assert(err, IsNil)
	check.Assert(obtainedCliByName, NotNil)

	obtainedCliById, err := cl.GetContentLibraryItemById(check.TestName())
	check.Assert(err, IsNil)
	check.Assert(obtainedCliById, NotNil)
	check.Assert(*obtainedCliById.ContentLibraryItem, DeepEquals, *obtainedCliByName.ContentLibraryItem)

	obtainedCliById, err = vcd.client.GetContentLibraryItemById(check.TestName())
	check.Assert(err, IsNil)
	check.Assert(obtainedCliById, NotNil)
	check.Assert(*obtainedCliById.ContentLibraryItem, DeepEquals, *obtainedCliByName.ContentLibraryItem)

	files, err := obtainedCliById.GetFiles(nil)
	check.Assert(err, IsNil)
	check.Assert(len(files), Not(Equals), 0)

	// Not found errors
	_, err = cl.GetContentLibraryItemByName("notexist")
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = cl.GetContentLibraryItemById("urn:vcloud:contentLibraryItem:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetContentLibraryItemById("urn:vcloud:contentLibraryItem:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(ContainsNotFound(err), Equals, true)
}
