//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
	"net/url"
	"path/filepath"
)

// TODO: TM: Test upload failures:
//       https://github.com/vmware/go-vcloud-director/pull/717#discussion_r1853767609

// Test_ContentLibraryItemProvider tests CRUD operations for a Content Library Item when uploading in a PROVIDER Content Library
func (vcd *TestVCD) Test_ContentLibraryItemProvider(check *C) {
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

	cl, clCleanup := getOrCreateContentLibrary(vcd, sc, nil, check)
	check.Assert(err, IsNil)
	defer clCleanup()

	testCases := []ContentLibraryItemUploadArguments{
		{
			FilePath: "../test-resources/test_vapp_template.ova",
		},
		{
			FilePath: "../test-resources/test.iso",
		},
		{
			FilePath: "../test-resources/test_vapp_template_ovf/descriptor.ovf",
			OvfFilesPaths: []string{
				"../test-resources/test_vapp_template_ovf/yVMFromVcd-disk1.vmdk",
			},
		},
	}

	for _, args := range testCases {
		testContentLibraryItem(vcd, cl, args, check)
	}

	// Not found errors
	_, err = cl.GetContentLibraryItemByName("notexist")
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = cl.GetContentLibraryItemById("urn:vcloud:contentLibraryItem:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetContentLibraryItemById("urn:vcloud:contentLibraryItem:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(ContainsNotFound(err), Equals, true)
}

// Test_ContentLibraryItemTenant tests CRUD operations for a Content Library Item when uploading in a TENANT Content Library
func (vcd *TestVCD) Test_ContentLibraryItemTenant(check *C) {
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

	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()

	// A Region Quota is needed to have Storage classes available in the Organization
	_, regionQuotaCleanup := createRegionQuota(vcd, org, region, check)
	defer regionQuotaCleanup()

	sc, err := region.GetStorageClassByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sc, NotNil)

	cl, clCleanup := getOrCreateContentLibrary(vcd, sc, org, check)
	check.Assert(err, IsNil)
	defer clCleanup()

	testCases := []ContentLibraryItemUploadArguments{
		{
			FilePath: "../test-resources/test_vapp_template.ova",
		},
		{
			FilePath: "../test-resources/test.iso",
		},
		{
			FilePath: "../test-resources/test_vapp_template_ovf/descriptor.ovf",
			OvfFilesPaths: []string{
				"../test-resources/test_vapp_template_ovf/yVMFromVcd-disk1.vmdk",
			},
		},
	}

	for _, args := range testCases {
		testContentLibraryItem(vcd, cl, args, check)
	}
}

func testContentLibraryItem(vcd *TestVCD, cl *ContentLibrary, args ContentLibraryItemUploadArguments, check *C) {
	cli, err := cl.CreateContentLibraryItem(&types.ContentLibraryItem{
		Name:        check.TestName(),
		Description: check.TestName(),
	}, args)
	check.Assert(err, IsNil)
	check.Assert(cli, NotNil)
	AddToCleanupListOpenApi(cli.ContentLibraryItem.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraryItems+cli.ContentLibraryItem.ID)

	if filepath.Ext(args.FilePath) == ".iso" {
		check.Assert(cli.ContentLibraryItem.ItemType, Equals, "ISO")
	} else {
		check.Assert(cli.ContentLibraryItem.ItemType, Equals, "TEMPLATE")
	}
	check.Assert(cli.ContentLibraryItem.Name, Equals, check.TestName())
	check.Assert(cli.ContentLibraryItem.Description, Equals, check.TestName())
	check.Assert(cli.ContentLibraryItem.Version, Equals, 1)
	check.Assert(cli.ContentLibraryItem.CreationDate, Not(Equals), "")

	// Check that the files are indeed uploaded (list of files is empty)
	files, err := getContentLibraryItemFiles(cli)
	check.Assert(err, IsNil)
	check.Assert(len(files), Equals, 0)

	// Content library deletion should fail with force=false and recursive=false
	err = cl.Delete(false, false)
	check.Assert(err, NotNil)

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
	//check.Assert(updatedCli.ContentLibraryItem.Version, Equals, obtainedCliById.ContentLibraryItem.Version)
	check.Assert(updatedCli.ContentLibraryItem.CreationDate, Equals, obtainedCliById.ContentLibraryItem.CreationDate)
	check.Assert(updatedCli.ContentLibraryItem.ItemType, Equals, obtainedCliById.ContentLibraryItem.ItemType)

	err = cli.DeleteWithForce(filepath.Ext(args.FilePath) != ".iso")
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_ContentLibraryDownloadBundle(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	quotas, err := vcd.client.GetAllRegionQuotas(nil)
	check.Assert(err, IsNil)
	projectID := ""
	var context *TenantContext
	for _, quota := range quotas {
		namespaces, err := quota.GetAllNamespaces(nil)
		check.Assert(err, IsNil)
		for _, ns := range namespaces {
			if ns.ProjectAssignment != nil && ns.ProjectAssignment.ID != "" && quota.TmVdc.Org != nil {
				projectID = ns.ProjectAssignment.ID
				context = &TenantContext{OrgId: quota.TmVdc.Org.ID, OrgName: quota.TmVdc.Org.Name}
				break
			}
		}
		if projectID != "" {
			break
		}
	}
	if projectID == "" {
		check.Skip("no project discoverable from existing region-quota namespaces")
	}
	for _, async := range []bool{false, true} {
		config := &types.ContentLibraryDownloadRequest{BundleName: fmt.Sprintf("sdk-test-%t", async), ProjectID: projectID, MetadataFileContent: "e30="}
		// Both requests generate metadata-only bundles: nil and explicit empty image lists.
		if async {
			config.ImageIdentifiers = []string{}
		}
		options := TmOperationOptions{TenantContext: context}
		var task *Task
		var err error
		if async {
			task, err = vcd.client.GenerateContentLibraryDownloadBundleAsync(config, options)
			tmTestWaitTask(check, task, err)
		} else {
			task, err = vcd.client.GenerateContentLibraryDownloadBundle(config, options)
			check.Assert(err, IsNil)
			check.Assert(task, NotNil)
		}
		check.Assert(task.Task.Status, Equals, "success")
		check.Assert(task.Task.Details, Not(Equals), "")
		download, err := url.Parse(task.Task.Details)
		check.Assert(err, IsNil)
		check.Assert(download.Host, Not(Equals), "")
		check.Assert(download.Scheme == "https" || download.Scheme == "http", Equals, true)
		if !async {
			check.Assert(config.ImageIdentifiers, IsNil)
		}
	}
}
