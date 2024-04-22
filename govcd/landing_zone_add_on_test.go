//go:build rde || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"os"
	"path/filepath"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_SolutionAddOn(check *C) {
	slz := createSlz(vcd, check)
	err := slz.Refresh()
	check.Assert(err, IsNil)

	isoPath := os.Getenv("DSE_ISO") //"vmware-vcd-ds-1.3.0-22829404.iso"
	if isoPath == "" {
		check.Skip("no .ISO defined")
	}

	// Upload image
	isoFileName := filepath.Base(isoPath)

	org, err := vcd.client.GetOrgById(slz.SolutionLandingZoneType.ID)
	check.Assert(err, IsNil)
	catalog, err := org.GetCatalogById(slz.SolutionLandingZoneType.Catalogs[0].ID, false)
	check.Assert(err, IsNil)
	catItem, err := catalog.GetCatalogItemByName(isoFileName, false)
	if ContainsNotFound(err) {
		uploadPieceSize := 10
		task, err := catalog.UploadMediaFile(isoFileName, "", isoPath, int64(uploadPieceSize)*1024*1024, false) // Convert from megabytes to bytes)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)

		catItem, err = catalog.GetCatalogItemByName(isoFileName, true)
		check.Assert(err, IsNil)
	}

	solutionAddOn, err := vcd.client.CreateSolutionAddOn(isoPath, "administrator", catItem.CatalogItem.ID, true, false)
	check.Assert(err, IsNil)
	PrependToCleanupListOpenApi(solutionAddOn.DefinedEntity.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+solutionAddOn.DefinedEntity.DefinedEntity.ID)

	// Get all
	allSolutionAddOns, err := vcd.client.GetAllSolutionAddons(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allSolutionAddOns), Equals, 1)
	check.Assert(allSolutionAddOns[0].Id(), Equals, solutionAddOn.Id())

	// Get all with filter
	queryParams := queryParameterFilterAnd("id=="+solutionAddOn.Id(), nil)
	filteredSolutionAddon, err := vcd.client.GetAllSolutionAddons(queryParams)
	check.Assert(err, IsNil)
	check.Assert(len(filteredSolutionAddon), Equals, 1)

	// By ID
	sao, err := vcd.client.GetSolutionAddonById(solutionAddOn.Id())
	check.Assert(err, IsNil)
	check.Assert(sao.Id(), Equals, solutionAddOn.Id())

	// Delete
	err = sao.Delete()
	check.Assert(err, IsNil)

	// Verify no more Add-Ons remaining
	allSolutionAddOns, err = vcd.client.GetAllSolutionAddons(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allSolutionAddOns), Equals, 0)
}
