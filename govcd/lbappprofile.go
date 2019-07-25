/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBAppProfile creates a load balancer application profile based on mandatory fields. It is a
// synchronous operation. It returns created object with all fields (including Id) populated or an error.
func (eGW *EdgeGateway) CreateLBAppProfile(lbAppProfileConfig *types.LbAppProfile) (*types.LbAppProfile, error) {
	if err := validateCreateLBAppProfile(lbAppProfileConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbAppProfilePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer application profile: %s", lbAppProfileConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// [/network/edges/edge-3/loadbalancer/config/applicationprofiles/applicationProfile-4]
	lbAppProfileID, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readAppProfile, err := eGW.ReadLBAppProfileByID(lbAppProfileID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application profile with Id (%s) after creation: %s",
			readAppProfile.Id, err)
	}
	return readAppProfile, nil
}

// ReadLBAppProfile is able to find the types.LbAppProfile type by Name and/or Id.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) ReadLBAppProfile(lbAppProfileConfig *types.LbAppProfile) (*types.LbAppProfile, error) {
	if err := validateReadLBAppProfile(lbAppProfileConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbAppProfilePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap response
	lbAppProfileResponse := &struct {
		LbAppProfiles []*types.LbAppProfile `xml:"applicationProfile"`
	}{}

	// This query returns all application profiles as the API does not have filtering options
	_, err = eGW.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load balancer application profile: %s", nil, lbAppProfileResponse)
	if err != nil {
		return nil, err
	}

	// Search for application profile by Id or by Name
	for _, profile := range lbAppProfileResponse.LbAppProfiles {
		// If Id was specified for lookup - look for the same Id
		if lbAppProfileConfig.Id != "" && profile.Id == lbAppProfileConfig.Id {
			return profile, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbAppProfileConfig.Name != "" && profile.Name == lbAppProfileConfig.Name {
			// We found it by name. Let's verify if search Id was specified and it matches the lookup object
			if lbAppProfileConfig.Id != "" && profile.Id != lbAppProfileConfig.Id {
				return nil, fmt.Errorf("load balancer application profile was found by name (%s)"+
					", but its Id (%s) does not match specified Id (%s)",
					profile.Name, profile.Id, lbAppProfileConfig.Id)
			}
			return profile, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// ReadLBAppProfileByID wraps ReadLBAppProfile and needs only an Id for lookup
func (eGW *EdgeGateway) ReadLBAppProfileByID(id string) (*types.LbAppProfile, error) {
	return eGW.ReadLBAppProfile(&types.LbAppProfile{Id: id})
}

// ReadLBAppProfileByName wraps ReadLBAppProfile and needs only a Name for lookup
func (eGW *EdgeGateway) ReadLBAppProfileByName(name string) (*types.LbAppProfile, error) {
	return eGW.ReadLBAppProfile(&types.LbAppProfile{Name: name})
}

// UpdateLBAppProfile updates types.LbAppProfile with all fields. At least name or Id must be specified.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) UpdateLBAppProfile(lbAppProfileConfig *types.LbAppProfile) (*types.LbAppProfile, error) {
	err := validateUpdateLBAppProfile(lbAppProfileConfig)
	if err != nil {
		return nil, err
	}

	lbAppProfileConfig.Id, err = eGW.getLBAppProfileIDByNameID(lbAppProfileConfig.Name, lbAppProfileConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer application profile: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbAppProfilePath + lbAppProfileConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer application profile : %s", lbAppProfileConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readAppProfile, err := eGW.ReadLBAppProfileByID(lbAppProfileConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application profile with Id (%s) after update: %s",
			readAppProfile.Id, err)
	}
	return readAppProfile, nil
}

// DeleteLBAppProfile is able to delete the types.LbAppProfile type by Name and/or Id.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) DeleteLBAppProfile(lbAppProfileConfig *types.LbAppProfile) error {
	err := validateDeleteLBAppProfile(lbAppProfileConfig)
	if err != nil {
		return err
	}

	lbAppProfileConfig.Id, err = eGW.getLBAppProfileIDByNameID(lbAppProfileConfig.Name, lbAppProfileConfig.Id)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer application profile: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbAppProfilePath + lbAppProfileConfig.Id)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	return eGW.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, "application/xml",
		"unable to delete application profile: %s", nil)
}

// DeleteLBAppProfileByID wraps DeleteLBAppProfile and requires only Id for deletion
func (eGW *EdgeGateway) DeleteLBAppProfileByID(id string) error {
	return eGW.DeleteLBAppProfile(&types.LbAppProfile{Id: id})
}

// DeleteLBAppProfileByName wraps DeleteLBAppProfile and requires only Name for deletion
func (eGW *EdgeGateway) DeleteLBAppProfileByName(name string) error {
	return eGW.DeleteLBAppProfile(&types.LbAppProfile{Name: name})
}

func validateCreateLBAppProfile(lbAppProfileConfig *types.LbAppProfile) error {
	if lbAppProfileConfig.Name == "" {
		return fmt.Errorf("load balancer application profile Name cannot be empty")
	}

	return nil
}

func validateReadLBAppProfile(lbAppProfileConfig *types.LbAppProfile) error {
	if lbAppProfileConfig.Id == "" && lbAppProfileConfig.Name == "" {
		return fmt.Errorf("to read load balancer application profile at least one of `Id`, `Name`" +
			" fields must be specified")
	}

	return nil
}

func validateUpdateLBAppProfile(lbAppProfileConfig *types.LbAppProfile) error {
	// Update and create have the same requirements for now
	return validateCreateLBAppProfile(lbAppProfileConfig)
}

func validateDeleteLBAppProfile(lbAppProfileConfig *types.LbAppProfile) error {
	// Read and delete have the same requirements for now
	return validateReadLBAppProfile(lbAppProfileConfig)
}

// getLBAppProfileIDByNameID checks if at least name or Id is set and returns the Id.
// If the Id is specified - it passes through the Id. If only name was specified
// it will lookup the object by name and return the Id.
func (eGW *EdgeGateway) getLBAppProfileIDByNameID(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or Id must be specific to find load balancer "+
			"application profile got name (%s) Id (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, Id must be found, because only Id can be used in request path
	readlbAppProfile, err := eGW.ReadLBAppProfileByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer application profile by name: %s", err)
	}
	return readlbAppProfile.Id, nil
}
