package govcd

import "fmt"

// Generic type explanations
// Common generic parameter names seen in this code
// O - Outer type that is in the `govcd` package. (e.g. 'IpSpace')
// I - Inner type the type that is being marshalled/unmarshalled (usually in `types` package. E.g. `types.IpSpace`)

// outerEntityWrapper is a type constraint that outer entities must implement in order to
// use generic CRUD functions defined in this file
type outerEntityWrapper[O any, I any] interface {
	// wrap is a value receiver function that must implement one thing for a concrete type - wrap
	// pointer to innter entity I and return pointer to outer entity O
	wrap(inner *I) *O
}

// createOuterEntity creates an outer entity with given inner entity config
func createOuterEntity[O outerEntityWrapper[O, I], I any](client *Client, outerEntity O, c crudConfig, innerEntityConfig *I) (*O, error) {
	if innerEntityConfig == nil {
		return nil, fmt.Errorf("entity config '%s' cannot be empty for create operation", c.entityName)
	}

	createdInnerEntity, err := createInnerEntity(client, c, innerEntityConfig)
	if err != nil {
		return nil, err
	}

	return outerEntity.wrap(createdInnerEntity), nil
}

// updateOuterEntity updates an outer entity with given inner entity config
func updateOuterEntity[O outerEntityWrapper[O, I], I any](client *Client, outerEntity O, c crudConfig, innerEntityConfig *I) (*O, error) {
	if innerEntityConfig == nil {
		return nil, fmt.Errorf("entity config '%s' cannot be empty for update operation", c.entityName)
	}

	updatedInnerEntity, err := updateInnerEntity(client, c, innerEntityConfig)
	if err != nil {
		return nil, err
	}

	return outerEntity.wrap(updatedInnerEntity), nil
}

// getOuterEntity retrieves a single outer entity
func getOuterEntity[O outerEntityWrapper[O, I], I any](client *Client, outerEntity O, c crudConfig) (*O, error) {
	retrievedInnerEntity, err := getInnerEntity[I](client, c)
	if err != nil {
		return nil, err
	}

	return outerEntity.wrap(retrievedInnerEntity), nil
}

// getAllOuterEntities retrieves all outer entities
func getAllOuterEntities[O outerEntityWrapper[O, I], I any](client *Client, outerEntity O, c crudConfig) ([]*O, error) {
	retrievedAllInnerEntities, err := getAllInnerEntities[I](client, c)
	if err != nil {
		return nil, err
	}

	wrappedOuterEntities := make([]*O, len(retrievedAllInnerEntities))
	for index, singleInnerEntity := range retrievedAllInnerEntities {
		// outerEntity.wrap() is a value receiver, therefore it creates a shallow copy for each call
		wrappedOuterEntities[index] = outerEntity.wrap(singleInnerEntity)
	}

	return wrappedOuterEntities, nil
}