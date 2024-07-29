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
	labelExternalEndpoint = "External Endpoint"
)

// ExternalEndpoint is a type for handling External Endpoints.
type ExternalEndpoint struct {
	ExternalEndpoint *types.ExternalEndpoint
	client           *Client
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (d ExternalEndpoint) wrap(inner *types.ExternalEndpoint) *ExternalEndpoint {
	d.ExternalEndpoint = inner
	return &d
}

// CreateExternalEndpoint creates an External Endpoint.
func (vcdClient *VCDClient) CreateExternalEndpoint(externalEndpoint *types.ExternalEndpoint) (*ExternalEndpoint, error) {
	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalEndpoints,
		entityLabel: labelExternalEndpoint,
	}
	outerType := ExternalEndpoint{client: &vcdClient.Client}
	return createOuterEntity(&vcdClient.Client, outerType, c, externalEndpoint)
}

// GetAllExternalEndpoints retrieves all available External Endpoints. Query parameters can be supplied to perform additional filtering.
func (vcdClient *VCDClient) GetAllExternalEndpoints(queryParameters url.Values) ([]*ExternalEndpoint, error) {
	return getAllExternalEndpoints(&vcdClient.Client, queryParameters)
}

// getAllExternalEndpoints retrieves all available External Endpoints. Query parameters can be supplied to perform additional filtering.
func getAllExternalEndpoints(client *Client, queryParameters url.Values) ([]*ExternalEndpoint, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalEndpoints,
		entityLabel:     labelExternalEndpoint,
		queryParameters: queryParameters,
	}

	outerType := ExternalEndpoint{client: client}
	return getAllOuterEntities[ExternalEndpoint, types.ExternalEndpoint](client, outerType, c)
}

// GetExternalEndpoint gets an External Endpoint by its unique combination of vendor, name and version.
func (vcdClient *VCDClient) GetExternalEndpoint(vendor, name, version string) (*ExternalEndpoint, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("vendor==%s;name==%s;version==%s", vendor, name, version))
	externalEndpoints, err := getAllExternalEndpoints(&vcdClient.Client, queryParameters)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("'vendor:name:version'", fmt.Sprintf("'%s:%s:%s'", vendor, name, version), externalEndpoints)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetExternalEndpointById gets an External Endpoint by its ID.
func (vcdClient *VCDClient) GetExternalEndpointById(id string) (*ExternalEndpoint, error) {
	c := crudConfig{
		entityLabel:    labelExternalEndpoint,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalEndpoints,
		endpointParams: []string{id},
	}

	outerType := ExternalEndpoint{client: &vcdClient.Client}
	return getOuterEntity[ExternalEndpoint, types.ExternalEndpoint](&vcdClient.Client, outerType, c)
}

// Update updates the receiver External Endpoint with the values given by the input.
// Note: Vendor, name and version can't be changed. Modifying them will have no effect.
func (externalEndpoint *ExternalEndpoint) Update(ep types.ExternalEndpoint) error {
	if externalEndpoint.ExternalEndpoint.ID == "" {
		return fmt.Errorf("ID of the receiver External Endpoint is empty")
	}

	if ep.ID != "" && ep.ID != externalEndpoint.ExternalEndpoint.ID {
		return fmt.Errorf("ID of the receiver External Endpoint and the input ID don't match")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalEndpoints,
		endpointParams: []string{externalEndpoint.ExternalEndpoint.ID},
		entityLabel:    labelExternalEndpoint,
	}

	result, err := updateInnerEntity(externalEndpoint.client, c, &ep)
	if err != nil {
		return err
	}
	// Only if there was no error in request we overwrite pointer receiver as otherwise it would
	// wipe out existing data
	externalEndpoint.ExternalEndpoint = result

	return nil
}

// Delete deletes the receiver External Endpoint.
// Note: To delete an External Endpoint, it must be disabled, otherwise the operation will fail.
func (externalEndpoint *ExternalEndpoint) Delete() error {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalEndpoints,
		endpointParams: []string{externalEndpoint.ExternalEndpoint.ID},
		entityLabel:    labelExternalEndpoint,
	}

	if err := deleteEntityById(externalEndpoint.client, c); err != nil {
		return err
	}

	externalEndpoint.ExternalEndpoint = &types.ExternalEndpoint{}
	return nil
}
