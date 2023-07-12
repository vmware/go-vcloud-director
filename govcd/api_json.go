/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// executeJsonRequest is a wrapper around regular API call operations, similar to client.ExecuteRequest, but with JSON payback
// Returns a http.Response object, which, in case of success, has its body still unread
// Caller function has the responsibility for closing the response body
func (client Client) executeJsonRequest(href, httpMethod string, inputStructure any, errorMessage string) (*http.Response, error) {

	text, err := json.MarshalIndent(inputStructure, " ", " ")
	if err != nil {
		return nil, err
	}
	requestHref, err := url.Parse(href)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	body := strings.NewReader(string(text))
	apiVersion := client.APIVersion
	headAccept := http.Header{}
	headAccept.Set("Accept", fmt.Sprintf("application/*+json;version=%s", apiVersion))
	headAccept.Set("Content-Type", "application/*+json")
	request := client.newRequest(nil, nil, httpMethod, *requestHref, body, apiVersion, headAccept)
	resp, err = client.Http.Do(request)
	if err != nil {
		return nil, fmt.Errorf(errorMessage, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		util.ProcessResponseOutput(util.CallFuncName(), resp, string(body))
		var jsonError types.OpenApiError
		err = json.Unmarshal(body, &jsonError)
		// By default, we return the whole response body as error message. This may also contain the stack trace
		message := string(body)
		// if the body contains a valid JSON representation of the error, we return a more agile message, using the
		// exposed fields, and hiding the stack trace from view
		if err == nil {
			message = fmt.Sprintf("%s - %s", jsonError.MinorErrorCode, jsonError.Message)
		}
		util.ProcessResponseOutput(util.CallFuncName(), resp, string(body))
		return resp, fmt.Errorf(errorMessage, message)
	}

	return checkRespWithErrType(types.BodyTypeJSON, resp, err, &types.Error{})
}

// closeBody is a wrapper function that should be used with "defer" after calling executeJsonRequest
func closeBody(resp *http.Response) {
	err := resp.Body.Close()
	if err != nil {
		util.Logger.Printf("error closing response body - Called by %s: %s\n", util.CallFuncName(), err)
	}
}
