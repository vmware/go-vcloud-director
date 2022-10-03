/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"strings"
)

// ----------------
// GET metadata
// ----------------

// GetMetadataByHref returns metadata from the given resource reference.
func (vcdClient *VCDClient) GetMetadataByHref(href string) (*types.Metadata, error) {
	return getMetadata(&vcdClient.Client, href)
}

// GetMetadata returns VM metadata.
func (vm *VM) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vm.client, vm.VM.HREF)
}

// GetMetadata returns VDC metadata.
// Note: Requires system administrator privileges.
func (vdc *Vdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vdc.client, getAdminURL(vdc.Vdc.HREF))
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
	return getMetadata(openApiOrgVdcNetwork.client, fmt.Sprintf("%s/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), strings.ReplaceAll(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID, "urn:vcloud:network:", "")))
}

// ----------------
// ADD metadata async
// ----------------

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

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VDC with the given key, value, type and visibility
// and returns the task.
// Note: Requires system administrator privileges.
func (vdc *Vdc) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vdc.client, getAdminURL(vdc.Vdc.HREF), key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given ProviderVdc with the given key, value, type and visibility
// and returns the task.
// Note: Requires system administrator privileges.
func (providerVdc *ProviderVdc) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(providerVdc.client,  providerVdc.ProviderVdc.HREF, key, value, typedValue, visibility, isSystem)
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VApp with the given key, value, type and visibility
// and returns the task.
func (vapp *VApp) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vapp.client, vapp.VApp.HREF, key, value, typedValue, visibility, isSystem )
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given VAppTemplate with the given key, value, type and visibility
// and returns the task.
func (vAppTemplate *VAppTemplate) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF, key, value, typedValue, visibility, isSystem )
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given MediaRecord with the given key, value, type and visibility
// and returns the task.
func (mediaRecord *MediaRecord) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF, key, value, typedValue, visibility, isSystem )
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Media with the given key, value, type and visibility
// and returns the task.
func (media *Media) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(media.client, media.Media.HREF, key, value, typedValue, visibility, isSystem )
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given AdminCatalog with the given key, value, type and visibility
// and returns the task.
func (adminCatalog *AdminCatalog) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF, key, value, typedValue, visibility, isSystem )
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given AdminOrg with the given key, value, type and visibility
// and returns the task.
func (adminOrg *AdminOrg) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(adminOrg.client, adminOrg.AdminOrg.HREF, key, value, typedValue, visibility, isSystem )
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Disk with the given key, value, type and visibility
// and returns the task.
func (disk *Disk) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(disk.client, disk.Disk.HREF, key, value, typedValue, visibility, isSystem )
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given OrgVDCNetwork with the given key, value, type and visibility
// and returns the task.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(orgVdcNetwork.client, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF), key, value, typedValue, visibility, isSystem )
}

// AddMetadataEntryWithVisibilityAsync adds metadata to the given Catalog Item with the given key, value, type and visibility
// and returns the task.
func (catalogItem *CatalogItem) AddMetadataEntryWithVisibilityAsync(key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	return addMetadata(catalogItem.client, catalogItem.CatalogItem.HREF, key, value, typedValue, visibility, isSystem )
}

// ----------------
// ADD metadata
// ----------------

// AddMetadataEntryWithVisibilityByHref adds metadata to the given resource reference with the given key, value, type and visibility
// and waits for completion.
func (vcdClient *VCDClient) AddMetadataEntryWithVisibilityByHref(href, key, value, typedValue, visibility string, isSystem bool) error {
	task, err := vcdClient.AddMetadataEntryWithVisibilityByHrefAsync(href, key, value, typedValue, visibility, isSystem)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// ----------------
// POST metadata async
// ----------------

// MergeMetadataWithVisibilityByHrefAsync updates the metadata entries present in the referenced entity and creates the ones not present, then
// returns the task.
func (vcdClient *VCDClient) MergeMetadataWithVisibilityByHrefAsync(href string, metadata map[string]types.MetadataTypedValue) (Task, error) {
	return mergeAllMetadata(&vcdClient.Client, href, metadata)
}

// ----------------
// POST metadata
// ----------------

// DeleteMetadataEntryByHref deletes metadata from the given resource reference, depending on key provided as input
// and waits for the task to finish.
func (vcdClient *VCDClient) DeleteMetadataEntryByHref(href, key string) error {
	task, err := vcdClient.DeleteMetadataEntryByHrefAsync(href, key)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// ----------------
// DELETE metadata async
// ----------------

// DeleteMetadataEntryByHrefAsync deletes metadata from the given resource reference, depending on key provided as input
// and returns a task.
func (vcdClient *VCDClient) DeleteMetadataEntryByHrefAsync(href, key string) (Task, error) {
	return deleteMetadata(&vcdClient.Client, key, href)
}

// ----------------
// DELETE metadata
// ----------------



// ----------------
// Generic private functions

// Generic function to retrieve metadata from VCD
func getMetadata(client *Client, requestUri string) (*types.Metadata, error) {
	metadata := &types.Metadata{}

	_, err := client.ExecuteRequest(requestUri+"/metadata/", http.MethodGet,
		types.MimeMetaData, "error retrieving metadata: %s", nil, metadata)

	return metadata, err
}

// NOTE: This "v2" is not v2 in terms of API versioning, it's just a new way of handling metadata.
// The idea is that once go-vcloud-director 3.0 is released, one can just remove "v1".

// addMetadata adds metadata to an entity.
// The function supports passing a value that requires a typed value that must be one of:
// types.MetadataStringValue, types.MetadataNumberValue, types.MetadataDateTimeValue and types.MetadataBooleanValue.
// Visibility also needs to be one of: types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility or types.MetadataReadWriteVisibility
func addMetadata(client *Client, requestUri, key, value, typedValue, visibility string, isSystem bool) (Task, error) {
	newMetadata := &types.MetadataValue{
		Xmlns: types.XMLNamespaceVCloud,
		Xsi:   types.XMLNamespaceXSI,
		TypedValue: &types.MetadataTypedValue{
			XsiType: typedValue,
			Value:   value,
		},
		Domain: &types.MetadataDomainTag{
			Visibility: visibility,
		},
	}

	if isSystem {
		newMetadata.Domain.Domain = "SYSTEM"
	}

	apiEndpoint := urlParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeMetaDataValue, "error adding metadata: %s", newMetadata)
}


// mergeAllMetadata merges the metadata key-values provided as parameter with existing entity metadata
func mergeAllMetadata(client *Client, requestUri string, metadata map[string]types.MetadataTypedValue) (Task, error) {
	var metadataToMerge []*types.MetadataEntry
	for key, value := range metadata {
		metadataToMerge = append(metadataToMerge, &types.MetadataEntry{
			Xmlns: types.XMLNamespaceVCloud,
			Xsi:   types.XMLNamespaceXSI,
			Key:   key,
			TypedValue: &value,
		})
	}

	newMetadata := &types.Metadata{
		Xmlns:         types.XMLNamespaceVCloud,
		Xsi:           types.XMLNamespaceXSI,
		MetadataEntry: metadataToMerge,
	}

	apiEndpoint := urlParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata"

	// Return the task
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeMetaData, "error adding metadata: %s", newMetadata)
}

// deleteMetadata Deletes metadata from an entity.
func deleteMetadata(client *Client, requestUri string, key string) (Task, error) {
	apiEndpoint := urlParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodDelete,
		"", "error deleting metadata: %s", nil)
}
