// +build vapp vm functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// guestPropertyGetSetter interface is used for covering tests in both VM and vApp guest property
type productSectionListGetSetter interface {
	GetProductSectionList(ctx context.Context) (*types.ProductSectionList, error)
	SetProductSectionList(ctx context.Context, productSection *types.ProductSectionList) (*types.ProductSectionList, error)
}

// propertyTester is a guest property setter accepting guestPropertyGetSetter interface for trying
// out settings on all objects implementing such interface
func propertyTester(ctx context.Context, vcd *TestVCD, check *C, object productSectionListGetSetter) {
	productSection := &types.ProductSectionList{
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

	productSection.SortByPropertyKeyName()

	gotproductSection, err := object.SetProductSectionList(ctx, productSection)
	check.Assert(err, IsNil)
	gotproductSection.SortByPropertyKeyName()

	getproductSection, err := object.GetProductSectionList(ctx)
	check.Assert(err, IsNil)
	getproductSection.SortByPropertyKeyName()

	// Check that values were set in API
	check.Assert(getproductSection, NotNil)
	check.Assert(getproductSection.ProductSection, NotNil)
	check.Assert(len(getproductSection.ProductSection.Property), Equals, 3)

	check.Assert(getproductSection.ProductSection.Property[0].Key, Equals, "asset_tag")
	check.Assert(getproductSection.ProductSection.Property[0].Label, Equals, "asset_tag_label")
	check.Assert(getproductSection.ProductSection.Property[0].Type, Equals, "string")
	check.Assert(getproductSection.ProductSection.Property[0].Value.Value, Equals, "xxxyyy")
	check.Assert(getproductSection.ProductSection.Property[0].DefaultValue, Equals, "asset_tag_default")
	check.Assert(getproductSection.ProductSection.Property[0].UserConfigurable, Equals, true)

	check.Assert(getproductSection.ProductSection.Property[1].Key, Equals, "guestinfo.config.bootstrap.ip")
	check.Assert(getproductSection.ProductSection.Property[1].Label, Equals, "guestinfo.config.bootstrap.ip_label")
	check.Assert(getproductSection.ProductSection.Property[1].Type, Equals, "string")
	check.Assert(getproductSection.ProductSection.Property[1].Value.Value, Equals, "192.168.12.180")
	check.Assert(getproductSection.ProductSection.Property[1].DefaultValue, Equals, "default_ip")
	check.Assert(getproductSection.ProductSection.Property[1].UserConfigurable, Equals, true)

	check.Assert(getproductSection.ProductSection.Property[2].Key, Equals, "sys_owner")
	check.Assert(getproductSection.ProductSection.Property[2].Label, Equals, "sys_owner_label")
	check.Assert(getproductSection.ProductSection.Property[2].Type, Equals, "string")
	check.Assert(getproductSection.ProductSection.Property[2].Value.Value, Equals, "test")
	check.Assert(getproductSection.ProductSection.Property[2].DefaultValue, Equals, "sys_owner_default")
	check.Assert(getproductSection.ProductSection.Property[2].UserConfigurable, Equals, false)

	// Ensure the object are deeply equal
	check.Assert(gotproductSection.ProductSection.Property, DeepEquals, productSection.ProductSection.Property)
	check.Assert(getproductSection, DeepEquals, gotproductSection)
}

// guestPropertyGetSetter interface is used for covering tests
type getGuestCustomizationSectionGetSetter interface {
	GetGuestCustomizationSection(ctx context.Context) (*types.GuestCustomizationSection, error)
	SetGuestCustomizationSection(ctx context.Context, guestCustomizationSection *types.GuestCustomizationSection) (*types.GuestCustomizationSection, error)
}

// guestCustomizationPropertyTester is a guest customization property get and setter accepting guestPropertyGetSetter interface for trying
// out settings on all objects implementing such interface
func guestCustomizationPropertyTester(ctx context.Context, vcd *TestVCD, check *C, object getGuestCustomizationSectionGetSetter) {
	setupedGuestCustomizationSection := &types.GuestCustomizationSection{
		Enabled: takeBoolPointer(true), JoinDomainEnabled: takeBoolPointer(false), UseOrgSettings: takeBoolPointer(false),
		DomainUserName: "", DomainName: "", DomainUserPassword: "",
		AdminPasswordEnabled: takeBoolPointer(true), AdminPassword: "adminPass", AdminPasswordAuto: takeBoolPointer(false),
		AdminAutoLogonEnabled: takeBoolPointer(true), AdminAutoLogonCount: 15, ResetPasswordRequired: takeBoolPointer(true),
		CustomizationScript: "ls", ComputerName: "Cname18"}

	guestCustomizationSection, err := object.SetGuestCustomizationSection(ctx, setupedGuestCustomizationSection)
	check.Assert(err, IsNil)

	// Check that values were set from API
	check.Assert(guestCustomizationSection, NotNil)

	check.Assert(*guestCustomizationSection.Enabled, Equals, true)
	check.Assert(*guestCustomizationSection.JoinDomainEnabled, Equals, false)
	check.Assert(*guestCustomizationSection.UseOrgSettings, Equals, false)
	check.Assert(guestCustomizationSection.DomainUserName, Equals, "")
	check.Assert(guestCustomizationSection.DomainName, Equals, "")
	check.Assert(guestCustomizationSection.DomainUserPassword, Equals, "")
	check.Assert(*guestCustomizationSection.AdminPasswordEnabled, Equals, true)
	check.Assert(*guestCustomizationSection.AdminPasswordAuto, Equals, false)
	check.Assert(guestCustomizationSection.AdminPassword, Equals, "adminPass")
	check.Assert(guestCustomizationSection.AdminAutoLogonCount, Equals, 15)
	check.Assert(*guestCustomizationSection.AdminAutoLogonEnabled, Equals, true)
	check.Assert(*guestCustomizationSection.ResetPasswordRequired, Equals, true)
	check.Assert(guestCustomizationSection.CustomizationScript, Equals, "ls")
	check.Assert(guestCustomizationSection.ComputerName, Equals, "Cname18")

	// Double check if GuestCustomizationSection retrieved from separate API is the same as the one
	// embedded directly in the VM
	vm := object.(*VM)

	// Refresh VM to have the latest structure
	err = vm.Refresh(ctx)
	check.Assert(err, IsNil)

	// Deep compare values retrieved from separate API to the ones embedded into VM
	guestCustomizationSection.Xmlns = "" // embedded VM structure does not have this field
	// Links amount may differ between structures
	guestCustomizationSection.Link = types.LinkList{}
	vm.VM.GuestCustomizationSection.Link = types.LinkList{}
	check.Assert(guestCustomizationSection, DeepEquals, vm.VM.GuestCustomizationSection)
}
