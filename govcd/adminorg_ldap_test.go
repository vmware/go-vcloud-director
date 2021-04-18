// +build user functional ALL
// +build !skipLong

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LDAP serves as a "subtest" framework for tests requiring LDAP configuration. It sets up LDAP
// server and configuration for Org and cleans up this test run.
//
// Prerequisites:
// * External network subnet must have access to internet
// * Correct DNS servers must be set for external network so that guest VM can resolve DNS records
func (vcd *TestVCD) Test_LDAP(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	ctx := context.Background()

	if !catalogItemIsPhotonOs(ctx, vcd) {
		check.Skip(fmt.Sprintf("Catalog item '%s' is not Photon OS", vcd.config.VCD.Catalog.CatalogItem))
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("[" + check.TestName() + "] external network not provided")
	}

	fmt.Println("Setting up LDAP")
	networkName, vappName, vmName := vcd.configureLdap(ctx, check)
	defer func() {
		fmt.Println("Unconfiguring LDAP")
		vcd.unconfigureLdap(ctx, check, networkName, vappName, vmName)
	}()

	// Run tests requiring LDAP from here.
	vcd.test_GroupCRUD(check)
	vcd.test_GroupFinderGetGenericEntity(check)

}

// configureLdap creates direct network, spawns Photon OS VM with LDAP server and configures vCD to
// use LDAP server
func (vcd *TestVCD) configureLdap(ctx context.Context, check *C) (string, string, string) {

	// Create direct network to expose LDAP server on external network
	directNetworkName := createDirectNetwork(ctx, vcd, check)

	// Launch LDAP server on external network
	ldapHostIp, vappName, vmName := createLdapServer(ctx, vcd, check, directNetworkName)

	// Configure vCD to use new LDAP server
	configureLdapForOrg(ctx, vcd, check, ldapHostIp)

	return directNetworkName, vappName, vmName
}

// unconfigureLdap cleans up LDAP configuration created by `configureLdap` immediately to reduce
// resource usage
func (vcd *TestVCD) unconfigureLdap(ctx context.Context, check *C, networkName, vAppName, vmName string) {

	// Get Org Vdc
	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVDCByName(ctx, vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	vapp, err := vdc.GetVAppByName(ctx, vAppName, false)
	check.Assert(err, IsNil)

	vm, err := vapp.GetVMByName(ctx, vmName, false)
	check.Assert(err, IsNil)

	// Remove VM
	task, err := vm.Undeploy(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	err = vapp.RemoveVM(ctx, *vm)
	check.Assert(err, IsNil)

	// undeploy and remove vApp
	task, err = vapp.Undeploy(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	task, err = vapp.Delete(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	// Remove network
	err = RemoveOrgVdcNetworkIfExists(ctx, *vcd.vdc, networkName)
	check.Assert(err, IsNil)

	// Clear LDAP configuration
	err = org.LdapDisable(ctx)
	check.Assert(err, IsNil)

}

// orgConfigureLdap sets up LDAP configuration in vCD org specified by vcd.config.VCD.Org variable
func configureLdapForOrg(ctx context.Context, vcd *TestVCD, check *C, ldapHostIp string) {
	fmt.Printf("# Configuring LDAP settings for Org '%s'", vcd.config.VCD.Org)

	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
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

	_, err = org.LdapConfigure(ctx, ldapSettings)
	check.Assert(err, IsNil)

	fmt.Println(" Done")
	AddToCleanupList("LDAP-configuration", "orgLdapSettings", org.AdminOrg.Name, check.TestName())
}

// createLdapServer spawns a vApp and photon OS VM. Using customization script it starts a testing
// LDAP server in docker container which has a few users and groups defined.
// In essence it creates two groups - "admin_staff" and "ship_crew" and a few users.
// More information about users and groups in: https://github.com/rroemhild/docker-test-openldap
func createLdapServer(ctx context.Context, vcd *TestVCD, check *C, directNetworkName string) (string, string, string) {
	vAppName := "ldap"
	// The customization script waits until IP address is set on the NIC because Guest tools run
	// script and network configuration together. If the script runs too quick - there is a risk
	// that network card is not yet configured and it will not be able to pull docker image from
	// remote. Guest tools could also be interrupted if the script below failed before NICs are
	// configured therefore it is run in background.
	// It waits until "inet" (not "inet6") is set and then runs docker container
	const ldapCustomizationScript = `
		{
			until ip a show eth0 | grep "inet "
			do
				sleep 1
			done
			systemctl enable docker
			systemctl start docker
			docker run --name ldap-server --restart=always --privileged -d -p 389:389 rroemhild/test-openldap
		} &
	`
	// Get Org, Vdc
	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVDCByName(ctx, vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Find catalog and catalog item
	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	catalogItem, err := catalog.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)

	fmt.Printf("# Creating RAW vApp '%s'", vAppName)
	vappTemplate, err := catalogItem.GetVAppTemplate(ctx)
	check.Assert(err, IsNil)
	// Compose Raw vApp
	err = vdc.ComposeRawVApp(ctx, vAppName)
	check.Assert(err, IsNil)
	vapp, err := vdc.GetVAppByName(ctx, vAppName, true)
	check.Assert(err, IsNil)
	// vApp was created - adding it to cleanup list (using prepend to remove it before direct
	// network removal)
	PrependToCleanupList(vAppName, "vapp", "", check.TestName())
	// Wait until vApp becomes configurable
	initialVappStatus, err := vapp.GetStatus(ctx)
	check.Assert(err, IsNil)
	if initialVappStatus != "RESOLVED" { // RESOLVED vApp is ready to accept operations
		err = vapp.BlockWhileStatus(ctx, initialVappStatus, vapp.client.MaxRetryTimeout)
		check.Assert(err, IsNil)
	}
	fmt.Printf(". Done\n")

	// Attach VDC network to vApp so that VMs can use it
	fmt.Printf("# Attaching network '%s'", directNetworkName)
	net, err := vdc.GetOrgVdcNetworkByName(ctx, directNetworkName, false)
	check.Assert(err, IsNil)
	task, err := vapp.AddRAWNetworkConfig(ctx, []*types.OrgVDCNetwork{net.OrgVDCNetwork})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
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
	ldapVm, err := spawnVM(ctx, "ldap-vm", 1024, *vdc, *vapp, desiredNetConfig, vappTemplate, check, ldapCustomizationScript, true)
	check.Assert(err, IsNil)

	// Must be deleted before vApp to avoid IP leak
	PrependToCleanupList(ldapVm.VM.Name, "vm", vAppName, check.TestName())

	// Got VM - ensure that TCP port for ldap service is open and reachable
	ldapHostIp := ldapVm.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
	fmt.Printf("# Waiting for server %s to respond on port 389: ", ldapHostIp)
	timerStart := time.Now()
	isLdapServiceUp := isTcpPortOpen(ldapHostIp, "389", vapp.client.MaxRetryTimeout)
	check.Assert(isLdapServiceUp, Equals, true)
	fmt.Printf("# Time taken to start LDAP container: %s\n", time.Since(timerStart))

	return ldapHostIp, vAppName, ldapVm.VM.Name
}

// createDirectNetwork creates a direct network attached to existing external network
func createDirectNetwork(ctx context.Context, vcd *TestVCD, check *C) string {
	networkName := check.TestName()
	fmt.Printf("# Creating direct network %s.", networkName)

	err := RemoveOrgVdcNetworkIfExists(ctx, *vcd.vdc, networkName)
	if err != nil {
		check.Skip(fmt.Sprintf("Error deleting network : %s", err))
	}

	externalNetwork, err := vcd.client.GetExternalNetworkByName(ctx, vcd.config.VCD.ExternalNetwork)
	check.Assert(err, IsNil)
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

	task, err := vcd.vdc.CreateOrgVDCNetwork(ctx, &networkConfig)
	if err != nil {
		fmt.Printf("error creating the network: %s", err)
	}
	check.Assert(err, IsNil)
	if task == (Task{}) {
		fmt.Printf("NULL task retrieved after network creation")
	}
	check.Assert(task.Task.HREF, Not(Equals), "")

	AddToCleanupList(networkName, "network", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, check.TestName())

	err = task.WaitInspectTaskCompletion(ctx, LogTask, 10)
	if err != nil {
		fmt.Printf("error performing task: %s", err)
	}
	check.Assert(err, IsNil)
	fmt.Println(" Done")
	return networkName
}
