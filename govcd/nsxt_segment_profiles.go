/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const (
	labelIpDiscoveryProfiles     = "IP Discovery Profiles"
	labelMacDiscoveryProfiles    = "MAC Discovery Profiles"
	labelSpoofGuardProfiles      = "Spoof Guard Profiles"
	labelQosProfiles             = "QoS Profiles"
	labelSegmentSecurityProfiles = "Segment Security Profiles"
)

// GetAllIpDiscoveryProfiles retrieves all IP Discovery Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllIpDiscoveryProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileIpDiscovery, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentIpDiscoveryProfiles,
		queryParameters: queryParameters,
		entityLabel:     labelIpDiscoveryProfiles,
	}
	return getAllInnerEntities[types.NsxtSegmentProfileIpDiscovery](&vcdClient.Client, c)
}

func (vcdClient *VCDClient) GetIpDiscoveryProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileIpDiscovery, error) {
	apiFilteredEntities, err := vcdClient.GetAllIpDiscoveryProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return localFilterOneOrError(labelIpDiscoveryProfiles, apiFilteredEntities, "DisplayName", name)
}

// GetAllMacDiscoveryProfiles retrieves all MAC Discovery Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllMacDiscoveryProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileMacDiscovery, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentMacDiscoveryProfiles,
		queryParameters: queryParameters,
		entityLabel:     labelMacDiscoveryProfiles,
	}
	return getAllInnerEntities[types.NsxtSegmentProfileMacDiscovery](&vcdClient.Client, c)
}

func (vcdClient *VCDClient) GetMacDiscoveryProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileMacDiscovery, error) {
	apiFilteredEntities, err := vcdClient.GetAllMacDiscoveryProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return localFilterOneOrError(labelMacDiscoveryProfiles, apiFilteredEntities, "DisplayName", name)
}

// GetAllSpoofGuardProfiles retrieves all Spoof Guard Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllSpoofGuardProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentSpoofGuard, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSpoofGuardProfiles,
		queryParameters: queryParameters,
		entityLabel:     labelSpoofGuardProfiles,
	}
	return getAllInnerEntities[types.NsxtSegmentProfileSegmentSpoofGuard](&vcdClient.Client, c)
}

func (vcdClient *VCDClient) GetSpoofGuardProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentSpoofGuard, error) {
	apiFilteredEntities, err := vcdClient.GetAllSpoofGuardProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return localFilterOneOrError(labelSpoofGuardProfiles, apiFilteredEntities, "DisplayName", name)
}

// GetAllQoSProfiles retrieves all QoS Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllQoSProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentQosProfile, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentQosProfiles,
		queryParameters: queryParameters,
		entityLabel:     labelQosProfiles,
	}
	return getAllInnerEntities[types.NsxtSegmentProfileSegmentQosProfile](&vcdClient.Client, c)
}

func (vcdClient *VCDClient) GetQoSProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentQosProfile, error) {
	apiFilteredEntities, err := vcdClient.GetAllQoSProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return localFilterOneOrError(labelQosProfiles, apiFilteredEntities, "DisplayName", name)
}

// GetAllSegmentSecurityProfiles retrieves all Segment Security Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllSegmentSecurityProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentSecurity, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSecurityProfiles,
		queryParameters: queryParameters,
		entityLabel:     labelSegmentSecurityProfiles,
	}
	return getAllInnerEntities[types.NsxtSegmentProfileSegmentSecurity](&vcdClient.Client, c)
}

func (vcdClient *VCDClient) GetSegmentSecurityProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentSecurity, error) {
	apiFilteredEntities, err := vcdClient.GetAllSegmentSecurityProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	return localFilterOneOrError(labelSegmentSecurityProfiles, apiFilteredEntities, "DisplayName", name)
}
