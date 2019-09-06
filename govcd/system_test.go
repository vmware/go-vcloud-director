// +build system functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Tests Org retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_SystemGetOrg(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_SystemGetOrg: Org name not given.")
		return
	}

	getByName := func(name string, refresh bool) (genericEntity, error) { return vcd.client.GetOrgByName(name) }
	getById := func(id string, refresh bool) (genericEntity, error) { return vcd.client.GetOrgById(id) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) { return vcd.client.GetOrgByNameOrId(id) }

	var def = getterTestDefinition{
		parentType:    "VCDClient",
		parentName:    "System",
		entityType:    "Org",
		entityName:    vcd.config.VCD.Org,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

// Tests AdminOrg retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_SystemGetAdminOrg(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_SystemGetAdminOrg: Org name not given.")
		return
	}

	getByName := func(name string, refresh bool) (genericEntity, error) { return vcd.client.GetAdminOrgByName(name) }
	getById := func(id string, refresh bool) (genericEntity, error) { return vcd.client.GetAdminOrgById(id) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) { return vcd.client.GetAdminOrgByNameOrId(id) }

	var def = getterTestDefinition{
		parentType:    "VCDClient",
		parentName:    "System",
		entityType:    "AdminOrg",
		entityName:    vcd.config.VCD.Org,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

// Tests the creation of an org with general settings,
// org vapp template settings, and orgldapsettings. Asserts an
// error if the task, fetching the org, or deleting the org fails
func (vcd *TestVCD) Test_CreateOrg(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	type testOrgData struct {
		name                     string
		enabled                  bool
		canPublishCatalogs       bool
		deployedVmQuota          int
		storedVmQuota            int
		delayAfterPowerOnSeconds int
		fullData                 bool
	}
	var orgList = []testOrgData{
		{"org1", true, false, 0, 0, 0, true},
		{"org2", true, true, 0, 0, 1, false},
		{"org3", false, false, 1, 1, 3, true},
		{"org4", true, true, 10, 10, 10, false},
		{"org5", false, true, 100, 100, 100, false},
	}

	fullSettings := &types.OrgSettings{
		OrgGeneralSettings: &types.OrgGeneralSettings{},
		OrgVAppTemplateSettings: &types.VAppTemplateLeaseSettings{
			DeleteOnStorageLeaseExpiration: true,
			StorageLeaseSeconds:            10,
		},
		OrgVAppLeaseSettings: &types.VAppLeaseSettings{
			PowerOffOnRuntimeLeaseExpiration: true,
			DeploymentLeaseSeconds:           1000000,
			DeleteOnStorageLeaseExpiration:   true,
			StorageLeaseSeconds:              1000000,
		},
		OrgLdapSettings: &types.OrgLdapSettingsType{
			OrgLdapMode: "NONE",
		},
	}
	for _, od := range orgList {
		var settings *types.OrgSettings
		if od.fullData {
			settings = fullSettings
		} else {
			settings = &types.OrgSettings{
				OrgGeneralSettings: &types.OrgGeneralSettings{},
			}
		}
		orgName := TestCreateOrg + "_" + od.name

		fmt.Printf("# org %s (enabled: %v - catalogs: %v [%d %d])\n", orgName, od.enabled, od.canPublishCatalogs, od.storedVmQuota, od.deployedVmQuota)
		settings.OrgGeneralSettings.CanPublishCatalogs = od.canPublishCatalogs
		settings.OrgGeneralSettings.DeployedVMQuota = od.deployedVmQuota
		settings.OrgGeneralSettings.StoredVMQuota = od.storedVmQuota
		settings.OrgGeneralSettings.DelayAfterPowerOnSeconds = od.delayAfterPowerOnSeconds
		task, err := CreateOrg(vcd.client, orgName, TestCreateOrg, TestCreateOrg, settings, od.enabled)
		check.Assert(err, IsNil)
		// After a successful creation, the entity is added to the cleanup list.
		// If something fails after this point, the entity will be removed
		AddToCleanupList(orgName, "org", "", "TestCreateOrg")
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		// fetch newly created org
		adminOrg, err := vcd.client.GetAdminOrgByName(orgName)
		check.Assert(err, IsNil)
		check.Assert(adminOrg, NotNil)
		check.Assert(adminOrg.AdminOrg.Name, Equals, orgName)
		check.Assert(adminOrg.AdminOrg.Description, Equals, TestCreateOrg)
		check.Assert(adminOrg.AdminOrg.IsEnabled, Equals, od.enabled)

		check.Assert(adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs, Equals, od.canPublishCatalogs)
		check.Assert(adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota, Equals, od.deployedVmQuota)
		check.Assert(adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.StoredVMQuota, Equals, od.storedVmQuota)
		check.Assert(adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.DelayAfterPowerOnSeconds, Equals, od.delayAfterPowerOnSeconds)
		// Delete, with force and recursive true
		err = adminOrg.Delete(true, true)
		check.Assert(err, IsNil)
		doesOrgExist(check, vcd)
	}
}

func (vcd *TestVCD) Test_CreateDeleteEdgeGateway(check *C) {

	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("No external network provided")
	}

	newEgwName := "CreateDeleteEdgeGateway"
	orgName := vcd.config.VCD.Org
	vdcName := vcd.config.VCD.Vdc
	egc := EdgeGatewayCreation{
		ExternalNetworks:          []string{vcd.config.VCD.ExternalNetwork},
		DefaultGateway:            vcd.config.VCD.ExternalNetwork,
		OrgName:                   orgName,
		VdcName:                   vdcName,
		AdvancedNetworkingEnabled: true,
	}

	testingRange := []string{"compact", "full"}
	for _, backingConf := range testingRange {
		egc.BackingConfiguration = backingConf
		egc.Name = newEgwName + "_" + backingConf
		egc.Description = egc.Name

		var edge EdgeGateway
		var task Task
		var err error
		builtWithDefaultGateway := true
		// Tests one edge gateway with default gateway, and one without
		// Also tests two different functions to create the gateway
		if backingConf == "full" {
			egc.DefaultGateway = vcd.config.VCD.ExternalNetwork
			edge, err = CreateEdgeGateway(vcd.client, egc)
			check.Assert(err, IsNil)
		} else {
			// The "compact" edge gateway is created without default gateway
			egc.DefaultGateway = ""
			builtWithDefaultGateway = false
			task, err = CreateEdgeGatewayAsync(vcd.client, egc)
			check.Assert(err, IsNil)
			err = task.WaitTaskCompletion()
			check.Assert(err, IsNil)
			newEdge, err := vcd.vdc.GetEdgeGatewayByName(egc.Name, true)
			check.Assert(err, IsNil)
			check.Assert(newEdge, NotNil)
			edge = *newEdge
		}

		AddToCleanupList(egc.Name, "edgegateway", orgName+"|"+vdcName, "Test_CreateDeleteEdgeGateway")

		check.Assert(edge.EdgeGateway.Name, Equals, egc.Name)
		// Edge gateway status:
		//  0 : being created
		//  1 : ready
		// -1 : creation error
		check.Assert(edge.EdgeGateway.Status, Equals, 1)

		check.Assert(edge.EdgeGateway.Configuration.AdvancedNetworkingEnabled, Equals, true)
		util.Logger.Printf("Edge Gateway:\n%s\n", prettyEdgeGateway(*edge.EdgeGateway))

		check.Assert(edge.HasDefaultGateway(), Equals, builtWithDefaultGateway)
		check.Assert(edge.HasAdvancedNetworking(), Equals, egc.AdvancedNetworkingEnabled)

		// testing both delete methods
		if backingConf == "full" {
			err = edge.Delete(true, true)
			check.Assert(err, IsNil)
		} else {
			task, err := edge.DeleteAsync(true, true)
			check.Assert(err, IsNil)
			err = task.WaitTaskCompletion()
			check.Assert(err, IsNil)
		}

		// Once deleted, look for the edge gateway again. It should return an error
		newEdge, err := vcd.vdc.GetEdgeGatewayByName(egc.Name, true)
		check.Assert(err, NotNil)
		check.Assert(newEdge, IsNil)
	}
}

func (vcd *TestVCD) Test_FindBadlyNamedStorageProfile(check *C) {
	reNotFound := `can't find any VDC Storage_profiles`
	_, err := vcd.vdc.FindStorageProfileReference("name with spaces")
	check.Assert(err, NotNil)
	check.Assert(err.Error(), Matches, reNotFound)
}

// Test getting network pool by href and vdc client
func (vcd *TestVCD) Test_GetNetworkPoolByHREF(check *C) {
	if vcd.config.VCD.ProviderVdc.NetworkPool == "" {
		check.Skip("Skipping test because network pool is not configured")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)

	// Get network pool by href
	foundNetworkPool, err := GetNetworkPoolByHREF(vcd.client, adminVdc.AdminVdc.NetworkPoolReference.HREF)
	check.Assert(err, IsNil)
	check.Assert(foundNetworkPool, Not(Equals), types.VMWNetworkPool{})
}

func init() {
	testingTags["system"] = "system_test.go"
}

func (vcd *TestVCD) Test_QueryOrgVdcNetworkByName(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	orgVdcNetwork, err := QueryOrgVdcNetworkByName(vcd.client, vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)
	check.Assert(len(orgVdcNetwork), Not(Equals), 0)
	check.Assert(orgVdcNetwork[0].Name, Equals, vcd.config.VCD.Network.Net1)
	check.Assert(orgVdcNetwork[0].ConnectedTo, Equals, vcd.config.VCD.EdgeGateway)
}

func (vcd *TestVCD) Test_QueryOrgVdcNetworkByNameWithSpace(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	networkName := "Test_QueryOrgVdcNetworkByNameWith Space"

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("[Test_CreateOrgVdcNetworkDirect] external network not provided")
	}
	externalNetwork, err := vcd.client.GetExternalNetworkByName(vcd.config.VCD.ExternalNetwork)
	if err != nil {
		check.Skip("[Test_CreateOrgVdcNetworkDirect] parent external network not found")
	}
	// Note that there is no IPScope for this type of network
	var networkConfig = types.OrgVDCNetwork{
		Xmlns: types.XMLNamespaceVCloud,
		Name:  networkName,
		Configuration: &types.NetworkConfiguration{
			FenceMode: types.FenceModeBridged,
			ParentNetwork: &types.Reference{
				HREF: externalNetwork.ExternalNetwork.HREF,
				Name: externalNetwork.ExternalNetwork.Name,
				Type: externalNetwork.ExternalNetwork.Type,
			},
			BackwardCompatibilityMode: true,
		},
		IsShared: false,
	}
	LogNetwork(networkConfig)

	task, err := vcd.vdc.CreateOrgVDCNetwork(&networkConfig)
	if err != nil {
		fmt.Printf("error creating the network: %s", err)
	}
	check.Assert(err, IsNil)
	if task == (Task{}) {
		fmt.Printf("NULL task retrieved after network creation")
	}
	check.Assert(task.Task.HREF, Not(Equals), "")

	AddToCleanupList(networkName,
		"network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name,
		"Test_CreateOrgVdcNetworkDirect")

	// err = task.WaitTaskCompletion()
	err = task.WaitInspectTaskCompletion(LogTask, 10)
	if err != nil {
		fmt.Printf("error performing task: %#v", err)
	}
	check.Assert(err, IsNil)

	orgVdcNetwork, err := QueryOrgVdcNetworkByName(vcd.client, networkName)
	check.Assert(err, IsNil)
	check.Assert(len(orgVdcNetwork), Not(Equals), 0)
	check.Assert(orgVdcNetwork[0].Name, Equals, networkName)
	check.Assert(orgVdcNetwork[0].ConnectedTo, Equals, externalNetwork.ExternalNetwork.Name)
}

func (vcd *TestVCD) Test_QueryProviderVdcEntities(check *C) {
	providerVdcName := vcd.config.VCD.ProviderVdc.Name
	networkPoolName := vcd.config.VCD.ProviderVdc.NetworkPool
	storageProfileName := vcd.config.VCD.ProviderVdc.StorageProfile
	if providerVdcName == "" {
		check.Skip("Skipping Provider VDC query: no provider VDC was given")
	}
	providerVdcs, err := vcd.client.QueryProviderVdcs()
	check.Assert(err, IsNil)
	check.Assert(len(providerVdcName) > 0, Equals, true)

	providerFound := false
	for _, providerVdc := range providerVdcs {
		if providerVdcName == providerVdc.Name {
			providerFound = true
		}

		if testVerbose {
			fmt.Printf("PVDC %s\n", providerVdc.Name)
			fmt.Printf("\t href    %s\n", providerVdc.HREF)
			fmt.Printf("\t status  %s\n", providerVdc.Status)
			fmt.Printf("\t enabled %v\n", providerVdc.IsEnabled)
			fmt.Println("")
		}
	}
	check.Assert(providerFound, Equals, true)

	if networkPoolName == "" {
		check.Skip("Skipping Network pool query: no network pool was given")
	}
	netPools, err := vcd.client.QueryNetworkPools()
	check.Assert(err, IsNil)
	check.Assert(len(netPools) > 0, Equals, true)
	networkPoolFound := false
	for _, networkPool := range netPools {
		if networkPoolName == networkPool.Name {
			networkPoolFound = true
		}
		if testVerbose {
			fmt.Printf("NP %s\n", networkPool.Name)
			fmt.Printf("\t href %s\n", networkPool.HREF)
			fmt.Printf("\t type %v\n", networkPool.NetworkPoolType)
			fmt.Println("")
		}
	}
	check.Assert(networkPoolFound, Equals, true)

	if storageProfileName == "" {
		check.Skip("Skipping storage profile query: no storage profile was given")
	}
	storageProfiles, err := vcd.client.QueryProviderVdcStorageProfiles()
	check.Assert(err, IsNil)
	check.Assert(len(storageProfiles) > 0, Equals, true)
	storageProfileFound := false
	for _, sp := range storageProfiles {
		if storageProfileName == sp.Name {
			storageProfileFound = true
		}
		if testVerbose {
			fmt.Printf("SP %s\n", sp.Name)
			fmt.Printf("\t enabled     %12v\n", sp.IsEnabled)
			fmt.Printf("\t storage     %12d\n", sp.StorageTotalMB)
			fmt.Printf("\t provisioned %12d\n", sp.StorageProvisionedMB)
			fmt.Printf("\t requested   %12d\n", sp.StorageRequestedMB)
			fmt.Printf("\t used        %12d\n", sp.StorageUsedMB)
			fmt.Println("")
		}
	}
	check.Assert(storageProfileFound, Equals, true)

}

func (vcd *TestVCD) Test_QueryProviderVdcByName(check *C) {
	if vcd.config.VCD.ProviderVdc.Name == "" {
		check.Skip("Skipping Provider VDC query: no provider VDC was given")
	}
	providerVdcs, err := QueryProviderVdcByName(vcd.client, vcd.config.VCD.ProviderVdc.Name)
	check.Assert(err, IsNil)
	check.Assert(len(providerVdcs) > 0, Equals, true)

	providerFound := false
	for _, providerVdc := range providerVdcs {
		if vcd.config.VCD.ProviderVdc.Name == providerVdc.Name {
			providerFound = true
		}

		if testVerbose {
			fmt.Printf("PVDC %s\n", providerVdc.Name)
			fmt.Printf("\t href    %s\n", providerVdc.HREF)
			fmt.Printf("\t status  %s\n", providerVdc.Status)
			fmt.Printf("\t enabled %v\n", providerVdc.IsEnabled)
			fmt.Println("")
		}
	}
	check.Assert(providerFound, Equals, true)

}

func (vcd *TestVCD) Test_QueryNetworkPoolByName(check *C) {
	if vcd.config.VCD.ProviderVdc.NetworkPool == "" {
		check.Skip("Skipping Provider VDC network pool query: no provider VDC network pool was given")
	}
	netPools, err := QueryNetworkPoolByName(vcd.client, vcd.config.VCD.ProviderVdc.NetworkPool)
	check.Assert(err, IsNil)
	check.Assert(len(netPools) > 0, Equals, true)

	networkPoolFound := false
	for _, networkPool := range netPools {
		if vcd.config.VCD.ProviderVdc.NetworkPool == networkPool.Name {
			networkPoolFound = true
		}
		if testVerbose {
			fmt.Printf("NP %s\n", networkPool.Name)
			fmt.Printf("\t href %s\n", networkPool.HREF)
			fmt.Printf("\t type %v\n", networkPool.NetworkPoolType)
			fmt.Println("")
		}
	}
	check.Assert(networkPoolFound, Equals, true)

}
