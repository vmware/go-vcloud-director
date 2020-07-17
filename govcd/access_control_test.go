// +build functional vapp catalog ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type controlAccessType interface {
	GetAccessControl() (*types.ControlAccessParams, error)
	SetAccessControl(params *types.ControlAccessParams) error
	RemoveAccessControl() error
	IsShared() bool
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

func testAccessControl(label string, accessible controlAccessType, params types.ControlAccessParams, expected types.ControlAccessParams, wantShared bool, check *C) error {

	fmt.Printf("-- %s\n", label)
	err := accessible.SetAccessControl(&params)
	if err != nil {
		return err
	}

	foundParams, err := accessible.GetAccessControl()
	if err != nil {
		return err
	}

	if testVerbose {
		text, err := json.MarshalIndent(foundParams, " ", " ")
		if err == nil {
			fmt.Printf("%s %s\n", label, text)
		}
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

	check.Assert(accessible.IsShared(), Equals, wantShared)

	return nil
}
