// +build functional openapi ALL

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_Roles(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	allExistingRoles, err := adminOrg.GetAllOpenApiRoles(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingRoles, NotNil)

	// Step 2 - Get all roles using query filters
	for _, oneRole := range allExistingRoles {

		// Step 2.1 - retrieve specific role by using FIQL filter
		// urlRef2, err := vcd.client.Client.BuildOpenApiEndpoint("1.0.0/roles")
		// check.Assert(err, IsNil)

		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+oneRole.ID)

		// expectOneRoleResultById := []*InlineRoles{{}}
		expectOneRoleResultById, err := adminOrg.GetAllOpenApiRoles(queryParams)
		check.Assert(err, IsNil)
		check.Assert(len(expectOneRoleResultById) == 1, Equals, true)

		// Step 2.2 - retrieve specific role by using endpoint
		exactItem, err := adminOrg.GetOpenApiRoleById(oneRole.ID)
		check.Assert(err, IsNil)

		check.Assert(err, IsNil)
		check.Assert(exactItem, NotNil)

		// Step 2.3 - compare struct retrieved by using filter and the one retrieved by exact endpoint ID
		check.Assert(oneRole, DeepEquals, expectOneRoleResultById[0])

	}

	// Step 3 - Create a new role and ensure it is created as specified by doing deep comparison

	newR := &OpenApiRole{
		client: adminOrg.client,
		Role: &types.Role{
			Name:        check.TestName(),
			Description: "Role created by test",
			// This BundleKey is being set by VCD even if it is not sent
			BundleKey: "com.vmware.vcloud.undefined.key",
			ReadOnly:  false,
		},
	}

	createdRole, err := newR.Create(newR.Role)
	check.Assert(err, IsNil)

	// Ensure supplied and created structs differ only by ID
	newR.Role.ID = createdRole.Role.ID
	check.Assert(createdRole.Role, DeepEquals, newR.Role)

	// Step 4 - updated created role
	newR.Role.Description = "Updated description"
	updatedRole, err := newR.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedRole.Role, DeepEquals, newR.Role)

	// Step 5 - delete created role
	err = updatedRole.Delete()
	check.Assert(err, IsNil)
	// Step 5 - try to read deleted role and expect error to contain 'ErrorEntityNotFound'
	// Read is tricky - it throws an error ACCESS_TO_RESOURCE_IS_FORBIDDEN when the resource with ID does not
	// exist therefore one cannot know what kind of error occurred.
	deletedRole, err := adminOrg.GetOpenApiRoleById(createdRole.Role.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(deletedRole, IsNil)
}
