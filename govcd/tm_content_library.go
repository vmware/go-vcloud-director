package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelContentLibrary = "Content Library"

type ContentLibrary struct {
	ContentLibrary *types.ContentLibrary
	vcdClient      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g ContentLibrary) wrap(inner *types.ContentLibrary) *ContentLibrary {
	g.ContentLibrary = inner
	return &g
}

// TODO: This one probably needs TenantContext, as can be created as Tenants
func (vcdClient *VCDClient) CreateContentLibrary(config *types.ContentLibrary) (*ContentLibrary, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("err")
	}
	c := crudConfig{
		entityLabel: labelContentLibrary,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
	}
	outerType := ContentLibrary{vcdClient: vcdClient}
	// FIXME: Workaround, this should be eventually refactored to match other OpenAPI endpoints.
	//        - Problem: The returned Task references a Catalog instead of a ContentLibrary, hence retrieving the resulting object
	//                from finished Task fails.
	//        - Solution: Retry fetching the entity again on error with the information inside of it
	result, err := createOuterEntity(&vcdClient.Client, outerType, c, config)
	if err != nil {
		if !strings.Contains(err.Error(), "urn:vcloud:catalog:") {
			return nil, err
		}
		// The created Content Library has the same UUID as the Catalog, which is present in the thrown error
		result, err = vcdClient.GetContentLibraryById(fmt.Sprintf("urn:vcloud:contentLibrary:%s", extractUuid(err.Error())))
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (vcdClient *VCDClient) GetAllContentLibraries(queryParameters url.Values) ([]*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:     labelContentLibrary,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		queryParameters: queryParameters,
	}

	outerType := ContentLibrary{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetContentLibraryByName(name string) (*ContentLibrary, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelContentLibrary)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllContentLibraries(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetContentLibraryById(singleEntity.ContentLibrary.Id)
}

func (vcdClient *VCDClient) GetContentLibraryById(id string) (*ContentLibrary, error) {
	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams: []string{id},
	}

	outerType := ContentLibrary{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (o *ContentLibrary) Update(contentLibraryConfig *types.ContentLibrary) (*ContentLibrary, error) {
	return nil, fmt.Errorf("not supported")
}

func (o *ContentLibrary) Delete() error {
	c := crudConfig{
		entityLabel:    labelContentLibrary,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries,
		endpointParams: []string{o.ContentLibrary.Id},
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
