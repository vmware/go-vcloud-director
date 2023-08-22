//go:build functional || openapi || uiPlugin || ALL

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"strings"
)

func init() {
	testingTags["uiPlugin"] = "ui_plugin_test.go"
}

// Test_UIPlugin tests all the possible operations that can be done with a UIPlugin object in VCD.
func (vcd *TestVCD) Test_UIPlugin(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	if vcd.config.Media.UiPluginPath == "" {
		check.Skip("The testing configuration property 'media.uiPluginPath' is empty")
	}

	testUIPluginMetadata, err := getPluginMetadata(vcd.config.Media.UiPluginPath)
	check.Assert(err, IsNil)

	// Add a plugin present on disk
	newUIPlugin, err := vcd.client.AddUIPlugin(vcd.config.Media.UiPluginPath, true)
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
	_, err = vcd.client.AddUIPlugin(vcd.config.Media.UiPluginPath, true)
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), "same pluginName-version-vendor"))

	// We refresh it to have the latest status
	newUIPlugin, err = vcd.client.GetUIPluginById(newUIPlugin.UIPluginMetadata.ID)
	check.Assert(err, IsNil)
	check.Assert(true, Equals, newUIPlugin.UIPluginMetadata.PluginStatus == "ready" || newUIPlugin.UIPluginMetadata.PluginStatus == "unavailable")

	// We check that the error returned by a non-existent ID is correct:
	_, err = vcd.client.GetUIPluginById("urn:vcloud:uiPlugin:00000000-0000-0000-0000-000000000000")
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), "could not find any UI plugin with ID"))

	// Retrieve the created plugin using different getters
	allUIPlugins, err := vcd.client.GetAllUIPlugins()
	check.Assert(err, IsNil)
	for _, plugin := range allUIPlugins {
		if plugin.IsTheSameAs(newUIPlugin) {
			plugin.UIPluginMetadata.PluginStatus = newUIPlugin.UIPluginMetadata.PluginStatus // We ignore status as it can be quite arbitrary
			check.Assert(plugin.UIPluginMetadata, DeepEquals, newUIPlugin.UIPluginMetadata)
		}
	}
	retrievedUIPlugin, err := vcd.client.GetUIPlugin(newUIPlugin.UIPluginMetadata.Vendor, newUIPlugin.UIPluginMetadata.PluginName, newUIPlugin.UIPluginMetadata.Version)
	check.Assert(err, IsNil)
	retrievedUIPlugin.UIPluginMetadata.PluginStatus = newUIPlugin.UIPluginMetadata.PluginStatus // We ignore status as it can be quite arbitrary
	check.Assert(retrievedUIPlugin.UIPluginMetadata, DeepEquals, newUIPlugin.UIPluginMetadata)

	// Publishing the plugin to all tenants
	err = newUIPlugin.PublishAll()
	check.Assert(err, IsNil)

	// Retrieving the published tenants, it should be at least the number of tenants provided in the test configuration + 1 (the System one)
	orgRefs, err := newUIPlugin.GetPublishedTenants()
	check.Assert(err, IsNil)
	check.Assert(orgRefs, NotNil)
	check.Assert(len(orgRefs) >= len(vcd.config.Tenants)+1, Equals, true)

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
