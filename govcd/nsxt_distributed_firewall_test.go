//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxtDistributedFirewall creates a list of distributed firewall rules with randomized
// parameters in two modes:
// * System user
// * Org Admin user
func (vcd *TestVCD) Test_NsxtDistributedFirewallRules(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)
	vcd.skipIfNotSysAdmin(check)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(nsxtExternalNetwork, NotNil)

	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)
	check.Assert(vdc, NotNil)
	check.Assert(vdcGroup, NotNil)

	// Run firewall tests as System user
	fmt.Println("# Running Distributed Firewall tests as 'System' user")
	test_NsxtDistributedFirewallRules(vcd, check, vdcGroup.VdcGroup.Id, vcd.client, vdc)

	// Prep Org admin user and run firewall tests
	userName := strings.ToLower(check.TestName())
	fmt.Printf("# Running Distributed Firewall tests as Org Admin user '%s'\n", userName)
	orgUserVcdClient, _, err := newOrgUserConnection(adminOrg, userName, "CHANGE-ME", vcd.config.Provider.Url, true)
	check.Assert(err, IsNil)
	orgUserOrgAdmin, err := orgUserVcdClient.GetAdminOrgById(adminOrg.AdminOrg.ID)
	check.Assert(err, IsNil)
	orgUserVdc, err := orgUserOrgAdmin.GetVDCById(vdc.Vdc.ID, false)
	check.Assert(err, IsNil)
	test_NsxtDistributedFirewallRules(vcd, check, vdcGroup.VdcGroup.Id, orgUserVcdClient, orgUserVdc)

	// Cleanup
	err = vdcGroup.Delete()
	check.Assert(err, IsNil)
	err = vdc.DeleteWait(true, true)
	check.Assert(err, IsNil)
}

func test_NsxtDistributedFirewallRules(vcd *TestVCD, check *C, vdcGroupId string, vcdClient *VCDClient, vdc *Vdc) {
	adminOrg, err := vcdClient.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

	vdcGroup, err := adminOrg.GetVdcGroupById(vdcGroupId)
	check.Assert(err, IsNil)

	_, err = vdcGroup.ActivateDfw()
	check.Assert(err, IsNil)

	// Get existing firewall rule configuration
	fwRules, err := vdcGroup.GetDistributedFirewall()
	check.Assert(err, IsNil)
	check.Assert(fwRules.DistributedFirewallRuleContainer.Values, NotNil)

	// Create some prerequisites and generate firewall rule configurations to feed them into config
	randomizedFwRuleDefs, ipSet, secGroup := createDistributedFirewallDefinitions(check, vcd, vdcGroup.VdcGroup.Id, vcdClient, vdc)

	fwRules.DistributedFirewallRuleContainer.Values = randomizedFwRuleDefs

	if testVerbose {
		dumpDistributedFirewallRulesToScreen(randomizedFwRuleDefs)
	}

	fwUpdated, err := vdcGroup.UpdateDistributedFirewall(fwRules.DistributedFirewallRuleContainer)
	check.Assert(err, IsNil)
	check.Assert(fwUpdated, Not(IsNil))

	check.Assert(len(fwUpdated.DistributedFirewallRuleContainer.Values), Equals, len(randomizedFwRuleDefs))

	// Check that all created rules have the same attributes and order
	for index := range fwUpdated.DistributedFirewallRuleContainer.Values {
		check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].Name, Equals, randomizedFwRuleDefs[index].Name)
		check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].Direction, Equals, randomizedFwRuleDefs[index].Direction)
		check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].IpProtocol, Equals, randomizedFwRuleDefs[index].IpProtocol)
		check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].Enabled, Equals, randomizedFwRuleDefs[index].Enabled)
		check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].Logging, Equals, randomizedFwRuleDefs[index].Logging)
		check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].Comments, Equals, randomizedFwRuleDefs[index].Comments)
		check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].ActionValue, Equals, randomizedFwRuleDefs[index].ActionValue)

		for fwGroupIndex := range fwUpdated.DistributedFirewallRuleContainer.Values[index].SourceFirewallGroups {
			check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].SourceFirewallGroups[fwGroupIndex].ID, Equals, randomizedFwRuleDefs[index].SourceFirewallGroups[fwGroupIndex].ID)
		}

		for fwGroupIndex := range fwUpdated.DistributedFirewallRuleContainer.Values[index].DestinationFirewallGroups {
			check.Assert(fwUpdated.DistributedFirewallRuleContainer.Values[index].DestinationFirewallGroups[fwGroupIndex].ID, Equals, randomizedFwRuleDefs[index].DestinationFirewallGroups[fwGroupIndex].ID)
		}

		// Ensure the same amount of Application Port Profiles are assigned and created
		check.Assert(len(fwUpdated.DistributedFirewallRuleContainer.Values), Equals, len(randomizedFwRuleDefs))
		definedAppPortProfileIds := extractIdsFromOpenApiReferences(randomizedFwRuleDefs[index].ApplicationPortProfiles)
		for _, appPortProfile := range fwUpdated.DistributedFirewallRuleContainer.Values[index].ApplicationPortProfiles {
			check.Assert(contains(appPortProfile.ID, definedAppPortProfileIds), Equals, true)
		}

		// Ensure the same amount of Network Context Profiles are assigned and created
		definedNetContextProfileIds := extractIdsFromOpenApiReferences(randomizedFwRuleDefs[index].NetworkContextProfiles)
		for _, networkContextProfile := range fwUpdated.DistributedFirewallRuleContainer.Values[index].NetworkContextProfiles {
			check.Assert(contains(networkContextProfile.ID, definedNetContextProfileIds), Equals, true)
		}
	}

	// Cleanup
	err = fwRules.DeleteAllRules()
	check.Assert(err, IsNil)
	// Check that rules were removed
	newRules, err := vdcGroup.GetDistributedFirewall()
	check.Assert(err, IsNil)
	check.Assert(len(newRules.DistributedFirewallRuleContainer.Values) == 0, Equals, true)

	// Cleanup remaining setup
	_, err = vdcGroup.DisableDefaultPolicy()
	check.Assert(err, IsNil)
	_, err = vdcGroup.DeactivateDfw()
	check.Assert(err, IsNil)
	err = ipSet.Delete()
	check.Assert(err, IsNil)
	err = secGroup.Delete()
	check.Assert(err, IsNil)
}

// createDistributedFirewallDefinitions creates some randomized firewall rule configurations to match possible configurations
func createDistributedFirewallDefinitions(check *C, vcd *TestVCD, vdcGroupId string, vcdClient *VCDClient, vdc *Vdc) ([]*types.DistributedFirewallRule, *NsxtFirewallGroup, *NsxtFirewallGroup) {
	// This number does not impact performance because all rules are created at once in the API
	numberOfRules := 40

	// Pre-Create Firewall Groups (IP Set and Security Group to randomly configure them)
	ipSet := preCreateVdcGroupIpSet(check, vcd, vdcGroupId, vdc)
	secGroup := preCreateVdcGroupSecurityGroup(check, vcd, vdcGroupId, vdc)
	fwGroupIds := []string{ipSet.NsxtFirewallGroup.ID, secGroup.NsxtFirewallGroup.ID}
	fwGroupRefs := convertSliceOfStringsToOpenApiReferenceIds(fwGroupIds)
	appPortProfileReferences := getRandomListOfAppPortProfiles(check, vcd)
	networkContextProfiles := getRandomListOfNetworkContextProfiles(check, vcd, vcdClient)

	firewallRules := make([]*types.DistributedFirewallRule, numberOfRules)
	for a := 0; a < numberOfRules; a++ {

		// Feed in empty value for source and destination or a firewall group
		src := pickRandomOpenApiRefOrEmpty(fwGroupRefs)
		var srcValue []types.OpenApiReference
		dst := pickRandomOpenApiRefOrEmpty(fwGroupRefs)
		var dstValue []types.OpenApiReference
		if src != (types.OpenApiReference{}) {
			srcValue = []types.OpenApiReference{src}
		}
		if dst != (types.OpenApiReference{}) {
			dstValue = []types.OpenApiReference{dst}
		}

		firewallRules[a] = &types.DistributedFirewallRule{
			Name:                      check.TestName() + strconv.Itoa(a),
			ActionValue:               pickRandomString([]string{"ALLOW", "DROP", "REJECT"}),
			Enabled:                   a%2 == 0,
			SourceFirewallGroups:      srcValue,
			DestinationFirewallGroups: dstValue,
			ApplicationPortProfiles:   appPortProfileReferences[0:a],
			IpProtocol:                pickRandomString([]string{"IPV6", "IPV4", "IPV4_IPV6"}),
			Logging:                   a%2 == 1,
			Direction:                 pickRandomString([]string{"IN", "OUT", "IN_OUT"}),
		}

		// Network Context Profile can usually work with up to one Application Profile therefore this
		// needs to be explicitly preset
		if a%5 == 1 { // Every fifth rule
			netCtxProfile := networkContextProfiles[0:a]
			networkContextProfile := make([]types.OpenApiReference, 0)
			for _, netCtxProf := range netCtxProfile {
				if netCtxProf.ID != "" {
					networkContextProfile = append(networkContextProfile, types.OpenApiReference{ID: netCtxProf.ID, Name: netCtxProf.Name})
				}
			}

			firewallRules[a].NetworkContextProfiles = networkContextProfile
			// firewallRules[a].ApplicationPortProfiles = appPortProfileReferences[0:1]
			firewallRules[a].ApplicationPortProfiles = nil

		}

		// API V36.2 introduced new field Comment which is shown in UI
		if vcd.client.Client.APIVCDMaxVersionIs(">= 36.2") {
			firewallRules[a].Comments = "Comment Rule"
		}

	}

	return firewallRules, ipSet, secGroup
}

func preCreateVdcGroupIpSet(check *C, vcd *TestVCD, ownerId string, nsxtVdc *Vdc) *NsxtFirewallGroup {
	ipSetDefinition := &types.NsxtFirewallGroup{
		Name:        check.TestName() + "ipset",
		Description: check.TestName() + "-Description",
		Type:        types.FirewallGroupTypeIpSet,
		OwnerRef:    &types.OpenApiReference{ID: ownerId},

		IpAddresses: []string{
			"12.12.12.1",
			"10.10.10.0/24",
			"11.11.11.1-11.11.11.2",
			// represents the block of IPv6 addresses from 2001:db8:0:0:0:0:0:0 to 2001:db8:0:ffff:ffff:ffff:ffff:ffff
			"2001:db8::/48",
			"2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
		},
	}

	// Create IP Set and add to cleanup if it was created
	createdIpSet, err := nsxtVdc.CreateNsxtFirewallGroup(ipSetDefinition)
	check.Assert(err, IsNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups + createdIpSet.NsxtFirewallGroup.ID
	PrependToCleanupListOpenApi(createdIpSet.NsxtFirewallGroup.Name, check.TestName(), openApiEndpoint)

	return createdIpSet
}

func preCreateVdcGroupSecurityGroup(check *C, vcd *TestVCD, ownerId string, nsxtVdc *Vdc) *NsxtFirewallGroup {
	fwGroupDefinition := &types.NsxtFirewallGroup{
		Name:        check.TestName() + "security-group",
		Description: check.TestName() + "-Description",
		Type:        types.FirewallGroupTypeSecurityGroup,
		OwnerRef:    &types.OpenApiReference{ID: ownerId},
	}

	// Create firewall group and add to cleanup if it was created
	createdSecGroup, err := nsxtVdc.CreateNsxtFirewallGroup(fwGroupDefinition)
	check.Assert(err, IsNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups + createdSecGroup.NsxtFirewallGroup.ID
	PrependToCleanupListOpenApi(createdSecGroup.NsxtFirewallGroup.Name, check.TestName(), openApiEndpoint)

	return createdSecGroup
}

func getRandomListOfNetworkContextProfiles(check *C, vcd *TestVCD, vdcClient *VCDClient) []types.OpenApiReference {
	networkContextProfiles, err := GetAllNetworkContextProfiles(&vcd.client.Client, nil)
	check.Assert(err, IsNil)
	openApiRefs := make([]types.OpenApiReference, 1)
	for _, networkContextProfile := range networkContextProfiles {
		// Skipping network context profile which has hardcoded destinations and throws error when used in firewall rules with specified destinations
		if strings.Contains(networkContextProfile.Description, "ALG") || strings.Contains(networkContextProfile.Description, "includes the URL categories") {
			continue
		}
		openApiRef := types.OpenApiReference{
			ID:   networkContextProfile.ID,
			Name: networkContextProfile.Name,
		}

		openApiRefs = append(openApiRefs, openApiRef)
	}

	return openApiRefs
}

func dumpDistributedFirewallRulesToScreen(rules []*types.DistributedFirewallRule) {
	fmt.Println("# The following firewall rules will be created")
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "Name\tDirection\tIP Protocol\tEnabled\tAction\tLogging\tSrc Count\tDst Count\tAppPortProfile Count\tNet Context Profile Count")

	for _, rule := range rules {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\t%t\t%d\t%d\t%d\t%d\n", rule.Name, rule.Direction, rule.IpProtocol,
			rule.Enabled, rule.Action, rule.Logging, len(rule.SourceFirewallGroups), len(rule.DestinationFirewallGroups), len(rule.ApplicationPortProfiles), len(rule.NetworkContextProfiles))
	}
	err := w.Flush()
	if err != nil {
		util.Logger.Printf("Error while dumping Distributed Firewall rules to screen: %s", err)
	}
}

// Test_NsxtDistributedFirewallRule tests the capability of managing Firewall Rules one by one using
// `DistributedFirewallRule` type.
func (vcd *TestVCD) Test_NsxtDistributedFirewallRule(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(nsxtExternalNetwork, NotNil)
	check.Assert(err, IsNil)

	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)
	check.Assert(vdc, NotNil)
	check.Assert(vdcGroup, NotNil)

	defer func() {
		// Cleanup
		err = vdcGroup.Delete()
		check.Assert(err, IsNil)
		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)
	}()

	fmt.Println("# Running Distributed Firewall tests for single Rule")
	test_NsxtDistributedFirewallRule(vcd, check, vdcGroup.VdcGroup.Id, vcd.client, vdc, true)
}

func (vcd *TestVCD) Test_NsxtDistributedFirewallWithDefaultRule(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(nsxtExternalNetwork, NotNil)
	check.Assert(err, IsNil)

	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)
	check.Assert(vdc, NotNil)
	check.Assert(vdcGroup, NotNil)

	defer func() {
		// Cleanup
		err = vdcGroup.Delete()
		check.Assert(err, IsNil)
		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)
	}()

	fmt.Println("# Running Distributed Firewall tests for single Rule (with default DFW rule enabled)")
	test_NsxtDistributedFirewallRuleAboveDefault(vcd, check, vdcGroup.VdcGroup.Id, vcd.client, vdc)
}

func test_NsxtDistributedFirewallRule(vcd *TestVCD, check *C, vdcGroupId string, vcdClient *VCDClient, vdc *Vdc, deleteDefaultFirewallRule bool) {
	adminOrg, err := vcdClient.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

	vdcGroup, err := adminOrg.GetVdcGroupById(vdcGroupId)
	check.Assert(err, IsNil)

	_, err = vdcGroup.ActivateDfw()
	check.Assert(err, IsNil)

	// Prep firewall rule sample to operate with
	randomizedFwRuleDefs, ipSet, secGroup := createDistributedFirewallDefinitions(check, vcd, vdcGroup.VdcGroup.Id, vcdClient, vdc)
	// defer cleanup function in case something goes wrong
	defer func() {
		dfw, err := vdcGroup.GetDistributedFirewall()
		check.Assert(err, IsNil)
		err = dfw.DeleteAllRules()
		check.Assert(err, IsNil)
		_, err = vdcGroup.DisableDefaultPolicy()
		check.Assert(err, IsNil)
		err = ipSet.Delete()
		check.Assert(err, IsNil)
		err = secGroup.Delete()
		check.Assert(err, IsNil)
	}()

	randomizedFwRuleSubSet := randomizedFwRuleDefs[0:5] // taking only first 5 rules to limit time of testing

	// removing default firewall rule which is created by VCD when vdcGroup.ActivateDfw() is executed
	if deleteDefaultFirewallRule {
		err = vdcGroup.DeleteAllDistributedFirewallRules()
		check.Assert(err, IsNil)
	}

	// Adding firewal rules one by one and checking that each of them is placed correctly
	testDistributedFirewallRuleSequence(check, randomizedFwRuleSubSet, vdcGroup, false)
	testDistributedFirewallRuleSequence(check, randomizedFwRuleSubSet, vdcGroup, true)
}

func test_NsxtDistributedFirewallRuleAboveDefault(vcd *TestVCD, check *C, vdcGroupId string, vcdClient *VCDClient, vdc *Vdc) {
	adminOrg, err := vcdClient.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

	vdcGroup, err := adminOrg.GetVdcGroupById(vdcGroupId)
	check.Assert(err, IsNil)

	_, err = vdcGroup.ActivateDfw()
	check.Assert(err, IsNil)

	// Prep firewall rule sample to operate with
	randomizedFwRuleDefs, ipSet, secGroup := createDistributedFirewallDefinitions(check, vcd, vdcGroup.VdcGroup.Id, vcdClient, vdc)
	// defer cleanup function in case something goes wrong
	defer func() {
		dfw, err := vdcGroup.GetDistributedFirewall()
		check.Assert(err, IsNil)
		err = dfw.DeleteAllRules()
		check.Assert(err, IsNil)
		_, err = vdcGroup.DisableDefaultPolicy()
		check.Assert(err, IsNil)
		err = ipSet.Delete()
		check.Assert(err, IsNil)
		err = secGroup.Delete()
		check.Assert(err, IsNil)
	}()

	// Get default rule by name (this name is automatically set by VCD)
	defaultRuleName := fmt.Sprintf("Default_VdcGroup_%s", vdcGroup.VdcGroup.Name)
	defaultRule, err := vdcGroup.GetDistributedFirewallRuleByName(defaultRuleName)
	check.Assert(err, IsNil)
	check.Assert(defaultRule, NotNil)

	_, rule1, err := vdcGroup.CreateDistributedFirewallRule(defaultRule.Rule.ID, randomizedFwRuleDefs[0])
	check.Assert(err, IsNil)

	_, rule2, err := vdcGroup.CreateDistributedFirewallRule(defaultRule.Rule.ID, randomizedFwRuleDefs[1])
	check.Assert(err, IsNil)

	// The order should be
	// * rule1 (created first, inserted above default rule)
	// * rule2 (created after rule1, inserted above default rule)
	// * Default DFW rule

	allRules, err := vdcGroup.GetDistributedFirewall()
	check.Assert(err, IsNil)
	check.Assert(len(allRules.DistributedFirewallRuleContainer.Values), Equals, 3)
	check.Assert(allRules.DistributedFirewallRuleContainer.Values[0].ID, Equals, rule1.Rule.ID)
	check.Assert(allRules.DistributedFirewallRuleContainer.Values[1].ID, Equals, rule2.Rule.ID)
	check.Assert(allRules.DistributedFirewallRuleContainer.Values[2].ID, Equals, defaultRule.Rule.ID)

	// Clean up created firewall rules for next phase
	err = vdcGroup.DeleteAllDistributedFirewallRules()
	check.Assert(err, IsNil)

}

// testDistributedFirewallRuleSequence tests the following:
// * create firewall rules one one by one
// * check that the order of firewall rules is the same as requested (or exactly reverse if
// reverseOrder=true)
// * check that all IDs of created firewall rules persisted during further updates (means that no
// firewall rules were recreated during addition of new ones)
func testDistributedFirewallRuleSequence(check *C, randomizedFwRuleSubSet []*types.DistributedFirewallRule, vdcGroup *VdcGroup, reverseOrder bool) {
	createdIdsFound := make(map[string]bool)
	fmt.Printf("# Creating '%d' rules one by one (reverseOrder: %t)\n", len(randomizedFwRuleSubSet), reverseOrder)
	previousRuleId := ""
	for _, rule := range randomizedFwRuleSubSet {
		if testVerbose {
			fmt.Printf("%s\t%s\t%s\t%t\t%s\t%t\t%d\t%d\t%d\t%d\n", rule.Name, rule.Direction, rule.IpProtocol,
				rule.Enabled, rule.Action, rule.Logging, len(rule.SourceFirewallGroups), len(rule.DestinationFirewallGroups), len(rule.ApplicationPortProfiles), len(rule.NetworkContextProfiles))
		}

		completeDfw, singleCreatedFwRule, err := vdcGroup.CreateDistributedFirewallRule(previousRuleId, rule)
		check.Assert(err, IsNil)
		check.Assert(completeDfw, NotNil)
		check.Assert(singleCreatedFwRule, NotNil)
		createdIdsFound[singleCreatedFwRule.Rule.ID] = false

		// caching ID to use as previous rule in case
		if reverseOrder {
			previousRuleId = singleCreatedFwRule.Rule.ID
		}
	}
	fmt.Printf("# Done creating '%d' rules one by one (reverseOrder: %t)\n", len(randomizedFwRuleSubSet), reverseOrder)

	// Retrieve all firewall rules and check that order matches
	allRules, err := vdcGroup.GetDistributedFirewall()
	check.Assert(err, IsNil)
	check.Assert(len(allRules.DistributedFirewallRuleContainer.Values), Equals, len(randomizedFwRuleSubSet))

	// check that rule order is exactly as expected (either reverse of randomizedFwRuleSubSet or exactly the same based on reverseOrder parameter)
	if reverseOrder {
		for ruleIndex, rule := range allRules.DistributedFirewallRuleContainer.Values {
			reverseRuleIndex := len(randomizedFwRuleSubSet) - ruleIndex - 1
			check.Assert(rule.Name, Equals, randomizedFwRuleSubSet[reverseRuleIndex].Name)
			createdIdsFound[rule.ID] = true
		}
	} else {
		for ruleIndex, rule := range allRules.DistributedFirewallRuleContainer.Values {
			check.Assert(rule.Name, Equals, randomizedFwRuleSubSet[ruleIndex].Name)
			createdIdsFound[rule.ID] = true
		}
	}

	// Check that all created IDs are in the final output (none of the firewall rules were recreated)
	for _, value := range createdIdsFound {
		check.Assert(value, Equals, true)
	}

	// Perform Update
	ruleById, err := vdcGroup.GetDistributedFirewallRuleById(allRules.DistributedFirewallRuleContainer.Values[0].ID)
	check.Assert(err, IsNil)

	updatedRuleName := check.TestName() + "-updated"
	ruleById.Rule.Name = updatedRuleName
	updatedRule, err := ruleById.Update(ruleById.Rule)
	check.Assert(err, IsNil)
	check.Assert(updatedRule.Rule.Name, Equals, updatedRuleName)

	// Delete
	err = updatedRule.Delete()
	check.Assert(err, IsNil)

	notFoundById, err := vdcGroup.GetDistributedFirewallRuleById(updatedRule.Rule.ID)
	check.Assert(err, NotNil)
	check.Assert(notFoundById, IsNil)

	// Clean up created firewall rules for next phase
	err = vdcGroup.DeleteAllDistributedFirewallRules()
	check.Assert(err, IsNil)
}
