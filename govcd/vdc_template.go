/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
)

type VdcTemplate struct {
	VdcTemplate *types.VMWVdcTemplate
	client      *Client
}

func (vcdClient *VCDClient) CreateVdcTemplate(input types.VMWVdcTemplate) (*VdcTemplate, error) {
	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires System Administrator privileges")
	}
	href := vcdClient.Client.VCDHREF
	href.Path += "/admin/extension/vdcTemplate"

	result := types.VMWVdcTemplate{}

	_, err := vcdClient.Client.ExecuteRequest(href.String(), http.MethodPost, types.MimeVdcTemplateXml,
		"error creating VDC Template: %s", input, &result)
	if err != nil {
		return nil, err
	}

	return &VdcTemplate{
		VdcTemplate: &result,
		client:      &vcdClient.Client,
	}, nil
}

func (vcdClient *VCDClient) GetVdcTemplateById() (*VdcTemplate, error) {
	// /api/admin/extension/vdcTemplate/876b75ff-1ee1-4508-a701-2d2ff3e6f662
	return nil, nil
}

func (vcdClient *VCDClient) GetVdcTemplateByName() (*VdcTemplate, error) {
	// /api/admin/extension/vdcTemplate/
	// loop per names
	return nil, nil
}

func (vdcTemplate *VdcTemplate) Update(input types.VMWVdcTemplate) (*VdcTemplate, error) {
	// PUT /api/admin/extension/vdcTemplate/876b75ff-1ee1-4508-a701-2d2ff3e6f662
	return nil, nil
}

func (vdcTemplate *VdcTemplate) Delete() error {
	// DELETE /api/admin/extension/vdcTemplate/876b75ff-1ee1-4508-a701-2d2ff3e6f662
	if vdcTemplate.VdcTemplate.HREF == "" {
		return fmt.Errorf("cannot delete the VDC Template, its HREF is empty")
	}

	_, err := vdcTemplate.client.ExecuteRequest(vdcTemplate.VdcTemplate.HREF, http.MethodDelete, "", "error deleting VDC Template: %s", nil, nil)
	if err != nil {
		return err
	}
	return nil
}
