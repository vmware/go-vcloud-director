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

const labelTmAviController = "TM Avi Controller"

// TmAviController manages an Avi Load Balancer Controller.
type TmAviController struct {
	TmAviController *types.TmAviController
	vcdClient       *VCDClient
	options         []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmAviController) wrap(inner *types.TmAviController) *TmAviController {
	g.TmAviController = inner
	return &g
}

// tmAviControllerConfig returns the CRUD configuration for Avi Load Balancer Controller operations. Optional params are appended to the endpoint.
func tmAviControllerConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmAviController,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmAviControllers,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "40.1",
	}.withTmOptions(options)
}

// CreateTmAviController creates an Avi Load Balancer Controller. Waits for task completion. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmAviController(config *types.TmAviController, options ...TmOperationOptions) (*TmAviController, error) {
	outerType := TmAviController{vcdClient: vcdClient, options: retainTmOptions(options)}
	return createOuterEntity(&vcdClient.Client, outerType, tmAviControllerConfig(options), config)
}

// CreateTmAviControllerAsync creates an Avi Load Balancer Controller. Returns the tracking task without waiting. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmAviControllerAsync(config *types.TmAviController, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmAviController config must not be nil")
	}
	return createInnerEntityAsync(&vcdClient.Client, tmAviControllerConfig(options), config)
}

// GetAllTmAviControllers retrieves all Avi Load Balancer Controllers. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmAviControllers(queryParameters url.Values, options ...TmOperationOptions) ([]*TmAviController, error) {
	c := tmAviControllerConfig(options)
	c.queryParameters = queryParameters
	outerType := TmAviController{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmAviControllerByName retrieves an Avi Load Balancer Controller by name. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviControllerByName(name string, options ...TmOperationOptions) (*TmAviController, error) {
	entities, err := vcdClient.GetAllTmAviControllers(nil, options...)
	if err != nil {
		return nil, err
	}
	return tmEntityByName(labelTmAviController, name, entities, func(e *TmAviController) string { return e.TmAviController.Name })
}

// GetTmAviControllerById retrieves an Avi Load Balancer Controller by ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviControllerById(id string, options ...TmOperationOptions) (*TmAviController, error) {
	outerType := TmAviController{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmAviControllerConfig(options, id))
}

// GetTmAviControllerByNameOrId retrieves an Avi Load Balancer Controller by name or ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviControllerByNameOrId(identifier string, options ...TmOperationOptions) (*TmAviController, error) {
	return tmEntityByNameOrId(identifier,
		func(id string) (*TmAviController, error) { return vcdClient.GetTmAviControllerById(id, options...) },
		func(name string) (*TmAviController, error) {
			return vcdClient.GetTmAviControllerByName(name, options...)
		})
}

// Update updates an Avi Load Balancer Controller. Waits for task completion. Requires TM API 40.1+.
func (o *TmAviController) Update(config *types.TmAviController, options ...TmOperationOptions) (*TmAviController, error) {
	options = inheritTmOptions(o.options, options)
	outerType := TmAviController{vcdClient: o.vcdClient, options: retainTmOptions(options)}
	return updateOuterEntity(&o.vcdClient.Client, outerType, tmAviControllerConfig(options, o.TmAviController.ID), config)
}

// UpdateAsync updates an Avi Load Balancer Controller. Returns the tracking task without waiting. Requires TM API 40.1+.
func (o *TmAviController) UpdateAsync(config *types.TmAviController, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	options = inheritTmOptions(o.options, options)
	return runOpenApiOperationAsync(&o.vcdClient.Client, tmAviControllerConfig(options, o.TmAviController.ID), http.MethodPut, config)
}

// Delete unregisters an Avi Load Balancer Controller. Waits for task completion. Requires TM API 40.1+.
func (o *TmAviController) Delete(force bool, options ...TmOperationOptions) error {
	options = inheritTmOptions(o.options, options)
	c := tmAviControllerConfig(options, o.TmAviController.ID)
	c.queryParameters = url.Values{"force": {strconv.FormatBool(force)}}
	return deleteEntityById(&o.vcdClient.Client, c)
}

// DeleteAsync unregisters an Avi Load Balancer Controller. Returns the tracking task without waiting. Requires TM API 40.1+.
func (o *TmAviController) DeleteAsync(force bool, options ...TmOperationOptions) (*Task, error) {
	options = inheritTmOptions(o.options, options)
	c := tmAviControllerConfig(options, o.TmAviController.ID)
	c.queryParameters = url.Values{"force": {strconv.FormatBool(force)}}
	return runOpenApiOperationAsync(&o.vcdClient.Client, c, http.MethodDelete, nil)
}

// Sync synchronizes the Avi Load Balancer Controller along with its associated Cloud and Service Engine Groups. Waits for task completion. Requires TM API 40.1+.
func (o *TmAviController) Sync(options ...TmOperationOptions) error {
	options = inheritTmOptions(o.options, options)
	return runOpenApiOperation(&o.vcdClient.Client, tmCrudConfig("TM Avi Controller sync", types.OpenApiEndpointTmAviControllerSync, "40.1", options, o.TmAviController.ID), http.MethodPost, nil)
}

// SyncAsync synchronizes the Avi Load Balancer Controller along with its associated Cloud and Service Engine Groups. Returns the tracking task without waiting. Requires TM API 40.1+.
func (o *TmAviController) SyncAsync(options ...TmOperationOptions) (*Task, error) {
	options = inheritTmOptions(o.options, options)
	return runOpenApiOperationAsync(&o.vcdClient.Client, tmCrudConfig("TM Avi Controller sync", types.OpenApiEndpointTmAviControllerSync, "40.1", options, o.TmAviController.ID), http.MethodPost, nil)
}
