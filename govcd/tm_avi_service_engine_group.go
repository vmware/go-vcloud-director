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

// TmAviServiceEngineGroup manages a Tenant Manager AviServiceEngineGroup.
type TmAviServiceEngineGroup struct {
	TmAviServiceEngineGroup *types.TmAviServiceEngineGroup
	vcdClient               *VCDClient
	options                 []TmOperationOptions
}

// TmAviServiceEngineGroupAssignment manages a Tenant Manager AviServiceEngineGroupAssignment.
type TmAviServiceEngineGroupAssignment struct {
	TmAviServiceEngineGroupAssignment *types.TmAviServiceEngineGroupAssignment
	vcdClient                         *VCDClient
	options                           []TmOperationOptions
}

// GetAllTmAviServiceEngineGroups retrieves all Avi Load Balancer Service Engine Groups in the system.
// Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmAviServiceEngineGroups(queryParameters url.Values, options ...TmOperationOptions) ([]*TmAviServiceEngineGroup, error) {
	c := crudConfig{
		entityLabel:     "loadBalancer/aviServiceEngineGroups",
		endpoint:        types.OpenApiPathVcf + "loadBalancer/aviServiceEngineGroups",
		requiresTm:      true,
		minimumVersion:  "40.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmAviServiceEngineGroup](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	wrapped := make([]*TmAviServiceEngineGroup, len(result))
	for i := range result {
		wrapped[i] = &TmAviServiceEngineGroup{TmAviServiceEngineGroup: result[i], vcdClient: vcdClient, options: retainTmOptions(options)}
	}
	return wrapped, nil
}

// GetTmAviServiceEngineGroupById retrieves Avi Load Balancer Service Engine Group. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviServiceEngineGroupById(aviServiceEngineGroupId string, options ...TmOperationOptions) (*TmAviServiceEngineGroup, error) {
	c := crudConfig{
		entityLabel:    "loadBalancer/aviServiceEngineGroups",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviServiceEngineGroups/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{aviServiceEngineGroupId},
	}
	result, err := getInnerEntity[types.TmAviServiceEngineGroup](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmAviServiceEngineGroup{TmAviServiceEngineGroup: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// CreateTmAviServiceEngineGroupAssignment creates a new Avi Service Engine Group Assignment. Waits for task
// completion. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmAviServiceEngineGroupAssignment(config *types.TmAviServiceEngineGroupAssignment, options ...TmOperationOptions) (*TmAviServiceEngineGroupAssignment, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmAviServiceEngineGroupAssignment config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "loadBalancer/aviServiceEngineGroupsAssignments",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviServiceEngineGroupsAssignments",
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	result, err := createInnerEntity(&vcdClient.Client, c.withTmOptions(options), config)
	if err != nil {
		return nil, err
	}
	return &TmAviServiceEngineGroupAssignment{TmAviServiceEngineGroupAssignment: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// CreateTmAviServiceEngineGroupAssignmentAsync creates a new Avi Service Engine Group Assignment. Returns
// the tracking task without waiting. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmAviServiceEngineGroupAssignmentAsync(config *types.TmAviServiceEngineGroupAssignment, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmAviServiceEngineGroupAssignment config must not be nil")
	}
	c := crudConfig{
		entityLabel:    "loadBalancer/aviServiceEngineGroupsAssignments",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviServiceEngineGroupsAssignments",
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, config)
}

// GetAllTmAviServiceEngineGroupAssignments retrieves the assignments for Avi Load Balancer Service Engine
// Groups in the system. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmAviServiceEngineGroupAssignments(queryParameters url.Values, options ...TmOperationOptions) ([]*TmAviServiceEngineGroupAssignment, error) {
	c := crudConfig{
		entityLabel:     "loadBalancer/aviServiceEngineGroupsAssignments",
		endpoint:        types.OpenApiPathVcf + "loadBalancer/aviServiceEngineGroupsAssignments",
		requiresTm:      true,
		minimumVersion:  "40.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmAviServiceEngineGroupAssignment](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	wrapped := make([]*TmAviServiceEngineGroupAssignment, len(result))
	for i := range result {
		wrapped[i] = &TmAviServiceEngineGroupAssignment{TmAviServiceEngineGroupAssignment: result[i], vcdClient: vcdClient, options: retainTmOptions(options)}
	}
	return wrapped, nil
}

// GetTmAviServiceEngineGroupAssignmentById retrieves a Avi Load Balancer Service Engine Group Assignment.
// Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviServiceEngineGroupAssignmentById(assignmentId string, options ...TmOperationOptions) (*TmAviServiceEngineGroupAssignment, error) {
	c := crudConfig{
		entityLabel:    "loadBalancer/aviServiceEngineGroupsAssignments",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviServiceEngineGroupsAssignments/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{assignmentId},
	}
	result, err := getInnerEntity[types.TmAviServiceEngineGroupAssignment](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmAviServiceEngineGroupAssignment{TmAviServiceEngineGroupAssignment: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// Delete deletes the specified Avi Service Engine Group Assignment. Waits for task completion. Requires TM
// API 40.1+.
func (o *TmAviServiceEngineGroupAssignment) Delete(options ...TmOperationOptions) error {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "loadBalancer/aviServiceEngineGroupsAssignments",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviServiceEngineGroupsAssignments/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmAviServiceEngineGroupAssignment.ID},
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}

// DeleteAsync deletes the specified Avi Service Engine Group Assignment. Returns the tracking task without
// waiting. Requires TM API 40.1+.
func (o *TmAviServiceEngineGroupAssignment) DeleteAsync(options ...TmOperationOptions) (*Task, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "loadBalancer/aviServiceEngineGroupsAssignments",
		endpoint:       types.OpenApiPathVcf + "loadBalancer/aviServiceEngineGroupsAssignments/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmAviServiceEngineGroupAssignment.ID},
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodDelete, nil)
}
