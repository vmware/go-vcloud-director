//go:build user || functional || ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LDAP serves as a "subtest" framework for tests requiring LDAP configuration. It sets up LDAP
// configuration for Org and cleans up this test run.
//
// Prerequisites:
// * LDAP server already installed
// * LDAP server IP set in TestConfig.VCD.LdapServer
func (vcd *TestVCD) Test_LDAP(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	vcd.checkSkipWhenApiToken(check)

	ldapHostIp := vcd.config.VCD.LdapServer
	if ldapHostIp == "" {
		check.Skip("[" + check.TestName() + "] LDAP server IP not provided in configuration")
	}
	// Due to a bug in VCD, when configuring LDAP service, Org publishing catalog settings `Publish external catalogs` and
	// `Subscribe to external catalogs ` gets disabled. For that reason we are getting the current values from those vars
	// to set them at the end of the test, to avoid interference with other tests.
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	publishExternalCatalogs := adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally
	subscribeToExternalCatalogs := adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanSubscribe

	fmt.Printf("Setting up LDAP (IP: %s)\n", ldapHostIp)
	err = configureLdapForOrg(vcd, adminOrg, ldapHostIp, check.TestName())
	check.Assert(err, IsNil)
	defer func() {
		fmt.Println("Unconfiguring LDAP")
		// Clear LDAP configuration
		err = adminOrg.LdapDisable()
		check.Assert(err, IsNil)

		// Due to the VCD bug mentioned above, we need to set the previous state from the publishing settings vars
		check.Assert(adminOrg.Refresh(), IsNil)

		adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally = publishExternalCatalogs
		adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanSubscribe = subscribeToExternalCatalogs

		task, err := adminOrg.Update()
		check.Assert(err, IsNil)

		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}()

	// Run tests requiring LDAP from here.
	vcd.test_GroupCRUD(check)
	vcd.test_GroupFinderGetGenericEntity(check)
	vcd.test_GroupUserListIsPopulated(check)
}

func (vcd *TestVCD) Test_LDAPSystem(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	vcd.checkSkipWhenApiToken(check)

	// Due to a bug in VCD, when configuring LDAP service, Org publishing catalog settings `Publish external catalogs` and
	// `Subscribe to external catalogs ` gets disabled. For that reason we are getting the current values from those vars
	// to set them at the end of the test, to avoid interference with other tests.
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	publishExternalCatalogs := adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally
	subscribeToExternalCatalogs := adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanSubscribe
	ldapSettings := types.OrgLdapSettingsType{
		OrgLdapMode:   "SYSTEM",
		CustomUsersOu: "ou=Foo,dc=domain,dc=local base DN",
	}

	_, err = adminOrg.LdapConfigure(&ldapSettings)
	check.Assert(err, IsNil)
	defer func() {
		fmt.Println("Unconfiguring LDAP")
		// Clear LDAP configuration
		err = adminOrg.LdapDisable()
		check.Assert(err, IsNil)

		// Due to the VCD bug mentioned above, we need to set the previous state from the publishing settings vars
		check.Assert(adminOrg.Refresh(), IsNil)

		adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally = publishExternalCatalogs
		adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanSubscribe = subscribeToExternalCatalogs

		task, err := adminOrg.Update()
		check.Assert(err, IsNil)

		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}()
}

// configureLdapForOrg sets up LDAP configuration in vCD org
func configureLdapForOrg(vcd *TestVCD, adminOrg *AdminOrg, ldapHostIp, testName string) error {
	fmt.Printf("# Configuring LDAP settings for Org '%s'", vcd.config.VCD.Org)

	// The below settings are tailored for LDAP docker testing image
	// https://github.com/rroemhild/docker-test-openldap
	ldapSettings := &types.OrgLdapSettingsType{
		OrgLdapMode: types.LdapModeCustom,
		CustomOrgLdapSettings: &types.CustomOrgLdapSettings{
			HostName:                ldapHostIp,
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

	_, err := adminOrg.LdapConfigure(ldapSettings)
	if err != nil {
		return err
	}
	fmt.Println(" Done")
	AddToCleanupList("LDAP-configuration", "orgLdapSettings", adminOrg.AdminOrg.Name, testName)
	return nil
}
