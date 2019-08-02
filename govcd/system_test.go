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

// Tests System methods vcd.GetOrg* by checking if the org object
// returned has the same name as the one provided in the config file.
// Asserts an error if the names don't match or if the function returned
// an error. Also tests an org that doesn't exist. Asserts an error
// if the function finds it or if the error is nil.
// Repeats the same operations fot GetOrgById, GetOrgByNameOrId, GetOrg
func (vcd *TestVCD) Test_GetOrgByNameOrId(check *C) {

	// Test Get by name
	orgName := vcd.config.VCD.Org
	org1, err1 := vcd.client.GetOrgByName(orgName)
	check.Assert(org1, NotNil)
	check.Assert(err1, IsNil)
	check.Assert(org1.Org.Name, Equals, orgName)
	orgId := org1.Org.ID
	// Tests Org that doesn't exist
	org1, err1 = vcd.client.GetOrgByName(INVALID_NAME)
	check.Assert(org1, IsNil)
	// When we explicitly search for a non existing item, we expect the error to be not nil
	check.Assert(err1, NotNil)
	check.Assert(IsNotFound(err1), Equals, true)

	// Test Get by ID
	org2, err2 := vcd.client.GetOrgById(orgId)
	check.Assert(org2, NotNil)
	check.Assert(err2, IsNil)
	check.Assert(org2.Org.Name, Equals, orgName)
	check.Assert(org2.Org.ID, Equals, orgId)
	org2, err2 = vcd.client.GetOrgById(invalidEntityId)
	check.Assert(org2, IsNil)
	check.Assert(err2, NotNil)
	check.Assert(IsNotFound(err2), Equals, true)

	// Test Get by name or ID using the ID
	org3, err3 := vcd.client.GetOrgByNameOrId(orgId)
	check.Assert(org3, NotNil)
	check.Assert(err3, IsNil)
	check.Assert(org3.Org.Name, Equals, orgName)
	check.Assert(org3.Org.ID, Equals, orgId)
	org3, err3 = vcd.client.GetOrgByNameOrId(invalidEntityId)
	check.Assert(org3, IsNil)
	check.Assert(err3, NotNil)
	check.Assert(IsNotFound(err3), Equals, true)

	// Test Get by name or ID using the name
	org4, err4 := vcd.client.GetOrgByNameOrId(orgName)
	check.Assert(org4, NotNil)
	check.Assert(err4, IsNil)
	check.Assert(org4.Org.Name, Equals, orgName)
	check.Assert(org4.Org.ID, Equals, orgId)
	org4, err4 = vcd.client.GetOrgByNameOrId(INVALID_NAME)
	check.Assert(org4, IsNil)
	check.Assert(err4, NotNil)
	check.Assert(IsNotFound(err4), Equals, true)
}

// Tests System methods vcd.GetAdminOrg* by checking if the adminOrg object
// returned has the same name as the one provided in the config file.
// Asserts an error if the names don't match or if the function returned
// an error. Also tests an admin org that doesn't exist. Asserts an error
// if the function finds it or if the error is nil.
// Repeats the same operations fot GetAdminOrgById, GetAdminOrgByNameOrId, GetAdminOrg
func (vcd *TestVCD) Test_GetAdminOrgByNameOrId(check *C) {

	// Test Get by name
	adminOrgName := vcd.config.VCD.Org
	adminOrg1, err1 := vcd.client.GetAdminOrgByName(adminOrgName)
	check.Assert(adminOrg1, NotNil)
	check.Assert(err1, IsNil)
	check.Assert(adminOrg1.AdminOrg.Name, Equals, adminOrgName)
	orgId := adminOrg1.AdminOrg.ID
	// Tests Org that doesn't exist
	adminOrg1, err1 = vcd.client.GetAdminOrgByName(INVALID_NAME)
	check.Assert(adminOrg1, IsNil)
	// When we explicitly search for a non existing item, we expect the error to be not nil
	check.Assert(err1, NotNil)
	check.Assert(IsNotFound(err1), Equals, true)

	// Test Get by ID
	adminOrg2, err2 := vcd.client.GetAdminOrgById(orgId)
	check.Assert(adminOrg2, NotNil)
	check.Assert(err2, IsNil)
	check.Assert(adminOrg2.AdminOrg.Name, Equals, adminOrgName)
	check.Assert(adminOrg2.AdminOrg.ID, Equals, orgId)
	adminOrg2, err2 = vcd.client.GetAdminOrgById(invalidEntityId)
	check.Assert(adminOrg2, IsNil)
	check.Assert(err2, NotNil)
	check.Assert(IsNotFound(err2), Equals, true)

	// Test Get by name or ID using the ID
	adminOrg3, err3 := vcd.client.GetAdminOrgByNameOrId(orgId)
	check.Assert(adminOrg3, NotNil)
	check.Assert(err3, IsNil)
	check.Assert(adminOrg3.AdminOrg.Name, Equals, adminOrgName)
	check.Assert(adminOrg3.AdminOrg.ID, Equals, orgId)
	adminOrg3, err3 = vcd.client.GetAdminOrgByNameOrId(invalidEntityId)
	check.Assert(adminOrg3, IsNil)
	check.Assert(err3, NotNil)
	check.Assert(IsNotFound(err3), Equals, true)

	// Test Get by name or ID using the name
	adminOrg4, err4 := vcd.client.GetAdminOrgByNameOrId(adminOrgName)
	check.Assert(adminOrg4, NotNil)
	check.Assert(err4, IsNil)
	check.Assert(adminOrg4.AdminOrg.Name, Equals, adminOrgName)
	check.Assert(adminOrg4.AdminOrg.ID, Equals, orgId)
	adminOrg4, err4 = vcd.client.GetAdminOrgByNameOrId(INVALID_NAME)
	check.Assert(adminOrg4, IsNil)
	check.Assert(err4, NotNil)
	check.Assert(IsNotFound(err4), Equals, true)
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
			edge, err = vcd.vdc.FindEdgeGateway(egc.Name)
			check.Assert(err, IsNil)
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
		edge, err = vcd.vdc.FindEdgeGateway(egc.Name)
		check.Assert(err, NotNil)
		check.Assert(edge, Equals, EdgeGateway{})
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

	adminVdc, err := adminOrg.GetAdminVdcByName(vcd.config.VCD.Vdc)
	check.Assert(adminVdc, Not(Equals), AdminVdc{})
	check.Assert(err, IsNil)

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
