/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

var dataSolutionRdeType = [3]string{"vmware", "dsConfig", "0.1"}
var dseRightsBundleName = "vmware:dataSolutionsRightsBundle" // Rights bundle name
// Name of Data Solutions Operator package. It cannot be published itself, but it is still seen in
// the list.
var defaultDsoName = "VCD Data Solutions" // Data Solutions Operator (DSO) name

// DataSolution represents Data Solution entities and their repository configurations as can be seen
// in "Container Registry" UI view
type DataSolution struct {
	DataSolution  *types.DataSolution
	DefinedEntity *DefinedEntity
	vcdClient     *VCDClient
}

// GetAllDataSolutions retrieves all Data Solutions
func (vcdClient *VCDClient) GetAllDataSolutions(queryParameters url.Values) ([]*DataSolution, error) {
	allDseInstances, err := vcdClient.GetAllRdes(dataSolutionRdeType[0], dataSolutionRdeType[1], dataSolutionRdeType[2], queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Data Solutions: %s", err)
	}

	results := make([]*DataSolution, len(allDseInstances))
	for index, rde := range allDseInstances {
		dseConfig, err := convertRdeEntityToAny[types.DataSolution](rde.DefinedEntity.Entity)
		if err != nil {
			return nil, fmt.Errorf("error converting RDE to Data Solution: %s", err)
		}

		results[index] = &DataSolution{
			vcdClient:     vcdClient,
			DefinedEntity: rde,
			DataSolution:  dseConfig,
		}
	}

	return results, nil
}

// GetDataSolutionById retrieves Data Solution by ID
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

// GetDataSolutionByName retrieves Data Solution by Name
func (vcdClient *VCDClient) GetDataSolutionByName(name string) (*DataSolution, error) {
	dseEntities, err := vcdClient.GetAllDataSolutions(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Data Solution with name '%s': %s", name, err)
	}

	for _, instance := range dseEntities {
		if instance.DataSolution.Spec.Artifacts[0]["name"].(string) == name {
			return instance, nil
		}
	}
	return nil, fmt.Errorf("%s Data Solution by name '%s' not found", ErrorEntityNotFound, name)
}

// Update Data Solution with given configuration
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

// Publish Data Solution to a tenant
// It is a bundle of operations that mimics what UI does when performing "publish" operation
// * Publish rights bundle 'vmware:dataSolutionsRightsBundle' to the tenant
// * Provision access to particular Data Solution
// * Always provision access to special Data Solution 'VCD Data Solutions'. This is the package that
// will install Data Solutions Operator (DSO) to Kubernetes cluster
// * Publish all templates of the given instance. Note: UI will not show instance templates until the
// tenant installs Data Solutions Operator (DSO)
func (ds *DataSolution) Publish(tenantId string) (*types.DefinedEntityAccess, *types.DefinedEntityAccess, []*types.DefinedEntityAccess, error) {
	if tenantId == "" {
		return nil, nil, nil, fmt.Errorf("error - tenant ID empty")
	}

	if ds.Name() == defaultDsoName {
		return nil, nil, nil, fmt.Errorf("cannot publish %s", defaultDsoName)
	}

	// The operation is idempotent and can be run multiple times which is what the UI does
	err := ds.PublishRightsBundle([]string{tenantId})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error publishing Rights Bundle: %s", err)
	}

	// Publish ACLs to a given Data Solution
	mainAcl, err := ds.PublishAccessControls(tenantId)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error publishing Access Controls to '%s': %s", ds.Name(), err)
	}

	// Additionally set the same ACL for Data Solutions Operator (DSO)
	dso, err := ds.vcdClient.GetDataSolutionByName(defaultDsoName)
	if err != nil {
		return mainAcl, nil, nil, err
	}

	dsoAcl, err := dso.PublishAccessControls(tenantId)
	if err != nil {
		return mainAcl, nil, nil, fmt.Errorf("error publishing Access Controls to '%s': %s", dso.Name(), err)
	}

	// PublishAllInstanceTemplates
	templateAcls, err := ds.PublishAllInstanceTemplates(tenantId)
	if err != nil {
		return mainAcl, dsoAcl, nil, fmt.Errorf("error publishing all Data Solution Instance Templates: %s", err)
	}

	return mainAcl, dsoAcl, templateAcls, nil
}

// Unpublish Data Solution to a slice of tenants
// It is a bundle of operations that mimics what UI does when unpublishing and attempts to revert
// what is done in 'Publish' method
// * Remove access from given Data Solution
//
// Note. This method (and UI) is asymmetric in comparison to 'Publish' operation. It _does not_ do
// the following operations:
// * Unpublish all Data Solution Templates
// * Unpublish access for Data Solutions Operator (DSO)
// * Unpublish rights bundle 'vmware:dataSolutionsRightsBundle'
func (ds *DataSolution) Unpublish(tenantId string) error {
	if tenantId == "" {
		return fmt.Errorf("error - tenant ID empty")
	}

	// ACLs
	err := ds.UnpublishAccessControls(tenantId)
	if err != nil {
		return fmt.Errorf("failed unpublishing Access Controls for %s: %s", ds.Name(), err)
	}

	return nil
}

// PublishRightsBundle publishes "vmware:dataSolutionsRightsBundle" rights bundle
func (ds *DataSolution) PublishRightsBundle(tenantIds []string) error {
	rightsBundle, err := ds.vcdClient.Client.GetRightsBundleByName(dseRightsBundleName)
	if err != nil {
		return fmt.Errorf("error retrieving Rights Bundle %s: %s", dseRightsBundleName, err)
	}

	references := convertSliceOfStringsToOpenApiReferenceIds(tenantIds)
	err = rightsBundle.PublishTenants(references)
	if err != nil {
		return fmt.Errorf("error publishing Rights Bundle '%s' to Tenants '%s': %s",
			dseRightsBundleName, strings.Join(tenantIds, ","), err)
	}

	return nil
}

// UnpublishRightsBundle removes "vmware:dataSolutionsRightsBundle" rights bundle from tenant
func (ds *DataSolution) UnpublishRightsBundle(tenantIds []string) error {
	rightsBundle, err := ds.vcdClient.Client.GetRightsBundleByName(dseRightsBundleName)
	if err != nil {
		return fmt.Errorf("error retrieving Rights Bundle %s: %s", dseRightsBundleName, err)
	}

	references := convertSliceOfStringsToOpenApiReferenceIds(tenantIds)
	err = rightsBundle.UnpublishTenants(references)
	if err != nil {
		return fmt.Errorf("error unpublishing %s for Tenants '%s': %s",
			dseRightsBundleName, strings.Join(tenantIds, ","), err)
	}

	return nil
}

// PublishAccessControls provisions ACL for a given tenant
func (ds *DataSolution) PublishAccessControls(tenantId string) (*types.DefinedEntityAccess, error) {
	acl := &types.DefinedEntityAccess{
		Tenant:        types.OpenApiReference{ID: tenantId},
		GrantType:     "MembershipAccessControlGrant",
		AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
		MemberID:      tenantId,
	}

	accessControl, err := ds.DefinedEntity.SetAccessControl(acl)
	if err != nil {
		return nil, fmt.Errorf("error setting Access Control for Data Solution '%s', Org ID %s: %s", tenantId, ds.Name(), err)
	}

	return accessControl, nil
}

// UnpublishAccessControls removes ACLs for a given tenant
func (ds *DataSolution) UnpublishAccessControls(tenantId string) error {
	acls, err := ds.GetAllAccessControlsForTenant(tenantId)
	if err != nil {
		return fmt.Errorf("error retrieving all Access Controls for Tenant '%s': %s", tenantId, err)
	}

	for _, acl := range acls {
		err = ds.DefinedEntity.DeleteAccessControl(acl)
		if err != nil {
			return fmt.Errorf("error deleting Access Control: %s", err)
		}
	}

	return nil
}

// GetAllAccessControls for a Data Solution
func (ds *DataSolution) GetAllAccessControls(queryParameters url.Values) ([]*types.DefinedEntityAccess, error) {
	allAcls, err := ds.DefinedEntity.GetAllAccessControls(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Access Controls for Data Solution %s: %s", ds.Name(), err)
	}

	return localFilter("Data Solution ACL", allAcls, "ObjectId", ds.RdeId())
}

// GetAccessControlById retrieves ACL by ID
func (ds *DataSolution) GetAccessControlById(id string) (*types.DefinedEntityAccess, error) {
	return ds.DefinedEntity.GetAccessControlById(id)
}

// GetAllAccessControlsForTenant retrieves all ACLs that apply for a specific Tenant
func (ds *DataSolution) GetAllAccessControlsForTenant(tenantId string) ([]*types.DefinedEntityAccess, error) {
	util.Logger.Printf("[TRACE] Data Solution '%s' getting Access Controls for tenant '%s'", ds.Name(), tenantId)
	allAcls, err := ds.GetAllAccessControls(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Access Controls for Data Solution: %s", err)
	}

	foundAcls := make([]*types.DefinedEntityAccess, 0)
	util.Logger.Printf("[TRACE] Data Solution '%s' looking for Access Controls for tenant '%s'", ds.Name(), tenantId)
	for _, acl := range allAcls {
		util.Logger.Printf("[TRACE] Data Solution '%s' checking Access Control ID '%s'", ds.Name(), acl.Id)
		if acl.Tenant.ID == tenantId {
			util.Logger.Printf("[TRACE] Data Solution '%s' Access Control '%s' matches tenant '%s'", ds.Name(), acl.Id, tenantId)
			foundAcls = append(foundAcls, acl)
		}
	}

	return foundAcls, nil
}

// RdeId is a shorthand function to retrieve parent RDE ID for a Data Solution.
func (dsCfg *DataSolution) RdeId() string {
	if dsCfg == nil || dsCfg.DefinedEntity == nil || dsCfg.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return dsCfg.DefinedEntity.DefinedEntity.ID
}

// Name extracts the name from inside RDE configuration. This name is used in UI and is not always
// the same as RDE name. It is guaranteed to persist.
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
