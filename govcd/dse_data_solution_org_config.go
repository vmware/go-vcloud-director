/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var dataSolutionOrgConfig = [3]string{"vmware", "dsOrgConfig", "0.1.1"}

type DataSolutionOrgConfig struct {
	DataSolutionOrgConfig *types.DataSolutionOrgConfig
	DefinedEntity         *DefinedEntity
	vcdClient             *VCDClient
}

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
