//go:build org || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	_ "embed"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"net/url"
	"strings"
	"time"
)

// Test_OrgOidcSettingsSystemAdminCreateWithWellKnownEndpoint configures OIDC
// with a wellknown endpoint.
func (vcd *TestVCD) Test_OrgOidcSettingsSystemAdminCreateWithWellKnownEndpoint(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}
	oidcServerUrl := validateAndGetOidcServerUrl(check, vcd)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := adminOrg.GetOpenIdConnectSettings()
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)
	check.Assert(settings.Enabled, Equals, false)
	check.Assert(settings.AccessTokenEndpoint, Equals, "")
	check.Assert(settings.UserInfoEndpoint, Equals, "")
	check.Assert(settings.UserAuthorizationEndpoint, Equals, "")
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")

	settings, err = setOIDCSettings(adminOrg, types.OrgOAuthSettings{
		ClientId:          "clientId",
		ClientSecret:      "clientSecret",
		Enabled:           true,
		MaxClockSkew:      60,
		WellKnownEndpoint: oidcServerUrl.String(),
	})
	check.Assert(err, IsNil)
	defer func() {
		deleteOIDCSettings(check, adminOrg)
	}()

	check.Assert(settings, NotNil)
	check.Assert(settings.Xmlns, Equals, "http://www.vmware.com/vcloud/v1.5")
	check.Assert(settings.Href, Equals, adminOrg.AdminOrg.HREF+"/settings/oauth")
	check.Assert(settings.Type, Equals, "application/vnd.vmware.admin.organizationOAuthSettings+xml")
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")
	check.Assert(settings.IssuerId, Not(Equals), "")
	check.Assert(settings.Enabled, Equals, true)
	check.Assert(settings.ClientId, Equals, "clientId")
	check.Assert(settings.ClientSecret, Equals, "clientSecret")
	check.Assert(settings.UserAuthorizationEndpoint, Not(Equals), "")
	check.Assert(settings.AccessTokenEndpoint, Not(Equals), "")
	check.Assert(settings.UserInfoEndpoint, Not(Equals), "")
	check.Assert(settings.ScimEndpoint, Equals, "")
	check.Assert(len(settings.Scope), Not(Equals), 0)
	check.Assert(settings.MaxClockSkew, Equals, 60)
	check.Assert(settings.WellKnownEndpoint, Not(Equals), "")
	check.Assert(settings.OIDCAttributeMapping, NotNil)
	check.Assert(settings.OAuthKeyConfigurations, NotNil)
	check.Assert(len(settings.OAuthKeyConfigurations.OAuthKeyConfiguration), Not(Equals), 0)
}

// Test_OrgOidcSettingsSystemAdminCreateWithWellKnownEndpointAndOverridingOptions configures OIDC
// with a wellknown endpoint, but overrides the obtained values with custom ones.
func (vcd *TestVCD) Test_OrgOidcSettingsSystemAdminCreateWithWellKnownEndpointAndOverridingOptions(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}
	oidcServerUrl := validateAndGetOidcServerUrl(check, vcd)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := adminOrg.GetOpenIdConnectSettings()
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")

	settings, err = setOIDCSettings(adminOrg, types.OrgOAuthSettings{
		ClientId:                  "clientId",
		ClientSecret:              "clientSecret",
		Enabled:                   true,
		MaxClockSkew:              60,
		AccessTokenEndpoint:       oidcServerUrl.Host + "/foo",
		UserAuthorizationEndpoint: oidcServerUrl.Host + "/foo2",
		WellKnownEndpoint:         oidcServerUrl.String(),
	})
	check.Assert(err, IsNil)
	defer func() {
		deleteOIDCSettings(check, adminOrg)
	}()

	check.Assert(settings, NotNil)
	check.Assert(settings.AccessTokenEndpoint, Equals, oidcServerUrl.Host+"/foo")
	check.Assert(settings.UserAuthorizationEndpoint, Equals, oidcServerUrl.Host+"/foo2")
	check.Assert(settings.Xmlns, Equals, "http://www.vmware.com/vcloud/v1.5")
	check.Assert(settings.Href, Equals, adminOrg.AdminOrg.HREF+"/settings/oauth")
	check.Assert(settings.Type, Equals, "application/vnd.vmware.admin.organizationOAuthSettings+xml")
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")
	check.Assert(settings.IssuerId, Not(Equals), "")
	check.Assert(settings.Enabled, Equals, true)
	check.Assert(settings.ClientId, Equals, "clientId")
	check.Assert(settings.ClientSecret, Equals, "clientSecret")
	check.Assert(settings.UserInfoEndpoint, Not(Equals), "")
	check.Assert(settings.ScimEndpoint, Equals, "")
	check.Assert(len(settings.Scope), Not(Equals), 0)
	check.Assert(settings.MaxClockSkew, Equals, 60)
	check.Assert(settings.WellKnownEndpoint, Not(Equals), "")
	check.Assert(settings.OIDCAttributeMapping, NotNil)
	check.Assert(settings.OAuthKeyConfigurations, NotNil)
	check.Assert(len(settings.OAuthKeyConfigurations.OAuthKeyConfiguration), Not(Equals), 0)
}

// Test_OrgOidcSettingsSystemAdminCreateWithCustomValues configures OIDC
// without the wellknown endpoint, by hand.
func (vcd *TestVCD) Test_OrgOidcSettingsSystemAdminCreateWithCustomValues(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}

	oidcServerUrl := validateAndGetOidcServerUrl(check, vcd)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := setOIDCSettings(adminOrg, types.OrgOAuthSettings{
		ClientId:                  "clientId",
		ClientSecret:              "clientSecret",
		Enabled:                   true,
		UserAuthorizationEndpoint: oidcServerUrl.Host + "/userAuth",
		AccessTokenEndpoint:       oidcServerUrl.Host + "/accessToken",
		IssuerId:                  oidcServerUrl.Host + "/issuerId",
		UserInfoEndpoint:          oidcServerUrl.Host + "/userInfo",
		MaxClockSkew:              60,
		Scope:                     []string{"foo", "bar"},
		OIDCAttributeMapping: &types.OIDCAttributeMapping{
			SubjectAttributeName:   "subject",
			EmailAttributeName:     "email",
			FullNameAttributeName:  "fullname",
			FirstNameAttributeName: "first",
			LastNameAttributeName:  "last",
			GroupsAttributeName:    "groups",
			RolesAttributeName:     "roles",
		},
		OAuthKeyConfigurations: &types.OAuthKeyConfigurationsList{
			OAuthKeyConfiguration: []types.OAuthKeyConfiguration{
				{
					KeyId:          "rsa1",
					Algorithm:      "RSA",
					Key:            "-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC9gXitSASYbVS56gBkQ3UOCS7F\n8SnFABs44sxXykt8DW4y1mxdyCcM0X/lVPf+DNfXbIISmPk/mqoRS9uZSuQIUtC2\n4iaGkWyUALvrq8FJcR8Krf5EtDt1W9AkLEREDJ7VkpJx/VoCd9ZNe8NFstAvbQ6+\nbM0Jg9lJJdr+VPNvywIDAQAB\n-----END PUBLIC KEY-----",
					ExpirationDate: "",
				},
			},
		},
	})
	check.Assert(err, IsNil)
	defer func() {
		deleteOIDCSettings(check, adminOrg)
	}()

	check.Assert(settings, NotNil)
	check.Assert(settings.Xmlns, Equals, "http://www.vmware.com/vcloud/v1.5")
	check.Assert(settings.Href, Equals, adminOrg.AdminOrg.HREF+"/settings/oauth")
	check.Assert(settings.Type, Equals, "application/vnd.vmware.admin.organizationOAuthSettings+xml")
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")
	check.Assert(settings.Enabled, Equals, true)
	check.Assert(settings.ClientId, Equals, "clientId")
	check.Assert(settings.ClientSecret, Equals, "clientSecret")
	check.Assert(settings.IssuerId, Equals, oidcServerUrl.Host+"/issuerId")
	check.Assert(settings.UserAuthorizationEndpoint, Equals, oidcServerUrl.Host+"/userAuth")
	check.Assert(settings.AccessTokenEndpoint, Equals, oidcServerUrl.Host+"/accessToken")
	check.Assert(settings.UserInfoEndpoint, Equals, oidcServerUrl.Host+"/userInfo")
	check.Assert(settings.ScimEndpoint, Equals, "")
	check.Assert(len(settings.Scope), Equals, 2)
	check.Assert(settings.MaxClockSkew, Equals, 60)
	check.Assert(settings.WellKnownEndpoint, Equals, "")
	check.Assert(settings.OIDCAttributeMapping, NotNil)
	check.Assert(settings.OIDCAttributeMapping.EmailAttributeName, Equals, "email")
	check.Assert(settings.OIDCAttributeMapping.LastNameAttributeName, Equals, "last")
	check.Assert(settings.OIDCAttributeMapping.FirstNameAttributeName, Equals, "first")
	check.Assert(settings.OIDCAttributeMapping.SubjectAttributeName, Equals, "subject")
	check.Assert(settings.OIDCAttributeMapping.GroupsAttributeName, Equals, "groups")
	check.Assert(settings.OIDCAttributeMapping.FullNameAttributeName, Equals, "fullname")
	check.Assert(settings.OIDCAttributeMapping.RolesAttributeName, Equals, "roles")
	check.Assert(settings.OAuthKeyConfigurations, NotNil)
	check.Assert(len(settings.OAuthKeyConfigurations.OAuthKeyConfiguration), Equals, 1)
	check.Assert(settings.OAuthKeyConfigurations.OAuthKeyConfiguration[0].KeyId, Equals, "rsa1")
	check.Assert(settings.OAuthKeyConfigurations.OAuthKeyConfiguration[0].Algorithm, Equals, "RSA")
	check.Assert(true, Equals, strings.Contains(settings.OAuthKeyConfigurations.OAuthKeyConfiguration[0].Key, "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC9gXitSASYbVS56gBkQ3UOCS7F"))
}

// Test_OrgOidcSettingsSystemAdminUpdate configures OIDC settings with a wellknown endpoint, then updates some values.
func (vcd *TestVCD) Test_OrgOidcSettingsSystemAdminUpdate(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}

	oidcServerUrl := validateAndGetOidcServerUrl(check, vcd)

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := adminOrg.GetOpenIdConnectSettings()
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)
	check.Assert(settings.Enabled, Equals, false)
	check.Assert(settings.AccessTokenEndpoint, Equals, "")
	check.Assert(settings.UserInfoEndpoint, Equals, "")
	check.Assert(settings.UserAuthorizationEndpoint, Equals, "")
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")

	settings, err = setOIDCSettings(adminOrg, types.OrgOAuthSettings{
		ClientId:          "clientId",
		ClientSecret:      "clientSecret",
		Enabled:           true,
		MaxClockSkew:      60,
		WellKnownEndpoint: oidcServerUrl.String(),
	})
	check.Assert(err, IsNil)
	defer func() {
		deleteOIDCSettings(check, adminOrg)
	}()
	check.Assert(settings, NotNil)

	updatedSettings, err := setOIDCSettings(adminOrg, types.OrgOAuthSettings{
		ClientId:     "clientId2",
		ClientSecret: "clientSecret2",
		Enabled:      false,
		MaxClockSkew: 120,
		OIDCAttributeMapping: &types.OIDCAttributeMapping{
			SubjectAttributeName:   "subject2",
			EmailAttributeName:     "email2",
			FullNameAttributeName:  "fullname2",
			FirstNameAttributeName: "first2",
			LastNameAttributeName:  "last2",
			GroupsAttributeName:    "groups2",
			RolesAttributeName:     "roles2",
		},
		WellKnownEndpoint: oidcServerUrl.String(),
	})
	check.Assert(err, IsNil)
	check.Assert(updatedSettings, NotNil)

	check.Assert(updatedSettings.Enabled, Equals, false)
	check.Assert(updatedSettings.ClientId, Equals, "clientId2")
	check.Assert(updatedSettings.ClientSecret, Equals, "clientSecret2")
	check.Assert(updatedSettings.MaxClockSkew, Equals, 120)
	check.Assert(updatedSettings.OIDCAttributeMapping, NotNil)
	check.Assert(updatedSettings.OIDCAttributeMapping.EmailAttributeName, Equals, "email2")
	check.Assert(updatedSettings.OIDCAttributeMapping.LastNameAttributeName, Equals, "last2")
	check.Assert(updatedSettings.OIDCAttributeMapping.FirstNameAttributeName, Equals, "first2")
	check.Assert(updatedSettings.OIDCAttributeMapping.SubjectAttributeName, Equals, "subject2")
	check.Assert(updatedSettings.OIDCAttributeMapping.GroupsAttributeName, Equals, "groups2")
	check.Assert(updatedSettings.OIDCAttributeMapping.FullNameAttributeName, Equals, "fullname2")
	check.Assert(updatedSettings.OIDCAttributeMapping.RolesAttributeName, Equals, "roles2")
}

func (vcd *TestVCD) Test_OrgOidcSettingsTenantCRUD(check *C) {
	_ = validateAndGetOidcServerUrl(check, vcd)

}

// Test_OrgOidcSettingsValidationErrors tests the validation rules when setting OpenID Connect Settings with AdminOrg.SetOpenIdConnectSettings
func (vcd *TestVCD) Test_OrgOidcSettingsValidationErrors(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

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
func setOIDCSettings(adminOrg *AdminOrg, settings types.OrgOAuthSettings) (*types.OrgOAuthSettings, error) {
	tries := 0
	var newSettings *types.OrgOAuthSettings
	var err error
	for tries < 3 {
		tries++
		newSettings, err = adminOrg.SetOpenIdConnectSettings(settings)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "could not establish a connection") || strings.Contains(err.Error(), "connect timed out") {
			time.Sleep(5 * time.Second)
		}
	}
	if err != nil {
		return nil, err
	}
	return newSettings, nil
}

// deleteOIDCSettings deletes the current OIDC settings for the given Organization
func deleteOIDCSettings(check *C, adminOrg *AdminOrg) {
	err := adminOrg.DeleteOpenIdConnectSettings()
	check.Assert(err, IsNil)

	settings, err := adminOrg.GetOpenIdConnectSettings()
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)
	check.Assert(settings.Enabled, Equals, false)
	check.Assert(settings.AccessTokenEndpoint, Equals, "")
	check.Assert(settings.UserInfoEndpoint, Equals, "")
	check.Assert(settings.UserAuthorizationEndpoint, Equals, "")
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")
}

func validateAndGetOidcServerUrl(check *C, vcd *TestVCD) *url.URL {
	if vcd.config.VCD.OidcServer.Url == "" || vcd.config.VCD.OidcServer.WellKnownEndpoint == "" {
		check.Skip("test requires OIDC configuration")
	}

	oidcServer, err := url.Parse(vcd.config.VCD.OidcServer.Url)
	if err != nil {
		check.Skip(check.TestName() + " requires OIDC Server URL and its well-known endpoint")
	}
	return oidcServer.JoinPath(vcd.config.VCD.OidcServer.WellKnownEndpoint)
}
