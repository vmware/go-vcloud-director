package govcd

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

// GetSecurityTaggedEntities Retrieves the list of entities that have at least one tag assigned to it.
// Besides entityType, additional supported filters are:
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

func GetTagValues() {}

func GetVMTags() {}

func UpdateSecurityTag() {}

func UpdateVMTags() {}
