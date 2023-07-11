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

// Deprecated: wrong implementation and result
// Use vdc.CreateVappFromTemplate instead
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

// GetVappTemplateRecord gets the corresponding vApp template record
func (vAppTemplate *VAppTemplate) GetVappTemplateRecord() (*types.QueryResultVappTemplateType, error) {
	queriedVappTemplates, err := queryVappTemplateListWithFilter(vAppTemplate.client, map[string]string{
		"id": vAppTemplate.VAppTemplate.ID,
	})
	if err != nil {
		return nil, err
	}
	if len(queriedVappTemplates) != 1 {
		return nil, fmt.Errorf("found %d vApp Templates with ID %s", len(queriedVappTemplates), vAppTemplate.VAppTemplate.ID)
	}
	return queriedVappTemplates[0], nil
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

// GetVAppTemplateByHref finds a vApp template by HREF
// On success, returns a pointer to the vApp template structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetVAppTemplateByHref(href string) (*VAppTemplate, error) {
	return getVAppTemplateByHref(&vcdClient.Client, href)
}

// GetVAppTemplateById finds a vApp Template by ID.
// On success, returns a pointer to the VAppTemplate structure and a nil error.
// On failure, returns a nil pointer and an error.
func (vcdClient *VCDClient) GetVAppTemplateById(vAppTemplateId string) (*VAppTemplate, error) {
	return getVAppTemplateById(&vcdClient.Client, vAppTemplateId)
}

// QuerySynchronizedVAppTemplateById Finds a vApp Template by its URN that is synchronized in the catalog.
// Returns types.QueryResultVMRecordType if it is found, returns ErrorEntityNotFound if not found, or an error if many are
// found.
func (vcdClient *VCDClient) QuerySynchronizedVAppTemplateById(vAppTemplateId string) (*types.QueryResultVappTemplateType, error) {
	queryType := types.QtVappTemplate
	if vcdClient.Client.IsSysAdmin {
		queryType = types.QtAdminVappTemplate
	}

	// this allows to query deployed and not deployed templates
	results, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type": queryType,
		"filter": "id==" + url.QueryEscape(extractUuid(vAppTemplateId)) +
			";status!=FAILED_CREATION;status!=UNKNOWN;status!=UNRECOGNIZED;status!=UNRESOLVED;status!=LOCAL_COPY_UNAVAILABLE&links=true",
		"filterEncoded": "true"})
	if err != nil {
		return nil, fmt.Errorf("[QueryVAppTemplateById] error quering vApp templates with ID %s: %s", vAppTemplateId, err)
	}

	vAppTemplateRecords := results.Results.VappTemplateRecord
	if vcdClient.Client.IsSysAdmin {
		vAppTemplateRecords = results.Results.AdminVappTemplateRecord
	}
	if len(vAppTemplateRecords) == 0 {
		return nil, ErrorEntityNotFound
	}

	if len(vAppTemplateRecords) > 1 {
		return nil, fmt.Errorf("[QueryVmInVAppTemplateByHref] found %d results with with ID: %s", len(vAppTemplateRecords), vAppTemplateId)
	}

	return vAppTemplateRecords[0], nil
}

// QueryVmInVAppTemplateByHref Finds a VM inside a vApp Template using the latter HREF.
// Returns types.QueryResultVMRecordType if it is found, returns ErrorEntityNotFound if not found, or an error if many are
// found.
func (vcdClient *VCDClient) QueryVmInVAppTemplateByHref(vAppTemplateHref, vmNameInTemplate string) (*types.QueryResultVMRecordType, error) {
	queryType := types.QtVm
	if vcdClient.Client.IsSysAdmin {
		queryType = types.QtAdminVm
	}

	// this allows to query deployed and not deployed templates
	results, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type": queryType,
		"filter": "container==" + url.QueryEscape(vAppTemplateHref) + ";name==" + url.QueryEscape(vmNameInTemplate) +
			";isVAppTemplate==true;status!=FAILED_CREATION;status!=UNKNOWN;status!=UNRECOGNIZED;status!=UNRESOLVED&links=true;",
		"filterEncoded": "true"})
	if err != nil {
		return nil, fmt.Errorf("[QueryVmInVAppTemplateByHref] error quering vApp templates with HREF %s:, VM name: %s: Error: %s", vAppTemplateHref, vmNameInTemplate, err)
	}

	vmResults := results.Results.VMRecord
	if vcdClient.Client.IsSysAdmin {
		vmResults = results.Results.AdminVMRecord
	}

	if len(vmResults) == 0 {
		return nil, ErrorEntityNotFound
	}

	if len(vmResults) > 1 {
		return nil, fmt.Errorf("[QueryVmInVAppTemplateByHref] found %d results with with HREF: %s, VM name: %s", len(vmResults), vAppTemplateHref, vmNameInTemplate)
	}

	return vmResults[0], nil
}

// QuerySynchronizedVmInVAppTemplateByHref Finds a catalog-synchronized VM inside a vApp Template using the latter HREF.
// Returns types.QueryResultVMRecordType if it is found and it's synchronized in the catalog.
// Returns ErrorEntityNotFound if not found, or an error if many are found.
func (vcdClient *VCDClient) QuerySynchronizedVmInVAppTemplateByHref(vAppTemplateHref, vmNameInTemplate string) (*types.QueryResultVMRecordType, error) {
	vmRecord, err := vcdClient.QueryVmInVAppTemplateByHref(vAppTemplateHref, vmNameInTemplate)
	if err != nil {
		return nil, err
	}
	if vmRecord.Status == "LOCAL_COPY_UNAVAILABLE" {
		return nil, fmt.Errorf("vApp template %s is not synchronized", extractUuid(vAppTemplateHref))
	}
	return vmRecord, nil
}
