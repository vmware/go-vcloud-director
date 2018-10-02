/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NetRefresh(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())

	network, err := vcd.vdc.FindVDCNetwork(vcd.config.VCD.Network)

	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network)
	save_network := network

	err = network.Refresh()

	check.Assert(err, IsNil)
	check.Assert(network.OrgVDCNetwork.Name, Equals, save_network.OrgVDCNetwork.Name)
	check.Assert(network.OrgVDCNetwork.HREF, Equals, save_network.OrgVDCNetwork.HREF)
	check.Assert(network.OrgVDCNetwork.Type, Equals, save_network.OrgVDCNetwork.Type)
	check.Assert(network.OrgVDCNetwork.ID, Equals, save_network.OrgVDCNetwork.ID)
	check.Assert(network.OrgVDCNetwork.Description, Equals, save_network.OrgVDCNetwork.Description)
	check.Assert(network.OrgVDCNetwork.EdgeGateway, DeepEquals, save_network.OrgVDCNetwork.EdgeGateway)
	check.Assert(network.OrgVDCNetwork.Status, Equals, save_network.OrgVDCNetwork.Status)

}
