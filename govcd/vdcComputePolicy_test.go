// +build functional openapi ALL

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VdcComputePolicies(check *C) {
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
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

	createdPolicy, err := org.CreateVdcComputePolicy(newComputePolicy.VdcComputePolicy)
	check.Assert(err, IsNil)
	check.Assert(createdPolicy.VdcComputePolicy.Name, Equals, newComputePolicy.VdcComputePolicy.Name)
	check.Assert(createdPolicy.VdcComputePolicy.Description, Equals, newComputePolicy.VdcComputePolicy.Description)

	AddToCleanupList(createdPolicy.VdcComputePolicy.ID, "vcdComputePolicy", vcd.org.Org.Name, "Test_VdcComputePolicies")

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

	createdPolicy2, err := org.CreateVdcComputePolicy(newComputePolicy2.VdcComputePolicy)
	check.Assert(err, IsNil)
	check.Assert(createdPolicy2.VdcComputePolicy.Name, Equals, newComputePolicy2.VdcComputePolicy.Name)
	check.Assert(createdPolicy2.VdcComputePolicy.Description, Equals, newComputePolicy2.VdcComputePolicy.Description)

	AddToCleanupList(createdPolicy2.VdcComputePolicy.ID, "vcdComputePolicy", vcd.org.Org.Name, "Test_VdcComputePolicies")
	// Step 2 - update
	createdPolicy2.VdcComputePolicy.Description = "Updated description"
	updatedPolicy, err := createdPolicy2.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedPolicy.VdcComputePolicy, DeepEquals, createdPolicy2.VdcComputePolicy)

	// Step 3 - Get all VDC compute policies
	allExistingPolicies, err := org.GetAllVdcComputePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingPolicies, NotNil)

	// Step 4 - Get all VDC compute policies using query filters
	for _, onePolicy := range allExistingPolicies {

		// Step 3.1 - retrieve  using FIQL filter
		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+onePolicy.ID)

		expectOnePolicyResultById, err := org.GetAllVdcComputePolicies(queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOnePolicyResultById) == 1, Equals, true)

		// Step 2.2 - retrieve
		exactItem, err := org.GetVdcComputePolicyById(onePolicy.ID)
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
	deletedPolicy, err := org.GetVdcComputePolicyById(createdPolicy.VdcComputePolicy.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy, IsNil)

	err = createdPolicy2.Delete()
	check.Assert(err, IsNil)
	deletedPolicy2, err := org.GetVdcComputePolicyById(createdPolicy2.VdcComputePolicy.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedPolicy2, IsNil)
}
