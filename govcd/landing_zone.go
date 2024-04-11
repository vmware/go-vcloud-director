/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelSolutionLandingZone = "Solution Landing Zone"

type SolutionLandingZone struct {
	SolutionLandingZoneType *types.SolutionLandingZoneType
	// DefinedEntity contains parent defined entity that contains SolutionLandingZoneType in
	// "Entity" field
	DefinedEntity *DefinedEntity
	vcdClient     *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (s SolutionLandingZone) wrap(inner *types.SolutionLandingZoneType) *SolutionLandingZone {
	s.SolutionLandingZoneType = inner
	return &s
}

// type SolutionLandingZoneType struct {
// 	Name     string     `json:"name,omitempty"`
// 	ID       string     `json:"id"`
// 	Catalogs []Catalogs `json:"catalogs"`
// 	Vdcs     []Vdcs     `json:"vdcs"`
// }

// type Catalogs struct {
// 	ID           string   `json:"id"`
// 	Name         string   `json:"name,omitempty"`
// 	Capabilities []string `json:"capabilities"`
// }

// type Networks struct {
// 	ID           string   `json:"id"`
// 	Name         string   `json:"name,omitempty"`
// 	IsDefault    bool     `json:"isDefault"`
// 	Capabilities []string `json:"capabilities"`
// }

// type StoragePolicies struct {
// 	ID           string   `json:"id"`
// 	Name         string   `json:"name,omitempty"`
// 	IsDefault    bool     `json:"isDefault"`
// 	Capabilities []string `json:"capabilities"`
// }

// type ComputePolicies struct {
// 	ID           string   `json:"id"`
// 	Name         string   `json:"name,omitempty"`
// 	IsDefault    bool     `json:"isDefault"`
// 	Capabilities []string `json:"capabilities"`
// }

// type Vdcs struct {
// 	ID              string            `json:"id"`
// 	Name            string            `json:"name,omitempty"`
// 	Capabilities    []string          `json:"capabilities"`
// 	IsDefault       bool              `json:"isDefault"`
// 	Networks        []Networks        `json:"networks"`
// 	StoragePolicies []StoragePolicies `json:"storagePolicies"`
// 	ComputePolicies []ComputePolicies `json:"computePolicies"`
// }

// UI does the following:
// Creates a new defined entity
// 1. POST https://HOST/cloudapi/1.0.0/entityTypes/urn:vcloud:type:vmware:solutions_organization:1.0.0
// 2. Retrieves all entities GET https://HOST/cloudapi/1.0.0/entities/types/vmware/solutions_organization
// 3. Resolves the entity POST https://HOST/cloudapi/1.0.0/entities/urn:vcloud:entity:vmware:solutions_organization:19032ea5-3e49-44e7-b601-5e37ecbbd190/resolve
// 4. Retrieves all entities GET https://HOST/cloudapi/1.0.0/entities/types/vmware/solutions_organization (probably to check if state is resolved)
// 5. Retrieves all behaviors GET https://HOST/cloudapi/1.0.0/entityTypes/urn:vcloud:type:vmware:solutions_add_on_instance:1.0.0/behaviors
// 6. Queries for existing solution add-ons - GET https://HOST/cloudapi/1.0.0/entities/types/vmware/solutions_add_on?filter=(entity.manifest.name!=solution-addon-landing-zone);(entity.manifest.name!=solutions-agent);(entity.manifest.name!=autoscale)
func (vcdClient *VCDClient) CreateSolutionLandingZone(slzCfg *types.SolutionLandingZoneType) (*SolutionLandingZone, error) {
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

	err = createdRdeEntity.Resolve()
	if err != nil {
		return nil, fmt.Errorf("error resolving Solutions add-on after creating: %s", err)
	}

	err = createdRdeEntity.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing RDE after resolving: %s", err)
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

func (vcdClient *VCDClient) GetAllSolutionLandingZones(queryParameters url.Values) ([]*SolutionLandingZone, error) {
	allSlzs, err := vcdClient.GetAllRdes("vmware", "solutions_organization", "1.0.0", queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all SLZs: %s", err)
	}

	results := make([]*SolutionLandingZone, len(allSlzs))
	for slzRdeIndex, slzRde := range allSlzs {

		slz, err := convertRdeToSlz(slzRde.DefinedEntity.Entity)
		if err != nil {
			return nil, fmt.Errorf("error converting RDE to SLZ: %s", err)
		}

		results[slzRdeIndex] = &SolutionLandingZone{
			vcdClient:               vcdClient,
			DefinedEntity:           slzRde,
			SolutionLandingZoneType: slz,
		}
	}

	return results, nil
}

func (vcdClient *VCDClient) GetSolutionLandingZoneById(id string) (*SolutionLandingZone, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be specified")
	}
	rde, err := getRdeById(&vcdClient.Client, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving RDE by ID: %s", err)
	}

	result, err := convertRdeToSlz(rde.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	packages := &SolutionLandingZone{
		SolutionLandingZoneType: result,
		vcdClient:               vcdClient,
		DefinedEntity:           rde,
	}

	return packages, nil
}

func (slz *SolutionLandingZone) Refresh() error {
	err := slz.DefinedEntity.Refresh()
	if err != nil {
		return err
	}

	// 5. Repackage created RDE "Entity" to more exact type
	result, err := convertRdeToSlz(slz.DefinedEntity.DefinedEntity.Entity)
	if err != nil {
		return err
	}

	slz.SolutionLandingZoneType = result

	return nil
}

func (slz *SolutionLandingZone) Update(slzCfg *types.SolutionLandingZoneType) (*SolutionLandingZone, error) {
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
func (slz *SolutionLandingZone) Delete() error {
	return slz.DefinedEntity.Delete()
}

func convertSlzToRde(slzCfg *types.SolutionLandingZoneType) (map[string]interface{}, error) {
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

func convertRdeToSlz(content map[string]interface{}) (*types.SolutionLandingZoneType, error) {
	jsonText2, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("error converting entity to SolutionLandingZone: %s", err)
	}

	result := &types.SolutionLandingZoneType{}
	err = json.Unmarshal(jsonText2, result)
	if err != nil {
		return nil, fmt.Errorf("error converting entity to SolutionLandingZone: %s", err)
	}

	return result, nil
}
