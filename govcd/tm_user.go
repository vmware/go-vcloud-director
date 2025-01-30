package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmUser = "User"

type TmUser struct {
	User      *types.TmUser
	vcdClient *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmUser) wrap(inner *types.TmUser) *TmUser {
	g.User = inner
	return &g
}

// CreateUser creates a new User with a given configuration
func (vcdClient *VCDClient) CreateUser(config *types.TmUser) (*TmUser, error) {
	c := crudConfig{
		entityLabel: labelTmUser,
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		requiresTm:  true,
	}
	outerType := TmUser{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllUsers retrieves all Users with an optional filter
func (vcdClient *VCDClient) GetAllUsers(queryParameters url.Values) ([]*TmUser, error) {
	c := crudConfig{
		entityLabel:     labelTmUser,
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmUser{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetUserByName retrieves User by Name
func (vcdClient *VCDClient) GetUserByName(name string) (*TmUser, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmUser)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllUsers(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetUserById(singleEntity.User.ID)
}

// GetUserById retrieves User by ID
func (vcdClient *VCDClient) GetUserById(id string) (*TmUser, error) {
	c := crudConfig{
		entityLabel:    labelTmUser,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmUser{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetUserByNameAndOrgId retrieves User by Name and Org ID
func (vcdClient *VCDClient) GetUserByNameAndOrgId(name, orgId string) (*TmUser, error) {
	if name == "" || orgId == "" {
		return nil, fmt.Errorf("%s lookup requires name and Org ID", labelTmUser)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("orgRef.id=="+orgId, queryParams)

	filteredEntities, err := vcdClient.GetAllUsers(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetUserById(singleEntity.User.ID)
}

// Update User with a given config
func (o *TmUser) Update(UserConfig *types.TmUser) (*TmUser, error) {
	c := crudConfig{
		entityLabel:    labelTmUser,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		endpointParams: []string{o.User.ID},
		requiresTm:     true,
	}
	outerType := TmUser{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, UserConfig)
}

// Delete User
func (o *TmUser) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmUser,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		endpointParams: []string{o.User.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
