// +build vdc functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_FindVDCNetwork(check *C) {
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}
	fmt.Printf("Running: %s\n", check.TestName())

	net, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, true)

	check.Assert(err, IsNil)
	check.Assert(net, NotNil)
	check.Assert(net.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network.Net1)
	check.Assert(net.OrgVDCNetwork.HREF, Not(Equals), "")

	// find Invalid Network
	net, err = vcd.vdc.GetOrgVdcNetworkByName("INVALID", false)
	check.Assert(err, NotNil)
}

// Tests Network retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_GetOrgVDCNetwork(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_GetOrgVDCNetwork: Org name not given")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_GetOrgVDCNetwork: VDC name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return vdc.GetOrgVdcNetworkByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return vdc.GetOrgVdcNetworkById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return vdc.GetOrgVdcNetworkByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "Vdc",
		parentName:    vcd.config.VCD.Vdc,
		entityType:    "OrgVDCNetwork",
		entityName:    vcd.config.VCD.Network.Net1,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

func (vcd *TestVCD) Test_NewVdc(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	err := vcd.vdc.Refresh()
	check.Assert(err, IsNil)

	check.Assert(vcd.vdc.Vdc.Link[0].Rel, Equals, "up")
	check.Assert(vcd.vdc.Vdc.Link[0].Type, Equals, "application/vnd.vmware.vcloud.org+xml")

	// fmt.Printf("allocation mem %#v\n\n",vcd.vdc.Vdc.AllocationModel)
	for _, resource := range vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity {

		// fmt.Printf("res %#v\n",resource)
		check.Assert(resource.Name, Not(Equals), "")
		check.Assert(resource.Type, Not(Equals), "")
		check.Assert(resource.HREF, Not(Equals), "")
	}

	// TODO: find which values are acceptable for AllocationModel
	// check.Assert(vcd.vdc.Vdc.AllocationModel, Equals, "AllocationPool")

	/*
		// TODO: Find the conditions that define valid ComputeCapacity
		for _, v := range vcd.vdc.Vdc.ComputeCapacity {
			check.Assert(v.CPU.Units, Equals, "MHz")
			check.Assert(v.CPU.Allocated, Equals, int64(30000))
			check.Assert(v.CPU.Limit, Equals, int64(30000))
			check.Assert(v.CPU.Reserved, Equals, int64(15000))
			check.Assert(v.CPU.Used, Equals, int64(0))
			check.Assert(v.CPU.Overhead, Equals, int64(0))
			check.Assert(v.Memory.Units, Equals, "MB")
			check.Assert(v.Memory.Allocated, Equals, int64(61440))
			check.Assert(v.Memory.Limit, Equals, int64(61440))
			check.Assert(v.Memory.Reserved, Equals, int64(61440))
			check.Assert(v.Memory.Used, Equals, int64(6144))
			check.Assert(v.Memory.Overhead, Equals, int64(95))
		}
	*/

	// Skipping this check, as we can't define the existence of a given vApp template beforehand
	/*
		check.Assert(vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity[0].Name, Equals, "vAppTemplate")
		check.Assert(vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity[0].Type, Equals, "application/vnd.vmware.vcloud.vAppTemplate+xml")
		check.Assert(vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity[0].HREF, Equals, "http://localhost:4444/api/vAppTemplate/vappTemplate-22222222-2222-2222-2222-222222222222")
	*/

	for _, availableNetworks := range vcd.vdc.Vdc.AvailableNetworks {
		for _, v2 := range availableNetworks.Network {
			check.Assert(v2.Name, Not(Equals), "")
			check.Assert(v2.Type, Equals, "application/vnd.vmware.vcloud.network+xml")
			check.Assert(v2.HREF, Not(Equals), "")
		}
	}

	/*

		// Skipping this check, as we don't have precise terms of comparison for this entity
		check.Assert(vcd.vdc.Vdc.NicQuota, Equals, 0)
		check.Assert(vcd.vdc.Vdc.NetworkQuota, Equals, 20)
		check.Assert(vcd.vdc.Vdc.UsedNetworkCount, Equals, 0)
		check.Assert(vcd.vdc.Vdc.VMQuota, Equals, 0)
		check.Assert(vcd.vdc.Vdc.IsEnabled, Equals, true)
	*/

	for _, v2 := range vcd.vdc.Vdc.VdcStorageProfiles.VdcStorageProfile {
		check.Assert(v2.Type, Equals, "application/vnd.vmware.vcloud.vdcStorageProfile+xml")
		check.Assert(v2.HREF, Not(Equals), "")
	}

}

// Tests ComposeVApp with given parameters in the config file.
// Throws an error if networks, catalog, catalog item, and
// storage preference are omitted from the config file.
func (vcd *TestVCD) Test_ComposeVApp(check *C) {
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}
	fmt.Printf("Running: %s\n", check.TestName())

	// Populate OrgVDCNetwork
	networks := []*types.OrgVDCNetwork{}
	net, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	networks = append(networks, net.OrgVDCNetwork)
	check.Assert(err, IsNil)
	// Populate Catalog
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)
	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catitem, NotNil)
	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	check.Assert(err, IsNil)
	// Get StorageProfileReference
	storageprofileref, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)
	// Compose VApp
	task, err := vcd.vdc.ComposeVApp(networks, vapptemplate, storageprofileref, TestComposeVapp, TestComposeVappDesc, true)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Tasks.Task[0].OperationName, Equals, "vdcComposeVapp")
	// Get VApp
	vapp, err := vcd.vdc.GetVAppByName(TestComposeVapp, true)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestComposeVapp, "vapp", "", "Test_ComposeVApp")
	// Once the operation is successful, we won't trigger a failure
	// until after the vApp deletion
	check.Check(vapp.VApp.Name, Equals, TestComposeVapp)
	check.Check(vapp.VApp.Description, Equals, TestComposeVappDesc)

	vapp_status, err := vapp.GetStatus()
	check.Check(err, IsNil)
	check.Check(vapp_status, Equals, "UNRESOLVED")
	// Let the VApp creation complete
	err = task.WaitTaskCompletion()
	if err != nil {
		panic(err)
	}
	err = vapp.BlockWhileStatus("UNRESOLVED", vapp.client.MaxRetryTimeout)
	check.Check(err, IsNil)
	vapp_status, err = vapp.GetStatus()
	check.Check(err, IsNil)
	check.Check(vapp_status, Equals, "POWERED_OFF")
	// Deleting VApp
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	if err != nil {
		panic(err)
	}
	check.Assert(err, IsNil)
	noSuchVapp, err := vcd.vdc.GetVAppByName(TestComposeVapp, true)
	check.Assert(err, NotNil)
	check.Assert(noSuchVapp, IsNil)

}

func (vcd *TestVCD) Test_FindVApp(check *C) {

	if vcd.vapp.VApp == nil {
		check.Skip("No vApp provided")
	}
	firstVapp, err := vcd.vdc.GetVAppByName(vcd.vapp.VApp.Name, false)

	check.Assert(err, IsNil)

	secondVapp, err := vcd.vdc.GetVAppById(firstVapp.VApp.ID, false)

	check.Assert(err, IsNil)

	check.Assert(secondVapp.VApp.Name, Equals, firstVapp.VApp.Name)
	check.Assert(secondVapp.VApp.HREF, Equals, firstVapp.VApp.HREF)
}

func (vcd *TestVCD) Test_FindMediaImage(check *C) {

	if vcd.config.Media.Media == "" {
		check.Skip("Skipping test because no media name given")
	}
	mediaImage, err := vcd.vdc.FindMediaImage(vcd.config.Media.Media)
	check.Assert(err, IsNil)
	if mediaImage == (MediaItem{}) {
		fmt.Printf("Media not found: %s\n", vcd.config.Media.Media)
	}
	check.Assert(mediaImage, Not(Equals), MediaItem{})

	check.Assert(mediaImage.MediaItem.Name, Equals, vcd.config.Media.Media)
	check.Assert(mediaImage.MediaItem.HREF, Not(Equals), "")

	// find Invalid Network
	mediaImage, err = vcd.vdc.FindMediaImage("INVALID")
	check.Assert(err, IsNil)
	check.Assert(mediaImage, Equals, MediaItem{})
}

// Tests function QueryVM by searching vm created
// by test suite
func (vcd *TestVCD) Test_QueryVM(check *C) {

	if vcd.vapp.VApp == nil {
		check.Skip("No Vapp provided")
	}

	// Find VM
	vapp := vcd.findFirstVapp()
	_, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}
	vm, err := vcd.vdc.QueryVM(vcd.vapp.VApp.Name, vmName)
	check.Assert(err, IsNil)

	check.Assert(vm.VM.Name, Equals, vmName)
}

func init() {
	testingTags["vdc"] = "vdc_test.go"
}

// Tests Edge Gateway retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_GetEdgeGateway(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_GetEdgeGateway: Org name not given")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_GetEdgeGateway: VDC name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return vdc.GetEdgeGatewayByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return vdc.GetEdgeGatewayById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return vdc.GetEdgeGatewayByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "Vdc",
		parentName:    vcd.config.VCD.Vdc,
		entityType:    "EdgeGateway",
		entityName:    vcd.config.VCD.EdgeGateway,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

// Tests vApp retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_GetVApp(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_GetVapp: Org name not given")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_GetVapp: VDC name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return vdc.GetVAppByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return vdc.GetVAppById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return vdc.GetVAppByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "Vdc",
		parentName:    vcd.config.VCD.Vdc,
		entityType:    "VApp",
		entityName:    TestSetUpSuite,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}
