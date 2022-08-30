//go:build functional || openapi || ALL
// +build functional openapi ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// TODO: Add VM Placement Policy name in vcd.config.VCD.PlacementPolicy
// This test retrieves a VM Placement Policy to check the behaviour of VM Groups and Logical VM Groups.
// With the information of this VM Placement Policy we test the read logic, then we create a Logical VM Group and
// finally we delete it.
func (vcd *TestVCD) Test_VmGroupsCRUD(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	computePolicies, err := vcd.client.Client.GetAllVdcComputePolicies(url.Values{
		"filter": []string{fmt.Sprintf("name==%s;policyType==VdcVmPolicy", vcd.config.VCD.PlacementPolicy)},
	})
	check.Assert(err, IsNil)
	check.Assert(len(computePolicies), Equals, 1)
	computePolicy := computePolicies[0]
	check.Assert(len(computePolicy.VdcComputePolicy.NamedVMGroups), Not(Equals), 0)
	check.Assert(len(computePolicy.VdcComputePolicy.NamedVMGroups[0]), Not(Equals), 0)

	vmGroup, err := vcd.client.GetVmGroupByNamedVmGroupId(computePolicy.VdcComputePolicy.NamedVMGroups[0][0].ID)
	check.Assert(err, IsNil)
	check.Assert(vmGroup.VmGroup.NamedVmGroupId, Equals, extractUuid(computePolicy.VdcComputePolicy.NamedVMGroups[0][0].ID))
	check.Assert(vmGroup.VmGroup.Name, Equals, computePolicy.VdcComputePolicy.NamedVMGroups[0][0].Name)

	pVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	logicalVmGroup, err := vcd.client.CreateLogicalVmGroup(types.LogicalVmGroup{
		Name: check.TestName(),
		NamedVmGroupReferences: types.OpenApiReferences{
			types.OpenApiReference{
				ID:   computePolicy.VdcComputePolicy.NamedVMGroups[0][0].ID,
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
