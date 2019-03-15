/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func GetExternalNetworkByName(vcdClient *VCDClient, networkName string) (*types.ExternalNetworkReference, error) {
	externalNetwork := NewExternalNetwork(&vcdClient.Client)
	err := externalNetwork.GetByName(networkName)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
	}
	return &types.ExternalNetworkReference{
		HREF: externalNetwork.ExternalNetwork.HREF,
		Type: externalNetwork.ExternalNetwork.Type,
		Name: externalNetwork.ExternalNetwork.Name,
	}, nil
}

func CreateExternalNetwork(vcdClient *VCDClient, externalNetwork *types.ExternalNetwork) (Task, error) {
	err := validateExternalNetwork(externalNetwork)
	if err != nil {
		return Task{}, err
	}

	output, err := xml.MarshalIndent(externalNetwork, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshalling xml: %s", err)
	}

	xmlData := bytes.NewBufferString(xml.Header + string(output))
	util.Logger.Printf("[TRACE] CreateExternalNetwork - xml payload: %s\n", xmlData)
	externalnetsHREF := vcdClient.Client.VCDHREF
	externalnetsHREF.Path += "/admin/extension/externalnets"
	req := vcdClient.Client.NewRequest(map[string]string{}, "POST", externalnetsHREF, xmlData)
	req.Header.Add("Content-Type", "application/vnd.vmware.admin.vmwexternalnet+xml")

	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[TRACE] error instantiating a new ExternalNetwork: %s", err)
		return Task{}, fmt.Errorf("error instantiating a new ExternalNetwork: %s", err)
	}

	task := NewTask(&vcdClient.Client)
	if err = decodeBody(resp, task.Task); err != nil {
		util.Logger.Printf("[TRACE] error decoding admin extension externalnets response: %s", err)
		return Task{}, fmt.Errorf("error decoding admin extension externalnets response: %s", err)
	}

	return *task, nil
}

func getExternalNetworkHref(client *Client) (string, error) {
	extensions, err := getExtension(client)
	if err != nil {
		return "", err
	}

	for _, extensionLink := range extensions.Link {
		if extensionLink.Type == "application/vnd.vmware.admin.vmwExternalNetworkReferences+xml" {
			return extensionLink.HREF, nil
		}
	}

	return "", errors.New("external network link isn't found")
}

func getExtension(client *Client) (*types.Extension, error) {
	extensions := &types.Extension{}

	extensionHREF := client.VCDHREF
	extensionHREF.Path += "/admin/extension/"
	req := client.NewRequest(map[string]string{}, "GET", extensionHREF, nil)
	resp, err := checkResp(client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[TRACE] error retrieving extension: %s", err)
		return extensions, fmt.Errorf("error retrieving extension: %s", err)
	}

	if err = decodeBody(resp, extensions); err != nil {
		util.Logger.Printf("[TRACE] error retrieving extension list: %s", err)
		return extensions, fmt.Errorf("error decoding extension list response: %s", err)
	}

	return extensions, nil
}

func queryVirtualCenters(vdcCli *VCDClient, filter string) ([]*types.QueryResultVirtualCenterRecordType, error) {
	results, err := vdcCli.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "virtualCenter",
		"filter": filter,
	})
	if err != nil {
		return nil, err
	}

	return results.Results.VirtualCenterRecord, nil
}
