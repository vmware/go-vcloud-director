//go:build vdc || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VdcTemplateCRUD(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}

	// Pre-requisites: We need information such as Provider VDC, External networks (Provider Gateways)
	// and Edge Clusters.
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	edgeCluster, err := vdc.GetNsxtEdgeClusterByName(vcd.config.VCD.Nsxt.NsxtEdgeCluster)
	check.Assert(err, IsNil)
	check.Assert(edgeCluster, NotNil)

	providerVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)
	check.Assert(providerVdc, NotNil)
	check.Assert(providerVdc.ProviderVdc.AvailableNetworks, NotNil)
	check.Assert(providerVdc.ProviderVdc.NetworkPoolReferences, NotNil)

	var networkRef *types.Reference
	for _, netRef := range providerVdc.ProviderVdc.AvailableNetworks.Network {
		if netRef.Name == vcd.config.VCD.Nsxt.ExternalNetwork {
			networkRef = netRef
			break
		}
	}
	check.Assert(networkRef, NotNil)

	var networkPoolRef *types.Reference
	for _, netPoolRef := range providerVdc.ProviderVdc.NetworkPoolReferences.NetworkPoolReference {
		if netPoolRef.Name == vcd.config.VCD.NsxtProviderVdc.NetworkPool {
			networkPoolRef = netPoolRef
			break
		}
	}
	check.Assert(networkPoolRef, NotNil)

	externalNetworkBindingId := fmt.Sprintf("urn:vcloud:binding:%s", uuid.NewString())
	gatewayEdgeClusterBindingId := fmt.Sprintf("urn:vcloud:binding:%s", uuid.NewString())
	servicesEdgeClusterBindingId := fmt.Sprintf("urn:vcloud:binding:%s", uuid.NewString())

	settings := types.VMWVdcTemplate{
		NetworkBackingType: "NSX_T",
		ProviderVdcReference: []*types.VMWVdcTemplateProviderVdcSpecification{{
			HREF: providerVdc.ProviderVdc.HREF,
			Binding: []*types.VMWVdcTemplateBinding{
				{
					Name: gatewayEdgeClusterBindingId,
					Value: &types.Reference{
						ID:   fmt.Sprintf("urn:vcloud:backingEdgeCluster:%s", edgeCluster.NsxtEdgeCluster.ID),
						HREF: fmt.Sprintf("urn:vcloud:backingEdgeCluster:%s", edgeCluster.NsxtEdgeCluster.ID),
						Type: "application/json",
					},
				},
				{
					Name: servicesEdgeClusterBindingId,
					Value: &types.Reference{
						ID:   fmt.Sprintf("urn:vcloud:backingEdgeCluster:%s", edgeCluster.NsxtEdgeCluster.ID),
						HREF: fmt.Sprintf("urn:vcloud:backingEdgeCluster:%s", edgeCluster.NsxtEdgeCluster.ID),
						Type: "application/json",
					},
				},
				{
					Name: externalNetworkBindingId,
					Value: &types.Reference{
						ID:   networkRef.ID,
						HREF: networkRef.HREF,
						Type: networkRef.Type,
					},
				},
			},
		}},
		Name:              check.TestName(),
		Description:       check.TestName(),
		TenantName:        check.TestName() + "_Tenant",
		TenantDescription: check.TestName() + "_Tenant",

		VdcTemplateSpecification: &types.VMWVdcTemplateSpecification{
			Type:                    types.VdcTemplateFlexType,
			NicQuota:                100,
			VmQuota:                 100,
			ProvisionedNetworkQuota: 1000,
			GatewayConfiguration: &types.VdcTemplateSpecificationGatewayConfiguration{
				Gateway: &types.EdgeGateway{
					Name:        check.TestName(),
					Description: check.TestName(),
					Configuration: &types.GatewayConfiguration{
						GatewayInterfaces: &types.GatewayInterfaces{GatewayInterface: []*types.GatewayInterface{
							{
								Name:          gatewayEdgeClusterBindingId,
								DisplayName:   gatewayEdgeClusterBindingId,
								Connected:     true,
								InterfaceType: "UPLINK",
								Network: &types.Reference{
									HREF: gatewayEdgeClusterBindingId,
								},
							},
						}},
						EdgeClusterConfiguration: &types.EdgeClusterConfiguration{PrimaryEdgeCluster: &types.Reference{HREF: gatewayEdgeClusterBindingId}},
					},
				},
				Network: &types.OrgVDCNetwork{
					Name:        check.TestName() + "_Net",
					Description: check.TestName() + "_Net",
					Configuration: &types.NetworkConfiguration{
						IPScopes: &types.IPScopes{IPScope: []*types.IPScope{
							{
								IsInherited:        false,
								Gateway:            "1.1.1.1",
								Netmask:            "255.255.240.0",
								SubnetPrefixLength: addrOf(20),
								IPRanges: &types.IPRanges{IPRange: []*types.IPRange{
									{
										StartAddress: "1.1.1.1",
										EndAddress:   "1.1.1.1",
									},
								}},
							},
						}},
						FenceMode: "natRouted",
					},
					IsShared: false,
				},
			},
			StorageProfile: []*types.VdcStorageProfile{
				{
					Name:    "Development2",
					Enabled: addrOf(true),
					Units:   "MB",
					Limit:   1024,
					Default: true,
				},
			},
			IsElastic:               addrOf(false),
			IncludeMemoryOverhead:   addrOf(true),
			ThinProvision:           true,
			FastProvisioningEnabled: true,
			NetworkPoolReference:    networkPoolRef,
			NetworkProfileConfiguration: &types.VdcTemplateNetworkProfile{
				ServicesEdgeCluster: &types.Reference{HREF: servicesEdgeClusterBindingId},
			},
			CpuAllocationMhz:           256,
			CpuLimitMhzPerVcpu:         1000,
			CpuLimitMhz:                256,
			MemoryAllocationMB:         1024,
			MemoryLimitMb:              1024,
			CpuGuaranteedPercentage:    20,
			MemoryGuaranteedPercentage: 30,
		},
	}

	template, err := vcd.client.CreateVdcTemplate(settings)
	check.Assert(err, IsNil)
	check.Assert(template, NotNil)

	defer func() {
		err = template.Delete()
		check.Assert(err, IsNil)
	}()

	templateById, err := vcd.client.GetVdcTemplateById(template.VdcTemplate.ID)
	check.Assert(err, IsNil)
	check.Assert(templateById, NotNil)
	check.Assert(templateById, DeepEquals, template)

	_, err = vcd.client.GetVdcTemplateById("urn:vcloud:vdctemplate:00000000-0000-0000-00000-000000000000")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	templateByName, err := vcd.client.GetVdcTemplateByName(template.VdcTemplate.Name)
	check.Assert(err, IsNil)
	check.Assert(templateByName, NotNil)
	check.Assert(templateByName, DeepEquals, templateById)

	_, err = vcd.client.GetVdcTemplateByName("IDoNotExist")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	settings.Description = "Updated"
	settings.VdcTemplateSpecification.CpuLimitMhz = 500
	settings.VdcTemplateSpecification.NicQuota = 500
	template, err = template.Update(settings)
	check.Assert(err, IsNil)
	check.Assert(template, NotNil)
	check.Assert(template.VdcTemplate.Description, Equals, "Updated")
	check.Assert(template.VdcTemplate.VdcTemplateSpecification.CpuLimitMhz, Equals, 500)
	check.Assert(template.VdcTemplate.VdcTemplateSpecification.NicQuota, Equals, 500)

	access, err := template.GetAccess()
	check.Assert(err, IsNil)
	check.Assert(access, NotNil)
	check.Assert(access.AccessSettings, IsNil)

	err = template.SetAccess([]string{org.AdminOrg.ID})
	check.Assert(err, IsNil)

	access, err = template.GetAccess()
	check.Assert(err, IsNil)
	check.Assert(access, NotNil)
	check.Assert(access.AccessSettings, NotNil)
	check.Assert(len(access.AccessSettings.AccessSetting), Equals, 1)
	check.Assert(access.AccessSettings.AccessSetting[0].Subject, NotNil)
	check.Assert(access.AccessSettings.AccessSetting[0].Subject.HREF, Equals, org.AdminOrg.HREF)
}

func (vcd *TestVCD) Test_VdcTemplateInstantiate(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}

	// Pre-requisites: We need information such as Provider VDC and External networks (Provider Gateways)
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	providerVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)
	check.Assert(providerVdc, NotNil)
	check.Assert(providerVdc.ProviderVdc.AvailableNetworks, NotNil)
	check.Assert(providerVdc.ProviderVdc.NetworkPoolReferences, NotNil)

	var networkRef *types.Reference
	for _, netRef := range providerVdc.ProviderVdc.AvailableNetworks.Network {
		if netRef.Name == vcd.config.VCD.Nsxt.ExternalNetwork {
			networkRef = netRef
			break
		}
	}
	check.Assert(networkRef, NotNil)

	var networkPoolRef *types.Reference
	for _, netPoolRef := range providerVdc.ProviderVdc.NetworkPoolReferences.NetworkPoolReference {
		if netPoolRef.Name == vcd.config.VCD.NsxtProviderVdc.NetworkPool {
			networkPoolRef = netPoolRef
			break
		}
	}
	check.Assert(networkPoolRef, NotNil)

	externalNetworkBindingId := fmt.Sprintf("urn:vcloud:binding:%s", uuid.NewString())

	template, err := vcd.client.CreateVdcTemplate(types.VMWVdcTemplate{
		NetworkBackingType: "NSX_T",
		ProviderVdcReference: []*types.VMWVdcTemplateProviderVdcSpecification{{
			HREF: providerVdc.ProviderVdc.HREF,
			Binding: []*types.VMWVdcTemplateBinding{
				{
					Name: externalNetworkBindingId,
					Value: &types.Reference{
						ID:   networkRef.ID,
						HREF: networkRef.HREF,
						Type: networkRef.Type,
					},
				},
			},
		}},
		Name:              check.TestName(),
		Description:       check.TestName(),
		TenantName:        check.TestName() + "_Tenant",
		TenantDescription: check.TestName() + "_Tenant",

		VdcTemplateSpecification: &types.VMWVdcTemplateSpecification{
			Type:                    types.VdcTemplateFlexType,
			NicQuota:                100,
			VmQuota:                 100,
			ProvisionedNetworkQuota: 1000,
			StorageProfile: []*types.VdcStorageProfile{
				{
					Name:    "Development2",
					Enabled: addrOf(true),
					Units:   "MB",
					Limit:   1024,
					Default: true,
				},
			},
			IsElastic:                  addrOf(false),
			IncludeMemoryOverhead:      addrOf(true),
			ThinProvision:              true,
			FastProvisioningEnabled:    true,
			CpuAllocationMhz:           256,
			CpuLimitMhzPerVcpu:         1000,
			CpuLimitMhz:                256,
			MemoryAllocationMB:         1024,
			MemoryLimitMb:              1024,
			CpuGuaranteedPercentage:    20,
			MemoryGuaranteedPercentage: 30,
		},
	})
	check.Assert(err, IsNil)
	check.Assert(template, NotNil)

	defer func() {
		err = template.Delete()
		check.Assert(err, IsNil)
	}()

	err = template.SetAccess([]string{org.AdminOrg.ID})
	check.Assert(err, IsNil)

	// Instantiate the VDC Template as System administrator
	vdcId, err := template.Instantiate(check.TestName(), check.TestName(), org.AdminOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(vdcId, Not(Equals), "")

	vdc, err := org.GetVDCById(vdcId, true)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	check.Assert(vdc.Vdc.ID, Equals, vdcId)

	err = vdc.DeleteWait(true, true)
	check.Assert(err, IsNil)

	// Instantiate the VDC Template as a Tenant
	if len(vcd.config.Tenants) > 0 {
		orgName := vcd.config.Tenants[0].SysOrg
		userName := vcd.config.Tenants[0].User
		password := vcd.config.Tenants[0].Password

		vcdClient := NewVCDClient(vcd.client.Client.VCDHREF, true)
		err := vcdClient.Authenticate(userName, password, orgName)
		check.Assert(err, IsNil)

		templateAsTenant, err := vcdClient.GetVdcTemplateByName(template.VdcTemplate.TenantName) // Careful, we must use the Tenant name now
		check.Assert(err, IsNil)
		check.Assert(templateAsTenant, NotNil)

		vdcId, err := templateAsTenant.Instantiate(check.TestName(), check.TestName(), org.AdminOrg.ID)
		check.Assert(err, IsNil)
		check.Assert(vdcId, Not(Equals), "")

		// Check that it really exists and delete it afterward. We do this as the System admin, as we don't need
		// the tenant user anymore.
		vdc, err := org.GetVDCById(vdcId, true)
		check.Assert(err, IsNil)
		check.Assert(vdc, NotNil)
		check.Assert(vdc.Vdc.ID, Equals, vdcId)

		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)

	}
}