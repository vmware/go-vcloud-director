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
	spt, err := genericCreateBareEntity("Segment Profile Template", &vcdClient.Client, endpoint, nil, segmentProfileConfig, nil, nil)
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
	allSegmentProfileTemplates, err := genericGetAllBareFilteredEntities[types.NsxtSegmentProfileTemplate]("Segment Profile Template", &client, endpoint, nil, queryFilter, nil)
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
	spt, err := genericGetSingleBareEntity[types.NsxtSegmentProfileTemplate]("Segment Profile Template", &vcdClient.Client, endpoint, []string{id}, nil, nil)
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
	updatedSpt, err := genericUpdateBareEntity("Segment Profile Template", &spt.VCDClient.Client, endpoint, []string{nsxtSegmentProfileTemplateConfig.ID}, nsxtSegmentProfileTemplateConfig, nil, nil)
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
	return deleteById("Segment Profile Template", &spt.VCDClient.Client, endpoint, []string{spt.NsxtSegmentProfileTemplate.ID}, nil, nil)
}
