//go:build functional || openapi || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
	"strings"
)

// Test_ExternalEndpoint tests the CRUD operations for the External Endpoints.
func (vcd *TestVCD) Test_ExternalEndpoint(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	name := strings.ReplaceAll(check.TestName(), ".", "") // Special characters like "." are not allowed

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

	retrievedEp, err := vcd.client.GetExternalEndpointById(createdEp.ExternalEndpoint.ID)
	check.Assert(err, IsNil)
	check.Assert(retrievedEp, NotNil)
	check.Assert(retrievedEp.ExternalEndpoint.Name, Equals, ep.Name)
	check.Assert(retrievedEp.ExternalEndpoint.Enabled, Equals, ep.Enabled)
	check.Assert(retrievedEp.ExternalEndpoint.Description, Equals, ep.Description)
	check.Assert(retrievedEp.ExternalEndpoint.Vendor, Equals, ep.Vendor)
	check.Assert(retrievedEp.ExternalEndpoint.Version, Equals, ep.Version)
	check.Assert(retrievedEp.ExternalEndpoint.RootUrl, Equals, ep.RootUrl)

	retrievedEp2, err := vcd.client.GetExternalEndpoint(retrievedEp.ExternalEndpoint.Vendor, retrievedEp.ExternalEndpoint.Name, retrievedEp.ExternalEndpoint.Version)
	check.Assert(err, IsNil)
	check.Assert(retrievedEp2, NotNil)
	check.Assert(*retrievedEp2.ExternalEndpoint, DeepEquals, *retrievedEp.ExternalEndpoint)

	allEps, err := vcd.client.GetAllExternalEndpoints(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allEps) > 0, Equals, true)
	pos := -1
	for i, ep := range allEps {
		if ep.ExternalEndpoint.ID == retrievedEp.ExternalEndpoint.ID {
			pos = i
		}
	}
	check.Assert(pos > -1, Equals, true)
	check.Assert(*allEps[pos].ExternalEndpoint, DeepEquals, *retrievedEp.ExternalEndpoint)

	ep.RootUrl = "https://www.broadcom.com/updated"
	err = createdEp.Update(ep)
	check.Assert(err, IsNil)
	check.Assert(createdEp, NotNil)
	check.Assert(createdEp.ExternalEndpoint.RootUrl, Equals, "https://www.broadcom.com/updated")
}
