//go:build providervdc || functional || ALL

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetNetworkPools(check *C) {

	if vcd.skipAdminTests {
		check.Skip("this test requires system administrator privileges")
	}
	knownNetworkPoolName := vcd.config.VCD.NsxtProviderVdc.NetworkPool
	networkPools, err := vcd.client.GetNetworkPoolSummaries(nil)
	check.Assert(err, IsNil)
	check.Assert(len(networkPools) > 0, Equals, true)

	checkNetworkPoolName := false
	foundNetworkPool := false
	if knownNetworkPoolName != "" {
		checkNetworkPoolName = true
	}

	for i, nps := range networkPools {
		if nps.Name == knownNetworkPoolName {
			foundNetworkPool = true
		}
		networkPoolById, err := vcd.client.GetNetworkPoolById(nps.Id)
		check.Assert(err, IsNil)
		check.Assert(networkPoolById, NotNil)
		check.Assert(networkPoolById.NetworkPool.Id, Equals, nps.Id)
		check.Assert(networkPoolById.NetworkPool.Name, Equals, nps.Name)

		networkPoolByName, err := vcd.client.GetNetworkPoolByName(nps.Name)
		check.Assert(err, IsNil)
		check.Assert(networkPoolByName, NotNil)
		check.Assert(networkPoolByName.NetworkPool.Id, Equals, nps.Id)
		check.Assert(networkPoolByName.NetworkPool.Name, Equals, nps.Name)
		if testVerbose {
			fmt.Printf("%d, %# v\n", i, pretty.Formatter(networkPoolByName.NetworkPool))
		}
	}
	if checkNetworkPoolName {
		check.Assert(foundNetworkPool, Equals, true)
	}

}
