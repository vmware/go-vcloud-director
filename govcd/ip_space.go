/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelIpSpace = "IP Space"

// IpSpace provides structured approach to allocating public and private IP addresses by preventing
// the use of overlapping IP addresses across organizations and organization VDCs.
//
// An IP space consists of a set of defined non-overlapping IP ranges and small CIDR blocks that are
// reserved and used during the consumption aspect of the IP space life cycle. An IP space can be
// either IPv4 or IPv6, but not both.
//
// Every IP space has an internal scope and an external scope. The internal scope of an IP space is
// a list of CIDR notations that defines the exact span of IP addresses in which all ranges and
// blocks must be contained in. The external scope defines the total span of IP addresses to which
// the IP space has access, for example the internet or a WAN.
type IpSpace struct {
	IpSpace   *types.IpSpace
	vcdClient *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g IpSpace) wrap(inner *types.IpSpace) *IpSpace {
	g.IpSpace = inner
	return &g
}

// CreateIpSpace creates IP Space with desired configuration
func (vcdClient *VCDClient) CreateIpSpace(ipSpaceConfig *types.IpSpace) (*IpSpace, error) {
	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces,
		entityLabel: labelIpSpace,
	}
	outerType := IpSpace{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, ipSpaceConfig)
}

// GetAllIpSpaceSummaries retrieve summaries of all IP Spaces with an optional filter
// Note. There is no API endpoint to get multiple IP Spaces with their full definitions. Only
// "summaries" endpoint exists, but it does not include all fields. To retrieve complete structure
// one can use `GetIpSpaceById` or `GetIpSpaceByName`
func (vcdClient *VCDClient) GetAllIpSpaceSummaries(queryParameters url.Values) ([]*IpSpace, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceSummaries,
		entityLabel:     labelIpSpace,
		queryParameters: queryParameters,
	}

	outerType := IpSpace{vcdClient: vcdClient}
	return getAllOuterEntities[IpSpace, types.IpSpace](&vcdClient.Client, outerType, c)
}

// GetIpSpaceByName retrieves IP Space with a given name
// Note. It will return an error if multiple IP Spaces exist with the same name
func (vcdClient *VCDClient) GetIpSpaceByName(name string) (*IpSpace, error) {
	if name == "" {
		return nil, fmt.Errorf("IP Space lookup requires name")
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllIpSpaceSummaries(queryParams)
	if err != nil {
		return nil, err
	}

	singleIpSpace, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetIpSpaceById(singleIpSpace.IpSpace.ID)
}

func (vcdClient *VCDClient) GetIpSpaceById(id string) (*IpSpace, error) {
	c := crudConfig{
		entityLabel:    labelIpSpace,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces,
		endpointParams: []string{id},
	}

	outerType := IpSpace{vcdClient: vcdClient}
	return getOuterEntity[IpSpace, types.IpSpace](&vcdClient.Client, outerType, c)
}

// GetIpSpaceByNameAndOrgId retrieves IP Space with a given name in a particular Org
// Note. Only PRIVATE IP spaces belong to Orgs
func (vcdClient *VCDClient) GetIpSpaceByNameAndOrgId(name, orgId string) (*IpSpace, error) {
	if name == "" || orgId == "" {
		return nil, fmt.Errorf("IP Space lookup requires name and Org ID")
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("orgRef.id=="+orgId, queryParams)

	filteredEntities, err := vcdClient.GetAllIpSpaceSummaries(queryParams)
	if err != nil {
		return nil, err
	}

	singleIpSpace, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetIpSpaceById(singleIpSpace.IpSpace.ID)
}

// Update updates IP Space with new config
func (ipSpace *IpSpace) Update(ipSpaceConfig *types.IpSpace) (*IpSpace, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces,
		endpointParams: []string{ipSpace.IpSpace.ID},
		entityLabel:    labelIpSpace,
	}
	outerType := IpSpace{vcdClient: ipSpace.vcdClient}
	return updateOuterEntity(&ipSpace.vcdClient.Client, outerType, c, ipSpaceConfig)
}

// Delete deletes IP Space
func (ipSpace *IpSpace) Delete() error {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces,
		endpointParams: []string{ipSpace.IpSpace.ID},
		entityLabel:    labelIpSpace,
	}
	return deleteEntityById(&ipSpace.vcdClient.Client, c)
}
