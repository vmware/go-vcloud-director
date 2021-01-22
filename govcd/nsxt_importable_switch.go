/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
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

	apiEndpoint := vdc.client.VCDHREF
	endpoint := apiEndpoint.Scheme + "://" + apiEndpoint.Host + "/network/orgvdcnetworks/importableswitches"
	// error below is ignored because it is a static endpoint
	urlRef, _ := url.Parse(endpoint)

	// request requires Org VDC ID to be specified as UUID, not as URN
	orgVdcId, err := getBareEntityUuid(vdc.Vdc.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get UUID from URN '%s': %s", vdc.Vdc.ID, err)
	}

	headAccept := http.Header{}
	headAccept.Set("Accept", types.JSONMime)
	request := vdc.client.newRequest(map[string]string{"orgVdc": orgVdcId}, nil, http.MethodGet, *urlRef, nil, vdc.client.APIVersion, headAccept)
	request.Header.Set("Accept", types.JSONMime)

	response, err := checkResp(vdc.client.Http.Do(request))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	nsxtImportableSwitches := []*types.NsxtImportableSwitch{}
	if err = decodeBody(types.BodyTypeJSON, response, &nsxtImportableSwitches); err != nil {
		return nil, err
	}

	wrappedNsxtImportableSwitches := make([]*NsxtImportableSwitch, len(nsxtImportableSwitches))
	for sliceIndex := range nsxtImportableSwitches {
		wrappedNsxtImportableSwitches[sliceIndex] = &NsxtImportableSwitch{
			NsxtImportableSwitch: nsxtImportableSwitches[sliceIndex],
			client:               vdc.client,
		}
	}

	return wrappedNsxtImportableSwitches, nil
}
