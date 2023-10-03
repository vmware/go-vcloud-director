//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtManager(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	vcd.skipIfNotSysAdmin(check)

	nsxtManager, err := vcd.client.GetNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(nsxtManager, NotNil)
}
