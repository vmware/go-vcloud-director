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
	check.Assert(true, Equals, strings.HasSuffix(settings.OrgRedirectUri, vcd.config.VCD.Org))

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
	check.Assert(true, Equals, strings.HasSuffix(settings.OrgRedirectUri, vcd.config.VCD.Org))
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
	check.Assert(true, Equals, strings.HasSuffix(settings.OrgRedirectUri, vcd.config.VCD.Org))

	accessTokenEndpoint := fmt.Sprintf("%s://%s/foo", oidcServerUrl.Scheme, oidcServerUrl.Host)
	userAuthorizationEndpoint := fmt.Sprintf("%s://%s/foo2", oidcServerUrl.Scheme, oidcServerUrl.Host)

	settings, err = setOIDCSettings(adminOrg, types.OrgOAuthSettings{
		ClientId:                  "clientId",
		ClientSecret:              "clientSecret",
		Enabled:                   true,
		MaxClockSkew:              60,
		AccessTokenEndpoint:       accessTokenEndpoint,
		UserAuthorizationEndpoint: userAuthorizationEndpoint,
		WellKnownEndpoint:         oidcServerUrl.String(),
	})
	check.Assert(err, IsNil)
	defer func() {
		deleteOIDCSettings(check, adminOrg)
	}()

	check.Assert(settings, NotNil)
	check.Assert(settings.AccessTokenEndpoint, Equals, accessTokenEndpoint)
	check.Assert(settings.UserAuthorizationEndpoint, Equals, userAuthorizationEndpoint)
	check.Assert(settings.Xmlns, Equals, "http://www.vmware.com/vcloud/v1.5")
	check.Assert(settings.Href, Equals, adminOrg.AdminOrg.HREF+"/settings/oauth")
	check.Assert(settings.Type, Equals, "application/vnd.vmware.admin.organizationOAuthSettings+xml")
	check.Assert(true, Equals, strings.HasSuffix(settings.OrgRedirectUri, vcd.config.VCD.Org))
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

	accessTokenEndpoint := fmt.Sprintf("%s://%s/accessToken", oidcServerUrl.Scheme, oidcServerUrl.Host)
	userAuthorizationEndpoint := fmt.Sprintf("%s://%s/userAuth", oidcServerUrl.Scheme, oidcServerUrl.Host)
	issuerId := fmt.Sprintf("%s://%s/issuerId", oidcServerUrl.Scheme, oidcServerUrl.Host)
	userInfoEndpoint := fmt.Sprintf("%s://%s/userInfo", oidcServerUrl.Scheme, oidcServerUrl.Host)

	expirationDate := "2123-12-31T01:59:59.000Z"
	dummyKey := "-----BEGIN PUBLIC KEY-----\n" +
		"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC9gXitSASYbVS56gBkQ3UOCS7F\n" +
		"8SnFABs44sxXykt8DW4y1mxdyCcM0X/lVPf+DNfXbIISmPk/mqoRS9uZSuQIUtC2\n" +
		"4iaGkWyUALvrq8FJcR8Krf5EtDt1W9AkLEREDJ7VkpJx/VoCd9ZNe8NFstAvbQ6+\n" +
		"bM0Jg9lJJdr+VPNvywIDAQAB\n" +
		"-----END PUBLIC KEY-----"

	settings, err := setOIDCSettings(adminOrg, types.OrgOAuthSettings{
		ClientId:                  "clientId",
		ClientSecret:              "clientSecret",
		Enabled:                   true,
		UserAuthorizationEndpoint: userAuthorizationEndpoint,
		AccessTokenEndpoint:       accessTokenEndpoint,
		IssuerId:                  issuerId,
		UserInfoEndpoint:          userInfoEndpoint,
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
					Key:            dummyKey,
					ExpirationDate: expirationDate,
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
	check.Assert(true, Equals, strings.HasSuffix(settings.OrgRedirectUri, vcd.config.VCD.Org))
	check.Assert(settings.Enabled, Equals, true)
	check.Assert(settings.ClientId, Equals, "clientId")
	check.Assert(settings.ClientSecret, Equals, "clientSecret")
	check.Assert(settings.IssuerId, Equals, issuerId)
	check.Assert(settings.UserAuthorizationEndpoint, Equals, userAuthorizationEndpoint)
	check.Assert(settings.AccessTokenEndpoint, Equals, accessTokenEndpoint)
	check.Assert(settings.UserInfoEndpoint, Equals, userInfoEndpoint)
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
	check.Assert(settings.OAuthKeyConfigurations.OAuthKeyConfiguration[0].Key, Equals, dummyKey)
	check.Assert(settings.OAuthKeyConfigurations.OAuthKeyConfiguration[0].ExpirationDate, Equals, expirationDate)
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
	check.Assert(true, Equals, strings.HasSuffix(settings.OrgRedirectUri, vcd.config.VCD.Org))

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

// Test_OrgOidcSettingsWithTenantUser configures OIDC settings with a tenant user instead of System administrator.
func (vcd *TestVCD) Test_OrgOidcSettingsWithTenantUser(check *C) {
	if len(vcd.config.Tenants) == 0 {
		check.Skip(check.TestName() + " requires at least one tenant in the configuration")
	}

	oidcServerUrl := validateAndGetOidcServerUrl(check, vcd)

	orgName := vcd.config.Tenants[0].SysOrg
	userName := vcd.config.Tenants[0].User
	password := vcd.config.Tenants[0].Password

	vcdClient := NewVCDClient(vcd.client.Client.VCDHREF, true)
	err := vcdClient.Authenticate(userName, password, orgName)
	check.Assert(err, IsNil)

	adminOrg, err := vcd.client.GetAdminOrgByName(orgName)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := setOIDCSettings(adminOrg, types.OrgOAuthSettings{
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
	check.Assert(true, Equals, strings.HasSuffix(settings.OrgRedirectUri, orgName))
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

	settings2, err := adminOrg.GetOpenIdConnectSettings()
	check.Assert(err, IsNil)
	check.Assert(settings2, DeepEquals, settings)
}

// Test_OrgOidcSettingsDifferentVersions tests the parameters that are only available in certain
// VCD versions, like the UI button label. This test only makes sense when it is run in several
// VCD versions.
func (vcd *TestVCD) Test_OrgOidcSettingsDifferentVersions(check *C) {
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
	check.Assert(true, Equals, strings.HasSuffix(settings.OrgRedirectUri, vcd.config.VCD.Org))

	s := types.OrgOAuthSettings{
		ClientId:          "clientId",
		ClientSecret:      "clientSecret",
		Enabled:           true,
		MaxClockSkew:      60,
		WellKnownEndpoint: oidcServerUrl.String(),
	}
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.1") {
		s.EnableIdTokenClaims = addrOf(true)
	}
	if vcd.client.Client.APIVCDMaxVersionIs(">= 38.0") {
		s.SendClientCredentialsAsAuthorizationHeader = addrOf(true)
		s.UsePKCE = addrOf(true)
	}
	if vcd.client.Client.APIVCDMaxVersionIs(">= 38.1") {
		s.CustomUiButtonLabel = addrOf("this is a test")
	}

	settings, err = setOIDCSettings(adminOrg, s)
	check.Assert(err, IsNil)
	defer func() {
		deleteOIDCSettings(check, adminOrg)
	}()

	check.Assert(settings, NotNil)
	if vcd.client.Client.APIVCDMaxVersionIs(">= 37.1") {
		check.Assert(settings.EnableIdTokenClaims, NotNil)
		check.Assert(*settings.EnableIdTokenClaims, Equals, true)
	} else {
		check.Assert(settings.EnableIdTokenClaims, IsNil)
	}
	if vcd.client.Client.APIVCDMaxVersionIs(">= 38.0") {
		check.Assert(settings.SendClientCredentialsAsAuthorizationHeader, NotNil)
		check.Assert(settings.UsePKCE, NotNil)
		check.Assert(*settings.SendClientCredentialsAsAuthorizationHeader, Equals, true)
		check.Assert(*settings.UsePKCE, Equals, true)
	} else {
		check.Assert(settings.SendClientCredentialsAsAuthorizationHeader, IsNil)
		check.Assert(settings.UsePKCE, IsNil)
	}
	if vcd.client.Client.APIVCDMaxVersionIs(">= 38.1") {
		check.Assert(settings.CustomUiButtonLabel, NotNil)
		check.Assert(*settings.CustomUiButtonLabel, Equals, "this is a test")
	} else {
		check.Assert(settings.CustomUiButtonLabel, IsNil)
	}
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
	for tries < 5 {
		tries++
		newSettings, err = adminOrg.SetOpenIdConnectSettings(settings)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "could not establish a connection") || strings.Contains(err.Error(), "connect timed out") {
			time.Sleep(10 * time.Second)
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
