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

func genericGetSingleEntity[T any](client *Client, endpoint, exactEndpoint string, queryParameters url.Values) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, err
	}

	typeResponse := new(T)
	err = client.OpenApiGetItem(apiVersion, urlRef, queryParameters, typeResponse, nil)
	if err != nil {
		return nil, err
	}

	return typeResponse, nil
}

// genericGetAllFilteredEntities can be used to retrieve a slice of any entity in the OpenAPI
// endpoints that are not nested into a parent type
//
// An example usage which can be found in nsxt_manager.go:
// endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentMacDiscoveryProfiles
// return genericGetAllFilteredEntities[types.NsxtSegmentProfileTemplateMacDiscovery](client, endpoint, queryParameters)
func genericGetAllFilteredEntities[T any](client *Client, endpoint string, queryParameters url.Values) ([]*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := make([]*T, 0)
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

func genericLocalFilter[T any](entities []*T, expectedFieldName, expectedFieldValue string) ([]*T, error) {
	if len(entities) == 0 {
		return nil, fmt.Errorf("zero entities provided for filtering")
	}

	filteredValues := make([]*T, 0)

	for _, entity := range entities {

		// Need to deference pointer because `reflect` package requires to work with types and not
		// pointers of types
		var entityValue T
		if entity != nil {
			entityValue = *entity
		} else {
			return nil, fmt.Errorf("given entity is a nil pointer")
		}

		value := reflect.ValueOf(entityValue)
		field := value.FieldByName(expectedFieldName)

		if !field.IsValid() {
			return nil, fmt.Errorf("the struct does not have the field '%s'", expectedFieldName)
		}

		if field.Type().Name() != "string" {
			return nil, fmt.Errorf("field '%s' is not string type, it has type '%s'", expectedFieldName, field.Type().Name())
		}

		if field.String() == expectedFieldValue {
			filteredValues = append(filteredValues, entity)
		}
	}

	return filteredValues, nil
}

func genericLocalFilterOneOrError[T any](entities []*T, expectedFieldName, expectedFieldValue string) (*T, error) {
	if expectedFieldName == "" || expectedFieldValue == "" {
		return nil, fmt.Errorf("expected field name and value must be specified")
	}

	filteredValues, err := genericLocalFilter(entities, expectedFieldName, expectedFieldValue)
	if err != nil {
		return nil, err
	}

	return oneOrError(expectedFieldName, expectedFieldValue, filteredValues)
}

// Two types of invocation are possible because the type T can be identified (it is a required parameter)
// * genericUpdateEntity[types.NsxtSegmentProfileTemplateDefaultDefinition](&client, endpoint, endpoint, entityConfig)
// * genericUpdateEntity(&client, endpoint, endpoint, entityConfig)
func genericUpdateEntity[T any](client *Client, endpoint, exactEndpoint string, entityConfig *T) (*T, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(exactEndpoint)
	if err != nil {
		return nil, err
	}

	updatedEntityConfig := new(T)
	err = client.OpenApiPutItem(apiVersion, urlRef, nil, entityConfig, updatedEntityConfig, nil)
	if err != nil {
		return nil, err
	}

	return updatedEntityConfig, nil
}
