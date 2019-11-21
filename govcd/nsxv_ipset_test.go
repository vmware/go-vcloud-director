// +build nsxv functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

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

	ipSetConfig := &types.EdgeIpSet{
		Name:        "test-ipset",
		Description: "test-ipset-description",
		// The below example demonstrates a problem when vCD API shuffles the below list and the
		// answer becomes ordered differently then it was submitted
		// IPAddresses:        "192.168.200.1-192.168.200.24,192.168.200.1,192.168.200.1/24",
		IPAddresses:        "192.168.200.1/24",
		InheritanceAllowed: takeBoolPointer(false),
	}

	createdIpSet, err := vdc.CreateNsxvIpSet(ipSetConfig)
	check.Assert(err, IsNil)
	check.Assert(createdIpSet.ID, Not(Equals), "") // Validate that ID was set
	// Check if the structure after creation is exactly the same, but with ID populated
	ipSetConfig.ID = createdIpSet.ID
	ipSetConfig.Revision = createdIpSet.Revision
	createdIpSet.XMLName.Local = ""
	check.Assert(createdIpSet, DeepEquals, ipSetConfig)

	// Get all IP sets
	ipSets, err := vdc.GetAllNsxvIpSets()
	check.Assert(err, IsNil)
	check.Assert(len(ipSets) > 0, Equals, true)

	// Update IP set field
	createdIpSet.InheritanceAllowed = takeBoolPointer(true)
	updatedIpSet, err := vcd.vdc.UpdateNsxvIpSet(createdIpSet)
	check.Assert(err, IsNil)

	// Check that only this field was changed
	updatedIpSet.XMLName.Local = ""
	// Because revisions are auto-incremented - this must also be incremented so that it matches
	createdIpSet.Revision = updatedIpSet.Revision
	check.Assert(updatedIpSet, DeepEquals, createdIpSet)

	// Delete created IP set
	err = vdc.DeleteNsxvIpSetById(createdIpSet.ID)
	check.Assert(err, IsNil)

	// Verify that the IP set cannot be found by ID and by Name
	_, err = vdc.GetNsxvIpSetById(createdIpSet.ID)
	check.Assert(IsNotFound(err), Equals, true)

	_, err = vdc.GetNsxvIpSetByName(createdIpSet.Name)
	check.Assert(IsNotFound(err), Equals, true)
}

func takeBoolPointer(value bool) *bool {
	return &value
}
