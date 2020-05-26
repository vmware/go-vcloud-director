/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_Ldap(check *C) {

	spinUpLdapServer(vcd, check, "direct-network")

}

func spinUpLdapServer(vcd *TestVCD, check *C, directNetworkName string) {
	vAppName := "ldap"
	const ldapCustomizationScript = "systemctl enable docker ; systemctl start docker ;" +
		"docker run --name ldap-server --restart=always --privileged -d -p 389:389 rroemhild/test-openldap"

	// Get Org Vdc
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
	// vApp was created - let's add it to cleanup list
	AddToCleanupList(vAppName, "vapp", "", check.TestName())
	// Wait until vApp becomes configurable
	initialVappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	if initialVappStatus != "RESOLVED" { // RESOLVED vApp is ready to accept operations
		err = vapp.BlockWhileStatus(initialVappStatus, vapp.client.MaxRetryTimeout)
		check.Assert(err, IsNil)
	}
	fmt.Printf(". Done\n")

	// Attach VDC network to vApp so that VMs can use it
	net, err := vdc.GetOrgVdcNetworkByName(directNetworkName, false)
	check.Assert(err, IsNil)
	task, err := vapp.AddRAWNetworkConfig([]*types.OrgVDCNetwork{net.OrgVDCNetwork})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")
	// EOF attach

	// Spawn VMs
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
	isLdapServiceUp := checkIfTcpPortIsOpen(ldapHostIp, "389", vapp.client.MaxRetryTimeout)
	check.Assert(isLdapServiceUp, Equals, true)
}
