/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetMetadata calls private function getMetadata() with vm.client and vm.VM.HREF
// which returns a *types.Metadata struct for provided VM input.
func (vm *VM) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vm.client, vm.VM.HREF)
}

// DeleteMetadata() function calls private function deleteMetadata() with vm.client and vm.VM.HREF
// which deletes metadata depending on key provided as input from VM.
func (vm *VM) DeleteMetadata(key string) (Task, error) {
	return deleteMetadata(vm.client, key, vm.VM.HREF)
}

// AddMetadata calls private function addMetadata() with vm.client and vm.VM.HREF
// which adds metadata key/value pair provided as input to VM.
func (vm *VM) AddMetadata(key string, value string) (Task, error) {
	return addMetadata(vm.client, types.MetadataStringValue, key, value, vm.VM.HREF)
}

// GetMetadata returns meta data for VDC.
func (vdc *Vdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vdc.client, getAdminVdcURL(vdc.Vdc.HREF))
}

// DeleteMetadata() function deletes metadata by key provided as input
func (vdc *Vdc) DeleteMetadata(key string) (Vdc, error) {
	task, err := deleteMetadata(vdc.client, key, getAdminVdcURL(vdc.Vdc.HREF))
	if err != nil {
		return Vdc{}, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return Vdc{}, err
	}

	err = vdc.Refresh()
	if err != nil {
		return Vdc{}, err
	}

	return *vdc, nil
}

// AddMetadata adds metadata key/value pair provided as input to VDC.
func (vdc *Vdc) AddMetadata(key string, value string) (Vdc, error) {
	task, err := addMetadata(vdc.client, types.MetadataStringValue, key, value, getAdminVdcURL(vdc.Vdc.HREF))
	if err != nil {
		return Vdc{}, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return Vdc{}, err
	}

	err = vdc.Refresh()
	if err != nil {
		return Vdc{}, err
	}

	return *vdc, nil
}

// AddMetadata adds metadata key/value pair provided as input to VDC.
// and returns task
func (vdc *Vdc) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(vdc.client, types.MetadataStringValue, key, value, getAdminVdcURL(vdc.Vdc.HREF))
}

// DeleteMetadata() function deletes metadata by key provided as input
// and returns task
func (vdc *Vdc) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(vdc.client, key, getAdminVdcURL(vdc.Vdc.HREF))
}

func getAdminVdcURL(vdcURL string) string {
	return strings.Split(vdcURL, "/api/vdc/")[0] + "/api/admin/vdc/" + strings.Split(vdcURL, "/api/vdc/")[1]
}

// GetMetadata calls private function getMetadata() with vapp.client and vapp.VApp.HREF
// which returns a *types.Metadata struct for provided vapp input.
func (vapp *VApp) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vapp.client, vapp.VApp.HREF)
}

func getMetadata(client *Client, requestUri string) (*types.Metadata, error) {
	metadata := &types.Metadata{}

	_, err := client.ExecuteRequest(requestUri+"/metadata/", http.MethodGet,
		types.MimeMetaData, "error retrieving metadata: %s", nil, metadata)

	return metadata, err
}

// DeleteMetadata() function calls private function deleteMetadata() with vapp.client and vapp.VApp.HREF
// which deletes metadata depending on key provided as input from vApp.
func (vapp *VApp) DeleteMetadata(key string) (Task, error) {
	return deleteMetadata(vapp.client, key, vapp.VApp.HREF)
}

// Deletes metadata (type MetadataStringValue) from the vApp
// TODO: Support all MetadataTypedValue types with this function
func deleteMetadata(client *Client, key string, requestUri string) (Task, error) {
	apiEndpoint := urlParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodDelete,
		"", "error deleting metadata: %s", nil)
}

// AddMetadata calls private function addMetadata() with vapp.client and vapp.VApp.HREF
// which adds metadata key/value pair provided as input
func (vapp *VApp) AddMetadata(key string, value string) (Task, error) {
	return addMetadata(vapp.client, types.MetadataStringValue, key, value, vapp.VApp.HREF)
}

// Adds metadata (type MetadataStringValue) to the vApp
// The function supports passing a typedValue. Use one of the consts defined.
// Only tested with MetadataStringValue and MetadataNumberValue.
// TODO: We might also need to add support to MetadataDateTimeValue and MetadataBooleanValue
func addMetadata(client *Client, typedValue, key, value, requestUri string) (Task, error) {
	newMetadata := &types.MetadataValue{
		Xmlns: types.XMLNamespaceVCloud,
		Xsi:   types.XMLNamespaceXSI,
		TypedValue: &types.TypedValue{
			XsiType: typedValue,
			Value:   value,
		},
	}

	apiEndpoint := urlParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeMetaDataValue, "error adding metadata: %s", newMetadata)
}

// GetMetadata calls private function getMetadata() with catalogItem.client and catalogItem.CatalogItem.HREF
// which returns a *types.Metadata struct for provided catalog item input.
func (vAppTemplate *VAppTemplate) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF)
}

// AddMetadata adds metadata key/value pair provided as input and returned update VAppTemplate
func (vAppTemplate *VAppTemplate) AddMetadata(key string, value string) (*VAppTemplate, error) {
	task, err := vAppTemplate.AddMetadataAsync(key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for vApp template task: %s", err)
	}

	err = vAppTemplate.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing vApp template: %s", err)
	}

	return vAppTemplate, nil
}

// AddMetadataAsync calls private function addMetadata() with vAppTemplate.client and vAppTemplate.VAppTemplate.HREF
// which adds metadata key/value pair provided as input.
func (vAppTemplate *VAppTemplate) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(vAppTemplate.client, types.MetadataStringValue, key, value, vAppTemplate.VAppTemplate.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
func (vAppTemplate *VAppTemplate) DeleteMetadata(key string) error {
	task, err := vAppTemplate.DeleteMetadataAsync(key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error completing delete metadata for vApp template task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with vAppTemplate.client and vAppTemplate.VAppTemplate.HREF
// which deletes metadata depending on key provided as input from catalog item.
func (vAppTemplate *VAppTemplate) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(vAppTemplate.client, key, vAppTemplate.VAppTemplate.HREF)
}

// GetMetadata calls private function getMetadata() with mediaItem.client and mediaItem.MediaItem.HREF
// which returns a *types.Metadata struct for provided media item input.
// Deprecated: Use MediaRecord.GetMetadata
func (mediaItem *MediaItem) GetMetadata() (*types.Metadata, error) {
	return getMetadata(mediaItem.vdc.client, mediaItem.MediaItem.HREF)
}

// AddMetadata adds metadata key/value pair provided as input.
// Deprecated: Use MediaRecord.AddMetadata
func (mediaItem *MediaItem) AddMetadata(key string, value string) (*MediaItem, error) {
	task, err := mediaItem.AddMetadataAsync(key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for media item task: %s", err)
	}

	err = mediaItem.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing media item: %s", err)
	}

	return mediaItem, nil
}

// AddMetadataAsync calls private function addMetadata() with mediaItem.client and mediaItem.MediaItem.HREF
// which adds metadata key/value pair provided as input.
// Deprecated: Use MediaRecord.AddMetadataAsync
func (mediaItem *MediaItem) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(mediaItem.vdc.client, types.MetadataStringValue, key, value, mediaItem.MediaItem.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
// Deprecated: Use MediaRecord.DeleteMetadata
func (mediaItem *MediaItem) DeleteMetadata(key string) error {
	task, err := mediaItem.DeleteMetadataAsync(key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error completing delete metadata for media item task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with mediaItem.client and mediaItem.MediaItem.HREF
// which deletes metadata depending on key provided as input from media item.
// Deprecated: Use MediaRecord.DeleteMetadataAsync
func (mediaItem *MediaItem) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(mediaItem.vdc.client, key, mediaItem.MediaItem.HREF)
}

// GetMetadata calls private function getMetadata() with MediaRecord.client and MediaRecord.MediaRecord.HREF
// which returns a *types.Metadata struct for provided media item input.
func (mediaRecord *MediaRecord) GetMetadata() (*types.Metadata, error) {
	return getMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF)
}

// AddMetadata adds metadata key/value pair provided as input.
func (mediaRecord *MediaRecord) AddMetadata(key string, value string) (*MediaRecord, error) {
	task, err := mediaRecord.AddMetadataAsync(key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for media item task: %s", err)
	}

	err = mediaRecord.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing media item: %s", err)
	}

	return mediaRecord, nil
}

// AddMetadataAsync calls private function addMetadata() with MediaRecord.client and MediaRecord.MediaRecord.HREF
// which adds metadata key/value pair provided as input.
func (mediaRecord *MediaRecord) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(mediaRecord.client, types.MetadataStringValue, key, value, mediaRecord.MediaRecord.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
func (mediaRecord *MediaRecord) DeleteMetadata(key string) error {
	task, err := mediaRecord.DeleteMetadataAsync(key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error completing delete metadata for media item task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with MediaRecord.client and MediaRecord.MediaRecord.HREF
// which deletes metadata depending on key provided as input from media item.
func (mediaRecord *MediaRecord) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(mediaRecord.client, key, mediaRecord.MediaRecord.HREF)
}

// GetMetadata calls private function getMetadata() with Media.client and Media.Media.HREF
// which returns a *types.Metadata struct for provided media item input.
func (media *Media) GetMetadata() (*types.Metadata, error) {
	return getMetadata(media.client, media.Media.HREF)
}

// AddMetadata adds metadata key/value pair provided as input.
func (media *Media) AddMetadata(key string, value string) (*Media, error) {
	task, err := media.AddMetadataAsync(key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for media item task: %s", err)
	}

	err = media.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing media item: %s", err)
	}

	return media, nil
}

// AddMetadataAsync calls private function addMetadata() with Media.client and Media.Media.HREF
// which adds metadata key/value pair provided as input.
func (media *Media) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(media.client, types.MetadataStringValue, key, value, media.Media.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
func (media *Media) DeleteMetadata(key string) error {
	task, err := media.DeleteMetadataAsync(key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error completing delete metadata for media item task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with Media.client and Media.Media.HREF
// which deletes metadata depending on key provided as input from media item.
func (media *Media) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(media.client, key, media.Media.HREF)
}

// GetMetadata calls private function getMetadata() with AdminCatalog.client and AdminCatalog.AdminCatalog.HREF
// which returns a *types.Metadata struct for provided adminCatalog item input.
func (adminCatalog *AdminCatalog) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF)
}

// AddMetadata adds metadata typedValue and key/value pair provided as input.
func (adminCatalog *AdminCatalog) AddMetadata(typedValue, key, value string) (*AdminCatalog, error) {
	task, err := adminCatalog.AddMetadataAsync(typedValue, key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for media item task: %s", err)
	}

	err = adminCatalog.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing media item: %s", err)
	}

	return adminCatalog, nil
}

// AddMetadataAsync calls private function addMetadata() with AdminCatalog.client and AdminCatalog.AdminCatalog.HREF
// which adds metadata typedvalue as well as its key/value pair provided as input.
func (catalog *AdminCatalog) AddMetadataAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(catalog.client, typedValue, key, value, catalog.AdminCatalog.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
func (adminCatalog *AdminCatalog) DeleteMetadata(key string) error {
	task, err := adminCatalog.DeleteMetadataAsync(key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error completing delete metadata for media item task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with AdminCatalog.client and AdminCatalog.AdminCatalog.HREF
// which deletes metadata depending on key provided as input from adminCatalog item.
func (adminCatalog *AdminCatalog) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(adminCatalog.client, key, adminCatalog.AdminCatalog.HREF)
}

// GetMetadata calls private function getMetadata() with Catalog.client and Catalog.Catalog.HREF
// which returns a *types.Metadata struct for provided catalog item input.
func (catalog *Catalog) GetMetadata() (*types.Metadata, error) {
	return getMetadata(catalog.client, catalog.Catalog.HREF)
}
