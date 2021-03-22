/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetMetadata calls private function getMetadata() with vm.client and vm.VM.HREF
// which returns a *types.Metadata struct for provided VM input.
func (vm *VM) GetMetadata(ctx context.Context) (*types.Metadata, error) {
	return getMetadata(ctx, vm.client, vm.VM.HREF)
}

// DeleteMetadata() function calls private function deleteMetadata() with vm.client and vm.VM.HREF
// which deletes metadata depending on key provided as input from VM.
func (vm *VM) DeleteMetadata(ctx context.Context, key string) (Task, error) {
	return deleteMetadata(ctx, vm.client, key, vm.VM.HREF)
}

// AddMetadata calls private function addMetadata() with vm.client and vm.VM.HREF
// which adds metadata key/value pair provided as input to VM.
func (vm *VM) AddMetadata(ctx context.Context, key string, value string) (Task, error) {
	return addMetadata(ctx, vm.client, key, value, vm.VM.HREF)
}

// GetMetadata returns meta data for VDC.
func (vdc *Vdc) GetMetadata(ctx context.Context) (*types.Metadata, error) {
	return getMetadata(ctx, vdc.client, getAdminVdcURL(vdc.Vdc.HREF))
}

// DeleteMetadata() function deletes metadata by key provided as input
func (vdc *Vdc) DeleteMetadata(ctx context.Context, key string) (Vdc, error) {
	task, err := deleteMetadata(ctx, vdc.client, key, getAdminVdcURL(vdc.Vdc.HREF))
	if err != nil {
		return Vdc{}, err
	}

	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return Vdc{}, err
	}

	err = vdc.Refresh(ctx)
	if err != nil {
		return Vdc{}, err
	}

	return *vdc, nil
}

// AddMetadata adds metadata key/value pair provided as input to VDC.
func (vdc *Vdc) AddMetadata(ctx context.Context, key string, value string) (Vdc, error) {
	task, err := addMetadata(ctx, vdc.client, key, value, getAdminVdcURL(vdc.Vdc.HREF))
	if err != nil {
		return Vdc{}, err
	}

	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return Vdc{}, err
	}

	err = vdc.Refresh(ctx)
	if err != nil {
		return Vdc{}, err
	}

	return *vdc, nil
}

// AddMetadata adds metadata key/value pair provided as input to VDC.
// and returns task
func (vdc *Vdc) AddMetadataAsync(ctx context.Context, key string, value string) (Task, error) {
	return addMetadata(ctx, vdc.client, key, value, getAdminVdcURL(vdc.Vdc.HREF))
}

// DeleteMetadata() function deletes metadata by key provided as input
// and returns task
func (vdc *Vdc) DeleteMetadataAsync(ctx context.Context, key string) (Task, error) {
	return deleteMetadata(ctx, vdc.client, key, getAdminVdcURL(vdc.Vdc.HREF))
}

func getAdminVdcURL(vdcURL string) string {
	return strings.Split(vdcURL, "/api/vdc/")[0] + "/api/admin/vdc/" + strings.Split(vdcURL, "/api/vdc/")[1]
}

// GetMetadata calls private function getMetadata() with vapp.client and vapp.VApp.HREF
// which returns a *types.Metadata struct for provided vapp input.
func (vapp *VApp) GetMetadata(ctx context.Context) (*types.Metadata, error) {
	return getMetadata(ctx, vapp.client, vapp.VApp.HREF)
}

func getMetadata(ctx context.Context, client *Client, requestUri string) (*types.Metadata, error) {
	metadata := &types.Metadata{}

	_, err := client.ExecuteRequest(ctx, requestUri+"/metadata/", http.MethodGet,
		types.MimeMetaData, "error retrieving metadata: %s", nil, metadata)

	return metadata, err
}

// DeleteMetadata() function calls private function deleteMetadata() with vapp.client and vapp.VApp.HREF
// which deletes metadata depending on key provided as input from vApp.
func (vapp *VApp) DeleteMetadata(ctx context.Context, key string) (Task, error) {
	return deleteMetadata(ctx, vapp.client, key, vapp.VApp.HREF)
}

// Deletes metadata (type MetadataStringValue) from the vApp
// TODO: Support all MetadataTypedValue types with this function
func deleteMetadata(ctx context.Context, client *Client, key string, requestUri string) (Task, error) {
	apiEndpoint, _ := url.ParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(ctx, apiEndpoint.String(), http.MethodDelete,
		"", "error deleting metadata: %s", nil)
}

// AddMetadata calls private function addMetadata() with vapp.client and vapp.VApp.HREF
// which adds metadata key/value pair provided as input
func (vapp *VApp) AddMetadata(ctx context.Context, key string, value string) (Task, error) {
	return addMetadata(ctx, vapp.client, key, value, vapp.VApp.HREF)
}

// Adds metadata (type MetadataStringValue) to the vApp
// TODO: Support all MetadataTypedValue types with this function
func addMetadata(ctx context.Context, client *Client, key string, value string, requestUri string) (Task, error) {
	newMetadata := &types.MetadataValue{
		Xmlns: types.XMLNamespaceVCloud,
		Xsi:   types.XMLNamespaceXSI,
		TypedValue: &types.TypedValue{
			XsiType: "MetadataStringValue",
			Value:   value,
		},
	}

	apiEndpoint, _ := url.ParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(ctx, apiEndpoint.String(), http.MethodPut,
		types.MimeMetaDataValue, "error adding metadata: %s", newMetadata)
}

// GetMetadata calls private function getMetadata() with catalogItem.client and catalogItem.CatalogItem.HREF
// which returns a *types.Metadata struct for provided catalog item input.
func (vAppTemplate *VAppTemplate) GetMetadata(ctx context.Context) (*types.Metadata, error) {
	return getMetadata(ctx, vAppTemplate.client, vAppTemplate.VAppTemplate.HREF)
}

// AddMetadata adds metadata key/value pair provided as input and returned update VAppTemplate
func (vAppTemplate *VAppTemplate) AddMetadata(ctx context.Context, key string, value string) (*VAppTemplate, error) {
	task, err := vAppTemplate.AddMetadataAsync(ctx, key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for vApp template task: %s", err)
	}

	err = vAppTemplate.Refresh(ctx)
	if err != nil {
		return nil, fmt.Errorf("error refreshing vApp template: %s", err)
	}

	return vAppTemplate, nil
}

// AddMetadataAsync calls private function addMetadata() with vAppTemplate.client and vAppTemplate.VAppTemplate.HREF
// which adds metadata key/value pair provided as input.
func (vAppTemplate *VAppTemplate) AddMetadataAsync(ctx context.Context, key string, value string) (Task, error) {
	return addMetadata(ctx, vAppTemplate.client, key, value, vAppTemplate.VAppTemplate.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
func (vAppTemplate *VAppTemplate) DeleteMetadata(ctx context.Context, key string) error {
	task, err := vAppTemplate.DeleteMetadataAsync(ctx, key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return fmt.Errorf("error completing delete metadata for vApp template task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with vAppTemplate.client and vAppTemplate.VAppTemplate.HREF
// which deletes metadata depending on key provided as input from catalog item.
func (vAppTemplate *VAppTemplate) DeleteMetadataAsync(ctx context.Context, key string) (Task, error) {
	return deleteMetadata(ctx, vAppTemplate.client, key, vAppTemplate.VAppTemplate.HREF)
}

// GetMetadata calls private function getMetadata() with mediaItem.client and mediaItem.MediaItem.HREF
// which returns a *types.Metadata struct for provided media item input.
// Deprecated: Use MediaRecord.GetMetadata
func (mediaItem *MediaItem) GetMetadata(ctx context.Context) (*types.Metadata, error) {
	return getMetadata(ctx, mediaItem.vdc.client, mediaItem.MediaItem.HREF)
}

// AddMetadata adds metadata key/value pair provided as input.
// Deprecated: Use MediaRecord.AddMetadata
func (mediaItem *MediaItem) AddMetadata(ctx context.Context, key string, value string) (*MediaItem, error) {
	task, err := mediaItem.AddMetadataAsync(ctx, key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for media item task: %s", err)
	}

	err = mediaItem.Refresh(ctx)
	if err != nil {
		return nil, fmt.Errorf("error refreshing media item: %s", err)
	}

	return mediaItem, nil
}

// AddMetadataAsync calls private function addMetadata() with mediaItem.client and mediaItem.MediaItem.HREF
// which adds metadata key/value pair provided as input.
// Deprecated: Use MediaRecord.AddMetadataAsync
func (mediaItem *MediaItem) AddMetadataAsync(ctx context.Context, key string, value string) (Task, error) {
	return addMetadata(ctx, mediaItem.vdc.client, key, value, mediaItem.MediaItem.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
// Deprecated: Use MediaRecord.DeleteMetadata
func (mediaItem *MediaItem) DeleteMetadata(ctx context.Context, key string) error {
	task, err := mediaItem.DeleteMetadataAsync(ctx, key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return fmt.Errorf("error completing delete metadata for media item task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with mediaItem.client and mediaItem.MediaItem.HREF
// which deletes metadata depending on key provided as input from media item.
// Deprecated: Use MediaRecord.DeleteMetadataAsync
func (mediaItem *MediaItem) DeleteMetadataAsync(ctx context.Context, key string) (Task, error) {
	return deleteMetadata(ctx, mediaItem.vdc.client, key, mediaItem.MediaItem.HREF)
}

// GetMetadata calls private function getMetadata() with MediaRecord.client and MediaRecord.MediaRecord.HREF
// which returns a *types.Metadata struct for provided media item input.
func (mediaRecord *MediaRecord) GetMetadata(ctx context.Context) (*types.Metadata, error) {
	return getMetadata(ctx, mediaRecord.client, mediaRecord.MediaRecord.HREF)
}

// AddMetadata adds metadata key/value pair provided as input.
func (mediaRecord *MediaRecord) AddMetadata(ctx context.Context, key string, value string) (*MediaRecord, error) {
	task, err := mediaRecord.AddMetadataAsync(ctx, key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for media item task: %s", err)
	}

	err = mediaRecord.Refresh(ctx)
	if err != nil {
		return nil, fmt.Errorf("error refreshing media item: %s", err)
	}

	return mediaRecord, nil
}

// AddMetadataAsync calls private function addMetadata() with MediaRecord.client and MediaRecord.MediaRecord.HREF
// which adds metadata key/value pair provided as input.
func (mediaRecord *MediaRecord) AddMetadataAsync(ctx context.Context, key string, value string) (Task, error) {
	return addMetadata(ctx, mediaRecord.client, key, value, mediaRecord.MediaRecord.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
func (mediaRecord *MediaRecord) DeleteMetadata(ctx context.Context, key string) error {
	task, err := mediaRecord.DeleteMetadataAsync(ctx, key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return fmt.Errorf("error completing delete metadata for media item task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with MediaRecord.client and MediaRecord.MediaRecord.HREF
// which deletes metadata depending on key provided as input from media item.
func (mediaRecord *MediaRecord) DeleteMetadataAsync(ctx context.Context, key string) (Task, error) {
	return deleteMetadata(ctx, mediaRecord.client, key, mediaRecord.MediaRecord.HREF)
}

// GetMetadata calls private function getMetadata() with Media.client and Media.Media.HREF
// which returns a *types.Metadata struct for provided media item input.
func (media *Media) GetMetadata(ctx context.Context) (*types.Metadata, error) {
	return getMetadata(ctx, media.client, media.Media.HREF)
}

// AddMetadata adds metadata key/value pair provided as input.
func (media *Media) AddMetadata(ctx context.Context, key string, value string) (*Media, error) {
	task, err := media.AddMetadataAsync(ctx, key, value)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return nil, fmt.Errorf("error completing add metadata for media item task: %s", err)
	}

	err = media.Refresh(ctx)
	if err != nil {
		return nil, fmt.Errorf("error refreshing media item: %s", err)
	}

	return media, nil
}

// AddMetadataAsync calls private function addMetadata() with Media.client and Media.Media.HREF
// which adds metadata key/value pair provided as input.
func (media *Media) AddMetadataAsync(ctx context.Context, key string, value string) (Task, error) {
	return addMetadata(ctx, media.client, key, value, media.Media.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
func (media *Media) DeleteMetadata(ctx context.Context, key string) error {
	task, err := media.DeleteMetadataAsync(ctx, key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return fmt.Errorf("error completing delete metadata for media item task: %s", err)
	}

	return nil
}

// DeleteMetadataAsync calls private function deleteMetadata() with Media.client and Media.Media.HREF
// which deletes metadata depending on key provided as input from media item.
func (media *Media) DeleteMetadataAsync(ctx context.Context, key string) (Task, error) {
	return deleteMetadata(ctx, media.client, key, media.Media.HREF)
}
