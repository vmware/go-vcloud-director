/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
)

type ExternalNetwork struct {
	ExternalNetwork *types.ExternalNetwork
	client          *Client
}

func NewExternalNetwork(cli *Client) *ExternalNetwork {
	return &ExternalNetwork{
		ExternalNetwork: new(types.ExternalNetwork),
		client:          cli,
	}
}

func (externalNetwork ExternalNetwork) GetByName(networkName string) error {
	extNetworkHREF, err := getExternalNetworkHref(externalNetwork.client)
	if err != nil {
		return err
	}

	extNetworkURL, err := url.ParseRequestURI(extNetworkHREF)
	if err != nil {
		return err
	}

	req := externalNetwork.client.NewRequest(map[string]string{}, "GET", *extNetworkURL, nil)
	resp, err := checkResp(externalNetwork.client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[TRACE] error retrieving external networks: %s", err)
		return fmt.Errorf("error retrieving external networks: %s", err)
	}

	extNetworkRefs := &types.ExternalNetworkReferences{}
	if err = decodeBody(resp, extNetworkRefs); err != nil {
		util.Logger.Printf("[TRACE] error retrieving external networks: %s", err)
		return fmt.Errorf("error decoding extension external networks: %s", err)
	}

	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		if netRef.Name == networkName {
			externalNetwork.ExternalNetwork.HREF = netRef.HREF
			return externalNetwork.Refresh()
		}
	}

	return fmt.Errorf("external network %s not found", networkName)
}

func (externalNetwork ExternalNetwork) Refresh() error {
	extNetworkURL, err := url.ParseRequestURI(externalNetwork.ExternalNetwork.HREF)
	if err != nil {
		return err
	}

	req := externalNetwork.client.NewRequest(map[string]string{}, "GET", *extNetworkURL, nil)
	resp, err := checkResp(externalNetwork.client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[TRACE] error retrieving external network: %s", err)
		return fmt.Errorf("error retrieving external network: %s", err)
	}

	if err = decodeBody(resp, externalNetwork.ExternalNetwork); err != nil {
		util.Logger.Printf("[TRACE] error decoding extension external network: %s", err)
		return fmt.Errorf("error decoding extension external network: %s", err)
	}
	return nil
}

func validateExternalNetwork(externalNetwork *types.ExternalNetwork) error {
	if externalNetwork.Name == "" {
		return errors.New("External Network missing required field: Name")
	}
	if externalNetwork.Xmlns == "" {
		return errors.New("External Network missing required field: Xmlns")
	}
	return nil
}

func (externalNetwork *ExternalNetwork) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] ExternalNetwork.Delete")
	externalNetworkUrl, err := url.ParseRequestURI(externalNetwork.ExternalNetwork.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parsing external network url: %s", err)
	}

	req := externalNetwork.client.NewRequest(map[string]string{}, "DELETE", *externalNetworkUrl, nil)
	resp, err := checkResp(externalNetwork.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error deleting external network: %s", err)
	}

	task := NewTask(externalNetwork.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	return *task, nil
}

func (externalNetwork *ExternalNetwork) DeleteWait() error {
	task, err := externalNetwork.Delete()
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("couldn't finish removing external network %#v", err)
	}
	return nil
}
