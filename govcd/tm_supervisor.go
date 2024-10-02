package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelSupervisor = "Supervisor"

// Supervisor is a type for handling VCF Supervisors
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

func (vcdClient *VCDClient) GetAllSupervisors(queryParameters url.Values) ([]*Supervisor, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointSupervisors,
		entityLabel:     labelSupervisor,
		queryParameters: queryParameters,
	}

	outerType := Supervisor{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetSupervisorById(id string) (*Supervisor, error) {
	c := crudConfig{
		entityLabel:    labelSupervisor,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointSupervisors,
		endpointParams: []string{id},
	}

	outerType := Supervisor{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

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
