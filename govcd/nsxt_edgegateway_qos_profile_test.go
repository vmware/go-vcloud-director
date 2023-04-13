//go:build ALL || openapi || functional || nsxt

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtEdgeGatewayQosProfiles(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointQosProfiles)

	skipNoNsxtConfiguration(vcd, check)
	if vcd.config.VCD.Nsxt.GatewayQosProfile == "" {
		check.Skip("No NSX-T Edge Gateway QoS Profile configured")
	}

	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers), Equals, 1)

	uuid, err := GetUuidFromHref(nsxtManagers[0].HREF, true)
	check.Assert(err, IsNil)
	urn, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", uuid)
	check.Assert(err, IsNil)

	// Fetch all profiles
	allQosProfiles, err := vcd.client.GetAllNsxtEdgeGatewayQosProfiles(urn, nil)
	check.Assert(err, IsNil)
	check.Assert(len(allQosProfiles) > 0, Equals, true)

	// Fetch one by one based on DisplayName
	for _, profile := range allQosProfiles {
		printVerbose("# Fetching QoS profile '%s' by Name\n", profile.NsxtEdgeGatewayQosProfile.DisplayName)
		qosProfile, err := vcd.client.GetNsxtEdgeGatewayQosProfileByDisplayName(urn, profile.NsxtEdgeGatewayQosProfile.DisplayName)
		check.Assert(err, IsNil)
		check.Assert(qosProfile, NotNil)
		check.Assert(qosProfile.NsxtEdgeGatewayQosProfile.ID, Equals, profile.NsxtEdgeGatewayQosProfile.ID)
	}
}
