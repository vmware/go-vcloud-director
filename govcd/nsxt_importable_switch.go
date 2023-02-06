/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// NsxtImportableSwitch is a read only object to retrieve NSX-T segments (importable switches) to be used for Org VDC
// imported network.
type NsxtImportableSwitch struct {
	NsxtImportableSwitch *types.NsxtImportableSwitch
	client               *Client
}

// GetNsxtImportableSwitchByName retrieves a particular NSX-T Segment by name available for that VDC
//
// Note. OpenAPI endpoint does not exist for this resource and by default endpoint
// "/network/orgvdcnetworks/importableswitches" returns only unused NSX-T importable switches (the ones that are not
// already consumed in Org VDC networks) and there is no way to get them all (including the used ones).
func (vdc *Vdc) GetNsxtImportableSwitchByName(name string) (*NsxtImportableSwitch, error) {
	if name == "" {
		return nil, fmt.Errorf("empty NSX-T Importable Switch name specified")
	}

	allNsxtImportableSwitches, err := vdc.GetAllNsxtImportableSwitches()
	if err != nil {
		return nil, fmt.Errorf("error getting all NSX-T Importable Switches for VDC '%s': %s", vdc.Vdc.Name, err)
	}

	var filteredNsxtImportableSwitches []*NsxtImportableSwitch
	for _, nsxtImportableSwitch := range allNsxtImportableSwitches {
		if nsxtImportableSwitch.NsxtImportableSwitch.Name == name {
			filteredNsxtImportableSwitches = append(filteredNsxtImportableSwitches, nsxtImportableSwitch)
		}
	}

	if len(filteredNsxtImportableSwitches) == 0 {
		// ErrorEntityNotFound is injected here for the ability to validate problem using ContainsNotFound()
		return nil, fmt.Errorf("%s: no NSX-T Importable Switch with name '%s' for Org VDC with ID '%s' found",
			ErrorEntityNotFound, name, vdc.Vdc.ID)
	}

	if len(filteredNsxtImportableSwitches) > 1 {
		return nil, fmt.Errorf("more than one (%d) NSX-T Importable Switch with name '%s' for Org VDC with ID '%s' found",
			len(filteredNsxtImportableSwitches), name, vdc.Vdc.ID)
	}

	return filteredNsxtImportableSwitches[0], nil
}

// GetNsxtImportableSwitchByName retrieves a particular NSX-T Segment by name available for that VDC
//
// Note. OpenAPI endpoint does not exist for this resource and by default endpoint
// "/network/orgvdcnetworks/importableswitches" returns only unused NSX-T importable switches (the ones that are not
// already consumed in Org VDC networks) and there is no way to get them all (including the used ones).
func (vdcGroup *VdcGroup) GetNsxtImportableSwitchByName(name string) (*NsxtImportableSwitch, error) {
	if name == "" {
		return nil, fmt.Errorf("empty NSX-T Importable Switch name specified")
	}

	allNsxtImportableSwitches, err := vdcGroup.GetAllNsxtImportableSwitches()
	if err != nil {
		return nil, fmt.Errorf("error getting all NSX-T Importable Switches for VDC Group '%s': %s", vdcGroup.VdcGroup.Name, err)
	}

	var filteredNsxtImportableSwitches []*NsxtImportableSwitch
	for _, nsxtImportableSwitch := range allNsxtImportableSwitches {
		if nsxtImportableSwitch.NsxtImportableSwitch.Name == name {
			filteredNsxtImportableSwitches = append(filteredNsxtImportableSwitches, nsxtImportableSwitch)
		}
	}

	if len(filteredNsxtImportableSwitches) == 0 {
		// ErrorEntityNotFound is injected here for the ability to validate problem using ContainsNotFound()
		return nil, fmt.Errorf("%s: no NSX-T Importable Switch with name '%s' for VDC Group with ID '%s' found",
			ErrorEntityNotFound, name, vdcGroup.VdcGroup.Id)
	}

	if len(filteredNsxtImportableSwitches) > 1 {
		return nil, fmt.Errorf("more than one (%d) NSX-T Importable Switch with name '%s' for VDC Group with ID '%s' found",
			len(filteredNsxtImportableSwitches), name, vdcGroup.VdcGroup.Id)
	}

	return filteredNsxtImportableSwitches[0], nil
}

// GetAllNsxtImportableSwitches retrieves all available importable switches which can be consumed for creating NSX-T
// "Imported" Org VDC network
//
// Note. OpenAPI endpoint does not exist for this resource and by default endpoint
// "/network/orgvdcnetworks/importableswitches" returns only unused NSX-T importable switches (the ones that are not
// already consumed in Org VDC networks) and there is no way to get them all.
func (vdcGroup *VdcGroup) GetAllNsxtImportableSwitches() ([]*NsxtImportableSwitch, error) {
	if vdcGroup.VdcGroup.Id == "" {
		return nil, fmt.Errorf("VDC Group must have ID populated to retrieve NSX-T importable switches")
	}
	// request requires Org VDC Group ID to be specified as UUID, not as URN
	orgVdcGroupId, err := getBareEntityUuid(vdcGroup.VdcGroup.Id)
	if err != nil {
		return nil, fmt.Errorf("could not get UUID from URN '%s': %s", vdcGroup.VdcGroup.Id, err)
	}
	filter := map[string]string{"vdcGroup": orgVdcGroupId}

	return getFilteredNsxtImportableSwitches(filter, vdcGroup.client)
}

// GetAllNsxtImportableSwitches retrieves all available importable switches which can be consumed for creating NSX-T
// "Imported" Org VDC network
//
// Note. OpenAPI endpoint does not exist for this resource and by default endpoint
// "/network/orgvdcnetworks/importableswitches" returns only unused NSX-T importable switches (the ones that are not
// already consumed in Org VDC networks) and there is no way to get them all.
func (vdc *Vdc) GetAllNsxtImportableSwitches() ([]*NsxtImportableSwitch, error) {
	if vdc.Vdc.ID == "" {
		return nil, fmt.Errorf("VDC must have ID populated to retrieve NSX-T importable switches")
	}
	// request requires Org VDC ID to be specified as UUID, not as URN
	orgVdcId, err := getBareEntityUuid(vdc.Vdc.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get UUID from URN '%s': %s", vdc.Vdc.ID, err)
	}
	filter := map[string]string{"orgVdc": orgVdcId}

	return getFilteredNsxtImportableSwitches(filter, vdc.client)
}

// GetFilteredNsxtImportableSwitches returns all available importable switches.
// One of the filters below is required (using plain UUID - not URN):
// * orgVdc
// * nsxTManager (only in VCD 10.3.0+)
//
// Note. OpenAPI endpoint does not exist for this resource and by default endpoint
// "/network/orgvdcnetworks/importableswitches" returns only unused NSX-T importable switches (the ones that are not
// already consumed in Org VDC networks) and there is no way to get them all.
func (vcdClient *VCDClient) GetFilteredNsxtImportableSwitches(filter map[string]string) ([]*NsxtImportableSwitch, error) {
	return getFilteredNsxtImportableSwitches(filter, &vcdClient.Client)
}

// GetFilteredNsxtImportableSwitchesByName builds on top of GetFilteredNsxtImportableSwitches and additionally performs
// client side filtering by Name
func (vcdClient *VCDClient) GetFilteredNsxtImportableSwitchesByName(filter map[string]string, name string) (*NsxtImportableSwitch, error) {
	importableSwitches, err := getFilteredNsxtImportableSwitches(filter, &vcdClient.Client)
	if err != nil {
		return nil, fmt.Errorf("error getting list of filtered Importable Switches: %s", err)
	}

	var foundImportableSwitch bool
	var foundSwitches []*NsxtImportableSwitch

	for index, impSwitch := range importableSwitches {
		if importableSwitches[index].NsxtImportableSwitch.Name == name {
			foundImportableSwitch = true
			foundSwitches = append(foundSwitches, impSwitch)
		}
	}

	if !foundImportableSwitch {
		return nil, fmt.Errorf("%s: Importable Switch with name '%s' not found", ErrorEntityNotFound, name)
	}

	if len(foundSwitches) > 1 {
		return nil, fmt.Errorf("found multiple Importable Switches with name '%s'", name)
	}

	return foundSwitches[0], nil
}

// getFilteredNsxtImportableSwitches is extracted so that it can be reused across multiple functions
func getFilteredNsxtImportableSwitches(filter map[string]string, client *Client) ([]*NsxtImportableSwitch, error) {
	apiEndpoint := client.VCDHREF
	endpoint := apiEndpoint.Scheme + "://" + apiEndpoint.Host + "/network/orgvdcnetworks/importableswitches/"
	// error below is ignored because it is a static endpoint
	urlRef, err := url.Parse(endpoint)
	if err != nil {
		util.Logger.Printf("[DEBUG - getFilteredNsxtImportableSwitches] error parsing URL: %s", err)
	}

	headAccept := http.Header{}
	headAccept.Set("Accept", types.JSONMime)
	request := client.newRequest(filter, nil, http.MethodGet, *urlRef, nil, client.APIVersion, headAccept)
	request.Header.Set("Accept", types.JSONMime)

	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Logger.Printf("error closing response Body [getFilteredNsxtImportableSwitches]: %s", err)
		}
	}(response.Body)

	var nsxtImportableSwitches []*types.NsxtImportableSwitch
	if err = decodeBody(types.BodyTypeJSON, response, &nsxtImportableSwitches); err != nil {
		return nil, err
	}

	wrappedNsxtImportableSwitches := make([]*NsxtImportableSwitch, len(nsxtImportableSwitches))
	for sliceIndex := range nsxtImportableSwitches {
		wrappedNsxtImportableSwitches[sliceIndex] = &NsxtImportableSwitch{
			NsxtImportableSwitch: nsxtImportableSwitches[sliceIndex],
			client:               client,
		}
	}

	return wrappedNsxtImportableSwitches, nil
}
