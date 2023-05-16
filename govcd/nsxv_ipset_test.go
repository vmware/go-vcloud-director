//go:build nsxv || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxvIpSet performes the following actions:
// 1. Creates an IP set and checks it was created
// 2. Tries to create duplicate name IP set and expects error
// 3. Gets all IP sets and ensures there are some
// 4. Validates GetByName, GetByID, GetByNameOrID
// 5. Updates IP set and validates only one field is changed
// 6. Deletes created IP set by ID
// 7. Deletes created IP set by Name
func (vcd *TestVCD) Test_NsxvIpSet(check *C) {

	if vcd.config.VCD.Org == "" {
		check.Skip(check.TestName() + ": Org name not given")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip(check.TestName() + ": VDC name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// 1. Create
	ipSetConfig := &types.EdgeIpSet{
		Name:        "test-ipset",
		Description: "test-ipset-description",
		// The below example demonstrates a problem when vCD API shuffles the below list and the
		// answer becomes ordered differently then it was submitted. If submitted as it was
		// returned, next time it shuffles it again so one can not rely on order of list returned.
		// IPAddresses: "192.168.200.1-192.168.200.24,192.168.200.1,192.168.200.1/24",
		IPAddresses:        "192.168.200.1/24",
		InheritanceAllowed: addrOf(false),
	}

	createdIpSet, err := vdc.CreateNsxvIpSet(ipSetConfig)
	check.Assert(err, IsNil)
	check.Assert(createdIpSet.ID, Not(Equals), "") // Validate that ID was set

	// Add to cleanup list
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name
	AddToCleanupList(createdIpSet.Name, "ipSet", parentEntity, check.TestName())

	// Check if the structure after creation is exactly the same, but with ID populated
	ipSetConfig.ID = createdIpSet.ID
	ipSetConfig.Revision = createdIpSet.Revision
	createdIpSet.XMLName.Local = ""
	check.Assert(createdIpSet, DeepEquals, ipSetConfig)

	// 2. Try to create another IP set with the same name and expect error
	_, err = vdc.CreateNsxvIpSet(ipSetConfig)
	check.Assert(err, ErrorMatches, ".*Another object with same name.* already exists in the current scope.*")

	// 3. Get all IP sets
	ipSets, err := vdc.GetAllNsxvIpSets()
	check.Assert(err, IsNil)
	check.Assert(len(ipSets) > 0, Equals, true)

	// 4. Get by Name, Id, NameOrId and check that all results are deeply equal
	ipSetByName, err := vdc.GetNsxvIpSetByName(createdIpSet.Name)
	check.Assert(err, IsNil)
	ipSetById, err := vdc.GetNsxvIpSetById(createdIpSet.ID)
	check.Assert(err, IsNil)
	ipSetByName2, err := vdc.GetNsxvIpSetByNameOrId(createdIpSet.Name)
	check.Assert(err, IsNil)
	ipSetById2, err := vdc.GetNsxvIpSetByNameOrId(createdIpSet.ID)
	check.Assert(err, IsNil)
	check.Assert(ipSetByName, DeepEquals, ipSetById)
	check.Assert(ipSetByName, DeepEquals, ipSetByName2)
	check.Assert(ipSetByName, DeepEquals, ipSetById2)

	// 5. Update IP set field
	createdIpSet.InheritanceAllowed = addrOf(true)
	updatedIpSet, err := vcd.vdc.UpdateNsxvIpSet(createdIpSet)
	check.Assert(err, IsNil)

	// Check that only this field was changed
	updatedIpSet.XMLName.Local = ""
	// Because revisions are auto-incremented - this must also be incremented so that it matches
	createdIpSet.Revision = updatedIpSet.Revision
	check.Assert(updatedIpSet, DeepEquals, createdIpSet)

	// 6. Delete created IP set by Id
	err = vdc.DeleteNsxvIpSetById(createdIpSet.ID)
	check.Assert(err, IsNil)

	// Verify that the IP set cannot be found by ID and by Name
	_, err = vdc.GetNsxvIpSetById(createdIpSet.ID)
	check.Assert(IsNotFound(err), Equals, true)

	_, err = vdc.GetNsxvIpSetByName(createdIpSet.Name)
	check.Assert(IsNotFound(err), Equals, true)

	// 7. Create another IP set and try to delete by name
	ipSet2, err := vdc.CreateNsxvIpSet(ipSetConfig)
	check.Assert(err, IsNil)

	err = vdc.DeleteNsxvIpSetByName(ipSet2.Name)
	check.Assert(err, IsNil)
}

// testCreateIpSet creates an IP set with given name and returns it which is useful in other tests
// when an IP set is needed to validate inputs.
func testCreateIpSet(name string, vdc *Vdc) (*types.EdgeIpSet, error) {
	ipSetConfig := &types.EdgeIpSet{
		Name:               name,
		Description:        "test-ipset-description",
		IPAddresses:        "192.168.200.1/24",
		InheritanceAllowed: addrOf(true),
	}

	return vdc.CreateNsxvIpSet(ipSetConfig)
}
