//go:build network || nsxt || functional || openapi || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtSegmentProfileTemplate(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	vcd.skipIfNotSysAdmin(check)

	nsxtManager, err := vcd.client.GetNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(nsxtManager, NotNil)

	nsxtManagerUrn, err := nsxtManager.Urn()
	check.Assert(err, IsNil)

	// Filter by NSX-T Manager
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("nsxTManagerRef.id==%s", nsxtManagerUrn), queryParams)

	// Lookup prerequisite profiles for Segment Profile template creation
	ipDiscoveryProfile, err := vcd.client.GetIpDiscoveryProfileByName(vcd.config.VCD.Nsxt.IpDiscoveryProfile, queryParams)
	check.Assert(err, IsNil)
	macDiscoveryProfile, err := vcd.client.GetMacDiscoveryProfileByName(vcd.config.VCD.Nsxt.MacDiscoveryProfile, queryParams)
	check.Assert(err, IsNil)
	spoofGuardProfile, err := vcd.client.GetSpoofGuardProfileByName(vcd.config.VCD.Nsxt.SpoofGuardProfile, queryParams)
	check.Assert(err, IsNil)
	qosProfile, err := vcd.client.GetQoSProfileByName(vcd.config.VCD.Nsxt.QosProfile, queryParams)
	check.Assert(err, IsNil)
	segmentSecurityProfile, err := vcd.client.GetSegmentSecurityProfileByName(vcd.config.VCD.Nsxt.SegmentSecurityProfile, queryParams)
	check.Assert(err, IsNil)

	config := &types.NsxtSegmentProfileTemplate{
		Name:                   check.TestName(),
		Description:            check.TestName() + "-description",
		IPDiscoveryProfile:     &types.Reference{ID: ipDiscoveryProfile.ID},
		MacDiscoveryProfile:    &types.Reference{ID: macDiscoveryProfile.ID},
		QosProfile:             &types.Reference{ID: qosProfile.ID},
		SegmentSecurityProfile: &types.Reference{ID: segmentSecurityProfile.ID},
		SpoofGuardProfile:      &types.Reference{ID: spoofGuardProfile.ID},
		SourceNsxTManagerRef:   &types.OpenApiReference{ID: nsxtManager.NsxtManager.ID},
	}

	createdSegmentProfileTemplate, err := vcd.client.CreateSegmentProfileTemplate(config)
	check.Assert(err, IsNil)
	check.Assert(createdSegmentProfileTemplate, NotNil)

	// Add to cleanup list
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates + createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID
	AddToCleanupListOpenApi(config.Name, check.TestName(), openApiEndpoint)

	// Retrieve segment profile template
	retrievedSpt, err := vcd.client.GetSegmentProfileTemplateById(createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID)
	check.Assert(err, IsNil)
	check.Assert(retrievedSpt.NsxtSegmentProfileTemplate, DeepEquals, createdSegmentProfileTemplate.NsxtSegmentProfileTemplate)

	// Get all and look for the required one
	allSpts, err := vcd.client.GetAllSegmentProfileTemplates(nil)
	check.Assert(err, IsNil)
	check.Assert(allSpts, NotNil)
	found := false
	for _, spt := range allSpts {
		if spt.NsxtSegmentProfileTemplate.ID == createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID {
			found = true
			break
		}
	}

	check.Assert(found, Equals, true)

	// Test update
	createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.Description = check.TestName() + "updated"
	updatedSegmentProfileTemplate, err := createdSegmentProfileTemplate.Update(createdSegmentProfileTemplate.NsxtSegmentProfileTemplate)
	check.Assert(err, IsNil)
	check.Assert(updatedSegmentProfileTemplate.NsxtSegmentProfileTemplate.Description, Equals, createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.Description)
	check.Assert(updatedSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID, Equals, createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID)

	// Delete
	err = createdSegmentProfileTemplate.Delete()
	check.Assert(err, IsNil)

	// Check that it returns sentinel error 'ErrorEntityNotFound' when an entity is not found
	notFoundSpt, err := vcd.client.GetSegmentProfileTemplateById(createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundSpt, IsNil)
}
