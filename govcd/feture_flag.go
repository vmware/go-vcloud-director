/*
 * Copyright 2025 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// {
// 	"id": "urn:vcloud:featureflag:CLASSIC_TENANT_CREATION",
// 	"name": "CLASSIC_TENANT_CREATION",
// 	"usage": "ALPHA",
// 	"enabled": true,
// 	"displayName": "Classic Tenant Creation",
// 	"displayDescription": "This feature enables the creation of classic tenant Organizations."
//   }

const labelFeatureFlag = "Org VDC Network Segment Profile"

func (vcdClient *VCDClient) UpdateFeatureFlag(entityConfig *types.FeatureFlag) (*types.FeatureFlag, error) {
	if entityConfig.ID == "" {
		return nil, fmt.Errorf("id must be specified to update feature flag")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFeatureFlags,
		endpointParams: []string{entityConfig.ID},
		entityLabel:    labelFeatureFlag,
	}
	return updateInnerEntity(&vcdClient.Client, c, entityConfig)
}

func (vcdClient *VCDClient) GetFeatureFlagById(featureFlagId string) (*types.FeatureFlag, error) {
	if featureFlagId == "" {
		return nil, fmt.Errorf("ID must be specified to update feature flag")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFeatureFlags,
		endpointParams: []string{featureFlagId},
		entityLabel:    labelFeatureFlag,
	}
	return getInnerEntity[types.FeatureFlag](&vcdClient.Client, c)
}

func (vcdClient *VCDClient) GetAllFeatureFlags() ([]*types.FeatureFlag, error) {

	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFeatureFlags,
		entityLabel: labelFeatureFlag,
	}
	return getAllInnerEntities[types.FeatureFlag](&vcdClient.Client, c)
}
