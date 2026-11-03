//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmRegionQuota(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()

	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()
	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()

	// Required information: Zones and classes
	regionZones, err := region.GetAllZones(nil)
	check.Assert(err, IsNil)
	check.Assert(len(regionZones) > 0, Equals, true)

	vmClasses, err := region.GetAllVmClasses(nil)
	check.Assert(err, IsNil)
	check.Assert(len(vmClasses) > 0, Equals, true)

	vdcType := &types.TmVdc{
		Name:        check.TestName(),
		Org:         &types.OpenApiReference{ID: org.TmOrg.ID},
		Region:      &types.OpenApiReference{ID: region.Region.ID},
		Supervisors: []types.OpenApiReference{{ID: supervisor.Supervisor.SupervisorID}},
		ZoneResourceAllocation: []*types.TmVdcZoneResourceAllocation{{
			Zone: &types.OpenApiReference{ID: regionZones[0].Zone.ID},
			ResourceAllocation: types.TmVdcResourceAllocation{
				CPUReservationMHz:    addrOf(int64(100)),
				CPULimitMHz:          addrOf(int64(500)),
				MemoryReservationMiB: addrOf(int64(256)),
				MemoryLimitMiB:       addrOf(int64(512)),
			},
		}},
	}

	createdVdc, err := vcd.client.CreateRegionQuota(vdcType)
	check.Assert(err, IsNil)
	check.Assert(createdVdc, NotNil)
	// Add to cleanup list
	PrependToCleanupListOpenApi(createdVdc.TmVdc.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmVdcs+createdVdc.TmVdc.ID)
	defer func() {
		err = createdVdc.DeleteWithForce(true)
		check.Assert(err, IsNil)
	}()

	// Assignment of VM Classes
	err = createdVdc.AssignVmClasses(&types.RegionVirtualMachineClasses{
		Values: types.OpenApiReferences{{Name: vmClasses[0].Name, ID: vmClasses[0].ID}},
	})

	check.Assert(err, IsNil)

	// Get TM VDC By Name
	byName, err := vcd.client.GetRegionQuotaByName(vdcType.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.TmVdc, DeepEquals, createdVdc.TmVdc)

	// Get TM VDC By Id
	byId, err := vcd.client.GetRegionQuotaById(createdVdc.TmVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmVdc, DeepEquals, createdVdc.TmVdc)

	// Get By Name and Org ID
	byNameAndOrgId, err := vcd.client.GetRegionQuotaByNameAndOrgId(createdVdc.TmVdc.Name, org.TmOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(byNameAndOrgId.TmVdc, DeepEquals, createdVdc.TmVdc)

	// Get By Name and Org ID in non existent Org
	byNameAndInvalidOrgId, err := vcd.client.GetRegionQuotaByNameAndOrgId(createdVdc.TmVdc.Name, "urn:vcloud:org:a93c9db9-0000-0000-0000-a8f7eeda85f9")
	check.Assert(err, NotNil)
	check.Assert(byNameAndInvalidOrgId, IsNil)

	// Not Found tests
	byNameInvalid, err := vcd.client.GetRegionQuotaByName("fake-name")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byNameInvalid, IsNil)

	byIdInvalid, err := vcd.client.GetRegionQuotaById("urn:vcloud:virtualDatacenter:5344b964-0000-0000-0000-d554913db643")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdInvalid, IsNil)

	// Update
	createdVdc.TmVdc.Name = check.TestName() + "-update"
	updatedVdc, err := createdVdc.Update(createdVdc.TmVdc)
	check.Assert(err, IsNil)
	check.Assert(updatedVdc.TmVdc, DeepEquals, createdVdc.TmVdc)
}

func (vcd *TestVCD) Test_TmRegionQuotaVmClassOperations(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	region, cleanup := tmTestRegion(vcd, check)
	defer cleanup()
	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()
	quota, quotaCleanup := createRegionQuota(vcd, org, region, check)
	defer quotaCleanup()
	classes, err := region.GetAllVmClasses(nil)
	check.Assert(err, IsNil)
	// Clear the test quota's initial assignment so POST has an observable effect.
	check.Assert(quota.AssignVmClasses(&types.RegionVirtualMachineClasses{Values: types.OpenApiReferences{}}), IsNil)
	payload := &types.RegionVirtualMachineClasses{Values: types.OpenApiReferences{{ID: classes[0].ID, Name: classes[0].Name}}}
	added, err := quota.AddVmClasses(payload)
	check.Assert(err, IsNil)
	check.Assert(len(added.Values), Equals, 1)
	check.Assert(added.Values[0].ID, Equals, classes[0].ID)
	assigned, err := quota.GetAllVmClasses(url.Values{"pageSize": {"1"}})
	check.Assert(err, IsNil)
	check.Assert(len(assigned), Equals, 1)
	check.Assert(assigned[0].ID, Equals, classes[0].ID)
	check.Assert(quota.AssignVmClasses(&types.RegionVirtualMachineClasses{Values: types.OpenApiReferences{}}), IsNil)
	added, err = vcd.client.AddVmClassesToRegionQuota(quota.TmVdc.ID, payload)
	check.Assert(err, IsNil)
	check.Assert(len(added.Values), Equals, 1)
	check.Assert(added.Values[0].ID, Equals, classes[0].ID)
	namespaces, err := quota.GetAllNamespaces(url.Values{"pageSize": {"1"}})
	check.Assert(err, IsNil)
	check.Assert(len(namespaces), Equals, 0)
	reservations, err := quota.GetAllReservedVmClasses(nil)
	check.Assert(err, IsNil)
	check.Assert(len(reservations), Equals, 0)
	updated, err := quota.UpdateReservedVmClasses(&types.TmVdcReservedVmClasses{Values: []types.TmVdcReservedVmClass{}})
	check.Assert(err, IsNil)
	check.Assert(len(updated.Values), Equals, 0)
}

func (vcd *TestVCD) Test_TmRegionQuotaReservedVmClasses(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	region := tmTestExistingRegion(vcd, check)
	available, err := vcd.client.GetAllTmReservedVmClasses(url.Values{"filter": {"region.id==" + region.Region.ID}})
	check.Assert(err, IsNil)
	var selected *types.TmReservedVmClass
	for _, row := range available {
		if row.AvailableCount > 0 && row.VirtualMachineClass != nil && row.Zone != nil {
			selected = row
			break
		}
	}
	if selected == nil {
		check.Skip("requires an available reserved VM class in tm.region")
	}
	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()
	// Use the reservation's zone, not an arbitrary first zone in the region.
	quota, err := vcd.client.CreateRegionQuota(&types.TmVdc{
		Name: check.TestName(), Org: &types.OpenApiReference{ID: org.TmOrg.ID}, Region: &types.OpenApiReference{ID: region.Region.ID},
		Supervisors:            region.Region.Supervisors,
		ZoneResourceAllocation: []*types.TmVdcZoneResourceAllocation{{Zone: selected.Zone}},
	})
	check.Assert(err, IsNil)
	defer func() { check.Check(quota.DeleteWithForce(false), IsNil) }()
	payload := &types.TmVdcReservedVmClasses{Values: []types.TmVdcReservedVmClass{{Name: selected.VirtualMachineClass.Name, Zone: selected.Zone, TotalCount: addrOf(int64(1))}}}
	updated, err := quota.UpdateReservedVmClasses(payload)
	check.Assert(err, IsNil)
	check.Assert(len(updated.Values), Equals, 1)
	check.Assert(updated.Values[0].TotalCount, NotNil)
	check.Assert(*updated.Values[0].TotalCount, Equals, int64(1))
	rows, err := quota.GetAllReservedVmClasses(url.Values{"pageSize": {"1"}})
	check.Assert(err, IsNil)
	check.Assert(len(rows), Equals, 1)
	check.Assert(rows[0].Name, Equals, selected.VirtualMachineClass.Name)
	check.Assert(rows[0].Zone.ID, Equals, selected.Zone.ID)
	check.Assert(rows[0].TotalCount, NotNil)
	check.Assert(*rows[0].TotalCount, Equals, int64(1))
	payload.Values[0].TotalCount = addrOf(int64(0))
	_, err = quota.UpdateReservedVmClasses(payload)
	check.Assert(err, IsNil)
	rows, err = quota.GetAllReservedVmClasses(nil)
	check.Assert(err, IsNil)
	for _, row := range rows {
		check.Assert(row.TotalCount, NotNil)
		check.Assert(*row.TotalCount, Equals, int64(0))
	}
}

func (vcd *TestVCD) Test_TmRegionQuotaNamespaces(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	quotas, err := vcd.client.GetAllRegionQuotas(nil)
	check.Assert(err, IsNil)
	found := false
	for _, quota := range quotas {
		namespaces, err := quota.GetAllNamespaces(url.Values{"pageSize": {"1"}}, tmTestOptions(vcd, check, "40.1"))
		check.Assert(err, IsNil)
		for _, ns := range namespaces {
			found = true
			check.Assert(ns.ID, Not(Equals), "")
			check.Assert(ns.Name, Not(Equals), "")
			check.Assert(ns.ProjectAssignment, NotNil)
			check.Assert(ns.Region, NotNil)
			check.Assert(ns.Region.ID, Equals, quota.TmVdc.Region.ID)
			check.Assert(len(ns.ZonalResourceAllocation) > 0, Equals, true)
		}
	}
	if !found {
		check.Skip("no existing namespaces found in region quotas")
	}
}

func (vcd *TestVCD) Test_TmRegionQuotaInfraPolicies(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	region := tmTestExistingRegion(vcd, check)
	quotas, err := vcd.client.GetAllRegionQuotas(nil)
	check.Assert(err, IsNil)
	policyID := ""
	for _, existing := range quotas {
		if existing.TmVdc.Region == nil || existing.TmVdc.Region.ID != region.Region.ID {
			continue
		}
		policies, err := existing.GetAllInfraPolicies(nil)
		check.Assert(err, IsNil)
		for _, policy := range policies {
			if policy.Status == "READY" && policy.InfraPolicy != nil && policy.InfraPolicy.ID != "" {
				policyID = policy.InfraPolicy.ID
				break
			}
		}
		if policyID != "" {
			break
		}
	}
	if policyID == "" {
		check.Skip("no ready infrastructure policy assigned to an existing quota in the region")
	}
	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()
	quota, quotaCleanup := createRegionQuota(vcd, org, region, check)
	defer quotaCleanup()
	updated, err := quota.UpdateInfraPolicies(&types.TmVdcInfraPolicies{Values: []types.TmVdcInfraPolicy{{InfraPolicy: &types.OpenApiReference{ID: policyID}}}})
	check.Assert(err, IsNil)
	check.Assert(len(updated.Values), Equals, 1)
	check.Assert(updated.Values[0].InfraPolicy.ID, Equals, policyID)
	rows, err := quota.GetAllInfraPolicies(url.Values{"pageSize": {"1"}})
	check.Assert(err, IsNil)
	check.Assert(len(rows), Equals, 1)
	check.Assert(rows[0].InfraPolicy.ID, Equals, policyID)
	_, err = quota.UpdateInfraPolicies(&types.TmVdcInfraPolicies{Values: []types.TmVdcInfraPolicy{}})
	check.Assert(err, IsNil)
	rows, err = quota.GetAllInfraPolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rows), Equals, 0)
}
