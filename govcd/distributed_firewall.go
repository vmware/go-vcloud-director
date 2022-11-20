/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// buildUrl uses the Client base URL to create a customised URL
func buildUrl(client *Client, elements ...string) (string, error) {
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

// IsDistributedFirewallEnabled checks whether a VDC has the distributed firewall functionality enabled
// This function fails with non-NSX-V VDCs
func (adminVdc *AdminVdc) IsDistributedFirewallEnabled() (bool, error) {
	initialUrl, err := buildUrl(adminVdc.client, "network", "firewall", "globalroot-0", "config")
	if err != nil {
		return false, err
	}

	params := map[string]string{
		"vdc": extractUuid(adminVdc.AdminVdc.ID),
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return false, err
	}

	req := adminVdc.client.NewRequest(params, http.MethodGet, *requestUrl, nil)

	resp, err := checkResp(adminVdc.client.Http.Do(req))
	if (resp != nil && resp.StatusCode == 400) || (err != nil && strings.Contains(err.Error(), "400 Bad Request")) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// Start temporary code to make the response appear in the logs
	var body []byte
	err = decodeBody(types.BodyTypeXML, resp, &body)
	if err != nil {
		return false, err
	}
	// end temporary code
	return true, nil
}

// EnableDistributedFirewall enables the distributed firewall functionality for this VDC
// It is an idempotent operation. It can be repeated on an already enabled VDC without errors.
// This function fails with non-NSX-V VDCs
func (adminVdc *AdminVdc) EnableDistributedFirewall() error {
	initialUrl, err := buildUrl(adminVdc.client, "network", "firewall", "vdc", extractUuid(adminVdc.AdminVdc.ID))
	if err != nil {
		return err
	}

	params := map[string]string{
		"vdc": extractUuid(adminVdc.AdminVdc.ID),
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return err
	}

	req := adminVdc.client.NewRequest(params, http.MethodPost, *requestUrl, nil)

	resp, err := checkResp(adminVdc.client.Http.Do(req))
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("[EnableDistributedFirewall] expected status code %d - received %d", http.StatusCreated, resp.StatusCode)
	}
	return nil
}

// DisableDistributedFirewall removes the distributed firewall functionality from a VDC
// This function fails with non-NSX-V VDCs
func (adminVdc *AdminVdc) DisableDistributedFirewall() error {
	initialUrl, err := buildUrl(adminVdc.client, "network", "firewall", "vdc", extractUuid(adminVdc.AdminVdc.ID))
	if err != nil {
		return err
	}

	params := map[string]string{
		"vdc": extractUuid(adminVdc.AdminVdc.ID),
	}

	requestUrl, err := url.ParseRequestURI(initialUrl)
	if err != nil {
		return err
	}

	req := adminVdc.client.NewRequest(params, http.MethodDelete, *requestUrl, nil)

	resp, err := checkResp(adminVdc.client.Http.Do(req))
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("[DisableDistributedFirewall] expected status code %d - received %d", http.StatusNoContent, resp.StatusCode)
	}
	return nil
}
