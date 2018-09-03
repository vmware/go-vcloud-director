/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_FindVDCNetwork(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())

	net, err := vcd.vdc.FindVDCNetwork(vcd.config.VCD.Network)

	check.Assert(err, IsNil)
	check.Assert(net, NotNil)
	check.Assert(net.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network)
	check.Assert(net.OrgVDCNetwork.HREF, Not(Equals), "")

	// find Invalid Network
	net, err = vcd.vdc.FindVDCNetwork("INVALID")
	check.Assert(err, NotNil)
}

func (vcd *TestVCD) Test_NewVdc(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	err := vcd.vdc.Refresh()
	check.Assert(err, IsNil)

	check.Assert(vcd.vdc.Vdc.Link[0].Rel, Equals, "up")
	check.Assert(vcd.vdc.Vdc.Link[0].Type, Equals, "application/vnd.vmware.vcloud.org+xml")

	// fmt.Printf("allocation mem %#v\n\n",vcd.vdc.Vdc.AllocationModel)
	for _, resource := range vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity {

		// fmt.Printf("res %#v\n",resource)
		check.Assert(resource.Name, Not(Equals), "")
		check.Assert(resource.Type, Not(Equals), "")
		check.Assert(resource.HREF, Not(Equals), "")
	}

	// TODO: find which values are acceptable for AllocationModel
	// check.Assert(vcd.vdc.Vdc.AllocationModel, Equals, "AllocationPool")

	/*
		// TODO: Find the conditions that define valid ComputeCapacity
		for _, v := range vcd.vdc.Vdc.ComputeCapacity {
			check.Assert(v.CPU.Units, Equals, "MHz")
			check.Assert(v.CPU.Allocated, Equals, int64(30000))
			check.Assert(v.CPU.Limit, Equals, int64(30000))
			check.Assert(v.CPU.Reserved, Equals, int64(15000))
			check.Assert(v.CPU.Used, Equals, int64(0))
			check.Assert(v.CPU.Overhead, Equals, int64(0))
			check.Assert(v.Memory.Units, Equals, "MB")
			check.Assert(v.Memory.Allocated, Equals, int64(61440))
			check.Assert(v.Memory.Limit, Equals, int64(61440))
			check.Assert(v.Memory.Reserved, Equals, int64(61440))
			check.Assert(v.Memory.Used, Equals, int64(6144))
			check.Assert(v.Memory.Overhead, Equals, int64(95))
		}
	*/

	// Skipping this check, as we can't define the existence of a given vApp template beforehand
	/*
		check.Assert(vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity[0].Name, Equals, "vAppTemplate")
		check.Assert(vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity[0].Type, Equals, "application/vnd.vmware.vcloud.vAppTemplate+xml")
		check.Assert(vcd.vdc.Vdc.ResourceEntities[0].ResourceEntity[0].HREF, Equals, "http://localhost:4444/api/vAppTemplate/vappTemplate-22222222-2222-2222-2222-222222222222")
	*/

	for _, v := range vcd.vdc.Vdc.AvailableNetworks {
		for _, v2 := range v.Network {
			check.Assert(v2.Name, Not(Equals), "")
			check.Assert(v2.Type, Equals, "application/vnd.vmware.vcloud.network+xml")
			check.Assert(v2.HREF, Not(Equals), "")
		}
	}

	/*

		// Skipping this check, as we don't have precise terms of comparison for this entity
		check.Assert(vcd.vdc.Vdc.NicQuota, Equals, 0)
		check.Assert(vcd.vdc.Vdc.NetworkQuota, Equals, 20)
		check.Assert(vcd.vdc.Vdc.UsedNetworkCount, Equals, 0)
		check.Assert(vcd.vdc.Vdc.VMQuota, Equals, 0)
		check.Assert(vcd.vdc.Vdc.IsEnabled, Equals, true)
	*/

	for _, v := range vcd.vdc.Vdc.VdcStorageProfiles {
		for _, v2 := range v.VdcStorageProfile {
			check.Assert(v2.Type, Equals, "application/vnd.vmware.vcloud.vdcStorageProfile+xml")
			check.Assert(v2.HREF, Not(Equals), "")
		}
	}

}

// Tests ComposeVApp with given parameters in the config file.
// Throws an error if networks, catalog, catalog item, and
// storage preference are omitted from the config file.
func (vcd *TestVCD) Test_ComposeVApp(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}
	fmt.Printf("Running: %s\n", check.TestName())

	// Populate OrgVDCNetwork
	networks := []*types.OrgVDCNetwork{}
	net, err := vcd.vdc.FindVDCNetwork(vcd.config.VCD.Network)
	networks = append(networks, net.OrgVDCNetwork)
	check.Assert(err, IsNil)
	// Populate Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)
	// Populate Catalog Item
	catitem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.Catalogitem)
	check.Assert(err, IsNil)
	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	check.Assert(err, IsNil)
	// Get StorageProfileReference
	storageprofileref, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)
	// Compose VApp
	temp_vapp_name := "go-vcloud-director-vapp-check"
	temp_vapp_description := "vapp created by tests"
	task, err := vcd.vdc.ComposeVApp(networks, vapptemplate, storageprofileref, temp_vapp_name, temp_vapp_description, true)
	check.Assert(err, IsNil)
	check.Assert(task.Task.OperationName, Equals, "vdcComposeVapp")
	// Get VApp
	vapp, err := vcd.vdc.FindVAppByName(temp_vapp_name)
	check.Assert(err, IsNil)
	// Once the operation is successful, we won't trigger a failure
	// until after the vApp deletion
	check.Check(vapp.VApp.Name, Equals, temp_vapp_name)
	check.Check(vapp.VApp.Description, Equals, temp_vapp_description)

	vapp_status, err := vapp.GetStatus()
	check.Check(err, IsNil)
	check.Check(vapp_status, Equals, "UNRESOLVED")
	// Let the VApp creation complete
	task.WaitTaskCompletion()
	vapp_status, err = vapp.GetStatus()
	check.Check(err, IsNil)
	check.Check(vapp_status, Equals, "POWERED_OFF")
	// Deleting VApp
	task, err = vapp.Delete()
	task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	no_such_vapp, err := vcd.vdc.FindVAppByName(temp_vapp_name)
	check.Assert(err, NotNil)
	check.Assert(no_such_vapp.VApp, IsNil)

}

func (vcd *TestVCD) Test_FindVApp(check *C) {

	if vcd.vapp.VApp == nil {
		check.Skip("No Vapp provided")
	}
	first_vapp, err := vcd.vdc.FindVAppByName(vcd.vapp.VApp.Name)

	check.Assert(err, IsNil)

	second_vapp, err := vcd.vdc.FindVAppByID(first_vapp.VApp.ID)

	check.Assert(err, IsNil)

	check.Assert(second_vapp.VApp.Name, Equals, first_vapp.VApp.Name)
	check.Assert(second_vapp.VApp.HREF, Equals, first_vapp.VApp.HREF)

}

var vdcExample = `
	<?xml version="1.0" ?>
	<Vdc href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000000" id="urn:vcloud:vdc:00000000-0000-0000-0000-000000000000" name="M916272752-5793" status="1" type="application/vnd.vmware.vcloud.vdc+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-in stance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
	  <Link href="http://localhost:4444/api/org/11111111-1111-1111-1111-111111111111" rel="up" type="application/vnd.vmware.vcloud.org+xml"/>
	  <Link href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000000/edgeGateways" rel="edgeGateways" type="application/vnd.vmware.vcloud.query.records+xml"/>
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
      <ResourceEntity href="http://localhost:4444/api/vApp/vapp-00000000-0000-0000-0000-000000000000" name="myVApp" type="application/vnd.vmware.vcloud.vApp+xml"/>
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
	    <VdcStorageProfile href="http://localhost:4444/api/vdcStorageProfile/88888888-8888-8888-8888-888888888888" name="storageProfile1" type="application/vnd.vmware.vcloud.vdcStorageProfile+xml"/>
	    <VdcStorageProfile href="http://localhost:4444/api/vdcStorageProfile/88888888-8888-8888-8888-888888888889" name="storageProfile2" type="application/vnd.vmware.vcloud.vdcStorageProfile+xml"/>
	  </VdcStorageProfiles>
	</Vdc>
	`
