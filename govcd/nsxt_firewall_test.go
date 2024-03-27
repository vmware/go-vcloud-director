//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxtFirewall creates 20 firewall rules with randomized parameters
func (vcd *TestVCD) Test_NsxtFirewall(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Get existing firewall rule configuration
	fwRules, err := edge.GetNsxtFirewall()
	check.Assert(err, IsNil)

	existingDefaultRuleCount := len(fwRules.NsxtFirewallRuleContainer.DefaultRules)
	existingSystemRuleCount := len(fwRules.NsxtFirewallRuleContainer.SystemRules)

	// Create some prerequisites and generate firewall rule configurations to feed them into config
	randomizedFwRuleDefs := createFirewallDefinitions(check, vcd)
	fwRules.NsxtFirewallRuleContainer.UserDefinedRules = randomizedFwRuleDefs

	if testVerbose {
		dumpFirewallRulesToScreen(randomizedFwRuleDefs)
	}

	fwCreated, err := edge.UpdateNsxtFirewall(fwRules.NsxtFirewallRuleContainer)
	check.Assert(err, IsNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + fmt.Sprintf(types.OpenApiEndpointNsxtFirewallRules, edge.EdgeGateway.ID)
	PrependToCleanupList(openApiEndpoint, "OpenApiEntityFirewall", edge.EdgeGateway.Name, check.TestName())

	check.Assert(fwCreated, Not(IsNil))
	check.Assert(len(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules), Equals, len(randomizedFwRuleDefs))

	// Check that all created rules are have the same attributes and order
	for index := range fwCreated.NsxtFirewallRuleContainer.UserDefinedRules {
		check.Assert(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].Name, Equals, randomizedFwRuleDefs[index].Name)
		check.Assert(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].Direction, Equals, randomizedFwRuleDefs[index].Direction)
		check.Assert(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].IpProtocol, Equals, randomizedFwRuleDefs[index].IpProtocol)
		check.Assert(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].Enabled, Equals, randomizedFwRuleDefs[index].Enabled)
		check.Assert(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].ActionValue, Equals, randomizedFwRuleDefs[index].ActionValue)
		if vcd.client.Client.IsSysAdmin {
			// Only system administrator can handle logging
			check.Assert(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].Logging, Equals, randomizedFwRuleDefs[index].Logging)
		}

		for fwGroupIndex := range fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].SourceFirewallGroups {
			check.Assert(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].SourceFirewallGroups[fwGroupIndex].ID, Equals, randomizedFwRuleDefs[index].SourceFirewallGroups[fwGroupIndex].ID)
		}

		for fwGroupIndex := range fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].DestinationFirewallGroups {
			check.Assert(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].DestinationFirewallGroups[fwGroupIndex].ID, Equals, randomizedFwRuleDefs[index].DestinationFirewallGroups[fwGroupIndex].ID)
		}

		// Ensure the same amount of Application Port Profiles are assigned and created
		check.Assert(len(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules), Equals, len(randomizedFwRuleDefs))
		definedAppPortProfileIds := extractIdsFromOpenApiReferences(randomizedFwRuleDefs[index].ApplicationPortProfiles)
		for _, appPortProfile := range fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[index].ApplicationPortProfiles {
			check.Assert(contains(appPortProfile.ID, definedAppPortProfileIds), Equals, true)
		}
	}

	// Delete a single rule by ID and check for two things:
	// * Rule with deleted ID should not be found in list post deletion
	// * There should be one less rule in the list
	deleteRuleId := fwCreated.NsxtFirewallRuleContainer.UserDefinedRules[3].ID
	err = fwCreated.DeleteRuleById(deleteRuleId)
	check.Assert(err, IsNil)

	allRulesPostDeletion, err := edge.GetNsxtFirewall()
	check.Assert(err, IsNil)

	check.Assert(len(allRulesPostDeletion.NsxtFirewallRuleContainer.UserDefinedRules), Equals, len(fwCreated.NsxtFirewallRuleContainer.UserDefinedRules)-1)
	for _, rule := range allRulesPostDeletion.NsxtFirewallRuleContainer.UserDefinedRules {
		check.Assert(rule.ID, Not(Equals), deleteRuleId)
	}

	err = fwRules.DeleteAllRules()
	check.Assert(err, IsNil)

	// Ensure no firewall rules left in user space post deletion, but the same amount of default and system rules still exist
	postDeleteCheck, err := edge.GetNsxtFirewall()
	check.Assert(err, IsNil)
	check.Assert(len(postDeleteCheck.NsxtFirewallRuleContainer.UserDefinedRules), Equals, 0)
	check.Assert(len(postDeleteCheck.NsxtFirewallRuleContainer.DefaultRules), Equals, existingDefaultRuleCount)
	check.Assert(len(postDeleteCheck.NsxtFirewallRuleContainer.SystemRules), Equals, existingSystemRuleCount)

}

// createFirewallDefinitions creates some randomized firewall rule configurations to match possible configurations
func createFirewallDefinitions(check *C, vcd *TestVCD) []*types.NsxtFirewallRule {
	// This number does not impact performance because all rules are created at once in the API
	numberOfRules := 20

	// Pre-Create Firewall Groups (IP Set and Security Group to randomly configure them)
	ipSet := preCreateIpSet(check, vcd)
	secGroup := preCreateSecurityGroup(check, vcd)
	fwGroupIds := []string{ipSet.NsxtFirewallGroup.ID, secGroup.NsxtFirewallGroup.ID}
	fwGroupRefs := convertSliceOfStringsToOpenApiReferenceIds(fwGroupIds)
	appPortProfileReferences := getRandomListOfAppPortProfiles(check, vcd)

	firewallRules := make([]*types.NsxtFirewallRule, numberOfRules)
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

		firewallRules[a] = &types.NsxtFirewallRule{
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
	}

	return firewallRules
}

func pickRandomString(in []string) string {
	randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(in))))
	return in[randomIndex.Uint64()]
}

// pickRandomOpenApiRefOrEmpty picks a random OpenAPI entity or an empty one
func pickRandomOpenApiRefOrEmpty(in []types.OpenApiReference) types.OpenApiReference {
	// Random value can be up to len+1 (len+1 is the special case when it should return an empty reference)
	randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(in)+1)))
	if randomIndex.Uint64() == uint64(len(in)) {
		return types.OpenApiReference{}
	}
	return in[randomIndex.Uint64()]
}

func preCreateIpSet(check *C, vcd *TestVCD) *NsxtFirewallGroup {
	nsxtVdc := vcd.nsxtVdc
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	ipSetDefinition := &types.NsxtFirewallGroup{
		Name:           check.TestName() + "ipset",
		Description:    check.TestName() + "-Description",
		Type:           types.FirewallGroupTypeIpSet,
		EdgeGatewayRef: &types.OpenApiReference{ID: edge.EdgeGateway.ID},

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
	AddToCleanupListOpenApi(createdIpSet.NsxtFirewallGroup.Name, check.TestName(), openApiEndpoint)

	return createdIpSet
}

func preCreateSecurityGroup(check *C, vcd *TestVCD) *NsxtFirewallGroup {
	nsxtVdc := vcd.nsxtVdc
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	fwGroupDefinition := &types.NsxtFirewallGroup{
		Name:           check.TestName() + "security-group",
		Description:    check.TestName() + "-Description",
		Type:           types.FirewallGroupTypeSecurityGroup,
		EdgeGatewayRef: &types.OpenApiReference{ID: edge.EdgeGateway.ID},
	}

	// Create firewall group and add to cleanup if it was created
	createdSecGroup, err := nsxtVdc.CreateNsxtFirewallGroup(fwGroupDefinition)
	check.Assert(err, IsNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups + createdSecGroup.NsxtFirewallGroup.ID
	AddToCleanupListOpenApi(check.TestName()+"sec-group", check.TestName(), openApiEndpoint)

	return createdSecGroup
}

func getRandomListOfAppPortProfiles(check *C, vcd *TestVCD) []types.OpenApiReference {
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	appProfileSlice, err := org.GetAllNsxtAppPortProfiles(nil, types.ApplicationPortProfileScopeSystem)
	check.Assert(err, IsNil)

	openApiRefs := make([]types.OpenApiReference, len(appProfileSlice))
	for index, appPortProfile := range appProfileSlice {
		openApiRefs[index].ID = appPortProfile.NsxtAppPortProfile.ID
		openApiRefs[index].Name = appPortProfile.NsxtAppPortProfile.Name
	}

	return openApiRefs
}

func dumpFirewallRulesToScreen(rules []*types.NsxtFirewallRule) {
	fmt.Println("# The following firewall rules will be created")
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "Name\tDirection\tIP Protocol\tEnabled\tAction\tLogging\tSrc Count\tDst Count\tAppPortProfile Count")

	for _, rule := range rules {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\t%t\t%d\t%d\t%d\n", rule.Name, rule.Direction, rule.IpProtocol,
			rule.Enabled, rule.ActionValue, rule.Logging, len(rule.SourceFirewallGroups), len(rule.DestinationFirewallGroups), len(rule.ApplicationPortProfiles))
	}
	err := w.Flush()
	if err != nil {
		util.Logger.Printf("Error while dumping Firewall rules to screen: %s", err)
	}
}
