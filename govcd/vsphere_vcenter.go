/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelVirtualCenter = "vCenter Server"

type VCenter struct {
	VSphereVCenter *types.VSphereVirtualCenter
	client         *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (v VCenter) wrap(inner *types.VSphereVirtualCenter) *VCenter {
	v.VSphereVCenter = inner
	return &v
}

// CreateVcenter
func (vcdClient *VCDClient) CreateVcenter(config *types.VSphereVirtualCenter) (*VCenter, error) {
	c := crudConfig{
		entityLabel: labelVirtualCenter,
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
	}
	outerType := VCenter{client: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllVCenters retrieves all vCenter servers based on optional query filtering
func (vcdClient *VCDClient) GetAllVCenters(queryParams url.Values) ([]*VCenter, error) {
	c := crudConfig{
		entityLabel:     labelVirtualCenter,
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		queryParameters: queryParams,
	}

	outerType := VCenter{client: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetVCenterByName retrieves vCenter server by name
func (vcdClient *VCDClient) GetVCenterByName(name string) (*VCenter, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelVirtualCenter)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	vCenters, err := vcdClient.GetAllVCenters(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, vCenters)
	if err != nil {
		return nil, err
	}

	return singleEntity, nil
}

// GetVCenterById retrieves vCenter server by id
func (vcdClient *VCDClient) GetVCenterById(id string) (*VCenter, error) {
	c := crudConfig{
		entityLabel:    labelVirtualCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{id},
	}

	outerType := VCenter{client: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (v *VCenter) Update(TmNsxtManagerConfig *types.VSphereVirtualCenter) (*VCenter, error) {
	c := crudConfig{
		entityLabel:    labelVirtualCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{v.VSphereVCenter.VcId},
	}
	outerType := VCenter{client: v.client}
	return updateOuterEntity(&v.client.Client, outerType, c, TmNsxtManagerConfig)
}

func (v *VCenter) Delete() error {
	c := crudConfig{
		entityLabel:    labelVirtualCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{v.VSphereVCenter.VcId},
	}
	return deleteEntityById(&v.client.Client, c)
}

func (v *VCenter) Disable() error {
	v.VSphereVCenter.IsEnabled = false
	_, err := v.Update(v.VSphereVCenter)
	return err
}

func (vcenter VCenter) GetVimServerUrl() (string, error) {
	return url.JoinPath(vcenter.client.Client.VCDHREF.String(), "admin", "extension", "vimServer", extractUuid(vcenter.VSphereVCenter.VcId))
}
