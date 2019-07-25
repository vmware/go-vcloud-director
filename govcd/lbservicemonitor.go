/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBServiceMonitor creates a load balancer service monitor based on mandatory fields. It is a synchronous
// operation. It returns created object with all fields (including Id) populated or an error.
func (egw *EdgeGateway) CreateLBServiceMonitor(lbMonitorConfig *types.LbMonitor) (*types.LbMonitor, error) {
	if err := validateCreateLBServiceMonitor(lbMonitorConfig); err != nil {
		return nil, err
	}

	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("edge gateway does not have advanced networking enabled")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbMonitorPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer service monitor: %s", lbMonitorConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// Location: [/network/edges/edge-3/loadbalancer/config/monitors/monitor-5]
	lbMonitorID, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readMonitor, err := egw.ReadLBServiceMonitorByID(lbMonitorID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve monitor with Id (%s) after creation: %s", lbMonitorID, err)
	}
	return readMonitor, nil
}

// ReadLBServiceMonitor is able to find the types.LbMonitor type by Name and/or Id.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) ReadLBServiceMonitor(lbMonitorConfig *types.LbMonitor) (*types.LbMonitor, error) {
	if err := validateReadLBServiceMonitor(lbMonitorConfig); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbMonitorPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "monitor response"
	lbMonitorResponse := &struct {
		LBMonitors []*types.LbMonitor `xml:"monitor"`
	}{}

	// This query returns all service monitors as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime, "unable to read Load Balancer monitor: %s", nil, lbMonitorResponse)
	if err != nil {
		return nil, err
	}

	// Search for monitor by Id or by Name
	for _, monitor := range lbMonitorResponse.LBMonitors {
		// If Id was specified for lookup - look for the same Id
		if lbMonitorConfig.Id != "" && monitor.Id == lbMonitorConfig.Id {
			return monitor, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbMonitorConfig.Name != "" && monitor.Name == lbMonitorConfig.Name {
			// We found it by name. Let's verify if search Id was specified and it matches the lookup object
			if lbMonitorConfig.Id != "" && monitor.Id != lbMonitorConfig.Id {
				return nil, fmt.Errorf("load balancer monitor was found by name (%s), but it's Id (%s) does not match specified Id (%s)",
					monitor.Name, monitor.Id, lbMonitorConfig.Id)
			}
			return monitor, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// ReadLBServiceMonitorByID wraps ReadLBServiceMonitor and needs only an Id for lookup
func (egw *EdgeGateway) ReadLBServiceMonitorByID(id string) (*types.LbMonitor, error) {
	return egw.ReadLBServiceMonitor(&types.LbMonitor{Id: id})
}

// ReadLBServiceMonitorByName wraps ReadLBServiceMonitor and needs only a Name for lookup
func (egw *EdgeGateway) ReadLBServiceMonitorByName(name string) (*types.LbMonitor, error) {
	return egw.ReadLBServiceMonitor(&types.LbMonitor{Name: name})
}

// UpdateLBServiceMonitor updates types.LbMonitor with all fields. At least name or Id must be specified.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) UpdateLBServiceMonitor(lbMonitorConfig *types.LbMonitor) (*types.LbMonitor, error) {
	err := validateUpdateLBServiceMonitor(lbMonitorConfig)
	if err != nil {
		return nil, err
	}

	lbMonitorConfig.Id, err = egw.getLBServiceMonitorIDByNameID(lbMonitorConfig.Name, lbMonitorConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer service monitor: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbMonitorPath + lbMonitorConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer service monitor : %s", lbMonitorConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readMonitor, err := egw.ReadLBServiceMonitorByID(lbMonitorConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve monitor with Id (%s) after update: %s", lbMonitorConfig.Id, err)
	}
	return readMonitor, nil
}

// DeleteLBServiceMonitor is able to delete the types.LbMonitor type by Name and/or Id.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) DeleteLBServiceMonitor(lbMonitorConfig *types.LbMonitor) error {
	err := validateDeleteLBServiceMonitor(lbMonitorConfig)
	if err != nil {
		return err
	}

	lbMonitorConfig.Id, err = egw.getLBServiceMonitorIDByNameID(lbMonitorConfig.Name, lbMonitorConfig.Id)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer service monitor: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbMonitorPath + lbMonitorConfig.Id)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	return egw.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete Service Monitor: %s", nil)
}

// DeleteLBServiceMonitorByID wraps DeleteLBServiceMonitor and requires only Id for deletion
func (egw *EdgeGateway) DeleteLBServiceMonitorByID(id string) error {
	return egw.DeleteLBServiceMonitor(&types.LbMonitor{Id: id})
}

// DeleteLBServiceMonitorByName wraps DeleteLBServiceMonitor and requires only Name for deletion
func (egw *EdgeGateway) DeleteLBServiceMonitorByName(name string) error {
	return egw.DeleteLBServiceMonitor(&types.LbMonitor{Name: name})
}

func validateCreateLBServiceMonitor(lbMonitorConfig *types.LbMonitor) error {
	if lbMonitorConfig.Name == "" {
		return fmt.Errorf("load balancer monitor Name cannot be empty")
	}

	if lbMonitorConfig.Timeout == 0 {
		return fmt.Errorf("load balancer monitor Timeout cannot be 0")
	}

	if lbMonitorConfig.Interval == 0 {
		return fmt.Errorf("load balancer monitor Interval cannot be 0")
	}

	if lbMonitorConfig.MaxRetries == 0 {
		return fmt.Errorf("load balancer monitor MaxRetries cannot be 0")
	}

	if lbMonitorConfig.Type == "" {
		return fmt.Errorf("load balancer monitor Type cannot be empty")
	}

	return nil
}

func validateReadLBServiceMonitor(lbMonitorConfig *types.LbMonitor) error {
	if lbMonitorConfig.Id == "" && lbMonitorConfig.Name == "" {
		return fmt.Errorf("to read load balancer service monitor at least one of `Id`, `Name` fields must be specified")
	}

	return nil
}

func validateUpdateLBServiceMonitor(lbMonitorConfig *types.LbMonitor) error {
	// Update and create have the same requirements for now
	return validateCreateLBServiceMonitor(lbMonitorConfig)
}

func validateDeleteLBServiceMonitor(lbMonitorConfig *types.LbMonitor) error {
	// Read and delete have the same requirements for now
	return validateReadLBServiceMonitor(lbMonitorConfig)
}

// getLBServiceMonitorIDByNameID checks if at least name or Id is set and returns the Id.
// If the Id is specified - it passes through the Id. If only name was specified
// it will lookup the object by name and return the Id.
func (egw *EdgeGateway) getLBServiceMonitorIDByNameID(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or Id must be specific to find load balancer "+
			"service monitor got name (%s) Id (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, Id must be found, because only Id can be used in request path
	readlbServiceMonitor, err := egw.ReadLBServiceMonitorByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer service monitor by name: %s", err)
	}
	return readlbServiceMonitor.Id, nil
}
