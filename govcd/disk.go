// Created by terry on 9/10/2018.

package govcd

import (
	"github.com/vmware/go-vcloud-director/types/v56"
	"fmt"
	"net/url"
	"bytes"
	"encoding/xml"
	"net/http"
	"github.com/go-siris/siris/core/errors"
	"io/ioutil"
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

func (d *Disk) IsAttached() (bool, error) {
	for _, diskLink := range d.Disk.Link {
		if diskLink.Type == types.MimeVMs {
			if diskLink.Rel == "down" {
				return false, nil
			} else {
				return true, nil
			}
		}
	}

	return false, errors.New("IsAttached link not found")
}

func (d *Disk) Delete() (Task, error) {
	var execLink *types.Link
	var err error

	for _, diskLink := range d.Disk.Link {
		if diskLink.Rel == types.RelRemove {
			execLink = diskLink
		}
	}

	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		panic(err)
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

	err = task.WaitTaskCompletion()
	if err != nil {
		return Task{}, fmt.Errorf("error performing task: %#v", err)
	}

	return *task, nil
}

func (vdc *Vdc) CreateDisk(diskInfo types.DiskCreateParamsDisk) (*Disk, error) {
	var err error

	//	var c Disk
	//	err = xml.Unmarshal([]byte(`
	//<?xml version="1.0" encoding="UTF-8"?>
	//<Disk xmlns="http://www.vmware.com/vcloud/v1.5" size="1024" status="0" name="HelloDisk" id="urn:vcloud:disk:7baeba09-53f4-4230-97c3-9f28bc8247da" type="application/vnd.vmware.vcloud.disk+xml" href="https://vcloud.supereffort.com/api/disk/7baeba09-53f4-4230-97c3-9f28bc8247da" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://vcloud.supereffort.com/api/v1.5/schema/master.xsd">
	//  <Link rel="up" type="application/vnd.vmware.vcloud.vdc+xml" href="https://vcloud.supereffort.com/api/vdc/59bbe518-0244-42c0-9a8d-e699e00da9cb"/>
	//  <Link rel="remove" href="https://vcloud.supereffort.com/api/disk/7baeba09-53f4-4230-97c3-9f28bc8247da"/>
	//  <Link rel="edit" type="application/vnd.vmware.vcloud.disk+xml" href="https://vcloud.supereffort.com/api/disk/7baeba09-53f4-4230-97c3-9f28bc8247da"/>
	//  <Link rel="down" type="application/vnd.vmware.vcloud.owner+xml" href="https://vcloud.supereffort.com/api/disk/7baeba09-53f4-4230-97c3-9f28bc8247da/owner"/>
	//  <Link rel="down" type="application/vnd.vmware.vcloud.vms+xml" href="https://vcloud.supereffort.com/api/disk/7baeba09-53f4-4230-97c3-9f28bc8247da/attachedVms"/>
	//  <Link rel="down" type="application/vnd.vmware.vcloud.metadata+xml" href="https://vcloud.supereffort.com/api/disk/7baeba09-53f4-4230-97c3-9f28bc8247da/metadata"/>
	//  <Description>123</Description>
	//  <Tasks>
	//      <Task status="running" startTime="2018-10-09T14:31:55.058+08:00" serviceNamespace="com.vmware.vcloud" operationName="vdcCreateDisk" operation="Creating Disk HelloDisk(7baeba09-53f4-4230-97c3-9f28bc8247da)" expiryTime="2019-01-07T14:31:55.058+08:00" cancelRequested="false" name="task" id="urn:vcloud:task:5d77da44-8cae-4766-9729-bc1b1c6674fa" type="application/vnd.vmware.vcloud.task+xml" href="https://vcloud.supereffort.com/api/task/5d77da44-8cae-4766-9729-bc1b1c6674fa">
	//          <Link rel="task:cancel" href="https://vcloud.supereffort.com/api/task/5d77da44-8cae-4766-9729-bc1b1c6674fa/action/cancel"/>
	//          <Owner type="application/vnd.vmware.vcloud.disk+xml" name="HelloDisk" href="https://vcloud.supereffort.com/api/disk/7baeba09-53f4-4230-97c3-9f28bc8247da"/>
	//          <User type="application/vnd.vmware.admin.user+xml" name="admin" href="https://vcloud.supereffort.com/api/admin/user/fcd70928-d325-474b-ab1d-c14c6067f528"/>
	//          <Organization type="application/vnd.vmware.vcloud.org+xml" name="IOT" href="https://vcloud.supereffort.com/api/org/f24042ee-b9cd-4227-9ecf-b9ed56af9d1a"/>
	//          <Details/>
	//      </Task>
	//  </Tasks>
	//  <StorageProfile type="application/vnd.vmware.vcloud.vdcStorageProfile+xml" name="Enhanced" href="https://vcloud.supereffort.com/api/vdcStorageProfile/5ab7252a-9b95-4614-82fa-d5a38ce937a5"/>
	//  <Owner type="application/vnd.vmware.vcloud.owner+xml">
	//      <User type="application/vnd.vmware.admin.user+xml" name="admin" href="https://vcloud.supereffort.com/api/admin/user/fcd70928-d325-474b-ab1d-c14c6067f528"/>
	//  </Owner>
	//</Disk>
	//`), &c)
	//	fmt.Println(c.Owner.User, err)
	//	return
	var execLink *types.Link

	for _, vdcLink := range vdc.Vdc.Link {
		if vdcLink.Rel == types.RelAdd && vdcLink.Type == types.MimeDiskCreateParams {
			execLink = vdcLink
		}
	}

	reqUrl, err := url.ParseRequestURI(execLink.HREF)
	if err != nil {
		panic(err)
	}

	diskCreateParamsType := types.DiskCreateParamsType{
		Xmlns: types.NsVCloud,
		Disk:  diskInfo,
	}

	xmlPayload, err := xml.Marshal(diskCreateParamsType)
	fmt.Println(execLink.Type)
	x := bytes.NewBufferString(xml.Header + string(xmlPayload))
	fmt.Println(x.String())
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

	task := NewTask(vdc.client)
	for _, taskItem := range disk.Disk.Tasks {
		fmt.Println(taskItem)
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		if err != nil {
			return nil, fmt.Errorf("error performing task: %#v", err)
		}
	}

	//disk.Delete()

	return disk, nil
}

func (vdc *Vdc) FindDiskByHREF(href string) (*Disk, error) {

	reqUrl, err := url.ParseRequestURI(href)
	if err != nil {
		panic(err)
	}

	req := vdc.client.NewRequest(nil, http.MethodGet, *reqUrl, nil)
	resp, err := checkResp(vdc.client.Http.Do(req))

	if err != nil {
		return nil, fmt.Errorf("error find disk: %s", err)
	}

	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))

	disk := NewDisk(vdc.client)
	if err = decodeBody(resp, disk.Disk); err != nil {
		return nil, fmt.Errorf("error decoding find disk response: %s", err)
	}

	return disk, nil
}
