/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
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

func (vdc *Vdc) InstantiateVAppTemplate(ctx context.Context, template *types.InstantiateVAppTemplateParams) error {
	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return fmt.Errorf("error getting vdc href: %s", err)
	}
	vdcHref.Path += "/action/instantiateVAppTemplate"

	var vapp types.VApp

	_, err = vdc.client.ExecuteRequest(ctx, vdcHref.String(), http.MethodPost,
		types.MimeInstantiateVappTemplateParams, "error instantiating a new template: %s", template, &vapp)
	if err != nil {
		return err
	}

	task := NewTask(vdc.client)
	for _, taskItem := range vapp.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion(ctx)
		if err != nil {
			return fmt.Errorf("error performing task: %s", err)
		}
	}
	return nil
}

// Refresh refreshes the vApp template item information by href
func (vAppTemplate *VAppTemplate) Refresh(ctx context.Context) error {

	if vAppTemplate.VAppTemplate == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	url := vAppTemplate.VAppTemplate.HREF
	if url == "nil" {
		return fmt.Errorf("cannot refresh, HREF is empty")
	}

	vAppTemplate.VAppTemplate = &types.VAppTemplate{}

	_, err := vAppTemplate.client.ExecuteRequest(ctx, url, http.MethodGet,
		"", "error retrieving vApp template item: %s", nil, vAppTemplate.VAppTemplate)

	return err
}
