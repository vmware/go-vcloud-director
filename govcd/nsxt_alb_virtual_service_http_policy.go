/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelNsxtAlbVsHttpRequestRules = "ALB Virtual Service HTTP Request Rules"
const labelNsxtAlbVsHttpResponseRules = "ALB Virtual Service HTTP Response Rules"
const labelNsxtAlbVsHttpSecurityRules = "ALB Virtual Service HTTP Security Rules"

// GetAllHttpRequestRules returns all ALB Virtual Service HTTP Request Rules
func (nsxtAlbVirtualService *NsxtAlbVirtualService) GetAllHttpRequestRules(queryParameters url.Values) ([]*types.AlbVsHttpRequestRule, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpRequestRules,
		entityLabel:     labelNsxtAlbVsHttpRequestRules,
		endpointParams:  []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
		queryParameters: queryParameters,
	}

	return getAllInnerEntities[types.AlbVsHttpRequestRule](&nsxtAlbVirtualService.vcdClient.Client, c)
}

// UpdateHttpRequestRules sets ALB Virtual Service HTTP Request Rules
func (nsxtAlbVirtualService *NsxtAlbVirtualService) UpdateHttpRequestRules(config *types.AlbVsHttpRequestRules) (*types.AlbVsHttpRequestRules, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpRequestRules,
		entityLabel:    labelNsxtAlbVsHttpRequestRules,
		endpointParams: []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
	}

	return updateInnerEntity(&nsxtAlbVirtualService.vcdClient.Client, c, config)
}

// GetAllHttpRequestRules returns all ALB Virtual Service HTTP Response Rules
func (nsxtAlbVirtualService *NsxtAlbVirtualService) GetAllHttpResponseRules(queryParameters url.Values) ([]*types.AlbVsHttpResponseRule, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpResponseRules,
		entityLabel:     labelNsxtAlbVsHttpResponseRules,
		endpointParams:  []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
		queryParameters: queryParameters,
	}

	return getAllInnerEntities[types.AlbVsHttpResponseRule](&nsxtAlbVirtualService.vcdClient.Client, c)
}

// UpdateHttpRequestRules sets ALB Virtual Service HTTP Response Rules
func (nsxtAlbVirtualService *NsxtAlbVirtualService) UpdateHttpResponseRules(config *types.AlbVsHttpResponseRules) (*types.AlbVsHttpResponseRules, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpResponseRules,
		entityLabel:    labelNsxtAlbVsHttpResponseRules,
		endpointParams: []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
	}

	return updateInnerEntity(&nsxtAlbVirtualService.vcdClient.Client, c, config)
}

// GetAllHttpRequestRules returns all ALB Virtual Service HTTP Security Rules
func (nsxtAlbVirtualService *NsxtAlbVirtualService) GetAllHttpSecurityRules(queryParameters url.Values) ([]*types.AlbVsHttpSecurityRule, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpSecurityRules,
		entityLabel:     labelNsxtAlbVsHttpSecurityRules,
		endpointParams:  []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
		queryParameters: queryParameters,
	}

	return getAllInnerEntities[types.AlbVsHttpSecurityRule](&nsxtAlbVirtualService.vcdClient.Client, c)
}

// UpdateHttpRequestRules sets ALB Virtual Service HTTP Security Rules
func (nsxtAlbVirtualService *NsxtAlbVirtualService) UpdateHttpSecurityRules(config *types.AlbVsHttpSecurityRules) (*types.AlbVsHttpSecurityRules, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpSecurityRules,
		entityLabel:    labelNsxtAlbVsHttpSecurityRules,
		endpointParams: []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
	}

	return updateInnerEntity(&nsxtAlbVirtualService.vcdClient.Client, c, config)
}
