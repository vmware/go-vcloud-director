//go:build unit || ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestGetPseudoUUID(t *testing.T) {

	var seen = make(map[string]int)

	reUuid := regexp.MustCompile(`^[A-F0-9]{8}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{12}$`)
	for N := 0; N < 1000; N++ {
		uuid, _ := getPseudoUuid()
		if !reUuid.MatchString(uuid) {
			t.Logf("string %s doesn't look like a UUID", uuid)
			t.Fail()
		}
		previous, found := seen[uuid]
		if found {
			t.Logf("uuid %s already in the generated list at position %d", uuid, previous)
			t.Fail()
		}
		seen[uuid] = N
	}
}

func Test_getVnicIndexFromNetworkNameType(t *testing.T) {
	// Fake body - has record for network my-vdc-int-net2 twice.
	sampleBody := []byte(`
	<edgeInterfaces>
    <edgeInterface>
        <name>my-vdc-int-net</name>
        <type>internal</type>
        <index>1</index>
        <networkReference>
            <id>95bffe8e-7e67-452d-abf2-535ac298db2b</id>
            <name>my-vdc-int-net</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>10.10.10.5</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>95bffe8e-7e67-452d-abf2-535ac298db2b</portgroupId>
        <portgroupName>my-vdc-int-net</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>my-ext-network</name>
        <type>uplink</type>
        <index>0</index>
        <networkReference>
            <id>f2547dd9-e0f7-4d81-97c1-dd33e5e0fbbf</id>
            <name>my-ext-network</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>192.168.1.110</primaryAddress>
                <secondaryAddresses>
                    <ipAddress>192.168.1.118</ipAddress>
                    <ipAddress>192.168.1.115</ipAddress>
                    <ipAddress>192.168.1.116</ipAddress>
                    <ipAddress>192.168.1.117</ipAddress>
                </secondaryAddresses>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>f2547dd9-e0f7-4d81-97c1-dd33e5e0fbbf</portgroupId>
        <portgroupName></portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>Distributed Router Transit</name>
        <type>internal</type>
        <index>4</index>
        <networkReference>
            <id>00251e01-16ed-4367-a0cb-58195d21f367</id>
            <name>Distributed Router Transit</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>10.255.255.249</primaryAddress>
                <subnetMask>255.255.255.252</subnetMask>
                <subnetPrefixLength>30</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>00251e01-16ed-4367-a0cb-58195d21f367</portgroupId>
        <portgroupName>DLR_to_EDGE_my-edge-gw</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>my-vdc-int-net2</name>
        <type>internal</type>
        <index>3</index>
        <networkReference>
            <id>96a68fd4-c21a-41d6-98a4-fbf32c96480f</id>
            <name>my-vdc-int-net2</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>13.13.1.1</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>96a68fd4-c21a-41d6-98a4-fbf32c96480f</portgroupId>
        <portgroupName>my-vdc-int-net2</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>my-vdc-int-net2</name>
        <type>internal</type>
        <index>3</index>
        <networkReference>
            <id>96a68fd4-c21a-41d6-98a4-fbf32c96480f</id>
            <name>my-vdc-int-net2</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>13.13.1.1</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>96a68fd4-c21a-41d6-98a4-fbf32c96480f</portgroupId>
        <portgroupName>my-vdc-int-net2</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>subinterfaced-net</name>
        <type>subinterface</type>
        <index>10</index>
        <networkReference>
            <id>2d3e46cb-afe6-4725-9f0a-63514ebac840</id>
            <name>subinterfaced-net</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>9.9.9.1</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>2d3e46cb-afe6-4725-9f0a-63514ebac840</portgroupId>
        <portgroupName>subinterfaced-net</portgroupName>
        <isConnected>true</isConnected>
        <tunnelId>1</tunnelId>
    </edgeInterface>
    <edgeInterface>
        <name>subinterface2</name>
        <type>subinterface</type>
        <index>11</index>
        <networkReference>
            <id>3e10dd56-2a3a-47bd-aac8-07fc3f653baa</id>
            <name>subinterface2</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>55.55.55.1</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>3e10dd56-2a3a-47bd-aac8-07fc3f653baa</portgroupId>
        <portgroupName>subinterface2</portgroupName>
        <isConnected>true</isConnected>
        <tunnelId>2</tunnelId>
    </edgeInterface>
    <edgeInterface>
        <name>distributd-net</name>
        <type>distributed</type>
        <index>10</index>
        <networkReference>
            <id>0f1cc84b-517b-4435-9f7d-e42eacea1e19</id>
            <name>distributd-net</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>65.65.65.1</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>0f1cc84b-517b-4435-9f7d-e42eacea1e19</portgroupId>
        <portgroupName>distributd-net</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>distri-2</name>
        <type>distributed</type>
        <index>11</index>
        <networkReference>
            <id>8d09d23c-fd08-4c34-9ad7-21afb629cd99</id>
            <name>distri-2</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>77.77.77.1</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>8d09d23c-fd08-4c34-9ad7-21afb629cd99</portgroupId>
        <portgroupName>distri-2</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
</edgeInterfaces>`)
	vnicObject := &types.EdgeGatewayInterfaces{}
	_ = xml.Unmarshal(sampleBody, vnicObject)

	tests := []struct {
		name              string
		networkName       string
		networkType       string
		expectedVnicIndex *int
		hasError          bool
		expectedError     error
	}{
		{"ExtNetwork", "my-ext-network", types.EdgeGatewayVnicTypeUplink, addrOf(0), false, nil},
		{"OrgNetwork", "my-vdc-int-net", types.EdgeGatewayVnicTypeInternal, addrOf(1), false, nil},
		{"WithSubinterfaces", "subinterfaced-net", types.EdgeGatewayVnicTypeSubinterface, addrOf(10), false, nil},
		{"WithSubinterfaces2", "subinterface2", types.EdgeGatewayVnicTypeSubinterface, addrOf(11), false, nil},
		{"NonExistingUplink", "invalid-network-name", types.EdgeGatewayVnicTypeUplink, nil, true, ErrorEntityNotFound},
		{"NonExistingInternal", "invalid-network-name", types.EdgeGatewayVnicTypeInternal, nil, true, ErrorEntityNotFound},
		{"NonExistingSubinterface", "invalid-network-name", types.EdgeGatewayVnicTypeSubinterface, nil, true, ErrorEntityNotFound},
		{"NonExistingTrunk", "invalid-network-name", types.EdgeGatewayVnicTypeTrunk, nil, true, ErrorEntityNotFound},
		{"MoreThanOne", "my-vdc-int-net2", types.EdgeGatewayVnicTypeInternal, nil, true,
			fmt.Errorf("more than one (2) networks of type 'internal' with name 'my-vdc-int-net2' found")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnicIndex, err := getVnicIndexByNetworkNameAndType(tt.networkName, tt.networkType, vnicObject)

			if !tt.hasError && err != nil {
				t.Errorf("error was not expected: %s", err)
			}

			if tt.hasError && !strings.Contains(err.Error(), tt.expectedError.Error()) {
				t.Errorf("Got unexpected error: %s, expected: %s", err, tt.expectedError)
			}

			if vnicIndex != nil && tt.expectedVnicIndex != nil && *vnicIndex != *tt.expectedVnicIndex {
				t.Errorf("Got unexpected vNic name: %d, expected: %d", vnicIndex, tt.expectedVnicIndex)
			}

		})
	}

}

func Test_getNetworkNameTypeFromVnicIndex(t *testing.T) {
	// sample body for unit testing. The vNic 8 record is duplicated on purpose for testing.
	sampleBody := []byte(`
	<edgeInterfaces>
    <edgeInterface>
        <name>my-vdc-int-net</name>
        <type>internal</type>
        <index>1</index>
        <networkReference>
            <id>95bffe8e-7e67-452d-abf2-535ac298db2b</id>
            <name>my-vdc-int-net</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>10.10.10.5</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>95bffe8e-7e67-452d-abf2-535ac298db2b</portgroupId>
        <portgroupName>my-vdc-int-net</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>my-ext-network</name>
        <type>uplink</type>
        <index>0</index>
        <networkReference>
            <id>f2547dd9-e0f7-4d81-97c1-dd33e5e0fbbf</id>
            <name>my-ext-network</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>192.168.1.110</primaryAddress>
                <secondaryAddresses>
                    <ipAddress>192.168.1.118</ipAddress>
                    <ipAddress>192.168.1.115</ipAddress>
                    <ipAddress>192.168.1.116</ipAddress>
                    <ipAddress>192.168.1.117</ipAddress>
                </secondaryAddresses>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>f2547dd9-e0f7-4d81-97c1-dd33e5e0fbbf</portgroupId>
        <portgroupName>my-ext-network</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>Distributed Router Transit</name>
        <type>internal</type>
        <index>4</index>
        <networkReference>
            <id>00251e01-16ed-4367-a0cb-58195d21f367</id>
            <name>Distributed Router Transit</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>10.255.255.249</primaryAddress>
                <subnetMask>255.255.255.252</subnetMask>
                <subnetPrefixLength>30</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>00251e01-16ed-4367-a0cb-58195d21f367</portgroupId>
        <portgroupName>DLR_to_EDGE_my-edge-gw</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>my-vdc-int-net2</name>
        <type>internal</type>
        <index>3</index>
        <networkReference>
            <id>96a68fd4-c21a-41d6-98a4-fbf32c96480f</id>
            <name>my-vdc-int-net2</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>13.13.1.1</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>96a68fd4-c21a-41d6-98a4-fbf32c96480f</portgroupId>
        <portgroupName>my-vdc-int-net2</portgroupName>
        <isConnected>true</isConnected>
    </edgeInterface>
    <edgeInterface>
        <name>subinterfaced-net</name>
        <type>subinterface</type>
        <index>10</index>
        <networkReference>
            <id>2d3e46cb-afe6-4725-9f0a-63514ebac840</id>
            <name>subinterfaced-net</name>
            <type>com.vmware.vcloud.entity.network</type>
        </networkReference>
        <addressGroups>
            <addressGroup>
                <primaryAddress>9.9.9.1</primaryAddress>
                <subnetMask>255.255.255.0</subnetMask>
                <subnetPrefixLength>24</subnetPrefixLength>
            </addressGroup>
        </addressGroups>
        <portgroupId>2d3e46cb-afe6-4725-9f0a-63514ebac840</portgroupId>
        <portgroupName>subinterfaced-net</portgroupName>
        <isConnected>true</isConnected>
        <tunnelId>1</tunnelId>
    </edgeInterface>
</edgeInterfaces>`)
	vnicObject := &types.EdgeGatewayInterfaces{}
	_ = xml.Unmarshal(sampleBody, vnicObject)

	tests := []struct {
		name                string
		vnicIndex           int
		expectednetworkName string
		expectednetworkType string
		hasError            bool
		expectedError       error
	}{
		{"ExternalUplink", 0, "my-ext-network", "uplink", false, nil},
		{"Internal", 1, "my-vdc-int-net", "internal", false, nil},
		{"Subinterface", 10, "subinterfaced-net", "subinterface", false, nil},
		{"NonExistent", 219, "", "", true, ErrorEntityNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkName, networkType, err := getNetworkNameAndTypeByVnicIndex(tt.vnicIndex, vnicObject)

			if !tt.hasError && err != nil {
				t.Errorf("error was not expected: %s", err)
			}

			if tt.hasError && !strings.Contains(err.Error(), tt.expectedError.Error()) {
				t.Errorf("Got unexpected error: %s, expected: %s", err, tt.expectedError)
			}

			if networkName != tt.expectednetworkName {
				t.Errorf("Got unexpected network name: %s, expected: %s", networkName, tt.expectednetworkName)
			}

			if networkType != tt.expectednetworkType {
				t.Errorf("Got unexpected vNic name: %s, expected: %s", networkType, tt.expectednetworkType)
			}

		})
	}
}
