//go:build slz || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_CreateLandingZone(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("Solution Landing Zones are supported in VCD 10.4.1+")
	}

	if vcd.config.VCD.Nsxt.RoutedNetwork == "" {
		check.Skip("Solution Landing Zones require 'vcd.config.VCD.Nsxt.RoutedNetwork' to be present")
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCById(vcd.nsxtVdc.Vdc.ID, false)
	check.Assert(err, IsNil)

	orgNetwork, err := vcd.nsxtVdc.GetOpenApiOrgVdcNetworkByName(vcd.config.VCD.Nsxt.RoutedNetwork)
	check.Assert(err, IsNil)
	check.Assert(orgNetwork, NotNil)

	computePolicy, err := adminVdc.GetAllAssignedVdcComputePoliciesV2(nil)
	check.Assert(err, IsNil)
	check.Assert(computePolicy, NotNil)

	storageProfileRef, err := adminVdc.GetDefaultStorageProfileReference()
	check.Assert(err, IsNil)
	check.Assert(storageProfileRef, NotNil)

	catalog, err := adminOrg.GetCatalogByName(vcd.config.VCD.Catalog.NsxtBackedCatalogName, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	slzCfg := &types.SolutionLandingZoneType{
		Name: adminOrg.AdminOrg.Name,
		ID:   adminOrg.AdminOrg.ID,
		Vdcs: []types.SolutionLandingZoneVdc{
			{
				ID:           adminVdc.AdminVdc.ID,
				Name:         adminVdc.AdminVdc.Name,
				Capabilities: []string{},
				Networks: []types.SolutionLandingZoneVdcChild{
					{
						ID:           orgNetwork.OpenApiOrgVdcNetwork.ID,
						Name:         orgNetwork.OpenApiOrgVdcNetwork.Name,
						IsDefault:    true,
						Capabilities: []string{},
					},
				},
				ComputePolicies: []types.SolutionLandingZoneVdcChild{
					{
						ID:           computePolicy[0].VdcComputePolicyV2.ID,
						Name:         computePolicy[0].VdcComputePolicyV2.Name,
						IsDefault:    true,
						Capabilities: []string{},
					},
				},
				StoragePolicies: []types.SolutionLandingZoneVdcChild{
					{
						ID:           storageProfileRef.ID,
						Name:         storageProfileRef.Name,
						IsDefault:    true,
						Capabilities: []string{},
					},
				},
			},
		},
		Catalogs: []types.SolutionLandingZoneCatalog{
			{
				ID:           catalog.Catalog.ID,
				Name:         catalog.Catalog.Name,
				Capabilities: []string{},
			},
		},
	}

	slz, err := vcd.client.CreateSolutionLandingZone(slzCfg)
	check.Assert(err, IsNil)
	check.Assert(slz, NotNil)

	AddToCleanupListOpenApi(slz.DefinedEntity.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+slz.DefinedEntity.DefinedEntity.ID)

	err = slz.Refresh()
	check.Assert(err, IsNil)

	// Get all
	allEntries, err := vcd.client.GetAllSolutionLandingZones(nil)
	check.Assert(err, IsNil)

	check.Assert(len(allEntries), Equals, 1)
	check.Assert(allEntries[0].RdeId(), Equals, slz.RdeId())

	// Get by ID
	slzById, err := vcd.client.GetSolutionLandingZoneById(slz.RdeId())
	check.Assert(err, IsNil)
	check.Assert(slzById.RdeId(), Equals, slz.RdeId())

	// Get exactly one
	slzSingle, err := vcd.client.GetExactlyOneSolutionLandingZone()
	check.Assert(err, IsNil)
	check.Assert(slzSingle.RdeId(), Equals, slz.RdeId())

	// Update
	// Lookup one more Org network and add it
	orgNetwork2, err := vcd.nsxtVdc.GetOpenApiOrgVdcNetworkByName(vcd.config.VCD.Nsxt.IsolatedNetwork)
	check.Assert(err, IsNil)
	check.Assert(orgNetwork2, NotNil)

	slzCfg.Vdcs[0].Networks = append(slzCfg.Vdcs[0].Networks, types.SolutionLandingZoneVdcChild{
		ID:           orgNetwork2.OpenApiOrgVdcNetwork.ID,
		Name:         orgNetwork2.OpenApiOrgVdcNetwork.Name,
		IsDefault:    false,
		Capabilities: []string{},
	})

	updatedSlz, err := slz.Update(slzCfg)
	check.Assert(err, IsNil)
	check.Assert(len(updatedSlz.SolutionLandingZoneType.Vdcs[0].Networks), Equals, 2)

	err = slz.Delete()
	check.Assert(err, IsNil)

	// Check that no entry exists
	slzByIdErr, err := vcd.client.GetSolutionLandingZoneById(slz.RdeId())
	check.Assert(err, NotNil)
	check.Assert(slzByIdErr, IsNil)
}

func createSlz(vcd *TestVCD, check *C) *SolutionLandingZone {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCById(vcd.nsxtVdc.Vdc.ID, false)
	check.Assert(err, IsNil)
	orgNetwork, err := vcd.nsxtVdc.GetOpenApiOrgVdcNetworkByName(vcd.config.VCD.Nsxt.RoutedNetwork)
	check.Assert(err, IsNil)
	check.Assert(orgNetwork, NotNil)
	computePolicy, err := adminVdc.GetAllAssignedVdcComputePoliciesV2(nil)
	check.Assert(err, IsNil)
	check.Assert(computePolicy, NotNil)
	storageProfileRef, err := adminVdc.GetDefaultStorageProfileReference()
	check.Assert(err, IsNil)
	check.Assert(storageProfileRef, NotNil)
	catalog, err := adminOrg.GetCatalogByName(vcd.config.VCD.Catalog.NsxtBackedCatalogName, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	slzCfg := &types.SolutionLandingZoneType{
		Name: adminOrg.AdminOrg.Name,
		ID:   adminOrg.AdminOrg.ID,
		Vdcs: []types.SolutionLandingZoneVdc{
			{
				ID:           adminVdc.AdminVdc.ID,
				Name:         adminVdc.AdminVdc.Name,
				Capabilities: []string{},
				Networks: []types.SolutionLandingZoneVdcChild{
					{
						ID:           orgNetwork.OpenApiOrgVdcNetwork.ID,
						Name:         orgNetwork.OpenApiOrgVdcNetwork.Name,
						IsDefault:    true,
						Capabilities: []string{},
					},
				},
				ComputePolicies: []types.SolutionLandingZoneVdcChild{
					{
						ID:           computePolicy[0].VdcComputePolicyV2.ID,
						Name:         computePolicy[0].VdcComputePolicyV2.Name,
						IsDefault:    true,
						Capabilities: []string{},
					},
				},
				StoragePolicies: []types.SolutionLandingZoneVdcChild{
					{
						ID:           storageProfileRef.ID,
						Name:         storageProfileRef.Name,
						IsDefault:    true,
						Capabilities: []string{},
					},
				},
			},
		},
		Catalogs: []types.SolutionLandingZoneCatalog{
			{
				ID:           catalog.Catalog.ID,
				Name:         catalog.Catalog.Name,
				Capabilities: []string{},
			},
		},
	}
	slz, err := vcd.client.CreateSolutionLandingZone(slzCfg)
	check.Assert(err, IsNil)
	check.Assert(slz, NotNil)

	AddToCleanupListOpenApi(slz.DefinedEntity.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+slz.DefinedEntity.DefinedEntity.ID)

	return slz
}
