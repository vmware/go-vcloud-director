package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelOpenApiUser = "User"

type OpenApiUser struct {
	User          *types.OpenApiUser
	vcdClient     *VCDClient
	TenantContext *TenantContext
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g OpenApiUser) wrap(inner *types.OpenApiUser) *OpenApiUser {
	g.User = inner
	return &g
}

// CreateUser creates a new User with a given configuration
func (vcdClient *VCDClient) CreateUser(config *types.OpenApiUser, ctx *TenantContext) (*OpenApiUser, error) {
	c := crudConfig{
		entityLabel:      labelOpenApiUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointUsers,
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}
	outerType := OpenApiUser{vcdClient: vcdClient, TenantContext: ctx}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllUsers retrieves all Users with an optional filter
func (vcdClient *VCDClient) GetAllUsers(queryParameters url.Values, ctx *TenantContext) ([]*OpenApiUser, error) {
	c := crudConfig{
		entityLabel:      labelOpenApiUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointUsers,
		queryParameters:  queryParameters,
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}

	outerType := OpenApiUser{vcdClient: vcdClient, TenantContext: ctx}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetUserByName retrieves User by Name
func (vcdClient *VCDClient) GetUserByName(username string, ctx *TenantContext) (*OpenApiUser, error) {
	if username == "" {
		return nil, fmt.Errorf("%s lookup requires username", labelOpenApiUser)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "username=="+username)

	filteredEntities, err := vcdClient.GetAllUsers(queryParams, ctx)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("username", username, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetUserById(singleEntity.User.ID, ctx)
}

// GetUserById retrieves User by ID
func (vcdClient *VCDClient) GetUserById(id string, ctx *TenantContext) (*OpenApiUser, error) {
	c := crudConfig{
		entityLabel:      labelOpenApiUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointUsers,
		endpointParams:   []string{id},
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}

	outerType := OpenApiUser{vcdClient: vcdClient, TenantContext: ctx}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update User with a given config
func (o *OpenApiUser) Update(cfg *types.OpenApiUser) (*OpenApiUser, error) {
	c := crudConfig{
		entityLabel:      labelOpenApiUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointUsers,
		endpointParams:   []string{o.User.ID},
		additionalHeader: getTenantContextHeader(o.TenantContext),
		requiresTm:       true,
	}
	outerType := OpenApiUser{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, cfg)
}

// Delete User
func (o *OpenApiUser) Delete() error {
	c := crudConfig{
		entityLabel:      labelOpenApiUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointUsers,
		endpointParams:   []string{o.User.ID},
		additionalHeader: getTenantContextHeader(o.TenantContext),
		requiresTm:       true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
