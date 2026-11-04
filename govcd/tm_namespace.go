// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TmNamespace manages a Tenant Manager Namespace.
type TmNamespace struct {
	TmNamespace *types.TmNamespace
	vcdClient   *VCDClient
	options     []TmOperationOptions
}

// GetAllTmNamespaceSummaries retrieves all Namespace summaries. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetAllTmNamespaceSummaries(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmNamespaceSummary, error) {
	c := crudConfig{
		entityLabel:     "namespaceSummaries",
		endpoint:        types.OpenApiPathVcf + "namespaceSummaries",
		requiresTm:      true,
		minimumVersion:  "40.0",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmNamespaceSummary](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateTmNamespace creates a new Namespace. Waits for task completion. Requires TM API 40.0+.
func (vcdClient *VCDClient) CreateTmNamespace(config *types.TmNamespace, options ...TmOperationOptions) (*TmNamespace, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmNamespace config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "namespaces",
		endpoint:       types.OpenApiPathVcf + "namespaces",
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	result, err := createInnerEntity(&vcdClient.Client, c.withTmOptions(options), config)
	if err != nil {
		return nil, err
	}
	return &TmNamespace{TmNamespace: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// CreateTmNamespaceAsync creates a new Namespace. Returns the tracking task without waiting. Requires TM
// API 40.0+.
func (vcdClient *VCDClient) CreateTmNamespaceAsync(config *types.TmNamespace, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmNamespace config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "namespaces",
		endpoint:       types.OpenApiPathVcf + "namespaces",
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// GetTmNamespaceById retrieves a Namespace. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetTmNamespaceById(namespaceId string, options ...TmOperationOptions) (*TmNamespace, error) {
	c := crudConfig{
		entityLabel:    "namespaces",
		endpoint:       types.OpenApiPathVcf + "namespaces/%s",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{namespaceId},
	}
	result, err := getInnerEntity[types.TmNamespace](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmNamespace{TmNamespace: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// Update updates a Namespace. Waits for task completion. Requires TM API 40.0+.
func (o *TmNamespace) Update(config *types.TmNamespace, options ...TmOperationOptions) (*TmNamespace, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "namespaces",
		endpoint:       types.OpenApiPathVcf + "namespaces/%s",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{o.TmNamespace.ID},
	}
	if err := runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config); err != nil {
		return nil, err
	}
	result, err := getInnerEntity[types.TmNamespace](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmNamespace{TmNamespace: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// UpdateAsync updates a Namespace. Returns the tracking task without waiting. Requires TM API 40.0+.
func (o *TmNamespace) UpdateAsync(config *types.TmNamespace, options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "namespaces",
		endpoint:       types.OpenApiPathVcf + "namespaces/%s",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{o.TmNamespace.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config)
}

// Delete deletes a specified namespace. Waits for task completion. Requires TM API 40.0+. With detachOnly,
// the namespace remains in vCenter.
func (o *TmNamespace) Delete(detachOnly bool, options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "namespaces",
		endpoint:        types.OpenApiPathVcf + "namespaces/%s",
		requiresTm:      true,
		minimumVersion:  "40.0",
		endpointParams:  []string{o.TmNamespace.ID},
		queryParameters: url.Values{"detachOnly": {strconv.FormatBool(detachOnly)}},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// DeleteAsync deletes a specified namespace. Returns the tracking task without waiting. Requires TM API
// 40.0+. With detachOnly, the namespace remains in vCenter.
func (o *TmNamespace) DeleteAsync(detachOnly bool, options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "namespaces",
		endpoint:        types.OpenApiPathVcf + "namespaces/%s",
		requiresTm:      true,
		minimumVersion:  "40.0",
		endpointParams:  []string{o.TmNamespace.ID},
		queryParameters: url.Values{"detachOnly": {strconv.FormatBool(detachOnly)}},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// GetAllCapabilities retrieves capabilities of a specific namespace. Requires TM API 40.1+.
func (o *TmNamespace) GetAllCapabilities(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmRegionCapability, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "namespaces/capabilities",
		endpoint:        types.OpenApiPathVcf + "namespaces/%s/capabilities",
		requiresTm:      true,
		minimumVersion:  "40.1",
		endpointParams:  []string{o.TmNamespace.ID},
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmRegionCapability](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ImportTmNamespace imports an existing Namespace from vCenter. Waits for task completion. Requires TM API
// 40.0+.
func (vcdClient *VCDClient) ImportTmNamespace(config *types.TmNamespaceImportDetails, options ...TmOperationOptions) error {
	if config == nil {
		return fmt.Errorf("ImportTmNamespace config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "namespaces/import",
		endpoint:       types.OpenApiPathVcf + "namespaces/import",
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// ImportTmNamespaceAsync imports an existing Namespace from vCenter. Returns the tracking task without
// waiting. Requires TM API 40.0+.
func (vcdClient *VCDClient) ImportTmNamespaceAsync(config *types.TmNamespaceImportDetails, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("ImportTmNamespace config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "namespaces/import",
		endpoint:       types.OpenApiPathVcf + "namespaces/import",
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// GetAllTmNamespaceStorageClassMetrics retrieves paginated list of namespace storage class usage objects.
// Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmNamespaceStorageClassMetrics(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmNamespaceStorageClassMetric, error) {
	c := crudConfig{
		entityLabel:     "namespaceStorageClassMetrics",
		endpoint:        types.OpenApiPathVcf + "namespaceStorageClassMetrics",
		requiresTm:      true,
		minimumVersion:  "40.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmNamespaceStorageClassMetric](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return result, nil
}
