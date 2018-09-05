/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
	"net/http"
	neturl "net/url"
	"regexp"
	"strings"
	"time"
)

// OrgVDCNetwork an org vdc network client
type OrgVDCNetwork struct {
	OrgVDCNetwork *types.OrgVDCNetwork
	client        *Client
}

// NewOrgVDCNetwork creates an org vdc network client
func NewOrgVDCNetwork(c *Client) *OrgVDCNetwork {
	return &OrgVDCNetwork{
		OrgVDCNetwork: new(types.OrgVDCNetwork),
		client:        c,
	}
}

func (orgVdcNet *OrgVDCNetwork) Refresh() error {
	if orgVdcNet.OrgVDCNetwork.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	url, _ := neturl.ParseRequestURI(orgVdcNet.OrgVDCNetwork.HREF)

	req := orgVdcNet.client.NewRequest(map[string]string{}, "GET", *url, nil)

	resp, err := checkResp(orgVdcNet.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error retrieving task: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	orgVdcNet.OrgVDCNetwork = &types.OrgVDCNetwork{}

	if err = decodeBody(resp, orgVdcNet.OrgVDCNetwork); err != nil {
		return fmt.Errorf("error decoding task response: %s", err)
	}

	// The request was successful
	return nil
}

func (orgVdcNet *OrgVDCNetwork) Delete() (Task, error) {
	err := orgVdcNet.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("Error refreshing network: %s", err)
	}
	pathArr := strings.Split(orgVdcNet.OrgVDCNetwork.HREF, "/")
	apiEndpoint, _ := neturl.ParseRequestURI(orgVdcNet.OrgVDCNetwork.HREF)
	apiEndpoint.Path = "/api/admin/network/" + pathArr[len(pathArr)-1]

	var resp *http.Response
	for {
		req := orgVdcNet.client.NewRequest(map[string]string{}, "DELETE", *apiEndpoint, nil)
		resp, err = checkResp(orgVdcNet.client.Http.Do(req))
		if err != nil {
			if match, _ := regexp.MatchString("is busy, cannot proceed with the operation.$", err.Error()); match {
				time.Sleep(3 * time.Second)
				continue
			}
			return Task{}, fmt.Errorf("error deleting Network: %s", err)
		}
		break
	}

	task := NewTask(orgVdcNet.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (vdc *Vdc) CreateOrgVDCNetwork(networkConfig *types.OrgVDCNetwork) error {
	for _, av := range vdc.Vdc.Link {
		if av.Rel == "add" && av.Type == "application/vnd.vmware.vcloud.orgVdcNetwork+xml" {
			url, err := neturl.ParseRequestURI(av.HREF)
			//return fmt.Errorf("Test output: %#v")

			if err != nil {
				return fmt.Errorf("error decoding vdc response: %s", err)
			}

			output, err := xml.MarshalIndent(networkConfig, "  ", "    ")
			if err != nil {
				return fmt.Errorf("error marshaling OrgVDCNetwork compose: %s", err)
			}

			//return fmt.Errorf("Test output: %s\n%#v", b, v.c)

			var resp *http.Response
			for {
				buffer := bytes.NewBufferString(xml.Header + string(output))
				util.GovcdLogger.Printf("[DEBUG] VCD Client configuration: %s", buffer)
				req := vdc.client.NewRequest(map[string]string{}, "POST", *url, buffer)
				req.Header.Add("Content-Type", av.Type)
				resp, err = checkResp(vdc.client.Http.Do(req))
				if err != nil {
					if match, _ := regexp.MatchString("is busy, cannot proceed with the operation.$", err.Error()); match {
						time.Sleep(3 * time.Second)
						continue
					}
					return fmt.Errorf("error instantiating a new OrgVDCNetwork: %s", err)
				}
				break
			}
			newstuff := NewOrgVDCNetwork(vdc.client)
			if err = decodeBody(resp, newstuff.OrgVDCNetwork); err != nil {
				return fmt.Errorf("error decoding orgvdcnetwork response: %s", err)
			}
			task := NewTask(vdc.client)
			for _, t := range newstuff.OrgVDCNetwork.Tasks.Task {
				task.Task = t
				err = task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf("Error performing task: %#v", err)
				}
			}
		}
	}
	return nil
}
