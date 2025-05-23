// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

type Results struct {
	Results *types.QueryResultRecordsType
	client  *Client
}

func NewResults(cli *Client) *Results {
	return &Results{
		Results: new(types.QueryResultRecordsType),
		client:  cli,
	}
}

func (vcdClient *VCDClient) Query(params map[string]string) (Results, error) {

	req := vcdClient.Client.NewRequest(params, http.MethodGet, vcdClient.QueryHREF, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vcdClient.Client.APIVersion)

	return getResult(&vcdClient.Client, req)
}

func (vdc *Vdc) Query(params map[string]string) (Results, error) {
	queryUrl := vdc.client.VCDHREF
	queryUrl.Path += "/query"
	req := vdc.client.NewRequest(params, http.MethodGet, queryUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vdc.client.APIVersion)

	return getResult(vdc.client, req)
}

// QueryWithNotEncodedParams uses Query API to search for requested data
func (client *Client) QueryWithNotEncodedParams(params map[string]string, notEncodedParams map[string]string) (Results, error) {
	return client.QueryWithNotEncodedParamsWithApiVersion(params, notEncodedParams, client.APIVersion)
}

func (client *Client) QueryWithNotEncodedParamsWithHeaders(params map[string]string, notEncodedParams map[string]string, headers map[string]string) (Results, error) {
	return client.QueryWithNotEncodedParamsWithApiVersionWithHeaders(params, notEncodedParams, client.APIVersion, headers)
}

// QueryWithNotEncodedParams uses Query API to search for requested data
func (client *Client) QueryWithNotEncodedParamsWithApiVersion(params map[string]string, notEncodedParams map[string]string, apiVersion string) (Results, error) {
	return client.QueryWithNotEncodedParamsWithApiVersionWithHeaders(params, notEncodedParams, apiVersion, nil)
}

func (client *Client) QueryWithNotEncodedParamsWithApiVersionWithHeaders(params map[string]string, notEncodedParams map[string]string, apiVersion string, headers map[string]string) (Results, error) {
	queryUrl := client.VCDHREF
	queryUrl.Path += "/query"

	req := client.NewRequestWitNotEncodedParamsWithApiVersion(params, notEncodedParams, http.MethodGet, queryUrl, nil, apiVersion)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+apiVersion)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return getResult(client, req)
}

func (vcdClient *VCDClient) QueryWithNotEncodedParams(params map[string]string, notEncodedParams map[string]string) (Results, error) {
	return vcdClient.Client.QueryWithNotEncodedParams(params, notEncodedParams)
}

func (vdc *Vdc) QueryWithNotEncodedParams(params map[string]string, notEncodedParams map[string]string) (Results, error) {
	return vdc.client.QueryWithNotEncodedParams(params, notEncodedParams)
}

func (vcdClient *VCDClient) QueryWithNotEncodedParamsWithApiVersion(params map[string]string, notEncodedParams map[string]string, apiVersion string) (Results, error) {
	return vcdClient.Client.QueryWithNotEncodedParamsWithApiVersion(params, notEncodedParams, apiVersion)
}

func (vdc *Vdc) QueryWithNotEncodedParamsWithApiVersion(params map[string]string, notEncodedParams map[string]string, apiVersion string) (Results, error) {
	return vdc.client.QueryWithNotEncodedParamsWithApiVersion(params, notEncodedParams, apiVersion)
}

func getResult(client *Client, request *http.Request) (Results, error) {
	resp, err := checkResp(client.Http.Do(request))
	if err != nil {
		return Results{}, fmt.Errorf("error retrieving query: %s", err)
	}

	results := NewResults(client)

	if err = decodeBody(types.BodyTypeXML, resp, results.Results); err != nil {
		return Results{}, fmt.Errorf("error decoding query results: %s", err)
	}

	return *results, nil
}
