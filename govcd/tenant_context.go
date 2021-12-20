/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TenantContext stores the information needed for an object to be used in the context of a given organization
type TenantContext struct {
	OrgId   string // The bare ID (without prefix) of an organization
	OrgName string // The organization name
}

// organization is an abstraction of types Org and AdminOrg
type organization interface {
	orgId() string
	orgName() string
	tenantContext() (*TenantContext, error)
	fullObject() interface{}
}

//lint:ignore U1000 for future usage
type genericVdc interface {
	vdcId() string
	vdcName() string
	vdcParent() interface{}
}

//lint:ignore U1000 for future usage
type genericCatalog interface {
	catalogId() string
	catalogName() string
	catalogParent() interface{}
}

// Implementation of organization interface for Org
func (org *Org) orgId() string                          { return org.Org.ID }
func (org *Org) orgName() string                        { return org.Org.Name }
func (org *Org) tenantContext() (*TenantContext, error) { return org.getTenantContext() }
func (org *Org) fullObject() interface{}                { return org }

// Implementation of organization interface for AdminOrg
func (adminOrg *AdminOrg) orgId() string                          { return adminOrg.AdminOrg.ID }
func (adminOrg *AdminOrg) orgName() string                        { return adminOrg.AdminOrg.Name }
func (adminOrg *AdminOrg) tenantContext() (*TenantContext, error) { return adminOrg.getTenantContext() }
func (adminOrg *AdminOrg) fullObject() interface{}                { return adminOrg }

// Implementation of genericVdc interface for Vdc
func (vdc *Vdc) vdcId() string          { return vdc.Vdc.ID }
func (vdc *Vdc) vdcName() string        { return vdc.Vdc.Name }
func (vdc *Vdc) vdcParent() interface{} { return vdc.parent }

// Implementation of genericVdc interface for AdminVdc
func (adminVdc *AdminVdc) vdcId() string          { return adminVdc.AdminVdc.ID }
func (adminVdc *AdminVdc) vdcName() string        { return adminVdc.AdminVdc.Name }
func (adminVdc *AdminVdc) vdcParent() interface{} { return adminVdc.parent }

// Implementation of genericCatalog interface for AdminCatalog
func (adminCatalog *AdminCatalog) catalogId() string          { return adminCatalog.AdminCatalog.ID }
func (adminCatalog *AdminCatalog) catalogName() string        { return adminCatalog.AdminCatalog.Name }
func (adminCatalog *AdminCatalog) catalogParent() interface{} { return adminCatalog.parent }

// Implementation of genericCatalog interface for AdminCatalog
func (catalog *Catalog) catalogId() string          { return catalog.Catalog.ID }
func (catalog *Catalog) catalogName() string        { return catalog.Catalog.Name }
func (catalog *Catalog) catalogParent() interface{} { return catalog.parent }

// getTenantContext returns the tenant context information for an Org
// If the information was not stored, it gets created and stored for future use
func (org *Org) getTenantContext() (*TenantContext, error) {
	if org.TenantContext == nil {
		id, err := getBareEntityUuid(org.Org.ID)
		if err != nil {
			return nil, err
		}
		org.TenantContext = &TenantContext{
			OrgId:   id,
			OrgName: org.Org.Name,
		}
	}
	return org.TenantContext, nil
}

// getTenantContext returns the tenant context information for an AdminOrg
// If the information was not stored, it gets created and stored for future use
func (org *AdminOrg) getTenantContext() (*TenantContext, error) {
	if org.TenantContext == nil {
		id, err := getBareEntityUuid(org.AdminOrg.ID)
		if err != nil {
			return nil, err
		}
		org.TenantContext = &TenantContext{
			OrgId:   id,
			OrgName: org.AdminOrg.Name,
		}
	}
	return org.TenantContext, nil
}

// getTenantContext retrieves the tenant context for an AdminVdc
func (vdc *AdminVdc) getTenantContext() (*TenantContext, error) {
	org := vdc.parent

	if org == nil {
		return nil, fmt.Errorf("VDC %s has no parent", vdc.AdminVdc.Name)
	}
	return org.tenantContext()
}

// getTenantContext retrieves the tenant context for a VDC
func (vdc *Vdc) getTenantContext() (*TenantContext, error) {
	org := vdc.parent

	if org == nil {
		return nil, fmt.Errorf("VDC %s has no parent", vdc.Vdc.Name)
	}
	return org.tenantContext()
}

// getTenantContext retrieves the tenant context for an AdminCatalog
func (catalog *AdminCatalog) getTenantContext() (*TenantContext, error) {
	org := catalog.parent

	if org == nil {
		return nil, fmt.Errorf("catalog %s has no parent", catalog.AdminCatalog.Name)
	}
	return org.tenantContext()
}

// getTenantContext retrieves the tenant context for a Catalog
func (catalog *Catalog) getTenantContext() (*TenantContext, error) {
	org := catalog.parent

	if org == nil {
		return nil, fmt.Errorf("catalog %s has no parent", catalog.Catalog.Name)
	}
	return org.tenantContext()
}

// getTenantContextHeader returns a map of strings containing the tenant context items
// needed to be used in http.Request.Header
func getTenantContextHeader(tenantContext *TenantContext) map[string]string {
	if tenantContext == nil {
		return nil
	}
	if tenantContext.OrgName == "" || strings.EqualFold(tenantContext.OrgName, "system") {
		return nil
	}
	return map[string]string{
		types.HeaderTenantContext: tenantContext.OrgId,
		types.HeaderAuthContext:   tenantContext.OrgName,
	}
}

// getTenantContextFromHeader does the opposite of getTenantContextHeader:
// given a header, returns a TenantContext
func getTenantContextFromHeader(header map[string]string) *TenantContext {
	if len(header) == 0 {
		return nil
	}
	tenantContext, okTenant := header[types.HeaderTenantContext]
	AuthContext, okAuth := header[types.HeaderAuthContext]
	if okTenant && okAuth {
		return &TenantContext{
			OrgId:   tenantContext,
			OrgName: AuthContext,
		}
	}
	return nil
}

// getTenantContext retrieves the tenant context for a VdcGroup
func (vdcGroup *VdcGroup) getTenantContext() (*TenantContext, error) {
	org := vdcGroup.parent

	if org == nil {
		return nil, fmt.Errorf("VDC group %s has no parent", vdcGroup.VdcGroup.Name)
	}
	return org.tenantContext()
}
