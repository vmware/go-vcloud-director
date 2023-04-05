//go:build ALL || openapi || functional || nsxt

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtEdgeGatewayQosProfiles(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)

	// skip for API versions less than 36.2
	if vcd.client.Client.APIVCDMaxVersionIs("< 36.2") {
		check.Skip("Test_GetAllNsxtEdgeGatewayQosProfiles requires at least API v36.2 (vCD 10.3.2+)")
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
	if len(allQosProfiles) == 0 {
		check.Skip("No QoS profiles found")
	}

	// Fetch one by one based on DisplayName
	for _, profile := range allQosProfiles {
		printVerbose("# Fetching QoS profile '%s' by Name\n", profile.NsxtEdgeGatewayQosProfile.DisplayName)
		qosProfile, err := vcd.client.GetNsxtEdgeGatewayQosProfileByDisplayName(urn, profile.NsxtEdgeGatewayQosProfile.DisplayName)
		check.Assert(err, IsNil)
		check.Assert(qosProfile, NotNil)
		check.Assert(qosProfile.NsxtEdgeGatewayQosProfile.ID, Equals, profile.NsxtEdgeGatewayQosProfile.ID)
	}
}
