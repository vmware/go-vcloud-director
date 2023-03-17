//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxtIpSet tests out IP Set capabilities using Firewall Group endpoint
func (vcd *TestVCD) Test_NsxtIpSet(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointFirewallGroups)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	ipSetDefinition := &types.NsxtFirewallGroup{
		Name:        check.TestName(),
		Description: check.TestName() + "-Description",
		Type:        types.FirewallGroupTypeIpSet,
		OwnerRef:    &types.OpenApiReference{ID: edge.EdgeGateway.ID},

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

	check.Assert(createdIpSet.NsxtFirewallGroup.ID, Not(Equals), "")
	check.Assert(createdIpSet.NsxtFirewallGroup.EdgeGatewayRef.Name, Equals, vcd.config.VCD.Nsxt.EdgeGateway)

	check.Assert(createdIpSet.NsxtFirewallGroup.Description, Equals, ipSetDefinition.Description)
	check.Assert(createdIpSet.NsxtFirewallGroup.Name, Equals, ipSetDefinition.Name)
	check.Assert(createdIpSet.NsxtFirewallGroup.Type, Equals, ipSetDefinition.Type)

	// Update and compare
	createdIpSet.NsxtFirewallGroup.Description = "updated-description"
	createdIpSet.NsxtFirewallGroup.Name = check.TestName() + "-updated"

	updatedIpSet, err := createdIpSet.Update(createdIpSet.NsxtFirewallGroup)
	check.Assert(err, IsNil)
	check.Assert(updatedIpSet.NsxtFirewallGroup, DeepEquals, createdIpSet.NsxtFirewallGroup)

	check.Assert(updatedIpSet, DeepEquals, createdIpSet)

	// Get all Firewall Groups and check if the created one is there
	allIpSets, err := org.GetAllNsxtFirewallGroups(nil, types.FirewallGroupTypeIpSet)
	check.Assert(err, IsNil)
	fwGroupFound := false
	for i := range allIpSets {
		if allIpSets[i].NsxtFirewallGroup.ID == updatedIpSet.NsxtFirewallGroup.ID {
			fwGroupFound = true
			break
		}
	}
	check.Assert(fwGroupFound, Equals, true)

	// Check if all retrieval functions get the same
	orgIpSetByName, err := org.GetNsxtFirewallGroupByName(updatedIpSet.NsxtFirewallGroup.Name, types.FirewallGroupTypeIpSet)
	check.Assert(err, IsNil)
	orgIpSetById, err := org.GetNsxtFirewallGroupById(updatedIpSet.NsxtFirewallGroup.ID)
	check.Assert(err, IsNil)
	check.Assert(orgIpSetByName.NsxtFirewallGroup, DeepEquals, orgIpSetById.NsxtFirewallGroup)

	// Get Firewall Group using VDC
	vdcIpSetByName, err := nsxtVdc.GetNsxtFirewallGroupByName(updatedIpSet.NsxtFirewallGroup.Name, types.FirewallGroupTypeIpSet)
	check.Assert(err, IsNil)
	vdcIpSetById, err := nsxtVdc.GetNsxtFirewallGroupById(updatedIpSet.NsxtFirewallGroup.ID)
	check.Assert(err, IsNil)
	check.Assert(vdcIpSetByName.NsxtFirewallGroup, DeepEquals, vdcIpSetById.NsxtFirewallGroup)
	check.Assert(vdcIpSetById.NsxtFirewallGroup, DeepEquals, orgIpSetById.NsxtFirewallGroup)

	// Get Firewall Group using Edge Gateway
	edgeIpSetByName, err := edge.GetNsxtFirewallGroupByName(updatedIpSet.NsxtFirewallGroup.Name, types.FirewallGroupTypeIpSet)
	check.Assert(err, IsNil)
	edgeIpSetById, err := edge.GetNsxtFirewallGroupById(updatedIpSet.NsxtFirewallGroup.ID)
	check.Assert(err, IsNil)
	check.Assert(edgeIpSetByName.NsxtFirewallGroup, DeepEquals, orgIpSetByName.NsxtFirewallGroup)
	check.Assert(edgeIpSetById.NsxtFirewallGroup, DeepEquals, edgeIpSetByName.NsxtFirewallGroup)

	// Get Firewall Group using VDC Group
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(nsxtExternalNetwork, NotNil)

	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)
	egwDefinition := &types.OpenAPIEdgeGateway{
		Name:        "nsx-for-IpSet-edge",
		Description: "nsx-for-IpSet-edge-description",
		OwnerRef: &types.OpenApiReference{
			ID: vdc.Vdc.ID,
		},
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{{
			UplinkID: nsxtExternalNetwork.ExternalNetwork.ID,
			Subnets: types.OpenAPIEdgeGatewaySubnets{Values: []types.OpenAPIEdgeGatewaySubnetValue{{
				Gateway:      "1.1.1.1",
				PrefixLength: 24,
				Enabled:      true,
			}}},
			Connected: true,
			Dedicated: false,
		}},
	}

	// Create Edge Gateway in VDC Group
	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdc:.*`)
	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	PrependToCleanupListOpenApi(createdEdge.EdgeGateway.Name, check.TestName(), openApiEndpoint)

	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Equals, egwDefinition.OwnerRef.ID)

	movedGateway, err := createdEdge.MoveToVdcOrVdcGroup(vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Equals, vdcGroup.VdcGroup.Id)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdcGroup:.*`)

	ipSetDefinition.Name = check.TestName() + "VdcGroup"
	ipSetDefinition.OwnerRef.ID = vdcGroup.VdcGroup.Id
	createdIpSetInVdcGroup, err := createdEdge.CreateNsxtFirewallGroup(ipSetDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdIpSetInVdcGroup, NotNil)
	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups + createdIpSetInVdcGroup.NsxtFirewallGroup.ID
	AddToCleanupListOpenApi(createdIpSet.NsxtFirewallGroup.Name, check.TestName(), openApiEndpoint)
	vdcGroupIpSetByName, err := vdcGroup.GetNsxtFirewallGroupByName(createdIpSetInVdcGroup.NsxtFirewallGroup.Name, types.FirewallGroupTypeIpSet)
	check.Assert(err, IsNil)
	vdcGroupIpSetById, err := vdcGroup.GetNsxtFirewallGroupById(createdIpSetInVdcGroup.NsxtFirewallGroup.ID)
	check.Assert(err, IsNil)
	check.Assert(vdcGroupIpSetByName.NsxtFirewallGroup, DeepEquals, vdcGroupIpSetById.NsxtFirewallGroup)
	check.Assert(vdcGroupIpSetById.NsxtFirewallGroup, DeepEquals, vdcGroupIpSetByName.NsxtFirewallGroup)

	associatedVms, err := edgeIpSetByName.GetAssociatedVms()
	// IP_SET type Firewall Groups do not have VM associations and throw an error on API call.
	// The error is: only Security Groups have associated VMs. This Firewall Group has type 'IP_SET'
	// Not hardcodeing it here because it may change and break the test.
	check.Assert(err, NotNil)
	check.Assert(associatedVms, IsNil)

	// Remove
	err = createdIpSet.Delete()
	check.Assert(err, IsNil)
	err = vdcGroupIpSetByName.Delete()
	check.Assert(err, IsNil)

	// Create IP Set using Edge Gateway method
	ipSetDefinition.Name = check.TestName() + "-using-edge-gateway-type"
	ipSetDefinition.OwnerRef.ID = edge.EdgeGateway.ID

	// Create IP Set and add to cleanup if it was created
	edgeCreatedIpSet, err := nsxtVdc.CreateNsxtFirewallGroup(ipSetDefinition)
	check.Assert(err, IsNil)
	openApiEndpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups + edgeCreatedIpSet.NsxtFirewallGroup.ID
	AddToCleanupListOpenApi(createdIpSet.NsxtFirewallGroup.Name, check.TestName(), openApiEndpoint)

	check.Assert(edgeCreatedIpSet.NsxtFirewallGroup.ID, Not(Equals), "")
	check.Assert(edgeCreatedIpSet.NsxtFirewallGroup.OwnerRef.Name, Equals, edge.EdgeGateway.Name)

	err = edgeCreatedIpSet.Delete()
	check.Assert(err, IsNil)

	// Remove Edge Gateway
	err = movedGateway.Delete()
	check.Assert(err, IsNil)
}
