package govcd

import (
	"fmt"
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
func oneOrError[E any](key, value string, entitySlice []*E) (*E, error) {
	if len(entitySlice) > 1 {
		return nil, fmt.Errorf("got more than one entity by %s '%s' %d", key, value, len(entitySlice))
	}

	if len(entitySlice) == 0 {
		// No entity found - returning ErrorEntityNotFound as it must be wrapped in the returned error
		return nil, fmt.Errorf("%s: got zero entities by %s '%s'", ErrorEntityNotFound, key, value)
	}

	return entitySlice[0], nil
}

// localFilter performs filtering of a type E based on a field name `fieldName` and its
// expected string value `expectedFieldValue`. Common use case for GetAllX methods where API does
// not support filtering and it must be done on the client side.
//
// Note. The field name `fieldName` must be present in a given type E (letter casing is important)
func localFilter[E any](entityLabel string, entities []*E, fieldName, expectedFieldValue string) ([]*E, error) {
	if len(entities) == 0 {
		return nil, fmt.Errorf("zero entities provided for filtering")
	}

	filteredValues := make([]*E, 0)
	for _, entity := range entities {

		// Need to deference pointer because `reflect` package requires to work with types and not
		// pointers to types
		var entityValue E
		if entity != nil {
			entityValue = *entity
		} else {
			return nil, fmt.Errorf("given entity for %s is a nil pointer", entityLabel)
		}

		value := reflect.ValueOf(entityValue)
		field := value.FieldByName(fieldName)

		if !field.IsValid() {
			return nil, fmt.Errorf("the struct for %s does not have the field '%s'", entityLabel, fieldName)
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

// localFilterOneOrError performs local filtering using `genericLocalFilter()` and
// additionally verifies that only a single result is present using `oneOrError()`. Common use case
// for GetXByName methods where API does not support filtering and it must be done on client side.
func localFilterOneOrError[E any](entityLabel string, entities []*E, fieldName, expectedFieldValue string) (*E, error) {
	if fieldName == "" || expectedFieldValue == "" {
		return nil, fmt.Errorf("expected field name and value must be specified to filter %s", entityLabel)
	}

	filteredValues, err := localFilter(entityLabel, entities, fieldName, expectedFieldValue)
	if err != nil {
		return nil, err
	}

	return oneOrError(fieldName, expectedFieldValue, filteredValues)
}
