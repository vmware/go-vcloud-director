//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxtDynamicSecurityGroup tests out CRUD of Dynamic NSX-T Security Group
//
// Note. Dynamic Security Group is one type of Firewall Group. Other types are IP-Set and Static
// Security Group.
func (vcd *TestVCD) Test_NsxtDynamicSecurityGroup(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	vdcGroup, err := adminOrg.GetVdcGroupByName(vcd.config.VCD.Nsxt.VdcGroup)
	check.Assert(err, IsNil)

	dynamicSecGroupDefinition := &types.NsxtFirewallGroup{
		Name:        check.TestName(),
		Description: check.TestName() + "-Description",
		TypeValue:   types.FirewallGroupTypeVmCriteria,
		OwnerRef:    &types.OpenApiReference{ID: vdcGroup.VdcGroup.Id},
		VmCriteria: []types.NsxtFirewallGroupVmCriteria{
			{
				VmCriteriaRule: []types.NsxtFirewallGroupVmCriteriaRule{
					{
						AttributeType:  "VM_TAG",
						Operator:       "EQUALS",
						AttributeValue: "string",
					}, // Boolean AND
					{
						AttributeType:  "VM_TAG",
						Operator:       "CONTAINS",
						AttributeValue: "substring",
					}, // Boolean AND
					{
						AttributeType:  "VM_TAG",
						Operator:       "STARTS_WITH",
						AttributeValue: "substring",
					}, // Boolean AND
					{
						AttributeType:  "VM_TAG",
						Operator:       "ENDS_WITH",
						AttributeValue: "substring",
					}, // Boolean AND
				},
			}, // Boolean OR
			{
				VmCriteriaRule: []types.NsxtFirewallGroupVmCriteriaRule{
					{
						AttributeType:  "VM_NAME",
						Operator:       "CONTAINS",
						AttributeValue: "substring",
					}, // Boolean AND
					{
						AttributeType:  "VM_NAME",
						Operator:       "STARTS_WITH",
						AttributeValue: "substring",
					}, // Boolean AND
				},
			},
		},
	}

	createdDynamicGroup, err := vdcGroup.CreateNsxtFirewallGroup(dynamicSecGroupDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdDynamicGroup, NotNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups + createdDynamicGroup.NsxtFirewallGroup.ID
	AddToCleanupListOpenApi(createdDynamicGroup.NsxtFirewallGroup.Name, check.TestName(), openApiEndpoint)

	check.Assert(createdDynamicGroup.NsxtFirewallGroup.ID, Not(Equals), "")
	check.Assert(createdDynamicGroup.NsxtFirewallGroup.OwnerRef.Name, Equals, vcd.config.VCD.Nsxt.VdcGroup)
	check.Assert(createdDynamicGroup.NsxtFirewallGroup.TypeValue, Equals, types.FirewallGroupTypeVmCriteria)

	// Update
	createdDynamicGroup.NsxtFirewallGroup.Description = "updated-description"
	createdDynamicGroup.NsxtFirewallGroup.Name = check.TestName() + "-updated"

	updatedDynamicGroup, err := createdDynamicGroup.Update(createdDynamicGroup.NsxtFirewallGroup)
	check.Assert(err, IsNil)
	check.Assert(updatedDynamicGroup, NotNil)
	check.Assert(updatedDynamicGroup.NsxtFirewallGroup, DeepEquals, createdDynamicGroup.NsxtFirewallGroup)

	check.Assert(updatedDynamicGroup, DeepEquals, createdDynamicGroup)

	// Remove Dynamic Security Group
	err = updatedDynamicGroup.Delete()
	check.Assert(err, IsNil)
}
