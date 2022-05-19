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

// MetadataCompatible allows consumers of this library to consider all structs that implement
// this interface to be the same type
type MetadataCompatible interface {
	GetMetadata() (*types.Metadata, error)
	AddMetadataEntry(typedValue, key, value string) error
	MergeMetadata(typedValue string, metadata map[string]interface{}) error
	DeleteMetadataEntry(key string) error
}

// ----------------
// REST API metadata CRUD functions

// GetMetadataByHref returns metadata from the given resource reference.
func (vcdClient *VCDClient) GetMetadataByHref(href string) (*types.Metadata, error) {
	return getMetadata(&vcdClient.Client, href)
}

// AddMetadataEntryByHref adds metadata typedValue and key/value pair provided as input to the given resource reference,
// then waits for the task to finish.
func (vcdClient *VCDClient) AddMetadataEntryByHref(href, typedValue, key, value string) error {
	task, err := vcdClient.AddMetadataEntryByHrefAsync(href, typedValue, key, value)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// AddMetadataEntryByHrefAsync adds metadata typedValue and key/value pair provided as input to the given resource reference
// and returns the task.
func (vcdClient *VCDClient) AddMetadataEntryByHrefAsync(href, typedValue, key, value string) (Task, error) {
	return addMetadata(&vcdClient.Client, typedValue, key, value, href)
}

// MergeMetadataByHrefAsync merges metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// and returns the task.
func (vcdClient *VCDClient) MergeMetadataByHrefAsync(href, typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(&vcdClient.Client, typedValue, metadata, href)
}

// MergeMetadataByHref merges metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vcdClient *VCDClient) MergeMetadataByHref(href, typedValue string, metadata map[string]interface{}) error {
	task, err := vcdClient.MergeMetadataByHrefAsync(href, typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntryByHref deletes metadata from the given resource reference, depending on key provided as input
// and waits for the task to finish.
func (vcdClient *VCDClient) DeleteMetadataEntryByHref(href, key string) error {
	task, err := vcdClient.DeleteMetadataEntryByHrefAsync(href, key)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// DeleteMetadataEntryByHrefAsync deletes metadata from the given resource reference, depending on key provided as input
// and returns a task.
func (vcdClient *VCDClient) DeleteMetadataEntryByHrefAsync(href, key string) (Task, error) {
	return deleteMetadata(&vcdClient.Client, key, href)
}

// GetMetadata returns VM metadata.
func (vm *VM) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vm.client, vm.VM.HREF)
}

// AddMetadataEntry adds VM metadata typedValue and key/value pair provided as input
// and waits for the task to finish.
func (vm *VM) AddMetadataEntry(typedValue, key, value string) error {
	task, err := vm.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = vm.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// AddMetadataEntryAsync adds VM metadata typedValue and key/value pair provided as input
// and returns the task.
func (vm *VM) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(vm.client, typedValue, key, value, vm.VM.HREF)
}

// MergeMetadataAsync merges VM metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then returns the task.
func (vm *VM) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(vm.client, typedValue, metadata, vm.VM.HREF)
}

// MergeMetadata merges VM metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vm *VM) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := vm.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes VM metadata by key provided as input and waits for the task to finish.
func (vm *VM) DeleteMetadataEntry(key string) error {
	task, err := vm.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = vm.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// DeleteMetadataEntryAsync deletes VM metadata depending on key provided as input
// and returns the task.
func (vm *VM) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(vm.client, key, vm.VM.HREF)
}

// GetMetadata returns Vdc metadata.
// Note: Requires system administrator privileges.
func (vdc *Vdc) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vdc.client, getAdminURL(vdc.Vdc.HREF))
}

// AddMetadataEntry adds Vdc metadata typedValue and key/value pair provided as input
// and waits for the task to finish.
// Note: Requires system administrator privileges.
func (vdc *Vdc) AddMetadataEntry(typedValue, key, value string) error {
	task, err := vdc.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = vdc.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// AddMetadataEntryAsync adds Vdc metadata typedValue and key/value pair provided as input and returns the task.
// Note: Requires system administrator privileges.
func (vdc *Vdc) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(vdc.client, typedValue, key, value, getAdminURL(vdc.Vdc.HREF))
}

// MergeMetadataAsync merges VM metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
// Note: Requires system administrator privileges.
func (vdc *Vdc) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(vdc.client, typedValue, metadata, getAdminURL(vdc.Vdc.HREF))
}

// MergeMetadata merges VM metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
// Note: Requires system administrator privileges.
func (vdc *Vdc) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := vdc.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes Vdc metadata by key provided as input and waits for
// the task to finish.
// Note: Requires system administrator privileges.
func (vdc *Vdc) DeleteMetadataEntry(key string) error {
	task, err := vdc.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = vdc.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// DeleteMetadataEntryAsync deletes Vdc metadata depending on key provided as input and returns the task.
// Note: Requires system administrator privileges.
func (vdc *Vdc) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(vdc.client, key, getAdminURL(vdc.Vdc.HREF))
}

// GetMetadata returns VApp metadata.
func (vapp *VApp) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vapp.client, vapp.VApp.HREF)
}

// AddMetadataEntry adds VApp metadata typedValue and key/value pair provided as input
// and waits for the task to finish.
func (vapp *VApp) AddMetadataEntry(typedValue, key, value string) error {
	task, err := vapp.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = vapp.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// AddMetadataEntryAsync adds VApp metadata typedValue and key/value pair provided as input and returns the task.
func (vapp *VApp) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(vapp.client, typedValue, key, value, vapp.VApp.HREF)
}

// MergeMetadataAsync merges VApp metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vapp *VApp) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(vapp.client, typedValue, metadata, vapp.VApp.HREF)
}

// MergeMetadata merges VApp metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vapp *VApp) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := vapp.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes VApp metadata by key provided as input and waits for
// the task to finish.
func (vapp *VApp) DeleteMetadataEntry(key string) error {
	task, err := vapp.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = vapp.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// DeleteMetadataEntryAsync deletes VApp metadata depending on key provided as input and returns the task.
func (vapp *VApp) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(vapp.client, key, vapp.VApp.HREF)
}

// GetMetadata returns VAppTemplate metadata.
func (vAppTemplate *VAppTemplate) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vAppTemplate.client, vAppTemplate.VAppTemplate.HREF)
}

// AddMetadataEntry adds VAppTemplate metadata typedValue and key/value pair provided as input and
// waits for the task to finish.
func (vAppTemplate *VAppTemplate) AddMetadataEntry(typedValue, key, value string) error {
	task, err := vAppTemplate.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = vAppTemplate.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// AddMetadataEntryAsync adds VAppTemplate metadata typedValue and key/value pair provided as input
// and returns the task.
func (vAppTemplate *VAppTemplate) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(vAppTemplate.client, typedValue, key, value, vAppTemplate.VAppTemplate.HREF)
}

// MergeMetadataAsync merges VAppTemplate metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vAppTemplate *VAppTemplate) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(vAppTemplate.client, typedValue, metadata, vAppTemplate.VAppTemplate.HREF)
}

// MergeMetadata merges VAppTemplate metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (vAppTemplate *VAppTemplate) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := vAppTemplate.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes VAppTemplate metadata depending on key provided as input
// and waits for the task to finish.
func (vAppTemplate *VAppTemplate) DeleteMetadataEntry(key string) error {
	task, err := vAppTemplate.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = vAppTemplate.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// DeleteMetadataEntryAsync deletes VAppTemplate metadata depending on key provided as input
// and returns the task.
func (vAppTemplate *VAppTemplate) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(vAppTemplate.client, key, vAppTemplate.VAppTemplate.HREF)
}

// GetMetadata returns MediaRecord metadata.
func (mediaRecord *MediaRecord) GetMetadata() (*types.Metadata, error) {
	return getMetadata(mediaRecord.client, mediaRecord.MediaRecord.HREF)
}

// AddMetadataEntry adds MediaRecord metadata typedValue and key/value pair provided as input and
// waits for the task to finish.
func (mediaRecord *MediaRecord) AddMetadataEntry(typedValue, key, value string) error {
	task, err := mediaRecord.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = mediaRecord.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// AddMetadataEntryAsync adds MediaRecord metadata typedValue and key/value pair provided as input
// and returns the task.
func (mediaRecord *MediaRecord) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(mediaRecord.client, typedValue, key, value, mediaRecord.MediaRecord.HREF)
}

// MergeMetadataAsync merges MediaRecord metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (mediaRecord *MediaRecord) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(mediaRecord.client, typedValue, metadata, mediaRecord.MediaRecord.HREF)
}

// MergeMetadata merges MediaRecord metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (mediaRecord *MediaRecord) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := mediaRecord.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes MediaRecord metadata depending on key provided as input
// and waits for the task to finish.
func (mediaRecord *MediaRecord) DeleteMetadataEntry(key string) error {
	task, err := mediaRecord.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = mediaRecord.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// DeleteMetadataEntryAsync deletes MediaRecord metadata depending on key provided as input
// and returns the task.
func (mediaRecord *MediaRecord) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(mediaRecord.client, key, mediaRecord.MediaRecord.HREF)
}

// GetMetadata returns Media metadata.
func (media *Media) GetMetadata() (*types.Metadata, error) {
	return getMetadata(media.client, media.Media.HREF)
}

// AddMetadataEntry adds Media metadata typedValue and key/value pair provided as input
// and waits for the task to finish.
func (media *Media) AddMetadataEntry(typedValue, key, value string) error {
	task, err := media.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = media.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// AddMetadataEntryAsync adds Media metadata typedValue and key/value pair provided as input
// and returns the task.
func (media *Media) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(media.client, typedValue, key, value, media.Media.HREF)
}

// MergeMetadataAsync merges Media metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (media *Media) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(media.client, typedValue, metadata, media.Media.HREF)
}

// MergeMetadata merges Media metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (media *Media) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := media.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes Media metadata depending on key provided as input
// and waits for the task to finish.
func (media *Media) DeleteMetadataEntry(key string) error {
	task, err := media.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = media.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// DeleteMetadataEntryAsync deletes Media metadata depending on key provided as input
// and returns the task.
func (media *Media) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(media.client, key, media.Media.HREF)
}

// GetMetadata returns Catalog metadata.
func (catalog *Catalog) GetMetadata() (*types.Metadata, error) {
	return getMetadata(catalog.client, catalog.Catalog.HREF)
}

// GetMetadata returns AdminCatalog metadata.
func (adminCatalog *AdminCatalog) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminCatalog.client, adminCatalog.AdminCatalog.HREF)
}

// AddMetadataEntry adds AdminCatalog metadata typedValue and key/value pair provided as input
// and waits for the task to finish.
func (adminCatalog *AdminCatalog) AddMetadataEntry(typedValue, key, value string) error {
	task, err := adminCatalog.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = adminCatalog.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// AddMetadataEntryAsync adds AdminCatalog metadata typedValue and key/value pair provided as input
// and returns the task.
func (adminCatalog *AdminCatalog) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(adminCatalog.client, typedValue, key, value, adminCatalog.AdminCatalog.HREF)
}

// MergeMetadataAsync merges AdminCatalog metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminCatalog *AdminCatalog) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(adminCatalog.client, typedValue, metadata, adminCatalog.AdminCatalog.HREF)
}

// MergeMetadata merges AdminCatalog metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminCatalog *AdminCatalog) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := adminCatalog.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes AdminCatalog metadata depending on key provided as input
// and waits for the task to finish.
func (adminCatalog *AdminCatalog) DeleteMetadataEntry(key string) error {
	task, err := adminCatalog.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	err = adminCatalog.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// DeleteMetadataEntryAsync deletes AdminCatalog metadata depending on key provided as input
// and returns a task.
func (adminCatalog *AdminCatalog) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(adminCatalog.client, key, adminCatalog.AdminCatalog.HREF)
}

// GetMetadata returns the Org metadata of the corresponding organization seen as administrator
func (org *Org) GetMetadata() (*types.Metadata, error) {
	return getMetadata(org.client, org.Org.HREF)
}

// GetMetadata returns the AdminOrg metadata of the corresponding organization seen as administrator
func (adminOrg *AdminOrg) GetMetadata() (*types.Metadata, error) {
	return getMetadata(adminOrg.client, adminOrg.AdminOrg.HREF)
}

// AddMetadataEntry adds AdminOrg metadata key/value pair provided as input to the corresponding organization seen as administrator
// and waits for completion.
func (adminOrg *AdminOrg) AddMetadataEntry(typedValue, key, value string) error {
	task, err := adminOrg.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// AddMetadataEntryAsync adds AdminOrg metadata key/value pair provided as input to the corresponding organization seen as administrator
// and returns a task.
func (adminOrg *AdminOrg) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(adminOrg.client, typedValue, key, value, adminOrg.AdminOrg.HREF)
}

// MergeMetadataAsync merges AdminOrg metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminOrg *AdminOrg) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(adminOrg.client, typedValue, metadata, adminOrg.AdminOrg.HREF)
}

// MergeMetadata merges AdminOrg metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (adminOrg *AdminOrg) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := adminOrg.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes metadata of the corresponding organization with the given key, and waits for completion
func (adminOrg *AdminOrg) DeleteMetadataEntry(key string) error {
	task, err := adminOrg.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error completing delete metadata for organization task: %s", err)
	}

	return nil
}

// DeleteMetadataEntryAsync deletes metadata of the corresponding organization with the given key, and returns
// a task.
func (adminOrg *AdminOrg) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(adminOrg.client, key, adminOrg.AdminOrg.HREF)
}

// GetMetadata returns the metadata of the corresponding independent disk
func (disk *Disk) GetMetadata() (*types.Metadata, error) {
	return getMetadata(disk.client, disk.Disk.HREF)
}

// AddMetadataEntry adds metadata key/value pair provided as input to the corresponding independent disk and waits for completion.
func (disk *Disk) AddMetadataEntry(typedValue, key, value string) error {
	task, err := disk.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// AddMetadataEntryAsync adds metadata key/value pair provided as input to the corresponding independent disk and returns a task.
func (disk *Disk) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(disk.client, typedValue, key, value, disk.Disk.HREF)
}

// MergeMetadataAsync merges Disk metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (disk *Disk) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(disk.client, typedValue, metadata, disk.Disk.HREF)
}

// MergeMetadata merges Disk metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (disk *Disk) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := disk.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes metadata of the corresponding independent disk with the given key, and waits for completion
func (disk *Disk) DeleteMetadataEntry(key string) error {
	task, err := disk.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error completing delete metadata for independent disk task: %s", err)
	}

	return nil
}

// DeleteMetadataEntryAsync deletes metadata of the corresponding independent disk with the given key, and returns
// a task.
func (disk *Disk) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(disk.client, key, disk.Disk.HREF)
}

// GetMetadata returns OrgVDCNetwork metadata.
func (orgVdcNetwork *OrgVDCNetwork) GetMetadata() (*types.Metadata, error) {
	return getMetadata(orgVdcNetwork.client, orgVdcNetwork.OrgVDCNetwork.HREF)
}

// AddMetadataEntry adds OrgVDCNetwork metadata typedValue and key/value pair provided as input
// and waits for the task to finish.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) AddMetadataEntry(typedValue, key, value string) error {
	task, err := orgVdcNetwork.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// AddMetadataEntryAsync adds OrgVDCNetwork metadata typedValue and key/value pair provided as input
// and returns the task.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(orgVdcNetwork.client, typedValue, key, value, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF))
}

// MergeMetadataAsync merges OrgVDCNetwork metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(orgVdcNetwork.client, typedValue, metadata, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF))
}

// MergeMetadata merges OrgVDCNetwork metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := orgVdcNetwork.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes OrgVDCNetwork metadata depending on key provided as input
// and waits for the task to finish.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) DeleteMetadataEntry(key string) error {
	task, err := orgVdcNetwork.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// DeleteMetadataEntryAsync deletes OrgVDCNetwork metadata depending on key provided as input
// and returns a task.
// Note: Requires system administrator privileges.
func (orgVdcNetwork *OrgVDCNetwork) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(orgVdcNetwork.client, key, getAdminURL(orgVdcNetwork.OrgVDCNetwork.HREF))
}

// GetMetadata returns CatalogItem metadata.
func (catalogItem *CatalogItem) GetMetadata() (*types.Metadata, error) {
	return getMetadata(catalogItem.client, catalogItem.CatalogItem.HREF)
}

// AddMetadataEntry adds CatalogItem metadata typedValue and key/value pair provided as input
// and waits for the task to finish.
func (catalogItem *CatalogItem) AddMetadataEntry(typedValue, key, value string) error {
	task, err := catalogItem.AddMetadataEntryAsync(typedValue, key, value)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// AddMetadataEntryAsync adds CatalogItem metadata typedValue and key/value pair provided as input
// and returns the task.
func (catalogItem *CatalogItem) AddMetadataEntryAsync(typedValue, key, value string) (Task, error) {
	return addMetadata(catalogItem.client, typedValue, key, value, catalogItem.CatalogItem.HREF)
}

// MergeMetadataAsync merges CatalogItem metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (catalogItem *CatalogItem) MergeMetadataAsync(typedValue string, metadata map[string]interface{}) (Task, error) {
	return mergeAllMetadata(catalogItem.client, typedValue, metadata, catalogItem.CatalogItem.HREF)
}

// MergeMetadata merges CatalogItem metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// then waits for the task to complete.
func (catalogItem *CatalogItem) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := catalogItem.MergeMetadataAsync(typedValue, metadata)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes CatalogItem metadata depending on key provided as input
// and waits for the task to finish.
func (catalogItem *CatalogItem) DeleteMetadataEntry(key string) error {
	task, err := catalogItem.DeleteMetadataEntryAsync(key)
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// DeleteMetadataEntryAsync deletes CatalogItem metadata depending on key provided as input
// and returns a task.
func (catalogItem *CatalogItem) DeleteMetadataEntryAsync(key string) (Task, error) {
	return deleteMetadata(catalogItem.client, key, catalogItem.CatalogItem.HREF)
}

// ----------------
// OpenAPI metadata functions

// GetMetadata returns OpenApiOrgVdcNetwork metadata.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is supported from v37.0 and is currently in alpha at the moment. See https://github.com/vmware/go-vcloud-director/pull/455
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) GetMetadata() (*types.Metadata, error) {
	return getMetadata(openApiOrgVdcNetwork.client, fmt.Sprintf("%s/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), strings.ReplaceAll(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID, "urn:vcloud:network:", "")))
}

// AddMetadataEntry adds OpenApiOrgVdcNetwork metadata typedValue and key/value pair provided as input
// and waits for the task to finish.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is supported from v37.0 and is currently in alpha at the moment. See https://github.com/vmware/go-vcloud-director/pull/455
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) AddMetadataEntry(typedValue, key, value string) error {
	task, err := addMetadata(openApiOrgVdcNetwork.client, typedValue, key, value, fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), strings.ReplaceAll(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID, "urn:vcloud:network:", "")))
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// MergeMetadata merges OpenApiOrgVdcNetwork metadata provided as a key-value map of type `typedValue` with the already present in VCD,
// and waits for the task to finish.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is supported from v37.0 and is currently in alpha at the moment. See https://github.com/vmware/go-vcloud-director/pull/455
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) MergeMetadata(typedValue string, metadata map[string]interface{}) error {
	task, err := mergeAllMetadata(openApiOrgVdcNetwork.client, typedValue, metadata, fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), strings.ReplaceAll(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID, "urn:vcloud:network:", "")))
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// DeleteMetadataEntry deletes OpenApiOrgVdcNetwork metadata depending on key provided as input
// and waits for the task to finish.
// TODO: This function is currently using XML API underneath as OpenAPI metadata is supported from v37.0 and is currently in alpha at the moment. // TODO: This function is currently using XML underneath as metadata is supported in v37.0 and at the moment is in alpha state. See https://github.com/vmware/go-vcloud-director/pull/455
func (openApiOrgVdcNetwork *OpenApiOrgVdcNetwork) DeleteMetadataEntry(key string) error {
	task, err := deleteMetadata(openApiOrgVdcNetwork.client, key, fmt.Sprintf("%s/admin/network/%s", openApiOrgVdcNetwork.client.VCDHREF.String(), strings.ReplaceAll(openApiOrgVdcNetwork.OpenApiOrgVdcNetwork.ID, "urn:vcloud:network:", "")))
	if err != nil {
		return err
	}

	return task.WaitTaskCompletion()
}

// ----------------
// Generic private functions

// Generic function to retrieve metadata from VCD
func getMetadata(client *Client, requestUri string) (*types.Metadata, error) {
	metadata := &types.Metadata{}

	_, err := client.ExecuteRequest(requestUri+"/metadata/", http.MethodGet,
		types.MimeMetaData, "error retrieving metadata: %s", nil, metadata)

	return metadata, err
}

// addMetadata adds metadata to an entity.
// The function supports passing a typedValue. Use one of the constants defined.
// Constants are types.MetadataStringValue, types.MetadataNumberValue, types.MetadataDateTimeValue and types.MetadataBooleanValue.
// Only tested with types.MetadataStringValue and types.MetadataNumberValue.
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

// mergeAllMetadata merges the metadata key-values provided as parameter with existing entity metadata
func mergeAllMetadata(client *Client, typedValue string, metadata map[string]interface{}, requestUri string) (Task, error) {
	var metadataToMerge []*types.MetadataEntry
	for _, value := range metadata {
		metadataToMerge = append(metadataToMerge, &types.MetadataEntry{
			Xmlns: types.XMLNamespaceVCloud,
			Xsi:   types.XMLNamespaceXSI,
			TypedValue: &types.TypedValue{
				XsiType: typedValue,
				Value:   value.(string),
			},
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
		types.MimeMetaDataValue, "error adding metadata: %s", newMetadata)
}

// deleteMetadata Deletes metadata from an entity.
func deleteMetadata(client *Client, key string, requestUri string) (Task, error) {
	apiEndpoint := urlParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodDelete,
		"", "error deleting metadata: %s", nil)
}

// ----------------
// Deprecations

// Deprecated: use VM.DeleteMetadataEntry.
func (vm *VM) DeleteMetadata(key string) (Task, error) {
	return deleteMetadata(vm.client, key, vm.VM.HREF)
}

// Deprecated: use VM.AddMetadataEntry.
func (vm *VM) AddMetadata(key string, value string) (Task, error) {
	return addMetadata(vm.client, types.MetadataStringValue, key, value, vm.VM.HREF)
}

// Deprecated: use Vdc.DeleteMetadataEntry.
func (vdc *Vdc) DeleteMetadata(key string) (Vdc, error) {
	task, err := deleteMetadata(vdc.client, key, getAdminURL(vdc.Vdc.HREF))
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

// Deprecated: use Vdc.AddMetadataEntry.
func (vdc *Vdc) AddMetadata(key string, value string) (Vdc, error) {
	task, err := addMetadata(vdc.client, types.MetadataStringValue, key, value, getAdminURL(vdc.Vdc.HREF))
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

// Deprecated: use Vdc.AddMetadataEntryAsync.
func (vdc *Vdc) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(vdc.client, types.MetadataStringValue, key, value, getAdminURL(vdc.Vdc.HREF))
}

// Deprecated: use Vdc.DeleteMetadataEntryAsync.
func (vdc *Vdc) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(vdc.client, key, getAdminURL(vdc.Vdc.HREF))
}

// Deprecated: use VApp.DeleteMetadataEntry.
func (vapp *VApp) DeleteMetadata(key string) (Task, error) {
	return deleteMetadata(vapp.client, key, vapp.VApp.HREF)
}

// Deprecated: use VApp.AddMetadataEntry
func (vapp *VApp) AddMetadata(key string, value string) (Task, error) {
	return addMetadata(vapp.client, types.MetadataStringValue, key, value, vapp.VApp.HREF)
}

// Deprecated: use VAppTemplate.AddMetadataEntry.
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

// Deprecated: use VAppTemplate.AddMetadataEntryAsync.
func (vAppTemplate *VAppTemplate) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(vAppTemplate.client, types.MetadataStringValue, key, value, vAppTemplate.VAppTemplate.HREF)
}

// Deprecated: use VAppTemplate.DeleteMetadataEntry.
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

// Deprecated: use VAppTemplate.DeleteMetadataEntryAsync.
func (vAppTemplate *VAppTemplate) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(vAppTemplate.client, key, vAppTemplate.VAppTemplate.HREF)
}

// Deprecated: use Media.AddMetadataEntry.
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

// Deprecated: use Media.AddMetadataEntryAsync.
func (media *Media) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(media.client, types.MetadataStringValue, key, value, media.Media.HREF)
}

// Deprecated: use Media.DeleteMetadataEntry.
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

// Deprecated: use Media.DeleteMetadataEntryAsync.
func (media *Media) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(media.client, key, media.Media.HREF)
}

// GetMetadata returns MediaItem metadata.
// Deprecated: Use MediaRecord.GetMetadata.
func (mediaItem *MediaItem) GetMetadata() (*types.Metadata, error) {
	return getMetadata(mediaItem.vdc.client, mediaItem.MediaItem.HREF)
}

// AddMetadata adds metadata key/value pair provided as input.
// Deprecated: Use MediaRecord.AddMetadata.
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

// Deprecated: use MediaItem.AddMetadataEntryAsync.
func (mediaItem *MediaItem) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(mediaItem.vdc.client, types.MetadataStringValue, key, value, mediaItem.MediaItem.HREF)
}

// DeleteMetadata deletes metadata depending on key provided as input from media item.
// Deprecated: Use MediaRecord.DeleteMetadata.
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

// DeleteMetadataAsync deletes metadata depending on key provided as input from MediaItem.
// Deprecated: Use MediaRecord.DeleteMetadataAsync.
func (mediaItem *MediaItem) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(mediaItem.vdc.client, key, mediaItem.MediaItem.HREF)
}

// Deprecated: use MediaRecord.AddMetadataEntry.
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

// Deprecated: use MediaRecord.AddMetadataEntryAsync.
func (mediaRecord *MediaRecord) AddMetadataAsync(key string, value string) (Task, error) {
	return addMetadata(mediaRecord.client, types.MetadataStringValue, key, value, mediaRecord.MediaRecord.HREF)
}

// Deprecated: use MediaRecord.DeleteMetadataEntry.
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

// Deprecated: use MediaRecord.DeleteMetadataEntryAsync.
func (mediaRecord *MediaRecord) DeleteMetadataAsync(key string) (Task, error) {
	return deleteMetadata(mediaRecord.client, key, mediaRecord.MediaRecord.HREF)
}
