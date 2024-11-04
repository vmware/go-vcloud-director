package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelSupervisor = "Supervisor"

// Supervisor is a type for reading available Supervisors
type Supervisor struct {
	Supervisor *types.Supervisor
	vcdClient  *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (s Supervisor) wrap(inner *types.Supervisor) *Supervisor {
	s.Supervisor = inner
	return &s
}

// GetAllSupervisors retrieves all available Supervisors
func (vcdClient *VCDClient) GetAllSupervisors(queryParameters url.Values) ([]*Supervisor, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointSupervisors,
		entityLabel:     labelSupervisor,
		queryParameters: queryParameters,
	}

	outerType := Supervisor{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetSupervisorById retrieves supervisor by ID
func (vcdClient *VCDClient) GetSupervisorById(id string) (*Supervisor, error) {
	c := crudConfig{
		entityLabel:    labelSupervisor,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointSupervisors,
		endpointParams: []string{id},
	}

	outerType := Supervisor{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetSupervisorByName retrieves Supervisor by name
func (vcdClient *VCDClient) GetSupervisorByName(name string) (*Supervisor, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelSupervisor)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllSupervisors(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleEntity, nil
}

// GetAllSupervisors returns all Supervisors that are available in this vCenter
func (v *VCenter) GetAllSupervisors(queryParameters url.Values) ([]*Supervisor, error) {
	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("virtualCenter.id==%s", v.VSphereVCenter.VcId), queryParams)
	return v.client.GetAllSupervisors(queryParams)
}

// GetSupervisorByName retrieves Supervisor by name in a given vCenter server
func (v *VCenter) GetSupervisorByName(name string) (*Supervisor, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("name==%s", name), queryParams)
	s, err := v.GetAllSupervisors(queryParams)
	if err != nil {
		return nil, err
	}

	return oneOrError("name", name, s)
}
