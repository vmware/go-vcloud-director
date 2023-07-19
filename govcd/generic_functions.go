package govcd

import (
	"fmt"
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
