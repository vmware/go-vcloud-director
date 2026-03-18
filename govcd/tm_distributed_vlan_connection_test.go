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

func (vcd *TestVCD) Test_TmDistributedVlanConnection(check *C) {
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

	ipSpace, ipSpaceCleanup := createTmIpSpace(vcd, region, check, "1", "0")
	defer ipSpaceCleanup()

	// Transforms the test name into a K8s compliant name
	k8sCompliantName := strings.ReplaceAll(strings.Split(strings.ToLower(check.TestName()), ".")[1], "_", "-")

	distributedVlanConnectionType := &types.TmDistributedVlanConnection{
		Name:            k8sCompliantName,
		Description:     check.TestName(),
		VlanId:          100,
		GatewayCidr:     "10.0.0.1/24",
		RegionRef:       types.OpenApiReference{ID: region.Region.ID},
		IpSpaceRef:      &types.OpenApiReference{ID: ipSpace.TmIpSpace.ID},
		SubnetExclusive: false,
	}

	createdDistributedVlanConnection, err := vcd.client.CreateTmDistributedVlanConnection(distributedVlanConnectionType)
	check.Assert(err, IsNil)
	check.Assert(createdDistributedVlanConnection, NotNil)
	// Add to cleanup list
	PrependToCleanupListOpenApi(createdDistributedVlanConnection.TmDistributedVlanConnection.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmDistributedVlanConnections+createdDistributedVlanConnection.TmDistributedVlanConnection.ID)

	// Get TM Distributed Vlan Connection By Name
	byName, err := vcd.client.GetTmDistributedVlanConnectionByName(distributedVlanConnectionType.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.TmDistributedVlanConnection, DeepEquals, createdDistributedVlanConnection.TmDistributedVlanConnection)

	// Get TM Distributed Vlan Connection By Id
	byId, err := vcd.client.GetTmDistributedVlanConnectionById(createdDistributedVlanConnection.TmDistributedVlanConnection.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmDistributedVlanConnection, DeepEquals, createdDistributedVlanConnection.TmDistributedVlanConnection)

	// Get TM Distributed Vlan Connection By Name and Region ID
	byNameAndRegionId, err := vcd.client.GetTmDistributedVlanConnectionByNameAndRegionId(createdDistributedVlanConnection.TmDistributedVlanConnection.Name, region.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(byNameAndRegionId.TmDistributedVlanConnection, DeepEquals, createdDistributedVlanConnection.TmDistributedVlanConnection)

	// Get TM Distributed Vlan Connection By Name and Region ID in non existent Region
	byNameAndInvalidRegionId, err := vcd.client.GetTmDistributedVlanConnectionByNameAndRegionId(createdDistributedVlanConnection.TmDistributedVlanConnection.Name, "urn:vcloud:region:a93c9db9-0000-0000-0000-a8f7eeda85f9")
	check.Assert(err, NotNil)
	check.Assert(byNameAndInvalidRegionId, IsNil)

	// Not Found tests
	byNameInvalid, err := vcd.client.GetTmDistributedVlanConnectionByName("fake-name")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byNameInvalid, IsNil)

	byIdInvalid, err := vcd.client.GetTmDistributedVlanConnectionById("urn:vcloud:distributedVlanConnection:5344b964-0000-0000-0000-d554913db643")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdInvalid, IsNil)

	// Update
	createdDistributedVlanConnection.TmDistributedVlanConnection.Name = k8sCompliantName + "-update"
	updatedVdc, err := createdDistributedVlanConnection.Update(createdDistributedVlanConnection.TmDistributedVlanConnection)
	check.Assert(err, IsNil)
	check.Assert(updatedVdc.TmDistributedVlanConnection, DeepEquals, createdDistributedVlanConnection.TmDistributedVlanConnection)

	// Delete
	err = createdDistributedVlanConnection.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetTmDistributedVlanConnectionByName(createdDistributedVlanConnection.TmDistributedVlanConnection.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)

	// Create async
	task, err := vcd.client.CreateTmDistributedVlanConnectionAsync(distributedVlanConnectionType)
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	byIdAsync, err := vcd.client.GetTmDistributedVlanConnectionById(task.Task.Owner.ID)
	check.Assert(err, IsNil)
	check.Assert(byIdAsync, NotNil)
	PrependToCleanupListOpenApi(byIdAsync.TmDistributedVlanConnection.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmDistributedVlanConnections+byIdAsync.TmDistributedVlanConnection.ID)

	err = byIdAsync.Delete()
	check.Assert(err, IsNil)
}
