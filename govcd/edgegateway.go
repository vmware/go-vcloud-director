/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	neturl "net/url"
	"regexp"
	"time"

	types "github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
)

type EdgeGateway struct {
	EdgeGateway *types.EdgeGateway
	client      *Client
}

func NewEdgeGateway(cli *Client) *EdgeGateway {
	return &EdgeGateway{
		EdgeGateway: new(types.EdgeGateway),
		client:      cli,
	}
}

func (eGW *EdgeGateway) AddDhcpPool(network *types.OrgVDCNetwork, dhcppool []interface{}) (Task, error) {
	newedgeconfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration
	util.GovcdLogger.Printf("[DEBUG] EDGE GATEWAY: %#v", newedgeconfig)
	util.GovcdLogger.Printf("[DEBUG] EDGE GATEWAY SERVICE: %#v", newedgeconfig.GatewayDhcpService)
	newdchpservice := &types.GatewayDhcpService{}
	if newedgeconfig.GatewayDhcpService == nil {
		newdchpservice.IsEnabled = true
	} else {
		newdchpservice.IsEnabled = newedgeconfig.GatewayDhcpService.IsEnabled

		for _, poolService := range newedgeconfig.GatewayDhcpService.Pool {

			// Kludgy IF to avoid deleting DNAT rules not created by us.
			// If matches, let's skip it and continue the loop
			if poolService.Network.HREF == network.HREF {
				continue
			}

			newdchpservice.Pool = append(newdchpservice.Pool, poolService)
		}
	}

	for _, item := range dhcppool {
		data := item.(map[string]interface{})

		if data["default_lease_time"] == nil {
			data["default_lease_time"] = 3600
		}

		if data["max_lease_time"] == nil {
			data["max_lease_time"] = 7200
		}

		dhcprule := &types.DhcpPoolService{
			IsEnabled: true,
			Network: &types.Reference{
				HREF: network.HREF,
				Name: network.Name,
			},
			DefaultLeaseTime: data["default_lease_time"].(int),
			MaxLeaseTime:     data["max_lease_time"].(int),
			LowIPAddress:     data["start_address"].(string),
			HighIPAddress:    data["end_address"].(string),
		}
		newdchpservice.Pool = append(newdchpservice.Pool, dhcprule)
	}

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:              "http://www.vmware.com/vcloud/v1.5",
		GatewayDhcpService: newdchpservice,
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	var resp *http.Response
	for {
		buffer := bytes.NewBufferString(xml.Header + string(output))

		apiEndpoint, _ := neturl.ParseRequestURI(eGW.EdgeGateway.HREF)
		apiEndpoint.Path += "/action/configureServices"

		req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)
		util.GovcdLogger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
		util.GovcdLogger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

		req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

		resp, err = checkResp(eGW.client.Http.Do(req))
		if err != nil {
			if match, _ := regexp.MatchString("is busy completing an operation.$", err.Error()); match {
				time.Sleep(3 * time.Second)
				continue
			}
			return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
		}
		break
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) RemoveNATMapping(nattype, externalIP, internalIP, port string) (Task, error) {
	return eGW.RemoveNATPortMapping(nattype, externalIP, port, internalIP, port)
}

func (eGW *EdgeGateway) RemoveNATPortMapping(nattype, externalIP, externalPort string, internalIP, internalPort string) (Task, error) {
	// Find uplink interface
	var uplink types.Reference
	for _, gi := range eGW.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gi.InterfaceType != "uplink" {
			continue
		}
		uplink = *gi.Network
	}

	newedgeconfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	// Take care of the NAT service
	newnatservice := &types.NatService{}

	newnatservice.IsEnabled = newedgeconfig.NatService.IsEnabled
	newnatservice.NatType = newedgeconfig.NatService.NatType
	newnatservice.Policy = newedgeconfig.NatService.Policy
	newnatservice.ExternalIP = newedgeconfig.NatService.ExternalIP

	for _, natRule := range newedgeconfig.NatService.NatRule {

		// Kludgy IF to avoid deleting DNAT rules not created by us.
		// If matches, let's skip it and continue the loop
		if natRule.RuleType == nattype &&
			natRule.GatewayNatRule.OriginalIP == externalIP &&
			natRule.GatewayNatRule.OriginalPort == externalPort &&
			natRule.GatewayNatRule.Interface.HREF == uplink.HREF {
			util.GovcdLogger.Printf("[DEBUG] REMOVING %s Rule: %#v", natRule.RuleType, natRule.GatewayNatRule)
			continue
		}
		util.GovcdLogger.Printf("[DEBUG] KEEPING %s Rule: %#v", natRule.RuleType, natRule.GatewayNatRule)
		newnatservice.NatRule = append(newnatservice.NatRule, natRule)
	}

	newedgeconfig.NatService = newnatservice

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:      "http://www.vmware.com/vcloud/v1.5",
		NatService: newnatservice,
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := neturl.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)
	util.GovcdLogger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
	util.GovcdLogger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		util.GovcdLogger.Printf("[DEBUG] Error is: %#v", err)
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) AddNATMapping(nattype, externalIP, internalIP, port string) (Task, error) {
	return eGW.AddNATPortMapping(nattype, externalIP, port, internalIP, port)
}

func (eGW *EdgeGateway) AddNATPortMapping(nattype, externalIP, externalPort string, internalIP, internalPort string) (Task, error) {
	return eGW.AddNATPortMappingWithUplink(nil, nattype, externalIP, externalPort, internalIP, internalPort)
}

func (eGW *EdgeGateway) getFirstUplink() types.Reference {
	var uplink types.Reference
	for _, gi := range eGW.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gi.InterfaceType != "uplink" {
			continue
		}
		uplink = *gi.Network
	}
	return uplink
}

func (eGW *EdgeGateway) AddNATPortMappingWithUplink(network *types.OrgVDCNetwork, nattype, externalIP, externalPort string, internalIP, internalPort string) (Task, error) {
	// if a network is provided take it, otherwise find first uplink on the edgegateway
	var uplinkRef string

	if network != nil {
		uplinkRef = network.HREF
	} else {
		uplinkRef = eGW.getFirstUplink().HREF
	}

	newedgeconfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	// Take care of the NAT service
	newnatservice := &types.NatService{}

	if newedgeconfig.NatService == nil {
		newnatservice.IsEnabled = true
	} else {
		newnatservice.IsEnabled = newedgeconfig.NatService.IsEnabled
		newnatservice.NatType = newedgeconfig.NatService.NatType
		newnatservice.Policy = newedgeconfig.NatService.Policy
		newnatservice.ExternalIP = newedgeconfig.NatService.ExternalIP

		for _, natRule := range newedgeconfig.NatService.NatRule {

			// Kludgy IF to avoid deleting DNAT rules not created by us.
			// If matches, let's skip it and continue the loop
			if natRule.RuleType == nattype &&
				natRule.GatewayNatRule.OriginalIP == externalIP &&
				natRule.GatewayNatRule.OriginalPort == externalPort &&
				natRule.GatewayNatRule.TranslatedIP == internalIP &&
				natRule.GatewayNatRule.TranslatedPort == internalPort &&
				natRule.GatewayNatRule.Interface.HREF == uplinkRef {
				continue
			}

			newnatservice.NatRule = append(newnatservice.NatRule, natRule)
		}
	}

	//add rule
	natRule := &types.NatRule{
		RuleType:  nattype,
		IsEnabled: true,
		GatewayNatRule: &types.GatewayNatRule{
			Interface: &types.Reference{
				HREF: uplinkRef,
			},
			OriginalIP:     externalIP,
			OriginalPort:   externalPort,
			TranslatedIP:   internalIP,
			TranslatedPort: internalPort,
			Protocol:       "tcp",
		},
	}
	newnatservice.NatRule = append(newnatservice.NatRule, natRule)

	newedgeconfig.NatService = newnatservice

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:      "http://www.vmware.com/vcloud/v1.5",
		NatService: newnatservice,
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := neturl.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)
	util.GovcdLogger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
	util.GovcdLogger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		util.GovcdLogger.Printf("[DEBUG] Error is: %#v", err)
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) CreateFirewallRules(defaultAction string, rules []*types.FirewallRule) (Task, error) {
	err := eGW.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error: %v\n", err)
	}

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		FirewallService: &types.FirewallService{
			IsEnabled:        true,
			DefaultAction:    defaultAction,
			LogDefaultAction: true,
			FirewallRule:     rules,
		},
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error: %v\n", err)
	}

	var resp *http.Response
	for {
		buffer := bytes.NewBufferString(xml.Header + string(output))

		apiEndpoint, _ := neturl.ParseRequestURI(eGW.EdgeGateway.HREF)
		apiEndpoint.Path += "/action/configureServices"

		req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)
		util.GovcdLogger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
		util.GovcdLogger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

		req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

		resp, err = checkResp(eGW.client.Http.Do(req))
		if err != nil {
			if match, _ := regexp.MatchString("is busy completing an operation.$", err.Error()); match {
				time.Sleep(3 * time.Second)
				continue
			}
			return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
		}
		break
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (eGW *EdgeGateway) Refresh() error {

	if eGW.EdgeGateway == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	url, _ := neturl.ParseRequestURI(eGW.EdgeGateway.HREF)

	req := eGW.client.NewRequest(map[string]string{}, "GET", *url, nil)

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error retreiving Edge Gateway: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	eGW.EdgeGateway = &types.EdgeGateway{}

	if err = decodeBody(resp, eGW.EdgeGateway); err != nil {
		return fmt.Errorf("error decoding Edge Gateway response: %s", err)
	}

	// The request was successful
	return nil
}

func (eGW *EdgeGateway) Remove1to1Mapping(internal, external string) (Task, error) {

	// Refresh EdgeGateway rules
	err := eGW.Refresh()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	var uplinkif string
	for _, gifs := range eGW.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gifs.InterfaceType == "uplink" {
			uplinkif = gifs.Network.HREF
		}
	}

	newedgeconfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	// Take care of the NAT service
	newnatservice := &types.NatService{}

	// Copy over the NAT configuration
	newnatservice.IsEnabled = newedgeconfig.NatService.IsEnabled
	newnatservice.NatType = newedgeconfig.NatService.NatType
	newnatservice.Policy = newedgeconfig.NatService.Policy
	newnatservice.ExternalIP = newedgeconfig.NatService.ExternalIP

	for N, natRule := range newedgeconfig.NatService.NatRule {

		// Kludgy IF to avoid deleting DNAT rules not created by us.
		// If matches, let's skip it and continue the loop
		if natRule.RuleType == "DNAT" &&
			natRule.GatewayNatRule.OriginalIP == external &&
			natRule.GatewayNatRule.TranslatedIP == internal &&
			natRule.GatewayNatRule.OriginalPort == "any" &&
			natRule.GatewayNatRule.TranslatedPort == "any" &&
			natRule.GatewayNatRule.Protocol == "any" &&
			natRule.GatewayNatRule.Interface.HREF == uplinkif {
			continue
		}

		// Kludgy IF to avoid deleting SNAT rules not created by us.
		// If matches, let's skip it and continue the loop
		if natRule.RuleType == "SNAT" &&
			natRule.GatewayNatRule.OriginalIP == internal &&
			natRule.GatewayNatRule.TranslatedIP == external &&
			natRule.GatewayNatRule.Interface.HREF == uplinkif {
			continue
		}

		// If doesn't match the above IFs, it's something we need to preserve,
		// let's add it to the new NatService struct
		newnatservice.NatRule = append(newnatservice.NatRule, newedgeconfig.NatService.NatRule[N])

	}

	// Fill the new NatService Section
	newedgeconfig.NatService = newnatservice

	// Take care of the Firewall service
	newfwservice := &types.FirewallService{}

	// Copy over the firewall configuration
	newfwservice.IsEnabled = newedgeconfig.FirewallService.IsEnabled
	newfwservice.DefaultAction = newedgeconfig.FirewallService.DefaultAction
	newfwservice.LogDefaultAction = newedgeconfig.FirewallService.LogDefaultAction

	for N, fwRule := range newedgeconfig.FirewallService.FirewallRule {

		// Kludgy IF to avoid deleting inbound FW rules not created by us.
		// If matches, let's skip it and continue the loop
		if fwRule.Policy == "allow" &&
			fwRule.Protocols.Any == true &&
			fwRule.DestinationPortRange == "Any" &&
			fwRule.SourcePortRange == "Any" &&
			fwRule.SourceIP == "Any" &&
			fwRule.DestinationIP == external {
			continue
		}

		// Kludgy IF to avoid deleting outbound FW rules not created by us.
		// If matches, let's skip it and continue the loop
		if fwRule.Policy == "allow" &&
			fwRule.Protocols.Any == true &&
			fwRule.DestinationPortRange == "Any" &&
			fwRule.SourcePortRange == "Any" &&
			fwRule.SourceIP == internal &&
			fwRule.DestinationIP == "Any" {
			continue
		}

		// If doesn't match the above IFs, it's something we need to preserve,
		// let's add it to the new FirewallService struct
		newfwservice.FirewallRule = append(newfwservice.FirewallRule, newedgeconfig.FirewallService.FirewallRule[N])

	}

	// Fill the new FirewallService Section
	newedgeconfig.FirewallService = newfwservice

	// Fix
	newedgeconfig.NatService.IsEnabled = true

	output, err := xml.MarshalIndent(newedgeconfig, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.GovcdLogger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := neturl.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) Create1to1Mapping(internal, external, description string) (Task, error) {

	// Refresh EdgeGateway rules
	err := eGW.Refresh()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	var uplinkif string
	for _, gifs := range eGW.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gifs.InterfaceType == "uplink" {
			uplinkif = gifs.Network.HREF
		}
	}

	newedgeconfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	snat := &types.NatRule{
		Description: description,
		RuleType:    "SNAT",
		IsEnabled:   true,
		GatewayNatRule: &types.GatewayNatRule{
			Interface: &types.Reference{
				HREF: uplinkif,
			},
			OriginalIP:   internal,
			TranslatedIP: external,
			Protocol:     "any",
		},
	}

	newedgeconfig.NatService.NatRule = append(newedgeconfig.NatService.NatRule, snat)

	dnat := &types.NatRule{
		Description: description,
		RuleType:    "DNAT",
		IsEnabled:   true,
		GatewayNatRule: &types.GatewayNatRule{
			Interface: &types.Reference{
				HREF: uplinkif,
			},
			OriginalIP:     external,
			OriginalPort:   "any",
			TranslatedIP:   internal,
			TranslatedPort: "any",
			Protocol:       "any",
		},
	}

	newedgeconfig.NatService.NatRule = append(newedgeconfig.NatService.NatRule, dnat)

	fwin := &types.FirewallRule{
		Description: description,
		IsEnabled:   true,
		Policy:      "allow",
		Protocols: &types.FirewallRuleProtocols{
			Any: true,
		},
		DestinationPortRange: "Any",
		DestinationIP:        external,
		SourcePortRange:      "Any",
		SourceIP:             "Any",
		EnableLogging:        false,
	}

	newedgeconfig.FirewallService.FirewallRule = append(newedgeconfig.FirewallService.FirewallRule, fwin)

	fwout := &types.FirewallRule{
		Description: description,
		IsEnabled:   true,
		Policy:      "allow",
		Protocols: &types.FirewallRuleProtocols{
			Any: true,
		},
		DestinationPortRange: "Any",
		DestinationIP:        "Any",
		SourcePortRange:      "Any",
		SourceIP:             internal,
		EnableLogging:        false,
	}

	newedgeconfig.FirewallService.FirewallRule = append(newedgeconfig.FirewallService.FirewallRule, fwout)

	output, err := xml.MarshalIndent(newedgeconfig, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.GovcdLogger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := neturl.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) AddIpsecVPN(ipsecVPNConfig *types.EdgeGatewayServiceConfiguration) (Task, error) {

	err := eGW.Refresh()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	output, err := xml.MarshalIndent(ipsecVPNConfig, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshaling ipsecVPNConfig compose: %s", err)
	}

	util.GovcdLogger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))
	util.GovcdLogger.Printf("[DEBUG] ipsecVPN configuration: %s", buffer)

	apiEndpoint, _ := neturl.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}
