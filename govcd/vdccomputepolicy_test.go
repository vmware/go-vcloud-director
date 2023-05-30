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

func (vcd *TestVCD) Test_VdcComputePolicies(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	client := &vcd.client.Client
	// Step 1 - Create a new VDC compute policies
	newComputePolicy := &VdcComputePolicy{
		client: client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:        check.TestName() + "_empty",
			Description: addrOf("Empty policy created by test"),
		},
	}

	createdPolicy, err := client.CreateVdcComputePolicy(newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy.VdcComputePolicy.ID, "vdcComputePolicy", "", check.TestName())

	check.Assert(createdPolicy.VdcComputePolicy.Name, Equals, newComputePolicy.VdcComputePolicy.Name)
	check.Assert(*createdPolicy.VdcComputePolicy.Description, Equals, *newComputePolicy.VdcComputePolicy.Description)

	newComputePolicy2 := &VdcComputePolicy{
		client: client,
		VdcComputePolicy: &types.VdcComputePolicy{
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
	}

	createdPolicy2, err := client.CreateVdcComputePolicy(newComputePolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)

	AddToCleanupList(createdPolicy2.VdcComputePolicy.ID, "vdcComputePolicy", "", check.TestName())

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
	createdPolicy2.VdcComputePolicy.Description = addrOf("Updated description")
	updatedPolicy, err := createdPolicy2.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedPolicy.VdcComputePolicy, DeepEquals, createdPolicy2.VdcComputePolicy)

	// Step 3 - Get all VDC compute policies
	allExistingPolicies, err := client.GetAllVdcComputePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingPolicies, NotNil)

	// Step 4 - Get all VDC compute policies using query filters
	for _, onePolicy := range allExistingPolicies {

		// Step 3.1 - retrieve  using FIQL filter
		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+onePolicy.VdcComputePolicy.ID)

		expectOnePolicyResultById, err := client.GetAllVdcComputePolicies(queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOnePolicyResultById) == 1, Equals, true)

		// Step 2.2 - retrieve
		exactItem, err := client.GetVdcComputePolicyById(onePolicy.VdcComputePolicy.ID)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		check.Assert(exactItem, NotNil)

		// Step 2.3 - compare struct retrieved by using filter and the one retrieved by exact ID
		check.Assert(onePolicy, DeepEquals, expectOnePolicyResultById[0])

	}

	// Step 5 - delete
	err = createdPolicy.Delete()
	check.Assert(err, IsNil)
	// Step 5 - try to read deleted VDC computed policy should end up with error 'ErrorEntityNotFound'
	deletedPolicy, err := client.GetVdcComputePolicyById(createdPolicy.VdcComputePolicy.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy, IsNil)

	err = createdPolicy2.Delete()
	check.Assert(err, IsNil)
	deletedPolicy2, err := client.GetVdcComputePolicyById(createdPolicy2.VdcComputePolicy.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy2, IsNil)
}

func (vcd *TestVCD) Test_SetAssignedComputePolicies(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	client := &vcd.client.Client
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	adminVdc, err := org.GetAdminVDCByName(vcd.vdc.Vdc.Name, false)
	if adminVdc == nil || err != nil {
		vcd.infoCleanup(notFoundMsg, "vdc", vcd.vdc.Vdc.Name)
	}

	// Step 1 - Create a new VDC compute policies
	newComputePolicy := &VdcComputePolicy{
		client: org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:                    check.TestName() + "1",
			Description:             addrOf("Policy created by Test_SetAssignedComputePolicies"),
			CoresPerSocket:          addrOf(1),
			CPUReservationGuarantee: takeFloatAddress(0.26),
			CPULimit:                addrOf(200),
		},
	}
	createdPolicy, err := client.CreateVdcComputePolicy(newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)
	AddToCleanupList(createdPolicy.VdcComputePolicy.ID, "vdcComputePolicy", "", check.TestName())

	newComputePolicy2 := &VdcComputePolicy{
		client: org.client,
		VdcComputePolicy: &types.VdcComputePolicy{
			Name:                    check.TestName() + "2",
			Description:             addrOf("Policy created by Test_SetAssignedComputePolicies"),
			CoresPerSocket:          addrOf(2),
			CPUReservationGuarantee: takeFloatAddress(0.52),
			CPULimit:                addrOf(400),
		},
	}
	createdPolicy2, err := client.CreateVdcComputePolicy(newComputePolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)
	AddToCleanupList(createdPolicy2.VdcComputePolicy.ID, "vdcComputePolicy", "", check.TestName())

	// Get default compute policy
	allAssignedComputePolicies, err := adminVdc.GetAllAssignedVdcComputePolicies(nil)
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

	assignedVdcComputePolicies, err := adminVdc.SetAssignedComputePolicies(policyReferences)
	check.Assert(err, IsNil)
	check.Assert(strings.SplitAfter(policyReferences.VdcComputePolicyReference[0].HREF, "vdcComputePolicy:")[1], Equals,
		strings.SplitAfter(assignedVdcComputePolicies.VdcComputePolicyReference[0].HREF, "vdcComputePolicy:")[1])
	check.Assert(strings.SplitAfter(policyReferences.VdcComputePolicyReference[1].HREF, "vdcComputePolicy:")[1], Equals,
		strings.SplitAfter(assignedVdcComputePolicies.VdcComputePolicyReference[1].HREF, "vdcComputePolicy:")[1])

	// cleanup assigned compute policies
	policyReferences = types.VdcComputePolicyReferences{VdcComputePolicyReference: []*types.Reference{
		{HREF: vdcComputePolicyHref.String() + defaultPolicyId}}}

	_, err = adminVdc.SetAssignedComputePolicies(policyReferences)
	check.Assert(err, IsNil)

	err = createdPolicy.Delete()
	check.Assert(err, IsNil)
	err = createdPolicy2.Delete()
	check.Assert(err, IsNil)
}
