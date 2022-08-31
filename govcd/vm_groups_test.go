//go:build functional || openapi || ALL
// +build functional openapi ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// This test checks the correct behaviour of the read, create and delete operations for VM Groups and Logical VM Groups.
func (vcd *TestVCD) Test_VmGroupsCRUD(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	if vcd.config.VCD.PlacementPolicyVmGroup == "" {
		check.Skip(fmt.Sprintf("%s test requires vcd.placementPolicyVmGroup configuration", check.TestName()))
	}

	vmGroup, err := vcd.client.GetVmGroupByName(vcd.config.VCD.PlacementPolicyVmGroup)
	check.Assert(err, IsNil)
	check.Assert(vmGroup.VmGroup.Name, Equals, vcd.config.VCD.PlacementPolicyVmGroup)

	vmGroup2, err := vcd.client.GetVmGroupByNamedVmGroupId(vmGroup.VmGroup.NamedVmGroupId)
	check.Assert(err, IsNil)
	check.Assert(vmGroup, DeepEquals, vmGroup2)

	// We need the Provider VDC to create a Logical VM Group
	pVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	logicalVmGroup, err := vcd.client.CreateLogicalVmGroup(types.LogicalVmGroup{
		Name: check.TestName(),
		NamedVmGroupReferences: types.OpenApiReferences{
			types.OpenApiReference{
				ID:   fmt.Sprintf("urn:vcloud:namedVmGroup:%s", vmGroup.VmGroup.NamedVmGroupId),
				Name: vmGroup.VmGroup.Name},
		},
		PvdcID: pVdc.ProviderVdc.ID,
	})
	check.Assert(err, IsNil)
	AddToCleanupList(logicalVmGroup.LogicalVmGroup.ID, "logicalVmGroup", "", check.TestName())

	retrievedLogicalVmGroup, err := vcd.client.GetLogicalVmGroupById(logicalVmGroup.LogicalVmGroup.ID)
	check.Assert(err, IsNil)
	check.Assert(retrievedLogicalVmGroup.LogicalVmGroup, DeepEquals, logicalVmGroup.LogicalVmGroup)

	err = logicalVmGroup.Delete()
	check.Assert(err, IsNil)
}
