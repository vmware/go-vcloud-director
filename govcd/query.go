/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
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

func (vcdCli *VCDClient) Query(ctx context.Context, params map[string]string) (Results, error) {

	req := vcdCli.Client.NewRequest(ctx, params, http.MethodGet, vcdCli.QueryHREF, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vcdCli.Client.APIVersion)

	return getResult(&vcdCli.Client, req)
}

func (vdc *Vdc) Query(ctx context.Context, params map[string]string) (Results, error) {
	queryUrl := vdc.client.VCDHREF
	queryUrl.Path += "/query"
	req := vdc.client.NewRequest(ctx, params, http.MethodGet, queryUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vdc.client.APIVersion)

	return getResult(vdc.client, req)
}

// QueryWithNotEncodedParams uses Query API to search for requested data
func (client *Client) QueryWithNotEncodedParams(ctx context.Context, params map[string]string, notEncodedParams map[string]string) (Results, error) {
	return client.QueryWithNotEncodedParamsWithApiVersion(ctx, params, notEncodedParams, client.APIVersion)
}

// QueryWithNotEncodedParams uses Query API to search for requested data
func (client *Client) QueryWithNotEncodedParamsWithApiVersion(ctx context.Context, params map[string]string, notEncodedParams map[string]string, apiVersion string) (Results, error) {
	queryUlr := client.VCDHREF
	queryUlr.Path += "/query"

	req := client.NewRequestWitNotEncodedParamsWithApiVersion(ctx, params, notEncodedParams, http.MethodGet, queryUlr, nil, apiVersion)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+apiVersion)

	return getResult(client, req)
}

func (vcdCli *VCDClient) QueryWithNotEncodedParams(ctx context.Context, params map[string]string, notEncodedParams map[string]string) (Results, error) {
	return vcdCli.Client.QueryWithNotEncodedParams(ctx, params, notEncodedParams)
}

func (vdc *Vdc) QueryWithNotEncodedParams(ctx context.Context, params map[string]string, notEncodedParams map[string]string) (Results, error) {
	return vdc.client.QueryWithNotEncodedParams(ctx, params, notEncodedParams)
}

func (vcdCli *VCDClient) QueryWithNotEncodedParamsWithApiVersion(ctx context.Context, params map[string]string, notEncodedParams map[string]string, apiVersion string) (Results, error) {
	return vcdCli.Client.QueryWithNotEncodedParamsWithApiVersion(ctx, params, notEncodedParams, apiVersion)
}

func (vdc *Vdc) QueryWithNotEncodedParamsWithApiVersion(ctx context.Context, params map[string]string, notEncodedParams map[string]string, apiVersion string) (Results, error) {
	return vdc.client.QueryWithNotEncodedParamsWithApiVersion(ctx, params, notEncodedParams, apiVersion)
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
