/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func GetExternalNetworkByName(vcdClient *VCDClient, networkName string) (*types.ExternalNetworkReference, error) {
	extNetworkRefs := &types.ExternalNetworkReferences{}

	extNetworkHREF, err := getExternalNetworkHref(vcdClient)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
	}

	extNetworkURL, err := url.ParseRequestURI(extNetworkHREF)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
	}

	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", *extNetworkURL, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[TRACE] error retrieving external networks: %s", err)
		return &types.ExternalNetworkReference{}, fmt.Errorf("error retrieving external networks: %s", err)
	}

	if err = decodeBody(resp, extNetworkRefs); err != nil {
		util.Logger.Printf("[TRACE] error retrieving  external networks: %s", err)
		return &types.ExternalNetworkReference{}, fmt.Errorf("error decoding extension  external networks: %s", err)
	}

	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		if netRef.Name == networkName {
			return netRef, nil
		}
	}

	return &types.ExternalNetworkReference{}, nil
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

	externalNetwork = new(types.ExternalNetwork)
	if err = decodeBody(resp, externalNetwork); err != nil {
		util.Logger.Printf("[TRACE] error decoding admin extension externalnets response: %s", err)
		return Task{}, fmt.Errorf("error decoding admin extension externalnets response: %s", err)
	}

	task := NewTask(&vcdClient.Client)
	task.Task = externalNetwork.Tasks.Task[0]
	return *task, nil
}

func getExternalNetworkHref(vcdClient *VCDClient) (string, error) {
	extensions, err := getExtension(vcdClient)
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

func getExtension(vcdClient *VCDClient) (*types.Extension, error) {
	extensions := &types.Extension{}

	extensionHREF := vcdClient.Client.VCDHREF
	extensionHREF.Path += "/admin/extension/"
	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", extensionHREF, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
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
