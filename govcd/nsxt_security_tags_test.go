package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"strings"
)

func (vcd *TestVCD) Test_GetSecurityTaggedEntities(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointSecurityTags)

	securityTaggedEntities, err := vcd.org.GetSecurityTaggedEntities("")
	check.Assert(err, IsNil)
	check.Assert(securityTaggedEntities, NotNil)
}

func (vcd *TestVCD) Test_GetSecurityTagValues(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointSecurityTags)

	securityTagValues, err := vcd.org.GetSecurityTagValues("")
	check.Assert(err, IsNil)
	check.Assert(securityTagValues, NotNil)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

	// Prep Org admin user and run firewall tests
	userName := strings.ToLower(check.TestName())
	fmt.Printf("# Running Get Security Tag Values test as Org Admin user '%s'\n", userName)
	orgUserVcdClient, err := newOrgUserConnection(adminOrg, userName, "CHANGE-ME", vcd.config.Provider.Url, true)
	check.Assert(err, IsNil)
	check.Assert(orgUserVcdClient, NotNil)

	orgUserOrg, err := orgUserVcdClient.GetOrgById(adminOrg.AdminOrg.ID)
	check.Assert(err, IsNil)

	securityTagValues, err = orgUserOrg.GetSecurityTagValues("")
	check.Assert(err, IsNil)
	check.Assert(securityTagValues, NotNil)
}

func (vcd *TestVCD) Test_GetVMSecurityTags(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointSecurityTags)

	securityTagValues, err := vcd.org.GetVMSecurityTags("urn:vcloud:vm:9f895262-2942-4826-8421-5ff3f1f53459")
	check.Assert(err, IsNil)
	check.Assert(securityTagValues, NotNil)
}
