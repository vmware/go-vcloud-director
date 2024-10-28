package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelSupervisorZone = "Supervisor Zone"

// Supervisor is a type for reading Supervisor Zones
type SupervisorZone struct {
	SupervisorZone *types.SupervisorZone
	vcdClient      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (s SupervisorZone) wrap(inner *types.SupervisorZone) *SupervisorZone {
	s.SupervisorZone = inner
	return &s
}

// GetAllSupervisorZones retrieves all Supervisor Zones in a given Supervisor
func (s *Supervisor) GetAllSupervisorZones(queryParameters url.Values) ([]*SupervisorZone, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointSupervisorZones,
		entityLabel:     labelSupervisorZone,
		queryParameters: queryParameters,
	}

	outerType := SupervisorZone{vcdClient: s.vcdClient}
	return getAllOuterEntities(&s.vcdClient.Client, outerType, c)
}

// GetSupervisorZoneById retrieves Supervisor by id
func (s *Supervisor) GetSupervisorZoneById(id string) (*SupervisorZone, error) {
	if id == "" {
		return nil, fmt.Errorf("'id' must be set")
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd("supervisor.id=="+s.Supervisor.SupervisorID, queryParams)

	c := crudConfig{
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointSupervisorZones,
		entityLabel:     labelSupervisorZone,
		queryParameters: queryParams,
		endpointParams:  []string{id},
	}

	outerType := SupervisorZone{vcdClient: s.vcdClient}
	return getOuterEntity(&s.vcdClient.Client, outerType, c)
}

// GetSupervisorZoneByName retrieves Supervisor Zone by a given name
func (s *Supervisor) GetSupervisorZoneByName(name string) (*SupervisorZone, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelSupervisor)
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd("supervisor.id=="+s.Supervisor.SupervisorID, queryParams)
	queryParams = queryParameterFilterAnd("name=="+name, queryParams)

	filteredEntities, err := s.GetAllSupervisorZones(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleEntity, nil
}
