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
//
// Note. API returns only unused DVPGs. If the DVPG is already consumed - it will not be returned.
type VcenterImportableDvpg struct {
	VcenterImportableDvpg *types.VcenterImportableDvpg
	client                *Client
}

// GetVcenterImportableDvpgByName retrieves a DVPG by name
//
// Note. API returns only unused DVPGs. If the DVPG is already consumed - it will not be returned.
func (vcdClient *VCDClient) GetVcenterImportableDvpgByName(name string) (*VcenterImportableDvpg, error) {
	if name == "" {
		return nil, fmt.Errorf("empty importable Distributed Virtual Port Group Name specified")
	}

	vcImportableDvpgs, err := vcdClient.GetAllVcenterImportableDvpgs(nil)
	if err != nil {
		return nil, fmt.Errorf("could not find Distributed Virtual Port Group with Name '%s' for vCenter with ID '%s': %s",
			name, "", err)
	}

	filteredVcImportableDvpgs := filterVcImportableDvpgsByName(name, vcImportableDvpgs)

	return oneOrError("name", name, filteredVcImportableDvpgs)
}

// GetAllVcenterImportableDvpgs retrieves all DVPGs that are available for import.
//
// Note. API returns only unused DVPGs. If the DVPG is already consumed - it will not be returned.
func (vcdClient *VCDClient) GetAllVcenterImportableDvpgs(queryParameters url.Values) ([]*VcenterImportableDvpg, error) {
	return getAllVcenterImportableDvpgs(&vcdClient.Client, queryParameters)
}

// GetVcenterImportableDvpgByName retrieves a DVPG that is available for import within the Org VDC.
func (vdc *Vdc) GetVcenterImportableDvpgByName(name string) (*VcenterImportableDvpg, error) {
	if name == "" {
		return nil, fmt.Errorf("empty importable Distributed Virtual Port Group Name specified")
	}

	vcImportableDvpgs, err := vdc.GetAllVcenterImportableDvpgs(nil)
	if err != nil {
		return nil, fmt.Errorf("could not find Distributed Virtual Port Group with name '%s': %s", name, err)
	}

	filteredVcImportableDvpgs := filterVcImportableDvpgsByName(name, vcImportableDvpgs)

	return oneOrError("name", name, filteredVcImportableDvpgs)
}

// GetAllVcenterImportableDvpgs retrieves all DVPGs that are available for import within the Org VDC.
//
// Note. API returns only unused DVPGs. If the DVPG is already consumed - it will not be returned.
func (vdc *Vdc) GetAllVcenterImportableDvpgs(queryParameters url.Values) ([]*VcenterImportableDvpg, error) {
	if vdc == nil || vdc.Vdc == nil || vdc.Vdc.ID == "" {
		return nil, fmt.Errorf("cannot get Importable DVPGs without VDC ID")
	}

	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("orgVdcId=="+vdc.Vdc.ID, queryParams)

	return getAllVcenterImportableDvpgs(vdc.client, queryParams)

}

func getAllVcenterImportableDvpgs(client *Client, queryParameters url.Values) ([]*VcenterImportableDvpg, error) {
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

// filterVcImportableDvpgsByName is created as a fix for local filtering instead of using
// FIQL filter (because it does not support it).
func filterVcImportableDvpgsByName(name string, allNVcImportableDvpgs []*VcenterImportableDvpg) []*VcenterImportableDvpg {
	filteredVcImportableDvpgs := make([]*VcenterImportableDvpg, 0)
	for _, VcImportableDvpg := range allNVcImportableDvpgs {
		if VcImportableDvpg.VcenterImportableDvpg.BackingRef.Name == name {
			filteredVcImportableDvpgs = append(filteredVcImportableDvpgs, VcImportableDvpg)
		}
	}

	return filteredVcImportableDvpgs
}
