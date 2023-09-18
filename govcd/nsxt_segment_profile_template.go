/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type NsxtSegmentProfileTemplate struct {
	NsxtSegmentProfileTemplate *types.NsxtSegmentProfileTemplate
	VCDClient                  *VCDClient
}

func (nsxtManager *NsxtManager) CreateSegmentProfileTemplate(segmentProfileConfig *types.NsxtSegmentProfileTemplate) (*NsxtSegmentProfileTemplate, error) {
	client := nsxtManager.VCDClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnSegmentProfile := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: &types.NsxtSegmentProfileTemplate{},
		VCDClient:                  nsxtManager.VCDClient,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, segmentProfileConfig, returnSegmentProfile.NsxtSegmentProfileTemplate, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Segment Profile Template: %s", err)
	}

	return returnSegmentProfile, nil
}

func (nsxtManager *NsxtManager) GetSegmentProfileTemplateById(id string) (*NsxtSegmentProfileTemplate, error) {
	client := nsxtManager.VCDClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty NSX-T Segment Profile Template ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	returnSegmentProfile := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: &types.NsxtSegmentProfileTemplate{},
		VCDClient:                  nsxtManager.VCDClient,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnSegmentProfile.NsxtSegmentProfileTemplate, nil)
	if err != nil {
		return nil, err
	}

	return returnSegmentProfile, nil
}

func (egw *NsxtSegmentProfileTemplate) Update(nsxtSegmentProfileTemplateConfig *types.NsxtSegmentProfileTemplate) (*NsxtSegmentProfileTemplate, error) {
	if !egw.VCDClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("only System Administrator can update NSX-T Segment Profile Template")
	}

	if nsxtSegmentProfileTemplateConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T Segment Profile Template without ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	apiVersion, err := egw.VCDClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := egw.VCDClient.Client.OpenApiBuildEndpoint(endpoint, nsxtSegmentProfileTemplateConfig.ID)
	if err != nil {
		return nil, err
	}

	returnEgw := &NsxtSegmentProfileTemplate{
		NsxtSegmentProfileTemplate: &types.NsxtSegmentProfileTemplate{},
		VCDClient:                  egw.VCDClient,
	}

	err = egw.VCDClient.Client.OpenApiPutItem(apiVersion, urlRef, nil, nsxtSegmentProfileTemplateConfig, returnEgw.NsxtSegmentProfileTemplate, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T Segment Profile Template: %s", err)
	}

	return returnEgw, nil
}

// Delete allows deleting NSX-TNSX-T Segment Profile Template for sysadmins
func (egw *NsxtSegmentProfileTemplate) Delete() error {
	if !egw.VCDClient.Client.IsSysAdmin {
		return fmt.Errorf("only Provider can delete NSX-T Segment Profile")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates
	endpointUrl := endpoint + egw.NsxtSegmentProfileTemplate.ID

	return deleteById(&egw.VCDClient.Client, endpoint, endpointUrl, "NSX-T Segment Profile Template")
}

func deleteById(client *Client, endpoint string, deletionUrl string, entityName string) error {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(deletionUrl)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting %s: %s", entityName, err)
	}

	return nil
}

func (vcdClient *VCDClient) GetGlobalDefaultSegmentProfileTemplates() (*types.NsxtSegmentProfileTemplateDefaultDefinition, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSpoofGuardProfiles
	return genericGetSingleEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, nil)
}

func (vcdClient *VCDClient) UpdateGlobalDefaultSegmentProfileTemplates(entityConfig *types.NsxtSegmentProfileTemplateDefaultDefinition) (*types.NsxtSegmentProfileTemplateDefaultDefinition, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplatesDefault
	return genericUpdateEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, entityConfig)
}
