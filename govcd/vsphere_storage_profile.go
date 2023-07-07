/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

/*
Note: These storage profile methods refer to storage profiles before they get assigned to a provider VDC.
This file, with related tests, was created before realizing that these calls do not retrieve the `*(Any)`
storage profile.
*/

// StorageProfile contains a storage profile in a given context (usually, a resource pool)
type StorageProfile struct {
	StorageProfile *types.OpenApiStorageProfile
	vcenter        *VCenter
	client         *VCDClient
}

// GetAllStorageProfiles retrieves all storage profiles existing in a given storage profile context
// Note: this function finds all *named* resource pools, but not the unnamed one [*(Any)]
func (vcenter VCenter) GetAllStorageProfiles(resourcePoolId string, queryParams url.Values) ([]*StorageProfile, error) {
	client := vcenter.client.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointStorageProfiles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vcenter.VSphereVCenter.VcId))
	if err != nil {
		return nil, err
	}

	retrieved := []*types.OpenApiStorageProfile{{}}

	if queryParams == nil {
		queryParams = url.Values{}
	}
	if resourcePoolId != "" {
		queryParams.Set("filter", fmt.Sprintf("_context==%s", resourcePoolId))
	}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParams, &retrieved, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting storage profile list: %s", err)
	}

	if len(retrieved) == 0 {
		return nil, nil
	}
	var returnList []*StorageProfile

	for _, sp := range retrieved {
		newSp := sp
		returnList = append(returnList, &StorageProfile{
			StorageProfile: newSp,
			vcenter:        &vcenter,
			client:         vcenter.client,
		})
	}
	return returnList, nil
}

// GetStorageProfileById retrieves a storage profile in the context of a given resource pool
func (vcenter VCenter) GetStorageProfileById(resourcePoolId, id string) (*StorageProfile, error) {
	storageProfiles, err := vcenter.GetAllStorageProfiles(resourcePoolId, nil)
	if err != nil {
		return nil, err
	}
	for _, sp := range storageProfiles {
		if sp.StorageProfile.Moref == id {
			return sp, nil
		}
	}
	return nil, fmt.Errorf("no storage profile found with ID '%s': %s", id, err)
}

// GetStorageProfileByName retrieves a storage profile in the context of a given resource pool
func (vcenter VCenter) GetStorageProfileByName(resourcePoolId, name string) (*StorageProfile, error) {
	storageProfiles, err := vcenter.GetAllStorageProfiles(resourcePoolId, nil)
	if err != nil {
		return nil, err
	}
	var found []*StorageProfile
	for _, sp := range storageProfiles {
		if sp.StorageProfile.Name == name {
			found = append(found, sp)
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("no storage profile found with name '%s': %s", name, ErrorEntityNotFound)
	}
	if len(found) > 1 {
		return nil, fmt.Errorf("more than one storage profile found with name '%s'", name)
	}
	return found[0], nil
}
