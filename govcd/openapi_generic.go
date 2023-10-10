package govcd

import (
	"fmt"
	"net/url"
	"strings"
)

// genericCreateBareEntity implements a common pattern for creating an entity throughout codebase
// Two types of invocation are possible because the type T can be identified (it is a required parameter)
// * genericCreateBareEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, entityConfig, entityName)
// * genericCreateBareEntity(&client, endpoint, endpoint, entityConfig, entityName)
// Parameters:
// * `client` is a *Client
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `endpointParams` is a slice of strings to replace or append for a given `endpoint`
// * `queryParameters` for the API call. Could be used for passing tenant context or other values
// * `entityName` is used for detailing error messages with an explicit entity name
func genericCreateBareEntity[T any](entityName string, client *Client, endpoint string, endpointParams []string, entityConfig *T, queryParameters url.Values, additionalHeader map[string]string) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for creating entity '%s': %s", entityName, err)
	}

	exactEndpoint, err := urlFromEndpoint(endpoint, endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", endpoint, strings.Join(endpointParams, ","), entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s' creation: %s", entityName, err)
	}

	createdEntityConfig := new(T)
	err = client.OpenApiPostItem(apiVersion, urlRef, queryParameters, entityConfig, createdEntityConfig, additionalHeader)
	if err != nil {
		return nil, fmt.Errorf("error creating entity of type '%s': %s", entityName, err)
	}

	return createdEntityConfig, nil
}

// genericUpdateBareEntity implements a common pattern for updating entity throughout codebase
// Two types of invocation are possible because the type T can be identified (it is a required parameter)
// * genericCreateBareEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, entityConfig, entityName)
// * genericCreateBareEntity(&client, endpoint, endpoint, entityConfig, entityName)
// Parameters:
// * `client` is a *Client
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `endpointParams` is a slice of strings to replace or append for a given `endpoint`
// * `queryParameters` for the API call. Could be used for passing tenant context or other values
// * `entityName` is used for detailing error messages with an explicit entity name
func genericUpdateBareEntity[T any](entityName string, client *Client, endpoint string, endpointParams []string, entityConfig *T, queryParameters url.Values, additionalHeader map[string]string) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for updating entity '%s': %s", entityName, err)
	}

	exactEndpoint, err := urlFromEndpoint(endpoint, endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", endpoint, strings.Join(endpointParams, ","), entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s' update: %s", entityName, err)
	}

	updatedEntityConfig := new(T)
	err = client.OpenApiPutItem(apiVersion, urlRef, queryParameters, entityConfig, updatedEntityConfig, additionalHeader)
	if err != nil {
		return nil, fmt.Errorf("error updating entity of type '%s': %s", entityName, err)
	}

	return updatedEntityConfig, nil
}

// genericGetSingleBareEntity is an implementation for a common pattern in our code where we have to
// retrieve bare entity (usually *types.XXXX) and does not need to be wrapped in a parent container.
// Parameters:
// * `client` is a *Client
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `endpointParams` is a slice of strings to replace or append for a given `endpoint`
// * `queryParameters` for the API call. Most common use case - applying a `filter`
// * `entityName` is used for detailing error messages with an explicit entity name
func genericGetSingleBareEntity[T any](entityName string, client *Client, endpoint string, endpointParams []string, queryParameters url.Values, additionalHeader map[string]string) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for entity '%s': %s", entityName, err)
	}

	exactEndpoint, err := urlFromEndpoint(endpoint, endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", endpoint, strings.Join(endpointParams, ","), entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s': %s", entityName, err)
	}

	typeResponse := new(T)
	err = client.OpenApiGetItem(apiVersion, urlRef, queryParameters, typeResponse, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving entity of type '%s': %s", entityName, additionalHeader)
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
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `endpointParams` is a slice of strings to replace or append for a given `endpoint`
// * `queryParameters` can be applied to API. `queryParameters` for the API call. Most common use case - applying a `filter`
// * `entityName` is used for detailing error messages with an explicit entity name
func genericGetAllBareFilteredEntities[T any](entityName string, client *Client, endpoint string, endpointParams []string, queryParameters url.Values, additionalHeader map[string]string) ([]*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for entity '%s': %s", entityName, err)
	}

	exactEndpoint, err := urlFromEndpoint(endpoint, endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", endpoint, strings.Join(endpointParams, ","), entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s': %s", entityName, err)
	}

	typeResponses := make([]*T, 0)
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all entities of type '%s': %s", entityName, additionalHeader)
	}

	return typeResponses, nil
}

// deleteById performs a common operation for OpenAPI endpoints that calls DELETE method for a given
// endpoint.
// Parameters:
// * `client` is a *Client
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `endpointParams` is a slice of strings to replace or append for a given `endpoint`
// * `queryParameters` for the API call. Could be used for passing tenant context or other values
// * `entityName` is used for detailing error messages with an explicit entity name
func deleteById(entityName string, client *Client, endpoint string, endpointParams []string, queryParameters url.Values, additionalHeader map[string]string) error {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	exactEndpoint, err := urlFromEndpoint(endpoint, endpointParams)
	if err != nil {
		return fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", endpoint, strings.Join(endpointParams, ","), entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, queryParameters, additionalHeader)

	if err != nil {
		return fmt.Errorf("error deleting %s: %s", entityName, err)
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
