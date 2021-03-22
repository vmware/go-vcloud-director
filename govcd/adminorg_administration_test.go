// +build org functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LDAP_Configuration tests LDAP configuration functions
func (vcd *TestVCD) Test_LDAP_Configuration(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	ctx := context.Background()

	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	ldapSettings := &types.OrgLdapSettingsType{
		OrgLdapMode: types.LdapModeCustom,
		CustomOrgLdapSettings: &types.CustomOrgLdapSettings{
			HostName:                "1.1.1.1",
			Port:                    389,
			SearchBase:              "dc=planetexpress,dc=com",
			AuthenticationMechanism: "SIMPLE",
			ConnectorType:           "OPEN_LDAP",
			Username:                "cn=admin,dc=planetexpress,dc=com",
			Password:                "GoodNewsEveryone",
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
	gotSettings, err := org.LdapConfigure(ctx, ldapSettings)
	check.Assert(err, IsNil)

	AddToCleanupList("LDAP-configuration", "orgLdapSettings", org.AdminOrg.Name, check.TestName())

	check.Assert(ldapSettings.CustomOrgLdapSettings.GroupAttributes, DeepEquals, gotSettings.CustomOrgLdapSettings.GroupAttributes)
	check.Assert(ldapSettings.CustomOrgLdapSettings.UserAttributes, DeepEquals, gotSettings.CustomOrgLdapSettings.UserAttributes)
	check.Assert(ldapSettings.CustomOrgLdapSettings.UserAttributes, DeepEquals, gotSettings.CustomOrgLdapSettings.UserAttributes)
	check.Assert(ldapSettings.CustomOrgLdapSettings.Username, DeepEquals, gotSettings.CustomOrgLdapSettings.Username)
	check.Assert(ldapSettings.CustomOrgLdapSettings.AuthenticationMechanism, DeepEquals, gotSettings.CustomOrgLdapSettings.AuthenticationMechanism)
	check.Assert(ldapSettings.CustomOrgLdapSettings.ConnectorType, DeepEquals, gotSettings.CustomOrgLdapSettings.ConnectorType)

	err = org.LdapDisable(ctx)
	check.Assert(err, IsNil)

	ldapConfig, err := org.GetLdapConfiguration(ctx)
	check.Assert(err, IsNil)

	check.Assert(ldapConfig.OrgLdapMode, Equals, types.LdapModeNone)

}
