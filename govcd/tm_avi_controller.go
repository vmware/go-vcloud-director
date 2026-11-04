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

// TmAviController manages a Tenant Manager AviController.
type TmAviController struct {
	TmAviController *types.TmAviController
	vcdClient       *VCDClient
	options         []TmOperationOptions
}

// CreateTmAviController register a new Avi Load Balancer Controller. Waits for task completion. Requires TM
// API 40.1+.
func (vcdClient *VCDClient) CreateTmAviController(config *types.TmAviController, options ...TmOperationOptions) (*TmAviController, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmAviController config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "loadBalancer/aviControllers",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviControllers",
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	result, err := createInnerEntity(&vcdClient.Client, c.withTmOptions(options), config)
	if err != nil {
		return nil, err
	}
	return &TmAviController{TmAviController: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// CreateTmAviControllerAsync register a new Avi Load Balancer Controller. Returns the tracking task without
// waiting. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmAviControllerAsync(config *types.TmAviController, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmAviController config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "loadBalancer/aviControllers",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviControllers",
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// GetAllTmAviControllers retrieves all registered Avi Load Balancer Controllers in the system. Requires TM
// API 40.1+.
func (vcdClient *VCDClient) GetAllTmAviControllers(queryParameters url.Values, options ...TmOperationOptions) ([]*TmAviController, error) {
	c := crudConfig{
		entityLabel:     "loadBalancer/aviControllers",
		endpoint:        types.OpenApiPathVcf + "loadBalancer/aviControllers",
		requiresTm:      true,
		minimumVersion:  "40.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmAviController](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	wrapped := make([]*TmAviController, len(result))
	for i := range result {
		wrapped[i] = &TmAviController{TmAviController: result[i], vcdClient: vcdClient, options: retainTmOptions(options)}
	}
	return wrapped, nil
}

// GetTmAviControllerById retrieves Avi Load Balancer Controller. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviControllerById(aviControllerId string, options ...TmOperationOptions) (*TmAviController, error) {
	c := crudConfig{
		entityLabel:    "loadBalancer/aviControllers",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviControllers/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{aviControllerId},
	}
	result, err := getInnerEntity[types.TmAviController](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmAviController{TmAviController: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// Update updates specified Avi Load Balancer Controller. Waits for task completion. Requires TM API 40.1+.
func (o *TmAviController) Update(config *types.TmAviController, options ...TmOperationOptions) (*TmAviController, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "loadBalancer/aviControllers",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviControllers/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmAviController.ID},
	}
	if err := runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config); err != nil {
		return nil, err
	}
	result, err := getInnerEntity[types.TmAviController](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmAviController{TmAviController: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// UpdateAsync updates specified Avi Load Balancer Controller. Returns the tracking task without waiting.
// Requires TM API 40.1+.
func (o *TmAviController) UpdateAsync(config *types.TmAviController, options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "loadBalancer/aviControllers",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviControllers/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmAviController.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config)
}

// Delete unregister the specified Avi Load Balancer Controller. Waits for task completion. Requires TM API
// 40.1+.
func (o *TmAviController) Delete(force bool, options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "loadBalancer/aviControllers",
		endpoint:        types.OpenApiPathVcf + "loadBalancer/aviControllers/%s",
		requiresTm:      true,
		minimumVersion:  "40.1",
		endpointParams:  []string{o.TmAviController.ID},
		queryParameters: url.Values{"force": {strconv.FormatBool(force)}},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// DeleteAsync unregister the specified Avi Load Balancer Controller. Returns the tracking task without
// waiting. Requires TM API 40.1+.
func (o *TmAviController) DeleteAsync(force bool, options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "loadBalancer/aviControllers",
		endpoint:        types.OpenApiPathVcf + "loadBalancer/aviControllers/%s",
		requiresTm:      true,
		minimumVersion:  "40.1",
		endpointParams:  []string{o.TmAviController.ID},
		queryParameters: url.Values{"force": {strconv.FormatBool(force)}},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// Sync synchronizes Avi Load Balancer Controller along with associated Cloud and Service Engine Groups.
// Waits for task completion. Requires TM API 40.1+.
func (o *TmAviController) Sync(options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "loadBalancer/aviControllers/sync",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviControllers/%s/sync",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmAviController.ID},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, nil)
}

// SyncAsync synchronizes Avi Load Balancer Controller along with associated Cloud and Service Engine
// Groups. Returns the tracking task without waiting. Requires TM API 40.1+.
func (o *TmAviController) SyncAsync(options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "loadBalancer/aviControllers/sync",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviControllers/%s/sync",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmAviController.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, nil)
}
