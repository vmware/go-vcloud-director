/*
 * Copyright 2025 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelFeatureFlag = "Feature Flag"

// UpdateFeatureFlag can update a feature flag with a given config. Can mostly be used to enable
// specific feature flags.
// The ID must be set (e.g. 'urn:vcloud:featureflag:CLASSIC_TENANT_CREATION' for Classic Tenants)
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

// GetFeatureFlagById returns a feature flag by ID. Sample ID -
// 'urn:vcloud:featureflag:CLASSIC_TENANT_CREATION'
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

// GetAllFeatureFlags retrieves all available feature flags
func (vcdClient *VCDClient) GetAllFeatureFlags() ([]*types.FeatureFlag, error) {
	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFeatureFlags,
		entityLabel: labelFeatureFlag,
	}
	return getAllInnerEntities[types.FeatureFlag](&vcdClient.Client, c)
}
