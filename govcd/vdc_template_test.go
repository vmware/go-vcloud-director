//go:build vdc || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VdcTemplate(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}

	providerVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)
	check.Assert(providerVdc, NotNil)

	vcd.client.CreateVdcTemplate(types.VMWVdcTemplate{
		NetworkBackingType:   "NSX_T",
		ProviderVdcReference: []*types.VMWVdcTemplateProviderVdcSpecification{{}},
		Name:                 check.TestName(),
		Description:          check.TestName(),
		TenantName:           check.TestName() + "_Tenant",
		TenantDescription:    check.TestName() + "_Tenant",

		VdcTemplateSpecification: &types.VMWVdcTemplateSpecification{
			AutomaticNetworkPoolReference: nil,
			FastProvisioningEnabled:       true,
			GatewayConfiguration:          nil,
			IncludeMemoryOverhead:         false,
			IsElastic:                     false,
			NetworkPoolReference:          nil,
			NetworkProfileConfiguration:   nil,
			NicQuota:                      0,
			ProvisionedNetworkQuota:       0,
			StorageProfile:                nil,
			ThinProvision:                 true,
			VmQuota:                       0,
		},
	})
}
