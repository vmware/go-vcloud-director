// +build edge nat functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NatRule(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	natRule := &types.EdgeSnatRule{
		Action:            "snat",
		TranslatedAddress: vcd.config.VCD.ExternalIp, // Edge gateway address
		OriginalAddress:   vcd.config.VCD.InternalIp,
		Enabled:           "true",
		// RuleType: "user",
		// Name:                        "asd",
		// SnatMatchDestinationAddress: "any",
		// LoggingEnabled:              "false",
		// OriginalPort:                "3380",
		// TranslatedPort:              "3380",
		// Protocol: "any",
		// Vnic: "0",
	}

	createdSnatRule, err := edge.CreateSnatRule(natRule)
	check.Assert(err, IsNil)

	gotSnatRule, err := edge.GetSnatRuleById(createdSnatRule.ID)
	check.Assert(err, IsNil)
	check.Assert(gotSnatRule.ID, Equals, createdSnatRule.ID)

	// Set ID and update nat rule with description
	natRule.ID = gotSnatRule.ID
	natRule.Description = "Description for SNAT rule"
	updatedSnatRule, err := edge.UpdateSnatRule(natRule)
	check.Assert(err, IsNil)
	check.Assert(updatedSnatRule.Description, Equals, natRule.Description)


	err = edge.DeleteSnatRuleById(gotSnatRule.ID)
	check.Assert(err, IsNil)

}
