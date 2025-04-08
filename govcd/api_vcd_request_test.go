//go:build api || openapi || functional || catalog || vapp || gateway || network || org || query || extnetwork || task || vm || vdc || system || disk || lb || lbAppRule || lbAppProfile || lbServerPool || lbServiceMonitor || lbVirtualServer || user || search || nsxv || nsxt || auth || affinity || role || alb || certificate || vdcGroup || metadata || providervdc || rde || vsphere || uiPlugin || cse || slz || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"

	. "gopkg.in/check.v1"
)

// Test_NewRequestWitNotEncodedParamsWithApiVersion verifies that api version override works
func (vcd *TestVCD) Test_NewRequestWitNotEncodedParamsWithApiVersion(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	queryUlr := vcd.client.Client.VCDHREF
	queryUlr.Path += "/query"

	apiVersion, err := vcd.client.Client.MaxSupportedVersion()
	check.Assert(err, IsNil)

	req := vcd.client.Client.NewRequestWitNotEncodedParamsWithApiVersion(nil, map[string]string{"type": "media",
		"filter": "name==any"}, http.MethodGet, queryUlr, nil, apiVersion)

	check.Assert(req.Header.Get("User-Agent"), Equals, vcd.client.Client.UserAgent)

	resp, err := checkResp(vcd.client.Client.Http.Do(req))
	check.Assert(err, IsNil)

	check.Assert(resp.Header.Get("Content-Type"), Equals, types.MimeQueryRecords+";version="+apiVersion)

	bodyBytes, err := rewrapRespBodyNoopCloser(resp)
	check.Assert(err, IsNil)

	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, string(bodyBytes))
	debugShowResponse(resp, bodyBytes)

	// Repeats the call without API version change
	req = vcd.client.Client.NewRequestWitNotEncodedParams(nil, map[string]string{"type": "media",
		"filter": "name==any"}, http.MethodGet, queryUlr, nil)

	resp, err = checkResp(vcd.client.Client.Http.Do(req))
	check.Assert(err, IsNil)

	// Checks that the regularAPI version was not affected by the previous call
	check.Assert(resp.Header.Get("Content-Type"), Equals, types.MimeQueryRecords+";version="+vcd.client.Client.APIVersion)

	bodyBytes, err = rewrapRespBodyNoopCloser(resp)
	check.Assert(err, IsNil)
	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, string(bodyBytes))
	debugShowResponse(resp, bodyBytes)

	fmt.Printf("Test: %s run with api Version: %s\n", check.TestName(), apiVersion)
}
