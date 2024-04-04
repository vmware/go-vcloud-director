//go:build api || functional || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
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

func (vcd *TestVCD) Test_ApiTokenCreation(check *C) {
	isApiTokenEnabled, err := vcd.client.Client.VersionEqualOrGreater("10.3.1", 3)
	check.Assert(err, IsNil)
	if !isApiTokenEnabled {
		check.Skip("This test requires VCD 10.3.1 or greater")
	}
	client := vcd.client

	token, err := client.CreateToken(vcd.config.Provider.SysOrg, check.TestName())
	check.Assert(err, IsNil)
	check.Assert(token, NotNil)
	check.Assert(token.Token.Type, Equals, "REFRESH")
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTokens + token.Token.ID
	AddToCleanupListOpenApi(token.Token.Name, check.TestName(), endpoint)

	tokenInfo, err := token.GetInitialApiToken()
	check.Assert(err, IsNil)
	check.Assert(tokenInfo.AccessToken, Not(Equals), "")
	check.Assert(tokenInfo.TokenType, Equals, "Bearer")

	tokenInfo, err = token.GetInitialApiToken()
	check.Assert(err, NotNil)
	check.Assert(tokenInfo, IsNil)

	err = token.Delete()
	check.Assert(err, IsNil)

	notFound, err := client.GetTokenById(token.Token.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFound, IsNil)
}

func (vcd *TestVCD) Test_GetFilteredTokensSysOrg(check *C) {
	isApiTokenEnabled, err := vcd.client.Client.VersionEqualOrGreater("10.3.1", 3)
	check.Assert(err, IsNil)
	if !isApiTokenEnabled {
		check.Skip("This test requires VCD 10.3.1 or greater")
	}
	client := vcd.client
	if !client.Client.IsSysAdmin {
		check.Skip("This test requires to be run by a SysAdmin")
	}

	token, err := client.CreateToken(vcd.config.Provider.SysOrg, check.TestName())
	check.Assert(err, IsNil)
	check.Assert(token, NotNil)
	check.Assert(token.Token.Type, Equals, "REFRESH")
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTokens
	AddToCleanupListOpenApi(token.Token.Name, check.TestName(), endpoint+token.Token.ID)

	queryParameters := &url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("(name==%s;owner.name==%s;(type==PROXY,type==REFRESH))", check.TestName(), "administrator"))

	tokens, err := client.GetAllTokens(*queryParameters)
	check.Assert(err, IsNil)
	check.Assert(len(tokens), Equals, 1)
	check.Assert(tokens[0].Token.Name, Equals, check.TestName())
	check.Assert(tokens[0].Token.Owner.Name, Equals, "administrator")

	newToken, err := client.GetTokenByNameAndUsername(check.TestName(), "administrator")
	check.Assert(err, IsNil)
	check.Assert(newToken.Token.Name, Equals, check.TestName())
	check.Assert(newToken.Token.Owner.Name, Equals, "administrator")

	err = newToken.Delete()
	check.Assert(err, IsNil)

	newToken, err = client.GetTokenByNameAndUsername(check.TestName(), "administrator")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(newToken, IsNil)
}

func (vcd *TestVCD) Test_GetFilteredTokensOrg(check *C) {
	isApiTokenEnabled, err := vcd.client.Client.VersionEqualOrGreater("10.3.1", 3)
	check.Assert(err, IsNil)
	if !isApiTokenEnabled {
		check.Skip("This test requires VCD 10.3.1 or greater")
	}

	if vcd.config.Tenants == nil || len(vcd.config.Tenants) < 2 {
		check.Skip("no tenants found in configuration")
	}

	orgName := vcd.config.Tenants[0].SysOrg
	userName := vcd.config.Tenants[0].User
	password := vcd.config.Tenants[0].Password

	vcdClient1 := NewVCDClient(vcd.client.Client.VCDHREF, true)
	err = vcdClient1.Authenticate(userName, password, orgName)
	check.Assert(err, IsNil)

	token, err := vcdClient1.CreateToken(vcd.config.Tenants[0].SysOrg, check.TestName())
	check.Assert(err, IsNil)
	check.Assert(token, NotNil)
	check.Assert(token.Token.Type, Equals, "REFRESH")
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTokens
	AddToCleanupListOpenApi(token.Token.Name, check.TestName(), endpoint+token.Token.ID)

	queryParameters := &url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("(name==%s;owner.name==%s;(type==PROXY,type==REFRESH))", check.TestName(), userName))

	tokens, err := vcdClient1.GetAllTokens(*queryParameters)
	check.Assert(err, IsNil)
	check.Assert(len(tokens), Equals, 1)
	check.Assert(tokens[0].Token.Name, Equals, check.TestName())
	check.Assert(tokens[0].Token.Owner.Name, Equals, userName)

	newToken, err := vcdClient1.GetTokenByNameAndUsername(check.TestName(), userName)
	check.Assert(err, IsNil)
	check.Assert(newToken.Token.Name, Equals, check.TestName())
	check.Assert(newToken.Token.Owner.Name, Equals, userName)

	err = newToken.Delete()
	check.Assert(err, IsNil)

	newToken, err = vcdClient1.GetTokenByNameAndUsername(check.TestName(), userName)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(newToken, IsNil)
}
