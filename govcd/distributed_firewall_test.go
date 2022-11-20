//go:build functional || ALL

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_DistributedFirewall(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	// Retrieve a NSX-T VDC
	nsxtVdc, err := org.GetAdminVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	err = nsxtVdc.EnableDistributedFirewall()
	// NSX-T VDCs don't support NSX-V distributed firewalls. We expect an error here.
	check.Assert(err, NotNil)

	enabled, err := nsxtVdc.IsDistributedFirewallEnabled()
	check.Assert(err, IsNil)
	check.Assert(enabled, Equals, false)

	// Retrieve a NSX-V VDC
	vdc, err := org.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// EnableDistributedFirewall is an idempotent operation. It can be repeated on an already enabled VDC
	// without errors.
	err = vdc.EnableDistributedFirewall()
	check.Assert(err, IsNil)

	enabled, err = vdc.IsDistributedFirewallEnabled()
	check.Assert(err, IsNil)
	check.Assert(enabled, Equals, true)

	// Repeat enable operation
	err = vdc.EnableDistributedFirewall()
	check.Assert(err, IsNil)

	enabled, err = vdc.IsDistributedFirewallEnabled()
	check.Assert(err, IsNil)
	check.Assert(enabled, Equals, true)

	err = vdc.DisableDistributedFirewall()
	check.Assert(err, IsNil)

	enabled, err = vdc.IsDistributedFirewallEnabled()
	check.Assert(err, IsNil)
	check.Assert(enabled, Equals, false)
}
