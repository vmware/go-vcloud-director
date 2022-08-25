//go:build providervdc || functional || ALL
// +build providervdc functional ALL

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
)

func init() {
	testingTags["providervdc"] = "provider_vdc_test.go"
}

func (vcd *TestVCD) Test_GetProviderVdc(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	var providerVdcs []*ProviderVdc
	providerVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)
	providerVdcs = append(providerVdcs, providerVdc)
	providerVdc, err = vcd.client.GetProviderVdcById(providerVdc.ProviderVdc.ID)
	check.Assert(err, IsNil)
	providerVdcs = append(providerVdcs, providerVdc)
	providerVdc, err = vcd.client.GetProviderVdcByHref(providerVdc.ProviderVdc.HREF)
	check.Assert(err, IsNil)
	providerVdcs = append(providerVdcs, providerVdc)

	// Common asserts
	for _, providerVdc := range providerVdcs {
		check.Assert(providerVdc.ProviderVdc.Name, Equals, vcd.config.VCD.NsxtProviderVdc.Name)
		check.Assert(providerVdc.ProviderVdc.StorageProfiles.ProviderVdcStorageProfile[0].Name, Equals, vcd.config.VCD.NsxtProviderVdc.StorageProfile)
	}

}
