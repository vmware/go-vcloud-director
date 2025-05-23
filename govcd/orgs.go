// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/url"
)

type OpenApiOrg struct {
	Org       *types.OpenApiOrg
	vcdClient *VCDClient
}

const LabelOrgs = "Organization"

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (org OpenApiOrg) wrap(inner *types.OpenApiOrg) *OpenApiOrg {
	org.Org = inner
	return &org
}

// GetAllOrgs retrieve all organizations visible to the user
// When 'multiSite' is set, it will also check the organizations available from associated sites
func (vcdClient *VCDClient) GetAllOrgs(queryParameters url.Values, multiSite bool) ([]*OpenApiOrg, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgs,
		entityLabel:     LabelOrgs,
		queryParameters: queryParameters,
	}
	if multiSite {
		c.additionalHeader = map[string]string{"Accept": "{{MEDIA_TYPE}};version={{API_VERSION}};multisite=global"}
	}

	outerType := OpenApiOrg{vcdClient: vcdClient}
	return getAllOuterEntities[OpenApiOrg, types.OpenApiOrg](&vcdClient.Client, outerType, c)
}
