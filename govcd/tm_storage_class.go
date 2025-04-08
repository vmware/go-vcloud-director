// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelStorageClass = "Storage Class"

// StorageClass defines the Storage Class data structure
type StorageClass struct {
	StorageClass *types.StorageClass
	vcdClient    *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g StorageClass) wrap(inner *types.StorageClass) *StorageClass {
	g.StorageClass = inner
	return &g
}

// GetAllStorageClasses retrieves all Storage Classes with the given query parameters, which allow setting filters
// and other constraints
func (vcdClient *VCDClient) GetAllStorageClasses(queryParameters url.Values) ([]*StorageClass, error) {
	c := crudConfig{
		entityLabel:     labelStorageClass,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointStorageClasses,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := StorageClass{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetStorageClassByName retrieves a Storage Class by name, that belongs to the given Region
func (r *Region) GetStorageClassByName(name string) (*StorageClass, error) {
	return getStorageClassByName(r.vcdClient, name, r.Region.ID)
}

// GetStorageClassByName retrieves a Storage Class by name
func (vcdClient *VCDClient) GetStorageClassByName(name string) (*StorageClass, error) {
	return getStorageClassByName(vcdClient, name, "")
}

func getStorageClassByName(vcdClient *VCDClient, name, regionId string) (*StorageClass, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelStorageClass)
	}
	filter := fmt.Sprintf("(name==%s)", name)
	if regionId != "" {
		filter = fmt.Sprintf("(name==%s;region.id==%s)", name, regionId)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", filter)

	filteredEntities, err := vcdClient.GetAllStorageClasses(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetStorageClassById(singleEntity.StorageClass.ID)
}

// GetStorageClassById retrieves a Storage Class by ID
func (vcdClient *VCDClient) GetStorageClassById(id string) (*StorageClass, error) {
	c := crudConfig{
		entityLabel:    labelStorageClass,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointStorageClasses,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := StorageClass{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}
