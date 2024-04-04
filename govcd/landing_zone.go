/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type SolutionLandingZone struct {
	SolutionLandingZoneType *SolutionLandingZoneType
	// DefinedEntity contains parent defined entity that contains SolutionLandingZoneType in
	// "Entity" field
	DefinedEntity *DefinedEntity
	vcdClient     *VCDClient
}

type SolutionLandingZoneType struct {
	Name     string     `json:"name,omitempty"`
	ID       string     `json:"id"`
	Catalogs []Catalogs `json:"catalogs"`
	Vdcs     []Vdcs     `json:"vdcs"`
}

type Catalogs struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Capabilities []any  `json:"capabilities"`
}

type Networks struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IsDefault    bool   `json:"isDefault"`
	Capabilities []any  `json:"capabilities"`
}

type StoragePolicies struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IsDefault    bool   `json:"isDefault"`
	Capabilities []any  `json:"capabilities"`
}

type ComputePolicies struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IsDefault    bool   `json:"isDefault"`
	Capabilities []any  `json:"capabilities"`
}

type Vdcs struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Capabilities    []any             `json:"capabilities"`
	IsDefault       bool              `json:"isDefault"`
	Networks        []Networks        `json:"networks"`
	StoragePolicies []StoragePolicies `json:"storagePolicies"`
	ComputePolicies []ComputePolicies `json:"computePolicies"`
}

func (vcdClient *VCDClient) CreateSolutionLandingZone(slzCfg *SolutionLandingZoneType) (*SolutionLandingZone, error) {
	// 1. Check that RDE type exists
	rdeType, err := vcdClient.GetRdeType("vmware", "solutions_organization", "1.0.0")
	if err != nil {
		return nil, fmt.Errorf("error retrieving RDE Type for Solution Landing zone: %s", err)
	}

	// 2. Convert more precise structure to fit DefinedEntity.DefinedEntity.Entity
	unmarshalledRdeEntityJson, err := convertSlzToRde(slzCfg)
	if err != nil {
		return nil, err
	}

	// 3. Construct payload
	entityCfg := &types.DefinedEntity{
		EntityType: "urn:vcloud:type:vmware:solutions_organization:1.0.0",
		Name:       "Solutions Organization",
		State:      addrOf("PRE_CREATED"),
		// Processed entity
		Entity: unmarshalledRdeEntityJson,
	}

	// 4. Create RDE
	createdRdeEntity, err := rdeType.CreateRde(*entityCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating RDE entity: %s", err)
	}

	// 5. Repackage created RDE "Entity" to more exact type
	result, err := convertRdeToSlz(createdRdeEntity.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	packages := SolutionLandingZone{
		SolutionLandingZoneType: result,
		vcdClient:               vcdClient,
		DefinedEntity:           createdRdeEntity,
	}

	return &packages, nil
}

func (slz *SolutionLandingZone) Update(slzCfg *SolutionLandingZoneType) (*SolutionLandingZone, error) {
	unmarshalledRdeEntityJson, err := convertSlzToRde(slzCfg)
	if err != nil {
		return nil, err
	}

	slz.DefinedEntity.DefinedEntity.Entity = unmarshalledRdeEntityJson

	err = slz.DefinedEntity.Update(*slz.DefinedEntity.DefinedEntity)
	if err != nil {
		return nil, err
	}

	result, err := convertRdeToSlz(slz.DefinedEntity.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	packages := SolutionLandingZone{
		SolutionLandingZoneType: result,
		vcdClient:               slz.vcdClient,
		DefinedEntity:           slz.DefinedEntity,
	}

	return &packages, nil
}

// Delete wraps
func (slz *SolutionLandingZone) Delete(slzCfg SolutionLandingZoneType) error {
	return slz.DefinedEntity.Delete()
}

func convertSlzToRde(slzCfg *SolutionLandingZoneType) (map[string]interface{}, error) {
	jsonText, err := json.Marshal(slzCfg)
	if err != nil {
		return nil, fmt.Errorf("error marshalling SLZ configuration :%s", err)
	}

	var unmarshalledRdeEntityJson map[string]interface{}
	err = json.Unmarshal(jsonText, &unmarshalledRdeEntityJson)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling SLZ configuration :%s", err)
	}

	return unmarshalledRdeEntityJson, nil
}

func convertRdeToSlz(content map[string]interface{}) (*SolutionLandingZoneType, error) {
	// content := createdRdeEntity.DefinedEntity.Entity
	jsonText2, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("error converting entity to SolutionLandingZone: %s", err)
	}

	result := &SolutionLandingZoneType{}
	err = json.Unmarshal(jsonText2, result)
	if err != nil {
		return nil, fmt.Errorf("error converting entity to SolutionLandingZone: %s", err)
	}

	return result, nil
}
