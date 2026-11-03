//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmRegionalNetworkingSetting(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()

	edgeClusterName := vcd.config.Tm.NsxtEdgeCluster
	if vcd.config.Tm.NsxtEdgeClusterSuffixRequired {
		edgeClusterName = fmt.Sprintf("%s-%s", vcd.config.Tm.NsxtEdgeCluster, nsxtManager.NsxtManagerOpenApi.ID[strings.LastIndex(nsxtManager.NsxtManagerOpenApi.ID, "-")+1:])
	}

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()

	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()
	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()

	pg, pgCleanup := createTmProviderGateway(vcd, region, check)
	defer pgCleanup()

	shortLogNamecleanup := setOrgShortLogname(vcd, org, check)
	defer shortLogNamecleanup()

	orgNetworkSettings := &types.TmRegionalNetworkingSetting{
		Name:               "test-terraform",
		OrgRef:             types.OpenApiReference{ID: org.TmOrg.ID},
		RegionRef:          types.OpenApiReference{ID: region.Region.ID},
		ProviderGatewayRef: &types.OpenApiReference{ID: pg.TmProviderGateway.ID},
	}

	// TODO: TM: This one "conflicts" with the CreateTmRegionalNetworkingSetting below, as there is work done in background
	// in NSX while the task in VCFA is finished, so it fails
	//rnsAsyncTask, err := vcd.client.CreateTmRegionalNetworkingSettingAsync(orgNetworkSettings)
	//check.Assert(err, IsNil)
	//check.Assert(rnsAsyncTask, NotNil)
	//err = rnsAsyncTask.WaitTaskCompletion()
	//check.Assert(err, IsNil)
	//
	//byIdAsync, err := vcd.client.GetTmRegionalNetworkingSettingById(rnsAsyncTask.Task.Owner.ID)
	//check.Assert(err, IsNil)
	//err = byIdAsync.Delete()
	//check.Assert(err, IsNil)

	rns, err := vcd.client.CreateTmRegionalNetworkingSetting(orgNetworkSettings)
	check.Assert(err, IsNil)
	check.Assert(rns, NotNil)
	PrependToCleanupListOpenApi(rns.TmRegionalNetworkingSetting.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmRegionalNetworkingSettings+rns.TmRegionalNetworkingSetting.ID)
	defer func() {
		err = rns.Delete()
		check.Assert(err, IsNil)

		_, err = vcd.client.GetTmRegionalNetworkingSettingById(rns.TmRegionalNetworkingSetting.ID)
		check.Assert(ContainsNotFound(err), Equals, true)
	}()

	// Get all
	allTmRs, err := vcd.client.GetAllTmRegionalNetworkingSettings(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allTmRs), Equals, 1)
	check.Assert(allTmRs[0].TmRegionalNetworkingSetting.Name, Equals, orgNetworkSettings.Name)

	// Get By Name
	byName, err := vcd.client.GetTmRegionalNetworkingSettingByName(orgNetworkSettings.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.TmRegionalNetworkingSetting, DeepEquals, allTmRs[0].TmRegionalNetworkingSetting)

	// By Id
	byId, err := vcd.client.GetTmRegionalNetworkingSettingById(byName.TmRegionalNetworkingSetting.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmRegionalNetworkingSetting, DeepEquals, allTmRs[0].TmRegionalNetworkingSetting)

	// Get by Name and Org ID
	byNameAndOrgId, err := vcd.client.GetTmRegionalNetworkingSettingByNameAndOrgId(orgNetworkSettings.Name, org.TmOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmRegionalNetworkingSetting, DeepEquals, byNameAndOrgId.TmRegionalNetworkingSetting)

	// Get by Name and Region ID
	byNameAndRegionId, err := vcd.client.GetTmRegionalNetworkingSettingByNameAndRegionId(orgNetworkSettings.Name, region.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmRegionalNetworkingSetting, DeepEquals, byNameAndRegionId.TmRegionalNetworkingSetting)

	// Update testing (only Name and Edge Cluster are updatable)
	// Lookup Edge Cluster to specify in update of Regional Network Settings
	edgeCluster, err := vcd.client.GetTmEdgeClusterByName(edgeClusterName)
	check.Assert(err, IsNil)
	check.Assert(edgeCluster, NotNil)

	orgNetworkSettings.ID = byName.TmRegionalNetworkingSetting.ID
	orgNetworkSettings.Name = orgNetworkSettings.Name + "-update"
	orgNetworkSettings.ServiceEdgeClusterRef = &types.OpenApiReference{ID: edgeCluster.TmEdgeCluster.ID}

	updated, err := byName.Update(orgNetworkSettings)
	check.Assert(err, IsNil)
	check.Assert(updated, NotNil)
	check.Assert(updated.TmRegionalNetworkingSetting.Name, Equals, orgNetworkSettings.Name)
	check.Assert(updated.TmRegionalNetworkingSetting.ServiceEdgeClusterRef.ID, Equals, edgeCluster.TmEdgeCluster.ID)

	// Testing VPC Connectivity Profile QoS
	vpcProfile, err := byName.GetDefaultVpcConnectivityProfile()
	check.Assert(err, IsNil)
	check.Assert(vpcProfile, NotNil)
	check.Assert(vpcProfile.ServiceEdgeClusterRef, NotNil)

	// fetch parent Edge Cluster for the VPC profile
	parentEdgeCluster, err := vcd.client.GetTmEdgeClusterById(vpcProfile.ServiceEdgeClusterRef.ID)
	check.Assert(err, IsNil)
	check.Assert(parentEdgeCluster, NotNil)

	// By default, the VPC QoS Profile contains values from parent Edge Cluster
	check.Assert(vpcProfile.QosConfig.EgressProfile.Type, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.Type)
	check.Assert(vpcProfile.QosConfig.EgressProfile.BurstSizeBytes, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.BurstSizeBytes)
	check.Assert(vpcProfile.QosConfig.EgressProfile.CommittedBandwidthMbps, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps)

	check.Assert(vpcProfile.QosConfig.IngressProfile.Type, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.Type)
	check.Assert(vpcProfile.QosConfig.IngressProfile.BurstSizeBytes, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.BurstSizeBytes)
	check.Assert(vpcProfile.QosConfig.IngressProfile.CommittedBandwidthMbps, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps)

	// Update VPC connectivity profile to custom values
	newQosConfig := &types.VpcConnectivityProfileQosConfig{
		IngressProfile: &types.VpcConnectivityProfileQosProfile{
			Type:                   "CUSTOM",
			CommittedBandwidthMbps: 8,
			BurstSizeBytes:         8,
		},
		EgressProfile: &types.VpcConnectivityProfileQosProfile{
			Type:                   "CUSTOM",
			CommittedBandwidthMbps: 8,
			BurstSizeBytes:         8,
		},
	}

	vpcProfile.QosConfig = newQosConfig // Leave body of VPC Connectivity profile as is, but update values to custom
	updatedVpcProfile, err := byName.UpdateDefaultVpcConnectivityProfile(vpcProfile)
	check.Assert(err, IsNil)
	check.Assert(updatedVpcProfile, NotNil)

	// Check that the new values are in place
	check.Assert(updatedVpcProfile.QosConfig.EgressProfile.Type, Equals, newQosConfig.EgressProfile.Type)
	check.Assert(updatedVpcProfile.QosConfig.EgressProfile.BurstSizeBytes, Equals, newQosConfig.EgressProfile.BurstSizeBytes)
	check.Assert(updatedVpcProfile.QosConfig.EgressProfile.CommittedBandwidthMbps, Equals, newQosConfig.EgressProfile.CommittedBandwidthMbps)
	check.Assert(updatedVpcProfile.QosConfig.IngressProfile.Type, Equals, newQosConfig.IngressProfile.Type)
	check.Assert(updatedVpcProfile.QosConfig.IngressProfile.BurstSizeBytes, Equals, newQosConfig.IngressProfile.BurstSizeBytes)
	check.Assert(updatedVpcProfile.QosConfig.IngressProfile.CommittedBandwidthMbps, Equals, newQosConfig.IngressProfile.CommittedBandwidthMbps)

	// Check that the parent Edge Cluster is not affected
	parentEdgeCluster, err = vcd.client.GetTmEdgeClusterById(vpcProfile.ServiceEdgeClusterRef.ID)
	check.Assert(err, IsNil)
	check.Assert(parentEdgeCluster, NotNil)

	check.Assert(updatedVpcProfile.QosConfig.EgressProfile.Type, Not(Equals), parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.Type)
	check.Assert(updatedVpcProfile.QosConfig.EgressProfile.BurstSizeBytes, Not(Equals), parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.BurstSizeBytes)
	check.Assert(updatedVpcProfile.QosConfig.EgressProfile.CommittedBandwidthMbps, Not(Equals), parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps)
	check.Assert(updatedVpcProfile.QosConfig.IngressProfile.Type, Not(Equals), parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.Type)
	check.Assert(updatedVpcProfile.QosConfig.IngressProfile.BurstSizeBytes, Not(Equals), parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.BurstSizeBytes)
	check.Assert(updatedVpcProfile.QosConfig.IngressProfile.CommittedBandwidthMbps, Not(Equals), parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps)

	// Update VPC QoS profile back to Edge Cluster defaults

	updatedVpcProfile.QosConfig = &parentEdgeCluster.TmEdgeCluster.DefaultQosConfig
	removedVpcQos, err := byName.UpdateDefaultVpcConnectivityProfile(updatedVpcProfile)
	check.Assert(err, IsNil)
	check.Assert(removedVpcQos, NotNil)

	// Check that the values are equal between Edge Cluster Qos and VPC QoS profile
	check.Assert(removedVpcQos.QosConfig.EgressProfile.Type, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.Type)
	check.Assert(removedVpcQos.QosConfig.EgressProfile.BurstSizeBytes, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.BurstSizeBytes)
	check.Assert(removedVpcQos.QosConfig.EgressProfile.CommittedBandwidthMbps, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps)
	check.Assert(removedVpcQos.QosConfig.IngressProfile.Type, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.Type)
	check.Assert(removedVpcQos.QosConfig.IngressProfile.BurstSizeBytes, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.BurstSizeBytes)
	check.Assert(removedVpcQos.QosConfig.IngressProfile.CommittedBandwidthMbps, Equals, parentEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps)
}

// Write tests use an isolated organization and setting. Discovered backing objects
// are only assigned and unassigned, never deleted.
func tmTestRegionalSetting(vcd *TestVCD, region *Region, check *C) (*TmRegionalNetworkingSetting, func()) {
	var cleanups []func()
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}
	ready := false
	defer func() {
		if !ready {
			cleanup()
		}
	}()
	org, orgCleanup := createOrg(vcd, check, false)
	cleanups = append(cleanups, orgCleanup)
	pg, pgCleanup := createTmProviderGateway(vcd, region, check)
	cleanups = append(cleanups, pgCleanup)
	logCleanup := setOrgShortLogname(vcd, org, check)
	cleanups = append(cleanups, logCleanup)
	setting, err := vcd.client.CreateTmRegionalNetworkingSetting(&types.TmRegionalNetworkingSetting{
		Name: check.TestName(), OrgRef: types.OpenApiReference{ID: org.TmOrg.ID}, RegionRef: types.OpenApiReference{ID: region.Region.ID}, ProviderGatewayRef: &types.OpenApiReference{ID: pg.TmProviderGateway.ID},
	})
	check.Assert(err, IsNil)
	PrependToCleanupListOpenApi(setting.TmRegionalNetworkingSetting.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmRegionalNetworkingSettings+setting.TmRegionalNetworkingSetting.ID)
	cleanups = append(cleanups, func() { check.Check(setting.Delete(), IsNil) })
	ready = true
	return setting, cleanup
}

func testTmRegionalAssignment[T any](check *C, config *T, options TmOperationOptions,
	list func(url.Values, ...TmOperationOptions) ([]*T, error),
	get func(string, ...TmOperationOptions) (*T, error),
	create func(*T, ...TmOperationOptions) error,
	createAsync func(*T, ...TmOperationOptions) (*Task, error),
	remove func(string, ...TmOperationOptions) error,
	removeAsync func(string, ...TmOperationOptions) (*Task, error),
	id func(*T) string, verify func(*T)) {
	for _, async := range []bool{false, true} {
		before, err := list(url.Values{"pageSize": {"1"}}, options)
		check.Assert(err, IsNil)
		existing := map[string]bool{}
		for _, row := range before {
			existing[id(row)] = true
		}
		if async {
			task, err := createAsync(config, options)
			tmTestWaitTask(check, task, err)
		} else {
			check.Assert(create(config, options), IsNil)
		}
		after, err := list(url.Values{"pageSize": {"1"}}, options)
		check.Assert(err, IsNil)
		check.Assert(len(after), Equals, len(before)+1)
		var created *T
		for _, row := range after {
			if !existing[id(row)] {
				check.Assert(created, IsNil)
				created = row
			}
		}
		check.Assert(created, NotNil)
		check.Assert(id(created), Not(Equals), "")
		verify(created)
		byID, err := get(id(created), options)
		check.Assert(err, IsNil)
		check.Assert(id(byID), Equals, id(created))
		verify(byID)
		if async {
			task, err := removeAsync(id(created), options)
			tmTestWaitTask(check, task, err)
		} else {
			check.Assert(remove(id(created), options), IsNil)
		}
		remaining, err := list(url.Values{"pageSize": {"1"}}, options)
		check.Assert(err, IsNil)
		check.Assert(len(remaining), Equals, len(before))
		for _, row := range remaining {
			check.Assert(existing[id(row)], Equals, true)
		}
		_, err = get(id(created), options)
		check.Assert(ContainsNotFound(err), Equals, true)
	}
}

func (vcd *TestVCD) Test_TmRegionalProviderGatewayConnections(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := TmOperationOptions{}

	region := tmTestExistingRegion(vcd, check)
	candidates, err := vcd.client.GetAllTmProviderGateways(nil)
	check.Assert(err, IsNil)
	var config *types.TmRegionalProviderGatewayConnection
	for _, candidate := range candidates {
		backing := candidate.TmProviderGateway
		if backing.RegionRef.ID != region.Region.ID || backing.Status != "REALIZED" {
			continue
		}
		if backing.OrgRef != nil || backing.AllowAdvertisingPrivateIpBlocks {
			continue
		}
		config = &types.TmRegionalProviderGatewayConnection{ProviderGatewayRef: &types.OpenApiReference{ID: backing.ID}, IsDefault: addrOf(false), AllowAdvertisingPrivateIpBlocks: addrOf(false)}
		break
	}
	if config == nil {
		check.Skip("no compatible ProviderGatewayConnection backing discovered in the region")
	}
	// Provision the test setting only after finding a compatible backing. Its
	// new provider gateway is distinct from the existing candidate above.
	setting, cleanup := tmTestRegionalSetting(vcd, region, check)
	defer cleanup()

	testTmRegionalAssignment(check, config, options, setting.GetAllProviderGatewayConnections, setting.GetProviderGatewayConnectionById,
		setting.CreateProviderGatewayConnection, setting.CreateProviderGatewayConnectionAsync, setting.DeleteProviderGatewayConnection, setting.DeleteProviderGatewayConnectionAsync,
		func(row *types.TmRegionalProviderGatewayConnection) string { return row.ID }, func(row *types.TmRegionalProviderGatewayConnection) {
			check.Assert(row.ProviderGatewayRef, NotNil)
			check.Assert(row.ProviderGatewayRef.ID, Equals, config.ProviderGatewayRef.ID)
		})
}

func (vcd *TestVCD) Test_TmRegionalSharedSubnets(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := TmOperationOptions{}

	region := tmTestExistingRegion(vcd, check)
	candidates, err := vcd.client.GetAllTmSharedSubnets(nil)
	check.Assert(err, IsNil)
	var config *types.TmRegionalSharedSubnet
	for _, candidate := range candidates {
		backing := candidate.TmSharedSubnet
		if backing.RegionRef.ID != region.Region.ID || backing.Status != "REALIZED" {
			continue
		}
		config = &types.TmRegionalSharedSubnet{SubnetRef: &types.OpenApiReference{ID: backing.ID}}
		break
	}
	if config == nil {
		check.Skip("no compatible SharedSubnet backing discovered in the region")
	}
	setting, cleanup := tmTestRegionalSetting(vcd, region, check)
	defer cleanup()

	testTmRegionalAssignment(check, config, options, setting.GetAllSharedSubnets, setting.GetSharedSubnetById,
		setting.CreateSharedSubnet, setting.CreateSharedSubnetAsync, setting.DeleteSharedSubnet, setting.DeleteSharedSubnetAsync,
		func(row *types.TmRegionalSharedSubnet) string { return row.ID }, func(row *types.TmRegionalSharedSubnet) {
			check.Assert(row.SubnetRef, NotNil)
			check.Assert(row.SubnetRef.ID, Equals, config.SubnetRef.ID)
		})
}

func (vcd *TestVCD) Test_TmRegionalDistributedVlanConnections(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := TmOperationOptions{}

	region := tmTestExistingRegion(vcd, check)
	candidates, err := vcd.client.GetAllTmDistributedVlanConnections(nil)
	check.Assert(err, IsNil)
	var config *types.TmRegionalDistributedVlanConnection
	for _, candidate := range candidates {
		backing := candidate.TmDistributedVlanConnection
		if backing.RegionRef.ID != region.Region.ID || backing.Status != "REALIZED" {
			continue
		}
		if backing.DedicatedOrgRef != nil || backing.DedicatedSharedSubnetRef != nil || backing.SubnetExclusive {
			continue
		}
		config = &types.TmRegionalDistributedVlanConnection{DistributedVlanConnectionRef: &types.OpenApiReference{ID: backing.ID}, IsDefault: addrOf(false), VlanId: addrOf(backing.VlanId), GatewayCidr: backing.GatewayCidr, IpSpaceRef: backing.IpSpaceRef}
		break
	}
	if config == nil {
		check.Skip("no compatible DistributedVlanConnection backing discovered in the region")
	}
	setting, cleanup := tmTestRegionalSetting(vcd, region, check)
	defer cleanup()

	testTmRegionalAssignment(check, config, options, setting.GetAllDistributedVlanConnections, setting.GetDistributedVlanConnectionById,
		setting.CreateDistributedVlanConnection, setting.CreateDistributedVlanConnectionAsync, setting.DeleteDistributedVlanConnection, setting.DeleteDistributedVlanConnectionAsync,
		func(row *types.TmRegionalDistributedVlanConnection) string { return row.ID }, func(row *types.TmRegionalDistributedVlanConnection) {
			check.Assert(row.DistributedVlanConnectionRef, NotNil)
			check.Assert(row.DistributedVlanConnectionRef.ID, Equals, config.DistributedVlanConnectionRef.ID)
		})
}

func (vcd *TestVCD) Test_TmRegionalDistributedVxlanConnections(check *C) {
	requireTmTestVersion(vcd, check, "9.1.1")
	options := TmOperationOptions{}
	settings, err := vcd.client.GetAllTmRegionalNetworkingSettings(nil)
	check.Assert(err, IsNil)
	found := false
	for _, setting := range settings {
		rows, err := setting.GetAllDistributedVxlanConnections(url.Values{"pageSize": {"1"}}, options)
		check.Assert(err, IsNil)
		for _, row := range rows {
			found = true
			check.Assert(row.ID, Not(Equals), "")
			actual, err := setting.GetDistributedVxlanConnectionById(row.ID, options)
			check.Assert(err, IsNil)
			check.Assert(actual.ID, Equals, row.ID)
			check.Assert(actual.DistributedVxlanConnectionRef, NotNil)
			check.Assert(actual.DistributedVxlanConnectionRef, DeepEquals, row.DistributedVxlanConnectionRef)
		}
	}
	if !found {
		check.Skip("no existing DistributedVxlanConnections assignments found")
	}
}

func (vcd *TestVCD) Test_TmRegionalAviSetting(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	settings, err := vcd.client.GetAllTmRegionalNetworkingSettings(nil)
	check.Assert(err, IsNil)
	if len(settings) == 0 {
		check.Skip("no regional networking settings available")
	}
	for _, setting := range settings {
		avi, err := setting.GetAviSetting()
		check.Assert(err, IsNil)
		check.Assert(avi, NotNil)
		if avi.Active {
			check.Assert(avi.ServiceEngineGroupMode, Not(Equals), "")
		}
	}
}
