// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/url"
)

const labelContentLibrary = "Content Library"

// ContentLibrary defines the Content Library data structure
type ContentLibrary struct {
	ContentLibrary *types.ContentLibrary
	vcdClient      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g ContentLibrary) wrap(inner *types.ContentLibrary) *ContentLibrary {
	g.ContentLibrary = inner
	return &g
}

// CreateContentLibrary creates a Content Library with the given tenant context. If tenant context is nil,
// it assumes that the Content Library to create is of Provider type.
// TODO: TM: Subscribed catalogs create Tasks for every OVA from the publisher. These are ignored at the moment, so a better mechanism must be implemented (like CreateCatalogFromSubscription for VCD catalogs)
func (vcdClient *VCDClient) CreateContentLibrary(config *types.ContentLibrary, ctx *TenantContext) (*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:      labelContentLibrary,
		endpoint:         types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}
	outerType := ContentLibrary{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// CreateContentLibrary creates a Content Library that belongs to the receiver Organization.
func (org *TmOrg) CreateContentLibrary(config *types.ContentLibrary) (*ContentLibrary, error) {
	return org.vcdClient.CreateContentLibrary(config, &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	})
}

// GetAllContentLibraries retrieves all Content Libraries with the given query parameters, which allow setting filters
// and other constraints. Tenant context can be nil, or can be used to retrieve the Content Libraries as a tenant.
func (vcdClient *VCDClient) GetAllContentLibraries(queryParameters url.Values, ctx *TenantContext) ([]*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:      labelContentLibrary,
		endpoint:         types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		queryParameters:  defaultPageSize(queryParameters, "64"), // Content Library endpoint forces to use maximum 64 items per page
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}

	outerType := ContentLibrary{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetAllContentLibraries retrieves all Content Libraries that belong to the receiver Organization
// and with the given query parameters, which allow setting filters and other constraints
func (org *TmOrg) GetAllContentLibraries(queryParameters url.Values) ([]*ContentLibrary, error) {
	return org.vcdClient.GetAllContentLibraries(queryParameters, &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	})
}

// GetContentLibraryByName retrieves a Content Library with the given name. Tenant context can be nil, or can be used to
// retrieve the Content Library as a tenant.
func (vcdClient *VCDClient) GetContentLibraryByName(name string, ctx *TenantContext) (*ContentLibrary, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("retrieving Content Libraries is only supported in TM")
	}

	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelContentLibrary)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllContentLibraries(queryParams, ctx)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetContentLibraryById(singleEntity.ContentLibrary.ID, ctx)
}

// GetContentLibraryByName retrieves a Content Library with the given name that belongs to the receiver Organization.
func (org *TmOrg) GetContentLibraryByName(name string) (*ContentLibrary, error) {
	return org.vcdClient.GetContentLibraryByName(name, &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	})
}

// GetContentLibraryById retrieves a Content Library with the given ID
func (vcdClient *VCDClient) GetContentLibraryById(id string, ctx *TenantContext) (*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:      labelContentLibrary,
		endpoint:         types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams:   []string{id},
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}

	outerType := ContentLibrary{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetContentLibraryById retrieves a Content Library with the given ID that belongs to the receiver Organization.
func (org *TmOrg) GetContentLibraryById(id string) (*ContentLibrary, error) {
	return org.vcdClient.GetContentLibraryById(id, &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	})
}

// Update updates an existing Content Library with the given configuration
func (o *ContentLibrary) Update(contentLibraryConfig *types.ContentLibrary) (*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams: []string{o.ContentLibrary.ID},
		requiresTm:     true,
	}
	outerType := ContentLibrary{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, contentLibraryConfig)
}

// Delete deletes the receiver Content Library.
// The 'recursive' flag deletes the Content Library, including its content library items, in a single operation.
// The 'force' flag forcefully deletes the Content Library and its items.
func (o *ContentLibrary) Delete(force, recursive bool) error {
	queryParams := url.Values{}
	queryParams.Add("force", fmt.Sprintf("%t", force))
	queryParams.Add("recursive", fmt.Sprintf("%t", recursive))

	c := crudConfig{
		entityLabel:     labelContentLibrary,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams:  []string{o.ContentLibrary.ID},
		queryParameters: queryParams,
		requiresTm:      true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
