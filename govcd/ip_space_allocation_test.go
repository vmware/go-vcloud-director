//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_IpSpaceIpAllocation tests out IP Space integration with other components
func (vcd *TestVCD) Test_IpSpaceIpAllocation(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointIpSpaceIpAllocations)

	ipSpace := createIpSpace(vcd, check)
	extNet := createExternalNetwork(vcd, check)

	// IP Space uplink (not directly referenced anywhere, but is required to make IP allocations)
	ipSpaceUplink := createIpSpaceUplink(vcd, check, extNet.ExternalNetwork.ID, ipSpace.IpSpace.ID)

	// Create NSX-T Edge Gateway
	edgeGw := createNsxtEdgeGateway(vcd, check, extNet.ExternalNetwork.ID)

	// Floating IP Allocation request
	floatingIpAllocationRequest := &types.IpSpaceIpAllocationRequest{
		Type:     "FLOATING_IP",
		Quantity: addrOf(1),
	}
	performIpAllocationChecks(vcd, check, ipSpace.IpSpace.ID, edgeGw.EdgeGateway.ID, floatingIpAllocationRequest)

	// Prefix allocation request
	prefixAllocationRequest := &types.IpSpaceIpAllocationRequest{
		Type:         "IP_PREFIX",
		Quantity:     addrOf(1),
		PrefixLength: addrOf(31),
	}
	performIpAllocationChecks(vcd, check, ipSpace.IpSpace.ID, edgeGw.EdgeGateway.ID, prefixAllocationRequest)

	// Cleanup
	err := edgeGw.Delete()
	check.Assert(err, IsNil)

	err = ipSpaceUplink.Delete()
	check.Assert(err, IsNil)

	err = extNet.Delete()
	check.Assert(err, IsNil)

	err = ipSpace.Delete()
	check.Assert(err, IsNil)
}

func performIpAllocationChecks(vcd *TestVCD, check *C, ipSpaceId, edgeGatewayId string, ipSpaceAllocationRequest *types.IpSpaceIpAllocationRequest) {
	// resulting slice must have 1 IP as the requested quantity is 1
	ipAllocationResult, err := vcd.org.IpSpaceAllocateIp(ipSpaceId, ipSpaceAllocationRequest)
	check.Assert(err, IsNil)
	check.Assert(len(ipAllocationResult), Equals, 1)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + fmt.Sprintf(types.OpenApiEndpointIpSpaceIpAllocations, ipSpaceId) + ipAllocationResult[0].ID
	PrependToCleanupListOpenApi("NSX-T IP Space IP Allocation", check.TestName(), openApiEndpoint)

	// Check IP allocation suggestion endpoint
	performIpSuggestionChecks(vcd, check, ipSpaceId, edgeGatewayId, ipSpaceAllocationRequest)

	// Get IP Allocation
	ipAllocation, err := vcd.org.GetIpSpaceAllocationByTypeAndValue(ipSpaceId, ipSpaceAllocationRequest.Type, ipAllocationResult[0].Value, nil)
	check.Assert(err, IsNil)
	check.Assert(ipAllocation, NotNil)
	check.Assert(ipAllocation.IpSpaceIpAllocation.UsageState, Equals, types.IpSpaceIpAllocationUnused)

	// Get IP Allocation by ID
	ipAllocationById, err := vcd.org.GetIpSpaceAllocationById(ipSpaceId, ipAllocation.IpSpaceIpAllocation.ID)
	check.Assert(err, IsNil)
	check.Assert(ipAllocationById, NotNil)
	check.Assert(ipAllocationById.IpSpaceIpAllocation, DeepEquals, ipAllocation.IpSpaceIpAllocation)

	// Set the IP for manual usage
	ipAllocation.IpSpaceIpAllocation.UsageState = types.IpSpaceIpAllocationUsedManual
	ipAllocation.IpSpaceIpAllocation.Description = "Manual usage description"
	updatedIpAllocationManual, err := ipAllocation.Update(ipAllocation.IpSpaceIpAllocation)
	check.Assert(err, IsNil)
	check.Assert(updatedIpAllocationManual.IpSpaceIpAllocation.ID, Equals, ipAllocation.IpSpaceIpAllocation.ID)
	check.Assert(updatedIpAllocationManual.IpSpaceIpAllocation.UsageState, Equals, types.IpSpaceIpAllocationUsedManual)

	// Removal manual allocation
	ipAllocation.IpSpaceIpAllocation.UsageState = types.IpSpaceIpAllocationUnused
	ipAllocation.IpSpaceIpAllocation.Description = ""
	releasedIpAllocation, err := updatedIpAllocationManual.Update(ipAllocation.IpSpaceIpAllocation)
	check.Assert(err, IsNil)
	check.Assert(releasedIpAllocation, NotNil)
	check.Assert(releasedIpAllocation.IpSpaceIpAllocation.UsageState, Equals, types.IpSpaceIpAllocationUnused)

	err = updatedIpAllocationManual.Delete()
	check.Assert(err, IsNil)

	// Get IP Space by ID
	ipSpace, err := vcd.client.GetIpSpaceById(ipSpaceId)
	check.Assert(err, IsNil)

	// Attempt to search for allocations when none exist
	allAllocations, err := ipSpace.GetAllIpSpaceAllocations(ipSpaceAllocationRequest.Type, nil)
	check.Assert(err, IsNil)
	check.Assert(len(allAllocations), Equals, 0)

	// allocate IP
	allocationByIpSpaceResult, err := ipSpace.AllocateIp(vcd.org.Org.ID, vcd.org.Org.Name, ipSpaceAllocationRequest)
	check.Assert(err, IsNil)
	check.Assert(len(allocationByIpSpaceResult), Equals, 1)

	// Remove
	ipAllocationByIpSpaceResult, err := vcd.org.GetIpSpaceAllocationByTypeAndValue(ipSpaceId, ipSpaceAllocationRequest.Type, allocationByIpSpaceResult[0].Value, nil)
	check.Assert(err, IsNil)
	check.Assert(ipAllocation, NotNil)

	err = ipAllocationByIpSpaceResult.Delete()
	check.Assert(err, IsNil)

}

func performIpSuggestionChecks(vcd *TestVCD, check *C, ipSpaceId, edgeGatewayId string, ipSpaceAllocationRequest *types.IpSpaceIpAllocationRequest) {
	// Get IP suggestions without additional filters
	floatingIpSuggestions, err := vcd.client.GetAllIpSpaceFloatingIpSuggestions(edgeGatewayId, nil)
	check.Assert(err, IsNil)
	check.Assert(len(floatingIpSuggestions) > 0, Equals, true)
	if ipSpaceAllocationRequest.Type == "FLOATING_IP" {
		check.Assert(len(floatingIpSuggestions[0].UnusedValues), Equals, 1)
	}

	// Get IP suggestions only for IPv4
	queryParams := url.Values{}
	queryParams.Set("filter", "ipType==IPV4")
	floatingIpSuggestionsWithFilterIpv4, err := vcd.client.GetAllIpSpaceFloatingIpSuggestions(edgeGatewayId, queryParams)
	check.Assert(err, IsNil)
	check.Assert(len(floatingIpSuggestionsWithFilterIpv4) > 0, Equals, true)
	if ipSpaceAllocationRequest.Type == "FLOATING_IP" {
		check.Assert(len(floatingIpSuggestionsWithFilterIpv4[0].UnusedValues), Equals, 1)
	}

	// Get IP suggestions only for IPv6
	queryParams.Set("filter", "ipType==IPV6")
	floatingIpSuggestionsWithFilterIpv6, err := vcd.client.GetAllIpSpaceFloatingIpSuggestions(edgeGatewayId, queryParams)
	check.Assert(err, IsNil)
	check.Assert(len(floatingIpSuggestionsWithFilterIpv6) > 0, Equals, false)

	// check IP suggestions with invalid Edge Gateway - it returns ACCESS_TO_RESOURCE_IS_FORBIDDEN
	// queryParams.Set("filter", fmt.Sprintf("gatewayId==%s", "urn:vcloud:gateway:00000000-0000-0000-0000-000000000000"))
	floatingIpSuggestions2, err := vcd.client.GetAllIpSpaceFloatingIpSuggestions("urn:vcloud:gateway:00000000-0000-0000-0000-000000000000", nil)
	check.Assert(strings.Contains(err.Error(), "ACCESS_TO_RESOURCE_IS_FORBIDDEN"), Equals, true)
	check.Assert(floatingIpSuggestions2, IsNil)

	// check with empty filter - it cannot be used this way (edge gateway is mandatory)
	floatingIpSuggestions3, err := vcd.client.GetAllIpSpaceFloatingIpSuggestions("", nil)
	check.Assert(strings.Contains(err.Error(), "edge gateway ID is mandatory"), Equals, true)
	check.Assert(floatingIpSuggestions3, IsNil)
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

	time.Sleep(3 * time.Second)
	err = vcd.client.Client.WaitForRouteAdvertisementTasks()
	check.Assert(err, IsNil)

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
