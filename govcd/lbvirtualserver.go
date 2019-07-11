/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBVirtualServer creates a load balancer virtual server based on mandatory fields. It is a
// synchronous operation. It returns created object with all fields (including ID) populated
// or an error.
// Name, Protocol, Port and IpAddress fields must be populated
func (eGW *EdgeGateway) CreateLBVirtualServer(lbVirtualServerConfig *types.LBVirtualServer) (*types.LBVirtualServer, error) {
	if err := validateCreateLBVirtualServer(lbVirtualServerConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBVirtualServerPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer virtual server: %s", lbVirtualServerConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// Location: [/network/edges/edge-3/loadbalancer/config/virtualservers/virtualServer-10]
	lbPoolID, err := extractNSXObjectIDfromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readPool, err := eGW.ReadLBVirtualServerByID(lbPoolID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve load balancer virtual server with ID (%s) after creation: %s",
			lbPoolID, err)
	}
	return readPool, nil
}

// ReadLBVirtualServer is able to find the types.LBVirtualServer type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) ReadLBVirtualServer(lbVirtualServerConfig *types.LBVirtualServer) (*types.LBVirtualServer, error) {
	if err := validateReadLBVirtualServer(lbVirtualServerConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBVirtualServerPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "virtual server response"
	lbVirtualServerResponse := &struct {
		LBPools []*types.LBVirtualServer `xml:"virtualServer"`
	}{}

	// This query returns all virtual servers as the API does not have filtering options
	_, err = eGW.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load balancer virtual server: %s", nil, lbVirtualServerResponse)
	if err != nil {
		return nil, err
	}

	// Search for virtual server by ID or by Name
	for _, virtualServer := range lbVirtualServerResponse.LBPools {
		// If ID was specified for lookup - look for the same ID
		if lbVirtualServerConfig.ID != "" && virtualServer.ID == lbVirtualServerConfig.ID {
			return virtualServer, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbVirtualServerConfig.Name != "" && virtualServer.Name == lbVirtualServerConfig.Name {
			// We found it by name. Let's verify if search ID was specified and it matches the lookup object
			if lbVirtualServerConfig.ID != "" && virtualServer.ID != lbVirtualServerConfig.ID {
				return nil, fmt.Errorf("load balancer virtual server was found by name (%s), but it's ID (%s) does not match specified ID (%s)",
					virtualServer.Name, virtualServer.ID, lbVirtualServerConfig.ID)
			}
			return virtualServer, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// ReadLBVirtualServerByName wraps ReadLBVirtualServer and needs only an ID for lookup
func (eGW *EdgeGateway) ReadLBVirtualServerByID(id string) (*types.LBVirtualServer, error) {
	return eGW.ReadLBVirtualServer(&types.LBVirtualServer{ID: id})
}

// ReadLBVirtualServerByName wraps ReadLBVirtualServer and needs only a Name for lookup
func (eGW *EdgeGateway) ReadLBVirtualServerByName(name string) (*types.LBVirtualServer, error) {
	return eGW.ReadLBVirtualServer(&types.LBVirtualServer{Name: name})
}

// UpdateLBVirtualServer updates types.LBVirtualServer with all fields. At least name or ID must be specified.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
// Name, Protocol, Port and IpAddress fields must be populated
func (eGW *EdgeGateway) UpdateLBVirtualServer(lbVirtualServerConfig *types.LBVirtualServer) (*types.LBVirtualServer, error) {
	err := validateUpdateLBVirtualServer(lbVirtualServerConfig)
	if err != nil {
		return nil, err
	}

	lbVirtualServerConfig.ID, err = eGW.getLBVirtualServerIDByNameID(lbVirtualServerConfig.Name, lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer virtual server: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBVirtualServerPath + lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer virtual server : %s", lbVirtualServerConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readPool, err := eGW.ReadLBVirtualServerByID(lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve virtual server with ID (%s) after update: %s", lbVirtualServerConfig.ID, err)
	}
	return readPool, nil
}

// DeleteLBVirtualServer is able to delete the types.LBVirtualServer type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) DeleteLBVirtualServer(lbVirtualServerConfig *types.LBVirtualServer) error {
	err := validateDeleteLBVirtualServer(lbVirtualServerConfig)
	if err != nil {
		return err
	}

	lbVirtualServerConfig.ID, err = eGW.getLBVirtualServerIDByNameID(lbVirtualServerConfig.Name, lbVirtualServerConfig.ID)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer virtual server: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBVirtualServerPath + lbVirtualServerConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	return eGW.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete load balancer virtual server: %s", nil)
}

// DeleteLBVirtualServerByID wraps DeleteLBVirtualServer and requires only ID for deletion
func (eGW *EdgeGateway) DeleteLBVirtualServerByID(id string) error {
	return eGW.DeleteLBVirtualServer(&types.LBVirtualServer{ID: id})
}

// DeleteLBVirtualServerByName wraps DeleteLBVirtualServer and requires only Name for deletion
func (eGW *EdgeGateway) DeleteLBVirtualServerByName(name string) error {
	return eGW.DeleteLBVirtualServer(&types.LBVirtualServer{Name: name})
}

func validateCreateLBVirtualServer(lbVirtualServerConfig *types.LBVirtualServer) error {
	if lbVirtualServerConfig.Name == "" {
		return fmt.Errorf("load balancer virtual server Name cannot be empty")
	}

	if lbVirtualServerConfig.IpAddress == "" {
		return fmt.Errorf("load balancer virtual server IpAddress cannot be empty")
	}

	if lbVirtualServerConfig.Protocol == "" {
		return fmt.Errorf("load balancer virtual server Protocol cannot be empty")
	}

	if lbVirtualServerConfig.Port == nil {
		return fmt.Errorf("load balancer virtual server Port cannot be empty")
	}

	return nil
}

func validateReadLBVirtualServer(lbVirtualServerConfig *types.LBVirtualServer) error {
	if lbVirtualServerConfig.ID == "" && lbVirtualServerConfig.Name == "" {
		return fmt.Errorf("to read load balancer virtual server at least one of `ID`, `Name` " +
			"fields must be specified")
	}

	return nil
}

func validateUpdateLBVirtualServer(lbVirtualServerConfig *types.LBVirtualServer) error {
	// Update and create have the same requirements for now
	return validateCreateLBVirtualServer(lbVirtualServerConfig)
}

func validateDeleteLBVirtualServer(lbVirtualServerConfig *types.LBVirtualServer) error {
	// Read and delete have the same requirements for now
	return validateReadLBVirtualServer(lbVirtualServerConfig)
}

// getLBVirtualServerIDByNameID checks if at least name or ID is set and returns the ID.
// If the ID is specified - it passes through the ID. If only name was specified
// it will lookup the object by name and return the ID.
func (eGW *EdgeGateway) getLBVirtualServerIDByNameID(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or ID must be specific to find load balancer "+
			"virtual server got name (%s) ID (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, ID must be found, because only ID can be used in request path
	readLbVirtualServer, err := eGW.ReadLBVirtualServerByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer virtual server by name: %s", err)
	}
	return readLbVirtualServer.ID, nil
}
