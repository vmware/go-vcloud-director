/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
)

// DEPRECATED please use GetExternalNetwork function instead
func GetExternalNetworkByName(vcdClient *VCDClient, networkName string) (*types.ExternalNetworkReference, error) {
	extNetworkRefs := &types.ExternalNetworkReferences{}

	extNetworkHREF, err := getExternalNetworkHref(&vcdClient.Client)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
	}

	_, err = vcdClient.Client.ExecuteRequest(extNetworkHREF, http.MethodGet,
		"", "error retrieving external networks: %s", nil, extNetworkRefs)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
	}

	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		if netRef.Name == networkName {
			return netRef, nil
		}
	}

	return &types.ExternalNetworkReference{}, nil
}

// If user specifies a valid external network name, then this returns a
// ExternalNetwork object. If no valid external network is found, it returns an empty
// ExternalNetwork and no error. Otherwise it returns an error and an empty
// ExternalNetwork object
func GetExternalNetwork(vcdClient *VCDClient, networkName string) (*ExternalNetwork, error) {

	if !vcdClient.Client.IsSysAdmin {
		return &ExternalNetwork{}, fmt.Errorf("functionality requires system administrator privileges")
	}

	extNetworkHREF, err := getExternalNetworkHref(&vcdClient.Client)
	if err != nil {
		return &ExternalNetwork{}, err
	}

	extNetworkRefs := &types.ExternalNetworkReferences{}
	_, err = vcdClient.Client.ExecuteRequest(extNetworkHREF, http.MethodGet,
		types.MimeNetworkConnectionSection, "error retrieving external networks: %s", nil, extNetworkRefs)
	if err != nil {
		return &ExternalNetwork{}, err
	}

	externalNetwork := NewExternalNetwork(&vcdClient.Client)

	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		if netRef.Name == networkName {
			externalNetwork.ExternalNetwork.HREF = netRef.HREF
			err = externalNetwork.Refresh()
			if err != nil {
				return &ExternalNetwork{}, err
			}
		}
	}

	return externalNetwork, nil

}

func CreateExternalNetwork(vcdClient *VCDClient, externalNetwork *types.ExternalNetwork) (Task, error) {

	if !vcdClient.Client.IsSysAdmin {
		return Task{}, fmt.Errorf("functionality requires system administrator privileges")
	}

	err := validateExternalNetwork(externalNetwork)
	if err != nil {
		return Task{}, err
	}

	externalNetHREF := vcdClient.Client.VCDHREF
	externalNetHREF.Path += "/admin/extension/externalnets"

	// Return the task
	return vcdClient.Client.ExecuteTaskRequest(externalNetHREF.String(), http.MethodPost,
		types.MimeExternalNetwork, "error instantiating a new ExternalNetwork: %s", externalNetwork)
}

func getExtension(client *Client) (*types.Extension, error) {
	extensions := &types.Extension{}

	extensionHREF := client.VCDHREF
	extensionHREF.Path += "/admin/extension/"

	_, err := client.ExecuteRequest(extensionHREF.String(), http.MethodGet,
		"", "error retrieving extension: %s", nil, extensions)

	return extensions, err
}
