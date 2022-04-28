package govcd

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

// GetSecurityTaggedEntities Retrieves the list of entities that have at least one tag assigned to it.
// Besides, entityType, additional supported filters are:
//   - tag - The tag to search by. I.e: filter=(tag==Web;entityType==vm)
// This function works from API v36.1 (VCD 10.3.1+)
func (org *Org) GetSecurityTaggedEntities(filter string) ([]*types.SecurityTaggedEntity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags
	apiVersion, err := org.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := org.client.OpenApiBuildEndpoint(endpoint, "/entities")
	if err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	queryParams.Set("filter", filter)
	rawValues, err := org.client.openApiGetAllPages(apiVersion, urlRef, queryParams, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	securityTaggedEntities := make([]*types.SecurityTaggedEntity, len(rawValues))

	for i, k := range rawValues {
		var temp types.SecurityTaggedEntity
		err = json.Unmarshal(k, &temp)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling SecurityTaggedEntity - %s", err)
		}
		securityTaggedEntities[i] = &temp
	}
	return securityTaggedEntities, nil
}

// GetSecurityTagValues Retrieves the list of security tags that are in the organization and can be reused to tag an entity.
// The list of tags include tags assigned to entities within the organization.
// This function works from API v36.1 (VCD 10.3.1+)
func (org *Org) GetSecurityTagValues(filter string) ([]*types.SecurityTagValue, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags
	apiVersion, err := org.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := org.client.OpenApiBuildEndpoint(endpoint, "/values")
	if err != nil {
		return nil, err
	}

	// If sysadmin, getting the org context and this method only works for org users
	orgContextHeaders := make(map[string]string)
	if org.client.IsSysAdmin {
		orgContextHeaders["X-VMWARE-VCLOUD-AUTH-CONTEXT"] = org.Org.Name
		orgContextHeaders["X-VMWARE-VCLOUD-TENANT-CONTEXT"] = org.Org.ID
	}

	queryParams := url.Values{}
	queryParams.Set("filter", filter)
	rawValues, err := org.client.openApiGetAllPages(apiVersion, urlRef, queryParams, nil, nil, orgContextHeaders)
	if err != nil {
		return nil, err
	}

	securityTaggedValues := make([]*types.SecurityTagValue, len(rawValues))

	for i, k := range rawValues {
		var temp types.SecurityTagValue
		err = json.Unmarshal(k, &temp)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling SecurityTaggedEntity - %s", err)
		}
		securityTaggedValues[i] = &temp
	}

	return securityTaggedValues, nil
}

// GetVMSecurityTags Retrieves the list of tags for a specific VM. If user has view right to the VM, user can view its tags.
// This function works from API v36.1 (VCD 10.3.1+)
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
// This function works from API v36.1 (VCD 10.3.1+)
func (org *Org) UpdateSecurityTag(securityTag *types.SecurityTag) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags
	apiVersion, err := org.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := org.client.OpenApiBuildEndpoint(endpoint, "/tag")
	if err != nil {
		return err
	}

	err = org.client.OpenApiPutItem(apiVersion, urlRef, nil, securityTag, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// UpdateVMSecurityTags updates the list of tags for a specific VM. An empty list of tags means to delete all dags
// for the VM. If user has edit permission on the VM, user can edit its tags.
// This function works from API v36.1 (VCD 10.3.1+)
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
