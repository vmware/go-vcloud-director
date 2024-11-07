package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelZone = "Zone"

// Zone represents region zones
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

// func (vcdClient *VCDClient) CreateZone(config *types.Zone) (*Zone, error) {
// 	c := crudConfig{
// 		entityLabel: labelZone,
// 		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointZones,
// 	}
// 	outerType := Zone{vcdClient: vcdClient}
// 	return createOuterEntity(&vcdClient.Client, outerType, c, config)
// }

func (vcdClient *VCDClient) GetAllZones(queryParameters url.Values) ([]*Zone, error) {
	c := crudConfig{
		entityLabel:     labelZone,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointZones,
		queryParameters: queryParameters,
	}

	outerType := Zone{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

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

func (vcdClient *VCDClient) GetZoneById(id string) (*Zone, error) {
	c := crudConfig{
		entityLabel:    labelZone,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointZones,
		endpointParams: []string{id},
	}

	outerType := Zone{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (r *Region) GetZoneByName(name string) (*Zone, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name ", labelZone)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("region.id=="+r.Region.ID, queryParams)

	filteredEntities, err := r.vcdClient.GetAllZones(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return r.vcdClient.GetZoneById(singleEntity.Zone.ID)
}

func (o *Zone) Update(ZoneConfig *types.Zone) (*Zone, error) {
	c := crudConfig{
		entityLabel:    labelZone,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointZones,
		endpointParams: []string{o.Zone.ID},
	}
	outerType := Zone{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, ZoneConfig)
}

func (o *Zone) Delete() error {
	c := crudConfig{
		entityLabel:    labelZone,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointZones,
		endpointParams: []string{o.Zone.ID},
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
