/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelVdcNetworkProfile = "VDC Network Profile"

// VDC Network profiles have 1:1 mapping with VDC - each VDC has an option to configure VDC Network
// Profiles. types.VdcNetworkProfile holds more information about possible configurations

// GetVdcNetworkProfile retrieves VDC Network Profile configuration
// vdc.Vdc.ID must be set and valid present
func (vdc *Vdc) GetVdcNetworkProfile() (*types.VdcNetworkProfile, error) {
	if vdc == nil || vdc.Vdc == nil || vdc.Vdc.ID == "" {
		return nil, fmt.Errorf("cannot lookup VDC Network Profile configuration without VDC ID")
	}

	return getVdcNetworkProfile(vdc.client, vdc.Vdc.ID)
}

// GetVdcNetworkProfile retrieves VDC Network Profile configuration
// vdc.Vdc.ID must be set and valid present
func (adminVdc *AdminVdc) GetVdcNetworkProfile() (*types.VdcNetworkProfile, error) {
	if adminVdc == nil || adminVdc.AdminVdc == nil || adminVdc.AdminVdc.ID == "" {
		return nil, fmt.Errorf("cannot lookup VDC Network Profile configuration without VDC ID")
	}

	return getVdcNetworkProfile(adminVdc.client, adminVdc.AdminVdc.ID)
}

// UpdateVdcNetworkProfile updates the VDC Network Profile configuration
//
// Note. Whenever updating VDC Network Profile it is required to send all fields (not only the
// changed ones) as VCD will remove other configuration. Best practice is to fetch current
// configuration of VDC Network Profile using GetVdcNetworkProfile, alter it with new values and
// submit it to UpdateVdcNetworkProfile.
func (vdc *Vdc) UpdateVdcNetworkProfile(vdcNetworkProfileConfig *types.VdcNetworkProfile) (*types.VdcNetworkProfile, error) {
	if vdc == nil || vdc.Vdc == nil || vdc.Vdc.ID == "" {
		return nil, fmt.Errorf("cannot update VDC Network Profile configuration without ID")
	}

	return updateVdcNetworkProfile(vdc.client, vdc.Vdc.ID, vdcNetworkProfileConfig)
}

// UpdateVdcNetworkProfile updates the VDC Network Profile configuration
func (adminVdc *AdminVdc) UpdateVdcNetworkProfile(vdcNetworkProfileConfig *types.VdcNetworkProfile) (*types.VdcNetworkProfile, error) {
	if adminVdc == nil || adminVdc.AdminVdc == nil || adminVdc.AdminVdc.ID == "" {
		return nil, fmt.Errorf("cannot update VDC Network Profile configuration without ID")
	}

	return updateVdcNetworkProfile(adminVdc.client, adminVdc.AdminVdc.ID, vdcNetworkProfileConfig)
}

// DeleteVdcNetworkProfile deletes VDC Network Profile Configuration
func (vdc *Vdc) DeleteVdcNetworkProfile() error {
	if vdc == nil || vdc.Vdc == nil || vdc.Vdc.ID == "" {
		return fmt.Errorf("cannot lookup VDC Network Profile without VDC ID")
	}

	return deleteVdcNetworkProfile(vdc.client, vdc.Vdc.ID)
}

// DeleteVdcNetworkProfile deletes VDC Network Profile Configuration
func (adminVdc *AdminVdc) DeleteVdcNetworkProfile() error {
	if adminVdc == nil || adminVdc.AdminVdc == nil || adminVdc.AdminVdc.ID == "" {
		return fmt.Errorf("cannot lookup VDC Network Profile without VDC ID")
	}

	return deleteVdcNetworkProfile(adminVdc.client, adminVdc.AdminVdc.ID)
}

func getVdcNetworkProfile(client *Client, vdcId string) (*types.VdcNetworkProfile, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcNetworkProfile,
		endpointParams: []string{vdcId},
		entityLabel:    labelVdcNetworkProfile,
	}
	return getInnerEntity[types.VdcNetworkProfile](client, c)
}

func updateVdcNetworkProfile(client *Client, vdcId string, vdcNetworkProfileConfig *types.VdcNetworkProfile) (*types.VdcNetworkProfile, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcNetworkProfile,
		endpointParams: []string{vdcId},
		entityLabel:    labelVdcNetworkProfile,
	}
	return updateInnerEntity(client, c, vdcNetworkProfileConfig)
}

func deleteVdcNetworkProfile(client *Client, vdcId string) error {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcNetworkProfile,
		endpointParams: []string{vdcId},
		entityLabel:    labelVdcNetworkProfile,
	}
	return deleteEntityById(client, c)
}
