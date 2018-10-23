/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 * Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/vmware/go-vcloud-director/types/v56"
	"net/http"
	"net/url"
)

// Independent disk
type Disk struct {
	Disk   *types.Disk
	client *Client
}

// Init independent disk struct
func NewDisk(cli *Client) *Disk {
	return &Disk{
		Disk:   new(types.Disk),
		client: cli,
	}
}

// Create an independent disk in VDC
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 102 - 103,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vdc *Vdc) CreateDisk(diskCreateParams *types.DiskCreateParams) (*Disk, error) {
	var err error
	var execLink *types.Link

	// Find the proper link for request
	for _, vdcLink := range vdc.Vdc.Link {
		if vdcLink.Rel == types.RelAdd && vdcLink.Type == types.MimeDiskCreateParams {
			execLink = vdcLink
			break
		}
	}

	if execLink == nil {
		return nil, fmt.Errorf("exec link not found")
	}

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parse URI: %s", err)
	}

	// Prepare the request payload
	diskCreateParams.Xmlns = types.NsVCloud

	xmlPayload, err := xml.Marshal(diskCreateParams)
	if err != nil {
		return nil, fmt.Errorf("error xml.Marshal: %s", err)
	}

	// Send Request
	req := vdc.client.NewRequest(nil, http.MethodPost, *reqUrl, bytes.NewBufferString(xml.Header+string(xmlPayload)))
	req.Header.Add("Content-Type", execLink.Type)
	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("error create disk: %s", err)
	}

	// Decode response
	disk := NewDisk(vdc.client)
	if err = decodeBody(resp, disk.Disk); err != nil {
		return nil, fmt.Errorf("error decoding create disk params response: %s", err)
	}

	// Return the disk
	return disk, nil
}

// Update an independent disk
// 1 Verify that the disk is not attached to a virtual machine.
// 2 Use newDiskInfo to change update the independent disk.
// 3 Return task of independent disk update
// Please verify the independent disk is not connected to any VM before call this function.
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 104 - 106,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (d *Disk) Update(newDiskInfo *types.Disk) (Task, error) {
	var err error
	var execLink *types.Link

	// Find the proper link for request
	for _, diskLink := range d.Disk.Link {
		if diskLink.Rel == types.RelEdit && diskLink.Type == types.MimeDisk {
			execLink = diskLink
			break
		}
	}

	if execLink == nil {
		return Task{}, fmt.Errorf("exec link not found")
	}

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parse URI: %s", err)
	}

	// Prepare the request payload
	xmlPayload, err := xml.Marshal(&types.Disk{
		Xmlns:          types.NsVCloud,
		Description:    newDiskInfo.Description,
		Size:           newDiskInfo.Size,
		Name:           newDiskInfo.Name,
		StorageProfile: newDiskInfo.StorageProfile,
		Owner:          newDiskInfo.Owner,
	})
	if err != nil {
		return Task{}, fmt.Errorf("error xml.Marshal: %s", err)
	}

	// Send request
	req := d.client.NewRequest(nil, http.MethodPut, *reqUrl, bytes.NewBufferString(xml.Header+string(xmlPayload)))
	req.Header.Add("Content-Type", execLink.Type)
	resp, err := checkResp(d.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error find disk: %s", err)
	}

	// Decode response
	task := NewTask(d.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding find disk response: %s", err)
	}

	// Return the task
	return *task, nil
}

// Remove an independent disk
// 1 Delete the independent disk. Make a DELETE request to the URL in the rel="remove" link in the Disk.
// 2 Return task of independent disk deletion.
// Please verify the independent disk is not connected to any VM before call this function.
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 106 - 107,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (d *Disk) Delete() (Task, error) {
	var err error
	var execLink *types.Link

	// Find the proper link for request
	for _, diskLink := range d.Disk.Link {
		if diskLink.Rel == types.RelRemove {
			execLink = diskLink
			break
		}
	}

	if execLink == nil {
		return Task{}, fmt.Errorf("exec link not found")
	}

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parse uri: %s", err)
	}

	// Make request
	req := d.client.NewRequest(nil, http.MethodDelete, *reqUrl, nil)
	resp, err := checkResp(d.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error delete disk: %s", err)
	}

	// Decode response
	task := NewTask(d.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding delete disk params response: %s", err)
	}

	// Return the task
	return *task, nil
}

// Refresh the disk information by disk href
func (d *Disk) Refresh() error {
	disk, err := FindDiskByHREF(d.client, d.Disk.HREF)
	if err != nil {
		return err
	}

	d.Disk = disk.Disk

	return nil
}

// Get a VM that attached the disk
// An independent disk can be attached to at most one virtual machine.
// If the disk doesn't attached to any VM, return empty VM reference and no error.
// Otherwise return the first VM reference and no error.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 107,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (d *Disk) AttachedVM() (*types.Reference, error) {
	var execLink *types.Link
	var err error

	// Find the proper link for request
	for _, diskLink := range d.Disk.Link {
		if diskLink.Type == types.MimeVMs {
			execLink = diskLink
			break
		}
	}

	if execLink == nil {
		return nil, fmt.Errorf("exec link not found")
	}

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parse uri: %s", err)
	}

	// Send request
	req := d.client.NewRequest(nil, http.MethodGet, *reqUrl, nil)
	req.Header.Add("Content-Type", execLink.Type)
	resp, err := checkResp(d.client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("error attached vms: %s", err)
	}

	// Decode request
	var vms = new(types.Vms)
	if err = decodeBody(resp, vms); err != nil {
		return nil, fmt.Errorf("error decoding find disk response: %s", err)
	}

	// If disk is not is attached to any VM
	if vms.VmReference == nil {
		return nil, nil
	}

	// An independent disk can be attached to at most one virtual machine so return the first result of VM reference
	return vms.VmReference, nil
}

// Find an independent disk by disk href in VDC
func (vdc *Vdc) FindDiskByHREF(href string) (*Disk, error) {
	return FindDiskByHREF(vdc.client, href)
}

// Find an independent disk by VDC client and disk href
func FindDiskByHREF(client *Client, href string) (*Disk, error) {
	// Parse request URI
	reqUrl, err := url.ParseRequestURI(href)
	if err != nil {
		return nil, fmt.Errorf("error parse URI: %s", err)
	}

	// Send request
	req := client.NewRequest(nil, http.MethodGet, *reqUrl, nil)
	resp, err := checkResp(client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("error find disk: %s", err)
	}

	// Decode response
	disk := NewDisk(client)
	if err = decodeBody(resp, disk.Disk); err != nil {
		return nil, fmt.Errorf("error decoding find disk response: %s", err)
	}

	// Return the disk
	return disk, nil
}
