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

// #nosec G101 -- These credentials are fake for testing purposes
const (
	secondVcdUrl              = "VCD_URL2"
	secondVcdUser             = "VCD_USER2"
	secondVcdPassword         = "VCD_PASSWORD2"
	secondVcdSysOrg           = "VCD_SYSORG2"
	secondVcdOrg2             = "VCD_ORG2"
	secondVcdOrgUser2         = "VCD_ORGUSER2"
	secondVcdOrgUserPassword2 = "VCD_ORGUSER_PASSWORD2"
)

func getClientConnectionFromEnv() (*VCDClient, error) {
	vcdUrl := os.Getenv(secondVcdUrl)
	user := os.Getenv(secondVcdUser)
	password := os.Getenv(secondVcdPassword)
	orgName := os.Getenv(secondVcdSysOrg)
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
	if orgName == "" {
		missing = append(missing, secondVcdSysOrg)
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing environment variables for connection: %v", missing)
	}

	return newUserConnection(vcdUrl, user, password, orgName, true)
}

/*
	Test_SiteAssociations will test the associations between two sites
	To run this test, make a shell script like the one below, filling the variables
	in addition to the VCD defined in govcd_test_config.yaml

$ cat connection.sh
export VCD_URL2=https://some-vcd-url.com/api
export VCD_USER2=administrator
export VCD_PASSWORD2='myPassword'
export VCD_SYSORG2=System
export VCD_ORG2=orgname2
export VCD_ORGUSER2=org-admin-name
export VCD_ORGUSER_PASSWORD2='myOrgAdminPassword'

$	source connection.sh
$ go test -tags functional -check.f Test_SiteAssociations -vcd-skip-vapp-creation -check.vv -timeout 0
*/
func (vcd *TestVCD) Test_SiteAssociations(check *C) {

	if !vcd.client.Client.IsSysAdmin {
		check.Skip(fmt.Sprintf("test %s requires system administrator privileges\n", check.TestName()))
	}
	if !vcd.client.Client.APIClientVersionIs(">= 37.0") {
		check.Skip(fmt.Sprintf("Minimum API version required for this test is 37.0. Found: %s", vcd.client.Client.APIVersion))
	}

	firstVcdClient := vcd.client
	secondVcdClient, err := getClientConnectionFromEnv()
	if err != nil {
		check.Skip(fmt.Sprintf("this test requires connection from a second VCD, identified by environment variables %s %s %s %s: %s",
			secondVcdUrl, secondVcdUser, secondVcdPassword, secondVcdSysOrg, err))
	}

	// The second VCD must be different from the first one
	check.Assert(os.Getenv(secondVcdUrl), Not(Equals), firstVcdClient.Client.VCDHREF.String())

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
	orgs1before, err := firstVcdClient.GetAllOrgs(nil, false)
	check.Assert(err, IsNil)
	orgs2before, err := secondVcdClient.GetAllOrgs(nil, false)
	check.Assert(err, IsNil)

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

	// STEP 4 get the list of associations and organizations
	associations1, err := firstVcdClient.Client.GetSiteAssociations()
	check.Assert(err, IsNil)
	check.Assert(len(associations1), Equals, len(associations1before)+1)
	associations2, err := secondVcdClient.Client.GetSiteAssociations()
	check.Assert(err, IsNil)
	check.Assert(len(associations2), Equals, len(associations2before)+1)

	// STEP 5 retrieve the specific associations that we have just created (used for removal)
	association1, err := firstVcdClient.Client.GetSiteAssociationBySiteId(secondVcdStructuredAssociationData.SiteID)
	check.Assert(err, IsNil)
	association2, err := secondVcdClient.Client.GetSiteAssociationBySiteId(firstVcdStructuredAssociationData.SiteID)
	check.Assert(err, IsNil)

	// STEP 6 trigger the site association removal (at the end of tests)
	defer func() {
		fmt.Println("removing site association 1")
		err = firstVcdClient.Client.RemoveSiteAssociation(association1.Href)
		check.Assert(err, IsNil)
		fmt.Println("removing site association 2")
		err = secondVcdClient.Client.RemoveSiteAssociation(association2.Href)
		check.Assert(err, IsNil)
	}()

	// STEP 7 Check that the association is complete
	status1, elapsed1, err := firstVcdClient.Client.CheckSiteAssociation(secondVcdStructuredAssociationData.SiteID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("site #1: status: %s - elapsed: %s\n", status1, elapsed1)
	status2, elapsed2, err := secondVcdClient.Client.CheckSiteAssociation(firstVcdStructuredAssociationData.SiteID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("site #2: status: %s - elapsed: %s\n", status2, elapsed2)

	// STEP 8 check number of organizations
	orgs1after, err := firstVcdClient.GetAllOrgs(nil, true)
	check.Assert(err, IsNil)
	orgs2after, err := secondVcdClient.GetAllOrgs(nil, true)
	check.Assert(err, IsNil)
	fmt.Printf("site #1 - Number of orgs before associations: %d - after association: %d\n", len(orgs1before), len(orgs1after))
	fmt.Printf("site #2 - Number of orgs before associations: %d - after association: %d\n", len(orgs2before), len(orgs2after))
	check.Assert(len(orgs1after), Equals, len(orgs1before)+len(orgs2before))
	check.Assert(len(orgs2after), Equals, len(orgs1before)+len(orgs2before))

	// STEP 9 get organization association data from both sides
	// NOTE: org association from different sites can only happen AFTER the two VCDs have been associated at site level

	if len(vcd.config.Tenants) == 0 {
		fmt.Println("no tenant user defined for this VCD")
		return
	}
	localUser, err := newUserConnection(
		firstVcdClient.Client.VCDHREF.String(),
		vcd.config.Tenants[0].User,
		vcd.config.Tenants[0].Password,
		vcd.config.Tenants[0].SysOrg, true)
	check.Assert(err, IsNil)
	localOrg, err := localUser.GetAdminOrgByName(vcd.config.Tenants[0].SysOrg)
	fmt.Printf("Using Org user '%s@%s' (site 1 %s)\n", vcd.config.Tenants[0].User, vcd.config.Tenants[0].SysOrg, firstVcdClient.Client.VCDHREF.String())
	check.Assert(err, IsNil)

	fmt.Println("org1 (site1) connected")
	remoteOrgName := os.Getenv(secondVcdOrg2)
	remoteOrgUserName := os.Getenv(secondVcdOrgUser2)
	remoteOrgPassword := os.Getenv(secondVcdOrgUserPassword2)
	if remoteOrgName == "" || remoteOrgPassword == "" || remoteOrgUserName == "" {
		fmt.Printf("one or more of [%s %s %s] was not defined\n", secondVcdOrg2, secondVcdOrgUser2, secondVcdOrgUserPassword2)
		return
	}
	_, err = secondVcdClient.GetOrgByName(remoteOrgName)
	if err != nil {
		fmt.Printf("Error retrieving Org '%s' in site %s\n", remoteOrgName, secondVcdClient.Client.VCDHREF.String())
	}
	check.Assert(err, IsNil)
	fmt.Printf("Using Org user '%s@%s' (site 2 %s)\n", remoteOrgUserName, remoteOrgName, secondVcdClient.Client.VCDHREF.String())
	remoteUser, err := newUserConnection(
		secondVcdClient.Client.VCDHREF.String(),
		remoteOrgUserName,
		remoteOrgPassword,
		remoteOrgName, true)
	check.Assert(err, IsNil)
	remoteOrg, err := remoteUser.GetAdminOrgByName(remoteOrgName)
	check.Assert(err, IsNil)
	fmt.Println("org2 (site2) connected")

	orgAssociationData1, err := localOrg.GetOrgAssociationData()
	check.Assert(err, IsNil)
	orgAssociationData2, err := remoteOrg.GetOrgAssociationData()
	check.Assert(err, IsNil)

	// STEP 10: set org association between two VCDs
	err = localOrg.SetOrgAssociation(*orgAssociationData2)
	check.Assert(err, IsNil)
	err = remoteOrg.SetOrgAssociation(*orgAssociationData1)
	check.Assert(err, IsNil)

	// STEP 12: retrieve the specific associations that we have just created (used for removal)
	orgAssociation1, err := localOrg.GetOrgAssociationByOrgId(orgAssociationData2.OrgID)
	check.Assert(err, IsNil)
	orgAssociation2, err := remoteOrg.GetOrgAssociationByOrgId(orgAssociationData1.OrgID)
	check.Assert(err, IsNil)

	// STEP 11: trigger association removal (at the end of the test: it will happen before the removal of site association)
	defer func() {
		fmt.Println("removing org association 1")
		err = localOrg.RemoveOrgAssociation(orgAssociation1.Href)
		check.Assert(err, IsNil)
		fmt.Println("removing org association 2")
		err = remoteOrg.RemoveOrgAssociation(orgAssociation2.Href)
		check.Assert(err, IsNil)
	}()

	// STEP 12: check org association connection
	status1, elapsed1, err = localOrg.CheckOrgAssociation(orgAssociationData2.OrgID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("org #1 '%s' (from site 1): status: %s - elapsed: %s\n", localOrg.AdminOrg.Name, status1, elapsed1)
	status2, elapsed2, err = remoteOrg.CheckOrgAssociation(orgAssociationData1.OrgID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("org #2 '%s' (from site 2): status: %s - elapsed: %s\n", remoteOrg.AdminOrg.Name, status2, elapsed2)

	// STEP 13: deferred org and site removal will happen here
}

func (vcd *TestVCD) Test_OrgAssociations(check *C) {

	// Note: this test runs regardless of `VCD_TEST_ORG_USER` state, as it uses explicit Org user connections
	// to perform its operations

	if len(vcd.config.Tenants) < 2 {
		check.Skip(fmt.Sprintf("not enough tenant structures defined in configuration. Two are requited. %d were found", len(vcd.config.Tenants)))
	}
	// Make sure that the tenants structure is populated
	for _, tenant := range vcd.config.Tenants {
		if tenant.User == "" || tenant.SysOrg == "" || tenant.Password == "" {
			check.Skip("One or more components in tenant structure are empty.")
			return
		}
	}

	firstOrgName := vcd.config.Tenants[0].SysOrg
	secondOrgName := vcd.config.Tenants[1].SysOrg

	// STEP 0: define two Org user connections
	firstVcdClient, err := newUserConnection(vcd.client.Client.VCDHREF.String(),
		vcd.config.Tenants[0].User,
		vcd.config.Tenants[0].Password,
		firstOrgName, true)
	check.Assert(err, IsNil)
	secondVcdClient, err := newUserConnection(vcd.client.Client.VCDHREF.String(),
		vcd.config.Tenants[1].User,
		vcd.config.Tenants[1].Password,
		secondOrgName, true)
	check.Assert(err, IsNil)
	fmt.Printf("Using user '%s@%s'\n", vcd.config.Tenants[0].User, firstOrgName)
	fmt.Printf("Using user '%s@%s'\n", vcd.config.Tenants[1].User, secondOrgName)

	// STEP 1: get organization association data from both sides, using their own Org users
	var firstOrg *AdminOrg
	var secondOrg *AdminOrg
	firstOrg, err = firstVcdClient.GetAdminOrgByName(firstOrgName)
	check.Assert(err, IsNil)
	secondOrg, err = secondVcdClient.GetAdminOrgByName(secondOrgName)
	check.Assert(err, IsNil)

	orgAssociationData1, err := firstOrg.GetOrgAssociationData()
	check.Assert(err, IsNil)
	rawOrgAssociationData1, err := firstOrg.GetOrgRawAssociationData()
	check.Assert(err, IsNil)
	orgAssociationData2, err := secondOrg.GetOrgAssociationData()
	check.Assert(err, IsNil)

	// Check that the raw data is the same as the structured data
	rawOrgAssociationData2, err := secondOrg.GetOrgRawAssociationData()
	check.Assert(err, IsNil)
	convertedAssociationData1, err := RawDataToStructuredXml[types.OrgAssociationMember](rawOrgAssociationData1)
	check.Assert(err, IsNil)
	check.Assert(orgAssociationData1.OrgID, Equals, convertedAssociationData1.OrgID)
	check.Assert(orgAssociationData1.OrgPublicKey, Equals, convertedAssociationData1.OrgPublicKey)
	convertedAssociationData2, err := RawDataToStructuredXml[types.OrgAssociationMember](rawOrgAssociationData2)
	check.Assert(err, IsNil)
	check.Assert(orgAssociationData2.OrgID, Equals, convertedAssociationData2.OrgID)
	check.Assert(orgAssociationData2.OrgPublicKey, Equals, convertedAssociationData2.OrgPublicKey)

	// Check number of networks for future comparison
	networks1Before, err := firstOrg.GetAllOpenApiOrgVdcNetworks(nil, false)
	check.Assert(err, IsNil)
	networks2Before, err := secondOrg.GetAllOpenApiOrgVdcNetworks(nil, false)
	check.Assert(err, IsNil)

	// STEP 2: set org associations within the same VCD
	err = firstOrg.SetOrgAssociation(*orgAssociationData2)
	check.Assert(err, IsNil)
	err = secondOrg.SetOrgAssociation(*orgAssociationData1)
	check.Assert(err, IsNil)

	// STEP 3: check association connection
	status1, elapsed1, err := firstOrg.CheckOrgAssociation(orgAssociationData2.OrgID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("org #1 (same site): status: %s - elapsed: %s\n", status1, elapsed1)
	status2, elapsed2, err := secondOrg.CheckOrgAssociation(orgAssociationData1.OrgID, 120*time.Second)
	check.Assert(err, IsNil)
	fmt.Printf("org #2 (same site): status: %s - elapsed: %s\n", status2, elapsed2)

	// STEP 4 retrieve the specific associations that we have just created (used for removal)
	orgAssociation1, err := firstOrg.GetOrgAssociationByOrgId(orgAssociationData2.OrgID)
	check.Assert(err, IsNil)
	orgAssociation2, err := secondOrg.GetOrgAssociationByOrgId(orgAssociationData1.OrgID)
	check.Assert(err, IsNil)

	// STEP 5: trigger association removal
	defer func() {
		check.Assert(err, IsNil)
		err = firstOrg.RemoveOrgAssociation(orgAssociation1.Href)
		check.Assert(err, IsNil)
		err = secondOrg.RemoveOrgAssociation(orgAssociation2.Href)
		check.Assert(err, IsNil)
	}()

	// STEP 6: check number of networks after association
	networks1After, err := firstOrg.GetAllOpenApiOrgVdcNetworks(nil, true)
	check.Assert(err, IsNil)
	networks2After, err := secondOrg.GetAllOpenApiOrgVdcNetworks(nil, true)
	check.Assert(err, IsNil)

	fmt.Printf("org #1 - Networks before associations: %d - after association: %d\n", len(networks1Before), len(networks1After))
	fmt.Printf("org #2 - Networks before associations: %d - after association: %d\n", len(networks2Before), len(networks2After))
	check.Assert(len(networks1After), Equals, len(networks1Before)+len(networks2Before))
	check.Assert(len(networks2After), Equals, len(networks1Before)+len(networks2Before))

	// STEP 7: deferred associations removal happens here
}
