// +build lb lbAppProfile functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LBAppProfile tests CRUD methods for load balancer application profile.
// The following things are tested if prerequisite Edge Gateway exists:
// 1. Creation of load balancer application profile
// 2. Get load balancer application profile by both ID and Name (application profile name must be unique in single edge gateway)
// 3. Update - change a single field and compare that configuration and result objects are deeply equal
// 4. Update - try and fail to update without mandatory field
// 5. Delete
func (vcd *TestVCD) Test_LBAppProfile(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	// Used for creating
	lbAppProfileConfig := &types.LbAppProfile{
		Name: TestLbAppProfile,
		Persistence: &types.LbAppProfilePersistence{
			Method: "sourceip",
			Expire: 13,
		},
		Template:                      "https",
		SslPassthrough:                false,
		InsertXForwardedForHttpHeader: false,
		ServerSslEnabled:              false,
	}

	err = deleteLbAppProfileIfExists(edge, lbAppProfileConfig.Name)
	check.Assert(err, IsNil)
	createdLbAppProfile, err := edge.CreateLbAppProfile(lbAppProfileConfig)
	check.Assert(err, IsNil)
	check.Assert(createdLbAppProfile.ID, Not(IsNil))
	check.Assert(createdLbAppProfile.Persistence.Method, Equals, lbAppProfileConfig.Persistence.Method)
	check.Assert(createdLbAppProfile.Template, Equals, lbAppProfileConfig.Template)
	check.Assert(createdLbAppProfile.Persistence.Expire, Equals, lbAppProfileConfig.Persistence.Expire)
	check.Assert(createdLbAppProfile.SslPassthrough, Equals, lbAppProfileConfig.SslPassthrough)
	check.Assert(createdLbAppProfile.InsertXForwardedForHttpHeader, Equals, lbAppProfileConfig.InsertXForwardedForHttpHeader)
	check.Assert(createdLbAppProfile.ServerSslEnabled, Equals, lbAppProfileConfig.ServerSslEnabled)

	// We created application profile successfully therefore let's add it to cleanup list
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(TestLbAppProfile, "lbAppProfile", parentEntity, check.TestName())

	// Lookup by both name and ID and compare that these are equal values
	lbAppProfileByID, err := edge.GetLbAppProfileById(createdLbAppProfile.ID)
	check.Assert(err, IsNil)
	check.Assert(lbAppProfileByID, Not(IsNil))

	lbAppProfileByName, err := edge.GetLbAppProfileByName(createdLbAppProfile.Name)
	check.Assert(err, IsNil)
	check.Assert(lbAppProfileByName, Not(IsNil))
	check.Assert(createdLbAppProfile.ID, Equals, lbAppProfileByName.ID)
	check.Assert(lbAppProfileByID.ID, Equals, lbAppProfileByName.ID)
	check.Assert(lbAppProfileByID.Name, Equals, lbAppProfileByName.Name)
	check.Assert(lbAppProfileByID.Persistence.Expire, Equals, lbAppProfileByName.Persistence.Expire)

	check.Assert(createdLbAppProfile.Template, Equals, lbAppProfileConfig.Template)

	// Test updating fields
	// Update persistence method
	lbAppProfileByID.Persistence.Method = "sourceip"
	updatedAppProfile, err := edge.UpdateLbAppProfile(lbAppProfileByID)
	check.Assert(err, IsNil)
	check.Assert(updatedAppProfile.Persistence.Method, Equals, lbAppProfileByID.Persistence.Method)

	// Update boolean value fields
	lbAppProfileByID.SslPassthrough = true
	lbAppProfileByID.InsertXForwardedForHttpHeader = true
	lbAppProfileByID.ServerSslEnabled = true
	updatedAppProfile, err = edge.UpdateLbAppProfile(lbAppProfileByID)
	check.Assert(err, IsNil)
	check.Assert(updatedAppProfile.SslPassthrough, Equals, lbAppProfileByID.SslPassthrough)
	check.Assert(updatedAppProfile.InsertXForwardedForHttpHeader, Equals, lbAppProfileByID.InsertXForwardedForHttpHeader)
	check.Assert(updatedAppProfile.ServerSslEnabled, Equals, lbAppProfileByID.ServerSslEnabled)

	// Verify that updated application profile and its configuration are identical
	check.Assert(updatedAppProfile, DeepEquals, lbAppProfileByID)

	// Try to set invalid algorithm hash and expect API to return error
	// Invalid persistence method invalid_method. Valid methods are: COOKIE|SSL-SESSIONID|SOURCEIP.
	lbAppProfileByID.Persistence.Method = "invalid_method"
	updatedAppProfile, err = edge.UpdateLbAppProfile(lbAppProfileByID)
	check.Assert(err, ErrorMatches, ".*Invalid persistence method .*Valid methods are:.*")

	// Update should fail without name
	lbAppProfileByID.Name = ""
	_, err = edge.UpdateLbAppProfile(lbAppProfileByID)
	check.Assert(err.Error(), Equals, "load balancer application profile Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLbAppProfile(&types.LbAppProfile{ID: createdLbAppProfile.ID})
	check.Assert(err, IsNil)

	// Ensure it is deleted
	_, err = edge.GetLbAppProfileById(createdLbAppProfile.ID)
	check.Assert(IsNotFound(err), Equals, true)
}

// deleteLbAppProfileIfExists is used to cleanup before creation of component. It returns error only if there was
// other error than govcd.ErrorEntityNotFound
func deleteLbAppProfileIfExists(edge EdgeGateway, name string) error {
	err := edge.DeleteLbAppProfileByName(name)
	if err != nil && !ContainsNotFound(err) {
		return err
	}
	if err != nil && ContainsNotFound(err) {
		return nil
	}

	fmt.Printf("# Removed leftover LB app profile '%s'\n", name)
	return nil
}
