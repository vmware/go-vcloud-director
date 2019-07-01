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

// Tests System function GetOrgByName by checking if the org object
// return has the same name as the one provided in the config file.
// Asserts an error if the names don't match or if the function returned
// an error. Also tests an org that doesn't exist. Asserts an error
// if the function finds it or if the error is not nil.
func (vcd *TestVCD) Test_GetOrgByName(check *C) {
	org, err := GetOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(org, Not(Equals), Org{})
	check.Assert(err, IsNil)
	check.Assert(org.Org.Name, Equals, vcd.config.VCD.Org)
	// Tests Org That doesn't exist
	org, err = GetOrgByName(vcd.client, INVALID_NAME)
	check.Assert(org, Equals, Org{})
	// When we explicitly search for a non existing item, we expect the error to be not nil
	check.Assert(err, NotNil)
}

// Tests System function GetAdminOrgByName by checking if the AdminOrg object
// return has the same name as the one provided in the config file. Asserts
// an error if the names don't match or if the function returned an error.
// Also tests an org that doesn't exist. Asserts an error
// if the function finds it or if the error is not nil.
func (vcd *TestVCD) Test_GetAdminOrgByName(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, vcd.config.VCD.Org)
	// Tests Org That doesn't exist
	org, err = GetAdminOrgByName(vcd.client, INVALID_NAME)
	check.Assert(org, Equals, AdminOrg{})
	check.Assert(err, NotNil)
}

// Tests the creation of an org with general settings,
// org vapp template settings, and orgldapsettings. Asserts an
// error if the task, fetching the org, or deleting the org fails
func (vcd *TestVCD) Test_CreateOrg(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	org, _ := GetAdminOrgByName(vcd.client, TestCreateOrg)
	if org != (AdminOrg{}) {
		err := org.Delete(true, true)
		check.Assert(err, IsNil)
	}
	settings := &types.OrgSettings{
		OrgGeneralSettings: &types.OrgGeneralSettings{
			CanPublishCatalogs:       true,
			DeployedVMQuota:          10,
			StoredVMQuota:            10,
			UseServerBootSequence:    true,
			DelayAfterPowerOnSeconds: 3,
		},
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
	task, err := CreateOrg(vcd.client, TestCreateOrg, TestCreateOrg, TestCreateOrg, settings, true)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestCreateOrg, "org", "", "TestCreateOrg")
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	// fetch newly created org
	org, err = GetAdminOrgByName(vcd.client, TestCreateOrg)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, TestCreateOrg)
	check.Assert(org.AdminOrg.Description, Equals, TestCreateOrg)
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	doesOrgExist(check, vcd)
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

	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVdcByName(vcd.config.VCD.Vdc)
	check.Assert(adminVdc, Not(Equals), AdminVdc{})
	check.Assert(err, IsNil)

	// Get network pool by href
	foundNetworkPool, err := GetNetworkPoolByHREF(vcd.client, adminVdc.AdminVdc.NetworkPoolReference.HREF)
	check.Assert(err, IsNil)
	check.Assert(foundNetworkPool, Not(Equals), types.VMWNetworkPool{})
}

// longer than the 128 characters so nothing can be named this
var INVALID_NAME = `*******************************************INVALID
					****************************************************
					************************`

func init() {
	testingTags["system"] = "system_test.go"
}
