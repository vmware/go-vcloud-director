/*
* @Author: frapposelli
* @Date:   2014-10-20 15:19:23
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:10:50
 */

package govcloudair

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestVDC(t *testing.T) {
	TestingT(t)
}

func (s *S) Test_RetrieveVDC(c *C) {
	testServer.Response(200, nil, vdcExample)

	vdc, err := s.client.RetrieveVDC()

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	for _, v := range vdc.Link {
		c.Assert(v.Rel, Equals, "up")
		c.Assert(v.Type, Equals, "application/vnd.vmware.vcloud.org+xml")
		c.Assert(v.HREF, Equals, "http://localhost:4444/api/org/11111111-1111-1111-1111-111111111111")
	}

	c.Assert(vdc.AllocationModel, Equals, "AllocationPool")

	for _, v := range vdc.ComputeCapacity {
		c.Assert(v.Cpu.Units, Equals, "MHz")
		c.Assert(v.Cpu.Allocated, Equals, 30000)
		c.Assert(v.Cpu.Limit, Equals, 30000)
		c.Assert(v.Cpu.Reserved, Equals, 15000)
		c.Assert(v.Cpu.Used, Equals, 0)
		c.Assert(v.Cpu.Overhead, Equals, 0)
		c.Assert(v.Memory.Units, Equals, "MB")
		c.Assert(v.Memory.Allocated, Equals, 61440)
		c.Assert(v.Memory.Limit, Equals, 61440)
		c.Assert(v.Memory.Reserved, Equals, 61440)
		c.Assert(v.Memory.Used, Equals, 6144)
		c.Assert(v.Memory.Overhead, Equals, 95)
	}

	for _, v := range vdc.ResourceEntities {
		for _, v2 := range v.ResourceEntity {
			c.Assert(v2.Name, Equals, "vAppTemplate")
			c.Assert(v2.Type, Equals, "application/vnd.vmware.vcloud.vAppTemplate+xml")
			c.Assert(v2.HREF, Equals, "http://localhost:4444/api/vAppTemplate/vappTemplate-22222222-2222-2222-2222-222222222222")
		}
	}

	for _, v := range vdc.AvailableNetworks {
		for _, v2 := range v.Network {
			c.Assert(v2.Name, Equals, "networkName")
			c.Assert(v2.Type, Equals, "application/vnd.vmware.vcloud.network+xml")
			c.Assert(v2.HREF, Equals, "http://localhost:4444/api/network/44444444-4444-4444-4444-4444444444444")
		}
	}

	for _, v := range vdc.Capabilities {
		for _, v2 := range v.SupportedHardwareVersions {
			c.Assert(v2, Equals, "vmx-10")
		}
	}

	c.Assert(vdc.NicQuota, Equals, 0)
	c.Assert(vdc.NetworkQuota, Equals, 20)
	c.Assert(vdc.UsedNetworkCount, Equals, 0)
	c.Assert(vdc.VmQuota, Equals, 0)
	c.Assert(vdc.IsEnabled, Equals, true)

	for _, v := range vdc.VdcStorageProfiles {
		for _, v2 := range v.VdcStorageProfile {
			c.Assert(v2.Name, Equals, "storageProfile")
			c.Assert(v2.Type, Equals, "application/vnd.vmware.vcloud.vdcStorageProfile+xml")
			c.Assert(v2.HREF, Equals, "http://localhost:4444/api/vdcStorageProfile/88888888-8888-8888-8888-888888888888")
		}
	}

}

func (s *S) Test_FindVDCNetworkId(c *C) {
	testServer.Response(200, nil, vdcExample)

	vdc, _ := s.client.RetrieveVDC()
	net := vdc.FindVDCNetworkId("networkName")

	_ = testServer.WaitRequest()

	c.Assert(net, Equals, "44444444-4444-4444-4444-4444444444444")

}

func (s *S) Test_FindVDCStorageProfileId(c *C) {
	testServer.Response(200, nil, vdcExample)

	vdc, _ := s.client.RetrieveVDC()
	net := vdc.FindVDCStorageProfileId("storageProfile")

	_ = testServer.WaitRequest()

	c.Assert(net, Equals, "88888888-8888-8888-8888-888888888888")

}

func (s *S) Test_FindVDCOrgId(c *C) {
	testServer.Response(200, nil, vdcExample)

	vdc, _ := s.client.RetrieveVDC()
	net := vdc.FindVDCOrgId()

	_ = testServer.WaitRequest()

	c.Assert(net, Equals, "11111111-1111-1111-1111-111111111111")

}

var vdcExample = `
	<?xml version="1.0" ?>
	<Vdc href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000000" id="urn:vcloud:vdc:00000000-0000-0000-0000-000000000000" name="M916272752-5793" status="1" type="application/vnd.vmware.vcloud.vdc+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-in stance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
	  <Link href="http://localhost:4444/api/org/11111111-1111-1111-1111-111111111111" rel="up" type="application/vnd.vmware.vcloud.org+xml"/>
	  <AllocationModel>AllocationPool</AllocationModel>
	  <ComputeCapacity>
	    <Cpu>
	      <Units>MHz</Units>
	      <Allocated>30000</Allocated>
	      <Limit>30000</Limit>
	      <Reserved>15000</Reserved>
	      <Used>0</Used>
	      <Overhead>0</Overhead>
	    </Cpu>
	    <Memory>
	      <Units>MB</Units>
	      <Allocated>61440</Allocated>
	      <Limit>61440</Limit>
	      <Reserved>61440</Reserved>
	      <Used>6144</Used>
	      <Overhead>95</Overhead>
	    </Memory>
	  </ComputeCapacity>
	  <ResourceEntities>
	    <ResourceEntity href="http://localhost:4444/api/vAppTemplate/vappTemplate-22222222-2222-2222-2222-222222222222" name="vAppTemplate" type="application/vnd.vmware.vcloud.vAppTemplate+xml"/>
	  </ResourceEntities>
	  <AvailableNetworks>
	    <Network href="http://localhost:4444/api/network/44444444-4444-4444-4444-4444444444444" name="networkName" type="application/vnd.vmware.vcloud.network+xml"/>
	  </AvailableNetworks>
	  <Capabilities>
	    <SupportedHardwareVersions>
	      <SupportedHardwareVersion>vmx-10</SupportedHardwareVersion>
	    </SupportedHardwareVersions>
	  </Capabilities>
	  <NicQuota>0</NicQuota>
	  <NetworkQuota>20</NetworkQuota>
	  <UsedNetworkCount>0</UsedNetworkCount>
	  <VmQuota>0</VmQuota>
	  <IsEnabled>true</IsEnabled>
	  <VdcStorageProfiles>
	    <VdcStorageProfile href="http://localhost:4444/api/vdcStorageProfile/88888888-8888-8888-8888-888888888888" name="storageProfile" type="application/vnd.vmware.vcloud.vdcStorageProfile+xml"/>
	  </VdcStorageProfiles>
	</Vdc>
	`
