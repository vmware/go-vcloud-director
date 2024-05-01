//go:build system || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"os"
	"strings"
	"time"
)

const (
	secondVcdUrl      = "VCD_URL2"
	secondVcdUser     = "VCD_USER2"
	secondVcdPassword = "VCD_PASSWORD2"
	secondVcdOrg      = "VCD_ORG2"
)

func getClientConnectionFromEnv() (*VCDClient, error) {
	vcdUrl := os.Getenv(secondVcdUrl)
	user := os.Getenv(secondVcdUser)
	password := os.Getenv(secondVcdPassword)
	orgName := os.Getenv(secondVcdOrg)
	if !strings.HasSuffix(vcdUrl, "/api") {
		return nil, fmt.Errorf("the VCD URL must terminate with '/api'")
	}
	var missing []string
	if vcdUrl == "" {
		missing = append(missing, secondVcdUrl)
	}
	if user == "" {
		missing = append(missing, secondVcdUser)
	}
	if password == "" {
		missing = append(missing, secondVcdPassword)
	}
	if password == "" {
		missing = append(missing, secondVcdOrg)
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing environment variables for connection: %v", missing)
	}

	return newUserConnection(vcdUrl, user, password, orgName, true)
}

func (vcd *TestVCD) Test_SiteAssociations(check *C) {

	if !vcd.client.Client.IsSysAdmin {
		check.Skip(fmt.Sprintf("test %s requires system administrator privileges\n", check.TestName()))
	}

	firstVcdClient := vcd.client
	secondVcdClient, err := getClientConnectionFromEnv()
	if err != nil {
		check.Skip(fmt.Sprintf("this test requires connection from a second VCD, identified by environment variables %s %s %s %s: %s",
			secondVcdUrl, secondVcdUser, secondVcdPassword, secondVcdOrg, err))
	}

	// The second VCD must be different from the first one
	check.Assert(os.Getenv(secondVcdUrl), Not(Equals), vcd.client.Client.VCDHREF.String())

	version1, _, err := firstVcdClient.Client.GetVcdVersion()
	check.Assert(err, IsNil)
	version2, _, err := secondVcdClient.Client.GetVcdVersion()
	check.Assert(err, IsNil)

	// Both VCDs must have the same version
	check.Assert(version1, Equals, version2)

	// STEP 1 Get the site association data from both VCDs
	firstVcdStructuredAssociationData, err := firstVcdClient.Client.GetSiteAssociationData()
	check.Assert(err, IsNil)
	firstVcdRawAssociationData, err := firstVcdClient.Client.GetSiteRawAssociationData()
	check.Assert(err, IsNil)
	secondVcdStructuredAssociationData, err := secondVcdClient.Client.GetSiteAssociationData()
	check.Assert(err, IsNil)
	secondVcdRawAssociationData, err := secondVcdClient.Client.GetSiteRawAssociationData()
	check.Assert(err, IsNil)

	// Check that the raw data is equivalent to the structured data
	firstConvertedAssociationData, err := RawDataToStructuredXml[types.SiteAssociationMember](firstVcdRawAssociationData)
	check.Assert(err, IsNil)
	check.Assert(firstConvertedAssociationData.SiteID, Equals, firstVcdStructuredAssociationData.SiteID)
	check.Assert(firstConvertedAssociationData.PublicKey, Equals, firstVcdStructuredAssociationData.PublicKey)
	check.Assert(firstConvertedAssociationData.RestEndpointCertificate, Equals, firstVcdStructuredAssociationData.RestEndpointCertificate)
	secondConvertedAssociationData, err := RawDataToStructuredXml[types.SiteAssociationMember](secondVcdRawAssociationData)
	check.Assert(err, IsNil)
	check.Assert(secondConvertedAssociationData.SiteID, Equals, secondVcdStructuredAssociationData.SiteID)
	check.Assert(secondConvertedAssociationData.PublicKey, Equals, secondVcdStructuredAssociationData.PublicKey)
	check.Assert(secondConvertedAssociationData.RestEndpointCertificate, Equals, secondVcdStructuredAssociationData.RestEndpointCertificate)

	// STEP 2 Get the list of current site associations from both sites for further comparison
	//orgs1before, err := firstVcdClient.QueryAllOrgs()
	//orgs1before, err := firstVcdClient.GetOrgList()
	//check.Assert(err, IsNil)
	//orgs2before, err := secondVcdClient.QueryAllOrgs()
	//orgs2before, err := secondVcdClient.GetOrgList()
	//check.Assert(err, IsNil)
	associations1before, err := firstVcdClient.Client.GetSiteAssociations()
	check.Assert(err, IsNil)
	associations2before, err := secondVcdClient.Client.GetSiteAssociations()
	check.Assert(err, IsNil)

	// STEP 3 Set the associations in both VCDs
	err = firstVcdClient.Client.SetSiteAssociation(*secondVcdStructuredAssociationData)
	check.Assert(err, IsNil)
	err = secondVcdClient.Client.SetSiteAssociation(*firstVcdStructuredAssociationData)
	check.Assert(err, IsNil)
	// Note: there is no call to AddToCleanupList, because we can't defer that action to a temporary client in a separate VCD

	// STEP 4 Check that the association is complete
	status1, elapsed1, err := firstVcdClient.Client.CheckSiteAssociation(secondVcdStructuredAssociationData.SiteID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("site #1: status: %s - elapsed: %s\n", status1, elapsed1)
	status2, elapsed2, err := secondVcdClient.Client.CheckSiteAssociation(firstVcdStructuredAssociationData.SiteID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("site #2: status: %s - elapsed: %s\n", status2, elapsed2)

	// STEP 5 get the list of associations
	associations1, err := firstVcdClient.Client.GetSiteAssociations()
	check.Assert(err, IsNil)
	check.Assert(len(associations1), Equals, len(associations1before)+1)
	associations2, err := secondVcdClient.Client.GetSiteAssociations()
	check.Assert(err, IsNil)
	check.Assert(len(associations2), Equals, len(associations2before)+1)

	//orgs1after, err := firstVcdClient.QueryAllOrgs()
	//orgs1after, err := firstVcdClient.GetOrgList()
	//check.Assert(err, IsNil)
	//orgs2after, err := secondVcdClient.QueryAllOrgs()
	//orgs2after, err := secondVcdClient.GetOrgList()
	//check.Assert(err, IsNil)
	//check.Assert(len(orgs1after.Org), Equals, len(orgs1before.Org)+len(orgs2before.Org))
	//check.Assert(len(orgs2after.Org), Equals, len(orgs1before.Org)+len(orgs2before.Org))

	// STEP 6 retrieve the specific associations that we have just created (used for removal)
	association1, err := firstVcdClient.Client.GetSiteAssociationBySiteId(secondVcdStructuredAssociationData.SiteID)
	check.Assert(err, IsNil)
	association2, err := secondVcdClient.Client.GetSiteAssociationBySiteId(firstVcdStructuredAssociationData.SiteID)
	check.Assert(err, IsNil)

	// STEP 7 remove site associations
	defer func() {
		err = firstVcdClient.Client.RemoveSiteAssociation(association1.Href)
		check.Assert(err, IsNil)
		err = secondVcdClient.Client.RemoveSiteAssociation(association2.Href)
		check.Assert(err, IsNil)
	}()

	// STEP 8 get organization association data from both sides

	var localOrg *AdminOrg
	// The local org –if possible– is accessed through Org admin
	if len(vcd.config.Tenants) > 0 {
		localUser, err := newUserConnection(firstVcdClient.Client.VCDHREF.String(),
			vcd.config.Tenants[0].User,
			vcd.config.Tenants[0].Password,
			vcd.config.Tenants[0].SysOrg, true)
		check.Assert(err, IsNil)
		localOrg, err = localUser.GetAdminOrgByName(vcd.config.Tenants[0].SysOrg)
		fmt.Println("using Org user for local org")
	} else {
		localOrg, err = firstVcdClient.GetAdminOrgByName(vcd.config.VCD.Org)
	}
	check.Assert(err, IsNil)
	remoteOrgs, err := secondVcdClient.GetOrgList()
	check.Assert(err, IsNil)
	check.Assert(len(remoteOrgs.Org) > 1, Equals, true)

	remoteOrgName := remoteOrgs.Org[0].Name
	if strings.EqualFold(remoteOrgName, "system") {
		remoteOrgName = remoteOrgs.Org[1].Name
	}
	remoteOrg, err := secondVcdClient.GetAdminOrgByName(remoteOrgName)
	check.Assert(err, IsNil)

	orgAssociationData1, err := localOrg.GetOrgAssociationData()
	check.Assert(err, IsNil)
	orgAssociationData2, err := remoteOrg.GetOrgAssociationData()
	check.Assert(err, IsNil)

	// STEP 9 set org association between two VCDs

	err = localOrg.SetOrgAssociation(*orgAssociationData2)
	check.Assert(err, IsNil)
	err = remoteOrg.SetOrgAssociation(*orgAssociationData1)
	check.Assert(err, IsNil)

	// STEP 10 check association connection

	status1, elapsed1, err = localOrg.CheckOrgAssociation(orgAssociationData2.OrgID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("org #1: status: %s - elapsed: %s\n", status1, elapsed1)
	status2, elapsed2, err = remoteOrg.CheckOrgAssociation(orgAssociationData1.OrgID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("org #2: status: %s - elapsed: %s\n", status2, elapsed2)

	// STEP 11 retrieve the specific associations that we have just created (used for removal)
	orgAssociation1, err := localOrg.GetOrgAssociationByOrgId(orgAssociationData2.OrgID)
	check.Assert(err, IsNil)
	orgAssociation2, err := remoteOrg.GetOrgAssociationByOrgId(orgAssociationData1.OrgID)
	check.Assert(err, IsNil)

	defer func() {
		err = localOrg.RemoveOrgAssociation(orgAssociation1.Href)
		check.Assert(err, IsNil)
		err = remoteOrg.RemoveOrgAssociation(orgAssociation2.Href)
		check.Assert(err, IsNil)
	}()

	/*
		siteStruct, err := firstVcdClient.Client.GetSite()
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

	*/
}
