//go:build network || functional || openapi || vdcGroup || nsxt || gateway || ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtVdcGroupOrgNetworks(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEdgeGateways)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(org, NotNil)
	check.Assert(err, IsNil)

	nsxtExternalNetwork, err := GetExternalNetworkV2ByName(vcd.client, vcd.config.VCD.Nsxt.ExternalNetwork)
	check.Assert(nsxtExternalNetwork, NotNil)
	check.Assert(err, IsNil)

	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)
	check.Assert(vdc, NotNil)
	check.Assert(vdcGroup, NotNil)

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
	check.Assert(createdEdge, NotNil)
	check.Assert(createdEdge.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdc:.*`)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways + createdEdge.EdgeGateway.ID
	PrependToCleanupListOpenApi(createdEdge.EdgeGateway.Name, check.TestName(), openApiEndpoint)

	// Move Edge Gateway to VDC Group
	movedGateway, err := createdEdge.MoveToVdcOrVdcGroup(vdcGroup.VdcGroup.Id)
	check.Assert(err, IsNil)
	check.Assert(movedGateway, NotNil)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Equals, vdcGroup.VdcGroup.Id)
	check.Assert(movedGateway.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdcGroup:.*`)

	mapOfNetworkConfigs := make(map[string]*types.OpenApiOrgVdcNetwork, 3)
	mapOfNetworkConfigs["isolated"] = buildIsolatedOrgVdcNetworkConfig(check, vcd, vdcGroup.VdcGroup.Id)
	mapOfNetworkConfigs["imported"] = buildImportedOrgVdcNetworkConfig(check, vcd, vdcGroup.VdcGroup.Id)
	mapOfNetworkConfigs["routed"] = buildRoutedOrgVdcNetworkConfig(check, vcd, movedGateway, vdcGroup.VdcGroup.Id)

	sliceOfCreatedNetworkConfigs := make(map[string]*OpenApiOrgVdcNetwork, 3)
	for index, orgVdcNetworkConfig := range mapOfNetworkConfigs {
		orgVdcNet, err := org.CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig)
		check.Assert(err, IsNil)
		check.Assert(orgVdcNet, NotNil)
		check.Assert(orgVdcNet.OpenApiOrgVdcNetwork.OwnerRef.ID, Equals, vdcGroup.VdcGroup.Id)

		sliceOfCreatedNetworkConfigs[index] = orgVdcNet

		// Use generic "OpenApiEntity" resource cleanup type
		openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + orgVdcNet.OpenApiOrgVdcNetwork.ID
		PrependToCleanupListOpenApi(orgVdcNet.OpenApiOrgVdcNetwork.Name, check.TestName(), openApiEndpoint)

		check.Assert(orgVdcNet.GetType(), Equals, orgVdcNetworkConfig.NetworkType)
	}

	// Move Edge Gateway back to VDC
	movedBackToVdcEdge, err := movedGateway.MoveToVdcOrVdcGroup(vdc.Vdc.ID)
	check.Assert(err, IsNil)
	check.Assert(movedBackToVdcEdge, NotNil)
	check.Assert(movedBackToVdcEdge.EdgeGateway.OwnerRef.ID, Matches, `^urn:vcloud:vdc:.*`)

	// Routed networks migrate to/from VDC Groups together with Edge Gateway therefore we need to
	// check that routed network owner ID is the same as Edge Gateway. Routed network must be
	// retrieved again so that it reflects latest information.
	routedOrgNetwork, err := org.GetOpenApiOrgVdcNetworkById(sliceOfCreatedNetworkConfigs["routed"].OpenApiOrgVdcNetwork.ID)
	check.Assert(err, IsNil)
	check.Assert(routedOrgNetwork, NotNil)
	check.Assert(routedOrgNetwork.OpenApiOrgVdcNetwork.OwnerRef.ID, Equals, movedBackToVdcEdge.EdgeGateway.OwnerRef.ID)

	// Remove all created networks
	for _, network := range sliceOfCreatedNetworkConfigs {
		err = network.Delete()
		check.Assert(err, IsNil)
	}

	// Remove Edge Gateway
	err = movedGateway.Delete()
	check.Assert(err, IsNil)

}

func buildIsolatedOrgVdcNetworkConfig(check *C, vcd *TestVCD, ownerId string) *types.OpenApiOrgVdcNetwork {
	isolatedOrgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName() + "-isolated",
		Description: check.TestName() + "-description",

		OwnerRef: &types.OpenApiReference{ID: ownerId},

		NetworkType: types.OrgVdcNetworkTypeIsolated,
		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "4.1.1.1",
					PrefixLength: 25,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "bar.foo",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "4.1.1.20",
								EndAddress:   "4.1.1.30",
							},
							{
								StartAddress: "4.1.1.40",
								EndAddress:   "4.1.1.50",
							},
							{
								StartAddress: "4.1.1.88",
								EndAddress:   "4.1.1.92",
							},
						}},
				},
			},
		},
	}

	return isolatedOrgVdcNetworkConfig
}

func buildImportedOrgVdcNetworkConfig(check *C, vcd *TestVCD, ownerId string) *types.OpenApiOrgVdcNetwork {
	logicalSwitch, err := vcd.nsxtVdc.GetNsxtImportableSwitchByName(vcd.config.VCD.Nsxt.NsxtImportSegment)
	check.Assert(err, IsNil)

	importedOrgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName() + "-imported",
		Description: check.TestName() + "-description",

		OwnerRef: &types.OpenApiReference{ID: ownerId},

		NetworkType: types.OrgVdcNetworkTypeOpaque,
		// BackingNetworkId contains NSX-T logical switch ID for Imported networks
		BackingNetworkId: logicalSwitch.NsxtImportableSwitch.ID,

		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "2.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "foo.bar",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "2.1.1.20",
								EndAddress:   "2.1.1.30",
							},
							{
								StartAddress: "2.1.1.40",
								EndAddress:   "2.1.1.50",
							},
						}},
				},
			},
		},
	}

	return importedOrgVdcNetworkConfig
}

func buildRoutedOrgVdcNetworkConfig(check *C, vcd *TestVCD, edgeGateway *NsxtEdgeGateway, ownerId string) *types.OpenApiOrgVdcNetwork {
	routedOrgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        check.TestName() + "-routed",
		Description: check.TestName() + "-description",

		OwnerRef: &types.OpenApiReference{ID: ownerId},

		NetworkType: types.OrgVdcNetworkTypeRouted,

		// Connection is used for "routed" network
		Connection: &types.Connection{
			RouterRef: types.OpenApiReference{
				ID: edgeGateway.EdgeGateway.ID,
			},
			ConnectionType: "INTERNAL",
		},
		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      "3.1.1.1",
					PrefixLength: 24,
					DNSServer1:   "8.8.8.8",
					DNSServer2:   "8.8.4.4",
					DNSSuffix:    "foo.bar",
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: []types.OrgVdcNetworkSubnetIPRangeValues{
							{
								StartAddress: "3.1.1.20",
								EndAddress:   "3.1.1.30",
							},
							{
								StartAddress: "3.1.1.40",
								EndAddress:   "3.1.1.50",
							},
							{
								StartAddress: "3.1.1.60",
								EndAddress:   "3.1.1.62",
							}, {
								StartAddress: "3.1.1.72",
								EndAddress:   "3.1.1.74",
							}, {
								StartAddress: "3.1.1.84",
								EndAddress:   "3.1.1.85",
							},
						}},
				},
			},
		},
	}

	return routedOrgVdcNetworkConfig
}

func test_CreateVdcGroup(check *C, adminOrg *AdminOrg, vcd *TestVCD) (*Vdc, *VdcGroup) {
	createdVdc := createNewVdc(vcd, check, check.TestName())

	createdVdcAsCandidate, err := adminOrg.GetAllNsxtVdcGroupCandidates(createdVdc.vdcId(),
		map[string][]string{"filter": []string{fmt.Sprintf("name==%s", url.QueryEscape(createdVdc.vdcName()))}})
	check.Assert(err, IsNil)
	check.Assert(createdVdcAsCandidate, NotNil)
	check.Assert(len(createdVdcAsCandidate) == 1, Equals, true)

	existingVdcAsCandidate, err := adminOrg.GetAllNsxtVdcGroupCandidates(createdVdc.vdcId(),
		map[string][]string{"filter": []string{fmt.Sprintf("name==%s", url.QueryEscape(vcd.nsxtVdc.vdcName()))}})
	check.Assert(err, IsNil)
	check.Assert(existingVdcAsCandidate, NotNil)
	check.Assert(len(existingVdcAsCandidate) == 1, Equals, true)

	vdcGroupConfig := &types.VdcGroup{
		Name:  check.TestName() + "Group",
		OrgId: adminOrg.orgId(),
		ParticipatingOrgVdcs: []types.ParticipatingOrgVdcs{
			types.ParticipatingOrgVdcs{
				VdcRef: types.OpenApiReference{
					ID: createdVdc.vdcId(),
				},
				SiteRef: (createdVdcAsCandidate)[0].SiteRef,
				OrgRef:  (createdVdcAsCandidate)[0].OrgRef,
			},
			types.ParticipatingOrgVdcs{
				VdcRef: types.OpenApiReference{
					ID: vcd.nsxtVdc.vdcId(),
				},
				SiteRef: (existingVdcAsCandidate)[0].SiteRef,
				OrgRef:  (existingVdcAsCandidate)[0].OrgRef,
			},
		},
		LocalEgress:                false,
		UniversalNetworkingEnabled: false,
		NetworkProviderType:        "NSX_T",
		Type:                       "LOCAL",
		//DfwEnabled: true, // ignored by API
	}

	vdcGroup, err := adminOrg.CreateVdcGroup(vdcGroupConfig)
	check.Assert(err, IsNil)
	check.Assert(vdcGroup, NotNil)
	check.Assert(vdcGroup.IsNsxt(), Equals, true)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups + vdcGroup.VdcGroup.Id
	PrependToCleanupListOpenApi(vdcGroup.VdcGroup.Name, check.TestName(), openApiEndpoint)

	return createdVdc, vdcGroup
}

func createNewVdc(vcd *TestVCD, check *C, vdcName string) *Vdc {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	pVdcs, err := QueryProviderVdcByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	if len(pVdcs) == 0 {
		check.Skip(fmt.Sprintf("No NSX-T Provider VDC found with name '%s'", vcd.config.VCD.NsxtProviderVdc.Name))
	}
	providerVdcHref := pVdcs[0].HREF
	pvdcStorageProfile, err := vcd.client.QueryProviderVdcStorageProfileByName(vcd.config.VCD.NsxtProviderVdc.StorageProfile, providerVdcHref)
	check.Assert(err, IsNil)
	check.Assert(pvdcStorageProfile, NotNil)
	providerVdcStorageProfileHref := pvdcStorageProfile.HREF

	networkPools, err := QueryNetworkPoolByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.NetworkPool)
	check.Assert(err, IsNil)
	if len(networkPools) == 0 {
		check.Skip(fmt.Sprintf("No network pool found with name '%s'", vcd.config.VCD.NsxtProviderVdc.NetworkPool))
	}

	networkPoolHref := networkPools[0].HREF
	trueValue := true
	vdcConfiguration := &types.VdcConfiguration{
		Name:            vdcName,
		Xmlns:           types.XMLNamespaceVCloud,
		AllocationModel: "Flex",
		ComputeCapacity: []*types.ComputeCapacity{
			&types.ComputeCapacity{
				CPU: &types.CapacityWithUsage{
					Units:     "MHz",
					Allocated: 1024,
					Limit:     1024,
				},
				Memory: &types.CapacityWithUsage{
					Allocated: 1024,
					Limit:     1024,
					Units:     "MB",
				},
			},
		},
		VdcStorageProfile: []*types.VdcStorageProfileConfiguration{&types.VdcStorageProfileConfiguration{
			Enabled: takeBoolPointer(true),
			Units:   "MB",
			Limit:   1024,
			Default: true,
			ProviderVdcStorageProfile: &types.Reference{
				HREF: providerVdcStorageProfileHref,
			},
		},
		},
		NetworkPoolReference: &types.Reference{
			HREF: networkPoolHref,
		},
		ProviderVdcReference: &types.Reference{
			HREF: providerVdcHref,
		},
		IsEnabled:             true,
		IsThinProvision:       true,
		UsesFastProvisioning:  true,
		IsElastic:             &trueValue,
		IncludeMemoryOverhead: &trueValue,
	}

	vdc, err := adminOrg.CreateOrgVdc(vdcConfiguration)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, check.TestName())
	return vdc
}
