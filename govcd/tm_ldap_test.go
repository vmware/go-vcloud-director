//go:build tm || functional || ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
	"net/url"
	"strings"
)

// Test_TmLdapSystem tests LDAP configuration in Provider (System) org. This test
// checks LDAP connection with SSL and without SSL
func (vcd *TestVCD) Test_TmLdapSystem(check *C) {
	skipNonTm(vcd, check)
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	vcd.checkSkipWhenApiToken(check)

	if vcd.config.Tm.Ldap.Host == "" || vcd.config.Tm.Ldap.Username == "" || vcd.config.Tm.Ldap.Password == "" || vcd.config.Tm.Ldap.Type == "" ||
		vcd.config.Tm.Ldap.Port == 0 || vcd.config.Tm.Ldap.BaseDistinguishedName == "" {
		check.Skip("LDAP testing configuration is required")
	}

	// Definition of the different use cases for LDAP configuration
	type testCase struct {
		isSsl bool
	}
	testCases := []testCase{
		{isSsl: false},
		{isSsl: true},
	}

	for _, t := range testCases {
		fmt.Printf("%s - %#v\n", check.TestName(), t)

		ldapSettings := types.TmLdapSettings{
			AuthenticationMechanism: "SIMPLE",
			ConnectorType:           vcd.config.Tm.Ldap.Type,
			CustomUiButtonLabel:     addrOf("Hello there"),
			GroupAttributes: &types.LdapGroupAttributesType{
				BackLinkIdentifier:   "objectSid",
				GroupName:            "cn",
				Membership:           "member",
				MembershipIdentifier: "dn",
				ObjectClass:          "group",
				ObjectIdentifier:     "objectGuid",
			},
			HostName:            vcd.config.Tm.Ldap.Host,
			IsSsl:               t.isSsl,
			MaxResults:          200,
			MaxUserGroups:       1015,
			PageSize:            200,
			PagedSearchDisabled: false,
			Password:            vcd.config.Tm.Ldap.Password,
			Port:                vcd.config.Tm.Ldap.Port,
			SearchBase:          vcd.config.Tm.Ldap.BaseDistinguishedName,
			UserAttributes: &types.LdapUserAttributesType{
				Email:                     "mail",
				FullName:                  "displayName",
				GivenName:                 "givenName",
				GroupBackLinkIdentifier:   "tokenGroups",
				GroupMembershipIdentifier: "dn",
				ObjectClass:               "user",
				ObjectIdentifier:          "objectGuid",
				Surname:                   "sn",
				Telephone:                 "telephoneNumber",
				UserName:                  "sAMAccountName",
			},
			UserName: vcd.config.Tm.Ldap.Username,
		}

		if t.isSsl {
			_, err := vcd.client.TmLdapConfigure(&ldapSettings, false)
			check.Assert(err, NotNil)
			check.Assert(strings.Contains(err.Error(), "cannot configure LDAP"), Equals, true)
		}

		receivedSettings, err := vcd.client.TmLdapConfigure(&ldapSettings, t.isSsl)
		check.Assert(err, IsNil)
		check.Assert(receivedSettings, NotNil)

		receivedSettings2, err := vcd.client.TmGetLdapConfiguration()
		check.Assert(err, IsNil)
		check.Assert(receivedSettings2, NotNil)
		check.Assert(receivedSettings2, DeepEquals, receivedSettings)

		// Update LDAP configuration. It should not trust any new certificate unless the host is changed
		ldapSettings.MaxUserGroups = ldapSettings.MaxUserGroups + 1
		receivedSettings2, err = vcd.client.TmLdapConfigure(&ldapSettings, t.isSsl)
		check.Assert(err, IsNil)
		check.Assert(receivedSettings2, NotNil)
		check.Assert(receivedSettings2.MaxUserGroups, Equals, receivedSettings.MaxUserGroups+1)

		// Clear LDAP configuration
		err = vcd.client.TmLdapDisable()
		check.Assert(err, IsNil)

		if t.isSsl {
			// Clean up trusted certificate
			certs, err := vcd.client.GetAllTrustedCertificates(url.Values{
				"filter": []string{fmt.Sprintf("alias==*%s*", vcd.config.Tm.Ldap.Host)},
			}, nil)
			check.Assert(err, IsNil)
			check.Assert(len(certs), Equals, 1) // Important to check that only one certificate was added
			err = certs[0].Delete()
			check.Assert(err, IsNil)
		}
	}
}

// Test_TmLdapOrg tests LDAP configuration in a regular Organization
func (vcd *TestVCD) Test_TmLdapOrg(check *C) {
	skipNonTm(vcd, check)
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	vcd.checkSkipWhenApiToken(check)

	// We are testing LDAP for a regular Organization
	cfg := &types.TmOrg{
		Name:        check.TestName(),
		DisplayName: check.TestName(),
		IsEnabled:   true,
	}
	org, err := vcd.client.CreateTmOrg(cfg)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	defer func() {
		err = org.Disable()
		check.Assert(err, IsNil)
		err = org.Delete()
		check.Assert(err, IsNil)
	}()

	// Add to cleanup list
	PrependToCleanupListOpenApi(org.TmOrg.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgs+org.TmOrg.ID)

	// Definition of the different use cases for LDAP configuration
	type testCase struct {
		isSsl bool
	}
	testCases := []testCase{
		{isSsl: false},
		{isSsl: true},
	}

	for _, t := range testCases {
		fmt.Printf("%s - %#v\n", check.TestName(), t)
		ldapSettings := &types.OrgLdapSettingsType{
			OrgLdapMode: types.LdapModeCustom,
			CustomOrgLdapSettings: &types.CustomOrgLdapSettings{
				HostName:                vcd.config.Tm.Ldap.Host,
				Port:                    vcd.config.Tm.Ldap.Port,
				SearchBase:              vcd.config.Tm.Ldap.BaseDistinguishedName,
				AuthenticationMechanism: "SIMPLE",
				IsSsl:                   t.isSsl,
				ConnectorType:           vcd.config.Tm.Ldap.Type,
				Username:                vcd.config.Tm.Ldap.Username,
				Password:                vcd.config.Tm.VcenterPassword,
				UserAttributes: &types.OrgLdapUserAttributes{
					ObjectClass:               "inetOrgPerson",
					ObjectIdentifier:          "uid",
					Username:                  "uid",
					Email:                     "mail",
					FullName:                  "cn",
					GivenName:                 "givenName",
					Surname:                   "sn",
					Telephone:                 "telephoneNumber",
					GroupMembershipIdentifier: "dn",
				},
				GroupAttributes: &types.OrgLdapGroupAttributes{
					ObjectClass:          "group",
					ObjectIdentifier:     "cn",
					GroupName:            "cn",
					Membership:           "member",
					MembershipIdentifier: "dn",
				},
			},
		}

		if t.isSsl {
			_, err := org.LdapConfigure(ldapSettings, false)
			check.Assert(err, NotNil)
			check.Assert(strings.Contains(err.Error(), "cannot configure LDAP"), Equals, true)
		}

		receivedSettings, err := org.LdapConfigure(ldapSettings, t.isSsl)
		check.Assert(err, IsNil)
		check.Assert(receivedSettings, NotNil)

		receivedSettings2, err := org.GetLdapConfiguration()
		check.Assert(err, IsNil)
		check.Assert(receivedSettings2, NotNil)
		check.Assert(receivedSettings2, DeepEquals, receivedSettings)

		// Update LDAP configuration. It should not trust any new certificate unless the host is changed
		ldapSettings.CustomOrgLdapSettings.SearchBase = vcd.config.Tm.Ldap.BaseDistinguishedName + ",DC=foo"
		receivedSettings2, err = org.LdapConfigure(ldapSettings, t.isSsl)
		check.Assert(err, IsNil)
		check.Assert(receivedSettings2, NotNil)
		check.Assert(receivedSettings2.CustomOrgLdapSettings.SearchBase, Equals, receivedSettings.CustomOrgLdapSettings.SearchBase+",DC=foo")

		ldapSettings.OrgLdapMode = types.LdapModeSystem
		ldapSettings.CustomOrgLdapSettings = nil
		receivedSettings, err = org.LdapConfigure(ldapSettings, t.isSsl)
		check.Assert(err, IsNil)
		check.Assert(receivedSettings, NotNil)

		receivedSettings2, err = org.GetLdapConfiguration()
		check.Assert(err, IsNil)
		check.Assert(receivedSettings, DeepEquals, receivedSettings2)

		// This is same to deletion
		ldapSettings.OrgLdapMode = types.LdapModeNone
		receivedSettings, err = org.LdapConfigure(ldapSettings, t.isSsl)
		check.Assert(err, IsNil)
		check.Assert(receivedSettings, NotNil)

		receivedSettings2, err = org.GetLdapConfiguration()
		check.Assert(err, IsNil)
		check.Assert(receivedSettings, DeepEquals, receivedSettings2)

		if t.isSsl {
			// Clean up trusted certificate. This step is not needed as it would be gone with the Org, but it's an extra check
			certs, err := org.GetAllTrustedCertificates(url.Values{
				"filter": []string{fmt.Sprintf("alias==*%s*", vcd.config.Tm.Ldap.Host)},
			})
			check.Assert(err, IsNil)
			check.Assert(len(certs), Equals, 1) // Important to check that only one certificate was added
			err = certs[0].Delete()
			check.Assert(err, IsNil)
		}
	}
}
