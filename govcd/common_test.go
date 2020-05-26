// +build api functional catalog vapp gateway network org query extnetwork task vm vdc system disk lb lbAppRule lbAppProfile lbServerPool lbServiceMonitor lbVirtualServer user nsxv ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// createAngGetResourcesForVmCreation creates vAPP and two VM for the testing
func (vcd *TestVCD) createAngGetResourcesForVmCreation(check *C, vmName string) (*Vdc, *EdgeGateway, VAppTemplate, *VApp, types.NetworkConnectionSection, error) {
	// Get org and vdc
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
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
	fmt.Printf("# Creating RAW vApp '%s'", vmName)
	vappTemplate, err := catalogItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	// Compose Raw vApp
	err = vdc.ComposeRawVApp(vmName)
	check.Assert(err, IsNil)
	vapp, err := vdc.GetVAppByName(vmName, true)
	check.Assert(err, IsNil)
	// vApp was created - let's add it to cleanup list
	AddToCleanupList(vmName, "vapp", "", "createTestVapp")
	// Wait until vApp becomes configurable
	initialVappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	if initialVappStatus != "RESOLVED" { // RESOLVED vApp is ready to accept operations
		err = vapp.BlockWhileStatus(initialVappStatus, vapp.client.MaxRetryTimeout)
		check.Assert(err, IsNil)
	}
	fmt.Printf(". Done\n")
	fmt.Printf("# Attaching VDC network '%s' to vApp '%s'", vcd.config.VCD.Network.Net1, vmName)
	// Attach VDC network to vApp so that VMs can use it
	net, err := vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	task, err := vapp.AddRAWNetworkConfig([]*types.OrgVDCNetwork{net.OrgVDCNetwork})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")
	// Spawn 2 VMs with python servers in the newly created vApp
	desiredNetConfig := types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 0
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModePool,
			Network:                 vcd.config.VCD.Network.Net1,
			NetworkConnectionIndex:  0,
		})
	return vdc, edge, vappTemplate, vapp, desiredNetConfig, err
}

// spawnVM spawns VMs in provided vApp from template and can also apply customizationScript if
// provided
func spawnVM(name string, memorySize int, vdc Vdc, vapp VApp, net types.NetworkConnectionSection, vAppTemplate VAppTemplate, check *C, customizationScript string) (VM, error) {
	fmt.Printf("# Spawning VM '%s'", name)
	task, err := vapp.AddNewVM(name, vAppTemplate, &net, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	vm, err := vapp.GetVMByName(name, true)
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	if customizationScript != "" {
		fmt.Printf("# Applying 2 vCPU and "+strconv.Itoa(memorySize)+"MB configuration for VM '%s'", name)
		task, err = vm.ChangeCPUCount(2)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)

		task, err = vm.ChangeMemorySize(memorySize)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		fmt.Printf(". Done\n")

		fmt.Printf("# Applying customization script for VM '%s'", name)
		task, err = vm.RunCustomizationScript(name, customizationScript)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		fmt.Printf(". Done\n")
	}

	fmt.Printf("# Powering on VM '%s'", name)
	task, err = vm.PowerOn()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	return *vm, nil
}

// isItemPhotonOs checks if a catalog item is Photon OS
func isItemPhotonOs(item CatalogItem) bool {
	vappTemplate, err := item.GetVAppTemplate()
	// Unable to get template - can validate it's Photon OS
	if err != nil {
		return false
	}
	// Photon OS template has exactly 1 child
	if len(vappTemplate.VAppTemplate.Children.VM) != 1 {
		return false
	}

	// If child name is not "Photon OS" it's not Photon OS
	if vappTemplate.VAppTemplate.Children.VM[0].Name != "Photon OS" {
		return false
	}

	return true
}

// cacheLoadBalancer is meant to store load balancer settings before any operations so that all
// configuration can be checked after manipulation
func testCacheLoadBalancer(edge EdgeGateway, check *C) (*types.LbGeneralParamsWithXml, string) {
	beforeLb, err := edge.GetLBGeneralParams()
	check.Assert(err, IsNil)
	beforeLbXml := testGetEdgeEndpointXML(types.LbConfigPath, edge, check)
	return beforeLb, beforeLbXml
}

// testGetEdgeEndpointXML is used for additional validation that modifying edge gateway endpoint
// does not change any single field. It returns an XML string of whole configuration
func testGetEdgeEndpointXML(endpoint string, edge EdgeGateway, check *C) string {

	httpPath, err := edge.buildProxiedEdgeEndpointURL(endpoint)
	check.Assert(err, IsNil)

	resp, err := edge.client.ExecuteRequestWithCustomError(httpPath, http.MethodGet, types.AnyXMLMime,
		fmt.Sprintf("unable to get XML from endpoint %s: %%s", endpoint), nil, &types.NSXError{})
	check.Assert(err, IsNil)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check.Assert(err, IsNil)

	return string(body)
}

// testCheckLoadBalancerConfig validates if both raw XML string and load balancer struct remain
// identical after settings manipulation.
func testCheckLoadBalancerConfig(beforeLb *types.LbGeneralParamsWithXml, beforeLbXml string, edge EdgeGateway, check *C) {
	afterLb, err := edge.GetLBGeneralParams()
	check.Assert(err, IsNil)

	afterLbXml := testGetEdgeEndpointXML(types.LbConfigPath, edge, check)

	// remove `<version></version>` tag from both XML represntation and struct for deep comparison
	// because this version changes with each update and will never be the same after a few
	// operations

	reVersion := regexp.MustCompile(`<version>\w*<\/version>`)
	beforeLbXml = reVersion.ReplaceAllLiteralString(beforeLbXml, "")
	afterLbXml = reVersion.ReplaceAllLiteralString(afterLbXml, "")

	beforeLb.Version = ""
	afterLb.Version = ""

	check.Assert(beforeLb, DeepEquals, afterLb)
	check.Assert(beforeLbXml, DeepEquals, afterLbXml)
}

// checkIfTcpPortIsOpen checks if remote TCP port is open or closed every 5 seconds until timeout is
// reached
func checkIfTcpPortIsOpen(host, port string, timeout int) bool {
	retryTimeout := timeout
	// due to the VMs taking long time to boot it needs to be at least 5 minutes
	// may be even more in slower environments
	if timeout < 5*60 { // 5 minutes
		retryTimeout = 5 * 60 // 5 minutes
	}
	timeOutAfterInterval := time.Duration(retryTimeout) * time.Second
	timeoutAfter := time.After(timeOutAfterInterval)
	tick := time.NewTicker(time.Duration(5) * time.Second)

	for {
		select {
		case <-timeoutAfter:
			fmt.Printf("x")
			return false
		case <-tick.C:
			timeout := time.Second * 3
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
			if err != nil {
				fmt.Printf(".")
			}
			// Connection established - the port is open
			if conn != nil {
				defer conn.Close()
				fmt.Printf(" OK\n")
				return true
			}
		}
	}

}
