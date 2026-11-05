// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/http"
	"net/url"
)

const labelTmProjectAssignment = "TM Project Assignment"

// TmProjectAssignment manages a Project Assignment.
type TmProjectAssignment struct {
	TmProjectAssignment *types.TmProjectAssignment
	vcdClient           *VCDClient
	options             []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmProjectAssignment) wrap(inner *types.TmProjectAssignment) *TmProjectAssignment {
	g.TmProjectAssignment = inner
	return &g
}

// tmProjectAssignmentConfig returns the CRUD configuration for Project Assignment operations. Optional params are appended to the endpoint.
func tmProjectAssignmentConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmProjectAssignment,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmProjectAssignments,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "40.0",
	}.withTmOptions(options)
}

// GetAllTmProjectAssignments retrieves all Project Assignments. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetAllTmProjectAssignments(queryParameters url.Values, options ...TmOperationOptions) ([]*TmProjectAssignment, error) {
	c := tmProjectAssignmentConfig(options)
	c.queryParameters = queryParameters
	outerType := TmProjectAssignment{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmProjectAssignmentByName retrieves a Project Assignment by name. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetTmProjectAssignmentByName(name string, options ...TmOperationOptions) (*TmProjectAssignment, error) {
	entities, err := vcdClient.GetAllTmProjectAssignments(nil, options...)
	if err != nil {
		return nil, err
	}
	return tmEntityByName(labelTmProjectAssignment, name, entities, func(e *TmProjectAssignment) string { return e.TmProjectAssignment.Name })
}

// GetTmProjectAssignmentById retrieves a Project Assignment by ID. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetTmProjectAssignmentById(id string, options ...TmOperationOptions) (*TmProjectAssignment, error) {
	outerType := TmProjectAssignment{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmProjectAssignmentConfig(options, id))
}

// GetTmProjectAssignmentByNameOrId retrieves a Project Assignment by name or ID. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetTmProjectAssignmentByNameOrId(identifier string, options ...TmOperationOptions) (*TmProjectAssignment, error) {
	return tmEntityByNameOrId(identifier,
		func(id string) (*TmProjectAssignment, error) {
			return vcdClient.GetTmProjectAssignmentById(id, options...)
		},
		func(name string) (*TmProjectAssignment, error) {
			return vcdClient.GetTmProjectAssignmentByName(name, options...)
		})
}

// Update updates a Project Assignment. Waits for task completion. Requires TM API 40.0+.
func (o *TmProjectAssignment) Update(config *types.TmProjectAssignment, options ...TmOperationOptions) (*TmProjectAssignment, error) {
	options = inheritTmOptions(o.options, options)
	outerType := TmProjectAssignment{vcdClient: o.vcdClient, options: retainTmOptions(options)}
	return updateOuterEntity(&o.vcdClient.Client, outerType, tmProjectAssignmentConfig(options, o.TmProjectAssignment.ID), config)
}

// UpdateAsync updates a Project Assignment. Returns the tracking task without waiting. Requires TM API 40.0+.
func (o *TmProjectAssignment) UpdateAsync(config *types.TmProjectAssignment, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	options = inheritTmOptions(o.options, options)
	return runOpenApiOperationAsync(&o.vcdClient.Client, tmProjectAssignmentConfig(options, o.TmProjectAssignment.ID), http.MethodPut, config)
}

// Delete deletes a Project Assignment. Waits for task completion. Requires TM API 40.0+.
func (o *TmProjectAssignment) Delete(options ...TmOperationOptions) error {
	options = inheritTmOptions(o.options, options)
	c := tmProjectAssignmentConfig(options, o.TmProjectAssignment.ID)
	return deleteEntityById(&o.vcdClient.Client, c)
}

// DeleteAsync deletes a Project Assignment. Returns the tracking task without waiting. Requires TM API 40.0+.
func (o *TmProjectAssignment) DeleteAsync(options ...TmOperationOptions) (*Task, error) {
	options = inheritTmOptions(o.options, options)
	c := tmProjectAssignmentConfig(options, o.TmProjectAssignment.ID)
	return runOpenApiOperationAsync(&o.vcdClient.Client, c, http.MethodDelete, nil)
}

// CreateTmProjectAssignment creates a Project Assignment. The project ID is supplied by the caller because the task
// owner does not identify the project. Waits for task completion. Requires TM API 40.0+.
func (vcdClient *VCDClient) CreateTmProjectAssignment(config *types.TmProjectAssignment, options ...TmOperationOptions) (*TmProjectAssignment, error) {
	if config == nil || config.ID == "" {
		return nil, fmt.Errorf("CreateTmProjectAssignment requires a config with a project ID")
	}
	if err := runOpenApiOperation(&vcdClient.Client, tmProjectAssignmentConfig(options), http.MethodPost, config); err != nil {
		return nil, err
	}
	return vcdClient.GetTmProjectAssignmentById(config.ID, options...)
}

// CreateTmProjectAssignmentAsync creates a Project Assignment. Returns the tracking task without waiting.
// Requires TM API 40.0+.
func (vcdClient *VCDClient) CreateTmProjectAssignmentAsync(config *types.TmProjectAssignment, options ...TmOperationOptions) (*Task, error) {
	if config == nil || config.ID == "" {
		return nil, fmt.Errorf("CreateTmProjectAssignment requires a config with a project ID")
	}
	return runOpenApiOperationAsync(&vcdClient.Client, tmProjectAssignmentConfig(options), http.MethodPost, config)
}

// GetAllTmProjectAccessLevels retrieves all Project Access Levels. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetAllTmProjectAccessLevels(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmProjectAccessLevel, error) {
	c := tmCrudConfig("TM Project Access Level", types.OpenApiEndpointTmProjectAccessLevels, "40.0", options)
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmProjectAccessLevel](&vcdClient.Client, c)
}

// GetAllMembers retrieves all Project Assignment Members. Requires TM API 40.0+.
func (o *TmProjectAssignment) GetAllMembers(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmProjectAssignmentMember, error) {
	options = inheritTmOptions(o.options, options)
	c := tmCrudConfig("TM Project Assignment member", types.OpenApiEndpointTmProjectAssignmentMembers, "40.0", options, o.TmProjectAssignment.ID)
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmProjectAssignmentMember](&o.vcdClient.Client, c)
}

// UpdateMembers updates all Project Assignment Members. Waits for task completion. Requires TM API 40.0+.
func (o *TmProjectAssignment) UpdateMembers(config *types.TmProjectAssignmentMembers, options ...TmOperationOptions) error {
	if config == nil {
		return fmt.Errorf("UpdateMembers config must not be nil")
	}
	options = inheritTmOptions(o.options, options)
	c := tmCrudConfig("TM Project Assignment member", types.OpenApiEndpointTmProjectAssignmentMembers, "40.0", options, o.TmProjectAssignment.ID)
	return runOpenApiOperation(&o.vcdClient.Client, c, http.MethodPut, config)
}

// UpdateMembersAsync updates all Project Assignment Members. Returns the tracking task without waiting.
// Requires TM API 40.0+.
func (o *TmProjectAssignment) UpdateMembersAsync(config *types.TmProjectAssignmentMembers, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("UpdateMembers config must not be nil")
	}
	options = inheritTmOptions(o.options, options)
	c := tmCrudConfig("TM Project Assignment member", types.OpenApiEndpointTmProjectAssignmentMembers, "40.0", options, o.TmProjectAssignment.ID)
	return runOpenApiOperationAsync(&o.vcdClient.Client, c, http.MethodPut, config)
}
