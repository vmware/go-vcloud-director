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

	name := strings.ReplaceAll(check.TestName(), ".", "")

	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointApiFilters)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointExternalEndpoints)

	ep := types.ExternalEndpoint{
		Name:        name,
		Version:     "1.0.0",
		Vendor:      "vmware",
		Enabled:     true,
		Description: "Description of" + name,
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

	createdAf, err := vcd.client.CreateApiFilter(&types.ApiFilter{
		ExternalSystem: &types.OpenApiReference{
			Name: ep.Name,
			ID:   ep.ID,
		},
		UrlMatcher: &types.UrlMatcher{
			UrlPattern: "/custom/.",
			UrlScope:   "EXT_API",
		},
	})
	check.Assert(err, IsNil)
	check.Assert(createdAf, NotNil)

	defer func() {
		err = createdAf.Delete()
		check.Assert(err, IsNil)
	}()
}
