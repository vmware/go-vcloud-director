//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtNatDnat(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	appPortProfiles, err := org.GetAllNsxtAppPortProfiles(nil, types.ApplicationPortProfileScopeSystem)
	check.Assert(err, IsNil)

	edgeGatewayPrimaryIp := ""
	if edge.EdgeGateway != nil && len(edge.EdgeGateway.EdgeGatewayUplinks) > 0 && len(edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values) > 0 {
		edgeGatewayPrimaryIp = edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP
	}
	check.Assert(edgeGatewayPrimaryIp, Not(Equals), "")

	natRuleDefinition := &types.NsxtNatRule{
		Name:              check.TestName() + "dnat",
		Description:       "description",
		Enabled:           true,
		RuleType:          types.NsxtNatRuleTypeDnat,
		ExternalAddresses: edgeGatewayPrimaryIp,
		InternalAddresses: "11.11.11.2",
		ApplicationPortProfile: &types.OpenApiReference{
			ID:   appPortProfiles[0].NsxtAppPortProfile.ID,
			Name: appPortProfiles[0].NsxtAppPortProfile.Name},
		SnatDestinationAddresses: "",
		Logging:                  vcd.client.Client.IsSysAdmin,
		DnatExternalPort:         "",
	}

	nsxtNatRuleChecks(natRuleDefinition, edge, check, vcd)
}

func (vcd *TestVCD) Test_NsxtNatDnatExternalPortPort(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	appPortProfiles, err := org.GetAllNsxtAppPortProfiles(nil, types.ApplicationPortProfileScopeSystem)
	check.Assert(err, IsNil)

	edgeGatewayPrimaryIp := ""
	if edge.EdgeGateway != nil && len(edge.EdgeGateway.EdgeGatewayUplinks) > 0 && len(edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values) > 0 {
		edgeGatewayPrimaryIp = edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP
	}
	check.Assert(edgeGatewayPrimaryIp, Not(Equals), "")

	natRuleDefinition := &types.NsxtNatRule{
		Name:              check.TestName() + "dnat",
		Description:       "description",
		Enabled:           true,
		RuleType:          types.NsxtNatRuleTypeDnat,
		ExternalAddresses: edgeGatewayPrimaryIp,
		InternalAddresses: "11.11.11.2",
		ApplicationPortProfile: &types.OpenApiReference{
			ID:   appPortProfiles[0].NsxtAppPortProfile.ID,
			Name: appPortProfiles[0].NsxtAppPortProfile.Name},
		SnatDestinationAddresses: "",
		Logging:                  vcd.client.Client.IsSysAdmin,
		DnatExternalPort:         "9898",
	}

	nsxtNatRuleChecks(natRuleDefinition, edge, check, vcd)
}

func (vcd *TestVCD) Test_NsxtNatDnatFirewallMatchPriority(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	appPortProfiles, err := org.GetAllNsxtAppPortProfiles(nil, types.ApplicationPortProfileScopeSystem)
	check.Assert(err, IsNil)

	edgeGatewayPrimaryIp := ""
	if edge.EdgeGateway != nil && len(edge.EdgeGateway.EdgeGatewayUplinks) > 0 && len(edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values) > 0 {
		edgeGatewayPrimaryIp = edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP
	}
	check.Assert(edgeGatewayPrimaryIp, Not(Equals), "")

	natRuleDefinition := &types.NsxtNatRule{
		Name:              check.TestName() + "dnat",
		Description:       "description",
		Enabled:           true,
		RuleType:          types.NsxtNatRuleTypeDnat,
		ExternalAddresses: edgeGatewayPrimaryIp,
		InternalAddresses: "11.11.11.2",
		ApplicationPortProfile: &types.OpenApiReference{
			ID:   appPortProfiles[0].NsxtAppPortProfile.ID,
			Name: appPortProfiles[0].NsxtAppPortProfile.Name},
		SnatDestinationAddresses: "",
		Logging:                  vcd.client.Client.IsSysAdmin,
		FirewallMatch:            types.NsxtNatRuleFirewallMatchExternalAddress,
		Priority:                 addrOf(248),
	}

	nsxtNatRuleChecks(natRuleDefinition, edge, check, vcd)
}

func (vcd *TestVCD) Test_NsxtNatNoDnat(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	edgeGatewayPrimaryIp := ""
	if edge.EdgeGateway != nil && len(edge.EdgeGateway.EdgeGatewayUplinks) > 0 && len(edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values) > 0 {
		edgeGatewayPrimaryIp = edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP
	}
	check.Assert(edgeGatewayPrimaryIp, Not(Equals), "")

	natRuleDefinition := &types.NsxtNatRule{
		Name:              check.TestName() + "no-dnat",
		Description:       "description",
		Enabled:           true,
		RuleType:          types.NsxtNatRuleTypeNoDnat,
		ExternalAddresses: edgeGatewayPrimaryIp,
	}

	nsxtNatRuleChecks(natRuleDefinition, edge, check, vcd)
}

func (vcd *TestVCD) Test_NsxtNatSnat(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	appPortProfiles, err := org.GetAllNsxtAppPortProfiles(nil, types.ApplicationPortProfileScopeSystem)
	check.Assert(err, IsNil)

	edgeGatewayPrimaryIp := ""
	if edge.EdgeGateway != nil && len(edge.EdgeGateway.EdgeGatewayUplinks) > 0 && len(edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values) > 0 {
		edgeGatewayPrimaryIp = edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP
	}
	check.Assert(edgeGatewayPrimaryIp, Not(Equals), "")

	natRuleDefinition := &types.NsxtNatRule{
		Name:                     check.TestName() + "snat",
		Description:              "description",
		Enabled:                  true,
		RuleType:                 types.NsxtNatRuleTypeSnat,
		ExternalAddresses:        edgeGatewayPrimaryIp,
		InternalAddresses:        "11.11.11.2",
		SnatDestinationAddresses: "11.11.11.4",
		ApplicationPortProfile: &types.OpenApiReference{
			ID:   appPortProfiles[1].NsxtAppPortProfile.ID,
			Name: appPortProfiles[1].NsxtAppPortProfile.Name},
	}

	nsxtNatRuleChecks(natRuleDefinition, edge, check, vcd)
}

func (vcd *TestVCD) Test_NsxtNatNoSnat(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	natRuleDefinition := &types.NsxtNatRule{
		Name:              check.TestName() + "no-snat",
		Description:       "description",
		Enabled:           true,
		RuleType:          types.NsxtNatRuleTypeNoSnat,
		ExternalAddresses: "",
		InternalAddresses: "11.11.11.2",
	}

	nsxtNatRuleChecks(natRuleDefinition, edge, check, vcd)
}

func (vcd *TestVCD) Test_NsxtNatPriorityAndFirewallMatch(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	edgeGatewayPrimaryIp := ""
	if edge.EdgeGateway != nil && len(edge.EdgeGateway.EdgeGatewayUplinks) > 0 && len(edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values) > 0 {
		edgeGatewayPrimaryIp = edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP
	}
	check.Assert(edgeGatewayPrimaryIp, Not(Equals), "")

	natRuleDefinition := &types.NsxtNatRule{
		Name:                     check.TestName() + "dnat",
		Description:              "description",
		Enabled:                  true,
		RuleType:                 types.NsxtNatRuleTypeDnat,
		ExternalAddresses:        edgeGatewayPrimaryIp,
		InternalAddresses:        "11.11.11.2",
		SnatDestinationAddresses: "",
		Logging:                  vcd.client.Client.IsSysAdmin,
		DnatExternalPort:         "",
		Priority:                 addrOf(100),
		FirewallMatch:            types.NsxtNatRuleFirewallMatchExternalAddress,
	}

	nsxtNatRuleChecks(natRuleDefinition, edge, check, vcd)
}

// Test_NsxtNatReflexive tests out REFLEXIVE rule type. This is only available in VCD 10.3 (API V36.0)
func (vcd *TestVCD) Test_NsxtNatReflexive(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	edgeGatewayPrimaryIp := ""
	if edge.EdgeGateway != nil && len(edge.EdgeGateway.EdgeGatewayUplinks) > 0 && len(edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values) > 0 {
		edgeGatewayPrimaryIp = edge.EdgeGateway.EdgeGatewayUplinks[0].Subnets.Values[0].PrimaryIP
	}
	check.Assert(edgeGatewayPrimaryIp, Not(Equals), "")

	natRuleDefinition := &types.NsxtNatRule{
		Name:        check.TestName() + "reflexive",
		Description: "description",
		Enabled:     true,
		//RuleType:          types.NsxtNatRuleTypeReflexive,
		Type:              types.NsxtNatRuleTypeReflexive,
		ExternalAddresses: edgeGatewayPrimaryIp,
		InternalAddresses: "11.11.11.2",
		Priority:          addrOf(100),
		FirewallMatch:     types.NsxtNatRuleFirewallMatchExternalAddress,
	}

	nsxtNatRuleChecks(natRuleDefinition, edge, check, vcd)
}

func nsxtNatRuleChecks(natRuleDefinition *types.NsxtNatRule, edge *NsxtEdgeGateway, check *C, vcd *TestVCD) {
	createdNatRule, err := edge.CreateNatRule(natRuleDefinition)
	check.Assert(err, IsNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + fmt.Sprintf(types.OpenApiEndpointNsxtNatRules, edge.EdgeGateway.ID) + createdNatRule.NsxtNatRule.ID
	AddToCleanupListOpenApi(createdNatRule.NsxtNatRule.Name, check.TestName(), openApiEndpoint)

	// check if created rule matches definition
	check.Assert(createdNatRule.IsEqualTo(natRuleDefinition), Equals, true)

	// Validate that supplied values are the same as read values
	natRuleDefinition.ID = createdNatRule.NsxtNatRule.ID                       // ID is always the difference
	natRuleDefinition.Priority = createdNatRule.NsxtNatRule.Priority           // Priority returns default value (0) for VCD 10.2.2+
	natRuleDefinition.FirewallMatch = createdNatRule.NsxtNatRule.FirewallMatch // FirewallMatch returns default value (MATCH_INTERNAL_ADDRESS) for VCD 10.2.2+
	natRuleDefinition.Version = createdNatRule.NsxtNatRule.Version             // Version will always be populated afterwards
	natRuleDefinition.Type = createdNatRule.NsxtNatRule.Type

	check.Assert(createdNatRule.NsxtNatRule, DeepEquals, natRuleDefinition)

	// Try to get NAT rules by name and by ID
	natRuleById, err := edge.GetNatRuleById(createdNatRule.NsxtNatRule.ID)
	check.Assert(err, IsNil)
	natRuleByName, err := edge.GetNatRuleByName(createdNatRule.NsxtNatRule.Name)
	check.Assert(err, IsNil)

	check.Assert(natRuleById.NsxtNatRule, DeepEquals, natRuleDefinition)
	check.Assert(natRuleByName.NsxtNatRule, DeepEquals, natRuleDefinition)

	// Try to update value
	createdNatRule.NsxtNatRule.Name = check.TestName() + "updated"
	updatedNatRule, err := createdNatRule.Update(createdNatRule.NsxtNatRule)
	check.Assert(err, IsNil)

	// validate that supplied values are new, but ID stays the same
	check.Assert(updatedNatRule.NsxtNatRule.ID, Equals, createdNatRule.NsxtNatRule.ID)
	check.Assert(updatedNatRule.NsxtNatRule.RuleType, Equals, createdNatRule.NsxtNatRule.RuleType)

	err = createdNatRule.Delete()
	check.Assert(err, IsNil)

	_, err = edge.GetNatRuleById(createdNatRule.NsxtNatRule.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
}
