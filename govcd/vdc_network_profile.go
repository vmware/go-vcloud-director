/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// VDC Network profiles have 1:1 mapping

// GetVdcNetworkProfile retrieves VDC Network Profile configuration
// vdc.Vdc.ID must be set and valid present
func (vdc *Vdc) GetVdcNetworkProfile() (*types.VdcNetworkProfile, error) {
	client := vdc.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcNetworkProfile
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if vdc == nil || vdc.Vdc == nil || vdc.Vdc.ID == "" {
		return nil, fmt.Errorf("cannot lookup VDC Network Profile configuration without VDC ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdc.Vdc.ID))
	if err != nil {
		return nil, err
	}

	returnObject := &types.VdcNetworkProfile{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject, nil)
	if err != nil {
		return nil, err
	}

	return returnObject, nil
}

// UpdateVdcNetworkProfile updates the VDC Network Profile configuration
//
// Note. It is VDC Network Profile configuration must have
func (vdc *Vdc) UpdateVdcNetworkProfile(vdcNetworkProfileConfig *types.VdcNetworkProfile) (*types.VdcNetworkProfile, error) {
	client := vdc.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcNetworkProfile
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if vdc == nil || vdc.Vdc == nil || vdc.Vdc.ID == "" {
		return nil, fmt.Errorf("cannot update VDC Network Profile configuration without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdc.Vdc.ID))
	if err != nil {
		return nil, err
	}

	returnObject := &types.VdcNetworkProfile{}
	err = client.OpenApiPutItem(apiVersion, urlRef, nil, vdcNetworkProfileConfig, returnObject, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating VDC Network Profile configuration: %s", err)
	}

	return returnObject, nil
}

// DeleteVdcNetworkProfile deletes VDC Network Profile Configuration
func (vdc *Vdc) DeleteVdcNetworkProfile() error {
	client := vdc.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcNetworkProfile
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if vdc == nil || vdc.Vdc == nil || vdc.Vdc.ID == "" {
		return fmt.Errorf("cannot lookup VDC Network Profile without VDC ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdc.Vdc.ID))
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting VDC Network Profile configuration: %s", err)
	}

	return nil
}
