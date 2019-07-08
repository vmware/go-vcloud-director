// +build lb lbAppProfile functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LBAppProfile tests CRUD methods for load balancer application profile.
// The following things are tested if prerequisite Edge Gateway exists:
// Creation of load balancer application profile
// Read load balancer application profile by both ID and Name (application profile name must be unique in single edge gateway)
// Update - change a single field and compare that configuration and result objects are deeply equal
// Update - try and fail to update without mandatory field
// Delete
func (vcd *TestVCD) Test_LBAppProfile(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	// Used for creating
	lbAppProfileConfig := &types.LBAppProfile{
		Name: TestLBAppProfile,
		Persistence: &types.LBAppProfilePersistence{
			Method: "sourceip",
			Expire: 13,
		},
		Template: "HTTPS",
	}

	createdLbAppProfile, err := edge.CreateLBAppProfile(lbAppProfileConfig)
	check.Assert(err, IsNil)
	check.Assert(createdLbAppProfile.ID, Not(IsNil))

	// We created application profile successfully therefore let's add it to cleanup list
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(TestLBAppProfile, "lbAppProfile", parentEntity, check.TestName())

	// Lookup by both name and ID and compare that these are equal values
	lbAppProfileByID, err := edge.ReadLBAppProfile(&types.LBAppProfile{ID: createdLbAppProfile.ID})
	check.Assert(err, IsNil)

	lbPoolByName, err := edge.ReadLBAppProfile(&types.LBAppProfile{Name: createdLbAppProfile.Name})
	check.Assert(err, IsNil)
	check.Assert(createdLbAppProfile.ID, Equals, lbPoolByName.ID)
	check.Assert(lbAppProfileByID.ID, Equals, lbPoolByName.ID)
	check.Assert(lbAppProfileByID.Name, Equals, lbPoolByName.Name)
	check.Assert(lbAppProfileByID.Persistence.Expire, Equals, lbPoolByName.Persistence.Expire)

	check.Assert(createdLbAppProfile.Template, Equals, lbAppProfileConfig.Template)

	// Test updating fields
	// Update persistence method
	lbAppProfileByID.Persistence.Method = "sourceip"
	updatedAppProfile, err := edge.UpdateLBAppProfile(lbAppProfileByID)
	check.Assert(err, IsNil)
	check.Assert(updatedAppProfile.Persistence.Method, Equals, lbAppProfileByID.Persistence.Method)

	// Verify that updated application profile and its configuration are identical
	check.Assert(updatedAppProfile, DeepEquals, lbAppProfileByID)

	// Try to set invalid algorithm hash and expect API to return error
	// Invalid persistence method invalid_method. Valid methods are: COOKIE|SSL-SESSIONID|SOURCEIP.
	lbAppProfileByID.Persistence.Method = "invalid_method"
	updatedAppProfile, err = edge.UpdateLBAppProfile(lbAppProfileByID)
	check.Assert(err, ErrorMatches, ".*Invalid persistence method .*Valid methods are:.*")

	// Update should fail without name
	lbAppProfileByID.Name = ""
	_, err = edge.UpdateLBAppProfile(lbAppProfileByID)
	check.Assert(err.Error(), Equals, "load balancer application profile Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLBAppProfile(&types.LBAppProfile{ID: createdLbAppProfile.ID})
	check.Assert(err, IsNil)
}
