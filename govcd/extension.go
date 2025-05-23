// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/http"
)

// Deprecated: please use GetExternalNetwork function instead
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
