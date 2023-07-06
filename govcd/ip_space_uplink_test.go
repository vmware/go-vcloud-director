//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_IpSpaceUplink(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointIpSpaceUplinks)

	// Create External Network (Provider Gateway)
	extNet := createExternalNetwork(vcd, check)

	// Create IP Space
	ipSpace := createIpSpace(vcd, check)

	// Create Uplink configuration
	uplinkConfig := &types.IpSpaceUplink{
		Name:               check.TestName(),
		Description:        "IP SPace Uplink for External Network (Provider Gateway)",
		ExternalNetworkRef: &types.OpenApiReference{ID: extNet.ExternalNetwork.ID},
		IPSpaceRef:         &types.OpenApiReference{ID: ipSpace.IpSpace.ID},
	}

	createdIpSpaceUplink, err := vcd.client.CreateIpSpaceUplink(uplinkConfig)
	check.Assert(err, IsNil)
	check.Assert(createdIpSpaceUplink, NotNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks + createdIpSpaceUplink.IpSpaceUplink.ID
	AddToCleanupListOpenApi(createdIpSpaceUplink.IpSpaceUplink.Name, check.TestName(), openApiEndpoint)

	// Operations on IP Space related entities trigger a separate task
	// 'ipSpaceUplinkRouteAdvertisementSync' which is better to finish before any other operations
	// as it might cause an error: busy completing an operation IP_SPACE_UPLINK_ROUTE_ADVERTISEMENT_SYNC
	// Sleeping a few seconds because the task is not immediately seen sometimes.
	time.Sleep(3 * time.Second)
	err = vcd.client.Client.WaitForRouteAdvertisementTasks()
	check.Assert(err, IsNil)

	// Get all IP Space Uplinks
	allIpSpaceUplinks, err := vcd.client.GetAllIpSpaceUplinks(extNet.ExternalNetwork.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(len(allIpSpaceUplinks) > 0, Equals, true)

	// Get by ID
	byId, err := vcd.client.GetIpSpaceUplinkById(createdIpSpaceUplink.IpSpaceUplink.ID)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)
	check.Assert(byId.IpSpaceUplink, DeepEquals, createdIpSpaceUplink.IpSpaceUplink)

	// Get by Name
	byName, err := vcd.client.GetIpSpaceUplinkByName(extNet.ExternalNetwork.ID, byId.IpSpaceUplink.Name)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)
	check.Assert(byName.IpSpaceUplink, DeepEquals, byId.IpSpaceUplink)

	// Update
	uplinkConfig.Name = check.TestName() + "updated"
	uplinkConfig.Description = uplinkConfig.Description + "updated"
	updatedUplinkConfig, err := createdIpSpaceUplink.Update(uplinkConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedUplinkConfig.IpSpaceUplink.ID, Equals, byId.IpSpaceUplink.ID)
	check.Assert(updatedUplinkConfig.IpSpaceUplink.ID, Equals, createdIpSpaceUplink.IpSpaceUplink.ID)
	check.Assert(updatedUplinkConfig.IpSpaceUplink.Name, Equals, uplinkConfig.Name)
	check.Assert(updatedUplinkConfig.IpSpaceUplink.Description, Equals, uplinkConfig.Description)
	check.Assert(updatedUplinkConfig.IpSpaceUplink.ExternalNetworkRef.ID, Equals, createdIpSpaceUplink.IpSpaceUplink.ExternalNetworkRef.ID)
	check.Assert(updatedUplinkConfig.IpSpaceUplink.IPSpaceRef.ID, Equals, createdIpSpaceUplink.IpSpaceUplink.IPSpaceRef.ID)

	// Read-only variables
	check.Assert(updatedUplinkConfig.IpSpaceUplink.IPSpaceType, Equals, types.IpSpacePublic)
	check.Assert(updatedUplinkConfig.IpSpaceUplink.Status, Equals, "REALIZED")

	time.Sleep(3 * time.Second)
	err = vcd.client.Client.WaitForRouteAdvertisementTasks()
	check.Assert(err, IsNil)

	err = createdIpSpaceUplink.Delete()
	check.Assert(err, IsNil)

	// Check that IP Space Uplink was deleted
	_, err = vcd.client.GetIpSpaceUplinkById(updatedUplinkConfig.IpSpaceUplink.ID)
	check.Assert(ContainsNotFound(err), Equals, true)

	err = extNet.Delete()
	check.Assert(err, IsNil)

	err = ipSpace.Delete()
	check.Assert(err, IsNil)
}

func createIpSpace(vcd *TestVCD, check *C) *IpSpace {
	ipSpaceConfig := &types.IpSpace{
		Name:                      check.TestName(),
		IPSpaceInternalScope:      []string{"22.0.0.0/24"},
		IPSpaceExternalScope:      "200.0.0.1/24",
		Type:                      types.IpSpacePublic,
		RouteAdvertisementEnabled: false,
		IPSpacePrefixes: []types.IPSpacePrefixes{
			{
				DefaultQuotaForPrefixLength: -1,
				IPPrefixSequence: []types.IPPrefixSequence{
					{
						StartingPrefixIPAddress: "22.0.0.200",
						PrefixLength:            31,
						TotalPrefixCount:        3,
					},
				},
			},
			{
				DefaultQuotaForPrefixLength: 2,
				IPPrefixSequence: []types.IPPrefixSequence{
					{
						StartingPrefixIPAddress: "22.0.0.100",
						PrefixLength:            30,
						TotalPrefixCount:        3,
					},
				},
			},
		},
		IPSpaceRanges: types.IPSpaceRanges{
			DefaultFloatingIPQuota: 3,
			IPRanges: []types.IpSpaceRangeValues{
				{
					StartIPAddress: "22.0.0.10",
					EndIPAddress:   "22.0.0.30",
				},
				{
					StartIPAddress: "22.0.0.32",
					EndIPAddress:   "22.0.0.34",
				},
			},
		},
	}

	createdIpSpace, err := vcd.client.CreateIpSpace(ipSpaceConfig)
	check.Assert(err, IsNil)
	check.Assert(createdIpSpace, NotNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces + createdIpSpace.IpSpace.ID
	AddToCleanupListOpenApi(createdIpSpace.IpSpace.Name, check.TestName(), openApiEndpoint)

	return createdIpSpace
}

func createExternalNetwork(vcd *TestVCD, check *C) *ExternalNetworkV2 {
	// NSX-T details
	man, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	nsxtManagerId, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", extractUuid(man[0].HREF))
	check.Assert(err, IsNil)

	backingId := getBackingIdByNameAndType(check, vcd.config.VCD.Nsxt.Tier0router, types.ExternalNetworkBackingTypeNsxtTier0Router, vcd, nsxtManagerId)

	net := &types.ExternalNetworkV2{
		Name:        check.TestName(),
		Description: "",
		NetworkBackings: types.ExternalNetworkV2Backings{Values: []types.ExternalNetworkV2Backing{
			{
				BackingID: backingId,
				NetworkProvider: types.NetworkProvider{
					ID: nsxtManagerId,
				},
				BackingTypeValue: types.ExternalNetworkBackingTypeNsxtTier0Router,
			},
		}},
		UsingIpSpace: addrOf(true),
	}

	createdNet, err := CreateExternalNetworkV2(vcd.client, net)
	check.Assert(err, IsNil)
	check.Assert(createdNet, NotNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks + createdNet.ExternalNetwork.ID
	AddToCleanupListOpenApi(createdNet.ExternalNetwork.Name, check.TestName(), openApiEndpoint)

	return createdNet
}
