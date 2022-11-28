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
	Etag          string
	enabled       bool    // internal flag that signifies whether the firewall is enabled
	client        *Client // internal usage client

	// Attributes below to be removed if we decide that we don't need services
	Services      []types.Application      // The list of services for this VDC
	ServiceGroups []types.ApplicationGroup // The list of service groups for this VDC
}

const (
	ProtocolTcp  = "TCP"
	ProtocolUdp  = "UDP"
	ProtocolIcmp = "ICMP"
)

var NsxvProtocolCodes = map[string]int{
	ProtocolTcp:  6,
	ProtocolUdp:  17,
	ProtocolIcmp: 1,
}

// NewNsxvDistributedFirewall creates a new NsxvDistributedFirewall
func NewNsxvDistributedFirewall(client *Client, vdcId string) *NsxvDistributedFirewall {
	return &NsxvDistributedFirewall{
		client: client,
		VdcId:  extractUuid(vdcId),
	}
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

// GetConfiguration retrieves the configuration of a distributed firewall
func (dfw *NsxvDistributedFirewall) GetConfiguration() (*types.FirewallConfiguration, error) {
	// Explicitly retrieving only the Layer 3 rules, as we don't need to deal with layer 2
	initialUrl, err := dfw.client.buildUrl("network", "firewall", "globalroot-0", "config", "layer3sections", dfw.VdcId)
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
	var config types.FirewallConfiguration
	err = decodeBody(types.BodyTypeXML, resp, &config.Layer3Sections)
	if err != nil {
		return nil, err
	}
	dfw.Etag = resp.Header.Get("etag")
	// The ETag header is needed for further operations. Rules insertion and update need to have a
	// header "If-Match" with the contents of the ETag from a previous read.
	// The same data can be found in the "GenerationNumber" within the section to update.
	// The value of the ETag changes at every GET
	if dfw.Etag == "" && config.Layer3Sections.Section.GenerationNumber != "" {
		dfw.Etag = config.Layer3Sections.Section.GenerationNumber
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
		return fmt.Errorf("[enable DistributedFirewall] expected status code %d - received %d", http.StatusCreated, resp.StatusCode)
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
		return fmt.Errorf("[disable DistributedFirewall] expected status code %d - received %d", http.StatusNoContent, resp.StatusCode)
	}
	dfw.Configuration = nil
	dfw.Services = nil
	dfw.ServiceGroups = nil
	dfw.enabled = false
	return nil
}

func (dfw *NsxvDistributedFirewall) UpdateConfiguration(rules []types.NsxvDistributedFirewallRule) (*types.FirewallConfiguration, error) {

	if dfw.Etag == "" {
		_, err := dfw.GetConfiguration()
		if err != nil {
			return nil, err
		}
		if dfw.Etag == "" {
			return nil, fmt.Errorf("error getting ETag from distributed firewall")
		}
	}
	initialUrl, err := dfw.client.buildUrl("network", "firewall", "globalroot-0", "config", "layer3sections", dfw.VdcId)
	if err != nil {
		return nil, err
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"If-Match": strings.Trim(dfw.Etag, `"`),
	}

	ruleSet := dfw.Configuration.Layer3Sections.Section
	// Check that there is a general rule with deny-all
	// If it is missing, add one.
	// If there is a general rule with accept-all, change it to deny-all

	//for _, newRule := range rules {
	// 1. Check that the rule is not already in the current set. If it is (complete equality, except the ID), skip (i.e., the current rule will continue to exist)
	// 2. If the rule is different, but with the same ID, the rule from the input replaces the internal one
	// 3. add the rule
	//}

	contentType := fmt.Sprintf("application/*+xml;version=%s", dfw.client.APIVersion)
	resp, err := dfw.client.ExecuteParamRequestWithCustomError(requestUrl.String(), params, http.MethodPut, contentType, "error updating NSX-V distributed firewall: %s", ruleSet, err)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[update DistributedFirewall] expected status code %d - received %d", http.StatusOK, resp.StatusCode)
	}
	//return dfw.GetConfiguration()
	return nil, fmt.Errorf("not fully implemented yet")
}

// ----------------------------------------------------------------------------------------------
// methods from here till the end of the file will be removed if we decide we don't need services
// ----------------------------------------------------------------------------------------------

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

// GetServiceGroups retrieves the list of services for the current VDC
// If `refresh` = false and the services were already retrieved in a previous operation,
// then it returns the internal values instead of fetching new ones
func (dfw *NsxvDistributedFirewall) GetServiceGroups(refresh bool) ([]types.ApplicationGroup, error) {
	if dfw.ServiceGroups != nil && !refresh {
		return dfw.ServiceGroups, nil
	}
	if dfw.VdcId == "" {
		return nil, fmt.Errorf("no AdminVdc set for this NsxvDistributedFirewall")
	}
	initialUrl, err := dfw.client.buildUrl("network", "services", "applicationgroup", "scope", extractUuid(dfw.VdcId))
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
		return nil, fmt.Errorf("error fetching the service groups: %s", resp.Status)
	}
	var applicationGroupList types.ApplicationGroupList
	err = decodeBody(types.BodyTypeXML, resp, &applicationGroupList)
	if err != nil {
		return nil, err
	}
	dfw.ServiceGroups = applicationGroupList.ApplicationGroup
	return applicationGroupList.ApplicationGroup, nil
}

// Refresh retrieves fresh values for the distributed firewall rules, services, and service groups
func (dfw *NsxvDistributedFirewall) Refresh() error {
	if dfw.VdcId == "" {
		return fmt.Errorf("no AdminVdc set for this NsxvDistributedFirewall")
	}

	_, err := dfw.GetServices(true)
	if err != nil {
		return err
	}

	_, err = dfw.GetServiceGroups(true)
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
// Returns ErrorEntityNotFound when the requested service was not found
func (dfw *NsxvDistributedFirewall) GetServiceByName(serviceName string) (*types.Application, error) {
	services, err := dfw.GetServices(false)
	if err != nil {
		return nil, err
	}
	var foundService types.Application
	for _, app := range services {
		if app.Name == serviceName {
			if foundService.ObjectID != "" {
				return nil, fmt.Errorf("more than one service found with name '%s'", serviceName)
			}
			foundService = app
		}
	}
	if foundService.ObjectID == "" {
		return nil, ErrorEntityNotFound
	}
	return &foundService, nil
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

// GetServiceGroupById retrieves a single service group, identified by its ID, for the current VDC
// If the list of service groups was already retrieved, it uses it, otherwise fetches new ones.
// Returns ErrorEntityNotFound when the requested service group was not found
func (dfw *NsxvDistributedFirewall) GetServiceGroupById(serviceGroupId string) (*types.ApplicationGroup, error) {
	serviceGroups, err := dfw.GetServiceGroups(false)
	if err != nil {
		return nil, err
	}
	for _, appGroup := range serviceGroups {
		if appGroup.ObjectID == serviceGroupId {
			return &appGroup, nil
		}
	}
	return nil, ErrorEntityNotFound
}

// GetServiceGroupByName retrieves a single service group, identified by its name, for the current VDC
// If the list of service groups was already retrieved, it uses it, otherwise fetches new ones.
// Returns ErrorEntityNotFound when the requested service group was not found
func (dfw *NsxvDistributedFirewall) GetServiceGroupByName(serviceGroupName string) (*types.ApplicationGroup, error) {
	serviceGroups, err := dfw.GetServiceGroups(false)
	if err != nil {
		return nil, err
	}
	var foundAppGroup types.ApplicationGroup
	for _, appGroup := range serviceGroups {
		if appGroup.Name == serviceGroupName {
			if foundAppGroup.ObjectID != "" {
				return nil, fmt.Errorf("more than one service group found with name %s", serviceGroupName)
			}
			foundAppGroup = appGroup
		}
	}
	if foundAppGroup.ObjectID == "" {
		return nil, ErrorEntityNotFound
	}
	return &foundAppGroup, nil
}

// GetServiceGroupsByRegex returns a list of services with their names matching the given regular expression
// It may return an empty list (without error)
func (dfw *NsxvDistributedFirewall) GetServiceGroupsByRegex(expression string) ([]types.ApplicationGroup, error) {
	serviceGroups, err := dfw.GetServiceGroups(false)
	if err != nil {
		return nil, err
	}
	searchRegex := regexp.MustCompile(expression)
	var found []types.ApplicationGroup
	for _, appGroup := range serviceGroups {
		if searchRegex.MatchString(appGroup.Name) {
			found = append(found, appGroup)
		}
	}
	return found, nil
}
