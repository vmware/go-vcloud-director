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

// slzRdeType sets Runtime Defined Entity Type to be used across multiple calls
var slzRdeType = [3]string{"vmware", "solutions_organization", "1.0.0"}

// SolutionLandingZone controls VCD Solution Add-On Landing Zone. It does so by wrapping RDE for
// entity types vmware:solutions_organization:1.0.0.
//
// Up to VCD 10.5.1.1 ,there can only be one single RDE instance for landing zone.
type SolutionLandingZone struct {
	// SolutionLandingZoneType defines internal content of RDE (`types.DefinedEntity.State`)
	SolutionLandingZoneType *types.SolutionLandingZoneType
	// DefinedEntity contains parent defined entity that contains SolutionLandingZoneType in
	// "Entity" field
	DefinedEntity *DefinedEntity
	vcdClient     *VCDClient
}

// CreateSolutionLandingZone configures VCD Solution Add-On Landing Zone. It does so by performing
// the following steps:
//
// 1. Creates Solution Landing Zone RDE based on type urn:vcloud:type:vmware:solutions_organization:1.0.0
// 2. Resolves the RDE
func (vcdClient *VCDClient) CreateSolutionLandingZone(slzCfg *types.SolutionLandingZoneType) (*SolutionLandingZone, error) {
	// 1. Check that RDE type exists
	rdeType, err := vcdClient.GetRdeType(slzRdeType[0], slzRdeType[1], slzRdeType[2])
	if err != nil {
		return nil, fmt.Errorf("error retrieving RDE Type for Solution Landing zone: %s", err)
	}

	// 2. Convert more precise structure to fit DefinedEntity.DefinedEntity.Entity
	unmarshalledRdeEntityJson, err := convertAnyToRdeEntity(slzCfg)
	if err != nil {
		return nil, err
	}

	// 3. Construct payload
	entityCfg := &types.DefinedEntity{
		EntityType: "urn:vcloud:type:" + strings.Join(slzRdeType[:], ":"),
		Name:       "Solutions Organization",
		State:      addrOf("PRE_CREATED"),
		// Processed solution landing zone
		Entity: unmarshalledRdeEntityJson,
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

	result, err := convertRdeEntityToAny[types.SolutionLandingZoneType](createdRdeEntity.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	returnType := SolutionLandingZone{
		SolutionLandingZoneType: result,
		vcdClient:               vcdClient,
		DefinedEntity:           createdRdeEntity,
	}

	return &returnType, nil
}

// GetAllSolutionLandingZones retrieves all Solution Landing Zones
//
// Note: Up to VCD 10.5.1.1 there can be only a single RDE entry (one SLZ per VCD)
func (vcdClient *VCDClient) GetAllSolutionLandingZones(queryParameters url.Values) ([]*SolutionLandingZone, error) {
	allSlzs, err := vcdClient.GetAllRdes(slzRdeType[0], slzRdeType[1], slzRdeType[2], queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all SLZs: %s", err)
	}

	results := make([]*SolutionLandingZone, len(allSlzs))
	for slzRdeIndex, slzRde := range allSlzs {

		slz, err := convertRdeEntityToAny[types.SolutionLandingZoneType](slzRde.DefinedEntity.Entity)
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

// GetExactlyOneSolutionLandingZone will get single Solution Landing Zone RDE or fail.
// There can be only one Solution Landing Zone in VCD, but because it is backed by RDE - it can
// occur that due to some error there is more than one RDE Entity
func (vcdClient *VCDClient) GetExactlyOneSolutionLandingZone() (*SolutionLandingZone, error) {
	allSlzs, err := vcdClient.GetAllSolutionLandingZones(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Solution Landing Zones: %s", err)
	}

	return oneOrError("rde", strings.Join(slzRdeType[:], ":"), allSlzs)
}

// GetSolutionLandingZoneById retrieves Solution Landing Zone by ID
//
// Note: defined entity ID must be used that can be accessed either by `SolutionLandingZone.Id()`
// method or directly in `SolutionLandingZone.DefinedEntity.DefinedEntity.ID` field
func (vcdClient *VCDClient) GetSolutionLandingZoneById(id string) (*SolutionLandingZone, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be specified")
	}
	rde, err := getRdeById(&vcdClient.Client, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving RDE by ID: %s", err)
	}

	result, err := convertRdeEntityToAny[types.SolutionLandingZoneType](rde.DefinedEntity.Entity)
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

// Refresh reloads parent RDE data
func (slz *SolutionLandingZone) Refresh() error {
	err := slz.DefinedEntity.Refresh()
	if err != nil {
		return err
	}

	// Repackage created RDE "Entity" to more exact type
	result, err := convertRdeEntityToAny[types.SolutionLandingZoneType](slz.DefinedEntity.DefinedEntity.Entity)
	if err != nil {
		return err
	}

	slz.SolutionLandingZoneType = result

	return nil
}

// RdeId is a shorthand to retrieve ID of parent runtime defined entity
func (slz *SolutionLandingZone) RdeId() string {
	if slz == nil || slz.DefinedEntity == nil || slz.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return slz.DefinedEntity.DefinedEntity.ID
}

// Update Solution Landing Zone
func (slz *SolutionLandingZone) Update(slzCfg *types.SolutionLandingZoneType) (*SolutionLandingZone, error) {
	unmarshalledRdeEntityJson, err := convertAnyToRdeEntity(slzCfg)
	if err != nil {
		return nil, err
	}

	slz.DefinedEntity.DefinedEntity.Entity = unmarshalledRdeEntityJson

	err = slz.DefinedEntity.Update(*slz.DefinedEntity.DefinedEntity)
	if err != nil {
		return nil, err
	}

	result, err := convertRdeEntityToAny[types.SolutionLandingZoneType](slz.DefinedEntity.DefinedEntity.Entity)
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

// Delete removes the RDE that defines Solution Landing Zone
func (slz *SolutionLandingZone) Delete() error {
	return slz.DefinedEntity.Delete()
}
