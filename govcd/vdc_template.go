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
	href := vcdClient.Client.VCDHREF
	href.Path += "/admin/extension/vdcTemplates"

	return genericVdcTemplateRequest(&vcdClient.Client, input, &href, http.MethodPost)
}

// Update updates an existing VDC Template with the given settings
func (vdcTemplate *VdcTemplate) Update(input types.VMWVdcTemplate) (*VdcTemplate, error) {
	href := vdcTemplate.client.VCDHREF
	href.Path += fmt.Sprintf("/admin/extension/vdcTemplate/%s", extractUuid(vdcTemplate.VdcTemplate.ID))

	return genericVdcTemplateRequest(vdcTemplate.client, input, &href, http.MethodPut)
}

// genericVdcTemplateRequest creates or updates a VDC Template with the given settings
func genericVdcTemplateRequest(client *Client, input types.VMWVdcTemplate, href *url.URL, method string) (*VdcTemplate, error) {
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires System Administrator privileges")
	}

	result := &types.VMWVdcTemplate{}

	resp, err := client.executeJsonRequest(href.String(), method, input, "error when performing a "+method+" for VDC Template: %s")
	defer closeBody(resp)
	if err != nil {
		return nil, err
	}

	vdcTemplate := VdcTemplate{
		VdcTemplate: result,
		client:      client,
	}

	err = decodeBody(types.BodyTypeJSON, resp, vdcTemplate.VdcTemplate)
	if err != nil {
		return nil, err
	}

	return &vdcTemplate, nil
}

// GetVdcTemplateById retrieves the VDC Template with the given ID
func (vcdClient *VCDClient) GetVdcTemplateById(id string) (*VdcTemplate, error) {
	href := vcdClient.Client.VCDHREF
	href.Path += "/admin/extension/vdcTemplate/" + extractUuid(id)

	result := &types.VMWVdcTemplate{}
	resp, err := vcdClient.Client.executeJsonRequest(href.String(), http.MethodGet, nil, "error getting VDC Template: %s")
	defer closeBody(resp)

	if err != nil {
		if strings.Contains(err.Error(), "RESOURCE_NOT_FOUND") || strings.Contains(err.Error(), "not exist") {
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

// GetVdcTemplateByName retrieves the VDC Template with the given name.
// NOTE: 'name' refers to the "System name", not "Tenant name".
func (vcdClient *VCDClient) GetVdcTemplateByName(name string) (*VdcTemplate, error) {
	queryType := types.QtAdminOrgVdcTemplate
	if !vcdClient.Client.IsSysAdmin {
		queryType = types.QtOrgVdcTemplate
	}
	results, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type":         queryType,
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
	return vcdClient.GetVdcTemplateById(results.Results.AdminOrgVdcTemplateRecord[0].HREF)
}

// Delete deletes the receiver VDC Template
func (vdcTemplate *VdcTemplate) Delete() error {
	if !vdcTemplate.client.IsSysAdmin {
		return fmt.Errorf("functionality requires System Administrator privileges")
	}
	if vdcTemplate.VdcTemplate.HREF == "" {
		return fmt.Errorf("cannot delete the VDC Template, its HREF is empty")
	}

	_, err := vdcTemplate.client.ExecuteRequest(vdcTemplate.VdcTemplate.HREF, http.MethodDelete, "", "error deleting VDC Template: %s", nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// SetAccess sets the Access control configuration for the receiver VDC Template,
// which specifies which Organizations can read it.
func (vdcTemplate *VdcTemplate) SetAccess(organizationIds []string) error {
	if !vdcTemplate.client.IsSysAdmin {
		return fmt.Errorf("functionality requires System Administrator privileges")
	}
	if vdcTemplate.VdcTemplate.HREF == "" {
		return fmt.Errorf("cannot delete the VDC Template, its HREF is empty")
	}
	accessSettings := make([]*types.AccessSetting, len(organizationIds))
	for i, organizationId := range organizationIds {
		accessSettings[i] = &types.AccessSetting{
			Subject: &types.LocalSubject{
				HREF: fmt.Sprintf("%s/org/%s", vdcTemplate.client.VCDHREF.String(), extractUuid(organizationId))},
			AccessLevel: types.ControlAccessReadOnly,
		}
	}
	payload := &types.ControlAccessParams{AccessSettings: &types.AccessSettingList{AccessSetting: accessSettings}}

	return vdcTemplate.client.setAccessControlWithHttpMethod(http.MethodPut, payload, vdcTemplate.VdcTemplate.HREF, "VDC Template", vdcTemplate.VdcTemplate.Name, nil)
}

// GetAccess retrieves the Control access configuration for the receiver VDC Template, which
// contains the Organizations that can read it.
func (vdcTemplate *VdcTemplate) GetAccess() (*types.ControlAccessParams, error) {
	if !vdcTemplate.client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires System Administrator privileges")
	}
	if vdcTemplate.VdcTemplate.HREF == "" {
		return nil, fmt.Errorf("cannot delete the VDC Template, its HREF is empty")
	}
	result := &types.ControlAccessParams{}
	href := fmt.Sprintf("%s/controlAccess", vdcTemplate.VdcTemplate.HREF)
	_, err := vdcTemplate.client.ExecuteRequest(href, http.MethodGet, types.AnyXMLMime, "error getting access of VDC Template: %s", nil, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Instantiate creates a new VDC from the template and returns its ID if the operation finishes successfully.
func (vdcTemplate *VdcTemplate) Instantiate(vdcName, description, organizationId string) (string, error) {
	if vdcName == "" {
		return "", fmt.Errorf("the VDC name is required to instantiate VDC Template '%s'", vdcTemplate.VdcTemplate.Name)
	}
	if organizationId == "" {
		return "", fmt.Errorf("the Organization ID is required to instantiate VDC Template '%s'", vdcTemplate.VdcTemplate.Name)
	}

	payload := &types.InstantiateVdcTemplateParams{
		Xmlns: types.XMLNamespaceVCloud,
		Name:  vdcName,
		Source: &types.Reference{
			HREF: vdcTemplate.VdcTemplate.HREF,
			Type: types.MimeVdcTemplateInstantiateType,
		},
	}
	if description != "" {
		payload.Description = description
	}

	href := vdcTemplate.client.VCDHREF
	href.Path += fmt.Sprintf("/org/%s/action/instantiate", extractUuid(organizationId))
	task, err := vdcTemplate.client.ExecuteTaskRequest(href.String(), http.MethodPost, types.MimeVdcTemplateInstantiate, "error getting access of VDC Template: %s", payload)
	if err != nil {
		return "", err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return "", err
	}
	if task.Task.Owner == nil {
		return "", fmt.Errorf("the VDC was instantiated but could not retrieve its ID from the finished task")
	}
	return task.Task.Owner.ID, nil
}
