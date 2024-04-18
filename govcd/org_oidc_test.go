//go:build org || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	_ "embed"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"strings"
	"time"
)

func (vcd *TestVCD) Test_OrgOidcSettingsSystemAdminCRUD(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}
	if vcd.config.VCD.OidcServer.Url == "" || vcd.config.VCD.OidcServer.WellKnownEndpoint == "" {
		check.Skip("test requires OIDC configuration")
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := adminOrg.GetOpenIdConnectSettings()
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")

	testValidationErrors(check, adminOrg)

	// To avoid test failures due to bad connectivity with the OIDC Provider server, we put some retries in place
	tries := 0
	for tries < 3 {
		tries++
		settings, err = adminOrg.SetOpenIdConnectSettings(types.OrgOAuthSettings{
			ClientId:          addrOf("a"),
			ClientSecret:      addrOf("b"),
			Enabled:           addrOf(true),
			WellKnownEndpoint: addrOf(vcd.config.VCD.OidcServer.Url + vcd.config.VCD.OidcServer.WellKnownEndpoint),
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

	// Be sure that the settings are always deleted
	defer func() {
		err = adminOrg.DeleteOpenIdConnectSettings()
		check.Assert(err, IsNil)
	}()

}

func (vcd *TestVCD) Test_OrgOidcSettingsTenantCRUD(check *C) {
	if vcd.config.VCD.OidcServer.Url == "" || vcd.config.VCD.OidcServer.WellKnownEndpoint == "" {
		check.Skip("test requires OIDC configuration")
	}
}

// testValidationErrors tests the validation rules when setting OpenID Connect Settings with AdminOrg.SetOpenIdConnectSettings
func testValidationErrors(check *C, adminOrg *AdminOrg) {
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
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
			},
			errorMsg: "the Access Token Endpoint is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
			},
			errorMsg: "the User Info Endpoint is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
			},
			errorMsg: "the Max Clock Skew is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
				MaxClockSkew:              addrOf(-1),
			},
			errorMsg: "the Max Clock Skew is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
				MaxClockSkew:              addrOf(60),
				OIDCAttributeMapping:      &types.OIDCAttributeMapping{},
			},
			errorMsg: "the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
				MaxClockSkew:              addrOf(60),
				OIDCAttributeMapping: &types.OIDCAttributeMapping{
					SubjectAttributeName: "a",
				},
			},
			errorMsg: "the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
				MaxClockSkew:              addrOf(60),
				OIDCAttributeMapping: &types.OIDCAttributeMapping{
					SubjectAttributeName: "a",
					EmailAttributeName:   "b",
				},
			},
			errorMsg: "the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
				MaxClockSkew:              addrOf(60),
				OIDCAttributeMapping: &types.OIDCAttributeMapping{
					SubjectAttributeName:  "a",
					EmailAttributeName:    "b",
					FullNameAttributeName: "c",
				},
			},
			errorMsg: "the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
				MaxClockSkew:              addrOf(60),
				OIDCAttributeMapping: &types.OIDCAttributeMapping{
					SubjectAttributeName:   "a",
					EmailAttributeName:     "b",
					FullNameAttributeName:  "c",
					FirstNameAttributeName: "d",
				},
			},
			errorMsg: "the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
				MaxClockSkew:              addrOf(60),
				OIDCAttributeMapping: &types.OIDCAttributeMapping{
					SubjectAttributeName:   "a",
					EmailAttributeName:     "b",
					FullNameAttributeName:  "c",
					FirstNameAttributeName: "d",
					LastNameAttributeName:  "e",
				},
			},
			errorMsg: "the OIDC Key Configuration is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  addrOf("clientId"),
				ClientSecret:              addrOf("clientSecret"),
				Enabled:                   addrOf(true),
				UserAuthorizationEndpoint: addrOf("https://dummy.url/authorize"),
				AccessTokenEndpoint:       addrOf("https://dummy.url/token"),
				UserInfoEndpoint:          addrOf("https://dummy.url/userinfo"),
				MaxClockSkew:              addrOf(60),
				OIDCAttributeMapping: &types.OIDCAttributeMapping{
					SubjectAttributeName:   "a",
					EmailAttributeName:     "b",
					FullNameAttributeName:  "c",
					FirstNameAttributeName: "d",
					LastNameAttributeName:  "e",
				},
				OAuthKeyConfigurations: &types.OAuthKeyConfigurationsList{},
			},
			errorMsg: "the OIDC Key Configuration is mandatory to configure OpenID Connect",
		},
	}

	for _, test := range tests {
		_, err := adminOrg.SetOpenIdConnectSettings(test.wrongConfig)
		check.Assert(err, NotNil)
		check.Assert(true, Equals, strings.Contains(err.Error(), test.errorMsg))
	}

}
