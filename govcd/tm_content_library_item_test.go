//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

// TODO: TM: Test upload failures:
//       https://github.com/vmware/go-vcloud-director/pull/717#discussion_r1853767609

// Test_ContentLibraryItemOva tests CRUD operations for a Content Library Item when uploading an OVA
func (vcd *TestVCD) Test_ContentLibraryItemOva(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	// Pre-requisites
	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	sc, err := region.GetStorageClassByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sc, NotNil)

	cl, clCleanup := getOrCreateContentLibrary(vcd, sc, check)
	check.Assert(err, IsNil)
	defer clCleanup()

	// Test begins
	cli, err := cl.CreateContentLibraryItem(&types.ContentLibraryItem{
		Name:        check.TestName(),
		Description: check.TestName(),
	}, ContentLibraryItemUploadArguments{
		FilePath: "../test-resources/test_vapp_template.ova",
	})
	check.Assert(err, IsNil)
	check.Assert(cli, NotNil)
	AddToCleanupListOpenApi(cli.ContentLibraryItem.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraryItems+cli.ContentLibraryItem.ID)
	check.Assert(cli.ContentLibraryItem.ItemType, Equals, "TEMPLATE")
	check.Assert(cli.ContentLibraryItem.Name, Equals, check.TestName())
	check.Assert(cli.ContentLibraryItem.Description, Equals, check.TestName())
	check.Assert(cli.ContentLibraryItem.Version, Equals, 1)
	check.Assert(cli.ContentLibraryItem.CreationDate, Not(Equals), "")

	// Content library deletion should fail with force=false and recursive=false
	err = cl.Delete(false, false)
	check.Assert(err, NotNil)

	// Defer deletion for a correct cleanup
	defer func() {
		err = cli.Delete()
		check.Assert(err, IsNil)
	}()

	allClis, err := cl.GetAllContentLibraryItems(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allClis) > 0, Equals, true)
	found := -1
	for i, k := range allClis {
		if k.ContentLibraryItem.ID == cli.ContentLibraryItem.ID {
			found = i
		}
	}
	check.Assert(found, Not(Equals), -1)
	check.Assert(allClis[found], NotNil)
	check.Assert(*allClis[found].ContentLibraryItem, DeepEquals, *cli.ContentLibraryItem)

	obtainedCliByName, err := cl.GetContentLibraryItemByName(check.TestName())
	check.Assert(err, IsNil)
	check.Assert(obtainedCliByName, NotNil)
	check.Assert(*obtainedCliByName.ContentLibraryItem, DeepEquals, *cli.ContentLibraryItem)

	obtainedCliById, err := cl.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
	check.Assert(err, IsNil)
	check.Assert(obtainedCliById, NotNil)
	check.Assert(*obtainedCliById.ContentLibraryItem, DeepEquals, *obtainedCliByName.ContentLibraryItem)

	obtainedCliById, err = vcd.client.GetContentLibraryItemById(cli.ContentLibraryItem.ID)
	check.Assert(err, IsNil)
	check.Assert(obtainedCliById, NotNil)
	check.Assert(*obtainedCliById.ContentLibraryItem, DeepEquals, *obtainedCliByName.ContentLibraryItem)

	updatedCli, err := cli.Update(&types.ContentLibraryItem{
		Name:        obtainedCliById.ContentLibraryItem.Name + "Updated",
		Description: obtainedCliById.ContentLibraryItem.Description + "Updated",
		ItemType:    obtainedCliById.ContentLibraryItem.ItemType, // We need to send the type, otherwise it fails
	})
	check.Assert(err, IsNil)
	check.Assert(updatedCli, NotNil)
	check.Assert(updatedCli.ContentLibraryItem.Name, Equals, obtainedCliById.ContentLibraryItem.Name+"Updated")
	check.Assert(updatedCli.ContentLibraryItem.Description, Equals, obtainedCliById.ContentLibraryItem.Description+"Updated")
	check.Assert(updatedCli.ContentLibraryItem.Version, Equals, obtainedCliById.ContentLibraryItem.Version)
	check.Assert(updatedCli.ContentLibraryItem.CreationDate, Equals, obtainedCliById.ContentLibraryItem.CreationDate)
	check.Assert(updatedCli.ContentLibraryItem.ItemType, Equals, obtainedCliById.ContentLibraryItem.ItemType)

	// Not found errors
	_, err = cl.GetContentLibraryItemByName("notexist")
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = cl.GetContentLibraryItemById("urn:vcloud:contentLibraryItem:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetContentLibraryItemById("urn:vcloud:contentLibraryItem:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(ContainsNotFound(err), Equals, true)
}

// Test_ContentLibraryItemIso tests CRUD operations for a Content Library Item when uploading an ISO file
func (vcd *TestVCD) Test_ContentLibraryItemIso(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	// Pre-requisites
	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	sc, err := region.GetStorageClassByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sc, NotNil)

	cl, clCleanup := getOrCreateContentLibrary(vcd, sc, check)
	check.Assert(err, IsNil)
	defer clCleanup()

	cli, err := cl.CreateContentLibraryItem(&types.ContentLibraryItem{
		Name:        check.TestName(),
		Description: check.TestName(),
	}, ContentLibraryItemUploadArguments{
		FilePath: "../test-resources/test.iso",
	})
	check.Assert(err, IsNil)
	check.Assert(cli, NotNil)
	AddToCleanupListOpenApi(cli.ContentLibraryItem.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraryItems+cli.ContentLibraryItem.ID)
	check.Assert(cli.ContentLibraryItem.ItemType, Equals, "ISO")
	check.Assert(cli.ContentLibraryItem.Name, Equals, check.TestName())
	check.Assert(cli.ContentLibraryItem.Description, Equals, check.TestName())
	check.Assert(cli.ContentLibraryItem.Version, Equals, 1)
	check.Assert(cli.ContentLibraryItem.CreationDate, Not(Equals), "")
}

// Test_ContentLibraryItemIso tests CRUD operations for a Content Library Item when uploading an OVF file
func (vcd *TestVCD) Test_ContentLibraryItemOvf(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	// Pre-requisites
	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	sc, err := region.GetStorageClassByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sc, NotNil)

	cl, clCleanup := getOrCreateContentLibrary(vcd, sc, check)
	check.Assert(err, IsNil)
	defer clCleanup()

	cli, err := cl.CreateContentLibraryItem(&types.ContentLibraryItem{
		Name:        check.TestName(),
		Description: check.TestName(),
	}, ContentLibraryItemUploadArguments{
		FilePath: "../test-resources/test_vapp_template_ovf/descriptor.ovf",
		OvfFilesPaths: []string{
			"../test-resources/test_vapp_template_ovf/yVMFromVcd-disk1.vmdk",
		},
	})
	check.Assert(err, IsNil)
	check.Assert(cli, NotNil)
	AddToCleanupListOpenApi(cli.ContentLibraryItem.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraryItems+cli.ContentLibraryItem.ID)
	check.Assert(cli.ContentLibraryItem.ItemType, Equals, "TEMPLATE")
	check.Assert(cli.ContentLibraryItem.Name, Equals, check.TestName())
	check.Assert(cli.ContentLibraryItem.Description, Equals, check.TestName())
	check.Assert(cli.ContentLibraryItem.Version, Equals, 1)
	check.Assert(cli.ContentLibraryItem.CreationDate, Not(Equals), "")
}
