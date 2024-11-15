package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelZone = "Region Zone"

// Zone represents Region Zones
type Zone struct {
	Zone      *types.Zone
	vcdClient *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g Zone) wrap(inner *types.Zone) *Zone {
	g.Zone = inner
	return &g
}

// GetAllZones retrieves all Region Zones
func (vcdClient *VCDClient) GetAllZones(queryParameters url.Values) ([]*Zone, error) {
	c := crudConfig{
		entityLabel:     labelZone,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointZones,
		queryParameters: queryParameters,
	}

	outerType := Zone{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetZoneByName retrieves Region Zone by name
func (vcdClient *VCDClient) GetZoneByName(name string) (*Zone, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelZone)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllZones(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetZoneById(singleEntity.Zone.ID)
}

// GetZoneById retrieves Region Zone by ID
func (vcdClient *VCDClient) GetZoneById(id string) (*Zone, error) {
	c := crudConfig{
		entityLabel:    labelZone,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointZones,
		endpointParams: []string{id},
	}

	outerType := Zone{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetAllZones retrieves all Region Zones within a particular Region
func (r *Region) GetAllZones(queryParameters url.Values) ([]*Zone, error) {
	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("region.id=="+r.Region.ID, queryParams)

	return r.vcdClient.GetAllZones(queryParams)
}

// GetZoneByName retrieves Region Zone by name within a particular Region
func (r *Region) GetZoneByName(name string) (*Zone, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name ", labelZone)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	filteredEntities, err := r.GetAllZones(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return r.vcdClient.GetZoneById(singleEntity.Zone.ID)
}
