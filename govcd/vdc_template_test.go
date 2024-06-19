//go:build vdc || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VdcTemplateCRUD(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}

	// Pre-requisites: We need information such as Provider VDC, External networks (Provider Gateways)
	// and Edge Clusters.
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	vdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
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

	// Bindings must be random UUIDs generated manually
	externalNetworkBindingId := "urn:vcloud:binding:b871f699-50e6-4a65-ab8b-428324735ff2"
	gatewayEdgeClusterBindingId := "urn:vcloud:binding:36038940-dbbf-4346-94d9-e99e54c8e43a"
	servicesEdgeClusterBindingId := "urn:vcloud:binding:8e7e2480-ba77-4dd0-a7d4-2d1155e4d087"

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
								IsInherited:           false,
								Gateway:               "1.1.1.1",
								Netmask:               "255.255.240.0",
								SubnetPrefixLengthInt: addrOf(20),
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
					Name:    vcd.config.VCD.NsxtProviderVdc.StorageProfile2,
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

	org, err := vcd.client.GetOrgByName(adminOrg.AdminOrg.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	adminTemplates, err := vcd.client.QueryAdminVdcTemplates()
	check.Assert(err, IsNil)
	check.Assert(true, Equals, len(adminTemplates) > 0)
	found := false
	for _, adminTemplate := range adminTemplates {
		if extractUuid(adminTemplate.HREF) == extractUuid(templateByName.VdcTemplate.HREF) {
			found = true
			break
		}
	}
	check.Assert(found, Equals, true)

	settings.Description = "Updated"
	settings.VdcTemplateSpecification.CpuLimitMhz = 500
	settings.VdcTemplateSpecification.NicQuota = 500
	template, err = template.Update(settings)
	check.Assert(err, IsNil)
	check.Assert(template, NotNil)
	check.Assert(template.VdcTemplate.Description, Equals, "Updated")
	check.Assert(template.VdcTemplate.VdcTemplateSpecification.CpuLimitMhz, Equals, 500)
	check.Assert(template.VdcTemplate.VdcTemplateSpecification.NicQuota, Equals, 500)

	access, err := template.GetAccessControl()
	check.Assert(err, IsNil)
	check.Assert(access, NotNil)
	check.Assert(access.AccessSettings, IsNil)

	err = template.SetAccessControl([]string{adminOrg.AdminOrg.ID})
	check.Assert(err, IsNil)

	access, err = template.GetAccessControl()
	check.Assert(err, IsNil)
	check.Assert(access, NotNil)
	check.Assert(access.AccessSettings, NotNil)
	check.Assert(len(access.AccessSettings.AccessSetting), Equals, 1)
	check.Assert(access.AccessSettings.AccessSetting[0].Subject, NotNil)
	check.Assert(access.AccessSettings.AccessSetting[0].Subject.HREF, Equals, adminOrg.AdminOrg.HREF)

	// Now that the tenant has permissions, the query should return it
	templates, err := org.QueryVdcTemplates()
	check.Assert(err, IsNil)
	check.Assert(true, Equals, len(adminTemplates) > 0)
	found = false
	for _, t := range templates {
		if extractUuid(t.HREF) == extractUuid(templateByName.VdcTemplate.HREF) {
			found = true
			break
		}
	}
	check.Assert(found, Equals, true)
}

func (vcd *TestVCD) Test_VdcTemplateInstantiate(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip("test requires system administrator privileges")
	}

	// Pre-requisites: We need information such as Provider VDC and External networks (Provider Gateways)
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

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

	// Bindings must be random UUIDs generated manually
	externalNetworkBindingId := "urn:vcloud:binding:10d82a8a-f0a9-4c98-a462-d6a1b65ae210"

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
					Name:    vcd.config.VCD.NsxtProviderVdc.StorageProfile2,
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

	err = template.SetAccessControl([]string{adminOrg.AdminOrg.ID})
	check.Assert(err, IsNil)

	// Instantiate the VDC Template as System administrator
	vdc, err := template.InstantiateVdc(check.TestName(), check.TestName(), adminOrg.AdminOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	defer func() {
		// Delete the instantiated VDC even on test errors
		err = vdc.DeleteWait(true, true)
		check.Assert(err, IsNil)
	}()
	check.Assert(vdc.Vdc.Name, Equals, check.TestName())
	check.Assert(vdc.Vdc.Description, Equals, check.TestName())

	org, err := vdc.getParentOrg()
	check.Assert(err, IsNil)
	check.Assert(adminOrg.AdminOrg.ID, Equals, org.orgId())

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

		_, err = vcdClient.QueryAdminVdcTemplates()
		check.Assert(err, NotNil)

		org, err := vcdClient.GetOrgByName(orgName)
		check.Assert(err, IsNil)
		check.Assert(org, NotNil)

		tenantTemplates, err := org.QueryVdcTemplates()
		check.Assert(err, IsNil)
		check.Assert(true, Equals, len(tenantTemplates) > 0)
		found := false
		for _, tenantTemplate := range tenantTemplates {
			if extractUuid(tenantTemplate.HREF) == extractUuid(templateAsTenant.VdcTemplate.HREF) {
				found = true
				break
			}
		}
		check.Assert(found, Equals, true)

		vdc2, err := templateAsTenant.InstantiateVdc(check.TestName()+"2", check.TestName()+"2", adminOrg.AdminOrg.ID)
		check.Assert(err, IsNil)
		check.Assert(vdc2, NotNil)
		defer func() {
			// Also delete the second instantiated VDC even on test errors. We need to retrieve it as Admin for that.
			adminVdc2, err := adminOrg.GetVDCById(vdc2.Vdc.ID, true)
			check.Assert(err, IsNil)
			err = adminVdc2.DeleteWait(true, true)
			check.Assert(err, IsNil)
		}()
		check.Assert(vdc2.Vdc.Name, Equals, check.TestName()+"2")
		check.Assert(vdc2.Vdc.Description, Equals, check.TestName()+"2")

		vdcOrg, err := vdc.getParentOrg()
		check.Assert(err, IsNil)
		check.Assert(adminOrg.AdminOrg.ID, Equals, vdcOrg.orgId())
	}
}
