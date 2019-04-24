package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VMupdateNicParameters(check *C) {
	//func Test_VMupdateNicParameters(t *testing.T) {

	tfConfig := []map[string]interface{}{
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
			"orgnetwork":         "multinic-net",
			"ip_allocation_mode": "MANUAL",
			"ip":                 "11.10.0.170",
			"is_primary":         false,
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
		PrimaryNetworkConnectionIndex: 1,
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

	// A sample NetworkConnectionSection object simulating API returning unordered list
	vcdConfig2 := types.NetworkConnectionSection{
		PrimaryNetworkConnectionIndex: 1,
		NetworkConnection: []*types.NetworkConnection{
			&types.NetworkConnection{
				Network:                 "multinic-net",
				NetworkConnectionIndex:  2,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:02",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
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
		},
	}

	// Mock VM struct
	c := Client{}
	vm := NewVM(&c)

	vcdCfg := &vcdConfig
	vm.updateNicParameters(tfConfig, vcdCfg)

	for loopIndex := range vcdCfg.NetworkConnection {
		nicSlot := vcdCfg.NetworkConnection[loopIndex].NetworkConnectionIndex

		check.Assert(vcdCfg.NetworkConnection[loopIndex].IPAddressAllocationMode, Equals, tfConfig[nicSlot]["ip_allocation_mode"].(string))
		check.Assert(vcdCfg.NetworkConnection[loopIndex].Network, Equals, tfConfig[nicSlot]["orgnetwork"].(string))
	}

	vcdCfg2 := &vcdConfig2

	vm.updateNicParameters(tfConfig, vcdCfg2)

	for loopIndex := range vcdCfg2.NetworkConnection {
		nicSlot := vcdCfg2.NetworkConnection[loopIndex].NetworkConnectionIndex

		check.Assert(vcdCfg2.NetworkConnection[loopIndex].IPAddressAllocationMode, Equals, tfConfig[nicSlot]["ip_allocation_mode"].(string))
		check.Assert(vcdCfg2.NetworkConnection[loopIndex].Network, Equals, tfConfig[nicSlot]["orgnetwork"].(string))
	}

}
