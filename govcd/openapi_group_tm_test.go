//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmOpenApiGroupLdap(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	ldapCfg := vcd.config.Tm.Ldap
	if ldapCfg.Host == "" || ldapCfg.Username == "" || ldapCfg.Password == "" || ldapCfg.Type == "" ||
		ldapCfg.Port == 0 || ldapCfg.BaseDistinguishedName == "" || ldapCfg.ImportGroup == "" {
		check.Skip("LDAP testing configuration (including 'tm.ldap.importGroup') is required")
	}

	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()

	// TODO: TM: Change to vcdClient.GetTmOrgById(orgId), requires implementing Role support for that type
	adminOrg, err := vcd.client.GetAdminOrgById(org.TmOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	roleOrgAdmin, err := adminOrg.GetRoleByName("Organization Administrator")
	check.Assert(err, IsNil)
	roleOrgUser, err := adminOrg.GetRoleByName("Organization User")
	check.Assert(err, IsNil)

	// Configure LDAP in the Org. The trusted certificate goes away with the Org
	ldapSettings := &types.OrgLdapSettingsType{
		OrgLdapMode: types.LdapModeCustom,
		CustomOrgLdapSettings: &types.CustomOrgLdapSettings{
			HostName:                ldapCfg.Host,
			Port:                    ldapCfg.Port,
			SearchBase:              ldapCfg.BaseDistinguishedName,
			AuthenticationMechanism: "SIMPLE",
			IsSsl:                   ldapCfg.IsSsl,
			ConnectorType:           ldapCfg.Type,
			Username:                ldapCfg.Username,
			Password:                ldapCfg.Password,
			UserAttributes: &types.OrgLdapUserAttributes{
				ObjectClass:               "user",
				ObjectIdentifier:          "objectGuid",
				Username:                  "sAMAccountName",
				Email:                     "mail",
				FullName:                  "displayName",
				GivenName:                 "givenName",
				Surname:                   "sn",
				Telephone:                 "telephoneNumber",
				GroupMembershipIdentifier: "dn",
				GroupBackLinkIdentifier:   "tokenGroups",
			},
			GroupAttributes: &types.OrgLdapGroupAttributes{
				ObjectClass:          "group",
				ObjectIdentifier:     "objectGuid",
				GroupName:            "cn",
				Membership:           "member",
				MembershipIdentifier: "dn",
				BackLinkIdentifier:   "objectSid",
			},
		},
	}
	_, err = org.LdapConfigure(ldapSettings, ldapCfg.IsSsl)
	check.Assert(err, IsNil)
	defer func() {
		err = org.LdapDisable()
		check.Assert(err, IsNil)
		if ldapCfg.IsSsl {
			certs, err := org.GetAllTrustedCertificates(url.Values{
				"filter": []string{fmt.Sprintf("alias==*%s*", ldapCfg.Host)},
			})
			check.Assert(err, IsNil)
			for _, cert := range certs {
				check.Assert(cert.Delete(), IsNil)
			}
		}
	}()

	tenantContext := &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	}

	// Create (import)
	groupConfig := &types.OpenApiGroup{
		Name:           ldapCfg.ImportGroup,
		ProviderType:   "LDAP",
		RoleEntityRefs: []types.OpenApiReference{{ID: roleOrgUser.Role.ID}},
		OrgEntityRef:   &types.OpenApiReference{ID: org.TmOrg.ID, Name: org.TmOrg.Name},
	}
	createdGroup, err := vcd.client.CreateGroup(groupConfig, tenantContext)
	check.Assert(err, IsNil)
	check.Assert(createdGroup, NotNil)
	check.Assert(createdGroup.Group, NotNil)
	check.Assert(createdGroup.Group.ID, Not(Equals), "")
	check.Assert(createdGroup.Group.Name, Equals, ldapCfg.ImportGroup)

	// GetAll
	allGroups, err := vcd.client.GetAllGroups(nil, tenantContext)
	check.Assert(err, IsNil)
	check.Assert(len(allGroups), Equals, 1)
	check.Assert(allGroups[0].Group.ID, Equals, createdGroup.Group.ID)

	// Get by Name
	byName, err := vcd.client.GetGroupByName(ldapCfg.ImportGroup, tenantContext)
	check.Assert(err, IsNil)
	check.Assert(byName.Group.ID, Equals, createdGroup.Group.ID)

	// Get by ID
	byId, err := vcd.client.GetGroupById(createdGroup.Group.ID, tenantContext)
	check.Assert(err, IsNil)
	check.Assert(byId.Group.ID, Equals, createdGroup.Group.ID)

	// Update role
	updateConfig := byId.Group
	updateConfig.RoleEntityRefs = []types.OpenApiReference{{ID: roleOrgAdmin.Role.ID}}
	updatedGroup, err := byId.Update(updateConfig)
	check.Assert(err, IsNil)
	check.Assert(updatedGroup, NotNil)
	check.Assert(updatedGroup.Group.ID, Equals, createdGroup.Group.ID)
	check.Assert(len(updatedGroup.Group.RoleEntityRefs), Equals, 1)
	check.Assert(updatedGroup.Group.RoleEntityRefs[0].ID, Equals, roleOrgAdmin.Role.ID)

	// Delete
	err = updatedGroup.Delete()
	check.Assert(err, IsNil)

	// Get by ID and fail
	byIdNotFound, err := vcd.client.GetGroupById(updatedGroup.Group.ID, tenantContext)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdNotFound, IsNil)
}
