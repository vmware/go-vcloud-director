package govcd

import (
	"fmt"
	"net/url"
	"strings"
)

// genericCrudConfig contains configuration that must be supplied when invoking generic functions
type genericCrudConfig struct {
	// endpoint in the usual format (e.g. types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentIpDiscoveryProfiles)
	endpoint string
	// endpointParams contains a slice of strings that will be used to contruct request URL. It will
	// initially replace '%s' placeholders in the `endpoint` (if any) and will add them as suffix
	// afterwards
	endpointParams []string
	// entityName contains friendly entity name that is used for logging meaningful errors
	entityName string
	// queryParameters will be passed as GET queries to the URL. Usually they are used for API filtering parameters
	queryParameters url.Values
	// additionalHeader can be used to pass additional headers for API calls. One of the common purposes is to pass
	// tenant context
	additionalHeader map[string]string
}

// genericCreateBareEntity implements a common pattern for creating an entity throughout codebase
// Two types of invocation are possible because the type T can be identified (it is a required parameter)
// * genericCreateBareEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, entityConfig, entityName)
// * genericCreateBareEntity(&client, endpoint, endpoint, entityConfig, entityName)
// Parameters:
// * `client` is a *Client
// * `entityConfig` is the new entity type
// * `c` holds settings for performing API call
func genericCreateBareEntity[T any](client *Client, entityConfig *T, c genericCrudConfig) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for creating entity '%s': %s", c.entityName, err)
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s' creation: %s", c.entityName, err)
	}

	createdEntityConfig := new(T)
	err = client.OpenApiPostItem(apiVersion, urlRef, c.queryParameters, entityConfig, createdEntityConfig, c.additionalHeader)
	if err != nil {
		return nil, fmt.Errorf("error creating entity of type '%s': %s", c.entityName, err)
	}

	return createdEntityConfig, nil
}

// genericUpdateBareEntity implements a common pattern for updating entity throughout codebase
// Two types of invocation are possible because the type T can be identified (it is a required parameter)
// * genericCreateBareEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, entityConfig, entityName)
// * genericCreateBareEntity(&client, endpoint, endpoint, entityConfig, entityName)
// Parameters:
// * `client` is a *Client
// * `entityConfig` is the new entity type
// * `c` holds settings for performing API call
func genericUpdateBareEntity[T any](client *Client, entityConfig *T, c genericCrudConfig) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for updating entity '%s': %s", c.entityName, err)
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s' update: %s", c.entityName, err)
	}

	updatedEntityConfig := new(T)
	err = client.OpenApiPutItem(apiVersion, urlRef, c.queryParameters, entityConfig, updatedEntityConfig, c.additionalHeader)
	if err != nil {
		return nil, fmt.Errorf("error updating entity of type '%s': %s", c.entityName, err)
	}

	return updatedEntityConfig, nil
}

// genericGetSingleBareEntity is an implementation for a common pattern in our code where we have to
// retrieve bare entity (usually *types.XXXX) and does not need to be wrapped in a parent container.
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
func genericGetSingleBareEntity[T any](client *Client, c genericCrudConfig) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for entity '%s': %s", c.entityName, err)
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s': %s", c.entityName, err)
	}

	typeResponse := new(T)
	err = client.OpenApiGetItem(apiVersion, urlRef, c.queryParameters, typeResponse, c.additionalHeader)
	if err != nil {
		return nil, fmt.Errorf("error retrieving entity of type '%s': %s", c.entityName, err)
	}

	return typeResponse, nil
}

// genericGetAllBareFilteredEntities can be used to retrieve a slice of any entity in the OpenAPI
// endpoints that are not nested into a parent type
//
// An example usage which can be found in nsxt_manager.go:
// endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentMacDiscoveryProfiles
// return genericGetAllBareFilteredEntities[types.NsxtSegmentProfileTemplateMacDiscovery](client, endpoint, queryParameters)
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
func genericGetAllBareFilteredEntities[T any](client *Client, c genericCrudConfig) ([]*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for entity '%s': %s", c.entityName, err)
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s': %s", c.entityName, err)
	}

	typeResponses := make([]*T, 0)
	err = client.OpenApiGetAllItems(apiVersion, urlRef, c.queryParameters, &typeResponses, c.additionalHeader)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all entities of type '%s': %s", c.entityName, err)
	}

	return typeResponses, nil
}

// deleteById performs a common operation for OpenAPI endpoints that calls DELETE method for a given
// endpoint.
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
func deleteById(client *Client, c genericCrudConfig) error {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return err
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, c.queryParameters, c.additionalHeader)

	if err != nil {
		return fmt.Errorf("error deleting %s: %s", c.entityName, err)
	}

	return nil
}

func urlFromEndpoint(endpoint string, endpointParams []string) (string, error) {
	// Count how many '%s' placeholders exist in the 'endpoint'
	placeholderCount := strings.Count(endpoint, "%s")

	// Validation. At the very minimum all placeholders must have their replacements - otherwise it
	// is an error as we never want to query an endpoint that still has placeholders '%s'
	if len(endpointParams) < placeholderCount {
		return "", fmt.Errorf("endpoint '%s' has unpopulated placeholders", endpoint)
	}

	// if there are no 'endpointParams' - exit with the same endpoint
	if len(endpointParams) == 0 {
		return endpoint, nil
	}

	// Loop over given endpointParams and replace placeholders at first. Afterwards - amend any
	// additional parameters to the end of endpoint
	for _, v := range endpointParams {
		// If there are placeholders '%s' to replace in the endpoint itself - do it
		if placeholderCount > 0 {
			endpoint = strings.Replace(endpoint, "%s", v, 1)
			placeholderCount = placeholderCount - 1
			continue
		}

		endpoint = endpoint + v
	}

	return endpoint, nil
}
