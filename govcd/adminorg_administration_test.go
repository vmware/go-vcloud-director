// +build user functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// configureLdap creates direct network, spawns Photon OS VM with LDAP server and configures vCD to
// use LDAP server
func (vcd *TestVCD) configureLdap(check *C) (string, string, string) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	// Create direct network to expose LDAP server on external network
	directNetworkName := createDirectNetwork(vcd, check)

	// Launch LDAP server on external network
	ldapHostIp, vappName, vmName := createLdapServer(vcd, check, directNetworkName)

	// Configure vCD to use new LDAP server
	orgConfigureLdap(vcd, check, ldapHostIp)

	return directNetworkName, vappName, vmName
}

// unconfigureLdap releases resources as soon as possible to avoid using external IPs or VMs during
// other test runs
func (vcd *TestVCD) unconfigureLdap(check *C, networkName, vAppName, vmName string) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	// Get Org Vdc
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	vapp, err := vdc.GetVAppByName(vAppName, false)
	check.Assert(err, IsNil)

	vm, err := vapp.GetVMByName(vmName, false)
	check.Assert(err, IsNil)

	// Remove VM
	task, err := vm.Undeploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	err = vapp.RemoveVM(*vm)
	check.Assert(err, IsNil)

	// Wait for vApp to complete task before removing itself
	// err = vapp.BlockWhileStatus("UNRESOLVED", vapp.client.MaxRetryTimeout)
	// check.Assert(err, IsNil)

	// undeploy and remove vApp
	task, err = vapp.Undeploy()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// Remove network
	err = RemoveOrgVdcNetworkIfExists(*vcd.vdc, networkName)
	check.Assert(err, IsNil)

	// Clear LDAP configuration
	err = org.LdapDisable()
	check.Assert(err, IsNil)

}

// vcdConfigureLdap sets up LDAP configuration in vCD org specified by vcd.config.VCD.Org variable
func orgConfigureLdap(vcd *TestVCD, check *C, ldapHostIp string) {
	fmt.Printf("# Configuring LDAP settings for Org '%s'", vcd.config.VCD.Org)

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
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
				ObjectClass:               "user",
				ObjectIdentifier:          "objectGuid",
				Username:                  "sAMAccountName",
				Email:                     "mail",
				FullName:                  "cn",
				GivenName:                 "givenName",
				Surname:                   "sn",
				Telephone:                 "telephone",
				GroupMembershipIdentifier: "dn",
				GroupBackLinkIdentifier:   "tokenGroups",
			},
			GroupAttributes: &types.OrgLdapGroupAttributes{
				ObjectClass:          "group",
				ObjectIdentifier:     "objectGuid",
				GroupName:            "cn",
				Membership:           "member",
				MembershipIdentifier: "dn",
			},
		},
	}

	err = org.LdapConfigure(ldapSettings)
	check.Assert(err, IsNil)

	fmt.Println(" Done")
	AddToCleanupList("LDAP-configuration", "orgLdapSettings", org.AdminOrg.Name, check.TestName())
}

// createLdapServer spawns a vApp and photon OS VM. Using customization script it starts a testing
// LDAP server in docker container which has a few users and groups defined.
// In essence it creates two groups - "admin_staff" and "ship_crew" and
// More information: https://github.com/rroemhild/docker-test-openldap
func createLdapServer(vcd *TestVCD, check *C, directNetworkName string) (string, string, string) {
	vAppName := "ldap"
	const ldapCustomizationScript = "systemctl enable docker ; systemctl start docker ;" +
		"docker run --name ldap-server --restart=always --privileged -d -p 389:389 rroemhild/test-openldap"

	// Get Org, Vdc
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Find catalog and catalog item
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	catalogItem, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	// Skip the test if catalog item is not Photon OS
	if !isItemPhotonOs(*catalogItem) {
		check.Skip(fmt.Sprintf("Skipping test because catalog item %s is not Photon OS",
			vcd.config.VCD.Catalog.CatalogItem))
	}
	fmt.Printf("# Creating RAW vApp '%s'", vAppName)
	vappTemplate, err := catalogItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	// Compose Raw vApp
	err = vdc.ComposeRawVApp(vAppName)
	check.Assert(err, IsNil)
	vapp, err := vdc.GetVAppByName(vAppName, true)
	check.Assert(err, IsNil)
	// vApp was created - adding it to cleanup list (using prepend to remove it before direct
	// network removal)
	PrependToCleanupList(vAppName, "vapp", "", check.TestName())
	// Wait until vApp becomes configurable
	initialVappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	if initialVappStatus != "RESOLVED" { // RESOLVED vApp is ready to accept operations
		err = vapp.BlockWhileStatus(initialVappStatus, vapp.client.MaxRetryTimeout)
		check.Assert(err, IsNil)
	}
	fmt.Printf(". Done\n")

	// Attach VDC network to vApp so that VMs can use it
	fmt.Printf("# Attaching network '%s'", directNetworkName)
	net, err := vdc.GetOrgVdcNetworkByName(directNetworkName, false)
	check.Assert(err, IsNil)
	task, err := vapp.AddRAWNetworkConfig([]*types.OrgVDCNetwork{net.OrgVDCNetwork})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	// Create VM
	desiredNetConfig := types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 0
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModePool,
			Network:                 directNetworkName,
			NetworkConnectionIndex:  0,
		})

	// LDAP docker container does not start if Photon OS VM does not have at least 1024 of RAM
	ldapVm, err := spawnVM("ldap-vm", 1024, *vdc, *vapp, desiredNetConfig, vappTemplate, check, ldapCustomizationScript)
	check.Assert(err, IsNil)

	// Must be deleted before vApp to avoid IP leak
	PrependToCleanupList(ldapVm.VM.Name, "vm", vAppName, check.TestName())

	// Got VM - ensure that TCP port for ldap service is open and reachable
	ldapHostIp := ldapVm.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
	fmt.Printf("# Waiting for server %s to respond on port 389: ", ldapHostIp)
	isLdapServiceUp := checkIfTcpPortIsOpen(ldapHostIp, "389", vapp.client.MaxRetryTimeout)
	check.Assert(isLdapServiceUp, Equals, true)

	return ldapHostIp, vAppName, ldapVm.VM.Name
}

// createDirectNetwork creates a direct network attached to existing external network
func createDirectNetwork(vcd *TestVCD, check *C) string {
	networkName := check.TestName()
	fmt.Printf("# Creating direct network %s.", networkName)

	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	err := RemoveOrgVdcNetworkIfExists(*vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("[" + check.TestName() + "] external network not provided")
	}
	externalNetwork, err := vcd.client.GetExternalNetworkByName(vcd.config.VCD.ExternalNetwork)
	if err != nil {
		check.Skip("[" + check.TestName() + "] parent network not found")
		return ""
	}
	// Note that there is no IPScope for this type of network
	description := "Created by govcd test"
	var networkConfig = types.OrgVDCNetwork{
		Xmlns:       types.XMLNamespaceVCloud,
		Name:        networkName,
		Description: description,
		Configuration: &types.NetworkConfiguration{
			FenceMode: types.FenceModeBridged,
			ParentNetwork: &types.Reference{
				HREF: externalNetwork.ExternalNetwork.HREF,
				Name: externalNetwork.ExternalNetwork.Name,
				Type: externalNetwork.ExternalNetwork.Type,
			},
			BackwardCompatibilityMode: true,
		},
		IsShared: false,
	}
	LogNetwork(networkConfig)

	task, err := vcd.vdc.CreateOrgVDCNetwork(&networkConfig)
	if err != nil {
		fmt.Printf("error creating the network: %s", err)
	}
	check.Assert(err, IsNil)
	if task == (Task{}) {
		fmt.Printf("NULL task retrieved after network creation")
	}
	check.Assert(task.Task.HREF, Not(Equals), "")

	AddToCleanupList(networkName,
		"network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, check.TestName())

	err = task.WaitInspectTaskCompletion(LogTask, 10)
	if err != nil {
		fmt.Printf("error performing task: %s", err)
	}
	check.Assert(err, IsNil)
	fmt.Println(" Done")
	return networkName
}
