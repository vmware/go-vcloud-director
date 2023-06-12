//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_IpSpaceIntegration tests out IP Space integration with other components
// It creates and integrates the following entities:
// * IP Space
// * External Network (Provider gateway backed by T0 router)
// * Configures IP Space Uplink for External Network
// * NSX-T Edge Gateway backed by the newly created External Network
// * IP Allocation - Manual and Automatic
// * Uses the IP address in DNAT rule
func (vcd *TestVCD) Test_IpSpaceIntegration(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointIpSpaces)

	ipSpace := createIpSpace(vcd, check)
	extNet := createExternalNetwork(vcd, check)

	// IP Space uplink (not directly needed anywhere)
	_ = createIpSpaceUplink(vcd, check, extNet.ExternalNetwork.ID, ipSpace.IpSpace.ID)

	// Create NSX-T Edge Gateway
	_ = createNsxtEdgeGateway(vcd, check, extNet.ExternalNetwork.ID)

	// IP Allocation request
	ipSpaceAllocationRequest := &types.IpSpaceIpAllocationRequest{
		Type:     "FLOATING_IP",
		Quantity: addrOf(1),
	}
	// resulting slice must have 1 IP as the requested quantity is 1
	ipAllocationResult, err := vcd.org.IpSpaceAllocateIp(ipSpace.IpSpace.ID, ipSpaceAllocationRequest)
	check.Assert(err, IsNil)
	check.Assert(len(ipAllocationResult), Equals, 1)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + fmt.Sprintf(types.OpenApiEndpointIpSpaceUplinksAllocations, ipSpace.IpSpace.ID) + ipAllocationResult[0].ID
	PrependToCleanupListOpenApi("NSX-T IP Space IP Allocation", check.TestName(), openApiEndpoint)

	// Get IP Allocation
	ipAllocation, err := vcd.org.GetIpSpaceAllocationByTypeAndValue(ipSpace.IpSpace.ID, "FLOATING_IP", ipAllocationResult[0].Value, nil)
	check.Assert(err, IsNil)
	check.Assert(ipAllocation, NotNil)
	check.Assert(ipAllocation.IpSpaceIpAllocation.UsageState, Equals, types.IpSpaceIpAllocationUnused)

	// Set the IP for manual usage

	ipAllocation.IpSpaceIpAllocation.UsageState = types.IpSpaceIpAllocationUsedManual
	ipAllocation.IpSpaceIpAllocation.Description = "Manual usage description"

	updatedIpAllocation, err := ipAllocation.Update(ipAllocation.IpSpaceIpAllocation)
	check.Assert(err, IsNil)
	check.Assert(updatedIpAllocation.IpSpaceIpAllocation.ID, Equals, ipAllocation.IpSpaceIpAllocation.ID)
	check.Assert(updatedIpAllocation.IpSpaceIpAllocation.UsageState, Equals, types.IpSpaceIpAllocationUsedManual)

}

func createIpSpaceUplink(vcd *TestVCD, check *C, extNetId, ipSpaceId string) *IpSpaceUplink {
	// Create Uplink configuration
	uplinkConfig := &types.IpSpaceUplink{
		Name:               check.TestName(),
		Description:        "IP SPace Uplink for External Network (Provider Gateway)",
		ExternalNetworkRef: &types.OpenApiReference{ID: extNetId},
		IPSpaceRef:         &types.OpenApiReference{ID: ipSpaceId},
	}

	createdIpSpaceUplink, err := vcd.client.CreateIpSpaceUplink(uplinkConfig)
	check.Assert(err, IsNil)
	check.Assert(createdIpSpaceUplink, NotNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks + createdIpSpaceUplink.IpSpaceUplink.ID
	AddToCleanupListOpenApi(createdIpSpaceUplink.IpSpaceUplink.Name, check.TestName(), openApiEndpoint)

	return createdIpSpaceUplink
}

func createNsxtEdgeGateway(vcd *TestVCD, check *C, extNetId string) *NsxtEdgeGateway {
	egwDefinition := &types.OpenAPIEdgeGateway{
		Name:        check.TestName(),
		Description: "nsx-t-edge-description",
		OrgVdc: &types.OpenApiReference{
			ID: vcd.nsxtVdc.Vdc.ID,
		},
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{
			types.EdgeGatewayUplinks{
				UplinkID: extNetId,
			},
		},
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	createdEdge, err := adminOrg.CreateNsxtEdgeGateway(egwDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdEdge.EdgeGateway.Name, Equals, egwDefinition.Name)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	// Using Prepend function so that Edge Gateway is removed before parent External Network is being removed
	PrependToCleanupListOpenApi("NSX-T Edge Gateway", check.TestName(), openApiEndpoint)

	return createdEdge
}
