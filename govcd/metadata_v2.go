/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net/http"
	"regexp"
	"strings"
)

// NOTE: This "v2" is not v2 in terms of API versioning, it's just a way to separate the functions that handle
// metadata in a complete way (v2, this file) and the deprecated functions that were incomplete (v1, they lacked
// "visibility" and "domain" handling).
//
// The idea is that once a new major version of go-vcloud-director is released, one can just remove "v1" file and perform
// a minor refactoring of the code here (probably renaming functions). Also, the code in "v2" is organized differently,
// as this is classified using "CRUD blocks" (meaning that all Create functions are together, same for Read... etc),
// which makes the code more readable.

// ------------------------------------------------------------------------------------------------
// GET metadata by key
// ------------------------------------------------------------------------------------------------

// GetMetadataByKeyAndHref returns metadata from the given resource reference, corresponding to the given key and domain.
func (vcdClient *VCDClient) GetMetadataByKeyAndHref(href, key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(&vcdClient.Client, href, "", key, isSystem)
}

// GetMetadataByKey returns VM metadata corresponding to the given key and domain.
func (vm *VM) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(vm.client, vm.VM.HREF, vm.VM.Name, key, isSystem)
}

// GetMetadataByKey returns VDC metadata corresponding to the given key and domain.
func (vdc *Vdc) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(vdc.client, vdc.Vdc.HREF, vdc.Vdc.Name, key, isSystem)
}

// GetMetadataByKey returns AdminVdc metadata corresponding to the given key and domain.
func (adminVdc *AdminVdc) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(adminVdc.client, adminVdc.AdminVdc.HREF, adminVdc.AdminVdc.Name, key, isSystem)
}

// GetMetadataByKey returns ProviderVdc metadata corresponding to the given key and domain.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(providerVdc.client, providerVdc.ProviderVdc.HREF, providerVdc.ProviderVdc.Name, key, isSystem)
}

// GetMetadataByKey returns VApp metadata corresponding to the given key and domain.
func (vapp *VApp) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(vapp.client, vapp.VApp.HREF, vapp.VApp.Name, key, isSystem)
}

// GetMetadataByKey returns VAppTemplate metadata corresponding to the given key and domain.
func (vAppTemplate *VAppTemplate) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, vAppTemplate.VAppTemplate.Name, key, isSystem)
}

// GetMetadataByKey returns MediaRecord metadata corresponding to the given key and domain.
func (mediaRecord *MediaRecord) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(mediaRecord.client, mediaRecord.MediaRecord.HREF, mediaRecord.MediaRecord.Name, key, isSystem)
}

// GetMetadataByKey returns Media metadata corresponding to the given key and domain.
func (media *Media) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(media.client, media.Media.HREF, media.Media.Name, key, isSystem)
}

// GetMetadataByKey returns Catalog metadata corresponding to the given key and domain.
func (catalog *Catalog) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(catalog.client, catalog.Catalog.HREF, catalog.Catalog.Name, key, isSystem)
}

// GetMetadataByKey returns AdminCatalog metadata corresponding to the given key and domain.
func (adminCatalog *AdminCatalog) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(adminCatalog.client, adminCatalog.AdminCatalog.HREF, adminCatalog.AdminCatalog.Name, key, isSystem)
}

// GetMetadataByKey returns the Org metadata corresponding to the given key and domain.
func (org *Org) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(org.client, org.Org.HREF, org.Org.Name, key, isSystem)
}

// GetMetadataByKey returns the AdminOrg metadata corresponding to the given key and domain.
// Note: Requires system administrator privileges.
func (adminOrg *AdminOrg) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(adminOrg.client, adminOrg.AdminOrg.HREF, adminOrg.AdminOrg.Name, key, isSystem)
}

// GetMetadataByKey returns the metadata corresponding to the given key and domain.
func (disk *Disk) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(disk.client, disk.Disk.HREF, disk.Disk.Name, key, isSystem)
}

// GetMetadataByKey returns OrgVDCNetwork metadata corresponding to the given key and domain.
func (orgVdcNetwork *OrgVDCNetwork) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(orgVdcNetwork.client, orgVdcNetwork.OrgVDCNetwork.HREF, orgVdcNetwork.OrgVDCNetwork.Name, key, isSystem)
}

// GetMetadataByKey returns CatalogItem metadata corresponding to the given key and domain.
func (catalogItem *CatalogItem) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	return getMetadataByKey(catalogItem.client, catalogItem.CatalogItem.HREF, catalogItem.CatalogItem.Name, key, isSystem)
}

// GetMetadataByKey returns OpenApiOrgVdcNetwork metadata corresponding to the given key and domain.
// NOTE: This function cannot retrieve metadata if the network belongs to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error) {
	href := fmt.Sprintf("%s/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	return getMetadataByKey(openApiOrgVdcNetwork.client, href, openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.Name, key, isSystem)
}

// ------------------------------------------------------------------------------------------------
// GET all metadata
// ------------------------------------------------------------------------------------------------

// GetMetadataByHref returns metadata from the given resource reference.
func (vcdClient *VCDClient) GetMetadataByHref(href string) (*types.Metadata, error) {
	return getMetadata(&vcdClient.Client, href, "")
}

// GetMetadata returns VM metadata.
func (vm *VM) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vm.client, vm.VM.HREF, vm.VM.Name)
}

// GetMetadata returns VDC metadata.
func (vdc *Vdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vdc.client, vdc.Vdc.HREF, vdc.Vdc.Name)
}

// GetMetadata returns AdminVdc metadata.
func (adminVdc *AdminVdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminVdc.client, adminVdc.AdminVdc.HREF, adminVdc.AdminVdc.Name)
}

// GetMetadata returns ProviderVdc metadata.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(providerVdc.client, providerVdc.ProviderVdc.HREF, providerVdc.ProviderVdc.Name)
}

// GetMetadata returns VApp metadata.
func (vapp *VApp) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vapp.client, vapp.VApp.HREF, vapp.VApp.Name)
}

// GetMetadata returns VAppTemplate metadata.
func (vAppTemplate *VAppTemplate) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, vAppTemplate.VAppTemplate.Name)
}

// GetMetadata returns MediaRecord metadata.
func (mediaRecord *MediaRecord) GetMetadata() (*types.Metadata, error) {
	return getMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF, mediaRecord.MediaRecord.Name)
}

// GetMetadata returns Media metadata.
func (media *Media) GetMetadata() (*types.Metadata, error) {
	return getMetadata(media.client, media.Media.HREF, media.Media.Name)
}

// GetMetadata returns Catalog metadata.
func (catalog *Catalog) GetMetadata() (*types.Metadata, error) {
	return getMetadata(catalog.client, catalog.Catalog.HREF, catalog.Catalog.Name)
}

// GetMetadata returns AdminCatalog metadata.
func (adminCatalog *AdminCatalog) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF, adminCatalog.AdminCatalog.Name)
}

// GetMetadata returns the Org metadata of the corresponding organization seen as administrator
func (org *Org) GetMetadata() (*types.Metadata, error) {
	return getMetadata(org.client, org.Org.HREF, org.Org.Name)
}

// GetMetadata returns the AdminOrg metadata of the corresponding organization seen as administrator
func (adminOrg *AdminOrg) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminOrg.client, adminOrg.AdminOrg.HREF, adminOrg.AdminOrg.Name)
}

// GetMetadata returns the metadata of the corresponding independent disk
func (disk *Disk) GetMetadata() (*types.Metadata, error) {
	return getMetadata(disk.client, disk.Disk.HREF, disk.Disk.Name)
}

// GetMetadata returns OrgVDCNetwork metadata.
func (orgVdcNetwork *OrgVDCNetwork) GetMetadata() (*types.Metadata, error) {
	return getMetadata(orgVdcNetwork.client, orgVdcNetwork.OrgVDCNetwork.HREF, orgVdcNetwork.OrgVDCNetwork.Name)
}

// GetMetadata returns CatalogItem metadata.
func (catalogItem *CatalogItem) GetMetadata() (*types.Metadata, error) {
	return getMetadata(catalogItem.client, catalogItem.CatalogItem.HREF, catalogItem.CatalogItem.Name)
}

// GetMetadata returns OpenApiOrgVdcNetwork metadata.
// NOTE: This function cannot retrieve metadata if the network belongs to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) GetMetadata() (*types.Metadata, error) {
	href := fmt.Sprintf("%s/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	return getMetadata(openApiOrgVdcNetwork.client, href, openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.Name)
}

// ------------------------------------------------------------------------------------------------
// ADD metadata async
// ------------------------------------------------------------------------------------------------

// AddMetadataEntryWithVisibilityByHrefAsync adds metadata to the given resource reference with the given key, value, type and visibility
// and returns the task.
func (vcdClient *VCDClient) AddMetadataEntryWithVisibilityByHrefAsync(href, key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(&vcdClient.Client, href, "", key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VM with the given key, value, type and visibility
// // and returns the task.
func (vm *VM) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vm.client, vm.VM.HREF, vm.VM.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given AdminVdc with the given key, value, type and visibility
// and returns the task.
func (adminVdc *AdminVdc) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(adminVdc.client, adminVdc.AdminVdc.HREF, adminVdc.AdminVdc.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given ProviderVdc with the given key, value, type and visibility
// and returns the task.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(providerVdc.client, providerVdc.ProviderVdc.HREF, providerVdc.ProviderVdc.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VApp with the given key, value, type and visibility
// and returns the task.
func (vapp *VApp) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vapp.client, vapp.VApp.HREF, vapp.VApp.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VAppTemplate with the given key, value, type and visibility
// and returns the task.
func (vAppTemplate *VAppTemplate) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, vAppTemplate.VAppTemplate.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given MediaRecord with the given key, value, type and visibility
// and returns the task.
func (mediaRecord *MediaRecord) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF, mediaRecord.MediaRecord.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Media with the given key, value, type and visibility
// and returns the task.
func (media *Media) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(media.client, media.Media.HREF, media.Media.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given AdminCatalog with the given key, value, type and visibility
// and returns the task.
func (adminCatalog *AdminCatalog) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF, adminCatalog.AdminCatalog.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given AdminOrg with the given key, value, type and visibility
// and returns the task.
func (adminOrg *AdminOrg) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(adminOrg.client, adminOrg.AdminOrg.HREF, adminOrg.AdminOrg.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Disk with the given key, value, type and visibility
// and returns the task.
func (disk *Disk) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(disk.client, disk.Disk.HREF, disk.Disk.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given OrgVDCNetwork with the given key, value, type and visibility
// and returns the task.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), orgVdcNetwork.OrgVDCNetwork.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Catalog Item with the given key, value, type and visibility
// and returns the task.
func (catalogItem *CatalogItem) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(catalogItem.client, catalogItem.CatalogItem.HREF, catalogItem.CatalogItem.Name, key, value, typedValue, visibility, isSystem)
}

// ------------------------------------------------------------------------------------------------
// ADD metadata
// ------------------------------------------------------------------------------------------------

// AddMetadataEntryWithVisibilityByHref adds metadata to the given resource reference with the given key, value, type and visibility
// and waits for completion.
func (vcdClient *VCDClient) AddMetadataEntryWithVisibilityByHref(href, key, value, typedValue, visibility string, isSystem bool) error {
	task, err := vcdClient.AddMetadataEntryWithVisibilityByHrefAsync(href, key, value, typedValue, visibility, isSystem)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// AddMetadataEntryWithVisibility adds metadata to the receiver VM and waits for the task to finish.
func (vm *VM) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(vm.client, vm.VM.HREF, vm.VM.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver AdminVdc and waits for the task to finish.
func (adminVdc *AdminVdc) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(adminVdc.client, adminVdc.AdminVdc.HREF, adminVdc.AdminVdc.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver ProviderVdc and waits for the task to finish.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(providerVdc.client, providerVdc.ProviderVdc.HREF, providerVdc.ProviderVdc.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver VApp and waits for the task to finish.
func (vapp *VApp) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(vapp.client, vapp.VApp.HREF, vapp.VApp.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver VAppTemplate and waits for the task to finish.
func (vAppTemplate *VAppTemplate) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, vAppTemplate.VAppTemplate.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver MediaRecord and waits for the task to finish.
func (mediaRecord *MediaRecord) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(mediaRecord.client, mediaRecord.MediaRecord.HREF, mediaRecord.MediaRecord.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver Media and waits for the task to finish.
func (media *Media) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(media.client, media.Media.HREF, media.Media.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver AdminCatalog and waits for the task to finish.
func (adminCatalog *AdminCatalog) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(adminCatalog.client, adminCatalog.AdminCatalog.HREF, adminCatalog.AdminCatalog.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver AdminOrg and waits for the task to finish.
func (adminOrg *AdminOrg) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(adminOrg.client, adminOrg.AdminOrg.HREF, adminOrg.AdminOrg.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver Disk and waits for the task to finish.
func (disk *Disk) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(disk.client, disk.Disk.HREF, disk.Disk.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver OrgVDCNetwork and waits for the task to finish.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), orgVdcNetwork.OrgVDCNetwork.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver CatalogItem and waits for the task to finish.
func (catalogItem *CatalogItem) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(catalogItem.client, catalogItem.CatalogItem.HREF, catalogItem.CatalogItem.Name, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver OpenApiOrgVdcNetwork and waits for the task to finish.
// Note: It doesn't add metadata to networks that belong to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	href := fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	task, err := addMetadata(openApiOrgVdcNetwork.client, href, openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.Name, key, value, typedValue, visibility, isSystem)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// ------------------------------------------------------------------------------------------------
// MERGE metadata async
// ------------------------------------------------------------------------------------------------

// MergeMetadataWithVisibilityByHrefAsync updates the metadata entries present in the referenced entity and creates the ones not present, then
// returns the task.
func (vcdClient *VCDClient) MergeMetadataWithVisibilityByHrefAsync(href string, metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(&vcdClient.Client, href, "", metadata)
}

// MergeMetadataWithMetadataValuesAsync merges VM metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then returns the task.
func (vm *VM) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(vm.client, vm.VM.HREF, vm.VM.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges AdminVdc metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminVdc *AdminVdc) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(adminVdc.client, adminVdc.AdminVdc.HREF, adminVdc.AdminVdc.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges Provider VDC metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(providerVdc.client, providerVdc.ProviderVdc.HREF, providerVdc.ProviderVdc.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges VApp metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vapp *VApp) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(vapp.client, vapp.VApp.HREF, vapp.VApp.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges VAppTemplate metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vAppTemplate *VAppTemplate) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, vAppTemplate.VAppTemplate.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges MediaRecord metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (mediaRecord *MediaRecord) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF, mediaRecord.MediaRecord.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges Media metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (media *Media) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(media.client, media.Media.HREF, media.Media.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges AdminCatalog metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminCatalog *AdminCatalog) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF, adminCatalog.AdminCatalog.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges AdminOrg metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminOrg *AdminOrg) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(adminOrg.client, adminOrg.AdminOrg.HREF, adminOrg.AdminOrg.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges Disk metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (disk *Disk) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(disk.client, disk.Disk.HREF, disk.Disk.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges OrgVDCNetwork metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), orgVdcNetwork.OrgVDCNetwork.Name, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges CatalogItem metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (catalogItem *CatalogItem) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(catalogItem.client, catalogItem.CatalogItem.HREF, catalogItem.CatalogItem.Name, metadata)
}

// ------------------------------------------------------------------------------------------------
// MERGE metadata
// ------------------------------------------------------------------------------------------------

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver VM and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (vm *VM) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(vm.client, vm.VM.HREF, vm.VM.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver AdminVdc and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (adminVdc *AdminVdc) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(adminVdc.client, adminVdc.AdminVdc.HREF, adminVdc.AdminVdc.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver ProviderVdc and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(providerVdc.client, providerVdc.ProviderVdc.HREF, providerVdc.ProviderVdc.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver VApp and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (vApp *VApp) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(vApp.client, vApp.VApp.HREF, vApp.VApp.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver VAppTemplate and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (vAppTemplate *VAppTemplate) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, vAppTemplate.VAppTemplate.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver MediaRecord and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (mediaRecord *MediaRecord) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(mediaRecord.client, mediaRecord.MediaRecord.HREF, mediaRecord.MediaRecord.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver Media and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (media *Media) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(media.client, media.Media.HREF, media.Media.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver AdminCatalog and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (adminCatalog *AdminCatalog) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(adminCatalog.client, adminCatalog.AdminCatalog.HREF, adminCatalog.AdminCatalog.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver AdminOrg and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (adminOrg *AdminOrg) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(adminOrg.client, adminOrg.AdminOrg.HREF, adminOrg.AdminOrg.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver Disk and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (disk *Disk) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(disk.client, disk.Disk.HREF, disk.Disk.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver OrgVDCNetwork and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), orgVdcNetwork.OrgVDCNetwork.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver CatalogItem and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (catalogItem *CatalogItem) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(catalogItem.client, catalogItem.CatalogItem.HREF, catalogItem.CatalogItem.Name, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver OpenApiOrgVdcNetwork and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
// Note: It doesn't merge metadata to networks that belong to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	href := fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	task, err := mergeAllMetadata(openApiOrgVdcNetwork.client, href, openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.Name, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// ------------------------------------------------------------------------------------------------
// DELETE metadata async
// ------------------------------------------------------------------------------------------------

// DeleteMetadataEntryWithDomainByHrefAsync deletes metadata from the given resource reference, depending on key provided as input
// and returns a task.
func (vcdClient *VCDClient) DeleteMetadataEntryWithDomainByHrefAsync(href, key string, isSystem bool) (Task, error) {
	return deleteMetadata(&vcdClient.Client, href, "", key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes VM metadata associated to the input key and returns the task.
func (vm *VM) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(vm.client, vm.VM.HREF, vm.VM.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes AdminVdc metadata associated to the input key and returns the task.
func (adminVdc *AdminVdc) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(adminVdc.client, adminVdc.AdminVdc.HREF, adminVdc.AdminVdc.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes ProviderVdc metadata associated to the input key and returns the task.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(providerVdc.client, providerVdc.ProviderVdc.HREF, providerVdc.ProviderVdc.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes VApp metadata associated to the input key and returns the task.
func (vapp *VApp) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(vapp.client, vapp.VApp.HREF, vapp.VApp.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes VAppTemplate metadata associated to the input key and returns the task.
func (vAppTemplate *VAppTemplate) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, vAppTemplate.VAppTemplate.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes MediaRecord metadata associated to the input key and returns the task.
func (mediaRecord *MediaRecord) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF, mediaRecord.MediaRecord.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes Media metadata associated to the input key and returns the task.
func (media *Media) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(media.client, media.Media.HREF, media.Media.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes AdminCatalog metadata associated to the input key and returns the task.
func (adminCatalog *AdminCatalog) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF, adminCatalog.AdminCatalog.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes AdminOrg metadata associated to the input key and returns the task.
func (adminOrg *AdminOrg) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(adminOrg.client, adminOrg.AdminOrg.HREF, adminOrg.AdminOrg.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes Disk metadata associated to the input key and returns the task.
func (disk *Disk) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(disk.client, disk.Disk.HREF, disk.Disk.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes OrgVDCNetwork metadata associated to the input key and returns the task.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), orgVdcNetwork.OrgVDCNetwork.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomainAsync deletes CatalogItem metadata associated to the input key and returns the task.
func (catalogItem *CatalogItem) DeleteMetadataEntryWithDomainAsync(key string, isSystem bool) (Task, error) {
	return deleteMetadata(catalogItem.client, catalogItem.CatalogItem.HREF, catalogItem.CatalogItem.Name, key, isSystem)
}

// ------------------------------------------------------------------------------------------------
// DELETE metadata
// ------------------------------------------------------------------------------------------------

// DeleteMetadataEntryWithDomainByHref deletes metadata from the given resource reference, depending on key provided as input
// and waits for the task to finish.
func (vcdClient *VCDClient) DeleteMetadataEntryWithDomainByHref(href, key string, isSystem bool) error {
	task, err := vcdClient.DeleteMetadataEntryWithDomainByHrefAsync(href, key, isSystem)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntryWithDomain deletes VM metadata associated to the input key and waits for the task to finish.
func (vm *VM) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(vm.client, vm.VM.HREF, vm.VM.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes AdminVdc metadata associated to the input key and waits for the task to finish.
// Note: Requires system administrator privileges.
func (adminVdc *AdminVdc) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(adminVdc.client, getAdminURL(adminVdc.AdminVdc.HREF), adminVdc.AdminVdc.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes ProviderVdc metadata associated to the input key and waits for the task to finish.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(providerVdc.client, providerVdc.ProviderVdc.HREF, providerVdc.ProviderVdc.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes VApp metadata associated to the input key and waits for the task to finish.
func (vApp *VApp) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(vApp.client, vApp.VApp.HREF, vApp.VApp.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes VAppTemplate metadata associated to the input key and waits for the task to finish.
func (vAppTemplate *VAppTemplate) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, vAppTemplate.VAppTemplate.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes MediaRecord metadata associated to the input key and waits for the task to finish.
func (mediaRecord *MediaRecord) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(mediaRecord.client, mediaRecord.MediaRecord.HREF, mediaRecord.MediaRecord.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes Media metadata associated to the input key and waits for the task to finish.
func (media *Media) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(media.client, media.Media.HREF, media.Media.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes AdminCatalog metadata associated to the input key and waits for the task to finish.
func (adminCatalog *AdminCatalog) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(adminCatalog.client, adminCatalog.AdminCatalog.HREF, adminCatalog.AdminCatalog.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes AdminOrg metadata associated to the input key and waits for the task to finish.
func (adminOrg *AdminOrg) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(adminOrg.client, adminOrg.AdminOrg.HREF, adminOrg.AdminOrg.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes Disk metadata associated to the input key and waits for the task to finish.
func (disk *Disk) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(disk.client, disk.Disk.HREF, disk.Disk.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes OrgVDCNetwork metadata associated to the input key and waits for the task to finish.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), orgVdcNetwork.OrgVDCNetwork.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes CatalogItem metadata associated to the input key and waits for the task to finish.
func (catalogItem *CatalogItem) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	return deleteMetadataAndWait(catalogItem.client, catalogItem.CatalogItem.HREF, catalogItem.CatalogItem.Name, key, isSystem)
}

// DeleteMetadataEntryWithDomain deletes OpenApiOrgVdcNetwork metadata associated to the input key and waits for the task to finish.
// Note: It doesn't delete metadata from networks that belong to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) DeleteMetadataEntryWithDomain(key string, isSystem bool) error {
	href := fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	task, err := deleteMetadata(openApiOrgVdcNetwork.client, href, openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.Name, key, isSystem)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// ------------------------------------------------------------------------------------------------
// Ignored metadata set/unset
// ------------------------------------------------------------------------------------------------

// SetMetadataToIgnore allows to update the metadata to be ignored in all metadata API calls with
// the given input. It returns the old IgnoredMetadata configuration from the client
func (vcdClient *VCDClient) SetMetadataToIgnore(ignoredMetadata []IgnoredMetadata) []IgnoredMetadata {
	result := vcdClient.Client.IgnoredMetadata
	vcdClient.Client.IgnoredMetadata = ignoredMetadata
	return result
}

// ------------------------------------------------------------------------------------------------
// Generic private functions
// ------------------------------------------------------------------------------------------------

// getMetadata is a generic function to retrieve metadata from VCD
func getMetadataByKey(client *Client, requestUri, name, key string, isSystem bool) (*types.MetadataValue, error) {
	metadata := &types.MetadataValue{}
	href := requestUri + "/metadata/"

	if isSystem {
		href += "SYSTEM/"
	}

	_, err := client.ExecuteRequest(href+key, http.MethodGet, types.MimeMetaData, "error retrieving metadata by key "+key+": %s", nil, metadata)
	if err != nil {
		return nil, err
	}
	return filterSingleMetadataEntry(key, requestUri, name, metadata, client.IgnoredMetadata)
}

// getMetadata is a generic function to retrieve metadata from VCD
func getMetadata(client *Client, requestUri, name string) (*types.Metadata, error) {
	metadata := &types.Metadata{}

	_, err := client.ExecuteRequest(requestUri+"/metadata/", http.MethodGet, types.MimeMetaData, "error retrieving metadata: %s", nil, metadata)
	if err != nil {
		return nil, err
	}
	return filterMetadata(metadata, requestUri, name, client.IgnoredMetadata)
}

// addMetadata adds metadata to an entity.
// If the metadata entry is of the SYSTEM domain (isSystem=true), one can set different types of Visibility:
// types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility but NOT types.MetadataReadWriteVisibility.
// If the metadata entry is of the GENERAL domain (isSystem=false), visibility is always types.MetadataReadWriteVisibility.
// In terms of typedValues, that must be one of:
// types.MetadataStringValue, types.MetadataNumberValue, types.MetadataDateTimeValue and types.MetadataBooleanValue.
func addMetadata(client *Client, requestUri, name, key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	apiEndpoint := urlParseRequestURI(requestUri)
	newMetadata := &types.MetadataValue{
		Xmlns: types.XMLNamespaceVCloud,
		Xsi:   types.XMLNamespaceXSI,
		TypedValue: &types.MetadataTypedValue{
			XsiType: typedValue,
			Value:   value,
		},
		Domain: &types.MetadataDomainTag{
			Visibility: visibility,
			Domain:     "SYSTEM",
		},
	}

	if isSystem {
		apiEndpoint.Path += "/metadata/SYSTEM/" + key
	} else {
		apiEndpoint.Path += "/metadata/" + key
		newMetadata.Domain.Domain = "GENERAL"
		if visibility != types.MetadataReadWriteVisibility {
			newMetadata.Domain.Visibility = types.MetadataReadWriteVisibility
		}
	}

	_, err := filterSingleMetadataEntry(key, requestUri, name, newMetadata, client.IgnoredMetadata)
	if err != nil {
		return Task{}, err
	}

	domain := newMetadata.Domain.Visibility
	task, err := client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut, types.MimeMetaDataValue, "error adding metadata: %s", newMetadata)

	// Workaround for ugly error returned by VCD: "API Error: 500: [ <uuid> ] visibility"
	if err != nil && strings.HasSuffix(err.Error(), "visibility") {
		err = fmt.Errorf("error adding metadata with key %s: visibility cannot be %s when domain is %s: %s", key, visibility, domain, err)
	}
	return task, err
}

// addMetadataAndWait adds metadata to an entity and waits for the task completion.
// The function supports passing a value that requires a typed value that must be one of:
// types.MetadataStringValue, types.MetadataNumberValue, types.MetadataDateTimeValue and types.MetadataBooleanValue.
// Visibility also needs to be one of: types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility or types.MetadataReadWriteVisibility
func addMetadataAndWait(client *Client, requestUri, name, key, value, typedValue, visibility string, isSystem bool) error {
	task, err := addMetadata(client, requestUri, name, key, value, typedValue, visibility, isSystem)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// mergeAllMetadata updates the metadata values that are already present in VCD and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// If the operation is successful, it returns the created task.
func mergeAllMetadata(client *Client, requestUri, name string, metadata map[string]types.MetadataValue) (Task, error) {
	var metadataToMerge []*types.MetadataEntry
	for key, value := range metadata {
		metadataToMerge = append(metadataToMerge, &types.MetadataEntry{
			Xmlns:      types.XMLNamespaceVCloud,
			Xsi:        types.XMLNamespaceXSI,
			Key:        key,
			TypedValue: value.TypedValue,
			Domain:     value.Domain,
		})
	}

	newMetadata := &types.Metadata{
		Xmlns:         types.XMLNamespaceVCloud,
		Xsi:           types.XMLNamespaceXSI,
		MetadataEntry: metadataToMerge,
	}

	apiEndpoint := urlParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata"

	filteredMetadata, err := filterMetadata(newMetadata, requestUri, name, client.IgnoredMetadata)
	if err != nil {
		return Task{}, err
	}
	if len(filteredMetadata.MetadataEntry) == 0 {
		return Task{}, fmt.Errorf("after filtering metadata, there is no metadata to merge")
	}

	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost, types.MimeMetaData, "error merging metadata: %s", filteredMetadata)
}

// mergeAllMetadata updates the metadata values that are already present in VCD and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func mergeMetadataAndWait(client *Client, requestUri, name string, metadata map[string]types.MetadataValue) error {
	task, err := mergeAllMetadata(client, requestUri, name, metadata)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// deleteMetadata deletes metadata associated to the input key from an entity referenced by its URI, then returns the
// task.
func deleteMetadata(client *Client, requestUri, name, key string, isSystem bool) (Task, error) {
	apiEndpoint := urlParseRequestURI(requestUri)
	if isSystem {
		apiEndpoint.Path += "/metadata/SYSTEM/" + key
	} else {
		apiEndpoint.Path += "/metadata/" + key
	}

	err := filterMetadataToDelete(client, key, requestUri, name, isSystem, client.IgnoredMetadata)
	if err != nil {
		return Task{}, err
	}

	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodDelete, "", "error deleting metadata: %s", nil)
}

// deleteMetadata deletes metadata associated to the input key from an entity referenced by its URI.
func deleteMetadataAndWait(client *Client, requestUri, name, key string, isSystem bool) error {
	task, err := deleteMetadata(client, requestUri, name, key, isSystem)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// IgnoredMetadata is a structure that defines the metadata entries that should be ignored by the VCD Client.
// The filtering works in such a way that all the non-nil pointers in an instance of this struct are evaluated with
// a logical AND.
// For example, ignoredMetadata.ObjectType = "org", ignoredMetadata.ObjectName = "foo" will ignore all metadata from
// Organizations whose name is "foo", with any key and any value.
// Note: This struct is only used by metadata_v2.go methods.
// Note 2: Filtering by ObjectName is not possible in the "ByHref" methods from VCDClient.
type IgnoredMetadata struct {
	ObjectType *string        // Type of the object that has the metadata as defined in the API documentation https://developer.vmware.com/apis/1601/vmware-cloud-director, for example "catalog", "disk", "org"...
	ObjectName *string        // Name of the object
	KeyRegex   *regexp.Regexp // A regular expression to filter out metadata keys
	ValueRegex *regexp.Regexp // A regular expression to filter out metadata values
}

func (im IgnoredMetadata) String() string {
	objectType := "<nil>"
	if im.ObjectType != nil {
		objectType = *im.ObjectType
	}

	objectName := "<nil>"
	if im.ObjectName != nil {
		objectName = *im.ObjectName
	}

	return fmt.Sprintf("IgnoredMetadata(ObjectType=%v, ObjectName=%v, KeyRegex=%v, ValueRegex=%v)", objectType, objectName, im.KeyRegex, im.ValueRegex)
}

// filterMetadata filters all metadata entries, given a slice of metadata that needs to be ignored. It doesn't
// alter the input metadata, but returns a copy of the filtered metadata.
func filterMetadata(allMetadata *types.Metadata, href, objectName string, metadataToIgnore []IgnoredMetadata) (*types.Metadata, error) {
	if len(metadataToIgnore) == 0 {
		return allMetadata, nil
	}

	result := &types.Metadata{
		XMLName:       allMetadata.XMLName,
		Xmlns:         allMetadata.Xmlns,
		HREF:          allMetadata.HREF,
		Type:          allMetadata.Type,
		Xsi:           allMetadata.Xsi,
		Link:          allMetadata.Link,
		MetadataEntry: nil,
	}

	var filteredMetadata []*types.MetadataEntry
	for _, originalEntry := range allMetadata.MetadataEntry {
		_, err := filterSingleMetadataEntry(originalEntry.Key, href, objectName, &types.MetadataValue{Domain: originalEntry.Domain, TypedValue: originalEntry.TypedValue}, metadataToIgnore)
		if err == nil || !strings.Contains(err.Error(), "ignored") {
			filteredMetadata = append(filteredMetadata, originalEntry)
		}
	}
	result.MetadataEntry = filteredMetadata
	return result, nil
}

// filterSingleMetadataEntry filters a single metadata entry given a slice of metadata that needs to be ignored. It doesn't
// alter the input metadata, but returns a copy of the filtered metadata.
func filterSingleMetadataEntry(key, href, objectName string, metadataEntry *types.MetadataValue, metadataToIgnore []IgnoredMetadata) (*types.MetadataValue, error) {
	if len(metadataToIgnore) == 0 {
		return metadataEntry, nil
	}

	objectType, err := getMetadataObjectTypeFromHref(href)
	if err != nil {
		return nil, err
	}
	for _, entryToIgnore := range metadataToIgnore {
		if entryToIgnore.ObjectType == nil && entryToIgnore.ObjectName == nil && entryToIgnore.KeyRegex == nil && entryToIgnore.ValueRegex == nil {
			continue
		}
		util.Logger.Printf("[DEBUG] Comparing metadata with key '%s' with ignored metadata filter '%s'", key, entryToIgnore)
		// We apply an optimistic approach here to simplify the conditions, so the metadata entry will always be ignored unless the filters
		// tell otherwise, that is, if they are nil (not all of them as per the condition above), if they're empty or if they don't match.
		// All the filtering options (type, name, keyRegex and valueRegex) must compute to true for the metadata to be ignored.
		if (entryToIgnore.ObjectType == nil || strings.TrimSpace(*entryToIgnore.ObjectType) == "" || *entryToIgnore.ObjectType == objectType) &&
			(entryToIgnore.ObjectName == nil || strings.TrimSpace(*entryToIgnore.ObjectName) == "" || strings.TrimSpace(objectName) == "" || *entryToIgnore.ObjectName == objectName) &&
			(entryToIgnore.KeyRegex == nil || entryToIgnore.KeyRegex.MatchString(key)) &&
			(entryToIgnore.ValueRegex == nil || entryToIgnore.ValueRegex.MatchString(metadataEntry.TypedValue.Value)) {
			util.Logger.Printf("[DEBUG] the metadata entry with key '%s' and value '%v' is being ignored", key, metadataEntry.TypedValue.Value)
			return nil, fmt.Errorf("the metadata entry with key '%s' and value '%v' is being ignored", key, metadataEntry.TypedValue.Value)
		}
	}
	return metadataEntry, nil
}

// filterMetadataToDelete filters a metadata entry that is going to be deleted, given a slice of metadata that needs to be ignored.
func filterMetadataToDelete(client *Client, key, href, objectName string, isSystem bool, metadataToIgnore []IgnoredMetadata) error {
	if len(metadataToIgnore) == 0 {
		return nil
	}

	objectType, err := getMetadataObjectTypeFromHref(href)
	if err != nil {
		return err
	}
	for _, entryToIgnore := range metadataToIgnore {
		if entryToIgnore.ObjectType == nil && entryToIgnore.ObjectName == nil && entryToIgnore.KeyRegex == nil && entryToIgnore.ValueRegex == nil {
			continue
		}

		if (entryToIgnore.ObjectType == nil || strings.TrimSpace(*entryToIgnore.ObjectType) == "" || *entryToIgnore.ObjectType == objectType) &&
			(entryToIgnore.ObjectName == nil || strings.TrimSpace(*entryToIgnore.ObjectName) == "" || strings.TrimSpace(objectName) == "" || *entryToIgnore.ObjectName == objectName) &&
			(entryToIgnore.KeyRegex == nil || entryToIgnore.KeyRegex.MatchString(key)) {

			// Entering here means that it is a good candidate to be ignored, but we need to know the metadata value
			// as we may be filtering by value
			ignore := true
			if entryToIgnore.ValueRegex != nil {
				metadataEntry, err := getMetadataByKey(client, href, objectName, key, isSystem)
				if err != nil {
					return err
				}
				if !entryToIgnore.ValueRegex.MatchString(metadataEntry.TypedValue.Value) {
					ignore = false
				}
			}

			if ignore {
				util.Logger.Printf("[DEBUG] can't delete metadata entry %s as it is ignored", key)
				return fmt.Errorf("can't delete metadata entry %s as it is ignored", key)
			}
			return nil
		}
	}
	return nil

}

// getMetadataObjectTypeFromHref returns the type of the object referenced by the input HREF.
// For example, "https://atl1-vcd-static-130-117.eng.vmware.com/api/admin/org/11582a00-16bb-4916-a42f-2d5e453ccf36"
// will return "org".
func getMetadataObjectTypeFromHref(href string) (string, error) {
	splitHref := strings.Split(href, "/")
	if len(splitHref) < 2 {
		return "", fmt.Errorf("could not find any object type in the provided HREF '%s'", href)
	}
	return splitHref[len(splitHref)-2], nil
}
