//go:build vm || unit || ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"reflect"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func init() {
	testingTags["unit"] = "vm_unit_test.go"
}

// Test_updateNicParameters_multinic is meant to check functionality of a complicated
// code structure used in vm.ChangeNetworkConfig which is abstracted into
// vm.updateNicParameters() method so that it does not contain any API calls, but
// only adjust the object which is meant to be sent to API. Initially we hit a bug
// which occurred only when API returned NICs in random order.
func Test_VMupdateNicParameters_multiNIC(t *testing.T) {

	// Mock VM struct
	c := Client{}
	vm := NewVM(&c)

	// Sample config which is rendered by .tf schema parsed
	tfCfg := []map[string]interface{}{
		map[string]interface{}{
			"network_name":       "multinic-net",
			"ip_allocation_mode": "POOL",
			"ip":                 "",
			"is_primary":         false,
		},
		map[string]interface{}{
			"network_name":       "multinic-net",
			"ip_allocation_mode": "DHCP",
			"ip":                 "",
			"is_primary":         true,
		},
		map[string]interface{}{
			"ip_allocation_mode": "NONE",
		},
		map[string]interface{}{
			"network_name":       "multinic-net2",
			"ip_allocation_mode": "MANUAL",
			"ip":                 "1.1.1.1",
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
			&types.NetworkConnection{
				Network:                 "multinic-net",
				NetworkConnectionIndex:  3,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:03",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
		},
	}

	// NIC configuration when API returns an ordered list
	vcdCfg := &vcdConfig
	err := vm.updateNicParameters(tfCfg, vcdCfg)
	if err != nil {
		t.Error(err)
	}

	// Test NIC updates when API returns an unordered list
	// Swap two &types.NetworkConnection so that it is not ordered correctly
	vcdConfig2 := vcdConfig
	vcdConfig2.NetworkConnection[2], vcdConfig2.NetworkConnection[0] = vcdConfig2.NetworkConnection[0], vcdConfig2.NetworkConnection[2]
	vcdCfg2 := &vcdConfig2
	err = vm.updateNicParameters(tfCfg, vcdCfg2)
	if err != nil {
		t.Error(err)
	}

	var tableTests = []struct {
		title     string
		tfConfig  []map[string]interface{}
		vcdConfig *types.NetworkConnectionSection
	}{
		{"Ordered NICs list", tfCfg, vcdCfg},
		{"Unordered NIC list", tfCfg, vcdCfg2},
	}

	for _, tableTest := range tableTests {
		t.Run(tableTest.title, func(t *testing.T) {
			// Check that primary interface is reset to 1 as hardcoded in tfCfg "is_primary" parameter
			if vcdCfg.PrimaryNetworkConnectionIndex != 1 {
				t.Errorf("PrimaryNetworkConnectionIndex expected: 1, got: %d", vcdCfg.PrimaryNetworkConnectionIndex)
			}

			for loopIndex := range tableTest.vcdConfig.NetworkConnection {
				vcdNic := tableTest.vcdConfig.NetworkConnection[loopIndex]
				vcdNicSlot := vcdNic.NetworkConnectionIndex
				tfNic := tableTest.tfConfig[vcdNicSlot]

				if vcdNic.IPAddressAllocationMode != tfNic["ip_allocation_mode"].(string) {
					t.Errorf("IPAddressAllocationMode expected: %s, got: %s", tfNic["ip_allocation_mode"].(string), vcdNic.IPAddressAllocationMode)
				}

				if vcdNic.IPAddressAllocationMode != tfNic["ip_allocation_mode"].(string) {
					t.Errorf("IPAddressAllocationMode expected: %s, got: %s", tfNic["ip_allocation_mode"].(string), vcdNic.IPAddressAllocationMode)
				}

				if vcdNic.IsConnected != true {
					t.Errorf("IsConnected expected: true, got: %t", vcdNic.IsConnected)
				}

				if vcdNic.NeedsCustomization != true {
					t.Errorf("NeedsCustomization expected: true, got: %t", vcdNic.NeedsCustomization)
				}

				if vcdNic.IPAddressAllocationMode != types.IPAllocationModeNone {
					if vcdNic.Network != tfNic["network_name"].(string) {
						t.Errorf("Network expected: %s, got: %s", tfNic["network_name"].(string), vcdNic.Network)
					}
				} else {
					if vcdNic.Network != "none" {
						t.Errorf("Network expected: none, got: %s", vcdNic.Network)
					}
				}
			}
		})
	}
}

// // Test_updateNicParameters_singleNIC is meant to check functionality when single NIC
// // is being configured and meant to check functionality so that the function is able
// // to cover legacy scenarios when Terraform provider was able to create single IP only.
// // TODO v3.0 this test should become irrelevant once `ip` and `network_name` parameters are removed.
func Test_VMupdateNicParameters_singleNIC(t *testing.T) {
	// Mock VM struct
	c := Client{}
	vm := NewVM(&c)

	tfCfgDHCP := []map[string]interface{}{
		map[string]interface{}{
			"network_name": "multinic-net",
			"ip":           "dhcp",
		},
	}

	tfCfgAllocated := []map[string]interface{}{
		map[string]interface{}{
			"network_name": "multinic-net",
			"ip":           "allocated",
		},
	}

	tfCfgNone := []map[string]interface{}{
		map[string]interface{}{
			"network_name": "multinic-net",
			"ip":           "none",
		},
	}

	tfCfgManual := []map[string]interface{}{
		map[string]interface{}{
			"network_name": "multinic-net",
			"ip":           "1.1.1.1",
		},
	}

	tfCfgInvalidIp := []map[string]interface{}{
		map[string]interface{}{
			"network_name": "multinic-net",
			"ip":           "invalidIp",
		},
	}

	tfCfgNoNetworkName := []map[string]interface{}{
		map[string]interface{}{
			"ip": "invalidIp",
		},
	}

	vcdConfig := types.NetworkConnectionSection{
		PrimaryNetworkConnectionIndex: 1,
		NetworkConnection: []*types.NetworkConnection{
			&types.NetworkConnection{
				Network:                 "singlenic-net",
				NetworkConnectionIndex:  0,
				IPAddress:               "",
				IsConnected:             true,
				MACAddress:              "00:00:00:00:00:00",
				IPAddressAllocationMode: "POOL",
				NetworkAdapterType:      "VMXNET3",
			},
		},
	}

	var tableTests = []struct {
		title                           string
		tfConfig                        []map[string]interface{}
		expectedIPAddressAllocationMode string
		expectedIPAddress               string
		mustNotError                    bool
	}{
		{"IPAllocationModeDHCP", tfCfgDHCP, types.IPAllocationModeDHCP, "Any", true},
		{"IPAllocationModePool", tfCfgAllocated, types.IPAllocationModePool, "Any", true},
		{"IPAllocationModeNone", tfCfgNone, types.IPAllocationModeNone, "Any", true},
		{"IPAllocationModeManual", tfCfgManual, types.IPAllocationModeManual, tfCfgManual[0]["ip"].(string), true},
		{"IPAllocationModeDHCPInvalidIP", tfCfgInvalidIp, types.IPAllocationModeDHCP, "Any", true},
		{"ErrNoNetworkName", tfCfgNoNetworkName, types.IPAllocationModeDHCP, "Any", false},
	}

	for _, tableTest := range tableTests {

		t.Run(tableTest.title, func(t *testing.T) {
			vcdCfg := &vcdConfig
			err := vm.updateNicParameters(tableTest.tfConfig, vcdCfg) // Execute parsing procedure

			// if we got an error which was expected abandon the subtest
			if err != nil && tableTest.mustNotError {
				t.Errorf("unexpected error got: %s", err)
				return
			}

			if vcdCfg.PrimaryNetworkConnectionIndex != 0 {
				t.Errorf("PrimaryNetworkConnectionIndex expected: 0, got: %d", vcdCfg.PrimaryNetworkConnectionIndex)
			}

			if vcdCfg.NetworkConnection[0].IPAddressAllocationMode != tableTest.expectedIPAddressAllocationMode {
				t.Errorf("IPAddressAllocationMode expected: %s, got: %s", tableTest.expectedIPAddressAllocationMode, vcdCfg.NetworkConnection[0].IPAddressAllocationMode)
			}

			if vcdCfg.NetworkConnection[0].IPAddress != tableTest.expectedIPAddress {
				t.Errorf("IPAddress expected: %s, got: %s", tableTest.expectedIPAddress, vcdCfg.NetworkConnection[0].IPAddress)
			}
		})

	}
}

// TestProductSectionList_SortByPropertyKeyName validates that a
// SortByPropertyKeyName() works on ProductSectionList and can handle empty properties as well as
// sort correctly
func TestProductSectionList_SortByPropertyKeyName(t *testing.T) {
	sliceProductSection := &types.ProductSectionList{
		ProductSection: &types.ProductSection{},
	}

	emptyProductSection := &types.ProductSectionList{
		ProductSection: &types.ProductSection{
			Info: "Custom properties",
		},
	}

	// unordered list for test
	sortOrder := &types.ProductSectionList{
		ProductSection: &types.ProductSection{
			Info: "Custom properties",
			Property: []*types.Property{
				&types.Property{
					UserConfigurable: false,
					Key:              "sys_owner",
					Label:            "sys_owner_label",
					Type:             "string",
					DefaultValue:     "sys_owner_default",
					Value:            &types.Value{Value: "test"},
				},
				&types.Property{
					UserConfigurable: true,
					Key:              "asset_tag",
					Label:            "asset_tag_label",
					Type:             "string",
					DefaultValue:     "asset_tag_default",
					Value:            &types.Value{Value: "xxxyyy"},
				},
				&types.Property{
					UserConfigurable: true,
					Key:              "guestinfo.config.bootstrap.ip",
					Label:            "guestinfo.config.bootstrap.ip_label",
					Type:             "string",
					DefaultValue:     "default_ip",
					Value:            &types.Value{Value: "192.168.12.180"},
				},
			},
		},
	}
	// correct state after ordering
	expectedSortedOrder := &types.ProductSectionList{
		ProductSection: &types.ProductSection{
			Info: "Custom properties",
			Property: []*types.Property{
				&types.Property{
					UserConfigurable: true,
					Key:              "asset_tag",
					Label:            "asset_tag_label",
					Type:             "string",
					DefaultValue:     "asset_tag_default",
					Value:            &types.Value{Value: "xxxyyy"},
				},
				&types.Property{
					UserConfigurable: true,
					Key:              "guestinfo.config.bootstrap.ip",
					Label:            "guestinfo.config.bootstrap.ip_label",
					Type:             "string",
					DefaultValue:     "default_ip",
					Value:            &types.Value{Value: "192.168.12.180"},
				},
				&types.Property{
					UserConfigurable: false,
					Key:              "sys_owner",
					Label:            "sys_owner_label",
					Type:             "string",
					DefaultValue:     "sys_owner_default",
					Value:            &types.Value{Value: "test"},
				},
			},
		},
	}

	tests := []struct {
		name          string
		setValue      *types.ProductSectionList
		expectedValue *types.ProductSectionList
	}{
		{name: "Empty", setValue: emptyProductSection, expectedValue: emptyProductSection},
		{name: "Slice", setValue: sliceProductSection, expectedValue: sliceProductSection},
		{name: "SortOrder", setValue: sortOrder, expectedValue: expectedSortedOrder},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setValue
			p.SortByPropertyKeyName()

			if !reflect.DeepEqual(p, tt.expectedValue) {
				t.Errorf("Objects were not deeply equal: \n%#+v\n, got:\n %#+v\n", tt.expectedValue, p)
			}

		})
	}
}
