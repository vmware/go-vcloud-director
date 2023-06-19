//go:build vsphere || functional || ALL

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetResourcePools(check *C) {

	vcenters, err := vcd.client.GetAllVcenters()
	check.Assert(err, IsNil)

	check.Assert(len(vcenters) > 0, Equals, true)

	vc := vcenters[0]

	resourcePools, err := vc.GetAllAvailableResourcePools()

	check.Assert(err, IsNil)

	for i, rp := range resourcePools {
		fmt.Printf("%2d %# v\n", i, pretty.Formatter(rp.ResourcePool))
		hw, err := rp.GetAvailableHardwareVersions()
		check.Assert(err, IsNil)
		fmt.Printf(" %# v\n", pretty.Formatter(hw))

	}
}
