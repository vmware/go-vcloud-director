// +build unit ALL

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
	// sample body for unit testing. vNic4 interface is made to contain duplicate network
	// name on purpose (not a real response from API)
	sampleBody := []byte(`
	<vnics>
  <vnic>
    <label>vNic_0</label>
    <name>my-ext-network</name>
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
    <mtu>1500</mtu>
    <type>uplink</type>
    <isConnected>true</isConnected>
    <index>0</index>
    <portgroupId>f2547dd9-e0f7-4d81-97c1-dd33e5e0fbbf</portgroupId>
    <portgroupName>my-ext-network</portgroupName>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_1</label>
    <name>vnic1</name>
    <addressGroups>
      <addressGroup>
        <primaryAddress>10.10.10.5</primaryAddress>
        <subnetMask>255.255.255.0</subnetMask>
        <subnetPrefixLength>24</subnetPrefixLength>
      </addressGroup>
    </addressGroups>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>true</isConnected>
    <index>1</index>
    <portgroupId>95bffe8e-7e67-452d-abf2-535ac298db2b</portgroupId>
    <portgroupName>my-vdc-int-net</portgroupName>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_2</label>
    <name>vdcf9daf2da-b4f9-4921-a2f4-d77a943a381c</name>
    <addressGroups/>
    <mtu>1600</mtu>
    <type>trunk</type>
    <subInterfaces>
      <subInterface>
        <isConnected>true</isConnected>
        <label>vNic_10</label>
        <name>vnic1-subinterface</name>
        <index>10</index>
        <tunnelId>1</tunnelId>
        <logicalSwitchId>9a222ba9-23bc-41bf-b10f-0168f87b80ab</logicalSwitchId>
        <logicalSwitchName>with-subinterfaces</logicalSwitchName>
        <enableSendRedirects>true</enableSendRedirects>
        <mtu>1500</mtu>
        <addressGroups>
          <addressGroup>
            <primaryAddress>3.3.3.1</primaryAddress>
            <subnetMask>255.255.255.0</subnetMask>
            <subnetPrefixLength>24</subnetPrefixLength>
          </addressGroup>
        </addressGroups>
      </subInterface>
      <subInterface>
        <isConnected>true</isConnected>
        <label>vNic_11</label>
        <name>vnic2-subinterface</name>
        <index>11</index>
        <tunnelId>2</tunnelId>
        <logicalSwitchId>aecdf098-13ea-4ddf-9c4b-6a06372aa163</logicalSwitchId>
        <logicalSwitchName>subinterface2</logicalSwitchName>
        <enableSendRedirects>true</enableSendRedirects>
        <mtu>1500</mtu>
        <addressGroups>
          <addressGroup>
            <primaryAddress>7.7.7.1</primaryAddress>
            <subnetMask>255.255.255.0</subnetMask>
            <subnetPrefixLength>24</subnetPrefixLength>
          </addressGroup>
        </addressGroups>
      </subInterface>
      <subInterface>
        <isConnected>true</isConnected>
        <label>vNic_12</label>
        <name>vnic3-subinterface</name>
        <index>12</index>
        <tunnelId>3</tunnelId>
        <logicalSwitchId>c489d85d-df17-409c-810e-12af5aa5c1c2</logicalSwitchId>
        <logicalSwitchName>subinterface-subnet-clash</logicalSwitchName>
        <enableSendRedirects>true</enableSendRedirects>
        <mtu>1500</mtu>
        <addressGroups>
          <addressGroup>
            <primaryAddress>4.3.3.1</primaryAddress>
            <subnetMask>255.255.255.0</subnetMask>
            <subnetPrefixLength>24</subnetPrefixLength>
          </addressGroup>
        </addressGroups>
      </subInterface>
    </subInterfaces>
    <isConnected>true</isConnected>
    <index>2</index>
    <portgroupId>dvportgroup-20228</portgroupId>
    <portgroupName>dvs.VCDVS-Trunk-Portgroup-vdcf9daf2da-b4f9-4921-a2f4-d77a943a381c</portgroupName>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_3</label>
    <name>vnic3</name>
    <addressGroups>
      <addressGroup>
        <primaryAddress>13.13.1.1</primaryAddress>
        <subnetMask>255.255.255.0</subnetMask>
        <subnetPrefixLength>24</subnetPrefixLength>
      </addressGroup>
    </addressGroups>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>true</isConnected>
    <index>3</index>
    <portgroupId>96a68fd4-c21a-41d6-98a4-fbf32c96480f</portgroupId>
    <portgroupName>my-vdc-int-net2</portgroupName>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_4</label>
    <name>vnic4</name>
    <addressGroups>
	<addressGroup>
		<primaryAddress>13.13.1.1</primaryAddress>
		<subnetMask>255.255.255.0</subnetMask>
		<subnetPrefixLength>24</subnetPrefixLength>
	</addressGroup>
	</addressGroups>
	<mtu>1500</mtu>
	<type>internal</type>
	<isConnected>true</isConnected>
	<index>3</index>
	<portgroupId>96a68fd4-c21a-41d6-98a4-fbf32c96480f</portgroupId>
	<portgroupName>my-vdc-int-net2</portgroupName>
	<enableProxyArp>false</enableProxyArp>
	<enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_5</label>
    <name>vnic5</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>5</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_6</label>
    <name>vnic6</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>6</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_7</label>
    <name>vnic7</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>7</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_8</label>
    <name>vnic8</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>8</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_9</label>
    <name>vnic9</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>9</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
</vnics>`)
	vnicObject := &types.EdgeGatewayVnics{}
	_ = xml.Unmarshal(sampleBody, vnicObject)

	tests := []struct {
		name              string
		networkName       string
		networkType       string
		expectedVnicIndex *int
		hasError          bool
		expectedError     error
	}{
		{"ExtNetwork", "my-ext-network", types.EdgeGatewayVnicTypeUplink, takeAddress(0), false, nil},
		{"OrgNetwork", "my-vdc-int-net", types.EdgeGatewayVnicTypeInternal, takeAddress(1), false, nil},
		{"WithSubinterfaces", "with-subinterfaces", types.EdgeGatewayVnicTypeSubinterface, takeAddress(10), false, nil},
		{"WithSubinterfaces2", "subinterface2", types.EdgeGatewayVnicTypeSubinterface, takeAddress(11), false, nil},
		{"Trunk", "dvs.VCDVS-Trunk-Portgroup-vdcf9daf2da-b4f9-4921-a2f4-d77a943a381c", types.EdgeGatewayVnicTypeTrunk, takeAddress(2), false, nil},
		{"NonExistingUplink", "invalid-network-name", types.EdgeGatewayVnicTypeUplink, nil, true, ErrorEntityNotFound},
		{"NonExistingInternal", "invalid-network-name", types.EdgeGatewayVnicTypeInternal, nil, true, ErrorEntityNotFound},
		{"NonExistingSubinterface", "invalid-network-name", types.EdgeGatewayVnicTypeSubinterface, nil, true, ErrorEntityNotFound},
		{"NonExistingTrunk", "invalid-network-name", types.EdgeGatewayVnicTypeTrunk, nil, true, ErrorEntityNotFound},
		{"MoreThanOne", "my-vdc-int-net2", types.EdgeGatewayVnicTypeInternal, nil, true,
			fmt.Errorf("more than one (2) networks of type 'internal' with name 'my-vdc-int-net2' found")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnicIndex, err := GetVnicIndexByNetworkNameAndType(tt.networkName, tt.networkType, vnicObject)

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
	<vnics>
  <vnic>
    <label>vNic_0</label>
    <name>my-ext-network</name>
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
    <mtu>1500</mtu>
    <type>uplink</type>
    <isConnected>true</isConnected>
    <index>0</index>
    <portgroupId>f2547dd9-e0f7-4d81-97c1-dd33e5e0fbbf</portgroupId>
    <portgroupName>my-ext-network</portgroupName>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_1</label>
    <name>vnic1</name>
    <addressGroups>
      <addressGroup>
        <primaryAddress>10.10.10.5</primaryAddress>
        <subnetMask>255.255.255.0</subnetMask>
        <subnetPrefixLength>24</subnetPrefixLength>
      </addressGroup>
    </addressGroups>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>true</isConnected>
    <index>1</index>
    <portgroupId>95bffe8e-7e67-452d-abf2-535ac298db2b</portgroupId>
    <portgroupName>my-vdc-int-net</portgroupName>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_2</label>
    <name>vdcf9daf2da-b4f9-4921-a2f4-d77a943a381c</name>
    <addressGroups/>
    <mtu>1600</mtu>
    <type>trunk</type>
    <subInterfaces>
      <subInterface>
        <isConnected>true</isConnected>
        <label>vNic_10</label>
        <name>vnic1-subinterface</name>
        <index>10</index>
        <tunnelId>1</tunnelId>
        <logicalSwitchId>9a222ba9-23bc-41bf-b10f-0168f87b80ab</logicalSwitchId>
        <logicalSwitchName>with-subinterfaces</logicalSwitchName>
        <enableSendRedirects>true</enableSendRedirects>
        <mtu>1500</mtu>
        <addressGroups>
          <addressGroup>
            <primaryAddress>3.3.3.1</primaryAddress>
            <subnetMask>255.255.255.0</subnetMask>
            <subnetPrefixLength>24</subnetPrefixLength>
          </addressGroup>
        </addressGroups>
      </subInterface>
      <subInterface>
        <isConnected>true</isConnected>
        <label>vNic_11</label>
        <name>vnic2-subinterface</name>
        <index>11</index>
        <tunnelId>2</tunnelId>
        <logicalSwitchId>aecdf098-13ea-4ddf-9c4b-6a06372aa163</logicalSwitchId>
        <logicalSwitchName>subinterface2</logicalSwitchName>
        <enableSendRedirects>true</enableSendRedirects>
        <mtu>1500</mtu>
        <addressGroups>
          <addressGroup>
            <primaryAddress>7.7.7.1</primaryAddress>
            <subnetMask>255.255.255.0</subnetMask>
            <subnetPrefixLength>24</subnetPrefixLength>
          </addressGroup>
        </addressGroups>
      </subInterface>
      <subInterface>
        <isConnected>true</isConnected>
        <label>vNic_12</label>
        <name>vnic3-subinterface</name>
        <index>12</index>
        <tunnelId>3</tunnelId>
        <logicalSwitchId>c489d85d-df17-409c-810e-12af5aa5c1c2</logicalSwitchId>
        <logicalSwitchName>subinterface-subnet-clash</logicalSwitchName>
        <enableSendRedirects>true</enableSendRedirects>
        <mtu>1500</mtu>
        <addressGroups>
          <addressGroup>
            <primaryAddress>4.3.3.1</primaryAddress>
            <subnetMask>255.255.255.0</subnetMask>
            <subnetPrefixLength>24</subnetPrefixLength>
          </addressGroup>
        </addressGroups>
      </subInterface>
    </subInterfaces>
    <isConnected>true</isConnected>
    <index>2</index>
    <portgroupId>dvportgroup-20228</portgroupId>
    <portgroupName>dvs.VCDVS-Trunk-Portgroup-vdcf9daf2da-b4f9-4921-a2f4-d77a943a381c</portgroupName>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_3</label>
    <name>vnic3</name>
    <addressGroups>
      <addressGroup>
        <primaryAddress>13.13.1.1</primaryAddress>
        <subnetMask>255.255.255.0</subnetMask>
        <subnetPrefixLength>24</subnetPrefixLength>
      </addressGroup>
    </addressGroups>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>true</isConnected>
    <index>3</index>
    <portgroupId>96a68fd4-c21a-41d6-98a4-fbf32c96480f</portgroupId>
    <portgroupName>my-vdc-int-net2</portgroupName>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_4</label>
    <name>vnic4</name>
    <addressGroups>
	<addressGroup>
		<primaryAddress>13.13.1.1</primaryAddress>
		<subnetMask>255.255.255.0</subnetMask>
		<subnetPrefixLength>24</subnetPrefixLength>
	</addressGroup>
	</addressGroups>
	<mtu>1500</mtu>
	<type>internal</type>
	<isConnected>true</isConnected>
	<index>3</index>
	<portgroupId>96a68fd4-c21a-41d6-98a4-fbf32c96480f</portgroupId>
	<portgroupName>my-vdc-int-net2</portgroupName>
	<enableProxyArp>false</enableProxyArp>
	<enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_5</label>
    <name>vnic5</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>5</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_6</label>
    <name>vnic6</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>6</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_7</label>
    <name>vnic7</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>7</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
    <label>vNic_8</label>
    <name>vnic8</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>8</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
  <vnic>
  <label>vNic_8</label>
  <name>vnic8</name>
  <addressGroups/>
  <mtu>1500</mtu>
  <type>internal</type>
  <isConnected>false</isConnected>
  <index>8</index>
  <enableProxyArp>false</enableProxyArp>
  <enableSendRedirects>true</enableSendRedirects>
</vnic>
  <vnic>
    <label>vNic_9</label>
    <name>vnic9</name>
    <addressGroups/>
    <mtu>1500</mtu>
    <type>internal</type>
    <isConnected>false</isConnected>
    <index>9</index>
    <enableProxyArp>false</enableProxyArp>
    <enableSendRedirects>true</enableSendRedirects>
  </vnic>
</vnics>`)
	vnicObject := &types.EdgeGatewayVnics{}
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
		{"Trunk", 2, "dvs.VCDVS-Trunk-Portgroup-vdcf9daf2da-b4f9-4921-a2f4-d77a943a381c", "trunk", false, nil},
		{"Subinterface", 10, "with-subinterfaces", "subinterface", false, nil},
		{"NonExistent", 219, "", "", true, ErrorEntityNotFound},
		{"DuplicateRecords", 8, "", "", true, fmt.Errorf("more than one networks found for vNic 8")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkName, networkType, err := GetNetworkNameAndTypeByVnicIndex(tt.vnicIndex, vnicObject)

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

// takeAddress is a helper which can gives address of `int`
func takeAddress(x int) *int {
	return &x
}
