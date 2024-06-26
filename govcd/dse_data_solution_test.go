//go:build slz || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"slices"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_Dse attempts to perform a lot of checks for code in one function because it is quite expensive
// to establish a Solution Add-On (measured to roughly 30mins)
func (vcd *TestVCD) Test_Dse(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("Solution Landing Zones are supported in VCD 10.4.1+")
	}

	if vcd.config.SolutionAddOn.Org == "" || vcd.config.SolutionAddOn.Catalog == "" || len(vcd.config.SolutionAddOn.DseSolutions) < 1 {
		check.Skip("DSE configuration is not present")
	}

	// Prerequisites
	// slz, addOn, addOnInstance := createDseAddonInstanceAndPublish(vcd, check)

	// defer func() {
	// 	_, err := addOnInstance.Publishing(nil, false)
	// 	check.Assert(err, IsNil)

	// 	_, err = addOnInstance.Delete()
	// 	check.Assert(err, IsNil)

	// 	err = addOn.Delete()
	// 	check.Assert(err, IsNil)

	// 	err = slz.Delete()
	// 	check.Assert(err, IsNil)
	// }()

	// _, err = vcd.client.GetSolutionAddonInstanceByName("TestAccSolutionAddonInstanceAndPublishing")
	// check.Assert(err, IsNil)

	dsNames := make([]string, 0)
	for dsName := range vcd.config.SolutionAddOn.DseSolutions {
		dsNames = append(dsNames, dsName)
	}

	// Lookup testing
	allDataSolutions, err := vcd.client.GetAllDataSolutions(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allDataSolutions), Equals, len(dsNames)+1) // +1 because of default "VCD Data Solutions"

	for _, ds := range allDataSolutions {
		printVerbose("# Checking Data Solution '%s' retrieval methods\n", ds.Name())
		if ds.Name() != defaultDsoName {
			check.Assert(slices.Contains(dsNames, ds.Name()), Equals, true)
		}
		check.Assert(strings.HasPrefix(ds.RdeId(), "urn:vcloud:entity:vmware:dsConfig:"), Equals, true)

		byId, err := vcd.client.GetDataSolutionById(ds.RdeId())
		check.Assert(err, IsNil)
		check.Assert(byId.DataSolution, DeepEquals, ds.DataSolution)

		byName, err := vcd.client.GetDataSolutionByName(ds.Name())
		check.Assert(err, IsNil)
		check.Assert(byName.DataSolution, DeepEquals, ds.DataSolution)
	}

	// Configure all Data Solutions except DSO
	for dsName, dsConfig := range vcd.config.SolutionAddOn.DseSolutions {
		printVerbose("# Configuring Data Solution '%s'\n", dsName)

		byName, err := vcd.client.GetDataSolutionByName(dsName)
		check.Assert(err, IsNil)

		cfg := byName.DataSolution

		if value, ok := dsConfig["chart_repository"]; ok {
			cfg.Spec.Artifacts[0]["chartRepository"] = value
		}
		if value, ok := dsConfig["version"]; ok {
			cfg.Spec.Artifacts[0]["version"] = value
		}
		if value, ok := dsConfig["package_name"]; ok {
			cfg.Spec.Artifacts[0]["packageName"] = value
		}

		if value, ok := dsConfig["package_repository"]; ok {
			cfg.Spec.Artifacts[0]["image"] = value
		}

		updatedDs, err := byName.Update(cfg)
		check.Assert(err, IsNil)

		if updatedDs.DefinedEntity.State() != "RESOLVED" {
			err = updatedDs.DefinedEntity.Resolve()
			check.Assert(err, IsNil)
		}
	}

	// Configure DSO
	printVerbose("# Configuring Default Data Solution '%s'\n", defaultDsoName)
	dsoByName, err := vcd.client.GetDataSolutionByName(defaultDsoName)
	check.Assert(err, IsNil)

	// Simulate using default values, but also configure registry
	cfg := dsoByName.DataSolution
	artifacts := cfg.Spec.Artifacts[0]

	if artifacts["defaultImage"] != nil {
		cfg.Spec.Artifacts[0]["image"] = artifacts["defaultImage"].(string)
	}

	if artifacts["defaultChartRepository"] != nil {
		cfg.Spec.Artifacts[0]["chartRepository"] = artifacts["defaultChartRepository"].(string)
	}
	if artifacts["defaultVersion"] != nil {
		cfg.Spec.Artifacts[0]["version"] = artifacts["defaultVersion"].(string)
	}

	if artifacts["defaultPackageName"] != nil {
		cfg.Spec.Artifacts[0]["packageName"] = artifacts["defaultPackageName"].(string)
	}

	auths := make(map[string]types.DseDockerAuth)
	auths[check.TestName()+"1"] = types.DseDockerAuth{Username: "user1", Password: "pass1", Description: "Test 1"}
	auths[check.TestName()+"2"] = types.DseDockerAuth{Username: "user2", Password: "pass2", Description: "Test 2"}
	cfg.Spec.DockerConfig = &types.DseDockerConfig{Auths: auths}

	updatedDs, err := dsoByName.Update(cfg)
	check.Assert(err, IsNil)

	if updatedDs.DefinedEntity.State() != "RESOLVED" {
		err = updatedDs.DefinedEntity.Resolve()
		check.Assert(err, IsNil)
	}

	// Publish to tenant
	recipientOrg, err := vcd.client.GetOrgByName(vcd.config.Cse.TenantOrg)
	check.Assert(err, IsNil)

	for dsName := range vcd.config.SolutionAddOn.DseSolutions {
		printVerbose("# Publishing Data Solution '%s' to tenant %s\n", dsName, recipientOrg.Org.Name)

		ds, err := vcd.client.GetDataSolutionByName(dsName)
		check.Assert(err, IsNil)

		dsAcl, dsoAcl, templateAcls, err := ds.Publish(recipientOrg.Org.ID)
		check.Assert(err, IsNil)
		check.Assert(dsAcl, NotNil)
		check.Assert(dsoAcl, NotNil)
		check.Assert(templateAcls, NotNil)
		check.Assert(len(templateAcls) > 1, Equals, true)

		printVerbose("# Unpublishing Data Solution '%s'\n", dsName)
		err = ds.Unpublish(recipientOrg.Org.ID)
		check.Assert(err, IsNil)
	}

}

func createDseAddonInstanceAndPublish(vcd *TestVCD, check *C) (*SolutionLandingZone, *SolutionAddOn, *SolutionAddOnInstance) {
	slz, addOn := createSlzAddOn(vcd, check)

	inputs := make(map[string]interface{})
	inputs["name"] = check.TestName()
	inputs["input-delete-previous-uiplugin-versions"] = false

	addOnInstance, _, err := addOn.CreateSolutionAddOnInstance(inputs)
	check.Assert(err, IsNil)

	PrependToCleanupListOpenApi(addOnInstance.DefinedEntity.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+addOnInstance.DefinedEntity.DefinedEntity.ID)

	scope := []string{vcd.config.Cse.TenantOrg}
	_, err = addOnInstance.Publishing(scope, false)
	check.Assert(err, IsNil)

	return slz, addOn, addOnInstance
}
