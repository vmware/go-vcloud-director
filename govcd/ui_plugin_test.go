//go:build functional || openapi || plugin || ALL

package govcd

import (
	"net/http"
	"net/textproto"
	"reflect"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func init() {
	testingTags["plugin"] = "ui_plugin_test.go"
}

func (vcd *TestVCD) Test_UIPlugin(check *C) {
	uiPlugin, err := vcd.client.AddUIPlugin("../test-resources/ui_plugin.zip")
	check.Assert(err, IsNil)

	err = uiPlugin.Delete()
	check.Assert(err, IsNil)
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
			want: &types.UIPluginMetadata{
				Vendor:      "VMware",
				License:     "BSD-2-Clause",
				Link:        "http://www.vmware.com/support",
				PluginName:  "Test Plugin",
				Version:     "1.2.3",
				Description: "Test Plugin description",
			},
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
