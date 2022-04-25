package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtSecurityTags(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointSecurityTags)

	securityTaggedEntities, err := vcd.org.GetSecurityTaggedEntities("")
	check.Assert(err, IsNil)
	check.Assert(securityTaggedEntities, NotNil)
}
