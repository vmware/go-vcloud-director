//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmRegionalNetworkingSetting(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
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
		Name:               "terraform-test",
		OrgRef:             types.OpenApiReference{ID: org.TmOrg.ID},
		RegionRef:          types.OpenApiReference{ID: region.Region.ID},
		ProviderGatewayRef: types.OpenApiReference{ID: pg.TmProviderGateway.ID},
	}

	rnsAsyncTask, err := vcd.client.CreateTmRegionalNetworkingSettingAsync(orgNetworkSettings)
	check.Assert(err, IsNil)
	check.Assert(rnsAsyncTask, NotNil)
	err = rnsAsyncTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	byIdAsync, err := vcd.client.GetTmRegionalNetworkingSettingById(rnsAsyncTask.Task.Owner.ID)
	check.Assert(err, IsNil)
	err = byIdAsync.Delete()
	check.Assert(err, IsNil)

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
	edgeCluster, err := vcd.client.GetTmEdgeClusterByName(vcd.config.Tm.NsxtEdgeCluster)
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
}
