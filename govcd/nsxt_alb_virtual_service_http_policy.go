/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelNsxtAlbVsHttpRequestRules = "NSX-T ALB Virtual Service HTTP Request Rules"
const labelNsxtAlbVsHttpResponseRules = "NSX-T ALB Virtual Service HTTP Response Rules"

func (nsxtAlbVirtualService *NsxtAlbVirtualService) GetAllHttpRequestRules(queryParameters url.Values) ([]*types.EdgeVirtualServiceHttpRequestRule, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpRequestRules,
		entityLabel:     labelNsxtAlbVsHttpRequestRules,
		endpointParams:  []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
		queryParameters: queryParameters,
	}

	return getAllInnerEntities[types.EdgeVirtualServiceHttpRequestRule](&nsxtAlbVirtualService.vcdClient.Client, c)
}

func (nsxtAlbVirtualService *NsxtAlbVirtualService) UpdateHttpRequestRules(config *types.EdgeVirtualServiceHttpRequestRules) (*types.EdgeVirtualServiceHttpRequestRules, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpRequestRules,
		entityLabel:    labelNsxtAlbVsHttpRequestRules,
		endpointParams: []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
	}

	return updateInnerEntity(&nsxtAlbVirtualService.vcdClient.Client, c, config)
}

func (nsxtAlbVirtualService *NsxtAlbVirtualService) GetAllHttpResponseRules(queryParameters url.Values) ([]*types.EdgeVirtualServiceHttpResponseRule, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpResponseRules,
		entityLabel:     labelNsxtAlbVsHttpResponseRules,
		endpointParams:  []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
		queryParameters: queryParameters,
	}

	return getAllInnerEntities[types.EdgeVirtualServiceHttpResponseRule](&nsxtAlbVirtualService.vcdClient.Client, c)
}

func (nsxtAlbVirtualService *NsxtAlbVirtualService) UpdateHttpResponseRules(config *types.EdgeVirtualServiceHttpResponseRules) (*types.EdgeVirtualServiceHttpResponseRules, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpResponseRules,
		entityLabel:    labelNsxtAlbVsHttpResponseRules,
		endpointParams: []string{nsxtAlbVirtualService.NsxtAlbVirtualService.ID},
	}

	return updateInnerEntity(&nsxtAlbVirtualService.vcdClient.Client, c, config)
}
