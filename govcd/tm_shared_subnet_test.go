//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"strings"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmSharedSubnet(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()

	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	// Transforms the test name into a K8s compliant name
	k8sCompliantName := strings.ReplaceAll(strings.Split(strings.ToLower(check.TestName()), ".")[1], "_", "-")

	sharedSubnetType := &types.TmSharedSubnet{
		Name:        k8sCompliantName,
		Description: check.TestName(),
		SubnetType:  "VLAN",
		VlanId:      100,
		GatewayCidr: "10.0.0.1/24",
		RegionRef:   types.OpenApiReference{ID: region.Region.ID},
	}

	createdSharedSubnet, err := vcd.client.CreateTmSharedSubnet(sharedSubnetType)
	check.Assert(err, IsNil)
	check.Assert(createdSharedSubnet, NotNil)
	// Add to cleanup list
	PrependToCleanupListOpenApi(createdSharedSubnet.TmSharedSubnet.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmSharedSubnets+createdSharedSubnet.TmSharedSubnet.ID)

	// Get TM Shared Subnet By Name
	byName, err := vcd.client.GetTmSharedSubnetByName(sharedSubnetType.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.TmSharedSubnet, DeepEquals, createdSharedSubnet.TmSharedSubnet)

	// Get TM Shared Subnet By Id
	byId, err := vcd.client.GetTmSharedSubnetById(createdSharedSubnet.TmSharedSubnet.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmSharedSubnet, DeepEquals, createdSharedSubnet.TmSharedSubnet)

	// Get TM Shared Subnet By Name and Region ID
	byNameAndRegionId, err := vcd.client.GetTmSharedSubnetByNameAndRegionId(createdSharedSubnet.TmSharedSubnet.Name, region.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(byNameAndRegionId.TmSharedSubnet, DeepEquals, createdSharedSubnet.TmSharedSubnet)

	// Get TM Shared Subnet By Name and Region ID in non existent Region
	byNameAndInvalidRegionId, err := vcd.client.GetTmSharedSubnetByNameAndRegionId(createdSharedSubnet.TmSharedSubnet.Name, "urn:vcloud:region:a93c9db9-0000-0000-0000-a8f7eeda85f9")
	check.Assert(err, NotNil)
	check.Assert(byNameAndInvalidRegionId, IsNil)

	// Not Found tests
	byNameInvalid, err := vcd.client.GetTmSharedSubnetByName("fake-name")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byNameInvalid, IsNil)

	byIdInvalid, err := vcd.client.GetTmSharedSubnetById("urn:vcloud:sharedSubnet:5344b964-0000-0000-0000-d554913db643")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdInvalid, IsNil)

	// Update
	createdSharedSubnet.TmSharedSubnet.Name = k8sCompliantName + "-update"
	updatedVdc, err := createdSharedSubnet.Update(createdSharedSubnet.TmSharedSubnet)
	check.Assert(err, IsNil)
	check.Assert(updatedVdc.TmSharedSubnet, DeepEquals, createdSharedSubnet.TmSharedSubnet)

	// Delete
	err = createdSharedSubnet.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetTmSharedSubnetByName(createdSharedSubnet.TmSharedSubnet.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)

	// Create async
	task, err := vcd.client.CreateTmSharedSubnetAsync(sharedSubnetType)
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	byIdAsync, err := vcd.client.GetTmSharedSubnetById(task.Task.Owner.ID)
	check.Assert(err, IsNil)
	check.Assert(byIdAsync, NotNil)
	PrependToCleanupListOpenApi(byIdAsync.TmSharedSubnet.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmSharedSubnets+createdSharedSubnet.TmSharedSubnet.ID)

	err = byIdAsync.Delete()
	check.Assert(err, IsNil)
}
