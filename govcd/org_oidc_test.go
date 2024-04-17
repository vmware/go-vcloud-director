//go:build org || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	_ "embed"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
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

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := adminOrg.GetOpenIdConnectSettings()
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)
	check.Assert(settings.OrgRedirectUri, Not(Equals), "")

	settings, err = adminOrg.SetOpenIdConnectSettings(types.OrgOAuthSettings{
		ClientId:          addrOf("a"),
		ClientSecret:      addrOf("b"),
		Enabled:           addrOf(true),
		WellKnownEndpoint: addrOf("http://10.196.34.27:8080/stf-oidc-server/.well-known/openid-configuration"),
	})
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)

	err = adminOrg.DeleteOpenIdConnectSettings()
	check.Assert(err, IsNil)

	//err = adminOrg.Delete(true, true)
	//check.Assert(err, IsNil)
}
