// Created by terry on 9/10/2018.

package govcd

import (
	"github.com/vmware/go-vcloud-director/types/v56"
	"fmt"
	"net/url"
	"bytes"
	"encoding/xml"
	"net/http"
)

type Disk struct {
	Disk   *types.DiskType
	client *Client
}

func NewDisk(cli *Client) *Disk {
	return &Disk{
		Disk:   new(types.DiskType),
		client: cli,
	}
}

func (d *Disk) AttachedVM() (*types.Reference, error) {
	var execLink *types.Link
	var err error

	for _, diskLink := range d.Disk.Link {
		if diskLink.Type == types.MimeVMs {
			execLink = diskLink
		}
	}

	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parse uri: %s", err)
	}

	req := d.client.NewRequest(nil, http.MethodGet, *reqUrl, nil)
	resp, err := checkResp(d.client.Http.Do(req))

	if err != nil {
		return nil, fmt.Errorf("error attached vms: %s", err)
	}

	var vmsType = new(types.VmsType)

	if err = decodeBody(resp, vmsType); err != nil {
		return nil, fmt.Errorf("error decoding find disk response: %s", err)
	}

	if len(vmsType.VmReference) <= 0 {
		return nil, nil
	}

	return vmsType.VmReference[0], nil
}

func (d *Disk) Delete() (Task, error) {
	var err error
	var execLink *types.Link

	for _, diskLink := range d.Disk.Link {
		if diskLink.Rel == types.RelRemove {
			execLink = diskLink
		}
	}

	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parse uri: %s", err)
	}

	req := d.client.NewRequest(nil, http.MethodDelete, *reqUrl, nil)
	resp, err := checkResp(d.client.Http.Do(req))

	if err != nil {
		return Task{}, fmt.Errorf("error delete disk: %s", err)
	}

	task := NewTask(d.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding delete disk params response: %s", err)
	}

	return *task, nil
}

func (vdc *Vdc) CreateDisk(diskInfo *types.DiskCreateParamsDisk) (*Disk, error) {
	var err error
	var execLink *types.Link

	for _, vdcLink := range vdc.Vdc.Link {
		if vdcLink.Rel == types.RelAdd && vdcLink.Type == types.MimeDiskCreateParams {
			execLink = vdcLink
		}
	}

	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parse uri create disk params: %s", err)
	}

	diskCreateParamsType := types.DiskCreateParamsType{
		Xmlns: types.NsVCloud,
		Disk:  diskInfo,
	}

	xmlPayload, err := xml.Marshal(diskCreateParamsType)

	req := vdc.client.NewRequest(nil, http.MethodPost, *reqUrl, bytes.NewBufferString(xml.Header+string(xmlPayload)))
	req.Header.Add("Content-Type", execLink.Type)
	resp, err := checkResp(vdc.client.Http.Do(req))

	if err != nil {
		return nil, fmt.Errorf("error create disk: %s", err)
	}

	disk := NewDisk(vdc.client)
	if err = decodeBody(resp, disk.Disk); err != nil {
		return nil, fmt.Errorf("error decoding create disk params response: %s", err)
	}

	return disk, nil
}

func (vdc *Vdc) FindDiskByHREF(href string) (*Disk, error) {
	reqUrl, err := url.ParseRequestURI(href)
	if err != nil {
		return nil, fmt.Errorf("error parse uri: %s", err)
	}

	req := vdc.client.NewRequest(nil, http.MethodGet, *reqUrl, nil)
	resp, err := checkResp(vdc.client.Http.Do(req))

	if err != nil {
		return nil, fmt.Errorf("error find disk: %s", err)
	}

	disk := NewDisk(vdc.client)
	if err = decodeBody(resp, disk.Disk); err != nil {
		return nil, fmt.Errorf("error decoding find disk response: %s", err)
	}

	return disk, nil
}
