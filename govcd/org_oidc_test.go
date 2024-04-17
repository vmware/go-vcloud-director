//go:build org || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	_ "embed"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"strings"
	"time"
)

func (vcd *TestVCD) Test_OrgOidcSettingsCRUD(check *C) {
	//orgName := check.TestName()
	//
	//task, err := CreateOrg(vcd.client, orgName, orgName, orgName, &types.OrgSettings{}, true)
	//check.Assert(err, IsNil)
	//check.Assert(task, NotNil)
	//AddToCleanupList(orgName, "org", "", check.TestName())
	//err = task.WaitTaskCompletion()
	//check.Assert(err, IsNil)

	oidcServerUrl := fmt.Sprintf("http://%s:8080/stf-oidc-server", vcd.config.VCD.LdapServer)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := adminOrg.GetOpenIdConnectSettings()
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")

	testFailures(check, adminOrg, oidcServerUrl)

	// To avoid test failures due to bad connectivity with the OIDC Provider server, we put some retries in place
	tries := 0
	for tries < 3 {
		tries++
		settings, err = adminOrg.SetOpenIdConnectSettings(types.OrgOAuthSettings{
			ClientId:          addrOf("a"),
			ClientSecret:      addrOf("b"),
			Enabled:           addrOf(true),
			WellKnownEndpoint: addrOf(oidcServerUrl + "/.well-known/openid-configuration"),
		})
		if err != nil {
			check.Assert(true, Equals, strings.Contains(err.Error(), "could not establish a connection") || strings.Contains(err.Error(), "connect timed out"))
		}
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)

	err = adminOrg.DeleteOpenIdConnectSettings()
	check.Assert(err, IsNil)

	//err = adminOrg.Delete(true, true)
	//check.Assert(err, IsNil)
}

func testFailures(check *C, adminOrg *AdminOrg, oidcServerUrl string) {
	tests := []struct {
		wrongConfig types.OrgOAuthSettings
		errorMsg    string
	}{
		{
			wrongConfig: types.OrgOAuthSettings{},
			errorMsg:    "the Client ID is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId: addrOf("clientId"),
			},
			errorMsg: "the Client Secret is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:     addrOf("clientId"),
				ClientSecret: addrOf("clientSecret"),
			},
			errorMsg: "the OpenID Connect input settings must specify either Enabled=true or Enabled=false",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:     addrOf("clientId"),
				ClientSecret: addrOf("clientSecret"),
				Enabled:      addrOf(true),
			},
			errorMsg: "the User Authorization Endpoint is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf(oidcServerUrl + "/authorize"),
			},
			errorMsg: "the Access Token Endpoint is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf(oidcServerUrl + "/authorize"),
				AccessTokenEndpoint:       addrOf(oidcServerUrl + "/token"),
			},
			errorMsg: "the User Info Endpoint is mandatory to configure OpenID Connect",
		},
	}

	for _, test := range tests {
		_, err := adminOrg.SetOpenIdConnectSettings(test.wrongConfig)
		check.Assert(err, NotNil)
		check.Assert(true, Equals, strings.Contains(err.Error(), test.errorMsg))
	}

}
