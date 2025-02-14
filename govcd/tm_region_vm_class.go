package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelRegionVirtualMachineClass = "Region VM Class"

// RegionVirtualMachineClass defines a Region VM Class in VCFA
type RegionVirtualMachineClass struct {
	RegionVirtualMachineClass *types.RegionVirtualMachineClass
	vcdClient                 *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g RegionVirtualMachineClass) wrap(inner *types.RegionVirtualMachineClass) *RegionVirtualMachineClass {
	g.RegionVirtualMachineClass = inner
	return &g
}

// GetAllRegionVirtualMachineClasses retrieves all Region VM Classes
func (vcdClient *VCDClient) GetAllRegionVirtualMachineClasses(queryParameters url.Values) ([]*RegionVirtualMachineClass, error) {
	c := crudConfig{
		entityLabel:     labelRegionVirtualMachineClass,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVmClasses,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := RegionVirtualMachineClass{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetRegionVirtualMachineClassByNameAndRegionId retrieves a Region VM Class by a given name and Region ID
func (vcdClient *VCDClient) GetRegionVirtualMachineClassByNameAndRegionId(name, regionId string) (*RegionVirtualMachineClass, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelRegionQuota)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("name==%s;region.id==%s", name, regionId))
	filteredEntities, err := vcdClient.GetAllRegionVirtualMachineClasses(queryParams)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("name + region.id", fmt.Sprintf("%s + %s", name, regionId), filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetRegionVirtualMachineClassById retrieves a Region VM Class by a given ID
func (vcdClient *VCDClient) GetRegionVirtualMachineClassById(id string) (*RegionVirtualMachineClass, error) {
	c := crudConfig{
		entityLabel:    labelRegionVirtualMachineClass,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVmClasses,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := RegionVirtualMachineClass{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}
