//go:build functional || openapi || plugin || ALL

package govcd

import (
	"fmt"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func init() {
	testingTags["plugin"] = "ui_plugin_test.go"
}

// This object is equivalent to the manifest.json that is inside the ../test-resources/ui_plugin.zip file
var testUIPluginMetadata = &types.UIPluginMetadata{
	Vendor:         "VMware",
	License:        "BSD-2-Clause",
	Link:           "http://www.vmware.com",
	PluginName:     "Test Plugin",
	Version:        "1.2.3",
	Description:    "Test Plugin description",
	ProviderScoped: true,
	TenantScoped:   true,
}

// Test_UIPlugin tests all the possible operations that can be done with a UIPlugin object in VCD.
func (vcd *TestVCD) Test_UIPlugin(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	// Add a plugin present on disk
	newUIPlugin, err := vcd.client.AddUIPlugin("../test-resources/ui_plugin.zip", true)
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(newUIPlugin.UIPluginMetadata.ID, check.TestName(), types.OpenApiEndpointExtensionsUi+newUIPlugin.UIPluginMetadata.ID)

	// Assert that the returned metadata from VCD corresponds to the one present inside the ZIP file.
	check.Assert(newUIPlugin.UIPluginMetadata.ID, Not(Equals), "")
	check.Assert(newUIPlugin.UIPluginMetadata.Vendor, Equals, testUIPluginMetadata.Vendor)
	check.Assert(newUIPlugin.UIPluginMetadata.License, Equals, testUIPluginMetadata.License)
	check.Assert(newUIPlugin.UIPluginMetadata.Link, Equals, testUIPluginMetadata.Link)
	check.Assert(newUIPlugin.UIPluginMetadata.PluginName, Equals, testUIPluginMetadata.PluginName)
	check.Assert(newUIPlugin.UIPluginMetadata.Version, Equals, testUIPluginMetadata.Version)
	check.Assert(newUIPlugin.UIPluginMetadata.Description, Equals, testUIPluginMetadata.Description)
	check.Assert(newUIPlugin.UIPluginMetadata.ProviderScoped, Equals, testUIPluginMetadata.ProviderScoped)
	check.Assert(newUIPlugin.UIPluginMetadata.TenantScoped, Equals, testUIPluginMetadata.TenantScoped)
	check.Assert(newUIPlugin.UIPluginMetadata.Enabled, Equals, true)

	// Try to add the same plugin twice, it should fail
	_, err = vcd.client.AddUIPlugin("../test-resources/ui_plugin.zip", true)
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), "same pluginName-version-vendor"))

	// We refresh it to have the latest status
	newUIPlugin, err = vcd.client.GetUIPluginById(newUIPlugin.UIPluginMetadata.ID)
	check.Assert(err, IsNil)
	check.Assert(newUIPlugin.UIPluginMetadata.PluginStatus, Equals, "ready")

	// Retrieve the created plugin using different getters
	allUIPlugins, err := vcd.client.GetAllUIPlugins()
	check.Assert(err, IsNil)
	for _, plugin := range allUIPlugins {
		if plugin.IsTheSameAs(newUIPlugin) {
			check.Assert(plugin.UIPluginMetadata, DeepEquals, newUIPlugin.UIPluginMetadata)
		}
	}
	retrievedUIPlugin, err := vcd.client.GetUIPlugin(newUIPlugin.UIPluginMetadata.Vendor, newUIPlugin.UIPluginMetadata.PluginName, newUIPlugin.UIPluginMetadata.Version)
	check.Assert(err, IsNil)
	check.Assert(retrievedUIPlugin.UIPluginMetadata, DeepEquals, newUIPlugin.UIPluginMetadata)

	// Publishing the plugin to all tenants
	err = newUIPlugin.PublishAll()
	check.Assert(err, IsNil)

	// Retrieving the published tenants, it should equal to the tenants provided in the test configuration + 1 (the System one)
	orgRefs, err := newUIPlugin.GetPublishedTenants()
	check.Assert(err, IsNil)
	check.Assert(orgRefs, NotNil)
	check.Assert(len(orgRefs), Equals, len(vcd.config.Tenants)+1)

	// Unpublishing the plugin from all the tenants
	err = newUIPlugin.UnpublishAll()
	check.Assert(err, IsNil)

	// Retrieving the published tenants, it should equal to 0
	orgRefs, err = newUIPlugin.GetPublishedTenants()
	check.Assert(err, IsNil)
	check.Assert(orgRefs, NotNil)
	check.Assert(len(orgRefs), Equals, 0)

	// Publishing/Unpublishing to/from a specific tenant, if available
	if len(vcd.config.Tenants) > 0 {
		existingOrg, err := vcd.client.GetOrgByName(vcd.config.Tenants[0].SysOrg)
		check.Assert(err, IsNil)
		existingOrgRefs := types.OpenApiReferences{{Name: existingOrg.Org.Name, ID: existingOrg.Org.ID}}

		// Publish to the retrieved tenant
		err = newUIPlugin.Publish(existingOrgRefs)
		check.Assert(err, IsNil)

		// Retrieving the published tenants, it should equal to the tenant above
		orgRefs, err = newUIPlugin.GetPublishedTenants()
		check.Assert(err, IsNil)
		check.Assert(orgRefs, NotNil)
		check.Assert(len(orgRefs), Equals, 1)
		check.Assert(orgRefs[0].Name, Equals, existingOrg.Org.Name)

		// Unpublishing from the same specific tenant
		err = newUIPlugin.Unpublish(existingOrgRefs)
		check.Assert(err, IsNil)

		// Retrieving the published tenants, it should equal to 0 as we just unpublished it
		orgRefs, err = newUIPlugin.GetPublishedTenants()
		check.Assert(err, IsNil)
		check.Assert(orgRefs, NotNil)
		check.Assert(len(orgRefs), Equals, 0)
	}

	// Check that the plugin can be disabled and its scope changed
	err = newUIPlugin.Update(false, false, false)
	check.Assert(err, IsNil)
	check.Assert(newUIPlugin.UIPluginMetadata.Enabled, Equals, false)
	check.Assert(newUIPlugin.UIPluginMetadata.ProviderScoped, Equals, false)
	check.Assert(newUIPlugin.UIPluginMetadata.TenantScoped, Equals, false)

	// Check that the plugin can be enabled again and its scope changed
	err = newUIPlugin.Update(true, true, true)
	check.Assert(err, IsNil)
	check.Assert(newUIPlugin.UIPluginMetadata.Enabled, Equals, true)
	check.Assert(newUIPlugin.UIPluginMetadata.ProviderScoped, Equals, true)
	check.Assert(newUIPlugin.UIPluginMetadata.TenantScoped, Equals, true)

	check.Assert(newUIPlugin.IsTheSameAs(retrievedUIPlugin), Equals, true)
	check.Assert(newUIPlugin.IsTheSameAs(&UIPlugin{UIPluginMetadata: &types.UIPluginMetadata{Vendor: "foo", Version: "1.2.3", PluginName: "bar"}}), Equals, false)

	// Delete the created plugin
	err = newUIPlugin.Delete()
	check.Assert(err, IsNil)
	check.Assert(*newUIPlugin.UIPluginMetadata, DeepEquals, types.UIPluginMetadata{})

	// Check that the plugin was correctly deleted
	_, err = vcd.client.GetUIPlugin(retrievedUIPlugin.UIPluginMetadata.Vendor, retrievedUIPlugin.UIPluginMetadata.PluginName, retrievedUIPlugin.UIPluginMetadata.Version)
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), ErrorEntityNotFound.Error()))
	_, err = vcd.client.GetUIPluginById(retrievedUIPlugin.UIPluginMetadata.ID)
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), ErrorEntityNotFound.Error()))
}

// Test_getPluginMetadata tests that getPluginMetadata can retrieve correctly the UI plugin metadata information
// stored inside the ZIP file.
func Test_getPluginMetadata(t *testing.T) {
	tests := []struct {
		name       string
		pluginPath string
		want       *types.UIPluginMetadata
		wantErr    bool
	}{
		{
			name:       "get ui plugin metadata",
			pluginPath: "../test-resources/ui_plugin.zip",
			want:       testUIPluginMetadata,
		},
		{
			name:       "invalid plugin",
			pluginPath: "../test-resources/udf_test.iso",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPluginMetadata(tt.pluginPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPluginMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPluginMetadata() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_getTransferIdFromHeader tests that getTransferIdFromHeader can retrieve correctly a transfer ID from the headers
// of any HTTP response.
func Test_getTransferIdFromHeader(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		want    string
		wantErr bool
	}{
		{
			name: "valid link in header",
			headers: http.Header{
				textproto.CanonicalMIMEHeaderKey("link"): {
					"<https://www.my-vcd.com/transfer/cb63b0f6-ba56-43a8-8fe3-a64f0b25e7e5/my-amazing-plugin1.0.zip>;rel=\"upload:default\";type=\"application/octet-stream\"",
				},
			},
			want:    "cb63b0f6-ba56-43a8-8fe3-a64f0b25e7e5/my-amazing-plugin1.0.zip",
			wantErr: false,
		},
		{
			name: "valid link in header with special URI",
			headers: http.Header{
				textproto.CanonicalMIMEHeaderKey("link"): {
					"<my-vcd:8080/transfer/cb63b0f6-ba56-43a8-8fe3-a64f0b25e7e5/my-amazing-plugin1.1.zip>;rel=\"upload:default\";type=\"application/octet-stream\"",
				},
			},
			want:    "cb63b0f6-ba56-43a8-8fe3-a64f0b25e7e5/my-amazing-plugin1.1.zip",
			wantErr: false,
		},
		{
			name: "empty header",
			headers: http.Header{
				textproto.CanonicalMIMEHeaderKey("link"): {
					"",
				},
			},
			wantErr: true,
		},
		{
			name: "empty link in header",
			headers: http.Header{
				textproto.CanonicalMIMEHeaderKey("link"): {
					"<>;rel=\"upload:default\";type=\"application/octet-stream\"",
				},
			},
			wantErr: true,
		},
		{
			name: "no link part in header",
			headers: http.Header{
				textproto.CanonicalMIMEHeaderKey("link"): {
					"rel=\"upload:default\";type=\"application/octet-stream\"",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid header",
			headers: http.Header{
				textproto.CanonicalMIMEHeaderKey("link"): {
					"Error",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTransferIdFromHeader(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTransferIdFromHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getTransferIdFromHeader() got = %v, want %v", got, tt.want)
			}
		})
	}
}
