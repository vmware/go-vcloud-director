/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetAllIpDiscoveryProfiles retrieves all IP Discovery Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllIpDiscoveryProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileIpDiscovery, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentIpDiscoveryProfiles
	return genericGetAllBareFilteredEntities[types.NsxtSegmentProfileIpDiscovery](&vcdClient.Client, endpoint, endpoint, queryParameters, "IP Discovery Profiles")
}

func (vcdClient *VCDClient) GetIpDiscoveryProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileIpDiscovery, error) {
	apiFilteredEntities, err := vcdClient.GetAllIpDiscoveryProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name, "Segment IP Discovery Profile")
}

// GetAllMacDiscoveryProfiles retrieves all MAC Discovery Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllMacDiscoveryProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileMacDiscovery, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentMacDiscoveryProfiles
	return genericGetAllBareFilteredEntities[types.NsxtSegmentProfileMacDiscovery](&vcdClient.Client, endpoint, endpoint, queryParameters, "MAC Discovery Profiles")
}

func (vcdClient *VCDClient) GetMacDiscoveryProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileMacDiscovery, error) {
	apiFilteredEntities, err := vcdClient.GetAllMacDiscoveryProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name, "Segment MAC Discovery Profile")
}

// GetAllSpoofGuardProfiles retrieves all Spoof Guard Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllSpoofGuardProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentSpoofGuard, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSpoofGuardProfiles
	return genericGetAllBareFilteredEntities[types.NsxtSegmentProfileSegmentSpoofGuard](&vcdClient.Client, endpoint, endpoint, queryParameters, "Spoof Guard Profiles")
}

func (vcdClient *VCDClient) GetSpoofGuardProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentSpoofGuard, error) {
	apiFilteredEntities, err := vcdClient.GetAllSpoofGuardProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name, "Segment Spoof Guard Profile")
}

// GetAllQoSProfiles retrieves all QoS Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllQoSProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentQosProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentQosProfiles
	return genericGetAllBareFilteredEntities[types.NsxtSegmentProfileSegmentQosProfile](&vcdClient.Client, endpoint, endpoint, queryParameters, "QoS Profiles")
}

func (vcdClient *VCDClient) GetQoSProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentQosProfile, error) {
	apiFilteredEntities, err := vcdClient.GetAllQoSProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name, "Segment QoS Profile")
}

// GetAllSegmentSecurityProfiles retrieves all Segment Security Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllSegmentSecurityProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentSecurity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSecurityProfiles
	return genericGetAllBareFilteredEntities[types.NsxtSegmentProfileSegmentSecurity](&vcdClient.Client, endpoint, endpoint, queryParameters, "Segment Security Profiles")
}

func (vcdClient *VCDClient) GetSegmentSecurityProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentSecurity, error) {
	apiFilteredEntities, err := vcdClient.GetAllSegmentSecurityProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return genericLocalFilterOneOrError(apiFilteredEntities, "DisplayName", name, "Segment Security Profile")
}
