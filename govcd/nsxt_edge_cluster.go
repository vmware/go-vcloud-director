/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtEdgeCluster is a logical grouping of NSX-T Edge virtual machines.
type NsxtEdgeCluster struct {
	NsxtEdgeCluster *types.NsxtEdgeCluster
	client          *Client
}

// GetNsxtEdgeClusterByName retrieves a particular NSX-T Edge Cluster by name available for that VDC
// Note: Multiple NSX-T Edge Clusters with the same name may exist.
func (vdc *Vdc) GetNsxtEdgeClusterByName(name string) (*NsxtEdgeCluster, error) {
	if name == "" {
		return nil, fmt.Errorf("empty NSX-T Edge Cluster name specified")
	}

	// Ideally FIQL filter could be used to filter on server side and get only desired result, but filtering on
	// 'name' is not yet supported. The only supported field for filtering is
	// _context==urn:vcloud:vdc:09722307-aee0-4623-af95-7f8e577c9ebc to specify parent Org VDC (This
	// automatically happens in GetAllNsxtEdgeClusters()). The below filter injection is left as documentation.
	/*
		queryParameters := copyOrNewUrlValues(nil)
		queryParameters.Add("filter", "name=="+name)
	*/

	nsxtEdgeClusters, err := vdc.GetAllNsxtEdgeClusters(nil)
	if err != nil {
		return nil, fmt.Errorf("could not find NSX-T Edge Cluster with name '%s' for Org VDC with id '%s': %s",
			name, vdc.Vdc.ID, err)
	}

	// TODO remove this when FIQL supports filtering on 'name'
	nsxtEdgeClusters = filterNsxtEdgeClusters(name, nsxtEdgeClusters)
	// EOF TODO remove this when FIQL supports filtering on 'name'

	if len(nsxtEdgeClusters) == 0 {
		// ErrorEntityNotFound is injected here for the ability to validate problem using ContainsNotFound()
		return nil, fmt.Errorf("%s: no NSX-T Tier-0 Edge Cluster with name '%s' for Org VDC with id '%s' found",
			ErrorEntityNotFound, name, vdc.Vdc.ID)
	}

	if len(nsxtEdgeClusters) > 1 {
		return nil, fmt.Errorf("more than one (%d) NSX-T Edge Cluster with name '%s' for Org VDC with id '%s' found",
			len(nsxtEdgeClusters), name, vdc.Vdc.ID)
	}

	return nsxtEdgeClusters[0], nil
}

// filterNsxtEdgeClusters is a helper to filter NSX-T Edge Clusters by name because the FIQL filter does not support
// filtering by name.
func filterNsxtEdgeClusters(name string, allNnsxtEdgeCluster []*NsxtEdgeCluster) []*NsxtEdgeCluster {
	filteredNsxtEdgeClusters := make([]*NsxtEdgeCluster, 0)
	for index, nsxtEdgeCluster := range allNnsxtEdgeCluster {
		if allNnsxtEdgeCluster[index].NsxtEdgeCluster.Name == name {
			filteredNsxtEdgeClusters = append(filteredNsxtEdgeClusters, nsxtEdgeCluster)
		}
	}

	return filteredNsxtEdgeClusters

}

// GetAllNsxtEdgeClusters retrieves all available Edge Clusters for a particular VDC
func (vdc *Vdc) GetAllNsxtEdgeClusters(queryParameters url.Values) ([]*NsxtEdgeCluster, error) {
	if vdc.Vdc.ID == "" {
		return nil, fmt.Errorf("VDC must have ID populated to retrieve NSX-T edge clusters")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeClusters
	minimumApiVersion, err := vdc.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vdc.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	// Get all NSX-T Edge clusters that are accessible to an organization VDC. The “_context” filter key must be set with
	// the ID of the VDC for which we want to get available Edge Clusters for.
	//
	// _context==urn:vcloud:vdc:09722307-aee0-4623-af95-7f8e577c9ebc

	// Create a copy of queryParameters so that original queryParameters are not mutated (because a map is always a
	// reference)
	queryParams := queryParameterFilterAnd("_context=="+vdc.Vdc.ID, queryParameters)

	typeResponses := []*types.NsxtEdgeCluster{{}}
	err = vdc.client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	returnObjects := make([]*NsxtEdgeCluster, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnObjects[sliceIndex] = &NsxtEdgeCluster{
			NsxtEdgeCluster: typeResponses[sliceIndex],
			client:          vdc.client,
		}
	}

	return returnObjects, nil
}
