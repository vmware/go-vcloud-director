/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtSegmentProfileTemplate contains a structure for configuring Segment Profile Templates
type NsxtSegmentProfileTemplate struct {
	NsxtSegmentProfileTemplate *types.NsxtSegmentProfileTemplate
	VCDClient                  *VCDClient
}

// CreateSegmentProfileTemplate creates a Segment Profile Template that can later be assigned to
// global VCD configuration, Org VDC or Org VDC Network
func (vcdClient *VCDClient) CreateSegmentProfileTemplate(segmentProfileConfig *types.NsxtSegmentProfileTemplate) (*NsxtSegmentProfileTemplate, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnSegmentProfile := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: &types.NsxtSegmentProfileTemplate{},
		VCDClient:                  vcdClient,
	}

	err = vcdClient.Client.OpenApiPostItem(apiVersion, urlRef, nil, segmentProfileConfig, returnSegmentProfile.NsxtSegmentProfileTemplate, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Segment Profile Template: %s", err)
	}

	return returnSegmentProfile, nil
}

// GetAllSegmentProfileTemplates retrieves all Segment Profile Templates
func (vcdClient *VCDClient) GetAllSegmentProfileTemplates(queryFilter url.Values) ([]*NsxtSegmentProfileTemplate, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates

	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtSegmentProfileTemplate{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryFilter, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*NsxtSegmentProfileTemplate, len(typeResponses))
	for sliceIndex, singleSegmentProfileTemplate := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtSegmentProfileTemplate{
			NsxtSegmentProfileTemplate: singleSegmentProfileTemplate,
			VCDClient:                  vcdClient,
		}
	}

	return wrappedResponses, nil
}

// GetSegmentProfileTemplateById retrieves Segment Profile Template by ID
func (vcdClient *VCDClient) GetSegmentProfileTemplateById(id string) (*NsxtSegmentProfileTemplate, error) {
	if id == "" {
		return nil, fmt.Errorf("empty NSX-T Segment Profile Template ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	wrappedSegmentProfile := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: &types.NsxtSegmentProfileTemplate{},
		VCDClient:                  vcdClient,
	}

	err = vcdClient.Client.OpenApiGetItem(apiVersion, urlRef, nil, wrappedSegmentProfile.NsxtSegmentProfileTemplate, nil)
	if err != nil {
		return nil, err
	}

	return wrappedSegmentProfile, nil
}

// GetSegmentProfileTemplateByName retrieves Segment Profile Template by ID
func (vcdClient *VCDClient) GetSegmentProfileTemplateByName(name string) (*NsxtSegmentProfileTemplate, error) {
	filterByName := copyOrNewUrlValues(nil)
	filterByName = queryParameterFilterAnd(fmt.Sprintf("name==%s", name), filterByName)

	allSegmentProfileTemplates, err := vcdClient.GetAllSegmentProfileTemplates(filterByName)
	if err != nil {
		return nil, err
	}

	singleSegmentProfileTemplate, err := oneOrError("name", name, allSegmentProfileTemplates)
	if err != nil {
		return nil, err
	}

	return singleSegmentProfileTemplate, nil
}

// Update Segment Profile Template
func (spt *NsxtSegmentProfileTemplate) Update(nsxtSegmentProfileTemplateConfig *types.NsxtSegmentProfileTemplate) (*NsxtSegmentProfileTemplate, error) {
	if nsxtSegmentProfileTemplateConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T Segment Profile Template without ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	apiVersion, err := spt.VCDClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := spt.VCDClient.Client.OpenApiBuildEndpoint(endpoint, nsxtSegmentProfileTemplateConfig.ID)
	if err != nil {
		return nil, err
	}

	returnSpt := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: &types.NsxtSegmentProfileTemplate{},
		VCDClient:                  spt.VCDClient,
	}

	err = spt.VCDClient.Client.OpenApiPutItem(apiVersion, urlRef, nil, nsxtSegmentProfileTemplateConfig, returnSpt.NsxtSegmentProfileTemplate, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating Edge Gateway: %s", err)
	}

	return returnSpt, nil
}

// Delete allows deleting NSX-T Segment Profile Template
func (spt *NsxtSegmentProfileTemplate) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	apiVersion, err := spt.VCDClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if spt.NsxtSegmentProfileTemplate.ID == "" {
		return fmt.Errorf("cannot delete Segment Profile Template without ID")
	}

	urlRef, err := spt.VCDClient.Client.OpenApiBuildEndpoint(endpoint, spt.NsxtSegmentProfileTemplate.ID)
	if err != nil {
		return err
	}

	err = spt.VCDClient.Client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting Segment Profile Template: %s", err)
	}

	return nil
}
