//go:build functional || openapi || vdcGroup || nsxt || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// tests creation of NSX-T VDCs group
func (vcd *TestVCD) Test_CreateVdcGroup(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	if vcd.config.VCD.Nsxt.Vdc == "" {
		check.Skip("Missing NSX-T config: No NSX-T VDC specified")
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVdcGroups)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	test_CreateVdcGroup(check, adminOrg, vcd)
}

// tests creation of NSX-T VDCs group
func (vcd *TestVCD) Test_NsxtVdcGroup(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	if vcd.config.VCD.Nsxt.Vdc == "" {
		check.Skip("Missing NSX-T config: No NSX-T VDC specified")
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVdcGroups)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	test_NsxtVdcGroup(check, adminOrg, vcd)
}

func test_NsxtVdcGroup(check *C, adminOrg *AdminOrg, vcd *TestVCD) {
	description := "vdc group created by test"

	_, err := adminOrg.CreateNsxtVdcGroup(check.TestName(), description, vcd.nsxtVdc.vdcId(), []string{vcd.vdc.vdcId()})
	check.Assert(err, NotNil)

	vdcGroup, err := adminOrg.CreateNsxtVdcGroup(check.TestName(), description, vcd.nsxtVdc.vdcId(), []string{vcd.nsxtVdc.vdcId()})
	check.Assert(err, IsNil)
	check.Assert(vdcGroup, NotNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups + vdcGroup.VdcGroup.Id
	PrependToCleanupListOpenApi(vdcGroup.VdcGroup.Name, check.TestName(), openApiEndpoint)

	check.Assert(vdcGroup.VdcGroup.Description, Equals, description)
	check.Assert(vdcGroup.VdcGroup.DfwEnabled, Equals, false)
	check.Assert(len(vdcGroup.VdcGroup.ParticipatingOrgVdcs), Equals, 1)
	check.Assert(vdcGroup.VdcGroup.OrgId, Equals, adminOrg.AdminOrg.ID)
	check.Assert(vdcGroup.VdcGroup.Name, Equals, check.TestName())
	check.Assert(vdcGroup.VdcGroup.LocalEgress, Equals, false)
	check.Assert(vdcGroup.VdcGroup.UniversalNetworkingEnabled, Equals, false)
	check.Assert(vdcGroup.VdcGroup.NetworkProviderType, Equals, "NSX_T")
	check.Assert(vdcGroup.VdcGroup.Type, Equals, "LOCAL")

	// check fetching by ID
	foundVdcGroup, err := adminOrg.GetVdcGroupById(vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)
	check.Assert(foundVdcGroup, NotNil)
	check.Assert(foundVdcGroup.VdcGroup.Name, Equals, vdcGroup.VdcGroup.Name)
	check.Assert(foundVdcGroup.VdcGroup.Description, Equals, vdcGroup.VdcGroup.Description)
	check.Assert(len(foundVdcGroup.VdcGroup.ParticipatingOrgVdcs), Equals, len(vdcGroup.VdcGroup.ParticipatingOrgVdcs))

	// check fetching all VDC groups
	allVdcGroups, err := adminOrg.GetAllVdcGroups(nil)
	check.Assert(err, IsNil)
	check.Assert(allVdcGroups, NotNil)

	if testVerbose {
		fmt.Printf("(org) how many VDC groups: %d\n", len(allVdcGroups))
		for i, oneVdcGroup := range allVdcGroups {
			fmt.Printf("%3d %-20s %-53s %s\n", i, oneVdcGroup.VdcGroup.Name, oneVdcGroup.VdcGroup.Id,
				oneVdcGroup.VdcGroup.Description)
		}
	}

	// check fetching VDC group by Name
	createdVdc := createNewVdc(vcd, check, check.TestName()+"_forUpdate")
	check.Assert(err, IsNil)
	check.Assert(createdVdc, NotNil)

	foundVdcGroup, err = adminOrg.GetVdcGroupByName(check.TestName())
	check.Assert(err, IsNil)
	check.Assert(foundVdcGroup, NotNil)
	check.Assert(foundVdcGroup.VdcGroup.Name, Equals, check.TestName())

	// check update
	newDescription := "newDescription"
	newName := check.TestName() + "newName"
	updatedVdcGroup, err := foundVdcGroup.Update(newName, newDescription, []string{createdVdc.vdcId()})
	check.Assert(err, IsNil)
	check.Assert(updatedVdcGroup, NotNil)
	check.Assert(updatedVdcGroup.VdcGroup.Name, Equals, newName)
	check.Assert(updatedVdcGroup.VdcGroup.Description, Equals, newDescription)
	check.Assert(updatedVdcGroup.VdcGroup.Id, Not(Equals), "")
	check.Assert(len(updatedVdcGroup.VdcGroup.ParticipatingOrgVdcs), Equals, 1)

	// activate and deactivate DFW
	enabledVdcGroup, err := updatedVdcGroup.ActivateDfw()
	check.Assert(err, IsNil)
	check.Assert(enabledVdcGroup, NotNil)
	check.Assert(enabledVdcGroup.VdcGroup.DfwEnabled, Equals, true)

	// disable default policy, otherwise deactivation of Dfw fails
	_, err = enabledVdcGroup.DisableDefaultPolicy()
	check.Assert(err, IsNil)
	defaultPolicy, err := enabledVdcGroup.GetDfwPolicies()
	check.Assert(err, IsNil)
	check.Assert(defaultPolicy, NotNil)
	check.Assert(*defaultPolicy.DefaultPolicy.Enabled, Equals, false)

	// also validate enable default policy
	_, err = enabledVdcGroup.EnableDefaultPolicy()
	check.Assert(err, IsNil)
	defaultPolicy, err = enabledVdcGroup.GetDfwPolicies()
	check.Assert(err, IsNil)
	check.Assert(defaultPolicy, NotNil)
	check.Assert(*defaultPolicy.DefaultPolicy.Enabled, Equals, true)

	_, err = enabledVdcGroup.DisableDefaultPolicy()
	check.Assert(err, IsNil)
	defaultPolicy, err = enabledVdcGroup.GetDfwPolicies()
	check.Assert(err, IsNil)
	check.Assert(defaultPolicy, NotNil)
	check.Assert(*defaultPolicy.DefaultPolicy.Enabled, Equals, false)

	disabledVdcGroup, err := updatedVdcGroup.DeactivateDfw()
	check.Assert(err, IsNil)
	check.Assert(disabledVdcGroup, NotNil)
	check.Assert(disabledVdcGroup.VdcGroup.DfwEnabled, Equals, false)

}

func (vcd *TestVCD) Test_GetVdcGroupByName_ValidatesSymbolsInName(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	if vcd.config.VCD.Nsxt.Vdc == "" {
		check.Skip("Missing NSX-T config: No NSX-T VDC specified")
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVdcGroups)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	test_GetVdcGroupByName_ValidatesSymbolsInName(check, adminOrg, vcd.nsxtVdc.vdcId())
}

func test_GetVdcGroupByName_ValidatesSymbolsInName(check *C, adminOrg *AdminOrg, vdcId string) {
	// When alias contains commas, semicolons, stars, or plus signs, the encoding may reject by the API when we try to Query it
	// Also, spaces present their own issues
	for _, symbol := range []string{";", ",", "+", " ", "*"} {

		name := fmt.Sprintf("Test%sVdcGroup", symbol)

		createdVdcGroup, err := adminOrg.CreateNsxtVdcGroup(name, "", vdcId, []string{vdcId})
		check.Assert(err, IsNil)
		check.Assert(createdVdcGroup, NotNil)

		openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups + createdVdcGroup.VdcGroup.Id
		PrependToCleanupListOpenApi(createdVdcGroup.VdcGroup.Name, check.TestName(), openApiEndpoint)

		check.Assert(createdVdcGroup, NotNil)
		check.Assert(createdVdcGroup.VdcGroup.Id, Not(Equals), "")
		check.Assert(createdVdcGroup.VdcGroup.Name, Equals, name)
		check.Assert(len(createdVdcGroup.VdcGroup.ParticipatingOrgVdcs), Equals, 1)

		foundVdcGroup, err := adminOrg.GetVdcGroupByName(name)
		check.Assert(err, IsNil)
		check.Assert(foundVdcGroup, NotNil)
		check.Assert(foundVdcGroup.VdcGroup.Name, Equals, name)

		err = foundVdcGroup.Delete()
		check.Assert(err, IsNil)
	}
}

// Test_NsxtVdcGroupWithOrgAdmin additionally tests Test_CreateVdcGroup,  Test_GetVdcGroupByName_ValidatesSymbolsInName
// and Test_NsxtVdcGroup using an org amin user with added rights which allows working with VDC groups.
func (vcd *TestVCD) Test_NsxtVdcGroupWithOrgAdmin(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	if vcd.config.VCD.Nsxt.Vdc == "" {
		check.Skip("Missing NSX-T config: No NSX-T VDC specified")
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVdcGroups)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	skipIfNeededRightsMissing(check, adminOrg)
	orgAdminClient, _, err := newOrgUserConnection(adminOrg, "test-user2", "CHANGE-ME", vcd.config.Provider.Url, true)
	check.Assert(err, IsNil)
	check.Assert(orgAdminClient, NotNil)

	orgAsOrgAdminUser, err := orgAdminClient.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(orgAsOrgAdminUser, NotNil)

	//run tests ad org Admin with needed rights
	test_NsxtVdcGroup(check, adminOrg, vcd)
	test_CreateVdcGroup(check, adminOrg, vcd)
	test_GetVdcGroupByName_ValidatesSymbolsInName(check, orgAsOrgAdminUser, vcd.nsxtVdc.vdcId())

}

// skipIfNeededRightsMissing checks if needed rights are configured
func skipIfNeededRightsMissing(check *C, adminOrg *AdminOrg) {
	defaultRightsBundle, err := adminOrg.client.GetRightsBundleByName("Default Rights Bundle")
	check.Assert(err, IsNil)
	check.Assert(defaultRightsBundle, NotNil)

	// add new rights to bundle
	var missingRights []string

	rightsBeforeChange, err := defaultRightsBundle.GetRights(nil)
	check.Assert(err, IsNil)
	for _, rightName := range []string{
		"vDC Group: Configure",
		"vDC Group: Configure Logging",
		"vDC Group: View",
		"Organization vDC Distributed Firewall: Enable/Disable",
		//"Security Tag Edit", 10.2 doesn't have it and for this kind testing not needed
	} {
		newRight, err := adminOrg.client.GetRightByName(rightName)
		check.Assert(err, IsNil)
		check.Assert(newRight, NotNil)
		foundRight := false
		for _, old := range rightsBeforeChange {
			if old.Name == rightName {
				foundRight = true
			}
		}
		if !foundRight {
			missingRights = append(missingRights, newRight.Name)
		}
	}

	if len(missingRights) > 0 {
		check.Skip(check.TestName() + "missing rights to run test: " + strings.Join(missingRights, ", "))
	}
}
