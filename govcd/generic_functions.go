package govcd

import (
	"fmt"
	"net/url"
	"reflect"
)

// oneOrError is used to cover up a common pattern in this codebase which is usually used in
// GetXByName functions.
// API endpoint returns N elements for an object we are looking (most commonly because API does not
// support filtering) and final filtering by Name must be done in code.
// After filtering returned entities one must be sure that exactly one was found and handle 3 cases:
// * If 0 entities are found - an error containing ErrorEntityNotFound must be returned
// * If >1 entities are found - an error containing the number of entities must be returned
// * If 1 entity was found - return it
//
// An example of code that was previously handled in non generic way - we had a lot of these
// occurrences throughout the code:
//
// if len(nsxtEdgeClusters) == 0 {
//     // ErrorEntityNotFound is injected here for the ability to validate problem using ContainsNotFound()
//     return nil, fmt.Errorf("%s: no NSX-T Tier-0 Edge Cluster with name '%s' for Org VDC with id '%s' found",
//             ErrorEntityNotFound, name, vdc.Vdc.ID)
// }

//	if len(nsxtEdgeClusters) > 1 {
//	        return nil, fmt.Errorf("more than one (%d) NSX-T Edge Cluster with name '%s' for Org VDC with id '%s' found",
//	                len(nsxtEdgeClusters), name, vdc.Vdc.ID)
//	}
func oneOrError[T any](key, name string, entitySlice []*T) (*T, error) {
	if len(entitySlice) > 1 {
		return nil, fmt.Errorf("got more than one entity by %s '%s' %d", key, name, len(entitySlice))
	}

	if len(entitySlice) == 0 {
		// No entity found - returning ErrorEntityNotFound as it must be wrapped in the returned error
		return nil, fmt.Errorf("%s: got zero entities by %s '%s'", ErrorEntityNotFound, key, name)
	}

	return entitySlice[0], nil
}

// genericGetSingleBareEntity is an implementation for a common pattern in our code where we have to
// retrieve bare entity (usually *types.XXXX) and does not need to be wrapped in a parent container.
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `exactEndpoint` is that same endpoint with placeholders filled (usually these are URNs of entities in the path)
// * `entityName` is used for detailing error messages with an explicit entity name
func genericGetSingleBareEntity[T any](client *Client, endpoint, exactEndpoint string, queryParameters url.Values, entityName string) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for entity '%s': %s", entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s': %s", entityName, err)
	}

	typeResponse := new(T)
	err = client.OpenApiGetItem(apiVersion, urlRef, queryParameters, typeResponse, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving entity of type '%s': %s", entityName, err)
	}

	return typeResponse, nil
}

// genericGetAllBareFilteredEntities can be used to retrieve a slice of any entity in the OpenAPI
// endpoints that are not nested into a parent type
//
// An example usage which can be found in nsxt_manager.go:
// endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentMacDiscoveryProfiles
// return genericGetAllBareFilteredEntities[types.NsxtSegmentProfileTemplateMacDiscovery](client, endpoint, queryParameters)
//
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `exactEndpoint` is that same endpoint with placeholders filled (usually these are URNs of entities)
// * `entityName` is used for detailing error messages with an explicit entity name
// * queryParameters can be applied to API. Usually these are filtering parameters
func genericGetAllBareFilteredEntities[T any](client *Client, endpoint, exactEndpoint string, queryParameters url.Values, entityName string) ([]*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for entity '%s': %s", entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s': %s", entityName, err)
	}

	typeResponses := make([]*T, 0)
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all entities of type '%s': %s", entityName, err)
	}

	return typeResponses, nil
}

// genericLocalFilter performs filtering of a type T based on a field name `fieldName` and its
// expected string value `expectedFieldValue`. Common use case for GetAllX methods where API does
// not support filtering and it must be done on client side.
//
// Note The field name `fieldName` must be present in a given type T
func genericLocalFilter[T any](entities []*T, fieldName, expectedFieldValue string, entityName string) ([]*T, error) {
	if len(entities) == 0 {
		return nil, fmt.Errorf("zero entities provided for filtering")
	}

	filteredValues := make([]*T, 0)

	for _, entity := range entities {

		// Need to deference pointer because `reflect` package requires to work with types and not
		// pointers to types
		var entityValue T
		if entity != nil {
			entityValue = *entity
		} else {
			return nil, fmt.Errorf("given entity for %s is a nil pointer", entityName)
		}

		value := reflect.ValueOf(entityValue)
		field := value.FieldByName(fieldName)

		if !field.IsValid() {
			return nil, fmt.Errorf("the struct for %s does not have the field '%s'", entityName, fieldName)
		}

		if field.Type().Name() != "string" {
			return nil, fmt.Errorf("field '%s' is not string type, it has type '%s'", fieldName, field.Type().Name())
		}

		if field.String() == expectedFieldValue {
			filteredValues = append(filteredValues, entity)
		}
	}

	return filteredValues, nil
}

// genericLocalFilterOneOrError performs local filtering using `genericLocalFilter()` and
// additionally verifies that only a single result is present using `oneOrError()`. Common use case
// for GetXByName methods where API does not support filtering and it must be done on client side.
func genericLocalFilterOneOrError[T any](entities []*T, fieldName, expectedFieldValue string, entityName string) (*T, error) {
	if fieldName == "" || expectedFieldValue == "" {
		return nil, fmt.Errorf("expected field name and value must be specified to filter %s", entityName)
	}

	filteredValues, err := genericLocalFilter(entities, fieldName, expectedFieldValue, entityName)
	if err != nil {
		return nil, err
	}

	return oneOrError(fieldName, expectedFieldValue, filteredValues)
}

// genericUpdateBareEntity implements a common pattern for updating entity throughout codebase
// Two types of invocation are possible because the type T can be identified (it is a required parameter)
// * genericUpdateBareEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, entityConfig)
// * genericUpdateBareEntity(&client, endpoint, endpoint, entityConfig)
//
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `exactEndpoint` is that same endpoint with placeholders filled (usually these are URNs of entities)
// * `entityName` is used for detailing error messages with an explicit entity name
func genericUpdateBareEntity[T any](client *Client, endpoint, exactEndpoint string, entityConfig *T, entityName string) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for updating entity '%s': %s", entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s' update: %s", entityName, err)
	}

	updatedEntityConfig := new(T)
	err = client.OpenApiPutItem(apiVersion, urlRef, nil, entityConfig, updatedEntityConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating entity of type '%s': %s", entityName, err)
	}

	return updatedEntityConfig, nil
}

// genericCreateBareEntity implements a common pattern for creating an entity throughout codebase
// Two types of invocation are possible because the type T can be identified (it is a required parameter)
// * genericUpdateBareEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, entityConfig)
// * genericUpdateBareEntity(&client, endpoint, endpoint, entityConfig)
//
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `exactEndpoint` is that same endpoint with placeholders filled (usually these are URNs of entities)
// * `entityName` is used for detailing error messages with an explicit entity name
func genericCreateBareEntity[T any](client *Client, endpoint, exactEndpoint string, entityConfig *T, entityName string) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting API version for creating entity '%s': %s", entityName, err)
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building API endpoint for entity '%s' creation: %s", entityName, err)
	}

	createdEntityConfig := new(T)
	err = client.OpenApiPostItem(apiVersion, urlRef, nil, entityConfig, createdEntityConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating entity of type '%s': %s", entityName, err)
	}

	return createdEntityConfig, nil
}

// deleteById performs a common operation for OpenAPI endpoints that calls DELETE method for a given
// endpoint.
//
// * `endpoint` is the endpoint as specified in `endpointMinApiVersions`
// * `exactEndpoint` is that same endpoint with placeholders filled (usually these are URNs of entities)
// * `entityName` is used for detailing error messages with an explicit entity name
func deleteById(client *Client, endpoint string, exactEndpoint string, entityName string) error {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting %s: %s", entityName, err)
	}

	return nil
}
