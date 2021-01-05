/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetAllOpenApiRights retrieves all rights using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
func (adminOrg *AdminOrg) GetAllOpenApiRights(queryParameters url.Values) ([]*types.Right, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRights
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	typeResponses := []*types.Right{{}}
	err = adminOrg.client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetRoleRights retrieves all rights belonging to a given Role. Query parameters can be supplied to perform additional
// filtering
func (role *Role) GetRoleRights(queryParameters url.Values) ([]*types.Right, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles + role.Role.ID + "/" + types.OpenApiEndpointRights
	minimumApiVersion, err := role.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := role.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.Right{{}}
	err = role.client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses, getTenantContextHeader(role.TenantContext))
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}
