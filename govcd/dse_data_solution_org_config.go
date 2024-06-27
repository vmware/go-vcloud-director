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

var dataSolutionOrgConfig = [3]string{"vmware", "dsOrgConfig", "0.1.1"}

// DataSolutionOrgConfig structure represents Data Solution Org Configuration. This configuration
// can carry additional information to Data Solutions if they require it. At the moment of
// implementation the only known Data Solution that requires this configuration is `Confluent
// Platform` that uses this structure to store licensing data.
type DataSolutionOrgConfig struct {
	DataSolutionOrgConfig *types.DataSolutionOrgConfig
	DefinedEntity         *DefinedEntity
	vcdClient             *VCDClient
}

// CreateDataSolutionOrgConfig creates Data Solution Org Configuration for a defined orgId
func (vcdClient *VCDClient) CreateDataSolutionOrgConfig(orgId string, cfg *types.DataSolutionOrgConfig) (*DataSolutionOrgConfig, error) {
	rdeType, err := vcdClient.GetRdeType(dataSolutionOrgConfig[0], dataSolutionOrgConfig[1], dataSolutionOrgConfig[2])
	if err != nil {
		return nil, fmt.Errorf("error retrieving RDE Type for VCD Data Solution Org Configuration: %s", err)
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

// GetAllDataSolutionOrgConfigs retrieves all available Data Solution Org Configs
func (vcdClient *VCDClient) GetAllDataSolutionOrgConfigs(queryParameters url.Values) ([]*DataSolutionOrgConfig, error) {
	allDseInstances, err := vcdClient.GetAllRdes(dataSolutionOrgConfig[0], dataSolutionOrgConfig[1], dataSolutionOrgConfig[2], queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Data Solution Org Configs: %s", err)
	}

	results := make([]*DataSolutionOrgConfig, len(allDseInstances))
	for index, rde := range allDseInstances {
		dsOrgConfig, err := convertRdeEntityToAny[types.DataSolutionOrgConfig](rde.DefinedEntity.Entity)
		if err != nil {
			return nil, fmt.Errorf("error converting RDE to Solution Add-on Instance: %s", err)
		}

		results[index] = &DataSolutionOrgConfig{
			vcdClient:             vcdClient,
			DefinedEntity:         rde,
			DataSolutionOrgConfig: dsOrgConfig,
		}
	}

	return results, nil
}

// GetAllDataSolutionOrgConfigs retrieves all available Data Solution Org Configs for a given Data
// Solution
func (ds *DataSolution) GetAllDataSolutionOrgConfigs() ([]*DataSolutionOrgConfig, error) {
	if ds == nil || ds.DataSolution == nil {
		return nil, fmt.Errorf("error - Data Solution structure is empty")
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.spec.solutionType==%s", ds.DataSolution.Spec.SolutionType), queryParams)
	queryParams = queryParameterFilterAnd("state==RESOLVED", queryParams)

	return ds.vcdClient.GetAllDataSolutionOrgConfigs(queryParams)
}

// GetDataSolutionOrgConfigForTenant retrieves all available Data Solution Org Configs for a given
// Data Solution and then uses local filter to find the one for given tenantId
func (ds *DataSolution) GetDataSolutionOrgConfigForTenant(tenantId string) (*DataSolutionOrgConfig, error) {
	if ds == nil || ds.DataSolution == nil {
		return nil, fmt.Errorf("error - Data Solution structure is empty")
	}

	if tenantId == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	allOrgConfigs, err := ds.GetAllDataSolutionOrgConfigs()
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Data Solution Org Configs: %s", err)
	}

	var foundOrgCfg *DataSolutionOrgConfig
	for _, orgCfg := range allOrgConfigs {
		// tenant ID is stored in RDE Name
		if orgCfg.DefinedEntity.DefinedEntity.Name == tenantId {
			foundOrgCfg = orgCfg
			break
		}
	}

	if foundOrgCfg == nil {
		return nil, fmt.Errorf("%s: could not find Data Solution '%s' Org Config for a given tenant '%s'",
			ErrorEntityNotFound, ds.Name(), tenantId)
	}

	return foundOrgCfg, nil

}

// Delete Data Solution Org Config
func (dsOrgCfg *DataSolutionOrgConfig) Delete() error {
	if dsOrgCfg.DefinedEntity == nil {
		return fmt.Errorf("error - parent Defined Entity is nil")
	}
	return dsOrgCfg.DefinedEntity.Delete()
}

// RdeId is a shortcut of SolutionEntity.DefinedEntity.DefinedEntity.ID
func (dsOrgCfg *DataSolutionOrgConfig) RdeId() string {
	if dsOrgCfg == nil || dsOrgCfg.DefinedEntity == nil || dsOrgCfg.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return dsOrgCfg.DefinedEntity.DefinedEntity.ID
}
