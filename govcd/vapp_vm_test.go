// +build vapp vm functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// guestPropertyGetSetter interface is used for covering tests in both VM and vApp guest property
type guestPropertyGetSetter interface {
	GetGuestProperties() (*types.ProductSectionList, error)
	SetGuestProperties(properties *types.ProductSectionList) (*types.ProductSectionList, error)
}

// propertyTester is a guest property setter accepting guestPropertyGetSetter interface for trying
// out settings on all objects implementing such interface
func propertyTester(vcd *TestVCD, check *C, object guestPropertyGetSetter) {
	vappProperties := &types.ProductSectionList{
		ProductSection: &types.ProductSection{
			Info: "Custom properties",
			Property: []*types.Property{
				&types.Property{
					UserConfigurable: false,
					Key:              "sys_owner",
					Label:            "sys_owner_label",
					Type:             "string",
					DefaultValue:     "sys_owner_default",
					Value:            &types.Value{Value: "test"},
				},
				&types.Property{
					UserConfigurable: true,
					Key:              "asset_tag",
					Label:            "asset_tag_label",
					Type:             "string",
					DefaultValue:     "asset_tag_default",
					Value:            &types.Value{Value: "xxxyyy"},
				},
				&types.Property{
					UserConfigurable: true,
					Key:              "guestinfo.config.bootstrap.ip",
					Label:            "guestinfo.config.bootstrap.ip_label",
					Type:             "string",
					DefaultValue:     "default_ip",
					Value:            &types.Value{Value: "192.168.12.180"},
				},
			},
		},
	}

	gotProperties, err := object.SetGuestProperties(vappProperties)
	check.Assert(err, IsNil)

	getProperties, err := object.GetGuestProperties()
	check.Assert(err, IsNil)

	// Check that values were set in API
	check.Assert(getProperties.ProductSection.Property[0].Key, Equals, "sys_owner")
	check.Assert(getProperties.ProductSection.Property[0].Label, Equals, "sys_owner_label")
	check.Assert(getProperties.ProductSection.Property[0].Type, Equals, "string")
	check.Assert(getProperties.ProductSection.Property[0].Value.Value, Equals, "test")
	check.Assert(getProperties.ProductSection.Property[0].DefaultValue, Equals, "sys_owner_default")
	check.Assert(getProperties.ProductSection.Property[0].UserConfigurable, Equals, false)

	check.Assert(getProperties.ProductSection.Property[1].Key, Equals, "asset_tag")
	check.Assert(getProperties.ProductSection.Property[1].Label, Equals, "asset_tag_label")
	check.Assert(getProperties.ProductSection.Property[1].Type, Equals, "string")
	check.Assert(getProperties.ProductSection.Property[1].Value.Value, Equals, "xxxyyy")
	check.Assert(getProperties.ProductSection.Property[1].DefaultValue, Equals, "asset_tag_default")
	check.Assert(getProperties.ProductSection.Property[1].UserConfigurable, Equals, true)

	check.Assert(getProperties.ProductSection.Property[2].Key, Equals, "guestinfo.config.bootstrap.ip")
	check.Assert(getProperties.ProductSection.Property[2].Label, Equals, "guestinfo.config.bootstrap.ip_label")
	check.Assert(getProperties.ProductSection.Property[2].Type, Equals, "string")
	check.Assert(getProperties.ProductSection.Property[2].Value.Value, Equals, "192.168.12.180")
	check.Assert(getProperties.ProductSection.Property[2].DefaultValue, Equals, "default_ip")
	check.Assert(getProperties.ProductSection.Property[2].UserConfigurable, Equals, true)

	// Ensure the object are deeply equal
	check.Assert(gotProperties.ProductSection.Property, DeepEquals, vappProperties.ProductSection.Property)
	check.Assert(getProperties, DeepEquals, gotProperties)
}
