/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"net/url"
	"strings"
)

type VdcTemplate struct {
	VdcTemplate *types.VMWVdcTemplate
	client      *Client
}

// CreateVdcTemplate creates a VDC Template with the given settings.
func (vcdClient *VCDClient) CreateVdcTemplate(input types.VMWVdcTemplate) (*VdcTemplate, error) {
	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires System Administrator privileges")
	}
	href := vcdClient.Client.VCDHREF
	href.Path += "/admin/extension/vdcTemplates"

	result := &types.VMWVdcTemplate{}

	resp, err := vcdClient.Client.executeJsonRequest(href.String(), http.MethodPost, input, "error creating VDC Template: %s")
	defer closeBody(resp)
	if err != nil {
		return nil, err
	}

	vdcTemplate := VdcTemplate{
		VdcTemplate: result,
		client:      &vcdClient.Client,
	}

	err = decodeBody(types.BodyTypeJSON, resp, vdcTemplate.VdcTemplate)
	if err != nil {
		return nil, err
	}

	return &vdcTemplate, nil
}

func (vcdClient *VCDClient) GetVdcTemplateById(id string) (*VdcTemplate, error) {
	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires System Administrator privileges")
	}

	href := vcdClient.Client.VCDHREF
	href.Path += "/admin/extension/vdcTemplate/" + extractUuid(id)

	result := &types.VMWVdcTemplate{}
	resp, err := vcdClient.Client.executeJsonRequest(href.String(), http.MethodGet, nil, "error getting VDC Template: %s")
	defer closeBody(resp)

	if err != nil {
		if strings.Contains(err.Error(), "RESOURCE_NOT_FOUND") {
			return nil, fmt.Errorf("%s: %s", ErrorEntityNotFound, err)
		}
		return nil, err
	}

	vdcTemplate := VdcTemplate{
		VdcTemplate: result,
		client:      &vcdClient.Client,
	}

	err = decodeBody(types.BodyTypeJSON, resp, vdcTemplate.VdcTemplate)
	if err != nil {
		return nil, err
	}

	return &vdcTemplate, nil
}

// GetVdcTemplateByName retrieves the VDC Template with the given name
func (vcdClient *VCDClient) GetVdcTemplateByName(name string) (*VdcTemplate, error) {
	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires System Administrator privileges")
	}
	results, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type":         "adminOrgVdcTemplate",
		"filter":       fmt.Sprintf("name==%s", url.QueryEscape(name)),
		"filterEncode": "true",
	})
	if err != nil {
		return nil, err
	}
	if len(results.Results.AdminOrgVdcTemplateRecord) == 0 {
		return nil, fmt.Errorf("could not find any VDC Template with name '%s': %s", name, ErrorEntityNotFound)
	}
	if len(results.Results.AdminOrgVdcTemplateRecord) > 1 {
		return nil, fmt.Errorf("expected one VDC Template with name '%s', but got %d", name, len(results.Results.AdminOrgVdcTemplateRecord))
	}
	return vcdClient.GetVdcTemplateById(extractUuid(results.Results.AdminOrgVdcTemplateRecord[0].HREF))
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
