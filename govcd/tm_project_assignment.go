// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TmProjectAssignment manages a Tenant Manager ProjectAssignment.
type TmProjectAssignment struct {
	TmProjectAssignment *types.TmProjectAssignment
	vcdClient           *VCDClient
	options             []TmOperationOptions
}

// GetAllTmProjectAccessLevels retrieves all Project Access Levels. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetAllTmProjectAccessLevels(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmProjectAccessLevel, error) {
	c := crudConfig{
		entityLabel:     "projectAccessLevels",
		endpoint:        types.OpenApiPathVcf + "projectAccessLevels",
		requiresTm:      true,
		minimumVersion:  "40.0",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmProjectAccessLevel](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAllTmProjectAssignments retrieves all Project Assignments. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetAllTmProjectAssignments(queryParameters url.Values, options ...TmOperationOptions) ([]*TmProjectAssignment, error) {
	c := crudConfig{
		entityLabel:     "projectAssignments",
		endpoint:        types.OpenApiPathVcf + "projectAssignments",
		requiresTm:      true,
		minimumVersion:  "40.0",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmProjectAssignment](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	wrapped := make([]*TmProjectAssignment, len(result))
	for i := range result {
		wrapped[i] = &TmProjectAssignment{TmProjectAssignment: result[i], vcdClient: vcdClient, options: retainTmOptions(options)}
	}
	return wrapped, nil
}

// CreateTmProjectAssignment creates a new Project Assignment. Waits for task completion. Requires TM API
// 40.0+.
func (vcdClient *VCDClient) CreateTmProjectAssignment(config *types.TmProjectAssignment, options ...TmOperationOptions) (*TmProjectAssignment, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmProjectAssignment config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "projectAssignments",
		endpoint:       types.OpenApiPathVcf + "projectAssignments",
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	if config.ID == "" {
		return nil, fmt.Errorf("CreateTmProjectAssignment requires a project ID")
	}
	// Project IDs are supplied by the caller; the task owner need not identify the project.
	if err := runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config); err != nil {
		return nil, err
	}
	result, err := getInnerEntity[types.TmProjectAssignment](&vcdClient.Client, c.item(config.ID).withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmProjectAssignment{TmProjectAssignment: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// CreateTmProjectAssignmentAsync creates a new Project Assignment. Returns the tracking task without
// waiting. Requires TM API 40.0+.
func (vcdClient *VCDClient) CreateTmProjectAssignmentAsync(config *types.TmProjectAssignment, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmProjectAssignment config must not be nil")
	}
	if config.ID == "" {
		return nil, fmt.Errorf("CreateTmProjectAssignment requires a project ID")
	}
	c := crudConfig{
		entityLabel:    "projectAssignments",
		endpoint:       types.OpenApiPathVcf + "projectAssignments",
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// GetTmProjectAssignmentById retrieves a Project Assignment. Requires TM API 40.0+.
func (vcdClient *VCDClient) GetTmProjectAssignmentById(projectAssignmentId string, options ...TmOperationOptions) (*TmProjectAssignment, error) {
	c := crudConfig{
		entityLabel:    "projectAssignments",
		endpoint:       types.OpenApiPathVcf + "projectAssignments/%s",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{projectAssignmentId},
	}
	result, err := getInnerEntity[types.TmProjectAssignment](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmProjectAssignment{TmProjectAssignment: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// Update updates a Project Assignment. Waits for task completion. Requires TM API 40.0+.
func (o *TmProjectAssignment) Update(config *types.TmProjectAssignment, options ...TmOperationOptions) (*TmProjectAssignment, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "projectAssignments",
		endpoint:       types.OpenApiPathVcf + "projectAssignments/%s",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{o.TmProjectAssignment.ID},
	}
	if err := runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config); err != nil {
		return nil, err
	}
	result, err := getInnerEntity[types.TmProjectAssignment](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmProjectAssignment{TmProjectAssignment: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// UpdateAsync updates a Project Assignment. Returns the tracking task without waiting. Requires TM API
// 40.0+.
func (o *TmProjectAssignment) UpdateAsync(config *types.TmProjectAssignment, options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("Update config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "projectAssignments",
		endpoint:       types.OpenApiPathVcf + "projectAssignments/%s",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{o.TmProjectAssignment.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config)
}

// Delete deletes a Project Assignment. Waits for task completion. Requires TM API 40.0+.
func (o *TmProjectAssignment) Delete(options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "projectAssignments",
		endpoint:       types.OpenApiPathVcf + "projectAssignments/%s",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{o.TmProjectAssignment.ID},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// DeleteAsync deletes a Project Assignment. Returns the tracking task without waiting. Requires TM API
// 40.0+.
func (o *TmProjectAssignment) DeleteAsync(options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "projectAssignments",
		endpoint:       types.OpenApiPathVcf + "projectAssignments/%s",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{o.TmProjectAssignment.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// GetAllMembers retrieves all Project Assignment Members. Requires TM API 40.0+.
func (o *TmProjectAssignment) GetAllMembers(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmProjectAssignmentMember, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:     "projectAssignments/members",
		endpoint:        types.OpenApiPathVcf + "projectAssignments/%s/members",
		requiresTm:      true,
		minimumVersion:  "40.0",
		endpointParams:  []string{o.TmProjectAssignment.ID},
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmProjectAssignmentMember](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateMembers updates all Project Assignment Members. Waits for task completion. Requires TM API 40.0+.
func (o *TmProjectAssignment) UpdateMembers(config *types.TmProjectAssignmentMembers, options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return fmt.Errorf("UpdateMembers config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "projectAssignments/members",
		endpoint:       types.OpenApiPathVcf + "projectAssignments/%s/members",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{o.TmProjectAssignment.ID},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config)
}

// UpdateMembersAsync updates all Project Assignment Members. Returns the tracking task without waiting.
// Requires TM API 40.0+.
func (o *TmProjectAssignment) UpdateMembersAsync(config *types.TmProjectAssignmentMembers, options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	if config == nil {
		return nil, fmt.Errorf("UpdateMembers config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "projectAssignments/members",
		endpoint:       types.OpenApiPathVcf + "projectAssignments/%s/members",
		requiresTm:     true,
		minimumVersion: "40.0",
		endpointParams: []string{o.TmProjectAssignment.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPut, config)
}
