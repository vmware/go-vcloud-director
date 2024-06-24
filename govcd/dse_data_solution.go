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
var defaultDsoName = "VCD Data Solutions"

type DataSolution struct {
	DataSolution  *types.DataSolution
	DefinedEntity *DefinedEntity
	vcdClient     *VCDClient
}

// RdeId is a shorthand function retrieve parent RDE ID for a Data Solution.
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

// Publish Data Solution to a slice of tenants
// It is a bundle of operations that mimics what UI does
func (ds *DataSolution) Publish(tenantIds []string) error {
	if ds.Name() == defaultDsoName {
		return fmt.Errorf("cannot publish Data Solutions Operator")
	}

	err := ds.PublishRightsBundle(tenantIds)
	if err != nil {
		return fmt.Errorf("error publishing Rights Bundle: %s", err)
	}

	// Publish ACLs to a given Data Solution
	_, err = ds.PublishAccessControls(tenantIds)
	if err != nil {
		return fmt.Errorf("error publishing Access Rights to '%s': %s", ds.Name(), err)
	}

	// If current Data Solution is not "VCD Data Solutions" - additionally set the same ACL for it
	// Note. Names will not change.
	dso, err := ds.vcdClient.GetDataSolutionByName(defaultDsoName)
	if err != nil {
		return err
	}

	_, err = dso.PublishAccessControls(tenantIds)
	if err != nil {
		return fmt.Errorf("error publishing Access Rights to '%s': %s", dso.Name(), err)
	}

	// PublishAllInstanceTemplates

	_, err = ds.PublishAllInstanceTemplates(tenantIds)
	if err != nil {
		return fmt.Errorf("error publishing ")
	}

	return nil
}

// Publish Data Solution to a slice of tenants
// It is a bundle of operations that mimics what UI does
func (ds *DataSolution) Unpublish(tenantIds []string) error {
	if ds.Name() == defaultDsoName {
		return fmt.Errorf("cannot unpublish %s", defaultDsoName)
	}

	// Templates

	// ACLs
	err := ds.UnpublishAccessControls(tenantIds)
	if err != nil {
		return fmt.Errorf("failed unpublishing Access Controls for %s: %s", ds.Name(), err)
	}

	// err := ds.PublishRightsBundle(tenantIds)
	// if err != nil {
	// 	return fmt.Errorf("error publishing Rights Bundle: %s", err)
	// }

	// Publish ACLs to a given Data Solution
	/* _, err := ds.PublishAccessControls(tenantIds)
	if err != nil {
		return fmt.Errorf("error publishing Access Rights to '%s': %s", ds.Name(), err)
	}

	// If current Data Solution is not "VCD Data Solutions" - additionally set the same ACL for it
	// Note. Names will not change.
	dso, err := ds.vcdClient.GetDataSolutionByName(defaultDsoName)
	if err != nil {
		return err
	}

	_, err = dso.PublishAccessControls(tenantIds)
	if err != nil {
		return fmt.Errorf("error publishing Access Rights to '%s': %s", dso.Name(), err)
	}

	// PublishAllInstanceTemplates

	_, err = ds.PublishAllInstanceTemplates(tenantIds)
	if err != nil {
		return fmt.Errorf("error publishing ")
	} */

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
		return fmt.Errorf("error publishing %s to Tenants '%s': %s",
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
	err = rightsBundle.PublishTenants(references)
	if err != nil {
		return fmt.Errorf("error publishing %s to Tenants '%s': %s",
			dseRightsBundleName, strings.Join(tenantIds, ","), err)
	}

	return nil
}

func (ds *DataSolution) PublishAccessControls(tenantIds []string) ([]*types.DefinedEntityAccess, error) {
	acls := make([]*types.DefinedEntityAccess, 0)

	for _, tenantId := range tenantIds {
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

		acls = append(acls, accessControl)
	}
	return acls, nil
}

func (ds *DataSolution) UnpublishAccessControls(tenantIds []string) error {
	acls, err := ds.GetAllAccessControlsForTenants(tenantIds)
	if err != nil {
		return fmt.Errorf("error retrieving all Access Controls for tenants '%s': %s", strings.Join(tenantIds, ","), err)
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
	return ds.DefinedEntity.GetAllAccessControls(queryParameters)
}

func (ds *DataSolution) GetAllAccessControlsForTenants(tenantIds []string) ([]*types.DefinedEntityAccess, error) {
	util.Logger.Printf("[TRACE] Data Solution '%s' getting Access Controls for tenants '%s'", ds.Name(), strings.Join(tenantIds, ","))
	allAcls, err := ds.GetAllAccessControls(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Access Controls for Data Solution: %s", err)
	}

	foundAcls := make([]*types.DefinedEntityAccess, 0)
	for _, tenantId := range tenantIds {
		util.Logger.Printf("[TRACE] Data Solution '%s' looking Access Controls for tenant '%s'", ds.Name(), tenantId)
		for _, acl := range allAcls {
			util.Logger.Printf("[TRACE] Data Solution '%s' checking Access Control ID '%s'", ds.Name(), acl.Id)
			if acl.Tenant.ID == tenantId {
				util.Logger.Printf("[TRACE] Data Solution '%s' Access Control '%s' matches tenant '%s'", ds.Name(), acl.Id, tenantId)
				foundAcls = append(foundAcls, acl)
			}
		}
	}

	return foundAcls, nil
}

// func (ds *DataSolution) PublishAllInstanceTemplates(tenantIds []string) ([]*types.DefinedEntityAccess, error) {
// 	allTemplates, err := ds.GetAllInstanceTemplates()
// 	if err != nil {
// 		return nil, fmt.Errorf("error retrieving all Data Solution Instance Templates: %s", err)
// 	}

// 	definedEntityAccess := make([]*types.DefinedEntityAccess, 0)
// 	for _, tenantId := range tenantIds {
// 		for _, template := range allTemplates {
// 			access, err := template.PublishAccessControls(tenantId)
// 			if err != nil {
// 				return nil, fmt.Errorf("error setting ACL to Data Solution Instance Template '%s': %s",
// 					template.DefinedEntity.DefinedEntity.Name, err)
// 			}

// 			definedEntityAccess = append(definedEntityAccess, access)
// 		}
// 	}

// 	return definedEntityAccess, nil
// }

// ####
/*  */
// func (dst *DataSolutionInstanceTemplate) Publish(tenantId string) (*types.DefinedEntityAccess, error) {
// 	acl := &types.DefinedEntityAccess{
// 		Tenant:        types.OpenApiReference{ID: tenantId},
// 		GrantType:     "MembershipAccessControlGrant",
// 		AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
// 		MemberID:      tenantId,
// 	}

// 	accessControl, err := dst.DefinedEntity.SetAccessControl(acl)
// 	if err != nil {
// 		return nil, fmt.Errorf("error setting Access Control for Data Solution %s: %s", dst.DefinedEntity.DefinedEntity.Name, err)
// 	}

// 	return accessControl, nil
// }

// ###########
