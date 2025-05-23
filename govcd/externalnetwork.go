// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"errors"
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
	"net/http"
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

	return "", errors.New("external network link wasn't found")
}

func (externalNetwork ExternalNetwork) Refresh() error {

	if !externalNetwork.client.IsSysAdmin {
		return fmt.Errorf("functionality requires System Administrator privileges")
	}

	_, err := externalNetwork.client.ExecuteRequest(externalNetwork.ExternalNetwork.HREF, http.MethodGet,
		"", "error refreshing external network: %s", nil, externalNetwork.ExternalNetwork)

	return err
}

func validateExternalNetwork(externalNetwork *types.ExternalNetwork) error {
	if externalNetwork.Name == "" {
		return errors.New("external Network missing required field: Name")
	}
	return nil
}

func (externalNetwork *ExternalNetwork) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] ExternalNetwork.Delete")

	if !externalNetwork.client.IsSysAdmin {
		return Task{}, fmt.Errorf("functionality requires System Administrator privileges")
	}

	// Return the task
	return externalNetwork.client.ExecuteTaskRequest(externalNetwork.ExternalNetwork.HREF, http.MethodDelete,
		"", "error deleting external network: %s", nil)
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
