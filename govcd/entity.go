/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

type genericGetter func(string, bool) (interface{}, error)

// getEntityByNameOrId finds a generic entity by Name Or ID
// On success, returns a pointer to the AdminVdc structure and a nil error
// On failure, returns a nil pointer and an error
// Example usage:
//
// func (org *Org) GetCatalogByNameOrId(identifier string, refresh bool) (*Catalog, error) {
// 	byName := func(name string, refresh bool) (interface{}, error) {
// 		return org.GetCatalogByName(name, refresh)
// 	}
// 	byId := func(id string, refresh bool) (interface{}, error) {
// 	  return org.GetCatalogById(id, refresh)
// 	}
// 	entity, err := getEntityByNameOrId(byName, byId, identifier, refresh)
// 	return entity.(*Catalog), err
// }
func getEntityByNameOrId(byName, byId genericGetter, identifier string, refresh bool) (interface{}, error) {

	var byNameErr, byIdErr error
	var entity interface{}

	entity, byIdErr = byId(identifier, refresh)
	if byIdErr == nil {
		// Found by ID
		return entity, nil
	}
	if IsNotFound(byIdErr) {
		// Not found by ID, try by name
		entity, byNameErr = byName(identifier, false)
		return entity, byNameErr
	} else {
		// On any other error, we return it
		return nil, byIdErr
	}
}
