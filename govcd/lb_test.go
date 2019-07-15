// +build unit lb functional integration ALL
// +build !skipLong
/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Setup vApp
// ??? Upload photon os ???
// Add 2 VMs with Photon OS and init script
// Setup load balancer
// Enable load balancing capabilities (to cover the enablement)
// Any firewall'ing at Edge Gateway?
// Perform HTTP requests and expect for both nodes to respond
// Cleanup load balancers
// Cleanup VMs
// Cleanup vApp
func (vcd *TestVCD) Test_LB(check *C) {

	// Sort out prerequisites
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVdcByName(vcd.config.VCD.Vdc)
	check.Assert(err, IsNil)

	catalog, err := org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)

	catalogItem, err := catalog.FindCatalogItem("photon-os-test")
	check.Assert(err, IsNil)
	if catalogItem.CatalogItem.Name == "" {
		// Upload photon os and prep template
		err = uploadImage(vcd, check)
		check.Assert(err, IsNil)
	}

	catalogItem, err = catalog.FindCatalogItem("photon-os-test")
	check.Assert(err, IsNil)

	vappTemplate, err := catalogItem.GetVAppTemplate()
	check.Assert(err, IsNil)

	// vApp and 2 VMs with photon-os
	// Compose vApp
	err = vdc.ComposeRawVApp(TestLB)
	check.Assert(err, IsNil)
	vapp, err := vdc.FindVAppByName(TestLB)
	check.Assert(err, IsNil)
	// Wait untill vApp becomes configuraable
	initialVappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	vapp.BlockWhileStatus(initialVappStatus, vcd.vapp.client.MaxRetryTimeout)

	// Attach VDC network to vApp so that VMs can use it
	net, err := vdc.FindVDCNetwork(vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)
	task, err := vapp.AddRAWNetworkConfig([]*types.OrgVDCNetwork{net.OrgVDCNetwork})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(TestLB, "vapp", "", "createTestVapp")

	// Spawn 2 VMs with python servers
	desiredNetConfig := types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 0
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModePool,
			Network:                 vcd.config.VCD.Network.Net1,
			NetworkConnectionIndex:  0,
		})

	vm1, err := runVM("FirstNode", vdc, vapp, desiredNetConfig, vappTemplate, check)
	vm2, err := runVM("SecondNode", vdc, vapp, desiredNetConfig, vappTemplate, check)

	ip1 := vm1.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
	ip2 := vm2.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress

	fmt.Println("vm1 net: ", ip1)
	fmt.Println("vm2 net: ", ip2)

	// time.Sleep(60 * time.Second)
	// VMs are up. IP addresses known. Let's setup load balancing

	// edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	// check.Assert(err, IsNil)

	vcd.buildLB(ip1, ip2, check)

	queryUrl := "http://" + vcd.config.VCD.ExternalIp + ":8000/server"

	fmt.Println("will query: ", queryUrl)
	// time.Sleep(15 * time.Second)

	for i := 1; i <= 60; i++ {
		resp, err := http.Get(queryUrl)
		fmt.Printf("http request error: %s", err)
		// check.Assert(err, IsNil)
		// defer resp.Body.Close()
		if err == nil {
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(resp.StatusCode)
			fmt.Println(string(body))
		}

		// check.Assert(err, IsNil)

		fmt.Println("sleeping 5 seconds")
		time.Sleep(5 * time.Second)
	}

	fmt.Println("Done")
	// panic("do not destroy setup")

	return
}

func runVM(name string, vdc Vdc, vapp VApp, net types.NetworkConnectionSection, vAppTemplate VAppTemplate, check *C) (VM, error) {
	task, err := vapp.AddNewVM(name, vAppTemplate, &net, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	vm1, err := vdc.FindVMByName(vapp, name)
	check.Assert(err, IsNil)

	task, err = vm1.RunCustomizationScript(name,
		"mkdir /tmp/node && cd /tmp/node && echo '"+name+"'>server && "+
			"/bin/systemctl stop iptables && /usr/bin/python3 -m http.server 8000 &")
	// vm1.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	// task, err = vapp.PowerOn()
	// err = task.WaitTaskCompletion()
	task, err = vm1.PowerOn()
	err = task.WaitTaskCompletion()

	fmt.Println("power on done: " + name)

	return vm1, nil

}

func uploadImage(vcd *TestVCD, check *C) error {
	itemName := "photon-os-test"
	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)
	uploadTask, err := catalog.UploadOvf(vcd.config.Media.PhotonOsOvaPath, itemName, "Photon OS image for testing", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_LB")

	catalog, err = org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
	return nil
}
