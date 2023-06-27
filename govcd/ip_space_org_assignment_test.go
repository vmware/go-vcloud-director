//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_IpSpaceOrgAssignment(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointIpSpaceOrgAssignments)

	ipSpace := createIpSpace(vcd, check)
	extNet := createExternalNetwork(vcd, check)

	// IP Space uplink (not directly referenced anywhere, but is required to make IP allocations)
	ipSpaceUplink := createIpSpaceUplink(vcd, check, extNet.ExternalNetwork.ID, ipSpace.IpSpace.ID)

	// Check if any Org assignments are found before Edge Gateway creation - there should be none as
	// Org assignments are implicitly created during Edge Gateway creation
	allOrgAssignments, err := ipSpace.GetAllOrgAssignments(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allOrgAssignments), Equals, 0)

	// Create NSX-T Edge Gateway
	edgeGw := createNsxtEdgeGateway(vcd, check, extNet.ExternalNetwork.ID)

	// After the Edge Gateway is created - one can find an implicitly created IP Space Org Assignment
	orgAssignmentByOrgName, err := ipSpace.GetOrgAssignmentByOrgName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(orgAssignmentByOrgName, NotNil)

	orgAssignmentByOrgId, err := ipSpace.GetOrgAssignmentByOrgId(vcd.org.Org.ID)
	check.Assert(err, IsNil)
	check.Assert(orgAssignmentByOrgId, NotNil)

	// Get Org Assignment by ID
	orgAssignmentById, err := ipSpace.GetOrgAssignmentById(orgAssignmentByOrgId.IpSpaceOrgAssignment.ID)
	check.Assert(err, IsNil)
	check.Assert(orgAssignmentById.IpSpaceOrgAssignment, DeepEquals, orgAssignmentByOrgId.IpSpaceOrgAssignment)

	// Get All org Assignments and check that there is exactly one - matching other lookup methods
	allOrgAssignments, err = ipSpace.GetAllOrgAssignments(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allOrgAssignments), Equals, 1)
	check.Assert(allOrgAssignments[0].IpSpaceOrgAssignment, DeepEquals, orgAssignmentByOrgId.IpSpaceOrgAssignment)
	check.Assert(allOrgAssignments[0].IpSpaceOrgAssignment, DeepEquals, orgAssignmentByOrgName.IpSpaceOrgAssignment)

	// Update
	orgAssignmentById.IpSpaceOrgAssignment.CustomQuotas = &types.IpSpaceOrgAssignmentQuotas{
		FloatingIPQuota: addrOf(10),
		IPPrefixQuotas: []types.IpSpaceOrgAssignmentIPPrefixQuotas{
			{
				PrefixLength: addrOf(31),
				Quota:        addrOf(11),
			},
			{
				PrefixLength: addrOf(30),
				Quota:        addrOf(12),
			},
		},
	}

	updatedOrgAssignmentCustomQuota, err := orgAssignmentById.Update(orgAssignmentById.IpSpaceOrgAssignment)
	check.Assert(err, IsNil)
	check.Assert(updatedOrgAssignmentCustomQuota.IpSpaceOrgAssignment.CustomQuotas.FloatingIPQuota, DeepEquals, orgAssignmentById.IpSpaceOrgAssignment.CustomQuotas.FloatingIPQuota)
	check.Assert(len(updatedOrgAssignmentCustomQuota.IpSpaceOrgAssignment.CustomQuotas.IPPrefixQuotas), DeepEquals, len(orgAssignmentById.IpSpaceOrgAssignment.CustomQuotas.IPPrefixQuotas))

	// Cleanup
	err = edgeGw.Delete()
	check.Assert(err, IsNil)

	err = ipSpaceUplink.Delete()
	check.Assert(err, IsNil)

	err = extNet.Delete()
	check.Assert(err, IsNil)

	err = ipSpace.Delete()
	check.Assert(err, IsNil)
}
