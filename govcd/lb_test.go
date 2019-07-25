// +build lb functional integration ALL
// +build !skipLong

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_Lb load balancer integration test
// 1. Validates that all needed parameters are here
// 2. Uploads or reuses media.photonOsOvaPath OVA image
// 3. Creates RAW vApp and attaches vDC network to it
// 4. Spawns two VMs with configuration script to server HTTP traffic
// 5. Sets up load balancer
// 6. Probes load balancer virtual server's external IP (edge gateway IP) for traffic
// being server in 2 VMs
// 7. Tears down
func (vcd *TestVCD) Test_Lb(check *C) {

	// Validate prerequisites
	validateTestLbPrerequisites(vcd, check)

	// Get org and vdc
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVdcByName(vcd.config.VCD.Vdc)
	check.Assert(err, IsNil)
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)

	// Find catalog and catalog item
	catalog, err := org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)
	catalogItem, err := catalog.FindCatalogItem(vcd.config.VCD.Catalog.CatalogItem)
	check.Assert(err, IsNil)

	// Skip the test if catalog item is not Photon OS
	if !isItemPhotonOs(catalogItem) {
		check.Skip(fmt.Sprintf("Skipping test because catalog item %s is not Photon OS",
			vcd.config.VCD.Catalog.CatalogItem))
	}

	fmt.Printf("# Creating RAW vApp '%s'", TestLb)
	vappTemplate, err := catalogItem.GetVAppTemplate()
	check.Assert(err, IsNil)

	// Compose Raw vApp
	err = vdc.ComposeRawVApp(TestLb)
	check.Assert(err, IsNil)
	vapp, err := vdc.FindVAppByName(TestLb)
	check.Assert(err, IsNil)
	// vApp was created - let's add it to cleanup list
	AddToCleanupList(TestLb, "vapp", "", "createTestVapp")

	// Wait until vApp becomes configurable
	initialVappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	vapp.BlockWhileStatus(initialVappStatus, vapp.client.MaxRetryTimeout)
	fmt.Printf(". Done\n")

	fmt.Printf("# Attaching vDC network '%s' to vApp '%s'", vcd.config.VCD.Network.Net1, TestLb)
	// Attach vDC network to vApp so that VMs can use it
	net, err := vdc.FindVDCNetwork(vcd.config.VCD.Network.Net1)
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

	vm1, err := spawnVM("FirstNode", vdc, vapp, desiredNetConfig, vappTemplate, check)
	vm2, err := spawnVM("SecondNode", vdc, vapp, desiredNetConfig, vappTemplate, check)

	// Get IPs alocated to the VMs
	ip1 := vm1.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
	ip2 := vm2.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress

	fmt.Printf("# VM '%s' got IP '%s' in vDC network %s\n", vm1.VM.Name, ip1, vcd.config.VCD.Network.Net1)
	fmt.Printf("# VM '%s' got IP '%s' in vDC network %s\n", vm2.VM.Name, ip2, vcd.config.VCD.Network.Net1)

	fmt.Printf("# Setting up load balancer for VMs: '%s' (%s), '%s' (%s)\n", vm1.VM.Name, ip1, vm2.VM.Name, ip2)

	fmt.Printf("# Creating firewall rule for load balancer virtual server access. ")
	ruleDescription := addFirewallRule(vdc, vcd, check)
	fmt.Printf("Done\n")

	// Build load balancer
	buildLb(edge, ip1, ip2, vcd, check)

	// Cache current load balancer settings for change validation in the end
	beforeLb, beforeLbXml := testCacheLoadBalancer(edge, check)

	// Enable load balancer globally
	fmt.Printf("# Enabling load balancer with acceleration: ")
	_, err = edge.UpdateLBGeneralParams(true, true, true, "warning")
	check.Assert(err, IsNil)
	fmt.Printf("Done\n")

	// Using external edge gateway IP for
	queryUrl := "http://" + vcd.config.VCD.ExternalIp + ":8000/server"
	fmt.Printf("# Querying load balancer for expected responses at %s\n", queryUrl)
	queryErr := checkLb(queryUrl, []string{vm1.VM.Name, vm2.VM.Name}, vapp.client.MaxRetryTimeout)

	// Remove firewall rule
	fmt.Printf("# Deleting firewall rule used for load balancer virtual server access. ")
	deleteFirewallRule(ruleDescription, vdc, vcd, check)
	fmt.Printf("Done\n")

	// Restore global load balancer configuration
	fmt.Printf("# Restoring load balancer global configuration: ")
	_, err = edge.UpdateLBGeneralParams(beforeLb.Enabled, beforeLb.AccelerationEnabled,
		beforeLb.Logging.Enable, beforeLb.Logging.LogLevel)
	check.Assert(err, IsNil)
	fmt.Printf("Done\n")

	// Validate load balancer configuration against initially cached version
	fmt.Printf("# Validating load balancer XML structure: ")
	testCheckLoadBalancerConfig(beforeLb, beforeLbXml, edge, check)
	fmt.Printf("Done\n")

	// Finally after some cleanups - check if querying succeeded
	check.Assert(queryErr, IsNil)
	return
}

// validateTestLbPrerequisites verifies the following:
// * Edge Gateway is set in config
// * ExternalIp is set in config (will be edge gateway external IP)
// * PhotonOsOvaPath is set (will be used for spawning VMs)
// * Edge Gateway can be found and it has advanced networking enabled (a must for load balancers)
func validateTestLbPrerequisites(vcd *TestVCD, check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}

	if vcd.config.VCD.ExternalIp == "" {
		check.Skip("Skipping test because no edge gateway external IP given")
	}

	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

}

// spawnVM spawns VMs in provided vApp from template and also applies customization script to
// spawn a Python 3 HTTP server
func spawnVM(name string, vdc Vdc, vapp VApp, net types.NetworkConnectionSection, vAppTemplate VAppTemplate, check *C) (VM, error) {
	fmt.Printf("# Spawning VM '%s'", name)
	task, err := vapp.AddNewVM(name, vAppTemplate, &net, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	vm, err := vdc.FindVMByName(vapp, name)
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	fmt.Printf("# Applying 2 vCPU and 512MB configuration for VM '%s'", name)
	task, err = vm.ChangeCPUCount(2)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()

	task, err = vm.ChangeMemorySize(512)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	fmt.Printf(". Done\n")

	fmt.Printf("# Applying customization script for VM '%s'", name)
	// The script below creates a file /tmp/node/server with single value `name` being set in it.
	// It also disables iptables and spawns simple Python 3 HTTP server listening on port 8000
	// in background which serves the just created `server` file.
	task, err = vm.RunCustomizationScript(name,
		"mkdir /tmp/node && cd /tmp/node && echo -n '"+name+"' > server && "+
			"/bin/systemctl stop iptables && /usr/bin/python3 -m http.server 8000 &")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	fmt.Printf("# Powering on VM '%s'", name)
	task, err = vm.PowerOn()
	err = task.WaitTaskCompletion()
	fmt.Printf(". Done\n")

	return vm, nil
}

// buildLB establishes an HTTP load balancer for 2 IPs specified as arguments
func buildLb(edge EdgeGateway, node1Ip, node2Ip string, vcd *TestVCD, check *C) {

	_, serverPoolId, appProfileId, _ := buildTestLBVirtualServerPrereqs(node1Ip, node2Ip, TestLb,
		check, vcd, edge)

	// Configure creation object including reference to service monitor
	lbVirtualServerConfig := &types.LbVirtualServer{
		Name: TestLb,
		// Load balancer virtual server serves on Edge gw IP
		IpAddress:            vcd.config.VCD.ExternalIp,
		Enabled:              true,
		AccelerationEnabled:  true,
		Protocol:             "http",
		Port:                 8000,
		ConnectionLimit:      5,
		ConnectionRateLimit:  10,
		ApplicationProfileId: appProfileId,
		DefaultPoolId:        serverPoolId,
	}

	_, err := edge.CreateLbVirtualServer(lbVirtualServerConfig)
	check.Assert(err, IsNil)

	// We created virtual server successfully therefore let's prepend it to cleanup list so that it
	// is deleted before the child components
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	PrependToCleanupList(TestLb, "lbVirtualServer", parentEntity, check.TestName())
}

// checkLb queries specified endpoint until it gets all responses in expectedResponses slice
func checkLb(queryUrl string, expectedResponses []string, maxRetryTimeout int) error {
	var err error
	var iterations int
	if len(expectedResponses) == 0 {
		return fmt.Errorf("no expected responses specified")
	}

	// due to the VMs taking long time to boot it needs to be at least 5 minutes
	// may be even more in slower environments
	sleepInterval := 5
	sleepIntervalDuration := time.Duration(sleepInterval) * time.Second
	if maxRetryTimeout < 5*60 { // 5 minutes
		iterations = 60
	} else {
		iterations = maxRetryTimeout / 5
	}

	fmt.Printf("# Waiting for the virtual server to accept responses (%s interval x %d iterations)"+
		"\n[_ = timeout, x = connection refused, ?(err) = unknown error, / = no nodes are up yet, "+
		". = no response from all nodes yet]: ", sleepIntervalDuration.String(), iterations)
	for i := 1; i <= iterations; i++ {
		var resp *http.Response
		resp, err = http.Get(queryUrl)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "i/o timeout"):
				fmt.Printf("_")
			case strings.Contains(err.Error(), "connect: connection refused"):
				fmt.Printf("x")
			case strings.Contains(err.Error(), "connect: network is unreachable"):
				fmt.Printf("/")
			default:
				fmt.Printf("?(%s)", err.Error())
			}
		}

		if err == nil {
			fmt.Printf(".") // progress bar when waiting for responses from all nodes
			body, _ := ioutil.ReadAll(resp.Body)
			// check if the element is in the list
			for index, value := range expectedResponses {
				if value == string(body) {
					expectedResponses = append(expectedResponses[:index], expectedResponses[index+1:]...)
					if len(expectedResponses) > 0 {
						fmt.Printf("\n# '%s' responded. Waiting for node(s) '%s': ",
							value, strings.Join(expectedResponses, ","))
					} else {
						fmt.Printf("\n# Last node '%s' responded. Exiting\n", value)
						return nil
					}
				}
			}
		}
		time.Sleep(sleepIntervalDuration)
	}

	return fmt.Errorf("timed out waiting for all nodes to be up: %s", err)
}

// addFirewallRule adds a firewall rule needed to access virtual server port on edge gateway
func addFirewallRule(vdc Vdc, vcd *TestVCD, check *C) string {
	description := "Created by: " + TestLb

	edge, err := vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)

	// Open up firewall to access edge gateway on load balancer port
	fwRule := &types.FirewallRule{
		IsEnabled:     true,
		Description:   description,
		Protocols:     &types.FirewallRuleProtocols{TCP: true},
		Port:          8000,
		DestinationIP: vcd.config.VCD.ExternalIp,
		SourceIP:      "any",
		SourcePort:    -1,
	}
	fwRules := []*types.FirewallRule{fwRule}
	task, err := edge.CreateFirewallRules("allow", fwRules)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	return description
}

// deleteFirewallRule removes firewall rule which was used for testing load balancer
func deleteFirewallRule(ruleDescription string, vdc Vdc, vcd *TestVCD, check *C) {
	edge, err := vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	rules := edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule
	for index, _ := range rules {
		if rules[index].Description == ruleDescription {
			rules = append(rules[:index], rules[index+1:]...)
		}
	}

	task, err := edge.CreateFirewallRules("allow", rules)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
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
