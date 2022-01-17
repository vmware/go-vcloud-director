/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type VAppTemplate struct {
	VAppTemplate *types.VAppTemplate
	client       *Client
}

func NewVAppTemplate(cli *Client) *VAppTemplate {
	return &VAppTemplate{
		VAppTemplate: new(types.VAppTemplate),
		client:       cli,
	}
}

func (vdc *Vdc) InstantiateVAppTemplate(template *types.InstantiateVAppTemplateParams) error {
	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return fmt.Errorf("error getting vdc href: %s", err)
	}
	vdcHref.Path += "/action/instantiateVAppTemplate"

	vapptemplate := NewVAppTemplate(vdc.client)

	_, err = vdc.client.ExecuteRequest(vdcHref.String(), http.MethodPut,
		types.MimeInstantiateVappTemplateParams, "error instantiating a new template: %s", template, vapptemplate)
	if err != nil {
		return err
	}

	task := NewTask(vdc.client)
	for _, taskItem := range vapptemplate.VAppTemplate.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error performing task: %s", err)
		}
	}
	return nil
}

// Refresh refreshes the vApp template item information by href
func (vAppTemplate *VAppTemplate) Refresh() error {

	if vAppTemplate.VAppTemplate == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	url := vAppTemplate.VAppTemplate.HREF
	if url == "nil" {
		return fmt.Errorf("cannot refresh, HREF is empty")
	}

	vAppTemplate.VAppTemplate = &types.VAppTemplate{}

	_, err := vAppTemplate.client.ExecuteRequest(url, http.MethodGet,
		"", "error retrieving vApp template item: %s", nil, vAppTemplate.VAppTemplate)

	return err
}

// Update updates the vApp template item information
// Returns vApp template and error.
func (vAppTemplate *VAppTemplate) Update() (*VAppTemplate, error) {
	if vAppTemplate.VAppTemplate == nil {
		return nil, fmt.Errorf("cannot update, Object is empty")
	}

	url := vAppTemplate.VAppTemplate.HREF
	if url == "nil" {
		return nil, fmt.Errorf("cannot update, HREF is empty")
	}

	task, err := vAppTemplate.UpdateAsync()
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error waiting for task completion after updating vApp template %s: %s", vAppTemplate.VAppTemplate.Name, err)
	}
	err = vAppTemplate.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing vApp template %s: %s", vAppTemplate.VAppTemplate.Name, err)
	}
	return vAppTemplate, nil
}

// UpdateAsync updates the vApp template item information
// Returns Task and error.
func (vAppTemplate *VAppTemplate) UpdateAsync() (Task, error) {

	if vAppTemplate.VAppTemplate == nil {
		return Task{}, fmt.Errorf("cannot update, Object is empty")
	}

	url := vAppTemplate.VAppTemplate.HREF
	if url == "nil" {
		return Task{}, fmt.Errorf("cannot update, HREF is empty")
	}

	vappTemplatePayload := types.VAppTemplateForUpdate{
		Xmlns:       types.XMLNamespaceVCloud,
		HREF:        vAppTemplate.VAppTemplate.HREF,
		ID:          vAppTemplate.VAppTemplate.ID,
		Name:        vAppTemplate.VAppTemplate.Name,
		GoldMaster:  vAppTemplate.VAppTemplate.GoldMaster,
		Description: vAppTemplate.VAppTemplate.Description,
		Link:        vAppTemplate.VAppTemplate.Link,
	}

	return vAppTemplate.client.ExecuteTaskRequest(url, http.MethodPut,
		types.MimeVAppTemplate, "error updating vApp template item: %s", vappTemplatePayload)
}
