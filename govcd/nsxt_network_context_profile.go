package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetAllNetworkContextProfiles retrieves a slice of types.NsxtNetworkContextProfile
// This function requires at least a filter value for 'context_id' which can be one of:
// * Org VDC ID - to get Network Context Profiles scoped for VDC
// * Network provider ID - to get Network Context Profiles scoped for attached NSX-T environment
// * VDC Group ID - to get Network Context Profiles scoped for attached NSX-T environment
func GetAllNetworkContextProfiles(client *Client, queryParameters url.Values) ([]*types.NsxtNetworkContextProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtNetworkContextProfile{}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetNetworkContextProfilesByScopeAndName retrieves a single NSX-T Network Context Profile by name
// and context ID. All fields - name, scope and contextId are mandatory
//
// contextId is mandatory and can be one off:
// * Org VDC ID - to get Network Context Profiles scoped for VDC
// * Network provider ID - to get Network Context Profiles scoped for attached NSX-T environment
// * VDC Group ID - to get Network Context Profiles scoped for attached NSX-T environment
//
// scope can be one off:
// * SYSTEM
// * PROVIDER
// * TENANT
func GetNetworkContextProfilesByNameScopeAndContext(client *Client, name, scope, contextId string) (*types.NsxtNetworkContextProfile, error) {
	if name == "" || contextId == "" || scope == "" {
		return nil, fmt.Errorf("error - 'name', 'scope' and 'contextId' must be specified")
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams.Add("filter", fmt.Sprintf("name==%s", name))
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", contextId), queryParams)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("scope==%s", scope), queryParams)

	allProfiles, err := GetAllNetworkContextProfiles(client, queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Network Context Profiles by name '%s', scope '%s' and context ID '%s': %s ",
			name, scope, contextId, err)
	}

	return returnSingleNetworkContextProfile(allProfiles)
}

func returnSingleNetworkContextProfile(allProfiles []*types.NsxtNetworkContextProfile) (*types.NsxtNetworkContextProfile, error) {
	if len(allProfiles) > 1 {
		return nil, fmt.Errorf("got more than 1 NSX-T Network Context Profile %d", len(allProfiles))
	}

	if len(allProfiles) < 1 {
		return nil, fmt.Errorf("%s: got 0 NSX-T Network Context Profiles", ErrorEntityNotFound)
	}

	return allProfiles[0], nil
}
