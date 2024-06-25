//go:build slz || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_SolutionAddOnInstanceAndPublishing(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("Solution Landing Zones are supported in VCD 10.4.1+")
	}

	if vcd.config.SolutionAddOn.Org == "" || vcd.config.SolutionAddOn.Catalog == "" {
		check.Skip("Solution Add-On configuration is not present")
	}

	// Prerequisites
	slz, addOn := createSlzAddOn(vcd, check)

	// Create Solution Add-On Instance
	inputs := make(map[string]interface{})
	inputs["name"] = check.TestName()
	inputs["input-delete-previous-uiplugin-versions"] = false

	addOnInstance, res, err := addOn.CreateSolutionAddOnInstance(inputs)
	check.Assert(err, IsNil)
	check.Assert(addOnInstance, NotNil)
	check.Assert(res, Not(Equals), "")

	// Get by Id
	addOnInstanceByName, err := vcd.client.GetSolutionAddOnInstanceById(addOnInstance.RdeId())
	check.Assert(err, IsNil)
	check.Assert(addOnInstanceByName.SolutionAddOnInstance.Name, Equals, addOnInstance.SolutionAddOnInstance.Name)

	// Get all by Name
	allAddOnInstancesByName, err := vcd.client.GetAllSolutionAddonInstancesByName(addOnInstance.SolutionAddOnInstance.Name)
	check.Assert(err, IsNil)
	check.Assert(len(allAddOnInstancesByName), Equals, 1)
	check.Assert(allAddOnInstancesByName[0].SolutionAddOnInstance.Name, Equals, addOnInstance.SolutionAddOnInstance.Name)

	// Get all instances of a specific Add-On
	allAddOnChildren, err := addOn.GetAllInstances()
	check.Assert(err, IsNil)
	check.Assert(len(allAddOnChildren), Equals, 1)
	check.Assert(allAddOnChildren[0].SolutionAddOnInstance.Name, Equals, addOnInstance.SolutionAddOnInstance.Name)

	// Get child instance by name
	addOnChildByName, err := addOn.GetInstanceByName(addOnInstance.SolutionAddOnInstance.Name)
	check.Assert(err, IsNil)
	check.Assert(addOnChildByName.RdeId(), Equals, addOnInstance.RdeId())

	// Get parent Solution Add-On
	parentSolutionAddOn, err := addOnInstance.GetParentSolutionAddOn()
	check.Assert(err, IsNil)
	check.Assert(parentSolutionAddOn.RdeId(), Equals, addOn.RdeId())

	// Publish Solution Add-On Instance
	scope := []string{vcd.config.Cse.TenantOrg}
	_, err = addOnInstance.Publishing(scope, false)
	check.Assert(err, IsNil)

	// Unpublish Solution Add-On Instance
	_, err = addOnInstance.Publishing(nil, false)
	check.Assert(err, IsNil)

	// Delete Solution Add-On Instance
	deleteInputs := make(map[string]interface{})
	deleteInputs["name"] = addOnInstance.SolutionAddOnInstance.AddonInstanceSolutionName
	deleteInputs["input-force-delete"] = true
	res2, err := addOnInstance.Delete(deleteInputs)
	check.Assert(err, IsNil)
	check.Assert(res2, Not(Equals), "")

	// Cleanup
	err = addOn.Delete()
	check.Assert(err, IsNil)

	err = slz.Delete()
	check.Assert(err, IsNil)
}

// createSlzAddOn depends on CSE build (having vcd.config.SolutionAddOn configuration present)
func createSlzAddOn(vcd *TestVCD, check *C) (*SolutionLandingZone, *SolutionAddOn) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.SolutionAddOn.Org)
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.config.SolutionAddOn.Vdc, false)
	check.Assert(err, IsNil)
	vdc, err := adminOrg.GetVDCByName(vcd.config.SolutionAddOn.Vdc, false)
	check.Assert(err, IsNil)

	orgNetwork, err := vdc.GetOpenApiOrgVdcNetworkByName(vcd.config.SolutionAddOn.RoutedNetwork)
	check.Assert(err, IsNil)
	check.Assert(orgNetwork, NotNil)
	computePolicy, err := adminVdc.GetAllAssignedVdcComputePoliciesV2(nil)
	check.Assert(err, IsNil)
	check.Assert(computePolicy, NotNil)
	storageProfileRef, err := adminVdc.GetDefaultStorageProfileReference()
	check.Assert(err, IsNil)
	check.Assert(storageProfileRef, NotNil)
	catalog, err := adminOrg.GetCatalogByName(vcd.config.SolutionAddOn.Catalog, false)
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

	cacheFilePath, err := fetchCacheFile(catalog, vcd.config.SolutionAddOn.AddonImageDse, check)
	check.Assert(err, IsNil)

	catItem, err := catalog.GetCatalogItemByName(vcd.config.SolutionAddOn.AddonImageDse, false)
	check.Assert(err, IsNil)

	createCfg := SolutionAddOnConfig{
		IsoFilePath:          cacheFilePath,
		User:                 "administrator",
		CatalogItemId:        catItem.CatalogItem.ID,
		AutoTrustCertificate: true,
	}
	solutionAddOn, err := vcd.client.CreateSolutionAddOn(createCfg)
	check.Assert(err, IsNil)
	PrependToCleanupListOpenApi(solutionAddOn.DefinedEntity.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+solutionAddOn.DefinedEntity.DefinedEntity.ID)

	return slz, solutionAddOn
}
