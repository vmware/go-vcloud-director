package govcd

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// ExtendedSessionInfo collects data regarding a VCD connection
type ExtendedSessionInfo struct {
	User           string
	Org            string
	Roles          []string
	Rights         []string
	Version        string
	ConnectionType string
}

// GetSessionInfo collects the basic session information for a VCD connection
func (client *Client) GetSessionInfo() (*types.CurrentSessionInfo, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSessionCurrent

	// We get the maximum supported version, as early versions of the API return less data
	apiVersion, err := client.MaxSupportedVersion()
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var info types.CurrentSessionInfo

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &info, nil)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// GetExtendedSessionInfo collects extended session information for support and debugging
// It will try to collect as much data as possible, failing only if the minimum data can't
// be collected.
func (vcdClient *VCDClient) GetExtendedSessionInfo() (*ExtendedSessionInfo, error) {
	var extendedSessionInfo ExtendedSessionInfo
	sessionInfo, err := vcdClient.Client.GetSessionInfo()
	if err != nil {
		return nil, err
	}
	switch {
	case vcdClient.Client.UsingBearerToken:
		extendedSessionInfo.ConnectionType = "Bearer token"
	case vcdClient.Client.UsingAccessToken:
		extendedSessionInfo.ConnectionType = "API Access token"
	default:
		extendedSessionInfo.ConnectionType = "Username + password"
	}
	version, err := vcdClient.Client.GetVcdFullVersion()
	if err == nil {
		extendedSessionInfo.Version = version.Version.String()
	}
	if sessionInfo.User.Name == "" {
		return nil, fmt.Errorf("no user reference found")
	}
	extendedSessionInfo.User = sessionInfo.User.Name

	if sessionInfo.Org.Name == "" {
		return nil, fmt.Errorf("no Org reference found")
	}
	extendedSessionInfo.Org = sessionInfo.Org.Name

	if len(sessionInfo.Roles) == 0 {
		return &extendedSessionInfo, nil
	}
	extendedSessionInfo.Roles = append(extendedSessionInfo.Roles, sessionInfo.Roles...)
	org, err := vcdClient.GetAdminOrgById(sessionInfo.Org.ID)
	if err != nil {
		return &extendedSessionInfo, err
	}
	for _, roleRef := range sessionInfo.RoleRefs {
		role, err := org.GetRoleById(roleRef.ID)
		if err != nil {
			continue
		}
		rights, err := role.GetRights(nil)
		if err != nil {
			continue
		}
		for _, right := range rights {
			extendedSessionInfo.Rights = append(extendedSessionInfo.Rights, right.Name)
		}
	}
	return &extendedSessionInfo, nil
}

// LogSessionInfo prints session information into the default logs
func (client *VCDClient) LogSessionInfo() {

	// If logging is disabled, there is no point in collecting session info
	if util.EnableLogging {
		info, err := client.GetExtendedSessionInfo()
		if err != nil {
			util.Logger.Printf("no session info collected: %s\n", err)
			return
		}
		text, err := json.MarshalIndent(info, " ", " ")
		if err != nil {
			util.Logger.Printf("error formatting session info %s\n", err)
			return
		}
		util.Logger.Println(strings.Repeat("*", 80))
		util.Logger.Println("START SESSION INFO")
		util.Logger.Println(strings.Repeat("*", 80))
		util.Logger.Printf("%s\n", text)
		util.Logger.Println(strings.Repeat("*", 80))
		util.Logger.Println("END SESSION INFO")
		util.Logger.Println(strings.Repeat("*", 80))
	}
}
