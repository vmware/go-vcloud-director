//go:build vdc || functional || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"reflect"
	"strings"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
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
	check.Assert(net, IsNil)
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
		check.Assert(vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity[0].Name, Equals, QtVappTemplate)
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

// Test_GetVDCHardwareVersion tests hardware version fetching functionality
func (vcd *TestVCD) Test_GetVDCHardwareVersion(check *C) {
	err := vcd.vdc.Refresh()
	check.Assert(err, IsNil)

	// vmx-18 is the latest version supported by 10.3.0, the oldest version we support.
	hwVersion, err := vcd.vdc.GetHardwareVersion("vmx-18")
	check.Assert(err, IsNil)
	check.Assert(hwVersion, NotNil)

	check.Assert(hwVersion.Name, Equals, "vmx-18")

	os, err := vcd.vdc.FindOsFromId(hwVersion, "sles10_64Guest")
	check.Assert(err, IsNil)
	check.Assert(os, NotNil)

	check.Assert(os.InternalName, Equals, "sles10_64Guest")
	check.Assert(os.Name, Equals, "SUSE Linux Enterprise 10 (64-bit)")
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
	check.Assert(task.Task.Tasks, NotNil)
	check.Assert(len(task.Task.Tasks.Task) > 0, Equals, true)
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

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

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

// Tests function QueryVM by searching vm created
// by test suite
func (vcd *TestVCD) Test_QueryVM(check *C) {

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

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

	if vcd.client.Client.IsSysAdmin {
		check.Assert(vm.VM.Moref, Not(Equals), "")
		check.Assert(strings.HasPrefix(vm.VM.Moref, "vm-"), Equals, true)
	}
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

// TestGetVappList tests all methods that retrieve a list of vApps
// vdc.GetVappList
// adminVdc.GetVappList
// client.QueryVappList
// client.SearchByFilter
func (vcd *TestVCD) TestGetVappList(check *C) {

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
	vapp := vcd.findFirstVapp()
	if vapp == (VApp{}) {
		check.Skip("no vApp found")
		return
	}

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	adminVdc, err := org.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)

	// Get the vApp list from VDC
	vappList := vdc.GetVappList()
	check.Assert(vappList, NotNil)
	check.Assert(len(vappList), Not(Equals), 0)

	// Get the vApp list from admin VDC
	vappAdminList := adminVdc.GetVappList()
	check.Assert(vappAdminList, NotNil)
	check.Assert(len(vappAdminList), Not(Equals), 0)
	check.Assert(len(vappAdminList), Equals, len(vappList))

	// Check that the known vApp is found in both lists
	foundVappInList := false
	foundVappInAdminList := false
	foundVappInQueryList := false
	for _, ref := range vappList {
		if ref.ID == vapp.VApp.ID {
			foundVappInList = true
		}
	}
	for _, ref := range vappAdminList {
		if ref.ID == vapp.VApp.ID {
			foundVappInAdminList = true
		}
	}
	check.Assert(foundVappInList, Equals, true)
	check.Assert(foundVappInAdminList, Equals, true)

	// Get the vApp list with a query (returns all vApps visible to user, non only the ones withing the current VDC)
	queryVappList, err := vcd.client.Client.QueryVappList()
	check.Assert(err, IsNil)
	check.Assert(queryVappList, NotNil)

	for _, qItem := range queryVappList {
		if qItem.HREF == vapp.VApp.HREF {
			foundVappInQueryList = true
		}
	}
	check.Assert(foundVappInQueryList, Equals, true)

	// Use the search engine to find the known vApp
	criteria := NewFilterDef()
	err = criteria.AddFilter(types.FilterNameRegex, TestSetUpSuite)
	check.Assert(err, IsNil)
	queryType := vcd.client.Client.GetQueryType(types.QtVapp)
	queryItems, _, err := vcd.client.Client.SearchByFilter(queryType, criteria)
	check.Assert(err, IsNil)
	check.Assert(queryItems, NotNil)
	check.Assert(len(queryItems), Not(Equals), 0)
	check.Assert(queryItems[0].GetHref(), Equals, vapp.VApp.HREF)

	// Use the search engine to also find the known VM
	vm, vmName := vcd.findFirstVm(vapp)
	check.Assert(vmName, Not(Equals), "")
	check.Assert(vm.HREF, Not(Equals), "")
	criteria = NewFilterDef()
	err = criteria.AddFilter(types.FilterNameRegex, vmName)
	check.Assert(err, IsNil)
	err = criteria.AddFilter(types.FilterParent, vapp.VApp.Name)
	check.Assert(err, IsNil)
	queryType = vcd.client.Client.GetQueryType(types.QtVm)
	queryItems, _, err = vcd.client.Client.SearchByFilter(queryType, criteria)
	check.Assert(err, IsNil)
	check.Assert(queryItems, NotNil)
	check.Assert(len(queryItems), Not(Equals), 0)
	check.Assert(vm.HREF, Equals, queryItems[0].GetHref())
}

// TestGetVdcCapabilities attempts to get a list of VDC capabilities
func (vcd *TestVCD) TestGetVdcCapabilities(check *C) {
	vdcCapabilities, err := vcd.vdc.GetCapabilities()
	check.Assert(err, IsNil)
	check.Assert(vdcCapabilities, NotNil)
	check.Assert(len(vdcCapabilities) > 0, Equals, true)
}

func (vcd *TestVCD) TestVdcIsNsxt(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	check.Assert(vcd.nsxtVdc.IsNsxt(), Equals, true)
	if vcd.vdc != nil {
		check.Assert(vcd.vdc.IsNsxt(), Equals, false)
	}
}

func (vcd *TestVCD) TestVdcIsNsxv(check *C) {
	check.Assert(vcd.vdc.IsNsxv(), Equals, true)
	// retrieve the same VDC as AdminVdc, to test the corresponding function
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	adminVdc, err := adminOrg.GetVDCByName(vcd.vdc.Vdc.Name, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdc.IsNsxv(), Equals, true)
	// if NSX-T is configured, we also check a NSX-T VDC
	if vcd.nsxtVdc != nil {
		check.Assert(vcd.nsxtVdc.IsNsxv(), Equals, false)
		nsxtAdminVdc, err := adminOrg.GetAdminVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
		check.Assert(err, IsNil)
		check.Assert(nsxtAdminVdc.IsNsxv(), Equals, false)
	}
}

func (vcd *TestVCD) TestCreateRawVapp(check *C) {
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	name := check.TestName()
	description := "test compose raw app"
	vapp, err := vdc.CreateRawVApp(name, description)
	check.Assert(err, IsNil)
	AddToCleanupList(name, "vapp", vdc.Vdc.Name, name)

	check.Assert(vapp.VApp.Name, Equals, name)
	check.Assert(vapp.VApp.Description, Equals, description)
	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) TestSetControlAccess(check *C) {
	// Set VDC sharing to everyone
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	readControlAccessParams, err := vdc.SetControlAccess(true, "ReadOnly", nil, true)
	check.Assert(err, IsNil)
	check.Assert(readControlAccessParams, NotNil)
	check.Assert(readControlAccessParams.IsSharedToEveryone, Equals, true)
	check.Assert(*readControlAccessParams.EveryoneAccessLevel, Equals, "ReadOnly")
	check.Assert(readControlAccessParams.AccessSettings, IsNil) // If not shared with users/groups, this will be nil

	// Set VDC sharing to one user
	orgUserRef := org.AdminOrg.Users.User[0]
	user, err := org.GetUserByName(orgUserRef.Name, false)
	check.Assert(err, IsNil)
	check.Assert(user, NotNil)

	accessSettings := []*types.AccessSetting{
		{
			AccessLevel: "ReadOnly",
			Subject: &types.LocalSubject{
				HREF: user.User.Href,
				Name: user.User.Name,
				Type: user.User.Type,
			},
		},
	}

	readControlAccessParams, err = vdc.SetControlAccess(false, "", accessSettings, true)
	check.Assert(err, IsNil)
	check.Assert(readControlAccessParams, NotNil)
	check.Assert(len(readControlAccessParams.AccessSettings.AccessSetting) > 0, Equals, true)
	check.Assert(assertVDCAccessSettings(accessSettings, readControlAccessParams.AccessSettings.AccessSetting), IsNil)

	// Check that fail if both isSharedToEveryone and accessSettings is passed
	readControlAccessParams, err = vdc.SetControlAccess(true, "ReadOnly", accessSettings, true)
	check.Assert(err, NotNil)
	check.Assert(readControlAccessParams, IsNil)

	// Check DeleteControlAccess
	readControlAccessParams, err = vdc.DeleteControlAccess(true)
	check.Assert(err, IsNil)
	check.Assert(readControlAccessParams.IsSharedToEveryone, Equals, false)
	check.Assert(readControlAccessParams.AccessSettings, IsNil)
}

func assertVDCAccessSettings(wanted, received []*types.AccessSetting) error {
	if len(wanted) != len(received) {
		return fmt.Errorf("wanted and received access settings are not the same length")
	}
	for _, receivedAccessSetting := range received {
		for i, wantedAccessSetting := range wanted {
			if reflect.DeepEqual(*wantedAccessSetting.Subject, *receivedAccessSetting.Subject) && (wantedAccessSetting.AccessLevel == receivedAccessSetting.AccessLevel) {
				break
			}
			if i == len(wanted)-1 {
				return fmt.Errorf("access settings for user %s were not found or are not correct", wantedAccessSetting.Subject.Name)
			}
		}
	}
	return nil
}

// TestVAppTemplateRetrieval tests that VDC receiver objects can search vApp Templates successfully.
func (vcd *TestVCD) TestVAppTemplateRetrieval(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Catalog.NsxtCatalogItem == "" {
		check.Skip(fmt.Sprintf("%s: Catalog Item not given. Test can't proceed", check.TestName()))
	}

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Test cases
	vAppTemplate, err := vdc.GetVAppTemplateByName(vcd.config.VCD.Catalog.NsxtCatalogItem)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.NsxtCatalogItem)
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(strings.Contains(vAppTemplate.VAppTemplate.Description, vcd.config.VCD.Catalog.CatalogItemDescription), Equals, true)
	}

	vAppTemplate, err = vcd.client.GetVAppTemplateById(vAppTemplate.VAppTemplate.ID)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.NsxtCatalogItem)
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(strings.Contains(vAppTemplate.VAppTemplate.Description, vcd.config.VCD.Catalog.CatalogItemDescription), Equals, true)
	}

	vAppTemplate, err = vdc.GetVAppTemplateByNameOrId(vAppTemplate.VAppTemplate.ID, false)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.NsxtCatalogItem)
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(strings.Contains(vAppTemplate.VAppTemplate.Description, vcd.config.VCD.Catalog.CatalogItemDescription), Equals, true)
	}

	vAppTemplate, err = vdc.GetVAppTemplateByNameOrId(vcd.config.VCD.Catalog.NsxtCatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.NsxtCatalogItem)
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(strings.Contains(vAppTemplate.VAppTemplate.Description, vcd.config.VCD.Catalog.CatalogItemDescription), Equals, true)
	}

	vAppTemplateRecord, err := vcd.client.QuerySynchronizedVAppTemplateById(vAppTemplate.VAppTemplate.ID)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplateRecord.Name, Equals, vAppTemplate.VAppTemplate.Name)
	check.Assert(vAppTemplateRecord.HREF, Equals, vAppTemplate.VAppTemplate.HREF)

	vmTemplateRecord, err := vcd.client.QuerySynchronizedVmInVAppTemplateByHref(vAppTemplate.VAppTemplate.HREF, "**")
	check.Assert(err, IsNil)
	check.Assert(vmTemplateRecord, NotNil)

	// Test non-existent vApp Template
	_, err = vdc.GetVAppTemplateByName("INVALID")
	check.Assert(err, NotNil)

	_, err = vcd.client.QuerySynchronizedVmInVAppTemplateByHref(vAppTemplate.VAppTemplate.HREF, "INVALID")
	check.Assert(err, Equals, ErrorEntityNotFound)
}

// TestMediaRetrieval tests that VDC receiver objects can search Media items successfully.
func (vcd *TestVCD) TestMediaRetrieval(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.Media.NsxtMedia == "" {
		check.Skip(fmt.Sprintf("%s: NSX-T Media item not given. Test can't proceed", check.TestName()))
	}

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.NsxtBackedCatalogName, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	mediaFromCatalog, err := catalog.GetMediaByName(vcd.config.Media.NsxtMedia, false)
	check.Assert(err, IsNil)
	check.Assert(mediaFromCatalog, NotNil)

	// Test cases
	mediaFromVdc, err := vcd.client.QueryMediaById(mediaFromCatalog.Media.ID)
	check.Assert(err, IsNil)
	check.Assert(mediaFromCatalog.Media.HREF, Equals, mediaFromVdc.MediaRecord.HREF)
	check.Assert(mediaFromCatalog.Media.Name, Equals, mediaFromVdc.MediaRecord.Name)

	// Test non-existent Media item
	mediaFromVdc, err = vcd.client.QueryMediaById("INVALID")
	check.Assert(err, NotNil)
	check.Assert(mediaFromVdc, IsNil)
}
