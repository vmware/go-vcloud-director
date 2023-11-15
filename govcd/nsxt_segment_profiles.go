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
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtSegmentProfileIpDiscovery{{}}
	err = vcdClient.Client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetIpDiscoveryProfileByName retrieves an NSX-T Segment IP Discovery Profile by given name
func (vcdClient *VCDClient) GetIpDiscoveryProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileIpDiscovery, error) {
	apiFilteredEntities, err := vcdClient.GetAllIpDiscoveryProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	filteredByName := make([]*types.NsxtSegmentProfileIpDiscovery, 0)
	for _, v := range apiFilteredEntities {
		if v.DisplayName == name {
			filteredByName = append(filteredByName, v)
		}
	}

	return oneOrError("displayName", name, filteredByName)
}

// GetAllMacDiscoveryProfiles retrieves all MAC Discovery Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllMacDiscoveryProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileMacDiscovery, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentMacDiscoveryProfiles
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtSegmentProfileMacDiscovery{{}}
	err = vcdClient.Client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetMacDiscoveryProfileByName retrieves an NSX-T Segment MAC Discovery Profile by given name
func (vcdClient *VCDClient) GetMacDiscoveryProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileMacDiscovery, error) {
	apiFilteredEntities, err := vcdClient.GetAllMacDiscoveryProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	filteredByName := make([]*types.NsxtSegmentProfileMacDiscovery, 0)
	for _, v := range apiFilteredEntities {
		if v.DisplayName == name {
			filteredByName = append(filteredByName, v)
		}
	}

	return oneOrError("displayName", name, filteredByName)
}

// GetAllSpoofGuardProfiles retrieves all Spoof Guard Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllSpoofGuardProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentSpoofGuard, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSpoofGuardProfiles
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtSegmentProfileSegmentSpoofGuard{{}}
	err = vcdClient.Client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetSpoofGuardProfileByName retrieves an NSX-T Segment Spoof Guard Profile by given name
func (vcdClient *VCDClient) GetSpoofGuardProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentSpoofGuard, error) {
	apiFilteredEntities, err := vcdClient.GetAllSpoofGuardProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	filteredByName := make([]*types.NsxtSegmentProfileSegmentSpoofGuard, 0)
	for _, v := range apiFilteredEntities {
		if v.DisplayName == name {
			filteredByName = append(filteredByName, v)
		}
	}

	return oneOrError("displayName", name, filteredByName)
}

// GetAllQoSProfiles retrieves all QoS Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllQoSProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentQosProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentQosProfiles
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtSegmentProfileSegmentQosProfile{{}}
	err = vcdClient.Client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetQoSProfileByName retrieves an NSX-T Segment QoS Profile by given name
func (vcdClient *VCDClient) GetQoSProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentQosProfile, error) {
	apiFilteredEntities, err := vcdClient.GetAllQoSProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	filteredByName := make([]*types.NsxtSegmentProfileSegmentQosProfile, 0)
	for _, v := range apiFilteredEntities {
		if v.DisplayName == name {
			filteredByName = append(filteredByName, v)
		}
	}

	return oneOrError("displayName", name, filteredByName)
}

// GetAllSegmentSecurityProfiles retrieves all Segment Security Profiles configured in an NSX-T manager.
// NSX-T manager ID (nsxTManagerRef.id), Org VDC ID (orgVdcId) or VDC Group ID (vdcGroupId) must be
// supplied as a filter. Results can also be filtered by a single profile ID
// (filter=nsxTManagerRef.id==nsxTManagerUrn;id==profileId).
func (vcdClient *VCDClient) GetAllSegmentSecurityProfiles(queryParameters url.Values) ([]*types.NsxtSegmentProfileSegmentSecurity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSecurityProfiles
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtSegmentProfileSegmentSecurity{{}}
	err = vcdClient.Client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetSegmentSecurityProfileByName retrieves an NSX-T Segment Security Profile by given name
func (vcdClient *VCDClient) GetSegmentSecurityProfileByName(name string, queryParameters url.Values) (*types.NsxtSegmentProfileSegmentSecurity, error) {
	apiFilteredEntities, err := vcdClient.GetAllSegmentSecurityProfiles(queryParameters) // API filtering by 'displayName' field is not supported
	if err != nil {
		return nil, err
	}

	filteredByName := make([]*types.NsxtSegmentProfileSegmentSecurity, 0)
	for _, v := range apiFilteredEntities {
		if v.DisplayName == name {
			filteredByName = append(filteredByName, v)
		}
	}

	return oneOrError("displayName", name, filteredByName)
}
