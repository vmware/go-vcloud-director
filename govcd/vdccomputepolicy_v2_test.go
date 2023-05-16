//go:build vdc || functional || openapi || ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VdcComputePoliciesV2(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	// Step 1 - Create a new VDC Compute Policy
	newComputePolicy := &VdcComputePolicyV2{
		client: &vcd.client.Client,
		VdcComputePolicyV2: &types.VdcComputePolicyV2{
			VdcComputePolicy: types.VdcComputePolicy{
				Name:        check.TestName() + "_empty",
				Description: addrOf("Empty policy created by test"),
			},
			PolicyType: "VdcVmPolicy",
		},
	}

	createdPolicy, err := vcd.client.CreateVdcComputePolicyV2(newComputePolicy.VdcComputePolicyV2)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy.VdcComputePolicyV2.ID, "vdcComputePolicy", "", check.TestName())

	check.Assert(createdPolicy.VdcComputePolicyV2.Name, Equals, newComputePolicy.VdcComputePolicyV2.Name)
	check.Assert(*createdPolicy.VdcComputePolicyV2.Description, Equals, *newComputePolicy.VdcComputePolicyV2.Description)

	newComputePolicy2 := &VdcComputePolicyV2{
		client: &vcd.client.Client,
		VdcComputePolicyV2: &types.VdcComputePolicyV2{
			VdcComputePolicy: types.VdcComputePolicy{
				Name:                       check.TestName(),
				Description:                addrOf("Not Empty policy created by test"),
				CPUSpeed:                   addrOf(100),
				CPUCount:                   addrOf(2),
				CoresPerSocket:             addrOf(1),
				CPUReservationGuarantee:    takeFloatAddress(0.26),
				CPULimit:                   addrOf(200),
				CPUShares:                  addrOf(5),
				Memory:                     addrOf(1600),
				MemoryReservationGuarantee: takeFloatAddress(0.5),
				MemoryLimit:                addrOf(1200),
				MemoryShares:               addrOf(500),
			},
			PolicyType: "VdcVmPolicy",
		},
	}

	createdPolicy2, err := vcd.client.CreateVdcComputePolicyV2(newComputePolicy2.VdcComputePolicyV2)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy2.VdcComputePolicyV2.ID, "vdcComputePolicy", "", check.TestName())

	check.Assert(createdPolicy2.VdcComputePolicyV2.Name, Equals, newComputePolicy2.VdcComputePolicyV2.Name)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.CPUSpeed, Equals, 100)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.CPUCount, Equals, 2)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.CoresPerSocket, Equals, 1)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.CPUReservationGuarantee, Equals, 0.26)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.CPULimit, Equals, 200)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.CPUShares, Equals, 5)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.Memory, Equals, 1600)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.MemoryReservationGuarantee, Equals, 0.5)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.MemoryLimit, Equals, 1200)
	check.Assert(*createdPolicy2.VdcComputePolicyV2.MemoryShares, Equals, 500)

	// Step 2 - Update
	createdPolicy2.VdcComputePolicyV2.Description = addrOf("Updated description")
	updatedPolicy, err := createdPolicy2.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedPolicy.VdcComputePolicyV2, DeepEquals, createdPolicy2.VdcComputePolicyV2)

	// Step 3 - Get all VDC compute policies
	allExistingPolicies, err := vcd.client.GetAllVdcComputePoliciesV2(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingPolicies, NotNil)

	// Step 4 - Get all VDC compute policies using query filters
	for _, onePolicy := range allExistingPolicies {

		// Step 3.1 - Retrieve  using FIQL filter
		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+onePolicy.VdcComputePolicyV2.ID)

		expectOnePolicyResultById, err := vcd.client.GetAllVdcComputePoliciesV2(queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOnePolicyResultById) == 1, Equals, true)

		// Step 2.2 - Retrieve
		exactItem, err := vcd.client.GetVdcComputePolicyV2ById(onePolicy.VdcComputePolicyV2.ID)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		check.Assert(exactItem, NotNil)

		// Step 2.3 - Compare struct retrieved by using filter and the one retrieved by exact ID
		check.Assert(onePolicy, DeepEquals, expectOnePolicyResultById[0])

	}

	// Step 5 - Delete
	err = createdPolicy.Delete()
	check.Assert(err, IsNil)
	// Step 5 - Try to read deleted VDC computed policy should end up with error 'ErrorEntityNotFound'
	deletedPolicy, err := vcd.client.GetVdcComputePolicyV2ById(createdPolicy.VdcComputePolicyV2.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy, IsNil)

	err = createdPolicy2.Delete()
	check.Assert(err, IsNil)
	deletedPolicy2, err := vcd.client.GetVdcComputePolicyV2ById(createdPolicy2.VdcComputePolicyV2.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy2, IsNil)
}

func (vcd *TestVCD) Test_SetAssignedComputePoliciesV2(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	adminVdc, err := org.GetAdminVDCByName(vcd.vdc.Vdc.Name, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)

	// Create a new VDC compute policies
	newComputePolicy := &VdcComputePolicyV2{
		client: &vcd.client.Client,
		VdcComputePolicyV2: &types.VdcComputePolicyV2{
			VdcComputePolicy: types.VdcComputePolicy{
				Name:                    check.TestName() + "1",
				Description:             addrOf("Policy created by Test_SetAssignedComputePolicies"),
				CoresPerSocket:          addrOf(1),
				CPUReservationGuarantee: takeFloatAddress(0.26),
				CPULimit:                addrOf(200),
			},
			PolicyType: "VdcVmPolicy",
		},
	}
	createdPolicy, err := vcd.client.CreateVdcComputePolicyV2(newComputePolicy.VdcComputePolicyV2)
	check.Assert(err, IsNil)
	AddToCleanupList(createdPolicy.VdcComputePolicyV2.ID, "vdcComputePolicy", "", check.TestName())

	newComputePolicy2 := &VdcComputePolicyV2{
		client: &vcd.client.Client,
		VdcComputePolicyV2: &types.VdcComputePolicyV2{
			VdcComputePolicy: types.VdcComputePolicy{
				Name:                    check.TestName() + "2",
				Description:             addrOf("Policy created by Test_SetAssignedComputePolicies"),
				CoresPerSocket:          addrOf(2),
				CPUReservationGuarantee: takeFloatAddress(0.52),
				CPULimit:                addrOf(400),
			},
			PolicyType: "VdcVmPolicy",
		},
	}
	createdPolicy2, err := vcd.client.CreateVdcComputePolicyV2(newComputePolicy2.VdcComputePolicyV2)
	check.Assert(err, IsNil)
	AddToCleanupList(createdPolicy2.VdcComputePolicyV2.ID, "vdcComputePolicy", "", check.TestName())

	// Get default compute policy
	allAssignedComputePolicies, err := adminVdc.GetAllAssignedVdcComputePoliciesV2(nil)
	check.Assert(err, IsNil)
	var defaultPolicyId string
	for _, assignedPolicy := range allAssignedComputePolicies {
		if assignedPolicy.VdcComputePolicyV2.ID == vcd.vdc.Vdc.DefaultComputePolicy.ID {
			defaultPolicyId = assignedPolicy.VdcComputePolicyV2.ID
		}
	}
	allAssignedComputePolicies, err = vcd.client.GetAllAssignedVdcComputePoliciesV2(adminVdc.AdminVdc.ID, nil)
	check.Assert(err, IsNil)
	for _, assignedPolicy := range allAssignedComputePolicies {
		if assignedPolicy.VdcComputePolicyV2.ID == vcd.vdc.Vdc.DefaultComputePolicy.ID {
			defaultPolicyId = assignedPolicy.VdcComputePolicyV2.ID
		}
	}

	vdcComputePolicyHref, err := org.client.OpenApiBuildEndpoint(types.OpenApiPathVersion2_0_0, types.OpenApiEndpointVdcComputePolicies)
	check.Assert(err, IsNil)

	// Assign compute policies to VDC
	policyReferences := types.VdcComputePolicyReferences{VdcComputePolicyReference: []*types.Reference{
		{HREF: vdcComputePolicyHref.String() + createdPolicy.VdcComputePolicyV2.ID},
		{HREF: vdcComputePolicyHref.String() + createdPolicy2.VdcComputePolicyV2.ID},
		{HREF: vdcComputePolicyHref.String() + defaultPolicyId}}}

	assignedVdcComputePolicies, err := adminVdc.SetAssignedComputePolicies(policyReferences)
	check.Assert(err, IsNil)
	check.Assert(strings.SplitAfter(policyReferences.VdcComputePolicyReference[0].HREF, "vdcComputePolicy:")[1], Equals,
		strings.SplitAfter(assignedVdcComputePolicies.VdcComputePolicyReference[0].HREF, "vdcComputePolicy:")[1])
	check.Assert(strings.SplitAfter(policyReferences.VdcComputePolicyReference[1].HREF, "vdcComputePolicy:")[1], Equals,
		strings.SplitAfter(assignedVdcComputePolicies.VdcComputePolicyReference[1].HREF, "vdcComputePolicy:")[1])

	// Cleanup assigned compute policies
	policyReferences = types.VdcComputePolicyReferences{VdcComputePolicyReference: []*types.Reference{
		{HREF: vdcComputePolicyHref.String() + defaultPolicyId}}}

	_, err = adminVdc.SetAssignedComputePolicies(policyReferences)
	check.Assert(err, IsNil)

	err = createdPolicy.Delete()
	check.Assert(err, IsNil)
	err = createdPolicy2.Delete()
	check.Assert(err, IsNil)
}

// Test_VdcVmPlacementPoliciesV2 is similar to Test_VdcComputePoliciesV2 but focused on VM Placement Policies
func (vcd *TestVCD) Test_VdcVmPlacementPoliciesV2(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	if vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup == "" {
		check.Skip("The configuration entry vcd.nsxt_provider_vdc.placementPolicyVmGroup is needed")
	}

	// We need the Provider VDC URN
	pVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	// We also need the VM Group to create a VM Placement Policy
	vmGroup, err := vcd.client.GetVmGroupByNameAndProviderVdcUrn(vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup, pVdc.ProviderVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(vmGroup.VmGroup.Name, Equals, vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup)

	// We'll also use a Logical VM Group to create the VM Placement Policy
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

	// Create a new VDC Compute Policy (VM Placement Policy)
	newComputePolicy := &VdcComputePolicyV2{
		client: &vcd.client.Client,
		VdcComputePolicyV2: &types.VdcComputePolicyV2{
			VdcComputePolicy: types.VdcComputePolicy{
				Name:        check.TestName() + "_empty",
				Description: addrOf("VM Placement Policy created by " + check.TestName()),
			},
			PolicyType: "VdcVmPolicy",
			PvdcNamedVmGroupsMap: []types.PvdcNamedVmGroupsMap{
				{
					NamedVmGroups: []types.OpenApiReferences{
						{
							types.OpenApiReference{
								Name: vmGroup.VmGroup.Name,
								ID:   fmt.Sprintf("%s:%s", vmGroupUrnPrefix, vmGroup.VmGroup.NamedVmGroupId),
							},
						},
					},
					Pvdc: types.OpenApiReference{
						Name: pVdc.ProviderVdc.Name,
						ID:   pVdc.ProviderVdc.ID,
					},
				},
			},
			PvdcLogicalVmGroupsMap: []types.PvdcLogicalVmGroupsMap{
				{
					LogicalVmGroups: types.OpenApiReferences{
						types.OpenApiReference{
							Name: logicalVmGroup.LogicalVmGroup.Name,
							ID:   logicalVmGroup.LogicalVmGroup.ID,
						},
					},
					Pvdc: types.OpenApiReference{
						Name: pVdc.ProviderVdc.Name,
						ID:   pVdc.ProviderVdc.ID,
					},
				},
			},
		},
	}

	createdPolicy, err := vcd.client.CreateVdcComputePolicyV2(newComputePolicy.VdcComputePolicyV2)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy.VdcComputePolicyV2.ID, "vdcComputePolicy", "", check.TestName())

	check.Assert(createdPolicy.VdcComputePolicyV2.Name, Equals, newComputePolicy.VdcComputePolicyV2.Name)
	check.Assert(*createdPolicy.VdcComputePolicyV2.Description, Equals, *newComputePolicy.VdcComputePolicyV2.Description)
	check.Assert(createdPolicy.VdcComputePolicyV2.PvdcLogicalVmGroupsMap, DeepEquals, newComputePolicy.VdcComputePolicyV2.PvdcLogicalVmGroupsMap)
	check.Assert(createdPolicy.VdcComputePolicyV2.PvdcNamedVmGroupsMap, DeepEquals, newComputePolicy.VdcComputePolicyV2.PvdcNamedVmGroupsMap)

	// Update the VM Placement Policy
	createdPolicy.VdcComputePolicyV2.Description = addrOf("Updated description")
	updatedPolicy, err := createdPolicy.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedPolicy.VdcComputePolicyV2, DeepEquals, createdPolicy.VdcComputePolicyV2)

	// Delete the VM Placement Policy and check it doesn't exist anymore
	err = createdPolicy.Delete()
	check.Assert(err, IsNil)
	deletedPolicy, err := vcd.client.GetVdcComputePolicyV2ById(createdPolicy.VdcComputePolicyV2.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy, IsNil)

	// Clean up
	err = logicalVmGroup.Delete()
	check.Assert(err, IsNil)
}

// Test_VdcDuplicatedVmPlacementPolicyGetsACleanError checks that when creating a duplicated VM Placement Policy, consumers
// of the SDK get a nicely formatted error.
// This test should not be needed once function `getFriendlyErrorIfVmPlacementPolicyAlreadyExists` is removed.
func (vcd *TestVCD) Test_VdcDuplicatedVmPlacementPolicyGetsACleanError(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.2") {
		check.Skip("The bug that this test checks for is fixed in 10.4.2")
	}
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	if vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup == "" {
		check.Skip("The configuration entry vcd.nsxt_provider_vdc.placementPolicyVmGroup is needed")
	}

	// We need the Provider VDC URN
	pVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	// We also need the VM Group to create a VM Placement Policy
	vmGroup, err := vcd.client.GetVmGroupByNameAndProviderVdcUrn(vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup, pVdc.ProviderVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(vmGroup.VmGroup.Name, Equals, vcd.config.VCD.NsxtProviderVdc.PlacementPolicyVmGroup)

	// Create a new VDC Compute Policy (VM Placement Policy)
	newComputePolicy := &VdcComputePolicyV2{
		client: &vcd.client.Client,
		VdcComputePolicyV2: &types.VdcComputePolicyV2{
			VdcComputePolicy: types.VdcComputePolicy{
				Name:        check.TestName(),
				Description: addrOf("VM Placement Policy created by " + check.TestName()),
			},
			PolicyType: "VdcVmPolicy",
			PvdcNamedVmGroupsMap: []types.PvdcNamedVmGroupsMap{
				{
					NamedVmGroups: []types.OpenApiReferences{
						{
							types.OpenApiReference{
								Name: vmGroup.VmGroup.Name,
								ID:   fmt.Sprintf("%s:%s", vmGroupUrnPrefix, vmGroup.VmGroup.NamedVmGroupId),
							},
						},
					},
					Pvdc: types.OpenApiReference{
						Name: pVdc.ProviderVdc.Name,
						ID:   pVdc.ProviderVdc.ID,
					},
				},
			},
		},
	}

	createdPolicy, err := vcd.client.CreateVdcComputePolicyV2(newComputePolicy.VdcComputePolicyV2)
	check.Assert(err, IsNil)
	check.Assert(createdPolicy, NotNil)

	AddToCleanupList(createdPolicy.VdcComputePolicyV2.ID, "vdcComputePolicy", "", check.TestName())

	_, err = vcd.client.CreateVdcComputePolicyV2(newComputePolicy.VdcComputePolicyV2)
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), "VM Placement Policy with name '"+check.TestName()+"' already exists"))

	err = createdPolicy.Delete()
	check.Assert(err, IsNil)
}
