//go:build vsphere || functional || ALL

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
	"strings"
)

func (vcd *TestVCD) Test_GetResourcePools(check *C) {

	if !vcd.client.Client.IsSysAdmin {
		check.Skip("this test requires system administrator privileges")
	}
	vcenters, err := vcd.client.GetAllVCenters(nil)
	check.Assert(err, IsNil)

	check.Assert(len(vcenters) > 0, Equals, true)

	vc := vcenters[0]

	allResourcePools, err := vc.GetAllResourcePools(nil)
	check.Assert(err, IsNil)

	for i, rp := range allResourcePools {
		rpByID, err := vc.GetResourcePoolById(rp.ResourcePool.Moref)
		check.Assert(err, IsNil)
		check.Assert(rpByID.ResourcePool.Moref, Equals, rp.ResourcePool.Moref)
		check.Assert(rpByID.ResourcePool.Name, Equals, rp.ResourcePool.Name)
		rpByName, err := vc.GetResourcePoolByName(rp.ResourcePool.Name)
		if err != nil && strings.Contains(err.Error(), "more than one") {
			if testVerbose {
				fmt.Printf("%s\n", err)
			}
			continue
		}
		check.Assert(err, IsNil)
		check.Assert(rpByName.ResourcePool.Moref, Equals, rp.ResourcePool.Moref)
		check.Assert(rpByName.ResourcePool.Name, Equals, rp.ResourcePool.Name)
		if testVerbose {
			fmt.Printf("%2d %# v\n", i, pretty.Formatter(rp.ResourcePool))
		}
		hw, err := rp.GetAvailableHardwareVersions()
		check.Assert(err, IsNil)
		if testVerbose {
			fmt.Printf("%s %#v\n", rp.ResourcePool.Name, hw)
		}
	}
}
