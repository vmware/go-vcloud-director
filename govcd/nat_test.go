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
		// RuleType: "user",
		// Name:                        "asd",
		Action:                      "snat", // mandatory
		// OriginalAddress:             "10.10.10.15",
		TranslatedAddress:           "192.168.1.110",	// mandatory
		// SnatMatchDestinationAddress: "any",
		Enabled:                     "true",	// mandatory
		// LoggingEnabled:              "false",
		// OriginalPort:                "3380",
		// TranslatedPort:              "3380",
		// Protocol: "any",
		// Vnic: "0",
	}

	_, err = edge.CreateSnatRule(natRule)
	check.Assert(err, IsNil)

}
