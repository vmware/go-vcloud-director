/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
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

	Services      []types.Application      // The list of services for this VDC
	ServiceGroups []types.ApplicationGroup // The list of service groups for this VDC
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

	var firewallSection types.FirewallSection
	err = decodeBody(types.BodyTypeXML, resp, &firewallSection)
	if err != nil {
		return nil, err
	}
	dfw.Etag = resp.Header.Get("etag")
	// The ETag header is needed for further operations. Rules insertion and update need to have a
	// header "If-Match" with the contents of the ETag from a previous read.
	// The same data can be found in the "GenerationNumber" within the section to update.
	// The value of the ETag changes at every GET
	if dfw.Etag == "" && firewallSection.GenerationNumber != "" {
		dfw.Etag = firewallSection.GenerationNumber
	}
	config.Layer3Sections = &types.Layer3Sections{Section: &firewallSection}
	dfw.Configuration = &config
	dfw.Configuration.Layer3Sections = config.Layer3Sections
	dfw.enabled = true
	return &config, nil
}

// IsEnabled returns true when the distributed firewall is enabled
func (dfw *NsxvDistributedFirewall) IsEnabled() (bool, error) {
	if dfw.VdcId == "" {
		return false, fmt.Errorf("no VDC set for this NsxvDistributedFirewall")
	}

	conf, err := dfw.GetConfiguration()
	if err != nil {
		return false, nil
	}
	if dfw.client.APIVersion == "36.0" {
		return conf != nil, nil
	}
	return true, nil
}

// Enable makes the distributed firewall available
// It fails with a non-NSX-V VDC
func (dfw *NsxvDistributedFirewall) Enable() error {
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
		// VCD 10.3.x sometimes returns an error even though the removal succeeds
		if dfw.client.APIVersion == "36.0" {
			conf, _ := dfw.GetConfiguration()
			if conf == nil {
				return nil
			}
		}
		return fmt.Errorf("error deleting Distributed firewall: %s", err)

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

// UpdateConfiguration will either create a new set of rules or update existing ones.
// If the firewall already contains rules, they are overwritten by the ones passed as parameters
func (dfw *NsxvDistributedFirewall) UpdateConfiguration(rules []types.NsxvDistributedFirewallRule) (*types.FirewallConfiguration, error) {

	oldConf, err := dfw.GetConfiguration()
	if err != nil {
		return nil, err
	}
	if dfw.Etag == "" {
		return nil, fmt.Errorf("error getting ETag from distributed firewall")
	}
	initialUrl, err := dfw.client.buildUrl("network", "firewall", "globalroot-0", "config", "layer3sections", dfw.VdcId)
	if err != nil {
		return nil, err
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return nil, err
	}

	var errorList []string
	for i := 0; i < len(rules); i++ {
		rules[i].SectionID = oldConf.Layer3Sections.Section.ID
		if rules[i].Direction == "" {
			errorList = append(errorList, fmt.Sprintf("missing Direction in rule n. %d ", i+1))
		}
		if rules[i].PacketType == "" {
			errorList = append(errorList, fmt.Sprintf("missing Packet Type in rule n. %d ", i+1))
		}
		if rules[i].Action == "" {
			errorList = append(errorList, fmt.Sprintf("missing Action in rule n. %d ", i+1))
		}
	}
	if len(errorList) > 0 {
		return nil, fmt.Errorf("missing required elements from rules: %s", strings.Join(errorList, "; "))
	}

	ruleSet := types.FirewallSection{
		ID:               oldConf.Layer3Sections.Section.ID,
		GenerationNumber: strings.Trim(dfw.Etag, `"`),
		Name:             dfw.VdcId,
		Rule:             rules,
	}

	var newRuleset types.FirewallSection

	dfw.client.SetCustomHeader(map[string]string{
		"If-Match": strings.Trim(oldConf.Layer3Sections.Section.GenerationNumber, `"`),
	})
	defer dfw.client.RemoveCustomHeader()

	contentType := fmt.Sprintf("application/*+xml;version=%s", dfw.client.APIVersion)

	resp, err := dfw.client.ExecuteRequest(requestUrl.String(), http.MethodPut, contentType,
		"error updating NSX-V distributed firewall: %s", ruleSet, &newRuleset)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[update DistributedFirewall] expected status code %d - received %d", http.StatusOK, resp.StatusCode)
	}
	return dfw.GetConfiguration()
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
	searchRegex, err := regexp.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("[GetServicesByRegex] error validating regular expression '%s': %s", expression, err)
	}
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
	searchRegex, err := regexp.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("[GetServiceGroupsByRegex] error validating regular expression '%s': %s", expression, err)
	}
	var found []types.ApplicationGroup
	for _, appGroup := range serviceGroups {
		if searchRegex.MatchString(appGroup.Name) {
			found = append(found, appGroup)
		}
	}
	return found, nil
}
