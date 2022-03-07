//go:build network || functional || openapi || vdcGroup || nsxt || gateway || ALL
// +build network functional openapi vdcGroup nsxt gateway ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func test_CreateVdcGroup(check *C, adminOrg *AdminOrg, vcd *TestVCD) (*Vdc, *VdcGroup) {
	createdVdc := createNewVdc(vcd, check, check.TestName())

	createdVdcAsCandidate, err := adminOrg.GetAllNsxtVdcGroupCandidates(createdVdc.vdcId(),
		map[string][]string{"filter": []string{fmt.Sprintf("name==%s", url.QueryEscape(createdVdc.vdcName()))}})
	check.Assert(err, IsNil)
	check.Assert(createdVdcAsCandidate, NotNil)
	check.Assert(len(createdVdcAsCandidate) == 1, Equals, true)

	existingVdcAsCandidate, err := adminOrg.GetAllNsxtVdcGroupCandidates(createdVdc.vdcId(),
		map[string][]string{"filter": []string{fmt.Sprintf("name==%s", url.QueryEscape(vcd.nsxtVdc.vdcName()))}})
	check.Assert(err, IsNil)
	check.Assert(existingVdcAsCandidate, NotNil)
	check.Assert(len(existingVdcAsCandidate) == 1, Equals, true)

	vdcGroupConfig := &types.VdcGroup{
		Name:  check.TestName() + "Group",
		OrgId: adminOrg.orgId(),
		ParticipatingOrgVdcs: []types.ParticipatingOrgVdcs{
			types.ParticipatingOrgVdcs{
				VdcRef: types.OpenApiReference{
					ID: createdVdc.vdcId(),
				},
				SiteRef: (createdVdcAsCandidate)[0].SiteRef,
				OrgRef:  (createdVdcAsCandidate)[0].OrgRef,
			},
			types.ParticipatingOrgVdcs{
				VdcRef: types.OpenApiReference{
					ID: vcd.nsxtVdc.vdcId(),
				},
				SiteRef: (existingVdcAsCandidate)[0].SiteRef,
				OrgRef:  (existingVdcAsCandidate)[0].OrgRef,
			},
		},
		LocalEgress:                false,
		UniversalNetworkingEnabled: false,
		NetworkProviderType:        "NSX_T",
		Type:                       "LOCAL",
		//DfwEnabled: true, // ignored by API
	}

	vdcGroup, err := adminOrg.CreateVdcGroup(vdcGroupConfig)
	check.Assert(err, IsNil)
	check.Assert(vdcGroup, NotNil)
	check.Assert(vdcGroup.IsNsxt(), Equals, true)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups + vdcGroup.VdcGroup.Id
	PrependToCleanupListOpenApi(vdcGroup.VdcGroup.Name, check.TestName(), openApiEndpoint)

	return createdVdc, vdcGroup
}

func createNewVdc(vcd *TestVCD, check *C, vdcName string) *Vdc {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	pVdcs, err := QueryProviderVdcByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)

	if len(pVdcs) == 0 {
		check.Skip(fmt.Sprintf("No NSX-T Provider VDC found with name '%s'", vcd.config.VCD.NsxtProviderVdc.Name))
	}
	providerVdcHref := pVdcs[0].HREF

	pvdcStorageProfile, err := vcd.client.QueryProviderVdcStorageProfileByName(vcd.config.VCD.NsxtProviderVdc.StorageProfile, providerVdcHref)

	check.Assert(err, IsNil)
	providerVdcStorageProfileHref := pvdcStorageProfile.HREF

	networkPools, err := QueryNetworkPoolByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.NetworkPool)
	check.Assert(err, IsNil)
	if len(networkPools) == 0 {
		check.Skip(fmt.Sprintf("No network pool found with name '%s'", vcd.config.VCD.NsxtProviderVdc.NetworkPool))
	}

	networkPoolHref := networkPools[0].HREF
	trueValue := true
	vdcConfiguration := &types.VdcConfiguration{
		Name:            vdcName,
		Xmlns:           types.XMLNamespaceVCloud,
		AllocationModel: "Flex",
		ComputeCapacity: []*types.ComputeCapacity{
			&types.ComputeCapacity{
				CPU: &types.CapacityWithUsage{
					Units:     "MHz",
					Allocated: 1024,
					Limit:     1024,
				},
				Memory: &types.CapacityWithUsage{
					Allocated: 1024,
					Limit:     1024,
					Units:     "MB",
				},
			},
		},
		VdcStorageProfile: []*types.VdcStorageProfileConfiguration{&types.VdcStorageProfileConfiguration{
			Enabled: takeBoolPointer(true),
			Units:   "MB",
			Limit:   1024,
			Default: true,
			ProviderVdcStorageProfile: &types.Reference{
				HREF: providerVdcStorageProfileHref,
			},
		},
		},
		NetworkPoolReference: &types.Reference{
			HREF: networkPoolHref,
		},
		ProviderVdcReference: &types.Reference{
			HREF: providerVdcHref,
		},
		IsEnabled:             true,
		IsThinProvision:       true,
		UsesFastProvisioning:  true,
		IsElastic:             &trueValue,
		IncludeMemoryOverhead: &trueValue,
	}

	vdc, err := adminOrg.CreateOrgVdc(vdcConfiguration)
	check.Assert(vdc, NotNil)
	check.Assert(err, IsNil)

	AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, check.TestName())
	return vdc
}
