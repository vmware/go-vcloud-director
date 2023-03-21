//go:build functional || openapi || ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_TenantContext checks that different members of a hierarchy
// with an Admin Org at its top are all reporting the same tenant context
// using different methods to retrieve each member
// When running with -vcd-verbose, you should see a long list of entities
// with the corresponding tenant context. If no errors occur, the tenant context
// values in all rows should be the same.
func (vcd *TestVCD) Test_TenantContext(check *C) {
	// Check the tenant context of the AdminOrg (top of the hierarchy)
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	check.Assert(adminOrg.TenantContext, NotNil)
	check.Assert(adminOrg.TenantContext.OrgId, Equals, extractUuid(adminOrg.AdminOrg.ID))
	check.Assert(adminOrg.TenantContext.OrgName, Equals, adminOrg.AdminOrg.Name)
	adminOrgTenantContext := adminOrg.TenantContext
	checkTenantContext(check, "adminOrg by name", adminOrgTenantContext, adminOrgTenantContext)

	adminOrgById, err := vcd.client.GetAdminOrgById(adminOrg.AdminOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(adminOrgById, NotNil)
	check.Assert(adminOrgById.TenantContext, NotNil)
	check.Assert(adminOrgById.TenantContext.OrgId, Equals, extractUuid(adminOrgById.AdminOrg.ID))
	check.Assert(adminOrgById.TenantContext.OrgName, Equals, adminOrgById.AdminOrg.Name)
	adminOrgByIdTenantContext := adminOrgById.TenantContext
	checkTenantContext(check, "adminOrg by ID", adminOrgByIdTenantContext, adminOrgTenantContext)

	adminOrgByNameOrId, err := vcd.client.GetAdminOrgByNameOrId(adminOrg.AdminOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(adminOrgByNameOrId, NotNil)
	check.Assert(adminOrgByNameOrId.TenantContext, NotNil)
	check.Assert(adminOrgByNameOrId.TenantContext.OrgId, Equals, extractUuid(adminOrgByNameOrId.AdminOrg.ID))
	check.Assert(adminOrgByNameOrId.TenantContext.OrgName, Equals, adminOrgByNameOrId.AdminOrg.Name)
	adminOrgByNameOrIdTenantContext := adminOrgByNameOrId.TenantContext
	checkTenantContext(check, "adminOrg by ID", adminOrgByNameOrIdTenantContext, adminOrgTenantContext)

	// Check the tenant context of the Org (top of the hierarchy)
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	check.Assert(org.TenantContext, NotNil)
	check.Assert(org.TenantContext.OrgId, Equals, extractUuid(org.Org.ID))
	check.Assert(org.TenantContext.OrgName, Equals, org.Org.Name)
	orgTenantContext := org.TenantContext
	checkTenantContext(check, "org by name", orgTenantContext, adminOrgTenantContext)

	orgById, err := vcd.client.GetOrgById(org.Org.ID)
	check.Assert(err, IsNil)
	check.Assert(orgById, NotNil)
	check.Assert(orgById.TenantContext, NotNil)
	check.Assert(orgById.TenantContext.OrgId, Equals, extractUuid(orgById.Org.ID))
	check.Assert(orgById.TenantContext.OrgName, Equals, orgById.Org.Name)
	orgTenantContext = orgById.TenantContext
	checkTenantContext(check, "org by ID", orgTenantContext, adminOrgTenantContext)

	orgByNameOrId, err := vcd.client.GetOrgByNameOrId(org.Org.ID)
	check.Assert(err, IsNil)
	check.Assert(orgByNameOrId, NotNil)
	check.Assert(orgByNameOrId.TenantContext, NotNil)
	check.Assert(orgByNameOrId.TenantContext.OrgId, Equals, extractUuid(orgByNameOrId.Org.ID))
	check.Assert(orgByNameOrId.TenantContext.OrgName, Equals, orgByNameOrId.Org.Name)
	orgTenantContext = orgByNameOrId.TenantContext
	checkTenantContext(check, "org by name or ID", orgTenantContext, adminOrgTenantContext)

	// Check that an admin VDC depending from our org has the same tenant context
	adminVdc, err := adminOrg.GetAdminVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdc, NotNil)
	adminVdcTenantContext, err := adminVdc.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "adminVdc by name", adminVdcTenantContext, adminOrgTenantContext)

	adminVdcById, err := adminOrg.GetAdminVDCById(adminVdc.AdminVdc.ID, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdcById, NotNil)
	adminVdcTenantContext, err = adminVdcById.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "adminVdc by ID", adminVdcTenantContext, adminOrgTenantContext)

	adminVdcByNameOrId, err := adminOrg.GetAdminVDCByNameOrId(adminVdc.AdminVdc.ID, false)
	check.Assert(err, IsNil)
	check.Assert(adminVdcByNameOrId, NotNil)
	adminVdcTenantContext, err = adminVdcByNameOrId.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "adminVdc by name or ID", adminVdcTenantContext, adminOrgTenantContext)

	// Check that a VDC depending from our org has the same tenant context
	vdc, err := adminOrg.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	vdcTenantContext, err := vdc.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "VDC by name", vdcTenantContext, adminOrgTenantContext)

	vdcById, err := adminOrg.GetVDCById(vdc.Vdc.ID, false)
	check.Assert(err, IsNil)
	check.Assert(vdcById, NotNil)
	vdcTenantContext, err = vdcById.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "VDC by ID", vdcTenantContext, adminOrgTenantContext)

	vdcByNameOrId, err := adminOrg.GetVDCByNameOrId(vdc.Vdc.ID, false)
	check.Assert(err, IsNil)
	check.Assert(vdcByNameOrId, NotNil)
	vdcTenantContext, err = vdcByNameOrId.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "VDC by name or ID", vdcTenantContext, adminOrgTenantContext)

	// Check that an admin catalog depending from our org has the same tenant context
	adminCatalog, err := adminOrg.GetAdminCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog, NotNil)
	adminCatalogTenantContext, err := adminCatalog.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "adminCatalog by name", adminCatalogTenantContext, adminOrgTenantContext)

	adminCatalogById, err := adminOrg.GetAdminCatalogById(adminCatalog.AdminCatalog.ID, false)
	check.Assert(err, IsNil)
	check.Assert(adminCatalogById, NotNil)
	adminCatalogTenantContext, err = adminCatalogById.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "adminCatalog by ID", adminCatalogTenantContext, adminOrgTenantContext)

	adminCatalogByNameOrId, err := adminOrg.GetAdminCatalogByNameOrId(adminCatalog.AdminCatalog.ID, false)
	check.Assert(err, IsNil)
	check.Assert(adminCatalogByNameOrId, NotNil)
	adminCatalogTenantContext, err = adminCatalogByNameOrId.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "adminCatalog by Name or ID", adminCatalogTenantContext, adminOrgTenantContext)

	// Check that a catalog depending from our org has the same tenant context
	catalog, err := adminOrg.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	catalogTenantContext, err := catalog.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "catalog by Name", catalogTenantContext, adminOrgTenantContext)

	catalogById, err := adminOrg.GetCatalogById(catalog.Catalog.ID, false)
	check.Assert(err, IsNil)
	check.Assert(catalogById, NotNil)
	catalogTenantContext, err = catalogById.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "catalog by ID", catalogTenantContext, adminOrgTenantContext)

	catalogByNameOrId, err := adminOrg.GetCatalogByNameOrId(catalog.Catalog.ID, false)
	check.Assert(err, IsNil)
	check.Assert(catalogByNameOrId, NotNil)
	catalogTenantContext, err = catalogByNameOrId.getTenantContext()
	check.Assert(err, IsNil)
	checkTenantContext(check, "catalog by ID", catalogTenantContext, adminOrgTenantContext)

	vappList := vdc.GetVappList()

	if len(vappList) > 0 {
		// Check that a vApp depending from our org has the same tenant context
		vapp, err := vdc.GetVAppByName(vappList[0].Name, false)
		check.Assert(err, IsNil)
		check.Assert(vapp, NotNil)
		vappTenantContext, err := vapp.getTenantContext()
		check.Assert(err, IsNil)
		checkTenantContext(check, "vapp by name", vappTenantContext, adminOrgTenantContext)

		vappById, err := vdc.GetVAppById(vapp.VApp.ID, false)
		check.Assert(err, IsNil)
		check.Assert(vappById, NotNil)
		vappTenantContext, err = vappById.getTenantContext()
		check.Assert(err, IsNil)
		checkTenantContext(check, "vapp by ID", vappTenantContext, adminOrgTenantContext)

		vappByNameOrId, err := vdc.GetVAppByNameOrId(vapp.VApp.ID, false)
		check.Assert(err, IsNil)
		check.Assert(vappByNameOrId, NotNil)
		vappTenantContext, err = vappByNameOrId.getTenantContext()
		check.Assert(err, IsNil)
		checkTenantContext(check, "vapp by name or ID", vappTenantContext, adminOrgTenantContext)

		vmList, err := vdc.QueryVmList(types.VmQueryFilterOnlyDeployed)
		check.Assert(err, IsNil)
		if len(vmList) > 0 {
			// Check that a VM depending from our org has the same tenant context
			vm, err := vcd.client.Client.GetVMByHref(vmList[0].HREF)
			check.Assert(err, IsNil)
			check.Assert(vm, NotNil)

			vmTenantContext, err := vm.getTenantContext()
			check.Assert(err, IsNil)
			checkTenantContext(check, "VM by Href", vmTenantContext, adminOrgTenantContext)
		}
	}

	// Check that a VM depending from our org has the same tenant context
	role, err := adminOrg.GetRoleByName("vApp Author")
	check.Assert(err, IsNil)
	check.Assert(role, NotNil)
	checkTenantContext(check, "role by name", role.TenantContext, adminOrgTenantContext)

	roleById, err := adminOrg.GetRoleById(role.Role.ID)
	check.Assert(err, IsNil)
	check.Assert(role, NotNil)
	checkTenantContext(check, "role by Id", roleById.TenantContext, adminOrgTenantContext)
}

func checkTenantContext(check *C, label string, tenantContext, parentTenantContext *TenantContext) {
	check.Assert(tenantContext, DeepEquals, parentTenantContext)
	check.Assert(tenantContext.OrgId, Equals, parentTenantContext.OrgId)
	check.Assert(tenantContext.OrgName, Equals, parentTenantContext.OrgName)
	if testVerbose {
		fmt.Printf("%-30s %-20s -> %s\n", label, tenantContext.OrgId, parentTenantContext.OrgName)
	}
}
