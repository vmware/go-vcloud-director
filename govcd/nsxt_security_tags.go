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

	v := url.Values{}
	v.Set("filter", filter)
	rawValues, err := org.client.openApiGetAllPages(apiVersion, urlRef, v, nil, nil, nil)
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

// Retrieves the list of security tags that are in the organization and can be reused to tag an entity.
// The list of tags include tags assigned to entities within the organization.
//This API is meant for organization user only (i.e. not system provider).
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

	v := url.Values{}
	v.Set("filter", filter)
	rawValues, err := org.client.openApiGetAllPages(apiVersion, urlRef, v, nil, nil, orgContextHeaders)
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

func GetVMTags() {}

func UpdateSecurityTag() {}

func UpdateVMTags() {}
