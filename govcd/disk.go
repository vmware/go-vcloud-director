// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

// Independent disk
type Disk struct {
	Disk   *types.Disk
	client *Client
}

// Independent disk query record
type DiskRecord struct {
	Disk   *types.DiskRecordType
	client *Client
}

// Init independent disk struct
func NewDisk(cli *Client) *Disk {
	return &Disk{
		Disk:   new(types.Disk),
		client: cli,
	}
}

// Create instance with reference to types.DiskRecordType
func NewDiskRecord(cli *Client) *DiskRecord {
	return &DiskRecord{
		Disk:   new(types.DiskRecordType),
		client: cli,
	}
}

// Create an independent disk in VDC
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 102 - 103,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vdc *Vdc) CreateDisk(diskCreateParams *types.DiskCreateParams) (Task, error) {
	util.Logger.Printf("[TRACE] Create disk, name: %s, size: %d \n",
		diskCreateParams.Disk.Name,
		diskCreateParams.Disk.SizeMb,
	)

	if diskCreateParams.Disk.Name == "" {
		return Task{}, fmt.Errorf("disk name is required")
	}

	var err error
	var createDiskLink *types.Link

	// Find the proper link for request
	for _, vdcLink := range vdc.Vdc.Link {
		if vdcLink.Rel == types.RelAdd && vdcLink.Type == types.MimeDiskCreateParams {
			util.Logger.Printf("[TRACE] Create disk - found the proper link for request, HREF: %s, name: %s, type: %s, id: %s, rel: %s \n",
				vdcLink.HREF,
				vdcLink.Name,
				vdcLink.Type,
				vdcLink.ID,
				vdcLink.Rel)
			createDiskLink = vdcLink
			break
		}
	}

	if createDiskLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for create disk in vdc Link")
	}

	// Prepare the request payload
	diskCreateParams.Xmlns = types.XMLNamespaceVCloud

	disk := NewDisk(vdc.client)

	_, err = vdc.client.ExecuteRequest(createDiskLink.HREF, http.MethodPost,
		createDiskLink.Type, "error create disk: %s", diskCreateParams, disk.Disk)
	if err != nil {
		return Task{}, err
	}
	// Obtain disk task
	if len(disk.Disk.Tasks.Task) == 0 {
		return Task{}, errors.New("error cannot find disk creation task in API response")
	}
	task := NewTask(vdc.client)
	if disk.Disk.Tasks == nil || len(disk.Disk.Tasks.Task) == 0 {
		return Task{}, fmt.Errorf("no task found after disk %s creation", diskCreateParams.Disk.Name)
	}
	task.Task = disk.Disk.Tasks.Task[0]

	util.Logger.Printf("[TRACE] AFTER CREATE DISK\n %s\n", prettyDisk(*disk.Disk))
	// Return the disk
	return *task, nil
}

// Update an independent disk
// 1 Verify the independent disk is not connected to any VM
// 2 Use newDiskInfo to change update the independent disk
// 3 Return task of independent disk update
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 104 - 106,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (disk *Disk) Update(newDiskInfo *types.Disk) (Task, error) {
	util.Logger.Printf("[TRACE] Update disk, name: %s, size: %d, HREF: %s \n",
		newDiskInfo.Name,
		newDiskInfo.SizeMb,
		disk.Disk.HREF,
	)

	var err error

	if newDiskInfo.Name == "" {
		return Task{}, fmt.Errorf("disk name is required")
	}

	// Verify the independent disk is not connected to any VM
	vmRef, err := disk.AttachedVM()
	if err != nil {
		return Task{}, fmt.Errorf("error find attached VM: %s", err)
	}
	if vmRef != nil {
		return Task{}, errors.New("error disk is attached")
	}

	var updateDiskLink *types.Link

	// Find the proper link for request
	for _, diskLink := range disk.Disk.Link {
		if diskLink.Rel == types.RelEdit && diskLink.Type == types.MimeDisk {
			util.Logger.Printf("[TRACE] Update disk - found the proper link for request, HREF: %s, name: %s, type: %s,id: %s, rel: %s \n",
				diskLink.HREF,
				diskLink.Name,
				diskLink.Type,
				diskLink.ID,
				diskLink.Rel)
			updateDiskLink = diskLink
			break
		}
	}

	if updateDiskLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for update disk in disk Link")
	}

	// Prepare the request payload
	xmlPayload := &types.Disk{
		Xmlns:          types.XMLNamespaceVCloud,
		Description:    newDiskInfo.Description,
		SizeMb:         newDiskInfo.SizeMb,
		Name:           newDiskInfo.Name,
		StorageProfile: newDiskInfo.StorageProfile,
		Owner:          newDiskInfo.Owner,
	}

	// Return the task
	return disk.client.ExecuteTaskRequest(updateDiskLink.HREF, http.MethodPut,
		updateDiskLink.Type, "error updating disk: %s", xmlPayload)
}

// Remove an independent disk
// 1 Verify the independent disk is not connected to any VM
// 2 Delete the independent disk. Make a DELETE request to the URL in the rel="remove" link in the Disk
// 3 Return task of independent disk deletion
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 106 - 107,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (disk *Disk) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] Delete disk, HREF: %s \n", disk.Disk.HREF)

	var err error

	// Verify the independent disk is not connected to any VM
	vmRef, err := disk.AttachedVM()
	if err != nil {
		return Task{}, fmt.Errorf("error find attached VM: %s", err)
	}
	if vmRef != nil {
		return Task{}, errors.New("error disk is attached")
	}

	var deleteDiskLink *types.Link

	// Find the proper link for request
	for _, diskLink := range disk.Disk.Link {
		if diskLink.Rel == types.RelRemove {
			util.Logger.Printf("[TRACE] Delete disk - found the proper link for request, HREF: %s, name: %s, type: %s,id: %s, rel: %s \n",
				diskLink.HREF,
				diskLink.Name,
				diskLink.Type,
				diskLink.ID,
				diskLink.Rel)
			deleteDiskLink = diskLink
			break
		}
	}

	if deleteDiskLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for delete disk in disk Link")
	}

	// Return the task
	return disk.client.ExecuteTaskRequest(deleteDiskLink.HREF, http.MethodDelete,
		"", "error delete disk: %s", nil)
}

// Refresh the disk information by disk href
func (disk *Disk) Refresh() error {
	if disk.Disk == nil || disk.Disk.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}
	util.Logger.Printf("[TRACE] Disk refresh, HREF: %s\n", disk.Disk.HREF)

	unmarshalledDisk := &types.Disk{}

	_, err := disk.client.ExecuteRequest(disk.Disk.HREF, http.MethodGet,
		"", "error refreshing independent disk: %s", nil, unmarshalledDisk)
	if err != nil {
		return err
	}
	disk.Disk = unmarshalledDisk

	// The request was successful
	return nil
}

// Get a VM that is attached the disk
// An independent disk can be attached to at most one virtual machine.
// If the disk isn't attached to any VM, return empty VM reference and no error.
// Otherwise return the first VM reference and no error.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 107,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (disk *Disk) AttachedVM() (*types.Reference, error) {
	util.Logger.Printf("[TRACE] Disk attached VM, HREF: %s\n", disk.Disk.HREF)

	var attachedVMLink *types.Link

	// Find the proper link for request
	for _, diskLink := range disk.Disk.Link {
		if diskLink.Type == types.MimeVMs {
			util.Logger.Printf("[TRACE] Disk attached VM - found the proper link for request, HREF: %s, name: %s, type: %s,id: %s, rel: %s \n",
				diskLink.HREF,
				diskLink.Name,
				diskLink.Type,
				diskLink.ID,
				diskLink.Rel)

			attachedVMLink = diskLink
			break
		}
	}

	if attachedVMLink == nil {
		return nil, fmt.Errorf("could not find request URL for attached vm in disk Link")
	}

	// Decode request
	var vms = new(types.Vms)

	_, err := disk.client.ExecuteRequest(attachedVMLink.HREF, http.MethodGet,
		attachedVMLink.Type, "error getting attached vms: %s", nil, vms)
	if err != nil {
		return nil, err
	}

	// If disk is not attached to any VM
	if len(vms.VmReference) == 0 {
		return nil, nil
	}

	// An independent disk can be attached to at most one virtual machine so return the first result of VM reference
	return vms.VmReference[0], nil
}

// Find an independent disk by disk href in VDC
// Deprecated: Use VDC.GetDiskByHref()
func (vdc *Vdc) FindDiskByHREF(href string) (*Disk, error) {
	util.Logger.Printf("[TRACE] VDC find disk By HREF: %s\n", href)

	return FindDiskByHREF(vdc.client, href)
}

// Find an independent disk by VDC client and disk href
// Deprecated: Use VDC.GetDiskByHref()
func FindDiskByHREF(client *Client, href string) (*Disk, error) {
	util.Logger.Printf("[TRACE] Find disk By HREF: %s\n", href)

	disk := NewDisk(client)

	_, err := client.ExecuteRequest(href, http.MethodGet,
		"", "error finding disk: %s", nil, disk.Disk)

	// Return the disk
	return disk, err

}

// QueryDisk find independent disk using disk name. Returns DiskRecord type
func (vdc *Vdc) QueryDisk(diskName string) (DiskRecord, error) {

	if diskName == "" {
		return DiskRecord{}, fmt.Errorf("disk name can not be empty")
	}

	typeMedia := "disk"
	if vdc.client.IsSysAdmin {
		typeMedia = "adminDisk"
	}

	results, err := vdc.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia,
		"filter": "name==" + url.QueryEscape(diskName) + ";vdc==" + vdc.vdcId(), "filterEncoded": "true"})
	if err != nil {
		return DiskRecord{}, fmt.Errorf("error querying disk %s", err)
	}

	diskResults := results.Results.DiskRecord
	if vdc.client.IsSysAdmin {
		diskResults = results.Results.AdminDiskRecord
	}

	newDisk := NewDiskRecord(vdc.client)

	if len(diskResults) == 1 {
		newDisk.Disk = diskResults[0]
	} else {
		return DiskRecord{}, fmt.Errorf("found results %d", len(diskResults))
	}

	return *newDisk, nil
}

// QueryDisks find independent disks using disk name. Returns list of DiskRecordType
func (vdc *Vdc) QueryDisks(diskName string) (*[]*types.DiskRecordType, error) {

	if diskName == "" {
		return nil, fmt.Errorf("disk name can't be empty")
	}

	typeMedia := "disk"
	if vdc.client.IsSysAdmin {
		typeMedia = "adminDisk"
	}

	results, err := vdc.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia,
		"filter": "name==" + url.QueryEscape(diskName) + ";vdc==" + vdc.vdcId(), "filterEncoded": "true"})
	if err != nil {
		return nil, fmt.Errorf("error querying disks %s", err)
	}

	diskResults := results.Results.DiskRecord
	if vdc.client.IsSysAdmin {
		diskResults = results.Results.AdminDiskRecord
	}

	return &diskResults, nil
}

// GetDiskByHref finds a Disk by HREF
// On success, returns a pointer to the Disk structure and a nil error
// On failure, returns a nil pointer and an error
func (vdc *Vdc) GetDiskByHref(diskHref string) (*Disk, error) {
	util.Logger.Printf("[TRACE] Get Disk By Href: %s\n", diskHref)
	Disk := NewDisk(vdc.client)

	_, err := vdc.client.ExecuteRequest(diskHref, http.MethodGet,
		"", "error retrieving Disk: %s", nil, Disk.Disk)
	if err != nil && (strings.Contains(err.Error(), "MajorErrorCode:403") || strings.Contains(err.Error(), "does not exist")) {
		return nil, ErrorEntityNotFound
	}
	if err != nil {
		return nil, err
	}
	return Disk, nil
}

// GetDisksByName finds one or more Disks by Name
// On success, returns a pointer to the Disk list and a nil error
// On failure, returns a nil pointer and an error
func (vdc *Vdc) GetDisksByName(diskName string, refresh bool) (*[]Disk, error) {
	util.Logger.Printf("[TRACE] Get Disk By Name: %s\n", diskName)
	var diskList []Disk
	if refresh {
		err := vdc.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceEntity := range resourceEntities.ResourceEntity {
			if resourceEntity.Name == diskName && resourceEntity.Type == "application/vnd.vmware.vcloud.disk+xml" {
				disk, err := vdc.GetDiskByHref(resourceEntity.HREF)
				if err != nil {
					return nil, err
				}
				diskList = append(diskList, *disk)
			}
		}
	}
	if len(diskList) == 0 {
		return nil, ErrorEntityNotFound
	}
	return &diskList, nil
}

// GetDiskById finds a Disk by ID
// On success, returns a pointer to the Disk structure and a nil error
// On failure, returns a nil pointer and an error
func (vdc *Vdc) GetDiskById(diskId string, refresh bool) (*Disk, error) {
	util.Logger.Printf("[TRACE] Get Disk By Id: %s\n", diskId)
	if refresh {
		err := vdc.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceEntity := range resourceEntities.ResourceEntity {
			if equalIds(diskId, resourceEntity.ID, resourceEntity.HREF) && resourceEntity.Type == "application/vnd.vmware.vcloud.disk+xml" {
				return vdc.GetDiskByHref(resourceEntity.HREF)
			}
		}
	}
	return nil, ErrorEntityNotFound
}

// Get a VMs HREFs that is attached to the disk
// An independent disk can be attached to at most one virtual machine.
// If the disk isn't attached to any VM, return empty slice.
// Otherwise return the list of VMs HREFs.
func (disk *Disk) GetAttachedVmsHrefs() ([]string, error) {
	util.Logger.Printf("[TRACE] GetAttachedVmsHrefs, HREF: %s\n", disk.Disk.HREF)

	var vmHrefs []string

	var attachedVMsLink *types.Link

	// Find the proper link for request
	for _, diskLink := range disk.Disk.Link {
		if diskLink.Type == types.MimeVMs {
			util.Logger.Printf("[TRACE] GetAttachedVmsHrefs - found the proper link for request, HREF: %s, name: %s, type: %s,id: %s, rel: %s \n",
				diskLink.HREF, diskLink.Name, diskLink.Type, diskLink.ID, diskLink.Rel)

			attachedVMsLink = diskLink
			break
		}
	}

	if attachedVMsLink == nil {
		return nil, fmt.Errorf("error GetAttachedVmsHrefs - could not find request URL for attached vm in disk Link")
	}

	// Decode request
	var vms = new(types.Vms)

	_, err := disk.client.ExecuteRequest(attachedVMsLink.HREF, http.MethodGet,
		attachedVMsLink.Type, "error GetAttachedVmsHrefs - error getting attached VMs: %s", nil, vms)
	if err != nil {
		return nil, err
	}

	// If disk is not attached to any VM
	if len(vms.VmReference) == 0 {
		return nil, nil
	}

	for _, value := range vms.VmReference {
		vmHrefs = append(vmHrefs, value.HREF)
	}

	return vmHrefs, nil
}
