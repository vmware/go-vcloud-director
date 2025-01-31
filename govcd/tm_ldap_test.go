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
	"regexp"
	"strings"
)

// Test_TmLdapSystemWithNoSslLdap tests LDAP configuration in Provider (System) org by using
// vCenter as LDAP (no SSL)
func (vcd *TestVCD) Test_TmLdapSystemWithNoSslLdap(check *C) {
	skipNonTm(vcd, check)
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	vcd.checkSkipWhenApiToken(check)

	ldapSettings := types.TmLdapSettings{
		AuthenticationMechanism: "SIMPLE",
		ConnectorType:           "ACTIVE_DIRECTORY",
		CustomUiButtonLabel:     addrOf("Hello there"),
		GroupAttributes: &types.LdapGroupAttributesType{
			BackLinkIdentifier:   "objectSid",
			GroupName:            "cn",
			Membership:           "member",
			MembershipIdentifier: "dn",
			ObjectClass:          "group",
			ObjectIdentifier:     "objectGuid",
		},
		HostName:            regexp.MustCompile(`https?://`).ReplaceAllString(vcd.config.Tm.VcenterUrl, ""),
		IsSsl:               false,
		MaxResults:          200,
		MaxUserGroups:       1015,
		PageSize:            200,
		PagedSearchDisabled: false,
		Password:            vcd.config.Tm.VcenterPassword,
		Port:                389,
		SearchBase:          "dc=vsphere,dc=local",
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
		UserName: "cn=Administrator,cn=Users,dc=vsphere,dc=local",
	}

	receivedSettings, err := vcd.client.TmLdapConfigure(&ldapSettings, false)
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, NotNil)

	receivedSettings2, err := vcd.client.TmGetLdapConfiguration()
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, DeepEquals, receivedSettings2)

	defer func() {
		fmt.Println("Unconfiguring LDAP")
		// Clear LDAP configuration
		err = vcd.client.TmLdapDisable()
		check.Assert(err, IsNil)
	}()
}

// Test_TmLdapSystemWithNoSslLdap tests LDAP configuration in Provider (System) org by using
// a configured LDAP that uses SSL
func (vcd *TestVCD) Test_TmLdapSystemWithSslLdap(check *C) {
	skipNonTm(vcd, check)
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	vcd.checkSkipWhenApiToken(check)

	if vcd.config.Tm.Ldap.Host == "" || vcd.config.Tm.Ldap.Username == "" || vcd.config.Tm.Ldap.Password == "" || vcd.config.Tm.Ldap.Type == "" ||
		vcd.config.Tm.Ldap.Port == 0 || vcd.config.Tm.Ldap.BaseDistinguishedName == "" {
		check.Skip("LDAP testing configuration is required")
	}

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
		IsSsl:               true,
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

	_, err := vcd.client.TmLdapConfigure(&ldapSettings, false)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "cannot configure LDAP"), Equals, true)

	receivedSettings, err := vcd.client.TmLdapConfigure(&ldapSettings, true)
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, NotNil)

	receivedSettings2, err := vcd.client.TmGetLdapConfiguration()
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, DeepEquals, receivedSettings2)

	defer func() {
		fmt.Println("Unconfiguring LDAP")
		// Clear LDAP configuration
		err = vcd.client.TmLdapDisable()
		check.Assert(err, IsNil)

		// Clean up trusted certificate
		certs, err := vcd.client.GetAllTrustedCertificates(url.Values{
			"filter": []string{fmt.Sprintf("alias==*%s*", vcd.config.Tm.Ldap.Host)},
		})
		check.Assert(err, IsNil)
		check.Assert(len(certs), Equals, 1)
		err = certs[0].Delete()
		check.Assert(err, IsNil)
	}()
}

// Test_TmLdapOrgWithVCenterLdap tests LDAP configuration in a regular Organization by using
// vCenter as LDAP
func (vcd *TestVCD) Test_TmLdapOrgWithVCenterLdap(check *C) {
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

	ldapSettings := &types.OrgLdapSettingsType{
		OrgLdapMode: types.LdapModeCustom,
		CustomOrgLdapSettings: &types.CustomOrgLdapSettings{
			HostName:                regexp.MustCompile(`https?://`).ReplaceAllString(vcd.config.Tm.VcenterUrl, ""),
			Port:                    389,
			SearchBase:              "dc=vsphere,dc=local",
			AuthenticationMechanism: "SIMPLE",
			ConnectorType:           "OPEN_LDAP",
			Username:                "cn=Administrator,cn=Users,dc=vsphere,dc=local",
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

	receivedSettings, err := org.LdapConfigure(ldapSettings, false)
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, NotNil)

	receivedSettings2, err := org.GetLdapConfiguration()
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, DeepEquals, receivedSettings2)

	ldapSettings.OrgLdapMode = types.LdapModeSystem
	ldapSettings.CustomOrgLdapSettings = nil
	receivedSettings, err = org.LdapConfigure(ldapSettings, false)
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, NotNil)

	receivedSettings2, err = org.GetLdapConfiguration()
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, DeepEquals, receivedSettings2)

	// This is same to deletion
	ldapSettings.OrgLdapMode = types.LdapModeNone
	receivedSettings, err = org.LdapConfigure(ldapSettings, false)
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, NotNil)

	receivedSettings2, err = org.GetLdapConfiguration()
	check.Assert(err, IsNil)
	check.Assert(receivedSettings, DeepEquals, receivedSettings2)

	err = org.LdapDisable()
	check.Assert(err, IsNil)
}
