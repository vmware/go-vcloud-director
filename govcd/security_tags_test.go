//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func init() {
	testingTags["vm"] = "security_tags_test.go"
}

func (vcd *TestVCD) Test_SecurityTags(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointSecurityTags)

	securityTagName1 := strings.ToLower(fmt.Sprintf("%s_%d", check.TestName(), 1)) // Security tags are always lowercase in server-side
	securityTagName2 := strings.ToLower(fmt.Sprintf("%s_%d", check.TestName(), 2))
	nonExistingSecurityTag := "icompletelymadeupthistag1234"

	// Get testing Org
	testingOrg, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(testingOrg, NotNil)

	// Get testing VM
	testingVM, _ := vcd.findFirstVm(*vcd.vapp)
	vm := &VM{
		VM:     &testingVM,
		client: vcd.org.client,
	}

	// Create a security tag using UpdateSecurityTag
	sentSecurityTag := &types.SecurityTag{
		Tag:      securityTagName1,
		Entities: []string{testingVM.ID},
	}

	receivedSecurityTag, err := testingOrg.UpdateSecurityTag(sentSecurityTag)
	check.Assert(err, IsNil)
	check.Assert(sentSecurityTag, DeepEquals, receivedSecurityTag)

	// Check that the security tag exist using Org.GetAllSecurityTaggedEntitiesByName
	securityTagEntities, err := testingOrg.GetAllSecurityTaggedEntitiesByName(securityTagName1)
	check.Assert(err, IsNil)
	check.Assert(len(securityTagEntities) > 0, Equals, true)

	// Check that ErrorEntityNotFound is returned if no entities where found with Org.GetAllSecurityTaggedEntitiesByName
	securityTagEntities, err = testingOrg.GetAllSecurityTaggedEntitiesByName(nonExistingSecurityTag)
	check.Assert(err, NotNil)
	check.Assert(err, Equals, ErrorEntityNotFound)
	check.Assert(securityTagEntities, IsNil)

	// Create a security tag using UpdateVMSecurityTags
	inputEntitySecurityTags := &types.EntitySecurityTags{
		Tags: []string{
			securityTagName1,
			securityTagName2,
		},
	}
	outputEntitySecurityTags, err := vm.UpdateVMSecurityTags(inputEntitySecurityTags)
	check.Assert(err, IsNil)
	check.Assert(outputEntitySecurityTags, NotNil)
	check.Assert(outputEntitySecurityTags, DeepEquals, inputEntitySecurityTags)

	// Check that the VM with security tags is retrieved using GetSecurityTaggedEntities
	securityTaggedEntities, err := testingOrg.GetAllSecurityTaggedEntities(nil)
	check.Assert(err, IsNil)
	check.Assert(len(securityTaggedEntities) > 0, Equals, true)

	var securityTaggedEntity types.SecurityTaggedEntity
	for _, v := range securityTaggedEntities {
		if v.ID == testingVM.ID {
			securityTaggedEntity = v
			break
		}
	}

	check.Assert(securityTaggedEntity, NotNil)

	// Check that security tags added before exist (As sysadm)
	securityTagValues, err := testingOrg.GetAllSecurityTagValues(nil)
	check.Assert(err, IsNil)
	check.Assert(len(securityTagValues) > 0, Equals, true)
	check.Assert(checkIfSecurityTagsExist(securityTagValues, securityTagName1, securityTagName2), Equals, true)

	// Check that security tags added before exist (As org adm)
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	userName := strings.ToLower(check.TestName())
	fmt.Printf("# Running Get Security Tag Values test as Org Admin user '%s'\n", userName)
	orgUserVcdClient, _, err := newOrgUserConnection(adminOrg, userName, "CHANGE-ME", vcd.config.Provider.Url, true)
	check.Assert(err, IsNil)
	check.Assert(orgUserVcdClient, NotNil)

	orgUserOrg, err := orgUserVcdClient.GetOrgById(adminOrg.AdminOrg.ID)
	check.Assert(err, IsNil)

	securityTagValues, err = orgUserOrg.GetAllSecurityTagValues(nil)
	check.Assert(err, IsNil)
	check.Assert(len(securityTagValues) > 0, Equals, true)
	check.Assert(checkIfSecurityTagsExist(securityTagValues, securityTagName1, securityTagName2), Equals, true)

	// Get security tags by VM
	entitySecurityTags, err := vm.GetVMSecurityTags()
	check.Assert(err, IsNil)
	check.Assert(securityTagValues, NotNil)
	check.Assert(contains(securityTagName1, entitySecurityTags.Tags), Equals, true)
	check.Assert(contains(securityTagName2, entitySecurityTags.Tags), Equals, true)

	// Remove tags
	sentSecurityTag = &types.SecurityTag{
		Tag:      securityTagName1,
		Entities: []string{},
	}

	receivedSecurityTag, err = testingOrg.UpdateSecurityTag(sentSecurityTag)
	check.Assert(err, IsNil)
	check.Assert(receivedSecurityTag.Tag, Equals, sentSecurityTag.Tag)
	check.Assert(len(receivedSecurityTag.Entities), Equals, 0)

	sentSecurityTag2 := &types.SecurityTag{
		Tag:      securityTagName2,
		Entities: []string{},
	}

	receivedSecurityTag2, err := testingOrg.UpdateSecurityTag(sentSecurityTag2)
	check.Assert(err, IsNil)
	check.Assert(receivedSecurityTag2.Tag, Equals, sentSecurityTag2.Tag)
	check.Assert(len(receivedSecurityTag2.Entities), Equals, 0)
}

func checkIfSecurityTagsExist(securityTagValues []types.SecurityTagValue, securityTagName ...string) bool {
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
