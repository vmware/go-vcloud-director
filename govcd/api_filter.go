/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

const (
	labelApiFilter = "API Filter"
)

// ApiFilter is a type for handling API Filters.
type ApiFilter struct {
	ApiFilter *types.ApiFilter
	client    *Client
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (d ApiFilter) wrap(inner *types.ApiFilter) *ApiFilter {
	d.ApiFilter = inner
	return &d
}

// CreateApiFilter creates an API Filter.
func (vcdClient *VCDClient) CreateApiFilter(apiFilter *types.ApiFilter) (*ApiFilter, error) {
	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointApiFilters,
		entityLabel: labelApiFilter,
	}
	outerType := ApiFilter{client: &vcdClient.Client}
	return createOuterEntity(&vcdClient.Client, outerType, c, apiFilter)
}

// GetAllApiFilters retrieves all available API Filters. Query parameters can be supplied to perform additional filtering.
func (vcdClient *VCDClient) GetAllApiFilters(queryParameters url.Values) ([]*ApiFilter, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointApiFilters,
		entityLabel:     labelApiFilter,
		queryParameters: queryParameters,
	}

	outerType := ApiFilter{client: &vcdClient.Client}
	return getAllOuterEntities[ApiFilter, types.ApiFilter](&vcdClient.Client, outerType, c)
}

// GetApiFilterById gets an API Filter by its ID.
func (vcdClient *VCDClient) GetApiFilterById(id string) (*ApiFilter, error) {
	c := crudConfig{
		entityLabel:    labelApiFilter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointApiFilters,
		endpointParams: []string{id},
	}

	outerType := ApiFilter{client: &vcdClient.Client}
	return getOuterEntity[ApiFilter, types.ApiFilter](&vcdClient.Client, outerType, c)
}

// Update updates the receiver API Filter with the values given by the input.
func (ApiFilter *ApiFilter) Update(ep types.ApiFilter) error {
	if ApiFilter.ApiFilter.ID == "" {
		return fmt.Errorf("ID of the receiver API Filter is empty")
	}

	if ep.ID != "" && ep.ID != ApiFilter.ApiFilter.ID {
		return fmt.Errorf("ID of the receiver API Filter and the input ID don't match")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointApiFilters,
		endpointParams: []string{ApiFilter.ApiFilter.ID},
		entityLabel:    labelApiFilter,
	}

	result, err := updateInnerEntity(ApiFilter.client, c, &ep)
	if err != nil {
		return err
	}
	// Only if there was no error in request we overwrite pointer receiver as otherwise it would
	// wipe out existing data
	ApiFilter.ApiFilter = result

	return nil
}

// Delete deletes the receiver API Filter.
func (ApiFilter *ApiFilter) Delete() error {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointApiFilters,
		endpointParams: []string{ApiFilter.ApiFilter.ID},
		entityLabel:    labelApiFilter,
	}

	if err := deleteEntityById(ApiFilter.client, c); err != nil {
		return err
	}

	ApiFilter.ApiFilter = &types.ApiFilter{}
	return nil
}
