/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtEdgeCluster
type NsxtEdgeCluster struct {
	NsxtEdgeCluster *types.NsxtEdgeCluster
	client          *Client
}

func (vdc *Vdc) GetNsxtEdgeClusterByName(name string) (*NsxtEdgeCluster, error) {
	if name == "" {
		return nil, fmt.Errorf("empty NSX-T Edge Cluster name specified")
	}

	// Ideally FIQL filter could be used to filter on server side and get only desired result, but filtering on
	// 'name' is not yet supported. The only supported field for filtering is
	// _context==urn:vcloud:vdc:09722307-aee0-4623-af95-7f8e577c9ebc to specify parent Org Vdc (This
	// automatically happens in GetAllNsxtEdgeClusters()). The below filter injection is left as documentation.
	/*
		queryParameters := copyOrNewUrlValues(nil)
		queryParameters.Add("filter", "name=="+name)
	*/

	nsxtEdgeClusters, err := vdc.GetAllNsxtEdgeClusters(nil)
	if err != nil {
		return nil, fmt.Errorf("could not find NSX-T Edge Cluster with name '%s' for Org Vdc with id '%s': %s",
			name, vdc.Vdc.ID, err)
	}

	// TODO remove this when FIQL supports filtering on 'name'
	nsxtEdgeClusters = filterNsxtEdgeClusters(name, nsxtEdgeClusters)
	// EOF TODO remove this when FIQL supports filtering on 'name'

	if len(nsxtEdgeClusters) == 0 {
		// ErrorEntityNotFound is injected here for the ability to validate problem using ContainsNotFound()
		return nil, fmt.Errorf("%s: no NSX-T Tier-0 Edge Cluster with name '%s' for Org Vdc with id '%s' found",
			ErrorEntityNotFound, name, vdc.Vdc.ID)
	}

	if len(nsxtEdgeClusters) > 1 {
		return nil, fmt.Errorf("more than one (%d) NSX-T Edge Cluster with name '%s' for Org Vdc with id '%s' found",
			len(nsxtEdgeClusters), name, vdc.Vdc.ID)
	}

	return nsxtEdgeClusters[0], nil
}

// GetNsxtEdgeClusterByName searches for NSX-T Edge Cluster by given name
// func (vcdCli *VCDClient) GetNsxtEdgeClusterByName(name, orgVdcId string) (*NsxtEdgeCluster, error) {
// 	if name == "" {
// 		return nil, fmt.Errorf("empty NSX-T Edge Cluster name specified")
// 	}
//
// 	// Ideally FIQL filter could be used to filter on server side and get only desired result, but filtering on
// 	// 'name' is not yet supported. The only supported field for filtering is
// 	// _context==urn:vcloud:vdc:09722307-aee0-4623-af95-7f8e577c9ebc to specify parent Org Vdc (This
// 	// automatically happens in GetAllNsxtEdgeClusters()). The below filter injection is left as documentation.
// 	/*
// 		queryParameters := copyOrNewUrlValues(nil)
// 		queryParameters.Add("filter", "name=="+name)
// 	*/
//
// 	nsxtEdgeClusters, err := vcdCli.GetAllNsxtEdgeClusters(orgVdcId, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not find NSX-T Edge Cluster with name '%s' for Org Vdc with id '%s': %s",
// 			name, orgVdcId, err)
// 	}
//
// 	// TODO remove this when FIQL supports filtering on 'name'
// 	nsxtEdgeClusters = filterNsxtEdgeClusters(name, nsxtEdgeClusters)
// 	// EOF TODO remove this when FIQL supports filtering on 'name'
//
// 	if len(nsxtEdgeClusters) == 0 {
// 		// ErrorEntityNotFound is injected here for the ability to validate problem using ContainsNotFound()
// 		return nil, fmt.Errorf("%s: no NSX-T Tier-0 Edge Cluster with name '%s' for Org Vdc with id '%s' found",
// 			ErrorEntityNotFound, name, orgVdcId)
// 	}
//
// 	if len(nsxtEdgeClusters) > 1 {
// 		return nil, fmt.Errorf("more than one (%d) NSX-T Edge Cluster with name '%s' for Org Vdc with id '%s' found",
// 			len(nsxtEdgeClusters), name, orgVdcId)
// 	}
//
// 	return nsxtEdgeClusters[0], nil
// }

func filterNsxtEdgeClusters(name string, allNnsxtEdgeCluster []*NsxtEdgeCluster) []*NsxtEdgeCluster {
	filteredNsxtEdgeClusters := make([]*NsxtEdgeCluster, 0)
	for index, nsxtTier0Router := range allNnsxtEdgeCluster {
		if allNnsxtEdgeCluster[index].NsxtEdgeCluster.Name == name {
			filteredNsxtEdgeClusters = append(filteredNsxtEdgeClusters, nsxtTier0Router)
		}
	}

	return filteredNsxtEdgeClusters

}

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

	// Get all Tier-0 routers that are accessible to an organization VDC. Routers that are already associated with an
	// External Network are filtered out. The “_context” filter key must be set with the id of the NSX-T manager for which
	// we want to get the Tier-0 routers for.
	//
	// _context==urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc

	// Create a copy of queryParameters so that original queryParameters are not mutated (because a map is always a
	// reference)
	queryParams := queryParameterFilterAnd("_context=="+vdc.Vdc.ID, queryParameters)

	typeResponses := []*types.NsxtEdgeCluster{{}}
	err = vdc.client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParams, &typeResponses)
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

// GetAllNsxtEdgeClusters retrieves all NSX-T Edge Clusters using OpenAPI endpoint. Query parameters can be
// supplied to perform additional filtering. By default it injects FIQL filter _context==orgVdcId (e.g.
// _context==urn:vcloud:vdc:09722307-aee0-4623-af95-7f8e577c9ebc) because it is mandatory to list child edge clusters.
//
// Note. IDs of returned NSX-T Edge Clusters are returned as plain UUIDs (instead of VCD URNs)
// func (vcdCli *VCDClient) GetAllNsxtEdgeClusters(orgVdcId string, queryParameters url.Values) ([]*NsxtEdgeCluster, error) {
// 	if !isUrn(orgVdcId) {
// 		return nil, fmt.Errorf("NSX-T manager ID is not URN (e.g. 'urn:vcloud:vdc:09722307-aee0-4623-af95-7f8e577c9ebc)', got: %s", orgVdcId)
// 	}
//
// 	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeClusters
// 	minimumApiVersion, err := vcdCli.Client.checkOpenApiEndpointCompatibility(endpoint)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	urlRef, err := vcdCli.Client.OpenApiBuildEndpoint(endpoint)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Get all Tier-0 routers that are accessible to an organization VDC. Routers that are already associated with an
// 	// External Network are filtered out. The “_context” filter key must be set with the id of the NSX-T manager for which
// 	// we want to get the Tier-0 routers for.
// 	//
// 	// _context==urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc
//
// 	// Create a copy of queryParameters so that original queryParameters are not mutated (because a map is always a
// 	// reference)
// 	queryParams := queryParameterFilterAnd("_context=="+orgVdcId, queryParameters)
//
// 	typeResponses := []*types.NsxtEdgeCluster{{}}
// 	err = vcdCli.Client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParams, &typeResponses)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	returnObjects := make([]*NsxtEdgeCluster, len(typeResponses))
// 	for sliceIndex := range typeResponses {
// 		returnObjects[sliceIndex] = &NsxtEdgeCluster{
// 			NsxtEdgeCluster: typeResponses[sliceIndex],
// 			client:          &vcdCli.Client,
// 		}
// 	}
//
// 	return returnObjects, nil
// }
