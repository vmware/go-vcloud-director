// +build lb functional integration ALL
// +build !skipLong

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LB load balancer integration test
// 1. Validates that all needed parameters are here
// 2. Uploads or reuses media.photonOsOvaPath OVA image
// 3. Creates RAW vApp and attaches vDC network to it
// 4. Spawns two VMs with configuration script to server HTTP traffic
// 5. Sets up load balancer
// 6. Probes load balancer virtual server's external IP (edge gateway IP) for traffic
// being server in 2 VMs
// 7. Tears down
func (vcd *TestVCD) Test_LB(check *C) {
	ctx := context.Background()

	// Validate prerequisites
	validateTestLbPrerequisites(ctx, vcd, check)

	vdc, edge, vappTemplate, vapp, desiredNetConfig, err := vcd.createAndGetResourcesForVmCreation(ctx, check, TestLb)
	check.Assert(err, IsNil)

	// The script below creates a file /tmp/node/server with single value `name` being set in it.
	// It also disables iptables and spawns simple Python 3 HTTP server listening on port 8000
	// in background which serves the just created `server` file.
	vm1CustomizationScript := "mkdir /tmp/node && cd /tmp/node && echo -n 'FirstNode' > server && " +
		"/bin/systemctl stop iptables && /usr/bin/python3 -m http.server 8000 &"
	vm2CustomizationScript := "mkdir /tmp/node && cd /tmp/node && echo -n 'SecondNode' > server && " +
		"/bin/systemctl stop iptables && /usr/bin/python3 -m http.server 8000 &"

	vm1, err := spawnVM(ctx, "FirstNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, vm1CustomizationScript, true)
	check.Assert(err, IsNil)
	vm2, err := spawnVM(ctx, "SecondNode", 512, *vdc, *vapp, desiredNetConfig, vappTemplate, check, vm2CustomizationScript, true)
	check.Assert(err, IsNil)

	// Get IPs allocated to the VMs
	ip1 := vm1.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
	ip2 := vm2.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress

	fmt.Printf("# VM '%s' got IP '%s' in vDC network %s\n", vm1.VM.Name, ip1, vcd.config.VCD.Network.Net1)
	fmt.Printf("# VM '%s' got IP '%s' in vDC network %s\n", vm2.VM.Name, ip2, vcd.config.VCD.Network.Net1)

	fmt.Printf("# Setting up load balancer for VMs: '%s' (%s), '%s' (%s)\n", vm1.VM.Name, ip1, vm2.VM.Name, ip2)

	fmt.Printf("# Creating firewall rule for load balancer virtual server access. ")
	ruleDescription := addFirewallRule(ctx, *vdc, vcd, check)
	fmt.Printf("Done\n")

	// Build load balancer
	buildLb(ctx, *edge, ip1, ip2, vcd, check)

	// Cache current load balancer settings for change validation in the end
	beforeLb, beforeLbXml := testCacheLoadBalancer(ctx, *edge, check)

	// Enable load balancer globally
	fmt.Printf("# Enabling load balancer with acceleration: ")
	_, err = edge.UpdateLBGeneralParams(ctx, true, true, true, "warning")
	check.Assert(err, IsNil)
	fmt.Printf("Done\n")

	// Using external edge gateway IP for
	queryUrl := "http://" + vcd.config.VCD.ExternalIp + ":8000/server"
	fmt.Printf("# Querying load balancer for expected responses at %s\n", queryUrl)
	queryErr := checkLb(queryUrl, []string{vm1.VM.Name, vm2.VM.Name}, vapp.client.MaxRetryTimeout)

	// Remove firewall rule
	fmt.Printf("# Deleting firewall rule used for load balancer virtual server access. ")
	deleteFirewallRule(ctx, ruleDescription, *vdc, vcd, check)
	fmt.Printf("Done\n")

	// Restore global load balancer configuration
	fmt.Printf("# Restoring load balancer global configuration: ")
	_, err = edge.UpdateLBGeneralParams(ctx, beforeLb.Enabled, beforeLb.AccelerationEnabled,
		beforeLb.Logging.Enable, beforeLb.Logging.LogLevel)
	check.Assert(err, IsNil)
	fmt.Printf("Done\n")

	// Validate load balancer configuration against initially cached version
	fmt.Printf("# Validating load balancer XML structure: ")
	testCheckLoadBalancerConfig(ctx, beforeLb, beforeLbXml, *edge, check)
	fmt.Printf("Done\n")

	// Finally after some cleanups - check if querying succeeded
	check.Assert(queryErr, IsNil)
}

// validateTestLbPrerequisites verifies the following:
// * Edge Gateway is set in config
// * ExternalIp is set in config (will be edge gateway external IP)
// * PhotonOsOvaPath is set (will be used for spawning VMs)
// * Edge Gateway can be found and it has advanced networking enabled (a must for load balancers)
func validateTestLbPrerequisites(ctx context.Context, vcd *TestVCD, check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}

	if vcd.config.VCD.ExternalIp == "" {
		check.Skip("Skipping test because no edge gateway external IP given")
	}

	edge, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

}

// buildLB establishes an HTTP load balancer for 2 IPs specified as arguments
func buildLb(ctx context.Context, edge EdgeGateway, node1Ip, node2Ip string, vcd *TestVCD, check *C) {

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

	err := deleteLbVirtualServerIfExists(edge, lbVirtualServerConfig.Name)
	check.Assert(err, IsNil)

	_, err = edge.CreateLbVirtualServer(ctx, lbVirtualServerConfig)
	check.Assert(err, IsNil)

	// We created virtual server successfully therefore let's prepend it to cleanup list so that it
	// is deleted before the child components
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	PrependToCleanupList(TestLb, "lbVirtualServer", parentEntity, check.TestName())
}

// checkLb queries specified endpoint until it gets all responses in expectedResponses slice
func checkLb(queryUrl string, expectedResponses []string, maxRetryTimeout int) error {
	var err error
	if len(expectedResponses) == 0 {
		return fmt.Errorf("no expected responses specified")
	}

	retryTimeout := maxRetryTimeout
	// due to the VMs taking long time to boot it needs to be at least 5 minutes
	// may be even more in slower environments
	if maxRetryTimeout < 5*60 { // 5 minutes
		retryTimeout = 5 * 60 // 5 minutes
	}

	timeOutAfterInterval := time.Duration(retryTimeout) * time.Second
	timeoutAfter := time.After(timeOutAfterInterval)
	tick := time.NewTicker(time.Duration(5) * time.Second)

	httpClient := &http.Client{Timeout: 5 * time.Second}

	fmt.Printf("# Waiting for the virtual server to accept responses (timeout after %s)"+
		"\n[_ = timeout, x = connection refused, ?(err) = unknown error, / = no nodes are up yet, "+
		". = no response from all nodes yet]: ", timeOutAfterInterval.String())

	for {
		select {
		case <-timeoutAfter:
			return fmt.Errorf("timed out waiting for all nodes to be up")
		case <-tick.C:
			var resp *http.Response
			resp, err = httpClient.Get(queryUrl)
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
				resp.Body.Close()
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
		}
	}

}

// addFirewallRule adds a firewall rule needed to access virtual server port on edge gateway
func addFirewallRule(ctx context.Context, vdc Vdc, vcd *TestVCD, check *C) string {
	description := "Created by: " + TestLb

	edge, err := vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
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
	task, err := edge.CreateFirewallRules(ctx, "allow", fwRules)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	return description
}

// deleteFirewallRule removes firewall rule which was used for testing load balancer
func deleteFirewallRule(ctx context.Context, ruleDescription string, vdc Vdc, vcd *TestVCD, check *C) {
	edge, err := vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	rules := edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule
	for index := range rules {
		if rules[index].Description == ruleDescription {
			rules = append(rules[:index], rules[index+1:]...)
		}
	}

	task, err := edge.CreateFirewallRules(ctx, "allow", rules)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
}
