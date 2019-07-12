// +build unit lb functional ALL 
// +build !skipLong
/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

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

	return
}

func testSetupVApp() {
	
}