//go:build network || nsxt || functional || openapi || ALL
// +build network nsxt functional openapi ALL

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtEdgeCreate(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxvVdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	nsxtVdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	if ContainsNotFound(err) {
		check.Skip(fmt.Sprintf("No NSX-T VDC (%s) found - skipping test", vcd.config.VCD.Nsxt.Vdc))
	}
	check.Assert(err, IsNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(err, IsNil)

	egwDefinition := &types.OpenAPIEdgeGateway{
		Name:        "nsx-t-edge",
		Description: "nsx-t-edge-description",
		OrgVdc: &types.OpenApiReference{
			ID: nsxtVdc.Vdc.ID,
		},
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{types.EdgeGatewayUplinks{
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

	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)

	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)

	createdEdge.EdgeGateway.Name = "renamed-edge"
	updatedEdge, err := createdEdge.Update(createdEdge.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(updatedEdge.EdgeGateway.Name, Equals, "renamed-edge")

	// FIQL filtering test
	queryParams := url.Values{}
	queryParams.Add("filter", "name==renamed-edge")
	//
	egws, err := adminOrg.GetAllNsxtEdgeGateways(queryParams)
	check.Assert(err, IsNil)
	check.Assert(len(egws) >= 1, Equals, true)

	// Lookup using different available methods
	e1, err := adminOrg.GetNsxtEdgeGatewayByName(updatedEdge.EdgeGateway.Name)
	check.Assert(err, IsNil)
	e2, err := org.GetNsxtEdgeGatewayByName(updatedEdge.EdgeGateway.Name)
	check.Assert(err, IsNil)
	e3, err := nsxtVdc.GetNsxtEdgeGatewayByName(updatedEdge.EdgeGateway.Name)
	check.Assert(err, IsNil)
	e4, err := adminOrg.GetNsxtEdgeGatewayById(updatedEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	e5, err := org.GetNsxtEdgeGatewayById(updatedEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)
	e6, err := nsxtVdc.GetNsxtEdgeGatewayById(updatedEdge.EdgeGateway.ID)
	check.Assert(err, IsNil)

	// Try to search for NSX-T edge gateway in NSX-V VDC and expect it to be not found
	expectNil, err := nsxvVdc.GetNsxtEdgeGatewayByName(updatedEdge.EdgeGateway.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(expectNil, IsNil)
	expectNil, err = nsxvVdc.GetNsxtEdgeGatewayById(updatedEdge.EdgeGateway.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(expectNil, IsNil)

	// Ensure all methods found the same edge gateway
	check.Assert(e1.EdgeGateway.ID, Equals, e2.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e3.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e4.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e5.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e6.EdgeGateway.ID)

	err = updatedEdge.Delete()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_NsxtEdgeVdcGroup(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(err, IsNil)

	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)

	egwDefinition := &types.OpenAPIEdgeGateway{
		Name:        "nsx-t-edge",
		Description: "nsx-t-edge-description",
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

	// Create Edge Gateway in VDC
	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdc:.*`)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	PrependToCleanupListOpenApi(createdEdge.EdgeGateway.Name, check.TestName(), openApiEndpoint)

	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Equals, egwDefinition.OwnerRef.ID)

	// Move Edge Gateway to VDC Group
	movedGateway, err := createdEdge.MoveToVdc(vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Equals, vdcGroup.VdcGroup.Id)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdcGroup:.*`)

	// Get by name and owner ID
	edgeByNameAndOwnerId, err := org.GetNsxtEdgeGatewayByNameAndOwnerId(createdEdge.EdgeGateway.Name, vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)

	// Check lookup of Edge Gateways in VDC Groups
	edgeInVdcGroup, err := vdcGroup.GetNsxtEdgeGatewayByName(createdEdge.EdgeGateway.Name)
	check.Assert(err, IsNil)

	// Ensure both methods for retrieval get the same object
	check.Assert(edgeByNameAndOwnerId.EdgeGateway, DeepEquals, edgeInVdcGroup.EdgeGateway)

	// Masking known variables that have change for deep check
	edgeInVdcGroup.EdgeGateway.OwnerRef.Name = ""
	check.Assert(edgeInVdcGroup.EdgeGateway, DeepEquals, createdEdge.EdgeGateway)

	// Move Edge Gateway back to VDC from VDC Group
	egwDefinition.OwnerRef.ID = vdc.Vdc.ID
	egwDefinition.ID = movedGateway.EdgeGateway.ID

	movedBackToVdcEdge, err := movedGateway.Update(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(movedBackToVdcEdge.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdc:.*`)

	// Ignore known to differ fields, but check that whole Edge Gateway structure remains the same
	// as it is important to perform update operations without impacting configuration itself

	// Fields to ignore on both sides
	createdEdge.EdgeGateway.OrgVdc = movedBackToVdcEdge.EdgeGateway.OrgVdc
	createdEdge.EdgeGateway.OwnerRef = movedBackToVdcEdge.EdgeGateway.OwnerRef
	check.Assert(movedBackToVdcEdge.EdgeGateway, DeepEquals, createdEdge.EdgeGateway)

	// Remove Edge Gateway
	err = movedBackToVdcEdge.Delete()
	check.Assert(err, IsNil)
}
