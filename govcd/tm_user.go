package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmUser = "User"

type TmUser struct {
	User          *types.TmUser
	vcdClient     *VCDClient
	TenantContext *TenantContext
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmUser) wrap(inner *types.TmUser) *TmUser {
	g.User = inner
	return &g
}

// CreateUser creates a new User with a given configuration
func (vcdClient *VCDClient) CreateUser(config *types.TmUser, ctx *TenantContext) (*TmUser, error) {
	c := crudConfig{
		entityLabel:      labelTmUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}
	outerType := TmUser{vcdClient: vcdClient, TenantContext: ctx}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllUsers retrieves all Users with an optional filter
func (vcdClient *VCDClient) GetAllUsers(queryParameters url.Values, ctx *TenantContext) ([]*TmUser, error) {
	c := crudConfig{
		entityLabel:      labelTmUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		queryParameters:  queryParameters,
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}

	outerType := TmUser{vcdClient: vcdClient, TenantContext: ctx}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetUserByName retrieves User by Name
func (vcdClient *VCDClient) GetUserByName(name string, ctx *TenantContext) (*TmUser, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmUser)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllUsers(queryParams, ctx)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetUserById(singleEntity.User.ID, ctx)
}

// GetUserById retrieves User by ID
func (vcdClient *VCDClient) GetUserById(id string, ctx *TenantContext) (*TmUser, error) {
	c := crudConfig{
		entityLabel:      labelTmUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		endpointParams:   []string{id},
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}

	outerType := TmUser{vcdClient: vcdClient, TenantContext: ctx}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update User with a given config
func (o *TmUser) Update(cfg *types.TmUser) (*TmUser, error) {
	c := crudConfig{
		entityLabel:      labelTmUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		endpointParams:   []string{o.User.ID},
		additionalHeader: getTenantContextHeader(o.TenantContext),
		requiresTm:       true,
	}
	outerType := TmUser{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, cfg)
}

// Delete User
func (o *TmUser) Delete() error {
	c := crudConfig{
		entityLabel:      labelTmUser,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsers,
		endpointParams:   []string{o.User.ID},
		additionalHeader: getTenantContextHeader(o.TenantContext),
		requiresTm:       true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}

// ChangePassword of a user
func (o *TmUser) ChangePassword(cfg *types.TmUserPasswordChange) error {
	c := crudConfig{
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmUsersPassword,
		endpointParams:   []string{o.User.ID},
		entityLabel:      labelDefinedEntityAccessControl,
		additionalHeader: getTenantContextHeader(o.TenantContext),
	}

	_, err := createInnerEntity(&o.vcdClient.Client, c, cfg)

	if err != nil {
		return fmt.Errorf("error updating %s password: %s", labelTmUser, err)
	}

	return nil
}
