//go:build slz || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"os"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_SolutionAddOn(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("Solution Landing Zones are supported in VCD 10.4.1+")
	}

	if vcd.config.VCD.Catalog.NsxtCatalogAddonDse == "" {
		check.Skip("missing 'VCD.Catalog.NsxtCatalogAddonDse' value")
	}
	catalogMediaName := vcd.config.VCD.Catalog.NsxtCatalogAddonDse

	slz := createSlz(vcd, check)
	err := slz.Refresh()
	check.Assert(err, IsNil)

	addOnCatalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.NsxtBackedCatalogName, false)
	check.Assert(err, IsNil)

	cacheFilePath, err := fetchCacheFile(addOnCatalog, catalogMediaName, check)
	check.Assert(err, IsNil)

	org, err := vcd.client.GetOrgById(slz.SolutionLandingZoneType.ID)
	check.Assert(err, IsNil)
	catalog, err := org.GetCatalogById(slz.SolutionLandingZoneType.Catalogs[0].ID, false)
	check.Assert(err, IsNil)
	catItem, err := catalog.GetCatalogItemByName(catalogMediaName, false)
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

	// Get all
	allSolutionAddOns, err := vcd.client.GetAllSolutionAddons(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allSolutionAddOns) >= 1, Equals, true) // VCD has a few baked in Solution Add-Ons (e.g. 'service-account-solutions-system-user', 'vmware.solution-addon-landing-zone-1.0.0')

	foundId := false
	for addOnIndex := range allSolutionAddOns {
		if allSolutionAddOns[addOnIndex].RdeId() == solutionAddOn.RdeId() {
			foundId = true
			break
		}
	}
	check.Assert(foundId, Equals, true)

	// Get all with filter
	queryParams := queryParameterFilterAnd("id=="+solutionAddOn.RdeId(), nil)
	filteredSolutionAddon, err := vcd.client.GetAllSolutionAddons(queryParams)
	check.Assert(err, IsNil)
	check.Assert(len(filteredSolutionAddon), Equals, 1)

	// By ID
	sao, err := vcd.client.GetSolutionAddonById(solutionAddOn.RdeId())
	check.Assert(err, IsNil)
	check.Assert(sao.RdeId(), Equals, solutionAddOn.RdeId())

	// By Name
	saoByName, err := vcd.client.GetSolutionAddonByName(solutionAddOn.DefinedEntity.DefinedEntity.Name)
	check.Assert(err, IsNil)
	check.Assert(saoByName.RdeId(), Equals, solutionAddOn.RdeId())

	// Update - pointing at wrong catalog image
	catItemPhoton, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.NsxtCatalogItem, false)
	check.Assert(err, IsNil)

	solutionAddOnUpdate := saoByName.SolutionAddOnEntity
	solutionAddOnUpdate.Origin.CatalogItemId = catItemPhoton.CatalogItem.ID

	updatedSao, err := sao.Update(solutionAddOnUpdate)
	check.Assert(err, IsNil)
	check.Assert(updatedSao.RdeId(), Equals, sao.RdeId())

	// Delete
	err = sao.Delete()
	check.Assert(err, IsNil)

	// Verify no more Add-Ons remaining
	allSolutionAddOnsAfterCleanup, err := vcd.client.GetAllSolutionAddons(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allSolutionAddOnsAfterCleanup), Equals, len(allSolutionAddOns)-1)

	err = slz.Delete()
	check.Assert(err, IsNil)
}

func fetchCacheFile(catalog *Catalog, fileName string, check *C) (string, error) {
	pwd, err := os.Getwd()
	check.Assert(err, IsNil)
	cacheDirPath := pwd + "/test-resources/cache"
	cacheFilePath := cacheDirPath + "/" + fileName
	printVerbose("# Using '%s' file to cache Solution Add-On\n", cacheFilePath)

	if _, err := os.Stat(cacheFilePath); errors.Is(err, os.ErrNotExist) || !dirExists(cacheDirPath) {
		// Create cache directory if it doesn't exist
		if fileInfo, err := os.Stat(cacheDirPath); os.IsNotExist(err) || !fileInfo.IsDir() {
			// test-resources/cache is a file, not a directory, it should be removed
			if !os.IsNotExist(err) && !fileInfo.IsDir() {
				fmt.Printf("# %s is a file, not a directory - removing\n", cacheDirPath)
				err := os.Remove(cacheDirPath)
				check.Assert(err, IsNil)
			}

			printVerbose("# Creating directory '%s'\n", cacheDirPath)
			err := os.Mkdir(cacheDirPath, 0750)
			check.Assert(err, IsNil)
		}

		fmt.Printf("# Downloading Solution Add-On '%s' from VCD...", fileName)
		addOnMediaItem, err := catalog.GetMediaByName(fileName, false)
		check.Assert(err, IsNil)

		addOn, err := addOnMediaItem.Download()
		check.Assert(err, IsNil)

		err = os.WriteFile(cacheFilePath, addOn, 0600)
		check.Assert(err, IsNil)
		addOn = nil // free memory
		fmt.Println("Done")
	} else {
		printVerbose("# File '%s' is present, not downloading\n", cacheFilePath)
	}

	return cacheFilePath, nil
}

// Checks if a directory exists
func dirExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	fileMode := f.Mode()
	return fileMode.IsDir()
}
