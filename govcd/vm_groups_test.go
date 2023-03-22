//go:build functional || openapi || ALL

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

	if vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup == "" {
		check.Skip(fmt.Sprintf("%s test requires vcd.nsxt_provider_vdc.placementPolicyVmGroup configuration", check.TestName()))
	}
	if vcd.config.VCD.NsxtProviderVdc.Name == "" {
		check.Skip(fmt.Sprintf("%s test requires vcd.nsxt_provider_vdc configuration", check.TestName()))
	}

	// We need the Provider VDC URN
	pVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	vmGroup, err := vcd.client.GetVmGroupByNameAndProviderVdcUrn(vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup, pVdc.ProviderVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(vmGroup.VmGroup.Name, Equals, vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup)

	vmGroup2, err := vcd.client.GetVmGroupById(vmGroup.VmGroup.ID)
	check.Assert(err, IsNil)
	check.Assert(vmGroup2, DeepEquals, vmGroup)

	vmGroup3, err := vcd.client.GetVmGroupByNamedVmGroupIdAndProviderVdcUrn(vmGroup2.VmGroup.NamedVmGroupId, pVdc.ProviderVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(vmGroup3, DeepEquals, vmGroup2)

	logicalVmGroup, err := vcd.client.CreateLogicalVmGroup(types.LogicalVmGroup{
		Name: check.TestName(),
		NamedVmGroupReferences: types.OpenApiReferences{
			types.OpenApiReference{
				ID:   fmt.Sprintf("%s:%s", vmGroupUrnPrefix, vmGroup.VmGroup.NamedVmGroupId),
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
