/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxvDistributedFirewall defines a distributed firewall for a NSX-V VDC
type NsxvDistributedFirewall struct {
	VdcId         string                       // The ID of the VDC
	Configuration *types.FirewallConfiguration // The latest firewall configuration
	Services      []types.Application          // The list of services for this VDC
	enabled       bool                         // internal flag that signifies whether the firewall is enabled
	client        *Client                      // internal usage client
}

// NewNsxvDistributedFirewall creates a new NsxvDistributedFirewall
func NewNsxvDistributedFirewall(client *Client, vdcId string) *NsxvDistributedFirewall {
	return &NsxvDistributedFirewall{
		client: client,
		VdcId:  extractUuid(vdcId),
	}
}

// GetConfiguration retrieves the configuration of a distributed firewall
func (dfw *NsxvDistributedFirewall) GetConfiguration() (*types.FirewallConfiguration, error) {
	initialUrl, err := dfw.client.buildUrl("network", "firewall", "globalroot-0", "config")
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"vdc": dfw.VdcId,
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return nil, err
	}

	req := dfw.client.NewRequest(params, http.MethodGet, *requestUrl, nil)

	resp, err := checkResp(dfw.client.Http.Do(req))
	if err != nil {
		return nil, err
	}
	var config types.FirewallConfiguration
	err = decodeBody(types.BodyTypeXML, resp, &config)
	if err != nil {
		return nil, err
	}
	dfw.Configuration = &config
	dfw.enabled = true
	return &config, nil
}

// IsEnabled returns true when the distributed firewall is enabled
func (dfw *NsxvDistributedFirewall) IsEnabled() (bool, error) {
	if dfw.VdcId == "" {
		return false, fmt.Errorf("no VDC set for this NsxvDistributedFirewall")
	}

	if dfw.Configuration != nil {
		return dfw.enabled, nil
	}
	_, err := dfw.GetConfiguration()
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Enable makes the distributed firewall available
// It requires system administrator privileges
// It fails with a non-NSX-V VDC
func (dfw *NsxvDistributedFirewall) Enable() error {
	if !dfw.client.IsSysAdmin {
		return fmt.Errorf("method 'Enable' requires system administrator privileges")
	}
	dfw.enabled = false
	if dfw.VdcId == "" {
		return fmt.Errorf("no AdminVdc set for this NsxvDistributedFirewall")
	}
	initialUrl, err := dfw.client.buildUrl("network", "firewall", "vdc", extractUuid(dfw.VdcId))
	if err != nil {
		return err
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return err
	}

	req := dfw.client.NewRequest(nil, http.MethodPost, *requestUrl, nil)

	resp, err := checkResp(dfw.client.Http.Do(req))
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("[enableDistributedFirewall] expected status code %d - received %d", http.StatusCreated, resp.StatusCode)
	}
	dfw.enabled = true
	return nil
}

// Disable removes the availability of a distributed firewall
// It requires system administrator privileges
// WARNING: it also removes all rules
func (dfw *NsxvDistributedFirewall) Disable() error {
	if dfw.VdcId == "" {
		return fmt.Errorf("no AdminVdc set for this NsxvDistributedFirewall")
	}
	initialUrl, err := dfw.client.buildUrl("network", "firewall", "vdc", extractUuid(dfw.VdcId))
	if err != nil {
		return err
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return err
	}

	req := dfw.client.NewRequest(nil, http.MethodDelete, *requestUrl, nil)

	resp, err := checkResp(dfw.client.Http.Do(req))
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("[disableDistributedFirewall] expected status code %d - received %d", http.StatusNoContent, resp.StatusCode)
	}
	dfw.Configuration = nil
	dfw.Services = nil
	dfw.enabled = false
	return nil
}

// GetServices retrieves the list of services for the current VCD
// If `refresh` = false and the services were already retrieved in a previous operation,
// then it returns the internal values instead of fetching new ones
func (dfw *NsxvDistributedFirewall) GetServices(refresh bool) ([]types.Application, error) {
	if dfw.Services != nil && !refresh {
		return dfw.Services, nil
	}
	if dfw.VdcId == "" {
		return nil, fmt.Errorf("no AdminVdc set for this NsxvDistributedFirewall")
	}
	initialUrl, err := dfw.client.buildUrl("network", "services", "application", "scope", extractUuid(dfw.VdcId))
	if err != nil {
		return nil, err
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return nil, err
	}

	req := dfw.client.NewRequest(nil, http.MethodGet, *requestUrl, nil)

	resp, err := checkResp(dfw.client.Http.Do(req))
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching the services: %s", resp.Status)
	}
	var applicationList types.ApplicationList
	err = decodeBody(types.BodyTypeXML, resp, &applicationList)
	if err != nil {
		return nil, err
	}
	dfw.Services = applicationList.Application
	return applicationList.Application, nil
}

// Refresh retrieves fresh values for the distributed firewall rules and services
func (dfw *NsxvDistributedFirewall) Refresh() error {
	if dfw.VdcId == "" {
		return fmt.Errorf("no AdminVdc set for this NsxvDistributedFirewall")
	}

	_, err := dfw.GetServices(true)
	if err != nil {
		return err
	}
	_, err = dfw.GetConfiguration()
	return err
}

// GetServiceById retrieves a single service, identified by its ID, for the current VDC
// If the list of services was already retrieved, it uses it, otherwise fetches new ones.
// Returns ErrorEntityNotFound when the requested services was not found
func (dfw *NsxvDistributedFirewall) GetServiceById(serviceId string) (*types.Application, error) {
	services, err := dfw.GetServices(false)
	if err != nil {
		return nil, err
	}
	for _, app := range services {
		if app.ObjectID == serviceId {
			return &app, nil
		}
	}
	return nil, ErrorEntityNotFound
}

// GetServiceByName retrieves a single service, identified by its name, for the current VDC
// If the list of services was already retrieved, it uses it, otherwise fetches new ones.
// Returns ErrorEntityNotFound when the requested services was not found
func (dfw *NsxvDistributedFirewall) GetServiceByName(serviceName string) (*types.Application, error) {
	services, err := dfw.GetServices(false)
	if err != nil {
		return nil, err
	}
	for _, app := range services {
		if app.Name == serviceName {
			return &app, nil
		}
	}
	return nil, ErrorEntityNotFound
}

// GetServicesByRegex returns a list of services with their names matching the given regular expression
// It may return an empty list (without error)
func (dfw *NsxvDistributedFirewall) GetServicesByRegex(expression string) ([]types.Application, error) {
	services, err := dfw.GetServices(false)
	if err != nil {
		return nil, err
	}
	searchRegex := regexp.MustCompile(expression)
	var found []types.Application
	for _, app := range services {
		if searchRegex.MatchString(app.Name) {
			found = append(found, app)
		}
	}
	return found, nil
}

// buildUrl uses the Client base URL to create a customised URL
func (client *Client) buildUrl(elements ...string) (string, error) {
	baseUrl := client.VCDHREF.String()
	if !IsValidUrl(baseUrl) {
		return "", fmt.Errorf("incorrect URL %s", client.VCDHREF.String())
	}
	if strings.HasSuffix(baseUrl, "/") {
		baseUrl = strings.TrimRight(baseUrl, "/")
	}
	if strings.HasSuffix(baseUrl, "/api") {
		baseUrl = strings.TrimRight(baseUrl, "/api")
	}
	return url.JoinPath(baseUrl, elements...)
}
