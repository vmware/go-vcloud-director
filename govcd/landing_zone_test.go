//go:build rde || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_CreateLandingZone(check *C) {

	// rdeId := "urn:vcloud:entity:vmware:solutions_organization:19032ea5-3e49-44e7-b601-5e37ecbbd190"

	// definedEntity, err := vcd.client.GetRdeById(rdeId)
	// check.Assert(err, IsNil)

	// err = definedEntity.Resolve()
	// check.Assert(err, IsNil)

	// err = definedEntity.Delete()
	// check.Assert(err, IsNil)

	// check.Assert(true, IsNil)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)

	adminVdc, err := adminOrg.GetAdminVDCById(vcd.nsxtVdc.Vdc.ID, false)
	check.Assert(err, IsNil)

	// orgNetwork, err := vcd.config.VCD.Nsxt.RoutedNetwork
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

	// AddToCleanupListOpenApi(slz.DefinedEntity.DefinedEntity.ID, check.TestName(), fmt.Sprintf(types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntitiesResolve, slz.DefinedEntity.DefinedEntity.ID))
	AddToCleanupListOpenApi(slz.DefinedEntity.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+slz.DefinedEntity.DefinedEntity.ID)

	// fmt.Println("sleeping")
	// time.Sleep(2 * time.Minute)

	err = slz.Refresh()
	check.Assert(err, IsNil)

	spew.Dump(slz.SolutionLandingZoneType)
	fmt.Println("====")
	spew.Dump(slz.DefinedEntity.DefinedEntity)

	err = slz.Delete()
	check.Assert(err, IsNil)

	// isoFileName := os.Getenv("DSE_ISO") //"vmware-vcd-ds-1.3.0-22829404.iso"
	// if isoFileName == "" {
	// 	check.Skip("no .ISO defined")
	// }
	// // WIP
	// rde, err := vcd.client.Client.CreateLandingZoneRde(isoFileName, "administrator", "TBA")
	// check.Assert(err, IsNil)
	// check.Assert(rde, NotNil)

	// contents, err := getContentsFromIsoFiles(isoFileName, wantedFiles)
	// check.Assert(err, IsNil)
	// for k, v := range contents {
	// 	fmt.Printf("%-15s: %-30s %d\n", k, v.foundFileName, len(v.contents))
	// }

}
