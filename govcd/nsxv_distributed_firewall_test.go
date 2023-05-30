//go:build functional || network || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"fmt"
	"strings"

	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxvDistributedFirewall(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	if vcd.config.VCD.Nsxt.Vdc != "" {
		if testVerbose {
			fmt.Println("Testing attempted access to NSX-T VDC")
		}
		// Retrieve a NSX-T VDC
		nsxtVdc, err := org.GetAdminVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
		check.Assert(err, IsNil)

		notWorkingDfw := NewNsxvDistributedFirewall(nsxtVdc.client, nsxtVdc.AdminVdc.ID)
		check.Assert(notWorkingDfw, NotNil)

		isEnabled, err := notWorkingDfw.IsEnabled()
		check.Assert(err, IsNil)
		check.Assert(isEnabled, Equals, false)

		err = notWorkingDfw.Enable()
		// NSX-T VDCs don't support NSX-V distributed firewalls. We expect an error here.
		check.Assert(err, NotNil)
		check.Assert(strings.Contains(err.Error(), "Forbidden"), Equals, true)
		if testVerbose {
			fmt.Printf("notWorkingDfw: %s\n", err)
		}
		config, err := notWorkingDfw.GetConfiguration()
		// Also this operation should fail
		check.Assert(err, NotNil)
		check.Assert(config, IsNil)
	}

	// Retrieve a NSX-V VDC
	vdc, err := org.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	dfw := NewNsxvDistributedFirewall(vdc.client, vdc.AdminVdc.ID)
	check.Assert(dfw, NotNil)

	// dfw.Enable is an idempotent operation. It can be repeated on an already enabled DFW
	// without errors.
	err = dfw.Enable()
	check.Assert(err, IsNil)

	enabled, err := dfw.IsEnabled()
	check.Assert(err, IsNil)
	check.Assert(enabled, Equals, true)

	config, err := dfw.GetConfiguration()
	check.Assert(err, IsNil)
	check.Assert(config, NotNil)
	if testVerbose {
		fmt.Printf("%# v\n", pretty.Formatter(config))
	}

	// Repeat enable operation
	err = dfw.Enable()
	check.Assert(err, IsNil)

	enabled, err = dfw.IsEnabled()
	check.Assert(err, IsNil)
	check.Assert(enabled, Equals, true)

	err = dfw.Disable()
	check.Assert(err, IsNil)
	enabled, err = dfw.IsEnabled()
	check.Assert(err, IsNil)
	check.Assert(enabled, Equals, false)

	// Also dfw.Disable is idempotent
	err = dfw.Disable()
	check.Assert(err, IsNil)

	enabled, err = dfw.IsEnabled()
	check.Assert(err, IsNil)
	check.Assert(enabled, Equals, false)
}

func (vcd *TestVCD) Test_NsxvDistributedFirewallUpdate(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	// Retrieve a NSX-V VDC
	adminVdc, err := org.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)

	dfw := NewNsxvDistributedFirewall(adminVdc.client, adminVdc.AdminVdc.ID)
	check.Assert(dfw, NotNil)
	enabled, err := dfw.IsEnabled()
	check.Assert(err, IsNil)
	//
	if enabled {
		check.Skip(fmt.Sprintf("VDC %s already contains a distributed firewall - skipping", vcd.config.VCD.Vdc))
	}

	vms, err := vdc.QueryVmList(types.VmQueryFilterOnlyDeployed)
	check.Assert(err, IsNil)

	sampleDestination := &types.Destinations{}
	if len(vms) > 0 {
		sampleDestination.Destination = []types.Destination{
			{
				Name:    vms[0].Name,
				Value:   extractUuid(vms[0].HREF),
				Type:    types.DFWElementVirtualMachine,
				IsValid: true,
			},
		}
	}
	err = dfw.Enable()
	check.Assert(err, IsNil)

	dnsService, err := dfw.GetServiceByName("DNS")
	check.Assert(err, IsNil)
	integrationServiceGroup, err := dfw.GetServiceGroupByName("MSSQL Integration Services")
	check.Assert(err, IsNil)

	network, err := vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	AddToCleanupList(vcd.config.VCD.Vdc, "nsxv_dfw", vcd.config.VCD.Org, check.TestName())
	rules := []types.NsxvDistributedFirewallRule{
		{
			Name:   "first",
			Action: types.DFWActionDeny,
			AppliedToList: &types.AppliedToList{
				AppliedTo: []types.AppliedTo{
					{
						Name:    adminVdc.AdminVdc.Name,
						Value:   adminVdc.AdminVdc.ID,
						Type:    "VDC",
						IsValid: true,
					},
				},
			},
			Direction:  types.DFWDirectionInout,
			PacketType: types.DFWPacketAny,
		},
		{
			Name:          "second",
			AppliedToList: &types.AppliedToList{},
			SectionID:     nil,
			Sources:       nil,
			Destinations:  nil,
			Services:      nil,
			Direction:     types.DFWDirectionIn,
			PacketType:    types.DFWPacketAny,
			Action:        types.DFWActionAllow,
		},
		{
			Name:          "third",
			Action:        types.DFWActionAllow,
			AppliedToList: &types.AppliedToList{},
			Sources: &types.Sources{
				Source: []types.Source{
					// Anonymous source
					{
						Name:  "10.10.10.1",
						Value: "10.10.10.1",
						Type:  types.DFWElementIpv4,
					},
					// Named source
					{
						Name:    network.OrgVDCNetwork.Name,
						Value:   extractUuid(network.OrgVDCNetwork.ID),
						Type:    types.DFWElementNetwork,
						IsValid: true,
					},
				},
			},
			Destinations: sampleDestination,
			Services: &types.Services{
				Service: []types.Service{
					// Anonymous service
					{
						IsValid:         true,
						SourcePort:      addrOf("1000"),
						DestinationPort: addrOf("1200"),
						Protocol:        addrOf(types.NsxvProtocolCodes[types.DFWProtocolTcp]),
						ProtocolName:    addrOf(types.DFWProtocolTcp),
					},
					// Named service
					{
						IsValid: true,
						Name:    dnsService.Name,
						Value:   dnsService.ObjectID,
						Type:    types.DFWServiceTypeApplication,
					},
					// Named service group
					{
						IsValid: true,
						Name:    integrationServiceGroup.Name,
						Value:   integrationServiceGroup.ObjectID,
						Type:    types.DFWServiceTypeApplicationGroup,
					},
				},
			},
			Direction:  types.DFWDirectionIn,
			PacketType: types.DFWPacketIpv4,
		},
	}

	updatedRules, err := dfw.UpdateConfiguration(rules)
	check.Assert(err, IsNil)
	check.Assert(updatedRules, NotNil)

	err = dfw.Disable()
	check.Assert(err, IsNil)

}

func (vcd *TestVCD) Test_NsxvServices(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	// Retrieve a NSX-V VDC
	vdc, err := org.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	dfw := NewNsxvDistributedFirewall(vdc.client, vdc.AdminVdc.ID)
	check.Assert(dfw, NotNil)

	services, err := dfw.GetServices(false)
	check.Assert(err, IsNil)
	check.Assert(services, NotNil)
	check.Assert(len(services) > 0, Equals, true)

	if testVerbose {
		fmt.Printf("services: %d\n", len(services))
		fmt.Printf("%# v\n", pretty.Formatter(services[0]))
	}

	serviceName := "PostgreSQL"
	serviceByName, err := dfw.GetServiceByName(serviceName)

	check.Assert(err, IsNil)
	check.Assert(serviceByName, NotNil)
	check.Assert(serviceByName.Name, Equals, serviceName)

	serviceById, err := dfw.GetServiceById(serviceByName.ObjectID)
	check.Assert(err, IsNil)
	check.Assert(serviceById.Name, Equals, serviceName)

	searchRegex := "M.SQL" // Finds, among others, names containing "MySQL" or "MSSQL"
	servicesByRegex, err := dfw.GetServicesByRegex(searchRegex)
	check.Assert(err, IsNil)
	check.Assert(len(servicesByRegex) > 1, Equals, true)

	searchRegex = "." // Finds all services
	servicesByRegex, err = dfw.GetServicesByRegex(searchRegex)
	check.Assert(err, IsNil)
	check.Assert(len(servicesByRegex), Equals, len(services))

	searchRegex = "^####--no-such-service--####$" // Finds no services
	servicesByRegex, err = dfw.GetServicesByRegex(searchRegex)
	check.Assert(err, IsNil)
	check.Assert(len(servicesByRegex), Equals, 0)

	serviceGroups, err := dfw.GetServiceGroups(false)
	check.Assert(err, IsNil)
	check.Assert(serviceGroups, NotNil)
	check.Assert(len(serviceGroups) > 0, Equals, true)
	if testVerbose {
		fmt.Printf("service groups: %d\n", len(serviceGroups))
		fmt.Printf("%# v\n", pretty.Formatter(serviceGroups[0]))
	}
	serviceGroupName := "Orchestrator"
	serviceGroupByName, err := dfw.GetServiceGroupByName(serviceGroupName)
	check.Assert(err, IsNil)
	check.Assert(serviceGroupByName, NotNil)
	check.Assert(serviceGroupByName.Name, Equals, serviceGroupName)

	serviceGroupById, err := dfw.GetServiceGroupById(serviceGroupByName.ObjectID)
	check.Assert(err, IsNil)
	check.Assert(serviceGroupById, NotNil)
	check.Assert(serviceGroupById.Name, Equals, serviceGroupName)

	searchRegex = "Oracle"
	serviceGroupsByRegex, err := dfw.GetServiceGroupsByRegex(searchRegex)
	check.Assert(err, IsNil)
	check.Assert(len(serviceGroupsByRegex) > 1, Equals, true)

	searchRegex = "."
	serviceGroupsByRegex, err = dfw.GetServiceGroupsByRegex(searchRegex)
	check.Assert(err, IsNil)
	check.Assert(len(serviceGroupsByRegex), Equals, len(serviceGroups))

	searchRegex = "^####--no-such-service-group--####$"
	serviceGroupsByRegex, err = dfw.GetServiceGroupsByRegex(searchRegex)
	check.Assert(err, IsNil)
	check.Assert(len(serviceGroupsByRegex), Equals, 0)
}
