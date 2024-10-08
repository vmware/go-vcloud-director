/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelvCenter = "vCenter"

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

func (vcdClient *VCDClient) CreateVcenter(config *types.VSphereVirtualCenter) (*VCenter, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("requires TM")
	}

	c := crudConfig{
		entityLabel: labelvCenter,
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
	}

	outerType := VCenter{client: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

func (vcdClient *VCDClient) GetAllVCenters(queryParameters url.Values) ([]*VCenter, error) {
	c := crudConfig{
		entityLabel:     labelvCenter,
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		queryParameters: queryParameters,
	}

	outerType := VCenter{client: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetVCenterByName(name string) (*VCenter, error) {
	vcenters, err := vcdClient.GetAllVCenters(nil)
	if err != nil {
		return nil, err
	}
	for _, vc := range vcenters {
		if vc.VSphereVCenter.Name == name {
			return vc, nil
		}
	}
	return nil, fmt.Errorf("vcenter %s not found: %s", name, ErrorEntityNotFound)
}

func (vcdClient *VCDClient) GetVCenterById(id string) (*VCenter, error) {

	c := crudConfig{
		entityLabel:    labelvCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{id},
	}

	outerType := VCenter{client: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (o *VCenter) Update(TmOrgConfig *types.VSphereVirtualCenter) (*VCenter, error) {
	c := crudConfig{
		entityLabel:    labelvCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{o.VSphereVCenter.VcId},
	}

	outerType := VCenter{client: o.client}
	return updateOuterEntity(&o.client.Client, outerType, c, TmOrgConfig)
}

func (o *VCenter) Delete() error {
	c := crudConfig{
		entityLabel:    labelvCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{o.VSphereVCenter.VcId},
	}
	return deleteEntityById(&o.client.Client, c)
}

func (vcenter VCenter) GetVimServerUrl() (string, error) {
	return url.JoinPath(vcenter.client.Client.VCDHREF.String(), "admin", "extension", "vimServer", extractUuid(vcenter.VSphereVCenter.VcId))
}
