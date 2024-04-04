//go:build auth || functional || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

// Test_SamlAdfsAuth checks if SAML ADFS login works using WS-TRUST endpoint
//
//	"/adfs/services/trust/13/usernamemixed".
//
// Credential variables must be specified in test configuration for it to work
// The steps of this test are:
// * Query object using test framework vCD connection
// * Create a new client with SAML authentication using specified org and query the same object
// using it to make sure access is granted
// * Compare results to ensure that it worked as it should
//
// Note. This test requires real environment setup to work. Unit testing is also available in
// `saml_auth_unit_test.go`
func (vcd *TestVCD) Test_SamlAdfsAuth(check *C) {
	cfg := vcd.config
	if cfg.Provider.SamlUser == "" || cfg.Provider.SamlPassword == "" || cfg.VCD.Org == "" {
		check.Skip("Skipping test because no Org, SamlUser, SamlPassword and was specified")
	}
	vcd.checkSkipWhenApiToken(check)

	// Get vDC details using existing vCD client
	org, err := vcd.client.GetOrgByName(cfg.VCD.Org)
	check.Assert(err, IsNil)

	vdc, err := org.GetVDCByName(cfg.VCD.Vdc, true)
	check.Assert(err, IsNil)

	// Get new vCD session and client using specifically SAML credentials
	samlVcdCli := NewVCDClient(vcd.client.Client.VCDHREF, true,
		WithSamlAdfs(true, cfg.Provider.SamlCustomRptId))
	err = samlVcdCli.Authenticate(cfg.Provider.SamlUser, cfg.Provider.SamlPassword, cfg.VCD.Org)
	check.Assert(err, IsNil)

	samlOrg, err := vcd.client.GetOrgByName(cfg.VCD.Org)
	check.Assert(err, IsNil)

	samlVdc, err := samlOrg.GetVDCByName(cfg.VCD.Vdc, true)
	check.Assert(err, IsNil)

	check.Assert(samlVdc, DeepEquals, vdc)

	// If SamlCustomRptId was not specified - try to feed VCD entity ID manually (this is usually
	// done automatically, but doing it to test this path is not broken)
	if cfg.Provider.SamlCustomRptId == "" {
		samlEntityId, err := getSamlEntityId(vcd.client, cfg.VCD.Org)
		check.Assert(err, IsNil)

		samlCustomRptVcdCli := NewVCDClient(vcd.client.Client.VCDHREF, true,
			WithSamlAdfs(true, samlEntityId))
		err = samlCustomRptVcdCli.Authenticate(cfg.Provider.SamlUser, cfg.Provider.SamlPassword, cfg.VCD.Org)
		check.Assert(err, IsNil)

		samlCustomRptOrg, err := vcd.client.GetOrgByName(cfg.VCD.Org)
		check.Assert(err, IsNil)

		samlCustomRptVdc, err := samlCustomRptOrg.GetVDCByName(cfg.VCD.Vdc, true)
		check.Assert(err, IsNil)

		check.Assert(samlCustomRptVdc, DeepEquals, samlVdc)
	}

}
