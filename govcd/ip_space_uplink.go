/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelIpSpaceUplink = "IP Space Uplink"

// IpSpaceUplink provides the capability to assign one or more IP Spaces as Uplinks to External
// Networks
type IpSpaceUplink struct {
	IpSpaceUplink *types.IpSpaceUplink
	vcdClient     *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (i IpSpaceUplink) wrap(inner *types.IpSpaceUplink) *IpSpaceUplink {
	i.IpSpaceUplink = inner
	return &i
}

// CreateIpSpaceUplink creates an IP Space Uplink with a given configuration
func (vcdClient *VCDClient) CreateIpSpaceUplink(ipSpaceUplinkConfig *types.IpSpaceUplink) (*IpSpaceUplink, error) {
	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks,
		entityLabel: labelIpSpaceUplink,
	}

	outerType := IpSpaceUplink{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, ipSpaceUplinkConfig)
}

// GetAllIpSpaceUplinks retrieves all IP Space Uplinks for a given External Network ID
//
// externalNetworkId is mandatory
func (vcdClient *VCDClient) GetAllIpSpaceUplinks(externalNetworkId string, queryParameters url.Values) ([]*IpSpaceUplink, error) {
	if externalNetworkId == "" {
		return nil, fmt.Errorf("mandatory External Network ID is empty")
	}

	queryparams := queryParameterFilterAnd(fmt.Sprintf("externalNetworkRef.id==%s", externalNetworkId), queryParameters)
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks,
		entityLabel:     labelIpSpaceUplink,
		queryParameters: queryparams,
	}

	outerType := IpSpaceUplink{vcdClient: vcdClient}
	return getAllOuterEntities[IpSpaceUplink, types.IpSpaceUplink](&vcdClient.Client, outerType, c)
}

// GetIpSpaceUplinkByName retrieves a single IP Space Uplink by Name in a given External Network
func (vcdClient *VCDClient) GetIpSpaceUplinkByName(externalNetworkId, name string) (*IpSpaceUplink, error) {
	queryParams := queryParameterFilterAnd(fmt.Sprintf("name==%s", name), nil)
	allIpSpaceUplinks, err := vcdClient.GetAllIpSpaceUplinks(externalNetworkId, queryParams)
	if err != nil {
		return nil, fmt.Errorf("error getting IP Space Uplink by Name '%s':%s", name, err)
	}

	return oneOrError("name", name, allIpSpaceUplinks)
}

// GetIpSpaceUplinkById retrieves IP Space Uplink with a given ID
func (vcdClient *VCDClient) GetIpSpaceUplinkById(id string) (*IpSpaceUplink, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks,
		endpointParams: []string{id},
		entityLabel:    labelIpSpaceUplink,
	}

	outerType := IpSpaceUplink{vcdClient: vcdClient}
	return getOuterEntity[IpSpaceUplink, types.IpSpaceUplink](&vcdClient.Client, outerType, c)
}

// Update IP Space Uplink
func (ipSpaceUplink *IpSpaceUplink) Update(ipSpaceUplinkConfig *types.IpSpaceUplink) (*IpSpaceUplink, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks,
		endpointParams: []string{ipSpaceUplink.IpSpaceUplink.ID},
		entityLabel:    labelIpSpaceUplink,
	}

	outerType := IpSpaceUplink{vcdClient: ipSpaceUplink.vcdClient}
	return updateOuterEntity(&ipSpaceUplink.vcdClient.Client, outerType, c, ipSpaceUplinkConfig)
}

// Delete IP Space Uplink
func (ipSpaceUplink *IpSpaceUplink) Delete() error {
	if ipSpaceUplink == nil || ipSpaceUplink.IpSpaceUplink == nil || ipSpaceUplink.IpSpaceUplink.ID == "" {
		return fmt.Errorf("IP Space Uplink must have ID")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks,
		endpointParams: []string{ipSpaceUplink.IpSpaceUplink.ID},
		entityLabel:    labelIpSpaceUplink,
	}

	return deleteEntityById(&ipSpaceUplink.vcdClient.Client, c)
}
