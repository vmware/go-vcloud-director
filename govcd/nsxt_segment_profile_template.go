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
	c := genericCrudConfig{
		endpoint:   types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates,
		entityName: "Segment Profile Template",
	}
	spt, err := genericCreateBareEntity(&vcdClient.Client, segmentProfileConfig, c)
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
	c := genericCrudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates,
		entityName:      "Segment Profile Template",
		queryParameters: queryFilter,
	}
	allSegmentProfileTemplates, err := genericGetAllBareFilteredEntities[types.NsxtSegmentProfileTemplate](&vcdClient.Client, c)
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

	c := genericCrudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates,
		endpointParams: []string{id},
		entityName:     "Segment Profile Template",
	}
	spt, err := genericGetSingleBareEntity[types.NsxtSegmentProfileTemplate](&vcdClient.Client, c)
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

	c := genericCrudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates,
		endpointParams: []string{nsxtSegmentProfileTemplateConfig.ID},
		entityName:     "Segment Profile Template",
	}
	updatedSpt, err := genericUpdateBareEntity(&spt.VCDClient.Client, nsxtSegmentProfileTemplateConfig, c)
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
	c := genericCrudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates,
		endpointParams: []string{spt.NsxtSegmentProfileTemplate.ID},
		entityName:     "Segment Profile Template",
	}
	return deleteById(&spt.VCDClient.Client, c)
}
