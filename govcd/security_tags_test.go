//go:build network || nsxt || functional || openapi || ALL
// +build network nsxt functional openapi ALL

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"strings"
)

func (vcd *TestVCD) Test_SecurityTags(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointSecurityTags)

	securityTagName1 := strings.ToLower(fmt.Sprintf("%s_%d", check.TestName(), 1)) // Security tags are always lowercase in server-side
	securityTagName2 := strings.ToLower(fmt.Sprintf("%s_%d", check.TestName(), 2))

	// Get testing VM ID
	testingVM, _ := vcd.findFirstVm(*vcd.vapp)
	vm := &VM{
		VM:     &testingVM,
		client: vcd.org.client,
	}

	// Create a security tag using UpdateSecurityTag
	err := vcd.client.UpdateSecurityTag(&types.SecurityTag{
		Tag:      securityTagName1,
		Entities: []string{testingVM.ID},
	})
	check.Assert(err, IsNil)

	// Create a security tag using UpdateVMSecurityTags
	inputEntitySecurityTags := &types.EntitySecurityTags{
		Tags: []string{
			securityTagName1,
			securityTagName2,
		},
	}
	outputEntitySecurityTags, err := vcd.client.UpdateVMSecurityTags(vm.VM.ID, inputEntitySecurityTags)
	check.Assert(err, IsNil)
	check.Assert(outputEntitySecurityTags, NotNil)
	check.Assert(outputEntitySecurityTags, DeepEquals, inputEntitySecurityTags)

	// Check that the VM with security tags is retrieved using GetSecurityTaggedEntities
	securityTaggedEntities, err := vcd.client.GetSecurityTaggedEntities("")
	check.Assert(err, IsNil)
	check.Assert(securityTaggedEntities, NotNil)

	var securityTaggedEntity *types.SecurityTaggedEntity
	for _, v := range securityTaggedEntities {
		if v.ID == testingVM.ID {
			securityTaggedEntity = v
			break
		}
	}

	check.Assert(securityTaggedEntity, NotNil)

	// Check that security tags added before exist (As sysadm)
	securityTagValues, err := vcd.org.GetSecurityTagValues("")
	check.Assert(err, IsNil)
	check.Assert(securityTagValues, NotNil)
	check.Assert(checkIfSecurityTagsExist(securityTagValues, securityTagName1, securityTagName2), Equals, true)

	// Check that security tags added before exist (As org adm)
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(adminOrg, NotNil)
	check.Assert(err, IsNil)

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
	check.Assert(checkIfSecurityTagsExist(securityTagValues, securityTagName1, securityTagName2), Equals, true)

	// Get security tags by VM
	entitySecurityTags, err := vcd.client.GetVMSecurityTags(vm.VM.ID)
	check.Assert(err, IsNil)
	check.Assert(securityTagValues, NotNil)
	check.Assert(contains(securityTagName1, entitySecurityTags.Tags), Equals, true)
	check.Assert(contains(securityTagName2, entitySecurityTags.Tags), Equals, true)

	// Remove tags
	err = vcd.client.UpdateSecurityTag(&types.SecurityTag{
		Tag:      securityTagName1,
		Entities: []string{},
	})
	check.Assert(err, IsNil)

	err = vcd.client.UpdateSecurityTag(&types.SecurityTag{
		Tag:      securityTagName2,
		Entities: []string{},
	})
	check.Assert(err, IsNil)

}

func checkIfSecurityTagsExist(securityTagValues []*types.SecurityTagValue, securityTagName ...string) bool {
	var numberFound int
	for _, v := range securityTagName {
		for _, tag := range securityTagValues {
			if tag.Tag == v {
				numberFound++
				break
			}
		}
	}

	return numberFound == len(securityTagName)
}
