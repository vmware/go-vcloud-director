// +build vdc functional openapi ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VdcComputePolicies(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	ctx := context.Background()
	if vcd.client.Client.APIVCDMaxVersionIs(ctx, "< 33.0") {
		check.Skip(fmt.Sprintf("Test %s requires VCD 10.0 (API version 33) or higher", check.TestName()))
	}

	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	// Step 1 - Create a new VDC compute policies
	newComputePolicy := &VdcComputePolicy{
		client: org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:        check.TestName() + "_empty",
			Description: "Empty policy created by test",
		},
	}

	createdPolicy, err := org.CreateVdcComputePolicy(ctx, newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_VdcComputePolicies")

	check.Assert(createdPolicy.VdcComputePolicy.Name, Equals, newComputePolicy.VdcComputePolicy.Name)
	check.Assert(createdPolicy.VdcComputePolicy.Description, Equals, newComputePolicy.VdcComputePolicy.Description)

	newComputePolicy2 := &VdcComputePolicy{
		client: org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:                       check.TestName(),
			Description:                "Not Empty policy created by test",
			CPUSpeed:                   takeIntAddress(100),
			CPUCount:                   takeIntAddress(2),
			CoresPerSocket:             takeIntAddress(1),
			CPUReservationGuarantee:    takeFloatAddress(0.26),
			CPULimit:                   takeIntAddress(200),
			CPUShares:                  takeIntAddress(5),
			Memory:                     takeIntAddress(1600),
			MemoryReservationGuarantee: takeFloatAddress(0.5),
			MemoryLimit:                takeIntAddress(1200),
			MemoryShares:               takeIntAddress(500),
		},
	}

	createdPolicy2, err := org.CreateVdcComputePolicy(ctx, newComputePolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy2.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_VdcComputePolicies")

	check.Assert(createdPolicy2.VdcComputePolicy.Name, Equals, newComputePolicy2.VdcComputePolicy.Name)
	check.Assert(*createdPolicy2.VdcComputePolicy.CPUSpeed, Equals, 100)
	check.Assert(*createdPolicy2.VdcComputePolicy.CPUCount, Equals, 2)
	check.Assert(*createdPolicy2.VdcComputePolicy.CoresPerSocket, Equals, 1)
	check.Assert(*createdPolicy2.VdcComputePolicy.CPUReservationGuarantee, Equals, 0.26)
	check.Assert(*createdPolicy2.VdcComputePolicy.CPULimit, Equals, 200)
	check.Assert(*createdPolicy2.VdcComputePolicy.CPUShares, Equals, 5)
	check.Assert(*createdPolicy2.VdcComputePolicy.Memory, Equals, 1600)
	check.Assert(*createdPolicy2.VdcComputePolicy.MemoryReservationGuarantee, Equals, 0.5)
	check.Assert(*createdPolicy2.VdcComputePolicy.MemoryLimit, Equals, 1200)
	check.Assert(*createdPolicy2.VdcComputePolicy.MemoryShares, Equals, 500)

	// Step 2 - update
	createdPolicy2.VdcComputePolicy.Description = "Updated description"
	updatedPolicy, err := createdPolicy2.Update(ctx)
	check.Assert(err, IsNil)
	check.Assert(updatedPolicy.VdcComputePolicy, DeepEquals, createdPolicy2.VdcComputePolicy)

	// Step 3 - Get all VDC compute policies
	allExistingPolicies, err := org.GetAllVdcComputePolicies(ctx, nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingPolicies, NotNil)

	// Step 4 - Get all VDC compute policies using query filters
	for _, onePolicy := range allExistingPolicies {

		// Step 3.1 - retrieve  using FIQL filter
		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+onePolicy.VdcComputePolicy.ID)

		expectOnePolicyResultById, err := org.GetAllVdcComputePolicies(ctx, queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOnePolicyResultById) == 1, Equals, true)

		// Step 2.2 - retrieve
		exactItem, err := org.GetVdcComputePolicyById(ctx, onePolicy.VdcComputePolicy.ID)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		check.Assert(exactItem, NotNil)

		// Step 2.3 - compare struct retrieved by using filter and the one retrieved by exact ID
		check.Assert(onePolicy, DeepEquals, expectOnePolicyResultById[0])

	}

	// Step 5 - delete
	err = createdPolicy.Delete(ctx)
	check.Assert(err, IsNil)
	// Step 5 - try to read deleted VDC computed policy should end up with error 'ErrorEntityNotFound'
	deletedPolicy, err := org.GetVdcComputePolicyById(ctx, createdPolicy.VdcComputePolicy.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy, IsNil)

	err = createdPolicy2.Delete(ctx)
	check.Assert(err, IsNil)
	deletedPolicy2, err := org.GetVdcComputePolicyById(ctx, createdPolicy2.VdcComputePolicy.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy2, IsNil)
}

func (vcd *TestVCD) Test_SetAssignedComputePolicies(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	ctx := context.Background()
	if vcd.client.Client.APIVCDMaxVersionIs(ctx, "< 33.0") {
		check.Skip(fmt.Sprintf("Test %s requires VCD 10.0 (API version 33) or higher", check.TestName()))
	}

	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	adminVdc, err := org.GetAdminVDCByName(ctx, vcd.vdc.Vdc.Name, false)
	if adminVdc == nil || err != nil {
		vcd.infoCleanup(notFoundMsg, "vdc", vcd.vdc.Vdc.Name)
	}

	// Step 1 - Create a new VDC compute policies
	newComputePolicy := &VdcComputePolicy{
		client: org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:                    check.TestName() + "1",
			Description:             "Policy created by Test_SetVdcComputePolicies",
			CoresPerSocket:          takeIntAddress(1),
			CPUReservationGuarantee: takeFloatAddress(0.26),
			CPULimit:                takeIntAddress(200),
		},
	}
	createdPolicy, err := org.CreateVdcComputePolicy(ctx, newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)
	AddToCleanupList(createdPolicy.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_VdcComputePolicies")

	newComputePolicy2 := &VdcComputePolicy{
		client: org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:                    check.TestName() + "2",
			Description:             "Policy created by Test_SetVdcComputePolicies",
			CoresPerSocket:          takeIntAddress(2),
			CPUReservationGuarantee: takeFloatAddress(0.52),
			CPULimit:                takeIntAddress(400),
		},
	}
	createdPolicy2, err := org.CreateVdcComputePolicy(ctx, newComputePolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)
	AddToCleanupList(createdPolicy2.VdcComputePolicy.ID, "vdcComputePolicy", vcd.org.Org.Name, "Test_VdcComputePolicies")

	// Get default compute policy
	allAssignedComputePolicies, err := adminVdc.GetAllAssignedVdcComputePolicies(ctx, nil)
	check.Assert(err, IsNil)
	var defaultPolicyId string
	for _, assignedPolicy := range allAssignedComputePolicies {
		if assignedPolicy.VdcComputePolicy.ID == vcd.vdc.Vdc.DefaultComputePolicy.ID {
			defaultPolicyId = assignedPolicy.VdcComputePolicy.ID
		}
	}

	vdcComputePolicyHref, err := org.client.OpenApiBuildEndpoint(types.OpenApiPathVersion1_0_0, types.OpenApiEndpointVdcComputePolicies)
	check.Assert(err, IsNil)

	// Assign compute policies to VDC
	policyReferences := types.VdcComputePolicyReferences{VdcComputePolicyReference: []*types.Reference{&types.Reference{HREF: vdcComputePolicyHref.String() + createdPolicy.VdcComputePolicy.ID},
		&types.Reference{HREF: vdcComputePolicyHref.String() + createdPolicy2.VdcComputePolicy.ID},
		{HREF: vdcComputePolicyHref.String() + defaultPolicyId}}}

	assignedVdcComputePolicies, err := adminVdc.SetAssignedComputePolicies(ctx, policyReferences)
	check.Assert(err, IsNil)
	check.Assert(strings.SplitAfter(policyReferences.VdcComputePolicyReference[0].HREF, "vdcComputePolicy:")[1], Equals,
		strings.SplitAfter(assignedVdcComputePolicies.VdcComputePolicyReference[0].HREF, "vdcComputePolicy:")[1])
	check.Assert(strings.SplitAfter(policyReferences.VdcComputePolicyReference[1].HREF, "vdcComputePolicy:")[1], Equals,
		strings.SplitAfter(assignedVdcComputePolicies.VdcComputePolicyReference[1].HREF, "vdcComputePolicy:")[1])

	// cleanup assigned compute policies
	policyReferences = types.VdcComputePolicyReferences{VdcComputePolicyReference: []*types.Reference{
		{HREF: vdcComputePolicyHref.String() + defaultPolicyId}}}

	_, err = adminVdc.SetAssignedComputePolicies(ctx, policyReferences)
	check.Assert(err, IsNil)
}
