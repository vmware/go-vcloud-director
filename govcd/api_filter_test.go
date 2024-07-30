//go:build functional || openapi || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"strings"
)

// Test_ApiFilter tests the CRUD operations for the API Filters.
func (vcd *TestVCD) Test_ApiFilter(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	name := strings.ReplaceAll(check.TestName(), ".", "") // Special characters like "." are not allowed

	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointApiFilters)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointExternalEndpoints)

	ep := types.ExternalEndpoint{
		Name:        name,
		Version:     "1.0.0",
		Vendor:      "vmware",
		Enabled:     true,
		Description: "Description of " + name,
		RootUrl:     "https://www.broadcom.com",
	}

	createdEp, err := vcd.client.CreateExternalEndpoint(&ep)
	check.Assert(err, IsNil)
	check.Assert(createdEp, NotNil)

	defer func() {
		ep.Enabled = false // Endpoint needs to be disabled before deleting
		err = createdEp.Update(ep)
		check.Assert(err, IsNil)
		err = createdEp.Delete()
		check.Assert(err, IsNil)
	}()

	af := &types.ApiFilter{
		ExternalSystem: &types.OpenApiReference{
			Name: createdEp.ExternalEndpoint.Name,
			ID:   createdEp.ExternalEndpoint.ID,
		},
		UrlMatcher: &types.UrlMatcher{
			UrlPattern: "/custom/.*",
			UrlScope:   "EXT_API",
		},
	}

	createdAf, err := vcd.client.CreateApiFilter(af)
	check.Assert(err, IsNil)
	check.Assert(createdAf, NotNil)

	defer func() {
		err = createdAf.Delete()
		check.Assert(err, IsNil)
	}()

	retrievedAf, err := vcd.client.GetApiFilterById(createdAf.ApiFilter.ID)
	check.Assert(err, IsNil)
	check.Assert(retrievedAf, NotNil)
	check.Assert(retrievedAf.ApiFilter.ID, Equals, createdAf.ApiFilter.ID)
	check.Assert(retrievedAf.ApiFilter.UrlMatcher, NotNil)
	check.Assert(retrievedAf.ApiFilter.UrlMatcher.UrlPattern, Equals, createdAf.ApiFilter.UrlMatcher.UrlPattern)
	check.Assert(retrievedAf.ApiFilter.UrlMatcher.UrlScope, Equals, createdAf.ApiFilter.UrlMatcher.UrlScope)
	check.Assert(retrievedAf.ApiFilter.ResponseContentType, Equals, "")
	check.Assert(retrievedAf.ApiFilter.ExternalSystem, NotNil)
	check.Assert(retrievedAf.ApiFilter.ExternalSystem.ID, Equals, createdEp.ExternalEndpoint.ID)
	check.Assert(retrievedAf.ApiFilter.ExternalSystem.Name, Equals, createdEp.ExternalEndpoint.Name)

	allAfs, err := vcd.client.GetAllApiFilters(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allAfs) > 0, Equals, true)
	pos := -1
	for i, af := range allAfs {
		if af.ApiFilter.ID == retrievedAf.ApiFilter.ID {
			pos = i
		}
	}
	check.Assert(pos > -1, Equals, true)
	check.Assert(*allAfs[pos].ApiFilter, DeepEquals, *retrievedAf.ApiFilter)

	af.UrlMatcher.UrlPattern = "/custom2/.*"
	err = createdAf.Update(*af)
	check.Assert(err, IsNil)
	check.Assert(createdAf, NotNil)
	check.Assert(createdAf.ApiFilter.UrlMatcher, NotNil)
	check.Assert(createdAf.ApiFilter.UrlMatcher.UrlPattern, Equals, af.UrlMatcher.UrlPattern)
}
