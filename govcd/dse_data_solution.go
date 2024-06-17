/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var dataSolutionRdeType = [3]string{"vmware", "dsConfig", "0.1"}
var dataSolutionTemplateInstanceRdeType = [3]string{"vmware", "dsInstanceTemplate", "0.1"}
var dataSolutionOrgConfig = [3]string{"vmware", "dsOrgConfig", "0.1.1"}
var dseRightsBundleName = "vmware:dataSolutionsRightsBundle" // Rights bundle name
var defaultDsoName = "VCD Data Solutions"                    // Name of Data Solutions Operator package.

type DataSolution struct {
	DataSolution  *types.DataSolution
	DefinedEntity *DefinedEntity
	vcdClient     *VCDClient
}

type DataSolutionInstanceTemplate struct {
	DataSolutionInstanceTemplate *types.DataSolutionInstanceTemplate
	DefinedEntity                *DefinedEntity
	vcdClient                    *VCDClient
}

type DataSolutionOrgConfig struct {
	DataSolutionOrgConfig *types.DataSolutionOrgConfig
	DefinedEntity         *DefinedEntity
	vcdClient             *VCDClient
}

func (dsCfg *DataSolution) RdeId() string {
	if dsCfg == nil || dsCfg.DefinedEntity == nil || dsCfg.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return dsCfg.DefinedEntity.DefinedEntity.ID
}

// Name extracts the name from inside RDE configuration. This name is used in UI and is not always
// the same as RDE name.
func (dsCfg *DataSolution) Name() string {
	if dsCfg.DataSolution == nil || dsCfg.DataSolution.Spec.Artifacts == nil || len(dsCfg.DataSolution.Spec.Artifacts) < 1 {
		return ""
	}

	nameString, ok := dsCfg.DataSolution.Spec.Artifacts[0]["name"].(string)
	if !ok {
		return ""
	}

	return nameString
}

func (vcdClient *VCDClient) GetAllDataSolutions(queryParameters url.Values) ([]*DataSolution, error) {
	allDseInstances, err := vcdClient.GetAllRdes(dataSolutionRdeType[0], dataSolutionRdeType[1], dataSolutionRdeType[2], queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Solution Add-on Instances: %s", err)
	}

	results := make([]*DataSolution, len(allDseInstances))
	for index, rde := range allDseInstances {
		dseConfig, err := convertRdeEntityToAny[types.DataSolution](rde.DefinedEntity.Entity)
		if err != nil {
			return nil, fmt.Errorf("error converting RDE to Solution Add-on Instance: %s", err)
		}

		results[index] = &DataSolution{
			vcdClient:     vcdClient,
			DefinedEntity: rde,
			DataSolution:  dseConfig,
		}
	}

	return results, nil
}

func (vcdClient *VCDClient) GetDataSolutionById(id string) (*DataSolution, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be specified")
	}
	rde, err := getRdeById(&vcdClient.Client, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Data Solution by ID: %s", err)
	}

	result, err := convertRdeEntityToAny[types.DataSolution](rde.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	packages := &DataSolution{
		DataSolution:  result,
		vcdClient:     vcdClient,
		DefinedEntity: rde,
	}

	return packages, nil
}

func (vcdClient *VCDClient) GetDataSolutionByName(name string) (*DataSolution, error) {
	dseEntities, err := vcdClient.GetAllDataSolutions(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving configuration item with name '%s': %s", name, err)
	}

	for _, instance := range dseEntities {
		if instance.DataSolution.Spec.Artifacts[0]["name"].(string) == name {
			return instance, nil
		}
	}
	return nil, fmt.Errorf("%s DSE Config Instance by name '%s' not found", ErrorEntityNotFound, name)
}

func (ds *DataSolution) Update(cfg *types.DataSolution) (*DataSolution, error) {
	unmarshalledRdeEntityJson, err := convertAnyToRdeEntity(cfg)
	if err != nil {
		return nil, err
	}

	ds.DefinedEntity.DefinedEntity.Entity = unmarshalledRdeEntityJson
	err = ds.DefinedEntity.Update(*ds.DefinedEntity.DefinedEntity)
	if err != nil {
		return nil, err
	}

	result, err := convertRdeEntityToAny[types.DataSolution](ds.DefinedEntity.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	packages := DataSolution{
		DataSolution:  result,
		vcdClient:     ds.vcdClient,
		DefinedEntity: ds.DefinedEntity,
	}

	return &packages, nil
}

// PublishRightsBundle publishes "vmware:dataSolutionsRightsBundle" rights bundle that is required
func (ds *DataSolution) PublishRightsBundle(tenantId string) error {
	rightsBundle, err := ds.vcdClient.Client.GetRightsBundleByName(dseRightsBundleName)
	if err != nil {
		return fmt.Errorf("error retrieving Rights Bundle %s: %s", dseRightsBundleName, err)
	}

	reference := []types.OpenApiReference{{
		ID: tenantId,
	}}

	err = rightsBundle.PublishTenants(reference)
	if err != nil {
		return fmt.Errorf("error publishing %s to Tenant '%s': %s",
			dseRightsBundleName, tenantId, err)
	}

	return nil
}

// UnpublishRightsBundle removes "vmware:dataSolutionsRightsBundle" rights bundle from tenant
func (ds *DataSolution) UnpublishRightsBundle(tenantId string) error {
	rightsBundle, err := ds.vcdClient.Client.GetRightsBundleByName(dseRightsBundleName)
	if err != nil {
		return fmt.Errorf("error retrieving Rights Bundle %s: %s", dseRightsBundleName, err)
	}

	reference := []types.OpenApiReference{{
		ID: tenantId,
	}}

	err = rightsBundle.UnpublishTenants(reference)
	if err != nil {
		return fmt.Errorf("error unpublishing %s from Tenant '%s': %s",
			dseRightsBundleName, tenantId, err)
	}

	return nil
}

// Publish Data Solution to a slice of tenants
// It is a bundle of operations that mimics what UI does
func (ds *DataSolution) Publish(tenantId string) error {
	err := ds.PublishRightsBundle(tenantId)
	if err != nil {
		return fmt.Errorf("error publishing Rights Bundle: %s", err)
	}

	// Publish ACLs to a given Data Solution
	_, err = ds.PublishAcls(tenantId)
	if err != nil {
		return fmt.Errorf("error Publishing Access Rights to '%s': %s", ds.Name(), err)
	}

	// If current Data Solution is not "VCD Data Solutions" - additionally set the same ACL for it
	// Note. Names will not change.
	if ds.Name() != defaultDsoName {
		dso, err := ds.vcdClient.GetDataSolutionByName(defaultDsoName)
		if err != nil {
			return err
		}

		_, err = dso.PublishAcls(tenantId)
		if err != nil {
			return fmt.Errorf("error Publishing Access Rights to '%s': %s", dso.Name(), err)
		}
	}

	// Publish all templates of a given

	dsInstanceTemplates, err := ds.GetAllInstanceTemplates()

	// acl := &types.DefinedEntityAccess{
	// 	Tenant:        types.OpenApiReference{ID: tenantId},
	// 	GrantType:     "MembershipAccessControlGrant",
	// 	AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
	// 	MemberID:      tenantId,
	// }

	// accessControl, err := ds.DefinedEntity.SetAccessControl(acl)
	// if err != nil {
	// 	return nil, fmt.Errorf("error setting Access Control for Data Solution %s: %s", ds.Name(), err)
	// }

	return nil
}

func (ds *DataSolution) PublishAcls(tenantId string) (*types.DefinedEntityAccess, error) {
	acl := &types.DefinedEntityAccess{
		Tenant:        types.OpenApiReference{ID: tenantId},
		GrantType:     "MembershipAccessControlGrant",
		AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
		MemberID:      tenantId,
	}

	accessControl, err := ds.DefinedEntity.SetAccessControl(acl)
	if err != nil {
		return nil, fmt.Errorf("error setting Access Control for Data Solution %s: %s", ds.Name(), err)
	}

	return accessControl, nil
}

// ####

func (vcdClient *VCDClient) GetAllInstanceTemplates(queryParameters url.Values) ([]*DataSolutionInstanceTemplate, error) {
	allDseInstanceTemplates, err := vcdClient.GetAllRdes(dataSolutionTemplateInstanceRdeType[0], dataSolutionTemplateInstanceRdeType[1], dataSolutionTemplateInstanceRdeType[2], queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Solution Add-on Instances: %s", err)
	}

	results := make([]*DataSolutionInstanceTemplate, len(allDseInstanceTemplates))
	for index, rde := range allDseInstanceTemplates {
		dseConfig, err := convertRdeEntityToAny[types.DataSolutionInstanceTemplate](rde.DefinedEntity.Entity)
		if err != nil {
			return nil, fmt.Errorf("error converting RDE to Solution Add-on Instance: %s", err)
		}

		results[index] = &DataSolutionInstanceTemplate{
			vcdClient:                    vcdClient,
			DefinedEntity:                rde,
			DataSolutionInstanceTemplate: dseConfig,
		}
	}

	return results, nil
}

func (ds *DataSolution) GetAllInstanceTemplates() ([]*DataSolutionInstanceTemplate, error) {
	if ds == nil || ds.DataSolution == nil {
		return nil, fmt.Errorf("error - Data Solution structure is empty")
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.spec.solutionType==%s", ds.DataSolution.Spec.SolutionType), queryParams)
	queryParams = queryParameterFilterAnd("state==RESOLVED", queryParams)

	return ds.vcdClient.GetAllInstanceTemplates(queryParams)
}

func (dse *DataSolutionInstanceTemplate) RdeId() string {
	if dse == nil || dse.DefinedEntity == nil || dse.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return dse.DefinedEntity.DefinedEntity.ID
}

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

// ###########

func (vcdClient *VCDClient) CreateDataSolutionOrgConfig(orgId string, cfg *types.DataSolutionOrgConfig) (*DataSolutionOrgConfig, error) {
	rdeType, err := vcdClient.GetRdeType(dataSolutionOrgConfig[0], dataSolutionOrgConfig[1], dataSolutionOrgConfig[2])
	if err != nil {
		return nil, fmt.Errorf("error retrieving RDE Type for VCD Data Solutions: %s", err)
	}

	// 2. Convert more precise structure to fit DefinedEntity.DefinedEntity.Entity
	unmarshalledRdeEntityJson, err := convertAnyToRdeEntity(cfg)
	if err != nil {
		return nil, err
	}

	// 3. Construct payload
	entityCfg := &types.DefinedEntity{
		EntityType: "urn:vcloud:type:" + strings.Join(dataSolutionOrgConfig[:], ":"),
		Name:       orgId, // Receiving Org ID is used as name
		State:      addrOf("PRE_CREATED"),
		Entity:     unmarshalledRdeEntityJson,
	}

	// 4. Create RDE
	createdRdeEntity, err := rdeType.CreateRde(*entityCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating RDE entity: %s", err)
	}

	// 5. Resolve RDE
	err = createdRdeEntity.Resolve()
	if err != nil {
		return nil, fmt.Errorf("error resolving Solutions add-on after creating: %s", err)
	}

	// 6. Reload RDE
	err = createdRdeEntity.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing RDE after resolving: %s", err)
	}

	result, err := convertRdeEntityToAny[types.DataSolutionOrgConfig](createdRdeEntity.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	returnType := DataSolutionOrgConfig{
		DataSolutionOrgConfig: result,
		vcdClient:             vcdClient,
		DefinedEntity:         createdRdeEntity,
	}

	return &returnType, nil
}
