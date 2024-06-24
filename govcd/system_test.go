//go:build system || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Tests Org retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_SystemGetOrg(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_SystemGetOrg: Org name not given")
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

	storageLeaseSeconds := 10
	deploymentLeaseSeconds := 1000000
	trueVal := true
	fullSettings := &types.OrgSettings{
		OrgGeneralSettings: &types.OrgGeneralSettings{},
		OrgVAppTemplateSettings: &types.VAppTemplateLeaseSettings{
			DeleteOnStorageLeaseExpiration: &trueVal,
			StorageLeaseSeconds:            &storageLeaseSeconds,
		},
		OrgVAppLeaseSettings: &types.VAppLeaseSettings{
			PowerOffOnRuntimeLeaseExpiration: &trueVal,
			DeploymentLeaseSeconds:           &deploymentLeaseSeconds,
			DeleteOnStorageLeaseExpiration:   &trueVal,
			StorageLeaseSeconds:              &storageLeaseSeconds,
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

		if vcd.client.Client.APIVCDMaxVersionIs("= 37.2") && !od.enabled {
			// TODO revisit once bug is fixed in VCD
			fmt.Println("[INFO] VCD 10.4.2 has a bug that prevents creating a disabled Org - Changing 'enabled' parameter to 'true'")
			od.enabled = true
		}
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
	vcd.skipIfNotSysAdmin(check)
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

		check.Assert(edge.EdgeGateway.Configuration.AdvancedNetworkingEnabled, NotNil)
		check.Assert(*edge.EdgeGateway.Configuration.AdvancedNetworkingEnabled, Equals, true)
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
		check.Assert(err, Equals, ErrorEntityNotFound)
		check.Assert(newEdge, IsNil)
	}
}

// Test_CreateDeleteEdgeGatewayAdvanced sets up external network which has multiple IP scopes and IP
// ranges defined. This helps to test edge gateway capabilities for multiple networks and scopes
func (vcd *TestVCD) Test_CreateDeleteEdgeGatewayAdvanced(check *C) {
	// Setup external network with multiple IP scopes and multiple ranges
	dnsSuffix := "some.net"
	skippingReason, externalNetwork, task, err := vcd.testCreateExternalNetwork(check.TestName(), check.TestName(), dnsSuffix)
	if skippingReason != "" {
		check.Skip(skippingReason)
	}

	check.Assert(err, IsNil)
	check.Assert(task.Task, Not(Equals), types.Task{})

	AddToCleanupList(externalNetwork.Name, "externalNetwork", "", check.TestName())
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// "Refresh" external network to fill in all fields (like HREF)
	extNet, err := vcd.client.GetExternalNetworkByName(externalNetwork.Name)
	check.Assert(err, IsNil)
	externalNetwork = extNet.ExternalNetwork

	edgeName := "Test-Multi-Scope-Gw"
	// Initialize edge gateway structure
	edgeGatewayConfig := &types.EdgeGateway{
		Name:        edgeName,
		Description: edgeName,
		Configuration: &types.GatewayConfiguration{
			HaEnabled:            addrOf(false),
			GatewayBackingConfig: "compact",
			GatewayInterfaces: &types.GatewayInterfaces{
				GatewayInterface: []*types.GatewayInterface{},
			},
			AdvancedNetworkingEnabled:  addrOf(true),
			DistributedRoutingEnabled:  addrOf(false),
			UseDefaultRouteForDNSRelay: addrOf(true),
		},
	}

	edgeGatewayConfig.Configuration.FipsModeEnabled = addrOf(false)

	// Create subnet participation structure
	subnetParticipation := make([]*types.SubnetParticipation, len(externalNetwork.Configuration.IPScopes.IPScope))
	// Loop over IP scopes
	for ipScopeIndex, ipScope := range externalNetwork.Configuration.IPScopes.IPScope {
		subnetParticipation[ipScopeIndex] = &types.SubnetParticipation{
			Gateway: ipScope.Gateway,
			Netmask: ipScope.Netmask,
			// IPAddress: string,			// Can be set to specify IP address of edge gateway
			// UseForDefaultRoute: bool,	// Can be specified to use subnet as default gateway
			IPRanges: &types.IPRanges{},
		}
	}

	// Setup network interface config
	networkConf := &types.GatewayInterface{
		Name:          externalNetwork.Name,
		DisplayName:   externalNetwork.Name,
		InterfaceType: "uplink",
		Network: &types.Reference{
			HREF: externalNetwork.HREF,
			ID:   externalNetwork.ID,
			Type: "application/vnd.vmware.admin.network+xml",
			Name: externalNetwork.Name,
		},
		UseForDefaultRoute:  true,
		SubnetParticipation: subnetParticipation,
	}

	// Sort by subnet participation gateway so that below injected variables are not being added to
	// incorrect network
	networkConf.SortBySubnetParticipationGateway()
	// Set static IP assignment
	networkConf.SubnetParticipation[0].IPAddress = "192.168.201.100"
	// Set default gateway subnet
	networkConf.SubnetParticipation[1].UseForDefaultRoute = true
	// Inject an IP range (in UI it is called "sub-allocated pools" in separate tab)
	networkConf.SubnetParticipation[0].IPRanges = &types.IPRanges{
		IPRange: []*types.IPRange{
			&types.IPRange{
				StartAddress: "192.168.201.120",
				EndAddress:   "192.168.201.130",
			},
		},
	}

	edgeGatewayConfig.Configuration.GatewayInterfaces.GatewayInterface =
		append(edgeGatewayConfig.Configuration.GatewayInterfaces.GatewayInterface, networkConf)

	orgName := vcd.config.VCD.Org
	vdcName := vcd.config.VCD.Vdc

	edge, err := CreateAndConfigureEdgeGateway(vcd.client, orgName, vdcName, edgeName, edgeGatewayConfig)
	check.Assert(err, IsNil)
	PrependToCleanupList(edge.EdgeGateway.Name, "edgegateway", orgName+"|"+vdcName, "Test_CreateDeleteEdgeGateway")

	// Patch known differences for comparison deep comparison
	edgeGatewayConfig.Configuration.GatewayInterfaces.GatewayInterface[0].SubnetParticipation[1].IPAddress = "192.168.231.3"
	edgeGatewayConfig.Configuration.GatewayInterfaces.GatewayInterface[0].Network.HREF =
		edge.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface[0].Network.HREF

	edgeGatewayConfig.Configuration.GatewayInterfaces.GatewayInterface[0].Network.HREF =
		edge.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface[0].Network.HREF

	//edgeGatewayConfig.Configuration.GatewayInterfaces.GatewayInterface[0].Network.ID = ""

	// Sort gateway interfaces so that comparison is easier
	edgeGatewayConfig.Configuration.GatewayInterfaces.GatewayInterface[0].SortBySubnetParticipationGateway()
	edge.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface[0].SortBySubnetParticipationGateway()

	check.Assert(edge.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface[0], DeepEquals,
		edgeGatewayConfig.Configuration.GatewayInterfaces.GatewayInterface[0])
	check.Assert(edge.EdgeGateway.Configuration.DistributedRoutingEnabled, NotNil)
	check.Assert(*edge.EdgeGateway.Configuration.DistributedRoutingEnabled, Equals, false)

	// FIPS mode is not being returned from API (neither when it is enabled, nor when disabled), but
	// does allow to turn it on.
	// check.Assert(edge.EdgeGateway.Configuration.FipsModeEnabled, NotNil)
	// check.Assert(*edge.EdgeGateway.Configuration.FipsModeEnabled, Equals, true)

	check.Assert(edge.EdgeGateway.Configuration.AdvancedNetworkingEnabled, NotNil)
	check.Assert(*edge.EdgeGateway.Configuration.AdvancedNetworkingEnabled, Equals, true)
	check.Assert(edge.EdgeGateway.Configuration.UseDefaultRouteForDNSRelay, NotNil)
	check.Assert(*edge.EdgeGateway.Configuration.UseDefaultRouteForDNSRelay, Equals, true)
	check.Assert(edge.EdgeGateway.Configuration.HaEnabled, NotNil)
	check.Assert(*edge.EdgeGateway.Configuration.HaEnabled, Equals, false)

	// Remove created objects to free them up
	err = edge.Delete(true, false)
	check.Assert(err, IsNil)

	err = extNet.DeleteWait()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_FindBadlyNamedStorageProfile(check *C) {
	reNotFound := `can't find any VDC Storage_profiles`
	_, err := vcd.vdc.FindStorageProfileReference("name with spaces")
	check.Assert(err, NotNil)
	check.Assert(err.Error(), Matches, reNotFound)
}

// Test getting network pool by href and vdc client
func (vcd *TestVCD) Test_GetNetworkPoolByHREF(check *C) {
	vcd.skipIfNotSysAdmin(check)
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
		Name: networkName,
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

	AddToCleanupList(networkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, check.TestName())

	// err = task.WaitTaskCompletion()
	err = task.WaitInspectTaskCompletion(LogTask, 10)
	if err != nil {
		fmt.Printf("error performing task: %s", err)
	}
	check.Assert(err, IsNil)

	orgVdcNetwork, err := QueryOrgVdcNetworkByName(vcd.client, networkName)
	check.Assert(err, IsNil)
	check.Assert(len(orgVdcNetwork), Not(Equals), 0)
	check.Assert(orgVdcNetwork[0].Name, Equals, networkName)
	check.Assert(orgVdcNetwork[0].ConnectedTo, Equals, externalNetwork.ExternalNetwork.Name)
	network, err := vcd.vdc.GetOrgVdcNetworkByName(networkName, true)
	check.Assert(err, IsNil)
	task, err = network.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_QueryProviderVdcEntities(check *C) {
	vcd.skipIfNotSysAdmin(check)
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
	storageProfiles, err := vcd.client.Client.QueryAllProviderVdcStorageProfiles()
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
	vcd.skipIfNotSysAdmin(check)
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

func (vcd *TestVCD) Test_QueryAdminOrgVdcStorageProfileByID(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("Skipping Admin VDC StorageProfile query: can't query as tenant user")
	}
	if vcd.config.VCD.StorageProfile.SP1 == "" {
		check.Skip("Skipping VDC StorageProfile query: no StorageProfile ID was given")
	}
	ref, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)
	expectedStorageProfileID, err := GetUuidFromHref(ref.HREF, true)
	check.Assert(err, IsNil)
	vdcStorageProfile, err := QueryAdminOrgVdcStorageProfileByID(vcd.client, ref.ID)
	check.Assert(err, IsNil)

	storageProfileFound := false

	storageProfileID, err := GetUuidFromHref(vdcStorageProfile.HREF, true)
	check.Assert(err, IsNil)
	if storageProfileID == expectedStorageProfileID {
		storageProfileFound = true
	}

	if testVerbose {
		fmt.Printf("StorageProfile %s\n", vdcStorageProfile.Name)
		fmt.Printf("\t href    %s\n", vdcStorageProfile.HREF)
		fmt.Printf("\t enabled %v\n", vdcStorageProfile.IsEnabled)
		fmt.Println("")
	}

	check.Assert(storageProfileFound, Equals, true)
}

func (vcd *TestVCD) Test_QueryOrgVdcStorageProfileByID(check *C) {
	if vcd.config.VCD.StorageProfile.SP1 == "" {
		check.Skip("Skipping VDC StorageProfile query: no StorageProfile ID was given")
	}

	// Setup Org user and connection
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	orgUserVcdClient, _, err := newOrgUserConnection(adminOrg, "query-org-vdc-storage-profile-by-id", "CHANGE-ME", vcd.config.Provider.Url, true)
	check.Assert(err, IsNil)

	ref, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)
	expectedStorageProfileID, err := GetUuidFromHref(ref.HREF, true)
	check.Assert(err, IsNil)
	vdcStorageProfile, err := QueryOrgVdcStorageProfileByID(orgUserVcdClient, ref.ID)
	check.Assert(err, IsNil)

	storageProfileFound := false

	storageProfileID, err := GetUuidFromHref(vdcStorageProfile.HREF, true)
	check.Assert(err, IsNil)
	if storageProfileID == expectedStorageProfileID {
		storageProfileFound = true
	}

	if testVerbose {
		fmt.Printf("StorageProfile %s\n", vdcStorageProfile.Name)
		fmt.Printf("\t href    %s\n", vdcStorageProfile.HREF)
		fmt.Printf("\t enabled %v\n", vdcStorageProfile.IsEnabled)
		fmt.Println("")
	}

	check.Assert(storageProfileFound, Equals, true)
}

func (vcd *TestVCD) Test_QueryNetworkPoolByName(check *C) {
	vcd.skipIfNotSysAdmin(check)
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

// Test_GetStorageProfile tests all the getters of Storage Profile
func (vcd *TestVCD) Test_GetStorageProfile(check *C) {
	if vcd.config.VCD.ProviderVdc.StorageProfile == "" {
		check.Skip("Skipping test because storage profile is not configured")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)

	// Get storage profile by href
	foundStorageProfile, err := vcd.client.Client.GetStorageProfileByHref(adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile[0].HREF)
	check.Assert(err, IsNil)
	check.Assert(foundStorageProfile, NotNil)
	check.Assert(foundStorageProfile.IopsSettings, NotNil)
	check.Assert(foundStorageProfile, Not(Equals), types.VdcStorageProfile{})
	check.Assert(foundStorageProfile.IopsSettings, Not(Equals), types.VdcStorageProfileIopsSettings{})

	// Get storage profile by ID
	foundStorageProfile2, err := vcd.client.GetStorageProfileById(foundStorageProfile.ID)
	check.Assert(err, IsNil)
	check.Assert(foundStorageProfile, DeepEquals, foundStorageProfile2)
}

func (vcd *TestVCD) Test_GetOrgList(check *C) {

	orgs, err := vcd.client.GetOrgList()
	check.Assert(err, IsNil)
	check.Assert(orgs, NotNil)

	if vcd.config.VCD.Org != "" {
		foundOrg := false
		for _, org := range orgs.Org {
			if org.Name == vcd.config.VCD.Org {
				foundOrg = true
			}
		}
		check.Assert(foundOrg, Equals, true)
	}
}

func (vcd *TestVCD) TestQueryAllVdcs(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	allVdcs, err := vcd.client.Client.QueryAllVdcs()
	check.Assert(err, IsNil)

	// Check for at least that many VDCs in VCD
	// expectedVdcCountInSystem = 1 NSX-V VDC
	expectedVdcCountInSystem := 1
	// If an NSX-T VDC exists - then expected count of VDCs is at least 2
	if vcd.config.VCD.Nsxt.Vdc != "" {
		expectedVdcCountInSystem++
	}

	if testVerbose {
		fmt.Printf("# List contains at least %d VDCs.", expectedVdcCountInSystem)
	}
	check.Assert(len(allVdcs) >= expectedVdcCountInSystem, Equals, true)
	// Check that known VDCs are inside the list

	knownVdcs := []string{vcd.config.VCD.Vdc}
	if vcd.config.VCD.Nsxt.Vdc != "" {
		knownVdcs = append(knownVdcs, vcd.config.VCD.Nsxt.Vdc)
	}

	foundVdcNames := make([]string, len(allVdcs))
	for vdcIndex, vdc := range allVdcs {
		foundVdcNames[vdcIndex] = vdc.Name
	}

	if testVerbose {
		fmt.Printf("# Checking result contains all known VDCs (%s).", strings.Join((knownVdcs), ", "))
	}
	for _, knownVdcName := range knownVdcs {
		check.Assert(contains(knownVdcName, foundVdcNames), Equals, true)
	}
}

func (vcd *TestVCD) Test_NsxtGlobalDefaultSegmentProfileTemplate(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	vcd.skipIfNotSysAdmin(check)

	nsxtManager, err := vcd.client.GetNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(nsxtManager, NotNil)
	nsxtManagerUrn, err := nsxtManager.Urn()
	check.Assert(err, IsNil)

	// Filter by NSX-T Manager
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("nsxTManagerRef.id==%s", nsxtManagerUrn), queryParams)

	// Lookup prerequisite profiles for Segment Profile template creation
	ipDiscoveryProfile, err := vcd.client.GetIpDiscoveryProfileByName(vcd.config.VCD.Nsxt.IpDiscoveryProfile, queryParams)
	check.Assert(err, IsNil)
	macDiscoveryProfile, err := vcd.client.GetMacDiscoveryProfileByName(vcd.config.VCD.Nsxt.MacDiscoveryProfile, queryParams)
	check.Assert(err, IsNil)
	spoofGuardProfile, err := vcd.client.GetSpoofGuardProfileByName(vcd.config.VCD.Nsxt.SpoofGuardProfile, queryParams)
	check.Assert(err, IsNil)
	qosProfile, err := vcd.client.GetQoSProfileByName(vcd.config.VCD.Nsxt.QosProfile, queryParams)
	check.Assert(err, IsNil)
	segmentSecurityProfile, err := vcd.client.GetSegmentSecurityProfileByName(vcd.config.VCD.Nsxt.SegmentSecurityProfile, queryParams)
	check.Assert(err, IsNil)

	config := &types.NsxtSegmentProfileTemplate{
		Name:                   check.TestName(),
		Description:            check.TestName() + "-description",
		IPDiscoveryProfile:     &types.Reference{ID: ipDiscoveryProfile.ID},
		MacDiscoveryProfile:    &types.Reference{ID: macDiscoveryProfile.ID},
		QosProfile:             &types.Reference{ID: qosProfile.ID},
		SegmentSecurityProfile: &types.Reference{ID: segmentSecurityProfile.ID},
		SpoofGuardProfile:      &types.Reference{ID: spoofGuardProfile.ID},
		SourceNsxTManagerRef:   &types.OpenApiReference{ID: nsxtManager.NsxtManager.ID},
	}

	createdSegmentProfileTemplate, err := vcd.client.CreateSegmentProfileTemplate(config)
	check.Assert(err, IsNil)
	check.Assert(createdSegmentProfileTemplate, NotNil)

	// Add to cleanup list
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates + createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID
	AddToCleanupListOpenApi(config.Name, check.TestName(), openApiEndpoint)

	// Set global profile template
	globalDefaultSegmentProfileConfig := &types.NsxtGlobalDefaultSegmentProfileTemplate{
		VappNetworkSegmentProfileTemplateRef: &types.OpenApiReference{ID: createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID},
		VdcNetworkSegmentProfileTemplateRef:  &types.OpenApiReference{ID: createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID},
	}

	updatedDefaults, err := vcd.client.UpdateGlobalDefaultSegmentProfileTemplates(globalDefaultSegmentProfileConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedDefaults, NotNil)
	check.Assert(updatedDefaults.VappNetworkSegmentProfileTemplateRef.ID, Equals, createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID)
	check.Assert(updatedDefaults.VdcNetworkSegmentProfileTemplateRef.ID, Equals, createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID)

	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtGlobalDefaultSegmentProfileTemplates
	PrependToCleanupList(openApiEndpoint, "OpenApiEntityGlobalDefaultSegmentProfileTemplate", "", check.TestName())

	// Cleanup
	resetDefaults, err := vcd.client.UpdateGlobalDefaultSegmentProfileTemplates(&types.NsxtGlobalDefaultSegmentProfileTemplate{})
	check.Assert(err, IsNil)
	check.Assert(resetDefaults, NotNil)
	check.Assert(resetDefaults.VappNetworkSegmentProfileTemplateRef, IsNil)
	check.Assert(resetDefaults.VdcNetworkSegmentProfileTemplateRef, IsNil)

	err = createdSegmentProfileTemplate.Delete()
	check.Assert(err, IsNil)
}

// Test retrieval of all Orgs
func (vcd *TestVCD) Test_QueryAllOrgs(check *C) {
	vcd.skipIfNotSysAdmin(check)
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_QueryOrgByName: Org Name not given")
		return
	}

	orgs, err := vcd.client.QueryAllOrgs()
	check.Assert(err, IsNil)
	check.Assert(orgs, NotNil)

	foundOrg := false
	for _, org := range orgs {
		if org.Name == vcd.config.VCD.Org {
			foundOrg = true
		}
	}
	check.Assert(foundOrg, Equals, true)
}

// Tests Org retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_QueryOrgByName(check *C) {
	vcd.skipIfNotSysAdmin(check)
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_QueryOrgByName: Org Name not given")
		return
	}

	org, err := vcd.client.QueryOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	orgFound := false
	if vcd.config.VCD.Org == org.Name {
		orgFound = true
	}

	if testVerbose {
		fmt.Printf("Org %s\n", org.Name)
		fmt.Printf("\t href    %s\n", org.HREF)
		fmt.Printf("\t enabled %v\n", org.IsEnabled)
		fmt.Println("")
	}

	check.Assert(orgFound, Equals, true)
}

// Tests Org retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_QueryOrgById(check *C) {
	vcd.skipIfNotSysAdmin(check)
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_QueryOrgByName: Org Name not given")
		return
	}

	namedOrg, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	orgFound := false
	if vcd.config.VCD.Org == namedOrg.Org.Name {

		idOrg, err := vcd.client.QueryOrgByID(namedOrg.Org.ID)
		check.Assert(err, IsNil)

		if idOrg.HREF == namedOrg.Org.HREF {
			orgFound = true
		}

		if testVerbose {
			fmt.Printf("Org %s\n", namedOrg.Org.Name)
			fmt.Printf("\t Org HREF (by Name): %s\n", namedOrg.Org.HREF)
			fmt.Printf("\t Org HREF (by ID): %s\n", idOrg.HREF)
			fmt.Println("")
		}
	}
	check.Assert(orgFound, Equals, true)
}
