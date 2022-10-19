/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net/http"
	"net/url"
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
		types.MimeInstantiateVappTemplateParams, "error instantiating a new vApp Template: %s", template, vapptemplate)
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
		"", "error retrieving vApp Template: %s", nil, vAppTemplate.VAppTemplate)

	return err
}

// GetCatalogName gets the catalog name to which the receiver vApp Template belongs
func (vAppTemplate *VAppTemplate) GetCatalogName() (string, error) {
	queriedVappTemplates, err := queryVappTemplateListWithFilter(vAppTemplate.client, map[string]string{
		"id": vAppTemplate.VAppTemplate.ID,
	})
	if err != nil {
		return "", err
	}
	if len(queriedVappTemplates) != 1 {
		return "", fmt.Errorf("found %d vApp Templates with ID %s", len(queriedVappTemplates), vAppTemplate.VAppTemplate.ID)
	}
	return queriedVappTemplates[0].CatalogName, nil
}

// GetVdcName gets the VDC name to which the receiver vApp Template belongs
func (vAppTemplate *VAppTemplate) GetVdcName() (string, error) {
	queriedVappTemplates, err := queryVappTemplateListWithFilter(vAppTemplate.client, map[string]string{
		"id": vAppTemplate.VAppTemplate.ID,
	})
	if err != nil {
		return "", err
	}
	if len(queriedVappTemplates) != 1 {
		return "", fmt.Errorf("found %d vApp Templates with ID %s", len(queriedVappTemplates), vAppTemplate.VAppTemplate.ID)
	}
	return queriedVappTemplates[0].VdcName, nil
}

// Update updates the vApp template item information.
// VCD also updates the associated Catalog Item, in order to be in sync with the receiver vApp Template entity.
// For example, updating a vApp Template name "A" to "B" will make VCD to also update the Catalog Item to be renamed to "B".
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
		return nil, fmt.Errorf("error waiting for task completion after updating vApp Template %s: %s", vAppTemplate.VAppTemplate.Name, err)
	}
	err = vAppTemplate.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing vApp Template %s: %s", vAppTemplate.VAppTemplate.Name, err)
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
		types.MimeVAppTemplate, "error updating vApp Template: %s", vappTemplatePayload)
}

// DeleteAsync deletes the VAppTemplate, returning the Task that monitors the deletion process, or an error
// if something wrong happened.
func (vAppTemplate *VAppTemplate) DeleteAsync() (Task, error) {
	util.Logger.Printf("[TRACE] Deleting vApp Template: %#v", vAppTemplate.VAppTemplate)

	vappTemplateHref := vAppTemplate.client.VCDHREF
	vappTemplateHref.Path += "/vAppTemplate/vappTemplate-" + extractUuid(vAppTemplate.VAppTemplate.ID)

	util.Logger.Printf("[TRACE] Url for deleting vApp Template: %#v and name: %s", vappTemplateHref, vAppTemplate.VAppTemplate.Name)

	return vAppTemplate.client.ExecuteTaskRequest(vappTemplateHref.String(), http.MethodDelete,
		"", "error deleting vApp Template: %s", nil)
}

// Delete deletes the VAppTemplate and waits for the deletion to finish, returning an error if something wrong happened.
func (vAppTemplate *VAppTemplate) Delete() error {
	task, err := vAppTemplate.DeleteAsync()
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for task completion after deleting vApp Template %s: %s", vAppTemplate.VAppTemplate.Name, err)
	}
	return nil
}
