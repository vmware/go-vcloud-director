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

	settings = setOIDCSettings(check, adminOrg, types.OrgOAuthSettings{
		ClientId:          "clientId",
		ClientSecret:      "clientSecret",
		Enabled:           true,
		MaxClockSkew:      60,
		WellKnownEndpoint: vcd.config.VCD.OidcServer.Url + vcd.config.VCD.OidcServer.WellKnownEndpoint,
	})
	check.Assert(settings.WellKnownEndpoint, NotNil)

	// Be sure that the settings are always deleted
	defer func() {
		err = adminOrg.DeleteOpenIdConnectSettings()
		check.Assert(err, IsNil)
	}()

	err = adminOrg.DeleteOpenIdConnectSettings()
	check.Assert(err, IsNil)

	// Re-configure manually, without the well-known endpoint
	newSettings := setOIDCSettings(check, adminOrg, types.OrgOAuthSettings{
		ClientId:     settings.ClientId,
		ClientSecret: settings.ClientSecret,

		UserAuthorizationEndpoint: settings.UserAuthorizationEndpoint,
		AccessTokenEndpoint:       settings.AccessTokenEndpoint,
		IssuerId:                  settings.IssuerId,
		UserInfoEndpoint:          settings.UserInfoEndpoint,
		MaxClockSkew:              60,
		Scope:                     settings.Scope,
		OIDCAttributeMapping:      settings.OIDCAttributeMapping,
		OAuthKeyConfigurations:    settings.OAuthKeyConfigurations,
	})
	check.Assert(newSettings.WellKnownEndpoint, NotNil)

	// Reconfigure without deleting
	newSettings = setOIDCSettings(check, adminOrg, types.OrgOAuthSettings{
		ClientId:     "changedClientId",
		ClientSecret: settings.ClientSecret,

		UserAuthorizationEndpoint: settings.UserAuthorizationEndpoint,
		AccessTokenEndpoint:       settings.AccessTokenEndpoint,
		IssuerId:                  settings.IssuerId,
		UserInfoEndpoint:          settings.UserInfoEndpoint,
		MaxClockSkew:              60,
		Scope:                     settings.Scope,
		OIDCAttributeMapping:      settings.OIDCAttributeMapping,
		OAuthKeyConfigurations:    settings.OAuthKeyConfigurations,
	})
	check.Assert(newSettings.ClientId, Equals, "changedClientId")

	// Disable OIDC
	newSettings = setOIDCSettings(check, adminOrg, types.OrgOAuthSettings{
		Enabled:                   false,
		ClientId:                  "changedClientId",
		ClientSecret:              settings.ClientSecret,
		UserAuthorizationEndpoint: settings.UserAuthorizationEndpoint,
		AccessTokenEndpoint:       settings.AccessTokenEndpoint,
		IssuerId:                  settings.IssuerId,
		UserInfoEndpoint:          settings.UserInfoEndpoint,
		MaxClockSkew:              60,
		Scope:                     settings.Scope,
		OIDCAttributeMapping:      settings.OIDCAttributeMapping,
		OAuthKeyConfigurations:    settings.OAuthKeyConfigurations,
	})
	check.Assert(newSettings.Enabled, Equals, false)

	err = adminOrg.DeleteOpenIdConnectSettings()
	check.Assert(err, IsNil)
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
				ClientId: "clientId",
			},
			errorMsg: "the Client Secret is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:     "clientId",
				ClientSecret: "clientSecret",
			},
			errorMsg: "the User Authorization Endpoint is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
			},
			errorMsg: "the Access Token Endpoint is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
			},
			errorMsg: "the User Info Endpoint is mandatory to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
				UserInfoEndpoint:          "https://dummy.url/userinfo",
				MaxClockSkew:              -1,
			},
			errorMsg: "the Max Clock Skew must be positive to correctly configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
				UserInfoEndpoint:          "https://dummy.url/userinfo",
				MaxClockSkew:              60,
				OIDCAttributeMapping:      &types.OIDCAttributeMapping{},
			},
			errorMsg: "the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
				UserInfoEndpoint:          "https://dummy.url/userinfo",
				MaxClockSkew:              60,
				OIDCAttributeMapping: &types.OIDCAttributeMapping{
					SubjectAttributeName: "a",
				},
			},
			errorMsg: "the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
				UserInfoEndpoint:          "https://dummy.url/userinfo",
				MaxClockSkew:              60,
				OIDCAttributeMapping: &types.OIDCAttributeMapping{
					SubjectAttributeName: "a",
					EmailAttributeName:   "b",
				},
			},
			errorMsg: "the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect",
		},
		{
			wrongConfig: types.OrgOAuthSettings{
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
				UserInfoEndpoint:          "https://dummy.url/userinfo",
				MaxClockSkew:              60,
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
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
				UserInfoEndpoint:          "https://dummy.url/userinfo",
				MaxClockSkew:              60,
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
				ClientId:     "clientId",
				ClientSecret: "clientSecret",

				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
				UserInfoEndpoint:          "https://dummy.url/userinfo",
				MaxClockSkew:              60,
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
				ClientId:                  "clientId",
				ClientSecret:              "clientSecret",
				UserAuthorizationEndpoint: "https://dummy.url/authorize",
				AccessTokenEndpoint:       "https://dummy.url/token",
				UserInfoEndpoint:          "https://dummy.url/userinfo",
				MaxClockSkew:              60,
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

// setOIDCSettings sets the given OIDC settings for the given Organization. It does this operation
// with some tries to avoid test failures due to network glitches.
func setOIDCSettings(check *C, adminOrg *AdminOrg, settings types.OrgOAuthSettings) *types.OrgOAuthSettings {
	tries := 0
	var newSettings *types.OrgOAuthSettings
	var err error
	for tries < 3 {
		tries++
		newSettings, err = adminOrg.SetOpenIdConnectSettings(settings)
		if err != nil {
			check.Assert(true, Equals, strings.Contains(err.Error(), "could not establish a connection") || strings.Contains(err.Error(), "connect timed out"))
		}
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	check.Assert(err, IsNil)
	check.Assert(newSettings, NotNil)
	return newSettings
}
