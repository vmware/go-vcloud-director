//go:build functional || openapi || vdcGroup || ALL
// +build functional openapi vdcGroup ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// create only support NSXT VDCs
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
func test_CreateVdcGroup(check *C, adminOrg *AdminOrg, vcd *TestVCD) {
	createdVdc := createNewVdc(vcd, check, check.TestName())

	createdVdcAsCandidate, err := adminOrg.GetAllNsxtCandidateVdcs(createdVdc.vdcId(),
		map[string][]string{"filter": []string{fmt.Sprintf("name==%s", url.QueryEscape(createdVdc.vdcName()))}})
	check.Assert(err, IsNil)
	check.Assert(createdVdcAsCandidate, NotNil)
	check.Assert(len(*createdVdcAsCandidate) == 1, Equals, true)

	existingVdcAsCandidate, err := adminOrg.GetAllNsxtCandidateVdcs(createdVdc.vdcId(),
		map[string][]string{"filter": []string{fmt.Sprintf("name==%s", url.QueryEscape(vcd.nsxtVdc.vdcName()))}})
	check.Assert(err, IsNil)
	check.Assert(existingVdcAsCandidate, NotNil)
	check.Assert(len(*existingVdcAsCandidate) == 1, Equals, true)

	vdcGroupConfig := &types.VdcGroup{
		Name:  check.TestName() + "Group",
		OrgId: adminOrg.orgId(),
		ParticipatingOrgVdcs: []types.ParticipatingOrgVdcs{
			types.ParticipatingOrgVdcs{
				VdcRef: types.OpenApiReference{
					ID: createdVdc.vdcId(),
				},
				SiteRef: (*createdVdcAsCandidate)[0].SiteRef,
				OrgRef:  (*createdVdcAsCandidate)[0].OrgRef,
			},
			types.ParticipatingOrgVdcs{
				VdcRef: types.OpenApiReference{
					ID: vcd.nsxtVdc.vdcId(),
				},
				SiteRef: (*existingVdcAsCandidate)[0].SiteRef,
				OrgRef:  (*existingVdcAsCandidate)[0].OrgRef,
			},
		},
		LocalEgress:                false,
		UniversalNetworkingEnabled: false,
		NetworkProviderType:        "NSX_T",
		Type:                       "LOCAL",
		//DfwEnabled: true, // ignored by API
	}

	vdcGroup, err := adminOrg.CreateVdcGroup(vdcGroupConfig)
	check.Assert(err, IsNil)
	check.Assert(vdcGroup, NotNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups + vdcGroup.VdcGroup.Id
	PrependToCleanupListOpenApi(vdcGroup.VdcGroup.Name, check.TestName(), openApiEndpoint)
}

func createNewVdc(vcd *TestVCD, check *C, vdcName string) *Vdc {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	pVdcs, err := QueryProviderVdcByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	if len(pVdcs) == 0 {
		check.Skip(fmt.Sprintf("No NSX-T Provider VDC found with name '%s'", vcd.config.VCD.NsxtProviderVdc.Name))
	}
	providerVdcHref := pVdcs[0].HREF

	pvdcStorageProfile, err := vcd.client.QueryProviderVdcStorageProfileByName(vcd.config.VCD.NsxtProviderVdc.StorageProfile, providerVdcHref)

	check.Assert(err, IsNil)
	providerVdcStorageProfileHref := pvdcStorageProfile.HREF

	networkPools, err := QueryNetworkPoolByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.NetworkPool)
	check.Assert(err, IsNil)
	if len(networkPools) == 0 {
		check.Skip(fmt.Sprintf("No network pool found with name '%s'", vcd.config.VCD.NsxtProviderVdc.NetworkPool))
	}

	networkPoolHref := networkPools[0].HREF
	trueValue := true
	vdcConfiguration := &types.VdcConfiguration{
		Name:            vdcName,
		Xmlns:           types.XMLNamespaceVCloud,
		AllocationModel: "Flex",
		ComputeCapacity: []*types.ComputeCapacity{
			&types.ComputeCapacity{
				CPU: &types.CapacityWithUsage{
					Units:     "MHz",
					Allocated: 1024,
					Limit:     1024,
				},
				Memory: &types.CapacityWithUsage{
					Allocated: 1024,
					Limit:     1024,
					Units:     "MB",
				},
			},
		},
		VdcStorageProfile: []*types.VdcStorageProfileConfiguration{&types.VdcStorageProfileConfiguration{
			Enabled: true,
			Units:   "MB",
			Limit:   1024,
			Default: true,
			ProviderVdcStorageProfile: &types.Reference{
				HREF: providerVdcStorageProfileHref,
			},
		},
		},
		NetworkPoolReference: &types.Reference{
			HREF: networkPoolHref,
		},
		ProviderVdcReference: &types.Reference{
			HREF: providerVdcHref,
		},
		IsEnabled:             true,
		IsThinProvision:       true,
		UsesFastProvisioning:  true,
		IsElastic:             &trueValue,
		IncludeMemoryOverhead: &trueValue,
	}

	vdc, err := adminOrg.CreateOrgVdc(vdcConfiguration)
	check.Assert(vdc, NotNil)
	check.Assert(err, IsNil)

	AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, check.TestName())
	return vdc
}

// create only support NSXT VDCs
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

	disabledVdcGroup, err := updatedVdcGroup.DeActivateDfw()
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

	rightsToRemove, orgAdminClient, err := newOrgAdminUserWithVdcGroupRightsConnection(check, adminOrg, "test-user", "CHANGE-ME", vcd.config.Provider.Url, true)
	check.Assert(err, IsNil)
	check.Assert(orgAdminClient, NotNil)

	//cleanup
	defer cleanupRightsAndBundle(check, adminOrg, rightsToRemove)

	orgAsOrgAdminUser, err := orgAdminClient.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(orgAsOrgAdminUser, NotNil)

	//run tests ad org Admin with needed rights
	test_NsxtVdcGroup(check, adminOrg, vcd)
	test_CreateVdcGroup(check, adminOrg, vcd)
	test_GetVdcGroupByName_ValidatesSymbolsInName(check, orgAsOrgAdminUser, vcd.nsxtVdc.vdcId())

}

// newOrgAdminUserWithVdcGroupRightsConnection creates a new Org Admin User with VDC group rights
// and returns a connection to it
func newOrgAdminUserWithVdcGroupRightsConnection(check *C, adminOrg *AdminOrg, userName, password, href string, insecure bool) ([]types.OpenApiReference, *VCDClient, error) {
	defaultRightsBundle, err := adminOrg.client.GetRightsBundleByName("Default Rights Bundle")
	check.Assert(err, IsNil)
	check.Assert(defaultRightsBundle, NotNil)

	// add new rights to bundle
	var rightsToAdd []types.OpenApiReference

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
		if foundRight {
			fmt.Printf("Right %s already in Default Rights Bundle\n", rightName)
			// ignore
		} else {
			rightsToAdd = append(rightsToAdd, types.OpenApiReference{Name: newRight.Name, ID: newRight.ID})
		}
	}

	if len(rightsToAdd) > 0 {
		err = defaultRightsBundle.AddRights(rightsToAdd)
		check.Assert(err, IsNil)
		rights, err := defaultRightsBundle.GetRights(nil)
		check.Assert(err, IsNil)
		check.Assert(len(rights), Equals, len(rightsBeforeChange)+len(rightsToAdd))
	}

	connection, err := newOrgUserConnection(adminOrg, userName, password, href, insecure)
	return rightsToAdd, connection, err
}

func cleanupRightsAndBundle(check *C, adminOrg *AdminOrg, rightsToRemove []types.OpenApiReference) {
	if len(rightsToRemove) > 0 {
		defaultRightsBundle, err := adminOrg.client.GetRightsBundleByName("Default Rights Bundle")
		check.Assert(err, IsNil)
		check.Assert(defaultRightsBundle, NotNil)

		rightsBeforeChange, err := defaultRightsBundle.GetRights(nil)
		check.Assert(err, IsNil)
		err = defaultRightsBundle.RemoveRights(rightsToRemove)
		check.Assert(err, IsNil)
		rightsAfterRemoval, err := defaultRightsBundle.GetRights(nil)
		check.Assert(err, IsNil)
		check.Assert(len(rightsAfterRemoval), Equals, len(rightsBeforeChange)-len(rightsToRemove))
	}
}
