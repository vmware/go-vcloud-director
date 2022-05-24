package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

// GetAllSecurityTaggedEntities Retrieves the list of entities that have at least one tag assigned to it.
// queryParameters allows users to pass filters: I.e: filter=(tag==Web;entityType==vm)
// This function works from API v36.0 (VCD 10.3.0+)
func (org *Org) GetAllSecurityTaggedEntities(queryParameters url.Values) ([]types.SecurityTaggedEntity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags
	apiVersion, err := org.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := org.client.OpenApiBuildEndpoint(endpoint, "/entities")
	if err != nil {
		return nil, err
	}

	tenantContext, err := org.getTenantContext()
	if err != nil {
		return nil, err
	}

	var securityTaggedEntities []types.SecurityTaggedEntity
	err = org.client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &securityTaggedEntities, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	return securityTaggedEntities, nil
}

// GetAllSecurityTaggedEntitiesByName wraps GetAllSecurityTaggedEntities and returns ErrorEntityNotFound if nothing was found
// This function works from API v36.0 (VCD 10.3.0+)
func (org *Org) GetAllSecurityTaggedEntitiesByName(securityTagName string) ([]types.SecurityTaggedEntity, error) {
	queryParameters := copyOrNewUrlValues(nil)
	queryParameters.Add("filter", "tag=="+securityTagName)

	securityTagEntities, err := org.GetAllSecurityTaggedEntities(queryParameters)
	if err != nil {
		return nil, err
	}

	if len(securityTagEntities) == 0 {
		return nil, ErrorEntityNotFound
	}

	return securityTagEntities, nil
}

// GetAllSecurityTagValues Retrieves the list of security tags that are in the organization and can be reused to tag an entity.
// The list of tags include tags assigned to entities within the organization.
// This function works from API v36.0 (VCD 10.3.0+)
func (org *Org) GetAllSecurityTagValues(queryParameters url.Values) ([]types.SecurityTagValue, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags
	apiVersion, err := org.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := org.client.OpenApiBuildEndpoint(endpoint, "/values")
	if err != nil {
		return nil, err
	}

	tenantContext, err := org.getTenantContext()
	if err != nil {
		return nil, err
	}

	var securityTaggedValues []types.SecurityTagValue
	err = org.client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &securityTaggedValues, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	return securityTaggedValues, nil
}

// GetVMSecurityTags Retrieves the list of tags for a specific VM. If user has view right to the VM, user can view its tags.
// This function works from API v36.0 (VCD 10.3.0+)
func (vm *VM) GetVMSecurityTags() (*types.EntitySecurityTags, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags
	apiVersion, err := vm.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vm.client.OpenApiBuildEndpoint(endpoint, fmt.Sprintf("/vm/%s", vm.VM.ID))
	if err != nil {
		return nil, err
	}

	var entitySecurityTags types.EntitySecurityTags
	err = vm.client.OpenApiGetItem(apiVersion, urlRef, nil, &entitySecurityTags, nil)
	if err != nil {
		return nil, err
	}

	return &entitySecurityTags, nil
}

// UpdateSecurityTag updates the entities associated with a Security Tag.
// Only the list of tagged entities can be updated. The name cannot be updated.
// Any other existing entities not in the list will be untagged.
// This function works from API v36.0 (VCD 10.3.0+)
func (org *Org) UpdateSecurityTag(securityTag *types.SecurityTag) (*types.SecurityTag, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags
	apiVersion, err := org.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := org.client.OpenApiBuildEndpoint(endpoint, "/tag")
	if err != nil {
		return nil, err
	}

	tenantContext, err := org.getTenantContext()
	if err != nil {
		return nil, err
	}

	err = org.client.OpenApiPutItem(apiVersion, urlRef, nil, securityTag, nil, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	queryParameters := copyOrNewUrlValues(nil)
	queryParameters.Add("filter", "tag=="+securityTag.Tag)
	readEntities, err := org.GetAllSecurityTaggedEntities(queryParameters)
	if err != nil {
		return nil, err
	}

	returnSecurityTags := &types.SecurityTag{
		Tag:      securityTag.Tag,
		Entities: make([]string, len(readEntities)),
	}

	for i, entity := range readEntities {
		returnSecurityTags.Entities[i] = entity.ID
	}

	return returnSecurityTags, nil
}

// UpdateVMSecurityTags updates the list of tags for a specific VM. An empty list of tags means to delete all tags
// for the VM. If user has edit permission on the VM, user can edit its tags.
// This function works from API v36.0 (VCD 10.3.0+)
func (vm *VM) UpdateVMSecurityTags(entitySecurityTags *types.EntitySecurityTags) (*types.EntitySecurityTags, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags
	apiVersion, err := vm.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vm.client.OpenApiBuildEndpoint(endpoint, fmt.Sprintf("/vm/%s", vm.VM.ID))
	if err != nil {
		return nil, err
	}

	var serverEntitySecurityTags types.EntitySecurityTags
	err = vm.client.OpenApiPutItem(apiVersion, urlRef, nil, entitySecurityTags, &serverEntitySecurityTags, nil)
	if err != nil {
		return nil, err
	}

	return &serverEntitySecurityTags, nil
}
