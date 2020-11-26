// +build network nsxt functional openapi ALL

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtEdgeCreate(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 34") {
		check.Skip("At least VCD 10.1 is required to create NSX-T edge gateway")
	}

	skipNoNsxtConfiguration(vcd, check)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)

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

	// Ensure all methods found the same edge gateway
	check.Assert(e1.EdgeGateway.ID, Equals, e2.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e3.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e4.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e5.EdgeGateway.ID)
	check.Assert(e1.EdgeGateway.ID, Equals, e6.EdgeGateway.ID)

	err = updatedEdge.Delete()
	check.Assert(err, IsNil)
}
