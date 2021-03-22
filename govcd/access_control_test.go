// +build functional vapp catalog ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kr/pretty"
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// accessControlType is an interface used to test access control for all entities that support it
type accessControlType interface {
	GetAccessControl(ctx context.Context, useTenantContext bool) (*types.ControlAccessParams, error)
	SetAccessControl(ctx context.Context, params *types.ControlAccessParams, useTenantContext bool) error
	RemoveAccessControl(ctx context.Context, useTenantContext bool) error
	IsShared(ctx context.Context, useTenantContext bool) (bool, error)
	GetId() string
}

type accessSettingMap map[string]*types.AccessSetting

func accessSettingsToMap(params types.ControlAccessParams) accessSettingMap {
	if params.AccessSettings == nil {
		return nil
	}
	var result = make(accessSettingMap)

	for _, setting := range params.AccessSettings.AccessSetting {

		result[setting.Subject.Name] = setting
	}

	return result
}

// testAccessControl runs an access control test on a target type, identified as an interface.
// * label is a test identifier
// * accessible is the entity being tested (such as a vApp or a catalog)
// * params are the access control parameters to be set
// * expected are the result parameters that we should find in the end result
// * wantShared is whether the final settings should result in the entity being shared
func testAccessControl(ctx context.Context, label string, accessible accessControlType, params types.ControlAccessParams, expected types.ControlAccessParams, wantShared bool, useTenantContext bool, check *C) error {

	if testVerbose {
		fmt.Printf("-- %s\n", label)
	}
	err := accessible.SetAccessControl(ctx, &params, useTenantContext)
	if err != nil {
		return err
	}

	foundParams, err := accessible.GetAccessControl(ctx, useTenantContext)
	if err != nil {
		return err
	}

	if testVerbose {
		text, err := json.MarshalIndent(foundParams, " ", " ")
		if err == nil {
			fmt.Printf("%s %s\n", label, text)
		}
	}
	if foundParams.IsSharedToEveryone != expected.IsSharedToEveryone {
		fmt.Printf("label    %s\n", label)
		fmt.Printf("found    %# v\n", pretty.Formatter(foundParams))
		fmt.Printf("expected %# v\n", pretty.Formatter(expected))
	}
	check.Assert(foundParams.IsSharedToEveryone, Equals, expected.IsSharedToEveryone)
	if expected.EveryoneAccessLevel != nil {
		check.Assert(foundParams.EveryoneAccessLevel, NotNil)
		check.Assert(*foundParams.EveryoneAccessLevel, Equals, *expected.EveryoneAccessLevel)
	}
	if expected.AccessSettings != nil {
		expectedMap := accessSettingsToMap(expected)
		foundMap := accessSettingsToMap(expected)

		check.Assert(foundParams.AccessSettings, NotNil)
		check.Assert(len(foundParams.AccessSettings.AccessSetting), Equals, len(expected.AccessSettings.AccessSetting))

		for k, v := range expectedMap {
			found, exists := foundMap[k]
			check.Assert(exists, Equals, true)
			if v.Subject.Name != "" {
				check.Assert(v.Subject.Name, Equals, found.Subject.Name)
			}
			check.Assert(v.Subject.Type, Not(Equals), "")
			check.Assert(v.Subject.HREF, Not(Equals), "")
			check.Assert(v.Subject.HREF, Equals, found.Subject.HREF)
			check.Assert(v.AccessLevel, Equals, found.AccessLevel)
		}
	}

	shared, err := accessible.IsShared(ctx, useTenantContext)
	if err != nil {
		return err
	}
	check.Assert(shared, Equals, wantShared)

	return nil
}
