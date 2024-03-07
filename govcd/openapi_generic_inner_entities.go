package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// crudConfig contains configuration that must be supplied when invoking generic functions that are defined
// in `openapi_generic_inner_entities.go` and `openapi_generic_outer_entities.go`
type crudConfig struct {
	// Mandatory parameters

	// entityLabel contains friendly entity name that is used for logging meaningful errors
	entityLabel string

	// endpoint in the usual format (e.g. types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentIpDiscoveryProfiles)
	endpoint string

	// Optional parameters

	// endpointParams contains a slice of strings that will be used to construct the request URL. It will
	// initially replace '%s' placeholders in the `endpoint` (if any) and will add them as suffix
	// afterwards
	endpointParams []string

	// queryParameters will be passed as GET queries to the URL. Usually they are used for API filtering parameters
	queryParameters url.Values
	// additionalHeader can be used to pass additional headers for API calls. One of the common purposes is to pass
	// tenant context
	additionalHeader map[string]string
}

// validate should catch errors in consuming generic CRUD functions and should never produce false
// positives.
func (c crudConfig) validate() error {
	// crudConfig misconfiguration - we can panic so that developer catches the problem during
	// development of this SDK
	if c.entityLabel == "" {
		panic("'entityName' must always be specified when initializing crudConfig")
	}

	if c.endpoint == "" {
		panic("'endpoint' must always be specified when initializing crudConfig")
	}

	// softer validations that consumers of this SDK can manipulate

	// If `endpointParams` is specified in `crudConfig` (it is not nil), then it must contain at
	// least a single non empty string parameter
	// Such validation should prevent cases where some ID is not speficied upon function call.
	// E.g.: endpointParams: []string{vdcId}, <--- vdcId comes from consumer of the SDK
	// If the user specified empty `vdcId` - we'd validate this
	for _, paramValue := range c.endpointParams {
		if paramValue == "" {
			return fmt.Errorf(`endpointParams were specified but they contain empty value "" for %s. %#v`,
				c.entityLabel, c.endpointParams)
		}
	}

	return nil
}

// createInnerEntity implements a common pattern for creating an entity throughout codebase
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
// * `innerConfig` is the new entity type
func createInnerEntity[I any](client *Client, c crudConfig, innerConfig *I) (*I, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for creating entity '%s': %s", c.entityLabel, err)
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityLabel, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s' creation: %s", c.entityLabel, err)
	}

	createdInnerEntityConfig := new(I)
	err = client.OpenApiPostItem(apiVersion, urlRef, c.queryParameters, innerConfig, createdInnerEntityConfig, c.additionalHeader)
	if err != nil {
		return nil, fmt.Errorf("error creating entity of type '%s': %s", c.entityLabel, err)
	}

	return createdInnerEntityConfig, nil
}

// updateInnerEntity implements a common pattern for updating entity throughout codebase
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
// * `innerConfig` is the new entity type
func updateInnerEntity[I any](client *Client, c crudConfig, innerConfig *I) (*I, error) {
	// Discarding returned headers to better match return signature for most common cases
	updatedInnerEntity, _, err := updateInnerEntityWithHeaders(client, c, innerConfig)
	return updatedInnerEntity, err
}

// updateInnerEntityWithHeaders implements a common pattern for updating entity throughout codebase
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
// * `innerConfig` is the new entity type
func updateInnerEntityWithHeaders[I any](client *Client, c crudConfig, innerConfig *I) (*I, http.Header, error) {
	if err := c.validate(); err != nil {
		return nil, nil, err
	}

	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting API version for updating entity '%s': %s", c.entityLabel, err)
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return nil, nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityLabel, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("error building API endpoint for entity '%s' update: %s", c.entityLabel, err)
	}

	updatedInnerEntityConfig := new(I)
	headers, err := client.OpenApiPutItemAndGetHeaders(apiVersion, urlRef, c.queryParameters, innerConfig, updatedInnerEntityConfig, c.additionalHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("error updating entity of type '%s': %s", c.entityLabel, err)
	}

	return updatedInnerEntityConfig, headers, nil
}

// getInnerEntity is an implementation for a common pattern in our code where we have to retrieve
// outer entity (usually *types.XXXX) and does not need to be wrapped in an inner container entity.
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
func getInnerEntity[I any](client *Client, c crudConfig) (*I, error) {
	// Discarding returned headers to better match return signature for most common cases
	innerEntity, _, err := getInnerEntityWithHeaders[I](client, c)
	return innerEntity, err
}

// getInnerEntityWithHeaders is an implementation for a common pattern in our code where we have to retrieve
// outer entity (usually *types.XXXX) and does not need to be wrapped in an inner container entity.
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
func getInnerEntityWithHeaders[I any](client *Client, c crudConfig) (*I, http.Header, error) {
	if err := c.validate(); err != nil {
		return nil, nil, err
	}

	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting API version for entity '%s': %s", c.entityLabel, err)
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return nil, nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityLabel, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("error building API endpoint for entity '%s': %s", c.entityLabel, err)
	}

	typeResponse := new(I)
	headers, err := client.OpenApiGetItemAndHeaders(apiVersion, urlRef, c.queryParameters, typeResponse, c.additionalHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving entity of type '%s': %s", c.entityLabel, err)
	}

	return typeResponse, headers, nil
}

// getAllInnerEntities can be used to retrieve a slice of any inner entities in the OpenAPI
// endpoints that are not nested in outer types
//
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
func getAllInnerEntities[I any](client *Client, c crudConfig) ([]*I, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for entity '%s': %s", c.entityLabel, err)
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return nil, fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityLabel, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s': %s", c.entityLabel, err)
	}

	typeResponses := make([]*I, 0)
	err = client.OpenApiGetAllItems(apiVersion, urlRef, c.queryParameters, &typeResponses, c.additionalHeader)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all entities of type '%s': %s", c.entityLabel, err)
	}

	return typeResponses, nil
}

// deleteEntityById performs a common operation for OpenAPI endpoints that calls DELETE method for a
// given endpoint.
// Note. It does not use generics for the operation, but is held in this file with other CRUD entries
// Parameters:
// * `client` is a *Client
// * `c` holds settings for performing API call
func deleteEntityById(client *Client, c crudConfig) error {
	if err := c.validate(); err != nil {
		return err
	}

	apiVersion, err := client.getOpenApiHighestElevatedVersion(c.endpoint)
	if err != nil {
		return err
	}

	exactEndpoint, err := urlFromEndpoint(c.endpoint, c.endpointParams)
	if err != nil {
		return fmt.Errorf("error building endpoint '%s' with given params '%s' for entity '%s': %s", c.endpoint, strings.Join(c.endpointParams, ","), c.entityLabel, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, c.queryParameters, c.additionalHeader)

	if err != nil {
		return fmt.Errorf("error deleting %s: %s", c.entityLabel, err)
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
