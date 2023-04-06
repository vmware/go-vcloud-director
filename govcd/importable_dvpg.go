/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// VcenterImportableDvpg is a read only structure that allows to get information about a Distributed
// Virtual Port Group (DVPG) network backing that is available for import.
type VcenterImportableDvpg struct {
	VcenterImportableDvpg *types.VcenterImportableDvpg
	client                *Client
}

// GetVcenterImportableDvpgByName
func (vcdClient *VCDClient) GetVcenterImportableDvpgByName(name, optionalVcenterId string) (*VcenterImportableDvpg, error) {
	if name == "" {
		return nil, fmt.Errorf("empty importable Distributed Virtual Port Group Name specified")
	}

	vcImportableDvpgs, err := vcdClient.GetAllVcenterImportableDvpgs(optionalVcenterId, nil)
	if err != nil {
		return nil, fmt.Errorf("could not find Distributed Virtual Port Group with Name '%s' for vCenter with ID '%s': %s",
			name, optionalVcenterId, err)
	}

	vcImportableDvpgs = filterVcImportableDvpgsInExternalNetworks(name, vcImportableDvpgs)

	return oneOrError("name", name, vcImportableDvpgs)
}

// GetAllVcenterImportableDvpgs retrieves all DVPGs that are available for import.
// They can be filtered by vCenter ID and/or Org vDC ID.
// The '_context' filter key is optional and can be set with the id of the vCenter from which to obtain the DVPG network backings.
// 'orgVdcId==[vdcUrn]' can be set as a filter to show importable DVPGs for an Org vDC.
func (vcdClient *VCDClient) GetAllVcenterImportableDvpgs(optionalVcenterId string, queryParameters url.Values) ([]*VcenterImportableDvpg, error) {
	return getAllVcenterImportableDvpgs(&vcdClient.Client, optionalVcenterId, queryParameters)
}

func (vdc *Vdc) GetVcenterImportableDvpgByName(name, optionalVcenterId string) (*VcenterImportableDvpg, error) {
	if name == "" {
		return nil, fmt.Errorf("empty importable Distributed Virtual Port Group Name specified")
	}

	vcImportableDvpgs, err := vdc.GetAllVcenterImportableDvpgs(optionalVcenterId, nil)
	if err != nil {
		return nil, fmt.Errorf("could not find NSX-T Tier-0 router with name '%s' for NSX-T manager with id '%s': %s",
			name, optionalVcenterId, err)
	}

	vcImportableDvpgs = filterVcImportableDvpgsInExternalNetworks(name, vcImportableDvpgs)

	return oneOrError("name", name, vcImportableDvpgs)
}

// GetAllVcenterImportableDvpgs retrieves all DVPGs that are available for import within the Org VDC.
func (vdc *Vdc) GetAllVcenterImportableDvpgs(optionalVcenterId string, queryParameters url.Values) ([]*VcenterImportableDvpg, error) {
	if vdc == nil || vdc.Vdc == nil || vdc.Vdc.ID == "" {
		return nil, fmt.Errorf("cannot get Importable DVPGs without VDC ID")
	}

	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("orgVdcId=="+vdc.Vdc.ID, queryParams)

	return getAllVcenterImportableDvpgs(vdc.client, optionalVcenterId, queryParams)

}

func getAllVcenterImportableDvpgs(client *Client, optionalVcenterId string, queryParameters url.Values) ([]*VcenterImportableDvpg, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableDvpgs
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	queryParams := copyOrNewUrlValues(queryParameters)

	// Inject vCenter ID if specified
	if optionalVcenterId != "" {
		queryParams = queryParameterFilterAnd("_context=="+optionalVcenterId, queryParams)
	}

	typeResponses := []*types.VcenterImportableDvpg{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	returnObjects := make([]*VcenterImportableDvpg, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnObjects[sliceIndex] = &VcenterImportableDvpg{
			VcenterImportableDvpg: typeResponses[sliceIndex],
			client:                client,
		}
	}

	return returnObjects, nil
}

// filterVcImportableDvpgsInExternalNetworks is created as a fix for local filtering instead of using
// FIQL filter (because it does not support it).
func filterVcImportableDvpgsInExternalNetworks(name string, allNVcImportableDvpgs []*VcenterImportableDvpg) []*VcenterImportableDvpg {
	filteredVcImportableDvpgs := make([]*VcenterImportableDvpg, 0)
	for index, VcImportableDvpg := range allNVcImportableDvpgs {
		if allNVcImportableDvpgs[index].VcenterImportableDvpg.BackingRef.Name == name {
			filteredVcImportableDvpgs = append(filteredVcImportableDvpgs, VcImportableDvpg)
		}
	}

	return filteredVcImportableDvpgs
}
