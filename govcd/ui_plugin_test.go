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
var testUiPluginMetadata = &types.UIPluginMetadata{
	Vendor:         "VMware",
	License:        "BSD-2-Clause",
	Link:           "http://www.vmware.com",
	PluginName:     "Test Plugin",
	Version:        "1.2.3",
	Description:    "Test Plugin description",
	ProviderScoped: true,
	TenantScoped:   true,
}

func (vcd *TestVCD) Test_UIPlugin(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	// Add a plugin present on disk
	uiPlugin, err := vcd.client.AddUIPlugin("../test-resources/ui_plugin.zip")
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(uiPlugin.UIPluginMetadata.ID, check.TestName(), types.OpenApiEndpointExtensionsUi+uiPlugin.UIPluginMetadata.ID)

	// Assert that the metadata is correctly read from the uploaded ZIP
	check.Assert(uiPlugin.UIPluginMetadata.ID, Not(Equals), "")
	check.Assert(uiPlugin.UIPluginMetadata.Vendor, Equals, testUiPluginMetadata.Vendor)
	check.Assert(uiPlugin.UIPluginMetadata.License, Equals, testUiPluginMetadata.License)
	check.Assert(uiPlugin.UIPluginMetadata.Link, Equals, testUiPluginMetadata.Link)
	check.Assert(uiPlugin.UIPluginMetadata.PluginName, Equals, testUiPluginMetadata.PluginName)
	check.Assert(uiPlugin.UIPluginMetadata.Version, Equals, testUiPluginMetadata.Version)
	check.Assert(uiPlugin.UIPluginMetadata.Description, Equals, testUiPluginMetadata.Description)
	check.Assert(uiPlugin.UIPluginMetadata.ProviderScoped, Equals, testUiPluginMetadata.ProviderScoped)
	check.Assert(uiPlugin.UIPluginMetadata.TenantScoped, Equals, testUiPluginMetadata.TenantScoped)

	// Try to add the same plugin twice, it should fail
	_, err = vcd.client.AddUIPlugin("../test-resources/ui_plugin.zip")
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), "same pluginName-version-vendor"))

	// Retrieve the created plugin using different getters
	allUIplugins, err := vcd.client.GetAllUIPlugins()
	check.Assert(err, IsNil)
	for _, plugin := range allUIplugins {
		if plugin.IsTheSameAs(uiPlugin) {
			check.Assert(plugin.UIPluginMetadata, DeepEquals, uiPlugin.UIPluginMetadata)
		}
	}
	anotherUiPlugin, err := vcd.client.GetUIPlugin(uiPlugin.UIPluginMetadata.Vendor, uiPlugin.UIPluginMetadata.PluginName, uiPlugin.UIPluginMetadata.Version)
	check.Assert(err, IsNil)
	check.Assert(anotherUiPlugin.UIPluginMetadata, DeepEquals, uiPlugin.UIPluginMetadata)
	anotherUiPlugin, err = vcd.client.GetUIPluginById(uiPlugin.UIPluginMetadata.ID)
	check.Assert(err, IsNil)
	check.Assert(anotherUiPlugin.UIPluginMetadata, DeepEquals, uiPlugin.UIPluginMetadata)

	// Publishing to all tenants
	err = uiPlugin.PublishAll()
	check.Assert(err, IsNil)

	// Retrieving the published tenants, it should equal to the tenants provided in the test configuration + 1 (the System one)
	orgRefs, err := uiPlugin.GetPublishedTenants()
	check.Assert(err, IsNil)
	check.Assert(orgRefs, NotNil)
	check.Assert(len(orgRefs), Equals, len(vcd.config.Tenants)+1)

	// Unpublishing from all the tenants
	err = uiPlugin.UnpublishAll()
	check.Assert(err, IsNil)

	// Retrieving the published tenants, it should equal to 0 as we just unpublished all
	orgRefs, err = uiPlugin.GetPublishedTenants()
	check.Assert(err, IsNil)
	check.Assert(orgRefs, NotNil)
	check.Assert(len(orgRefs), Equals, 0)

	// Publishing to a specific tenant
	if len(vcd.config.Tenants) > 0 {
		existingOrg, err := vcd.client.GetOrgByName(vcd.config.Tenants[0].SysOrg)
		check.Assert(err, IsNil)

		existingOrgRefs := types.OpenApiReferences{
			{
				Name: existingOrg.Org.Name,
				ID:   existingOrg.Org.ID,
			},
		}

		err = uiPlugin.Publish(existingOrgRefs)
		check.Assert(err, IsNil)

		// Retrieving the published tenants, it should equal to the tenant above
		orgRefs, err = uiPlugin.GetPublishedTenants()
		check.Assert(err, IsNil)
		check.Assert(orgRefs, NotNil)
		check.Assert(len(orgRefs), Equals, 1)
		check.Assert(orgRefs[0].Name, Equals, existingOrg.Org.Name)

		// Unpublishing from all the tenants
		err = uiPlugin.Unpublish(existingOrgRefs)
		check.Assert(err, IsNil)

		// Retrieving the published tenants, it should equal to 0 as we just unpublished it
		orgRefs, err = uiPlugin.GetPublishedTenants()
		check.Assert(err, IsNil)
		check.Assert(orgRefs, NotNil)
		check.Assert(len(orgRefs), Equals, 0)
	}

	// Delete the created plugin
	err = uiPlugin.Delete()
	check.Assert(err, IsNil)
	check.Assert(*uiPlugin.UIPluginMetadata, DeepEquals, types.UIPluginMetadata{})

	// Check that the plugin was correctly deleted
	_, err = vcd.client.GetUIPlugin(anotherUiPlugin.UIPluginMetadata.Vendor, anotherUiPlugin.UIPluginMetadata.PluginName, anotherUiPlugin.UIPluginMetadata.Version)
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), ErrorEntityNotFound.Error()))
	_, err = vcd.client.GetUIPluginById(anotherUiPlugin.UIPluginMetadata.ID)
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
			want:       testUiPluginMetadata,
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
