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

const labelTmAviServiceEngineGroup = "TM Avi Service Engine Group"

// TmAviServiceEngineGroup manages an Avi Load Balancer Service Engine Group.
type TmAviServiceEngineGroup struct {
	TmAviServiceEngineGroup *types.TmAviServiceEngineGroup
	vcdClient               *VCDClient
	options                 []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmAviServiceEngineGroup) wrap(inner *types.TmAviServiceEngineGroup) *TmAviServiceEngineGroup {
	g.TmAviServiceEngineGroup = inner
	return &g
}

// tmAviServiceEngineGroupConfig returns the CRUD configuration for Avi Load Balancer Service Engine Group operations. Optional params are appended to the endpoint.
func tmAviServiceEngineGroupConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmAviServiceEngineGroup,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmAviServiceEngineGroups,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "40.1",
	}.withTmOptions(options)
}

// GetAllTmAviServiceEngineGroups retrieves all Avi Load Balancer Service Engine Groups. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmAviServiceEngineGroups(queryParameters url.Values, options ...TmOperationOptions) ([]*TmAviServiceEngineGroup, error) {
	c := tmAviServiceEngineGroupConfig(options)
	c.queryParameters = queryParameters
	outerType := TmAviServiceEngineGroup{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmAviServiceEngineGroupByName retrieves an Avi Load Balancer Service Engine Group by name. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviServiceEngineGroupByName(name string, options ...TmOperationOptions) (*TmAviServiceEngineGroup, error) {
	entities, err := vcdClient.GetAllTmAviServiceEngineGroups(nil, options...)
	if err != nil {
		return nil, err
	}
	return tmEntityByName(labelTmAviServiceEngineGroup, name, entities, func(e *TmAviServiceEngineGroup) string { return e.TmAviServiceEngineGroup.Name })
}

// GetTmAviServiceEngineGroupById retrieves an Avi Load Balancer Service Engine Group by ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviServiceEngineGroupById(id string, options ...TmOperationOptions) (*TmAviServiceEngineGroup, error) {
	outerType := TmAviServiceEngineGroup{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmAviServiceEngineGroupConfig(options, id))
}

// GetTmAviServiceEngineGroupByNameOrId retrieves an Avi Load Balancer Service Engine Group by name or ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviServiceEngineGroupByNameOrId(identifier string, options ...TmOperationOptions) (*TmAviServiceEngineGroup, error) {
	return tmEntityByNameOrId(identifier,
		func(id string) (*TmAviServiceEngineGroup, error) {
			return vcdClient.GetTmAviServiceEngineGroupById(id, options...)
		},
		func(name string) (*TmAviServiceEngineGroup, error) {
			return vcdClient.GetTmAviServiceEngineGroupByName(name, options...)
		})
}

// #nosec G101 -- This is a label, not a credential
const labelTmAviServiceEngineGroupAssignment = "TM Avi Service Engine Group Assignment"

// TmAviServiceEngineGroupAssignment manages an Avi Load Balancer Service Engine Group Assignment.
type TmAviServiceEngineGroupAssignment struct {
	TmAviServiceEngineGroupAssignment *types.TmAviServiceEngineGroupAssignment
	vcdClient                         *VCDClient
	options                           []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmAviServiceEngineGroupAssignment) wrap(inner *types.TmAviServiceEngineGroupAssignment) *TmAviServiceEngineGroupAssignment {
	g.TmAviServiceEngineGroupAssignment = inner
	return &g
}

// tmAviServiceEngineGroupAssignmentConfig returns the CRUD configuration for Avi Load Balancer Service Engine Group Assignment operations. Optional params are appended to the endpoint.
func tmAviServiceEngineGroupAssignmentConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmAviServiceEngineGroupAssignment,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmAviServiceEngineGroupAssignments,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "40.1",
	}.withTmOptions(options)
}

// CreateTmAviServiceEngineGroupAssignment creates an Avi Load Balancer Service Engine Group Assignment. Waits for task completion. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmAviServiceEngineGroupAssignment(config *types.TmAviServiceEngineGroupAssignment, options ...TmOperationOptions) (*TmAviServiceEngineGroupAssignment, error) {
	outerType := TmAviServiceEngineGroupAssignment{vcdClient: vcdClient, options: retainTmOptions(options)}
	return createOuterEntity(&vcdClient.Client, outerType, tmAviServiceEngineGroupAssignmentConfig(options), config)
}

// CreateTmAviServiceEngineGroupAssignmentAsync creates an Avi Load Balancer Service Engine Group Assignment. Returns the tracking task without waiting. Requires TM API 40.1+.
func (vcdClient *VCDClient) CreateTmAviServiceEngineGroupAssignmentAsync(config *types.TmAviServiceEngineGroupAssignment, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("CreateTmAviServiceEngineGroupAssignment config must not be nil")
	}
	return createInnerEntityAsync(&vcdClient.Client, tmAviServiceEngineGroupAssignmentConfig(options), config)
}

// GetAllTmAviServiceEngineGroupAssignments retrieves all Avi Load Balancer Service Engine Group Assignments. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmAviServiceEngineGroupAssignments(queryParameters url.Values, options ...TmOperationOptions) ([]*TmAviServiceEngineGroupAssignment, error) {
	c := tmAviServiceEngineGroupAssignmentConfig(options)
	c.queryParameters = queryParameters
	outerType := TmAviServiceEngineGroupAssignment{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmAviServiceEngineGroupAssignmentById retrieves an Avi Load Balancer Service Engine Group Assignment by ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmAviServiceEngineGroupAssignmentById(id string, options ...TmOperationOptions) (*TmAviServiceEngineGroupAssignment, error) {
	outerType := TmAviServiceEngineGroupAssignment{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmAviServiceEngineGroupAssignmentConfig(options, id))
}

// Delete deletes an Avi Load Balancer Service Engine Group Assignment. Waits for task completion. Requires TM API 40.1+.
func (o *TmAviServiceEngineGroupAssignment) Delete(options ...TmOperationOptions) error {
	options = inheritTmOptions(o.options, options)
	c := tmAviServiceEngineGroupAssignmentConfig(options, o.TmAviServiceEngineGroupAssignment.ID)
	return deleteEntityById(&o.vcdClient.Client, c)
}

// DeleteAsync deletes an Avi Load Balancer Service Engine Group Assignment. Returns the tracking task without waiting. Requires TM API 40.1+.
func (o *TmAviServiceEngineGroupAssignment) DeleteAsync(options ...TmOperationOptions) (*Task, error) {
	options = inheritTmOptions(o.options, options)
	c := tmAviServiceEngineGroupAssignmentConfig(options, o.TmAviServiceEngineGroupAssignment.ID)
	return runOpenApiOperationAsync(&o.vcdClient.Client, c, http.MethodDelete, nil)
}
