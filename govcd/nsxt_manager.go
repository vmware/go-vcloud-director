/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type NsxtManager struct {
	NsxtManager *types.NsxtManager
	VCDClient   *VCDClient
	// Urn holds a URN value for NSX-T manager. None of the API endpoints return it, but filtering other entities requires that
	// Sample format: urn:vcloud:nsxtmanager:UUID
	//
	// Note:  this is being computed when retrieving the structure and will not be populated if this structure is initialized manually
	Urn string
}

// GetNsxtManagerByName searches for NSX-T managers available in VCD and returns the one that
// matches name
func (vcdClient *VCDClient) GetNsxtManagerByName(name string) (*NsxtManager, error) {
	nsxtManagers, err := vcdClient.QueryNsxtManagerByName(name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Manager by name '%s': %s", name, err)
	}

	// Double check that exactly one NSX-T Manager is found and throw error otherwise
	singleNsxtManager, err := oneOrError("name", name, nsxtManagers)
	if err != nil {
		return nil, err
	}

	resp, err := vcdClient.Client.executeJsonRequest(singleNsxtManager.HREF, http.MethodGet, nil, "error retrieving NSX-T Manager: %s")
	if err != nil {
		return nil, err
	}

	defer closeBody(resp)

	nsxtManager := NsxtManager{
		NsxtManager: &types.NsxtManager{},
		VCDClient:   vcdClient,
	}

	err = decodeBody(types.BodyTypeJSON, resp, nsxtManager.NsxtManager)
	if err != nil {
		return nil, err
	}

	// Populate computed URN field, which is not usually provided in API, but is useful in other
	// endpoints
	nsxtManager.Urn = nsxtManager.NsxtManager.ID
	if !isUrn(nsxtManager.NsxtManager.ID) {
		nsxtManagerUrn, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", nsxtManager.NsxtManager.ID)
		if err != nil {
			return nil, fmt.Errorf("error building NSX-T Manager URN from ID '%s': %s", nsxtManager.NsxtManager.ID, err)
		}
		nsxtManager.Urn = nsxtManagerUrn
	}

	return &nsxtManager, nil
}
