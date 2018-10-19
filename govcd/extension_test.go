/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetExternalNetwork(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	networkName := vcd.config.VCD.ExternalNetwork
	if networkName == "" {
		check.Skip("No external network provided")
	}
	externalNetwork, err := GetExternalNetworkByName(vcd.client, networkName)
	check.Assert(err, IsNil)
	LogExternalNetwork(*externalNetwork)
	check.Assert(externalNetwork.HREF, Not(Equals), "")
	expectedType := "application/vnd.vmware.admin.extension.network+xml"
	check.Assert(externalNetwork.Name, Equals, networkName)
	check.Assert(externalNetwork.Type, Equals, expectedType)
}
