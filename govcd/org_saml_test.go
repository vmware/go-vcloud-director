//go:build org || functional || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	_ "embed"
	"encoding/xml"
	"fmt"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

//go:embed test-resources/saml-test-idp.xml
var externalMetadata string

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

	if testVerbose {
		fmt.Printf("# 1 %# v\n", pretty.Formatter(settings))
	}

	metadata, err := adminOrg.GetServiceProviderSamlMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	if testVerbose {
		fmt.Printf("# 2 %# v\n", pretty.Formatter(metadata))
	}

	metadataText, err := xml.Marshal(metadata)
	check.Assert(err, IsNil)
	settings.SAMLMetadata = string(metadataText)
	settings.SamlSPEntityID = "dummyId"
	settings.Enabled = true
	settings.SamlAttributeMapping.EmailAttributeName = "email"
	settings.SamlAttributeMapping.UserNameAttributeName = "uname"
	settings.SamlAttributeMapping.FirstNameAttributeName = "fname"
	settings.SamlAttributeMapping.SurnameAttributeName = "lname"
	settings.SamlAttributeMapping.FullNameAttributeName = "fullname"
	settings.SamlAttributeMapping.RoleAttributeName = "role"
	settings.SamlAttributeMapping.GroupAttributeName = "group"
	// Use a service provider metadata, without proper namespace settings: expecting an error
	newSetting, err := adminOrg.SetFederationSettings(settings)
	check.Assert(err, NotNil)
	check.Assert(err.Error(), Matches, "(?i).*bad request.*is not a valid SAML 2.0 metadata document.*")
	check.Assert(newSetting, IsNil)

	// Add namespace definitions to the metadata, and this time it will pass
	newMetadataText, err := normalizeServiceProviderSamlMetadata(string(metadataText))
	check.Assert(err, IsNil)
	settings.SAMLMetadata = newMetadataText
	newSetting, err = adminOrg.SetFederationSettings(settings)
	check.Assert(err, IsNil)
	check.Assert(newSetting, NotNil)

	check.Assert(err, IsNil)
	settings.SAMLMetadata = externalMetadata
	newSetting, err = adminOrg.SetFederationSettings(settings)
	check.Assert(err, IsNil)
	check.Assert(newSetting, NotNil)
	check.Assert(newSetting.SamlSPEntityID, Equals, "dummyId")
	check.Assert(newSetting.Enabled, Equals, true)
	check.Assert(newSetting.SamlAttributeMapping.EmailAttributeName, Equals, "email")
	check.Assert(newSetting.SamlAttributeMapping.FirstNameAttributeName, Equals, "fname")
	check.Assert(newSetting.SamlAttributeMapping.SurnameAttributeName, Equals, "lname")
	check.Assert(newSetting.SamlAttributeMapping.FullNameAttributeName, Equals, "fullname")
	check.Assert(newSetting.SamlAttributeMapping.UserNameAttributeName, Equals, "uname")
	check.Assert(newSetting.SamlAttributeMapping.RoleAttributeName, Equals, "role")
	check.Assert(newSetting.SamlAttributeMapping.GroupAttributeName, Equals, "group")
	check.Assert(newSetting, NotNil)

	err = adminOrg.UnsetFederationSettings()
	check.Assert(err, IsNil)
	err = adminOrg.Refresh()
	check.Assert(err, IsNil)
	newSettings, err := adminOrg.GetFederationSettings()
	check.Assert(err, IsNil)
	check.Assert(newSettings.SamlSPEntityID, Equals, "dummyId")
	check.Assert(newSettings.Enabled, Equals, false)

	err = adminOrg.Disable()
	check.Assert(err, IsNil)
	err = adminOrg.Delete(true, true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) TestClient_RetrieveRemoteDoc(check *C) {
	// samltest.id is a well known test site for SAML services
	metadataUrl := "https://samltest.id/saml/idp"
	metadata, err := vcd.client.Client.RetrieveRemoteDocument(metadataUrl)
	if err != nil {
		check.Skip("samltest.id is not responding")
	}
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	errors := ValidateSamlServiceProviderMetadata(string(metadata))
	check.Assert(errors, IsNil)
}

func (vcd *TestVCD) TestClient_RetrieveRemoteSamlMetadata(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip("No organization found")
	}
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	metadataText, err := adminOrg.RetrieveServiceProviderSamlMetadata()
	check.Assert(err, IsNil)
	errors := ValidateSamlServiceProviderMetadata(metadataText)
	check.Assert(errors, IsNil)
}
