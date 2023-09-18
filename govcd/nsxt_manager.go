/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type NsxtManager struct {
	NsxtManager *types.NsxtManager
	VCDClient   *VCDClient
	// Urn holds a URN value for NSX-T manager. None of the API endpoints return it, but filtering other entities requires that
	// Sample format: urn:vcloud:nsxtmanager:UUID
	//
	// Note:  this is being computed when retrieving the structure and will not be populated if this structure is initialized manually
	Urn string
}

// GetAllIpDiscoveryProfiles retrieves all IP Discovery Profiles configured on an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllIpDiscoveryProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileTemplateIpDiscovery, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentIpDiscoveryProfiles
	return genericGetAllFilteredEntities[types.NsxtSegmentProfileTemplateIpDiscovery](&vcdClient.Client, endpoint, queryParameters)
}

func (vcdClient *VCDClient) GetIpDiscoveryProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileTemplateIpDiscovery, error) {
	apiFilteredEntities, err := vcdClient.GetAllIpDiscoveryProfiles(queryParameters)
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name)
}

// GetAllMacDiscoveryProfiles retrieves all MAC Discovery Profiles configured on an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllMacDiscoveryProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileTemplateMacDiscovery, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentMacDiscoveryProfiles
	return genericGetAllFilteredEntities[types.NsxtSegmentProfileTemplateMacDiscovery](&vcdClient.Client, endpoint, queryParameters)
}

func (vcdClient *VCDClient) GetMacDiscoveryProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileTemplateMacDiscovery, error) {
	apiFilteredEntities, err := vcdClient.GetAllMacDiscoveryProfiles(queryParameters)
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name)
}

// GetAllSpoofGuardProfiles retrieves all Spoof Guard Profiles configured on an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllSpoofGuardProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileTemplateSegmentSpoofGuard, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSpoofGuardProfiles
	return genericGetAllFilteredEntities[types.NsxtSegmentProfileTemplateSegmentSpoofGuard](&vcdClient.Client, endpoint, queryParameters)
}

func (vcdClient *VCDClient) GetSpoofGuardProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileTemplateSegmentSpoofGuard, error) {
	apiFilteredEntities, err := vcdClient.GetAllSpoofGuardProfiles(queryParameters)
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name)
}

// GetAllQoSProfiles retrieves all QoS Profiles configured on an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllQoSProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileTemplateSegmentQosProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentQosProfiles
	return genericGetAllFilteredEntities[types.NsxtSegmentProfileTemplateSegmentQosProfile](&vcdClient.Client, endpoint, queryParameters)
}

func (vcdClient *VCDClient) GetQoSProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileTemplateSegmentQosProfile, error) {
	apiFilteredEntities, err := vcdClient.GetAllQoSProfiles(queryParameters)
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name)
}

// GetAllSegmentSecurityProfiles retrieves all Segment Security Profiles configured on an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllSegmentSecurityProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileTemplateSegmentSecurity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSecurityProfiles
	return genericGetAllFilteredEntities[types.NsxtSegmentProfileTemplateSegmentSecurity](&vcdClient.Client, endpoint, queryParameters)
}

func (vcdClient *VCDClient) GetSegmentSecurityProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileTemplateSegmentSecurity, error) {
	apiFilteredEntities, err := vcdClient.GetAllSegmentSecurityProfiles(queryParameters)
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name)
}
