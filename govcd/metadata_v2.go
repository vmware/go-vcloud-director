/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
)

// NOTE: This "v2" is not v2 in terms of API versioning, it's just a new way of handling metadata.
// The idea is that once go-vcloud-director 3.0 is released, one can just remove "v1".

// ------------------------------------------------------------------------------------------------
// GET metadata
// ------------------------------------------------------------------------------------------------

// GetMetadataByHref returns metadata from the given resource reference.
func (vcdClient *VCDClient) GetMetadataByHref(href string) (*types.Metadata, error) {
	return getMetadata(&vcdClient.Client, href)
}

// GetMetadata returns VM metadata.
func (vm *VM) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vm.client, vm.VM.HREF)
}

// GetMetadata returns VDC metadata.
func (vdc *Vdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vdc.client, vdc.Vdc.HREF)
}

// GetMetadata returns AdminVdc metadata.
func (adminVdc *AdminVdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminVdc.client, adminVdc.AdminVdc.HREF)
}

// GetMetadata returns ProviderVdc metadata.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(providerVdc.client, providerVdc.ProviderVdc.HREF)
}

// GetMetadata returns VApp metadata.
func (vapp *VApp) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vapp.client, vapp.VApp.HREF)
}

// GetMetadata returns VAppTemplate metadata.
func (vAppTemplate *VAppTemplate) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF)
}

// GetMetadata returns MediaRecord metadata.
func (mediaRecord *MediaRecord) GetMetadata() (*types.Metadata, error) {
	return getMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF)
}

// GetMetadata returns Media metadata.
func (media *Media) GetMetadata() (*types.Metadata, error) {
	return getMetadata(media.client, media.Media.HREF)
}

// GetMetadata returns Catalog metadata.
func (catalog *Catalog) GetMetadata() (*types.Metadata, error) {
	return getMetadata(catalog.client, catalog.Catalog.HREF)
}

// GetMetadata returns AdminCatalog metadata.
func (adminCatalog *AdminCatalog) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF)
}

// GetMetadata returns the Org metadata of the corresponding organization seen as administrator
func (org *Org) GetMetadata() (*types.Metadata, error) {
	return getMetadata(org.client, org.Org.HREF)
}

// GetMetadata returns the AdminOrg metadata of the corresponding organization seen as administrator
func (adminOrg *AdminOrg) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminOrg.client, adminOrg.AdminOrg.HREF)
}

// GetMetadata returns the metadata of the corresponding independent disk
func (disk *Disk) GetMetadata() (*types.Metadata, error) {
	return getMetadata(disk.client, disk.Disk.HREF)
}

// GetMetadata returns OrgVDCNetwork metadata.
func (orgVdcNetwork *OrgVDCNetwork) GetMetadata() (*types.Metadata, error) {
	return getMetadata(orgVdcNetwork.client, orgVdcNetwork.OrgVDCNetwork.HREF)
}

// GetMetadata returns CatalogItem metadata.
func (catalogItem *CatalogItem) GetMetadata() (*types.Metadata, error) {
	return getMetadata(catalogItem.client, catalogItem.CatalogItem.HREF)
}

// GetMetadata returns OpenApiOrgVdcNetwork metadata.
// NOTE: This function cannot retrieve metadata if the network belongs to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) GetMetadata() (*types.Metadata, error) {
	href := fmt.Sprintf("%s/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	return getMetadata(openApiOrgVdcNetwork.client, href)
}

// ------------------------------------------------------------------------------------------------
// ADD metadata async
// ------------------------------------------------------------------------------------------------

// AddMetadataEntryWithVisibilityByHrefAsync adds metadata to the given resource reference with the given key, value, type and visibility
// and returns the task.
func (vcdClient *VCDClient) AddMetadataEntryWithVisibilityByHrefAsync(href, key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(&vcdClient.Client, href, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VM with the given key, value, type and visibility
//// and returns the task.
func (vm *VM) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vm.client, vm.VM.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given AdminVdc with the given key, value, type and visibility
// and returns the task.
func (adminVdc *AdminVdc) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(adminVdc.client, adminVdc.AdminVdc.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given ProviderVdc with the given key, value, type and visibility
// and returns the task.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(providerVdc.client, providerVdc.ProviderVdc.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VApp with the given key, value, type and visibility
// and returns the task.
func (vapp *VApp) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vapp.client, vapp.VApp.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VAppTemplate with the given key, value, type and visibility
// and returns the task.
func (vAppTemplate *VAppTemplate) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given MediaRecord with the given key, value, type and visibility
// and returns the task.
func (mediaRecord *MediaRecord) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Media with the given key, value, type and visibility
// and returns the task.
func (media *Media) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(media.client, media.Media.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given AdminCatalog with the given key, value, type and visibility
// and returns the task.
func (adminCatalog *AdminCatalog) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given AdminOrg with the given key, value, type and visibility
// and returns the task.
func (adminOrg *AdminOrg) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(adminOrg.client, adminOrg.AdminOrg.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Disk with the given key, value, type and visibility
// and returns the task.
func (disk *Disk) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(disk.client, disk.Disk.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given OrgVDCNetwork with the given key, value, type and visibility
// and returns the task.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Catalog Item with the given key, value, type and visibility
// and returns the task.
func (catalogItem *CatalogItem) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(catalogItem.client, catalogItem.CatalogItem.HREF, key, value, typedValue, visibility, isSystem)
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
	return addMetadataAndWait(vm.client, vm.VM.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver AdminVdc and waits for the task to finish.
func (adminVdc *AdminVdc) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(adminVdc.client, adminVdc.AdminVdc.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver ProviderVdc and waits for the task to finish.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(providerVdc.client, providerVdc.ProviderVdc.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver VApp and waits for the task to finish.
func (vapp *VApp) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(vapp.client, vapp.VApp.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver VAppTemplate and waits for the task to finish.
func (vAppTemplate *VAppTemplate) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver MediaRecord and waits for the task to finish.
func (mediaRecord *MediaRecord) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(mediaRecord.client, mediaRecord.MediaRecord.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver Media and waits for the task to finish.
func (media *Media) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(media.client, media.Media.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver AdminCatalog and waits for the task to finish.
func (adminCatalog *AdminCatalog) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(adminCatalog.client, adminCatalog.AdminCatalog.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver AdminOrg and waits for the task to finish.
func (adminOrg *AdminOrg) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(adminOrg.client, adminOrg.AdminOrg.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver Disk and waits for the task to finish.
func (disk *Disk) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(disk.client, disk.Disk.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver OrgVDCNetwork and waits for the task to finish.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver CatalogItem and waits for the task to finish.
func (catalogItem *CatalogItem) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	return addMetadataAndWait(catalogItem.client, catalogItem.CatalogItem.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibility adds metadata to the receiver OpenApiOrgVdcNetwork and waits for the task to finish.
// Note: It doesn't add metadata to networks that belong to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error {
	href := fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	task, err := addMetadata(openApiOrgVdcNetwork.client, href, key, value, typedValue, visibility, isSystem)
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
	return mergeAllMetadata(&vcdClient.Client, href, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges VM metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then returns the task.
func (vm *VM) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(vm.client, vm.VM.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges AdminVdc metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminVdc *AdminVdc) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(adminVdc.client, adminVdc.AdminVdc.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges Provider VDC metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(providerVdc.client, providerVdc.ProviderVdc.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges VApp metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vapp *VApp) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(vapp.client, vapp.VApp.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges VAppTemplate metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vAppTemplate *VAppTemplate) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges MediaRecord metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (mediaRecord *MediaRecord) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges Media metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (media *Media) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(media.client, media.Media.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges AdminCatalog metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminCatalog *AdminCatalog) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges AdminOrg metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminOrg *AdminOrg) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(adminOrg.client, adminOrg.AdminOrg.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges Disk metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (disk *Disk) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(disk.client, disk.Disk.HREF, metadata)
}

// MergeMetadataWithMetadataValuesAsync merges OrgVDCNetwork metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), metadata)
}

// MergeMetadataWithMetadataValuesAsync merges CatalogItem metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (catalogItem *CatalogItem) MergeMetadataWithMetadataValuesAsync(metadata map[string]types.MetadataValue) (Task, error) {
	return mergeAllMetadata(catalogItem.client, catalogItem.CatalogItem.HREF, metadata)
}

// ------------------------------------------------------------------------------------------------
// MERGE metadata
// ------------------------------------------------------------------------------------------------

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver VM and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (vm *VM) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(vm.client, vm.VM.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver AdminVdc and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (adminVdc *AdminVdc) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(adminVdc.client, adminVdc.AdminVdc.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver ProviderVdc and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(providerVdc.client, providerVdc.ProviderVdc.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver VApp and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (vApp *VApp) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(vApp.client, vApp.VApp.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver VAppTemplate and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (vAppTemplate *VAppTemplate) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver MediaRecord and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (mediaRecord *MediaRecord) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(mediaRecord.client, mediaRecord.MediaRecord.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver Media and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (media *Media) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(media.client, media.Media.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver AdminCatalog and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (adminCatalog *AdminCatalog) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(adminCatalog.client, adminCatalog.AdminCatalog.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver AdminOrg and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (adminOrg *AdminOrg) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(adminOrg.client, adminOrg.AdminOrg.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver Disk and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (disk *Disk) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(disk.client, disk.Disk.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver OrgVDCNetwork and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver CatalogItem and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func (catalogItem *CatalogItem) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	return mergeMetadataAndWait(catalogItem.client, catalogItem.CatalogItem.HREF, metadata)
}

// MergeMetadataWithMetadataValues updates the metadata values that are already present in the receiver OpenApiOrgVdcNetwork and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
// Note: It doesn't merge metadata to networks that belong to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error {
	href := fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	task, err := mergeAllMetadata(openApiOrgVdcNetwork.client, href, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// ------------------------------------------------------------------------------------------------
// DELETE metadata async
// ------------------------------------------------------------------------------------------------

// DeleteMetadataEntryByHrefAsync deletes metadata from the given resource reference, depending on key provided as input
// and returns a task.
func (vcdClient *VCDClient) DeleteMetadataEntryByHrefAsync(href, key string) (Task, error) {
	return deleteMetadata(&vcdClient.Client, href, key)
}

// DeleteMetadataEntryAsync deletes VM metadata associated to the input key and returns the task.
func (vm *VM) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(vm.client, vm.VM.HREF, key)
}

// DeleteMetadataEntryAsync deletes AdminVdc metadata associated to the input key and returns the task.
func (adminVdc *AdminVdc) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(adminVdc.client, adminVdc.AdminVdc.HREF, key)
}

// DeleteMetadataEntryAsync deletes ProviderVdc metadata associated to the input key and returns the task.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(providerVdc.client, providerVdc.ProviderVdc.HREF, key)
}

// DeleteMetadataEntryAsync deletes VApp metadata associated to the input key and returns the task.
func (vapp *VApp) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(vapp.client, vapp.VApp.HREF, key)
}

// DeleteMetadataEntryAsync deletes VAppTemplate metadata associated to the input key and returns the task.
func (vAppTemplate *VAppTemplate) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, key)
}

// DeleteMetadataEntryAsync deletes MediaRecord metadata associated to the input key and returns the task.
func (mediaRecord *MediaRecord) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF, key)
}

// DeleteMetadataEntryAsync deletes Media metadata associated to the input key and returns the task.
func (media *Media) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(media.client, media.Media.HREF, key)
}

// DeleteMetadataEntryAsync deletes AdminCatalog metadata associated to the input key and returns the task.
func (adminCatalog *AdminCatalog) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF, key)
}

// DeleteMetadataEntryAsync deletes AdminOrg metadata associated to the input key and returns the task.
func (adminOrg *AdminOrg) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(adminOrg.client, adminOrg.AdminOrg.HREF, key)
}

// DeleteMetadataEntryAsync deletes Disk metadata associated to the input key and returns the task.
func (disk *Disk) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(disk.client, disk.Disk.HREF, key)
}

// DeleteMetadataEntryAsync deletes OrgVDCNetwork metadata associated to the input key and returns the task.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), key)
}

// DeleteMetadataEntryAsync deletes CatalogItem metadata associated to the input key and returns the task.
func (catalogItem *CatalogItem) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(catalogItem.client, catalogItem.CatalogItem.HREF, key)
}

// ------------------------------------------------------------------------------------------------
// DELETE metadata
// ------------------------------------------------------------------------------------------------

// DeleteMetadataEntryByHref deletes metadata from the given resource reference, depending on key provided as input
// and waits for the task to finish.
func (vcdClient *VCDClient) DeleteMetadataEntryByHref(href, key string) error {
	task, err := vcdClient.DeleteMetadataEntryByHrefAsync(href, key)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes VM metadata associated to the input key and waits for the task to finish.
func (vm *VM) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(vm.client, vm.VM.HREF, key)
}

// DeleteMetadataEntry deletes AdminVdc metadata associated to the input key and waits for the task to finish.
// Note: Requires system administrator privileges.
func (adminVdc *AdminVdc) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(adminVdc.client, getAdminURL(adminVdc.AdminVdc.HREF), key)
}

// DeleteMetadataEntry deletes ProviderVdc metadata associated to the input key and waits for the task to finish.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(providerVdc.client, providerVdc.ProviderVdc.HREF, key)
}

// DeleteMetadataEntry deletes VApp metadata associated to the input key and waits for the task to finish.
func (vApp *VApp) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(vApp.client, vApp.VApp.HREF, key)
}

// DeleteMetadataEntry deletes VAppTemplate metadata associated to the input key and waits for the task to finish.
func (vAppTemplate *VAppTemplate) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, key)
}

// DeleteMetadataEntry deletes MediaRecord metadata associated to the input key and waits for the task to finish.
func (mediaRecord *MediaRecord) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(mediaRecord.client, mediaRecord.MediaRecord.HREF, key)
}

// DeleteMetadataEntry deletes Media metadata associated to the input key and waits for the task to finish.
func (media *Media) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(media.client, media.Media.HREF, key)
}

// DeleteMetadataEntry deletes AdminCatalog metadata associated to the input key and waits for the task to finish.
func (adminCatalog *AdminCatalog) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(adminCatalog.client, adminCatalog.AdminCatalog.HREF, key)
}

// DeleteMetadataEntry deletes AdminOrg metadata associated to the input key and waits for the task to finish.
func (adminOrg *AdminOrg) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(adminOrg.client, adminOrg.AdminOrg.HREF, key)
}

// DeleteMetadataEntry deletes Disk metadata associated to the input key and waits for the task to finish.
func (disk *Disk) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(disk.client, disk.Disk.HREF, key)
}

// DeleteMetadataEntry deletes OrgVDCNetwork metadata associated to the input key and waits for the task to finish.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), key)
}

// DeleteMetadataEntry deletes CatalogItem metadata associated to the input key and waits for the task to finish.
func (catalogItem *CatalogItem) DeleteMetadataEntry(key string) error {
	return deleteMetadataAndWait(catalogItem.client, catalogItem.CatalogItem.HREF, key)
}

// DeleteMetadataEntry deletes OpenApiOrgVdcNetwork metadata associated to the input key and waits for the task to finish.
// Note: It doesn't delete metadata from networks that belong to a VDC Group.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is still not supported.
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) DeleteMetadataEntry(key string) error {
	href := fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), extractUuid(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID))
	task, err := deleteMetadata(openApiOrgVdcNetwork.client, href, key)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// ------------------------------------------------------------------------------------------------
// Generic private functions
// ------------------------------------------------------------------------------------------------

// getMetadata is a generic function to retrieve metadata from VCD
func getMetadata(client *Client, requestUri string) (*types.Metadata, error) {
	metadata := &types.Metadata{}

	_, err := client.ExecuteRequest(requestUri+"/metadata/", http.MethodGet, types.MimeMetaData, "error retrieving metadata: %s", nil, metadata)
	return metadata, err
}

// addMetadata adds metadata to an entity.
// If the metadata entry is of the SYSTEM domain (isSystem=true), one can set different types of Visibility:
// types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility but NOT types.MetadataReadWriteVisibility.
// If the metadata entry is of the GENERAL domain (isSystem=false), visibility is always types.MetadataReadWriteVisibility.
// In terms of typedValues, that must be one of:
// types.MetadataStringValue, types.MetadataNumberValue, types.MetadataDateTimeValue and types.MetadataBooleanValue.
func addMetadata(client *Client, requestUri, key, value, typedValue, visibility string, isSystem bool) (Task, error) {

	if isSystem && visibility == types.MetadataReadWriteVisibility {
		return Task{}, fmt.Errorf("visibility cannot be %s when domain is SYSTEM", types.MetadataReadWriteVisibility)
	}

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

	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut, types.MimeMetaDataValue, "error adding metadata: %s", newMetadata)
}

// addMetadataAndWait adds metadata to an entity and waits for the task completion.
// The function supports passing a value that requires a typed value that must be one of:
// types.MetadataStringValue, types.MetadataNumberValue, types.MetadataDateTimeValue and types.MetadataBooleanValue.
// Visibility also needs to be one of: types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility or types.MetadataReadWriteVisibility
func addMetadataAndWait(client *Client, requestUri, key, value, typedValue, visibility string, isSystem bool) error {
	task, err := addMetadata(client, requestUri, key, value, typedValue, visibility, isSystem)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// mergeAllMetadata updates the metadata values that are already present in VCD and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// If the operation is successful, it returns the created task.
func mergeAllMetadata(client *Client, requestUri string, metadata map[string]types.MetadataValue) (Task, error) {
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

	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost, types.MimeMetaData, "error adding metadata: %s", newMetadata)
}

// mergeAllMetadata updates the metadata values that are already present in VCD and creates the ones not present.
// The input metadata map has a "metadata key"->"metadata value" relation.
// This function waits until merge finishes.
func mergeMetadataAndWait(client *Client, requestUri string, metadata map[string]types.MetadataValue) error {
	task, err := mergeAllMetadata(client, requestUri, metadata)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// deleteMetadata deletes metadata associated to the input key from an entity referenced by its URI, then returns the
// task.
func deleteMetadata(client *Client, requestUri string, key string) (Task, error) {
	apiEndpoint := urlParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodDelete, "", "error deleting metadata: %s", nil)
}

// deleteMetadata deletes metadata associated to the input key from an entity referenced by its URI.
func deleteMetadataAndWait(client *Client, requestUri string, key string) error {
	task, err := deleteMetadata(client, requestUri, key)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}