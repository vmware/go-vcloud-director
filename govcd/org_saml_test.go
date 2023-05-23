//go:build org || functional || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_OrgSamlSettingsCRUD(check *C) {

	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}
	orgName := check.TestName()

	task, err := CreateOrg(vcd.client, orgName, orgName, orgName, &types.OrgSettings{}, true)
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)
	AddToCleanupList(orgName, "org", "", check.TestName())
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	adminOrg, err := vcd.client.GetAdminOrgByName(orgName)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	settings, err := adminOrg.GetFederationSettings()
	check.Assert(err, IsNil)
	check.Assert(settings, NotNil)

	fmt.Printf("# 1 %# v\n", pretty.Formatter(settings))

	metadata, err := adminOrg.GetSamlMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	fmt.Printf("# 2 %# v\n", pretty.Formatter(metadata))

	/**/
	metadataText, err := xml.Marshal(metadata)
	check.Assert(err, IsNil)
	settings.SAMLMetadata = string(metadataText)
	settings.Enabled = true
	newSetting, err := adminOrg.SetFederationSettings(settings)
	check.Assert(err, IsNil)
	check.Assert(newSetting, NotNil)
	/**/
	err = adminOrg.Disable()
	check.Assert(err, IsNil)
	err = adminOrg.Delete(true, true)
	check.Assert(err, IsNil)
}
