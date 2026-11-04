// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/http"
	"net/url"
	"strconv"
)

const labelTmNamespace = "TM Namespace"

// TmNamespace manages a Namespace.
type TmNamespace struct {
	TmNamespace *types.TmNamespace
	vcdClient   *VCDClient
	options     []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmNamespace) wrap(inner *types.TmNamespace) *TmNamespace {
	g.TmNamespace = inner
	return &g
}

// tmNamespaceConfig returns the CRUD configuration for Namespace operations. Optional params are appended to the endpoint.
func tmNamespaceConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmNamespace,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmNamespaces,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "40.0",
	}.withTmOptions(options)
}

// CreateTmNamespace creates a Namespace. Waits for task completion. Requires TM API 40.0+.
func (vcdClient *VCDClient) CreateTmNamespace(config *types.TmNamespace, options ...TmOperationOptions) (*TmNamespace, error) {
	outerType := TmNamespace{vcdClient: vcdClient, options: retainTmOptions(options)}
	return createOuterEntity(&vcdClient.Client, outerType, tmNamespaceConfig(options), config)
}

// CreateTmNamespaceAsync creates a Namespace. Returns the tracking task without waiting. Requires TM API 40.0+.
func (vcdClient *VCDClient) CreateTmNamespaceAsync(config *types.TmNamespace, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmNamespace config must not be nil")
	}
	return createInnerEntityAsync(&vcdClient.Client, tmNamespaceConfig(options), config)
}

// GetAllTmNamespaces retrieves all Namespaces. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetAllTmNamespaces(queryParameters url.Values, options ...TmOperationOptions) ([]*TmNamespace, error) {
	c := tmNamespaceConfig(options)
	c.queryParameters = queryParameters
	outerType := TmNamespace{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmNamespaceByName retrieves a Namespace by name. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetTmNamespaceByName(name string, options ...TmOperationOptions) (*TmNamespace, error) {
	entities, err := vcdClient.GetAllTmNamespaces(nil, options...)
	if err != nil {
		return nil, err
	}
	return tmEntityByName(labelTmNamespace, name, entities, func(e *TmNamespace) string { return e.TmNamespace.Name })
}

// GetTmNamespaceById retrieves a Namespace by ID. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetTmNamespaceById(id string, options ...TmOperationOptions) (*TmNamespace, error) {
	outerType := TmNamespace{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmNamespaceConfig(options, id))
}

// GetTmNamespaceByNameOrId retrieves a Namespace by name or ID. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetTmNamespaceByNameOrId(identifier string, options ...TmOperationOptions) (*TmNamespace, error) {
	return tmEntityByNameOrId(identifier,
		func(id string) (*TmNamespace, error) { return vcdClient.GetTmNamespaceById(id, options...) },
		func(name string) (*TmNamespace, error) { return vcdClient.GetTmNamespaceByName(name, options...) })
}

// Update updates a Namespace. Waits for task completion. Requires TM API 40.0+.
func (o *TmNamespace) Update(config *types.TmNamespace, options ...TmOperationOptions) (*TmNamespace, error) {
	options = inheritTmOptions(o.options, options)
	outerType := TmNamespace{vcdClient: o.vcdClient, options: retainTmOptions(options)}
	return updateOuterEntity(&o.vcdClient.Client, outerType, tmNamespaceConfig(options, o.TmNamespace.ID), config)
}

// UpdateAsync updates a Namespace. Returns the tracking task without waiting. Requires TM API 40.0+.
func (o *TmNamespace) UpdateAsync(config *types.TmNamespace, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	options = inheritTmOptions(o.options, options)
	return runOpenApiOperationAsync(&o.vcdClient.Client, tmNamespaceConfig(options, o.TmNamespace.ID), http.MethodPut, config)
}

// Delete deletes a Namespace. Waits for task completion. Requires TM API 40.0+. With detachOnly, the namespace remains in vCenter.
func (o *TmNamespace) Delete(detachOnly bool, options ...TmOperationOptions) error {
	options = inheritTmOptions(o.options, options)
	c := tmNamespaceConfig(options, o.TmNamespace.ID)
	c.queryParameters = url.Values{"detachOnly": {strconv.FormatBool(detachOnly)}}
	return deleteEntityById(&o.vcdClient.Client, c)
}

// DeleteAsync deletes a Namespace. Returns the tracking task without waiting. Requires TM API 40.0+. With detachOnly, the namespace remains in vCenter.
func (o *TmNamespace) DeleteAsync(detachOnly bool, options ...TmOperationOptions) (*Task, error) {
	options = inheritTmOptions(o.options, options)
	c := tmNamespaceConfig(options, o.TmNamespace.ID)
	c.queryParameters = url.Values{"detachOnly": {strconv.FormatBool(detachOnly)}}
	return runOpenApiOperationAsync(&o.vcdClient.Client, c, http.MethodDelete, nil)
}

// GetAllTmNamespaceSummaries retrieves all Namespace summaries. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetAllTmNamespaceSummaries(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmNamespaceSummary, error) {
	c := tmCrudConfig("TM Namespace Summary", types.OpenApiEndpointTmNamespaceSummaries, "40.0", options)
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmNamespaceSummary](&vcdClient.Client, c)
}

// GetAllCapabilities retrieves the capabilities of a Namespace. Requires TM API 40.1+.
func (o *TmNamespace) GetAllCapabilities(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmRegionCapability, error) {
	options = inheritTmOptions(o.options, options)
	c := tmCrudConfig("TM Namespace capability", types.OpenApiEndpointTmNamespaceCapabilities, "40.1", options, o.TmNamespace.ID)
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmRegionCapability](&o.vcdClient.Client, c)
}

// ImportTmNamespace imports an existing Namespace from vCenter. Waits for task completion. Requires TM API 40.0+.
func (vcdClient *VCDClient) ImportTmNamespace(config *types.TmNamespaceImportDetails, options ...TmOperationOptions) error {
	if config == nil {
		return fmt.Errorf("ImportTmNamespace config must not be nil")
	}
	c := tmCrudConfig("TM Namespace import", types.OpenApiEndpointTmNamespaceImport, "40.0", options)
	return runOpenApiOperation(&vcdClient.Client, c, http.MethodPost, config)
}

// ImportTmNamespaceAsync imports an existing Namespace from vCenter. Returns the tracking task without waiting.
// Requires TM API 40.0+.
func (vcdClient *VCDClient) ImportTmNamespaceAsync(config *types.TmNamespaceImportDetails, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("ImportTmNamespace config must not be nil")
	}
	c := tmCrudConfig("TM Namespace import", types.OpenApiEndpointTmNamespaceImport, "40.0", options)
	return runOpenApiOperationAsync(&vcdClient.Client, c, http.MethodPost, config)
}

// GetAllTmNamespaceStorageClassMetrics retrieves the paginated list of Namespace storage class usage objects.
// Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmNamespaceStorageClassMetrics(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmNamespaceStorageClassMetric, error) {
	c := tmCrudConfig("TM Namespace storage class metric", types.OpenApiEndpointTmNamespaceStorageClassMetrics, "40.1", options)
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmNamespaceStorageClassMetric](&vcdClient.Client, c)
}
