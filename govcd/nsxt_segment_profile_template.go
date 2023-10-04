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
	spt, err := genericCreateBareEntity(&vcdClient.Client, endpoint, endpoint, segmentProfileConfig, "Segment Profile Template")
	if err != nil {
		return nil, err
	}

	returnSegmentProfile := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: spt,
		VCDClient:                  vcdClient,
	}

	return returnSegmentProfile, nil
}

// GetAllSegmentProfileTemplates retrieves all Segment Profile Templates
func (vcdClient *VCDClient) GetAllSegmentProfileTemplates(queryFilter url.Values) ([]*NsxtSegmentProfileTemplate, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	allSegmentProfileTemplates, err := genericGetAllBareFilteredEntities[types.NsxtSegmentProfileTemplate](&client, endpoint, endpoint, queryFilter, "Segment Profile Template")
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*NsxtSegmentProfileTemplate, len(allSegmentProfileTemplates))
	for sliceIndex, singleSegmentProfileTemplate := range allSegmentProfileTemplates {
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
	exactEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates + id
	spt, err := genericGetSingleBareEntity[types.NsxtSegmentProfileTemplate](&vcdClient.Client, endpoint, exactEndpoint, nil, "Segment Profile Template")
	if err != nil {
		return nil, err
	}

	wrappedSegmentProfile := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: spt,
		VCDClient:                  vcdClient,
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
	exactEndpoint := endpoint + nsxtSegmentProfileTemplateConfig.ID

	updatedSpt, err := genericUpdateBareEntity(&spt.VCDClient.Client, endpoint, exactEndpoint, nsxtSegmentProfileTemplateConfig, "Segment Profile Template")
	if err != nil {
		return nil, err
	}

	returnSpt := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: updatedSpt,
		VCDClient:                  spt.VCDClient,
	}

	return returnSpt, nil
}

// Delete allows deleting NSX-T Segment Profile Template
func (spt *NsxtSegmentProfileTemplate) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	exactEndpoint := endpoint + spt.NsxtSegmentProfileTemplate.ID

	return deleteById(&spt.VCDClient.Client, endpoint, exactEndpoint, "Segment Profile Template")
}
