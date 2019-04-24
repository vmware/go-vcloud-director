package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// Test_updateNicParameters_multinic is meant to check functionality of a complicated
// code structure used in vm.ChangeNetworkConfig which is abstracted into
// vm.updateNicParameters() method so that it does not contain any API calls, but
// only adjust the object. Initially we hit a bug which occurred only when API returned
// NICs in random order (which happens rarely and was only seen with 3 NICs).
func (vcd *TestVCD) Test_updateNicParameters_multiNIC(check *C) {

	// Mock VM struct
	c := Client{}
	vm := NewVM(&c)

	// Sample config which is rendered by .tf schema parsed
	tfCfg := []map[string]interface{}{
		map[string]interface{}{
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "POOL",
			"ip":                 "",
			"is_primary":         false,
		},
		map[string]interface{}{
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "DHCP",
			"ip":                 "",
			"is_primary":         true,
		},
		map[string]interface{}{
			"orgnetwork":         "multinic-net2",
			"ip_allocation_mode": "NONE",
			"ip":                 "",
			"is_primary":         false,
		},
	}

	// A sample NetworkConnectionSection object simulating API returning ordered list
	vcdConfig := types.NetworkConnectionSection{
		PrimaryNetworkConnectionIndex: 0,
		NetworkConnection: []*types.NetworkConnection{
			&types.NetworkConnection{
				Network:                 "multinic-net",
				NetworkConnectionIndex:  0,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:00",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
			&types.NetworkConnection{
				Network:                 "multinic-net",
				NetworkConnectionIndex:  1,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:01",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
			&types.NetworkConnection{
				Network:                 "multinic-net",
				NetworkConnectionIndex:  2,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:02",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
		},
	}

	// NIC configuration when API returns an ordered list
	vcdCfg := &vcdConfig
	vm.updateNicParameters(tfCfg, vcdCfg)

	// Test NIC updates when API returns an unordered list
	// Swap two &types.NetworkConnection so that it is not ordered correctly
	vcdConfig2 := vcdConfig
	vcdConfig2.NetworkConnection[2], vcdConfig2.NetworkConnection[0] = vcdConfig2.NetworkConnection[0], vcdConfig2.NetworkConnection[2]
	vcdCfg2 := &vcdConfig2
	vm.updateNicParameters(tfCfg, vcdCfg2)

	var tableTests = []struct {
		tfConfig []map[string]interface{}
		vcdConfig *types.NetworkConnectionSection
	}{
		{tfConfig: tfCfg, vcdConfig:vcdCfg},	// Ordered NIC list
		{tfConfig: tfCfg, vcdConfig:vcdCfg2},	// Unordered NIC list
	}

	for _, tableTest := range tableTests {

		// Check that primary interface is reset to 1 as hardcoded in tfCfg "is_primary" parameter
		check.Assert(vcdCfg.PrimaryNetworkConnectionIndex, Equals, 1)

		for loopIndex := range tableTest.vcdConfig.NetworkConnection {
			nicSlot := tableTest.vcdConfig.NetworkConnection[loopIndex].NetworkConnectionIndex

			check.Assert(tableTest.vcdConfig.NetworkConnection[loopIndex].IPAddressAllocationMode, Equals, tableTest.tfConfig[nicSlot]["ip_allocation_mode"].(string))
			check.Assert(tableTest.vcdConfig.NetworkConnection[loopIndex].IsConnected, Equals, true)
			check.Assert(tableTest.vcdConfig.NetworkConnection[loopIndex].NeedsCustomization, Equals, true)
			check.Assert(tableTest.vcdConfig.NetworkConnection[loopIndex].Network, Equals, tableTest.tfConfig[nicSlot]["orgnetwork"].(string))
		}
	}
}
