/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

var dataSolutionTemplateInstanceRdeType = [3]string{"vmware", "dsInstanceTemplate", "0.1"}

// DataSolutionInstanceTemplate represents Data Solution Instance Templates that come with Data Solutions
type DataSolutionInstanceTemplate struct {
	DataSolutionInstanceTemplate *types.DataSolutionInstanceTemplate
	DefinedEntity                *DefinedEntity
	vcdClient                    *VCDClient
}

// GetAllInstanceTemplates retrieves all Data Solution Instance Templates that are available in the
// system
func (vcdClient *VCDClient) GetAllInstanceTemplates(queryParameters url.Values) ([]*DataSolutionInstanceTemplate, error) {
	allDseInstanceTemplates, err := vcdClient.GetAllRdes(dataSolutionTemplateInstanceRdeType[0], dataSolutionTemplateInstanceRdeType[1], dataSolutionTemplateInstanceRdeType[2], queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Data Solution Instance Templates: %s", err)
	}

	results := make([]*DataSolutionInstanceTemplate, len(allDseInstanceTemplates))
	for index, rde := range allDseInstanceTemplates {
		dseConfig, err := convertRdeEntityToAny[types.DataSolutionInstanceTemplate](rde.DefinedEntity.Entity)
		if err != nil {
			return nil, fmt.Errorf("error converting RDE to Data Solution Instance Template: %s", err)
		}

		results[index] = &DataSolutionInstanceTemplate{
			vcdClient:                    vcdClient,
			DefinedEntity:                rde,
			DataSolutionInstanceTemplate: dseConfig,
		}
	}

	return results, nil
}

// GetAllInstanceTemplates retrieves all Data Solution Instance Templates for a given Data Solution
func (ds *DataSolution) GetAllInstanceTemplates() ([]*DataSolutionInstanceTemplate, error) {
	if ds == nil || ds.DataSolution == nil {
		return nil, fmt.Errorf("error - Data Solution structure is empty")
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.spec.solutionType==%s", ds.DataSolution.Spec.SolutionType), queryParams)
	queryParams = queryParameterFilterAnd("state==RESOLVED", queryParams)

	return ds.vcdClient.GetAllInstanceTemplates(queryParams)
}

// PublishAllInstanceTemplates creates Access Controls for all available Data Solution Instance Templates
func (ds *DataSolution) PublishAllInstanceTemplates(tenantId string) ([]*types.DefinedEntityAccess, error) {
	allTemplates, err := ds.GetAllInstanceTemplates()
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Data Solution Instance Templates: %s", err)
	}

	definedEntityAccess := make([]*types.DefinedEntityAccess, len(allTemplates))
	for templateIndex, template := range allTemplates {
		access, err := template.Publish(tenantId)
		if err != nil {
			return nil, fmt.Errorf("error setting ACL for Data Solution Instance Template '%s': %s",
				template.DefinedEntity.DefinedEntity.Name, err)
		}

		definedEntityAccess[templateIndex] = access
	}

	return definedEntityAccess, nil
}

// UnPublishAllInstanceTemplates removes all ACLs of a given Data Solution from specified tenantId
func (ds *DataSolution) UnPublishAllInstanceTemplates(tenantId string) error {
	allTemplates, err := ds.GetAllInstanceTemplates()
	if err != nil {
		return fmt.Errorf("error retrieving all Data Solution Instance Templates: %s", err)
	}

	for _, template := range allTemplates {
		err := template.Unpublish(tenantId)
		if err != nil {
			return fmt.Errorf("error removing ACL for Data Solution Instance Template '%s': %s",
				template.DefinedEntity.DefinedEntity.Name, err)
		}

	}

	return nil
}

// Name extracts the name from inside RDE configuration. This name is used in UI.
func (dst *DataSolutionInstanceTemplate) Name() string {
	if dst.DefinedEntity == nil || dst.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return dst.DefinedEntity.DefinedEntity.Name
}

// GetAllAccessControls retrieves all Access Controls for a given Data Solution Instance Template
func (dst *DataSolutionInstanceTemplate) GetAllAccessControls(queryParameters url.Values) ([]*types.DefinedEntityAccess, error) {
	return dst.DefinedEntity.GetAllAccessControls(queryParameters)
}

// GetAllAccessControlsForTenant retrieves all Access Controls for a given tenant
func (dst *DataSolutionInstanceTemplate) GetAllAccessControlsForTenant(tenantId string) ([]*types.DefinedEntityAccess, error) {
	util.Logger.Printf("[TRACE] Data Solution Instance Template '%s' getting Access Controls for tenant '%s'", dst.Name(), tenantId)
	allAcls, err := dst.GetAllAccessControls(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Access Controls for Data Solution Solution Instance Template: %s", err)
	}

	foundAcls := make([]*types.DefinedEntityAccess, 0)
	util.Logger.Printf("[TRACE] Data Solution Instance Template '%s' looking for Access Controls for tenant '%s'", dst.Name(), tenantId)
	for _, acl := range allAcls {
		util.Logger.Printf("[TRACE] Data Solution Instance Template '%s' checking Access Control ID '%s'", dst.Name(), acl.Id)
		if acl.Tenant.ID == tenantId {
			util.Logger.Printf("[TRACE] Data Solution Instance Template '%s' Access Control '%s' matches tenant '%s'", dst.Name(), acl.Id, tenantId)
			foundAcls = append(foundAcls, acl)
		}
	}

	return foundAcls, nil
}

// Publish a single Data Solution Instance Template to a given tenant
func (dst *DataSolutionInstanceTemplate) Publish(tenantId string) (*types.DefinedEntityAccess, error) {
	acl := &types.DefinedEntityAccess{
		Tenant:        types.OpenApiReference{ID: tenantId},
		GrantType:     "MembershipAccessControlGrant",
		AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
		MemberID:      tenantId,
	}

	accessControl, err := dst.DefinedEntity.SetAccessControl(acl)
	if err != nil {
		return nil, fmt.Errorf("error setting Access Control for Data Solution %s: %s", dst.DefinedEntity.DefinedEntity.Name, err)
	}

	return accessControl, nil
}

// Unpublish a single Data Solution Instance Template for a given tenant
func (dst *DataSolutionInstanceTemplate) Unpublish(tenantId string) error {
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("tenant.id==%s", tenantId), queryParams)
	acls, err := dst.DefinedEntity.GetAllAccessControls(queryParams)
	if err != nil {
		return fmt.Errorf("error getting Access Control for Data Solution Instance Template %s: %s", dst.DefinedEntity.DefinedEntity.Name, err)
	}

	for _, acl := range acls {
		err = dst.DefinedEntity.DeleteAccessControl(acl)
		if err != nil {
			return fmt.Errorf("error deleting Access Control: %s", err)
		}
	}

	return nil
}

// RdeId is a shortcut of SolutionEntity.DefinedEntity.DefinedEntity.ID
func (dst *DataSolutionInstanceTemplate) RdeId() string {
	if dst == nil || dst.DefinedEntity == nil || dst.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return dst.DefinedEntity.DefinedEntity.ID
}
