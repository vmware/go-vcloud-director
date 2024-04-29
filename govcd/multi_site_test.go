//go:build system || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"time"
)

func (vcd *TestVCD) Test_GetSiteAssociations(check *C) {

	if !vcd.client.Client.IsSysAdmin {
		check.Skip(fmt.Sprintf("test %s requires system administrator privileges\n", check.TestName()))
	}

	siteStruct, err := vcd.client.Client.GetSite()
	check.Assert(err, IsNil)
	fmt.Printf("CURRENT SITE %# v\n", pretty.Formatter(siteStruct))

	siteQueryAssociations, err := vcd.client.Client.QueryAllSiteAssociations(nil, nil)
	check.Assert(err, IsNil)
	for i, s := range siteQueryAssociations {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(s))
	}
	fmt.Println()
	orgQueryAssociations, err := vcd.client.Client.QueryAllOrgAssociations(nil, nil)
	check.Assert(err, IsNil)
	for i, s := range orgQueryAssociations {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(s))
	}

	fmt.Println()
	associationData, err := vcd.client.Client.GetSiteAssociationData()
	check.Assert(err, IsNil)
	fmt.Printf("---- %# v\n", pretty.Formatter(associationData))

	rawAssociationData, err := vcd.client.Client.GetSiteRawAssociationData()
	check.Assert(err, IsNil)
	fmt.Printf("%s\n", rawAssociationData)

	fmt.Println()
	siteAssociations, err := vcd.client.Client.GetSiteAssociations()
	check.Assert(err, IsNil)
	for i, a := range siteAssociations {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(a))
	}

	fmt.Println()
	//org, err := vcd.client.GetAdminOrgByName("gmaxia")
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	orgAssociations, err := org.GetOrgAssociations()
	check.Assert(err, IsNil)
	for i, s := range orgAssociations {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(s))
	}

	orgAssociationData, err := org.GetOrgAssociationData()
	check.Assert(err, IsNil)
	fmt.Printf("---- %# v\n", pretty.Formatter(orgAssociationData))
	orgRawAssociationData, err := org.GetOrgRawAssociationData()
	check.Assert(err, IsNil)
	fmt.Printf("---- %s\n", orgRawAssociationData)

	// TODO: change the test to be more generic

	siteFileName := "./multi-site/sc1-vcd-22-29.eng.vmware.com.xml"
	siteSettingData, err := ReadXmlDataFromFile[types.SiteAssociationMember](siteFileName)
	check.Assert(err, IsNil)
	check.Assert(siteSettingData, NotNil)
	err = vcd.client.Client.SetSiteAssociation(*siteSettingData)
	check.Assert(err, IsNil)
	time.Sleep(10 * time.Second)
	newSiteAssociations, err := vcd.client.Client.GetSiteAssociations()
	check.Assert(err, IsNil)
	for i, s := range newSiteAssociations {
		fmt.Printf("NEW SITE %d %# v\n", i, pretty.Formatter(s))
	}

	// This check should be performed only when a full site connection has been established (i.e. both sides have done the connection)
	status, elapsed, err := vcd.client.Client.CheckSiteAssociation(siteSettingData.SiteID, 120*time.Second)
	check.Assert(err, IsNil)
	check.Assert(status, Equals, "ACTIVE")
	fmt.Printf("elapsed: %s\n", elapsed)

	siteAssociationToDelete, err := vcd.client.Client.GetSiteAssociationBySiteId(siteSettingData.SiteID)
	check.Assert(err, IsNil)
	err = vcd.client.Client.RemoveSiteAssociation(siteAssociationToDelete.Href)
	check.Assert(err, IsNil)

	orgFileName := "./multi-site/datacloud-1.xml"
	orgSettingData, err := ReadXmlDataFromFile[types.OrgAssociationMember](orgFileName)
	check.Assert(err, IsNil)
	check.Assert(orgSettingData, NotNil)

	err = org.SetOrgAssociation(*orgSettingData)
	check.Assert(err, IsNil)
	time.Sleep(10 * time.Second)
	newOrgAssociations, err := org.GetOrgAssociations()
	check.Assert(err, IsNil)
	for i, s := range newOrgAssociations {
		fmt.Printf("NEW %d %# v\n", i, pretty.Formatter(s))
	}

	// This check should be performed only when a full org connection has been established (i.e. both sides have done the connection)
	status, elapsed, err = org.CheckOrgAssociation(orgSettingData.OrgID, 120*time.Second)
	check.Assert(err, IsNil)
	check.Assert(status, Equals, "ACTIVE")
	fmt.Printf("elapsed: %s\n", elapsed)

	orgAssociationToDelete, err := org.GetOrgAssociationByOrgId(orgSettingData.OrgID)
	check.Assert(err, IsNil)
	err = org.RemoveOrgAssociation(orgAssociationToDelete.Href)
	check.Assert(err, IsNil)
}
