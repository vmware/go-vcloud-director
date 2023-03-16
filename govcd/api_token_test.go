//go:build api || functional || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"os"

	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
)

// TestVCDClient_GetBearerTokenFromApiToken tests the token refresh operation
// To make it work, we need the following, or the test is skipped:
//   - VCD version 10.3.1 or greater
//   - environment variable TEST_VCD_API_TOKEN filled with a valid API token for that VCD
//   - If the API token was not set for the Organization defined in vcd.config.VCD.Org, the variable
//     TEST_VCD_ORG should be filled with the name of the Org for which the API token was set.
func (vcd *TestVCD) TestVCDClient_GetBearerTokenFromApiToken(check *C) {

	apiToken := os.Getenv("TEST_VCD_API_TOKEN")

	orgName := os.Getenv("TEST_VCD_ORG")
	if orgName == "" {
		orgName = vcd.config.VCD.Org
	}
	if orgName == "" {
		check.Skip("orgName not set")
	}
	if apiToken == "" {
		check.Skip(fmt.Sprintf("API token not set. Use TEST_VCD_API_TOKEN to indicate an API token for Org '%s'", orgName))
	}

	isApiTokenEnabled, err := vcd.client.Client.VersionEqualOrGreater("10.3.1", 3)
	check.Assert(err, IsNil)
	if !isApiTokenEnabled {
		check.Skip("This test requires VCD 10.3.1 or greater")
	}

	tokenInfo, err := vcd.client.GetBearerTokenFromApiToken(orgName, apiToken)
	check.Assert(err, IsNil)
	check.Assert(tokenInfo, NotNil)
	check.Assert(tokenInfo.AccessToken, Not(Equals), "")
	if testVerbose {
		fmt.Printf("%# v\n", pretty.Formatter(tokenInfo))
	}
	check.Assert(tokenInfo.ExpiresIn, Not(Equals), 0)
	check.Assert(tokenInfo.TokenType, Equals, "Bearer")
}
