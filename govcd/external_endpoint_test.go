//go:build functional || openapi || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_ExternalEndpoint tests the CRUD operations for the External Endpoints.
func (vcd *TestVCD) Test_ExternalEndpoint(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointExternalEndpoints)

	ep := types.ExternalEndpoint{
		Name:        check.TestName(),
		Version:     "1.0.0",
		Vendor:      "vmware",
		Enabled:     true,
		Description: check.TestName(),
		RootUrl:     "https://www.broadcom.com",
	}

	createdEp, err := vcd.client.CreateExternalEndpoint(&ep)
	check.Assert(err, IsNil)
	check.Assert(createdEp, NotNil)

	defer func() {
		err = createdEp.Delete()
		check.Assert(err, IsNil)
	}()

	retrievedEp, err := vcd.client.GetExternalEndpointById(createdEp.ExternalEndpoint.ID)
	check.Assert(err, IsNil)
	check.Assert(retrievedEp, NotNil)

}
