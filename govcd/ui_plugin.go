package govcd

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type UIPlugin struct {
	UIPluginMetadata *types.UIPluginMetadata
	client           *Client
}

// AddUIPlugin reads the plugin ZIP file located in the input path, obtains the inner metadata, sends it to
// VCD and performs the plugin upload.
func (vcdClient *VCDClient) AddUIPlugin(pluginPath string, enabled bool) (*UIPlugin, error) {
	if strings.TrimSpace(pluginPath) == "" {
		return nil, fmt.Errorf("plugin path must not be empty")
	}
	uiPluginMetadataPayload, err := getPluginMetadata(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("error retrieving the metadata for the given plugin %s: %s", pluginPath, err)
	}
	uiPluginMetadataPayload.Enabled = enabled
	uiPluginMetadata, err := createUIPlugin(&vcdClient.Client, uiPluginMetadataPayload)
	if err != nil {
		return nil, fmt.Errorf("error creating the UI plugin: %s", err)
	}
	err = uiPluginMetadata.upload(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("error uploading the UI plugin: %s", err)
	}

	return uiPluginMetadata, nil
}

// GetAllUIPlugins retrieves a slice with all the available UIPlugin objects present in VCD.
func (vcdClient *VCDClient) GetAllUIPlugins() ([]*UIPlugin, error) {
	endpoint := types.OpenApiEndpointExtensionsUi // This one is not versioned, hence not using types.OpenApiPathVersion1_0_0 or alike
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	var typeResponses []*types.UIPluginMetadata
	err = vcdClient.Client.OpenApiGetItem(apiVersion, urlRef, nil, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into UIPlugin types with client
	uiPlugins := make([]*UIPlugin, len(typeResponses))
	for sliceIndex := range typeResponses {
		uiPlugins[sliceIndex] = &UIPlugin{
			UIPluginMetadata: typeResponses[sliceIndex],
			client:           &vcdClient.Client,
		}
	}

	return uiPlugins, nil
}

// GetUIPluginById obtains a unique UIPlugin identified by its URN.
func (vcdClient *VCDClient) GetUIPluginById(id string) (*UIPlugin, error) {
	endpoint := types.OpenApiEndpointExtensionsUi // This one is not versioned, hence not using types.OpenApiPathVersion1_0_0 or alike
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	result := &UIPlugin{
		UIPluginMetadata: &types.UIPluginMetadata{},
		client:           &vcdClient.Client,
	}
	err = vcdClient.Client.OpenApiGetItem(apiVersion, urlRef, nil, result.UIPluginMetadata, nil)
	if err != nil {
		return nil, amendUIPluginGetByIdError(id, err)
	}

	return result, nil
}

// amendUIPluginGetByIdError is a workaround for a bug in VCD that causes the GET endpoint to return an ugly error 500 with a NullPointerException
// when the UI Plugin with given ID is not found
func amendUIPluginGetByIdError(id string, err error) error {
	if err != nil && strings.Contains(err.Error(), "NullPointerException") {
		return fmt.Errorf("could not find any UI plugin with ID '%s': %s", id, ErrorEntityNotFound)
	}
	return err
}

// GetUIPlugin obtains a unique UIPlugin identified by the combination of its vendor, plugin name and version.
func (vcdClient *VCDClient) GetUIPlugin(vendor, pluginName, version string) (*UIPlugin, error) {
	allUIPlugins, err := vcdClient.GetAllUIPlugins()
	if err != nil {
		return nil, err
	}
	for _, plugin := range allUIPlugins {
		if plugin.IsTheSameAs(&UIPlugin{UIPluginMetadata: &types.UIPluginMetadata{
			Vendor:     vendor,
			PluginName: pluginName,
			Version:    version,
		}}) {
			return plugin, nil
		}
	}

	return nil, fmt.Errorf("could not find any UI plugin with vendor '%s', pluginName '%s' and version '%s': %s", vendor, pluginName, version, ErrorEntityNotFound)
}

// GetPublishedTenants gets all the Organization references where the receiver UIPlugin is published.
func (uiPlugin *UIPlugin) GetPublishedTenants() (types.OpenApiReferences, error) {
	if strings.TrimSpace(uiPlugin.UIPluginMetadata.ID) == "" {
		return nil, fmt.Errorf("plugin ID is required but it is empty")
	}

	endpoint := types.OpenApiEndpointExtensionsUiTenants // This one is not versioned, hence not using types.OpenApiPathVersion1_0_0 or alike
	apiVersion, err := uiPlugin.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := uiPlugin.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, uiPlugin.UIPluginMetadata.ID))
	if err != nil {
		return nil, err
	}

	var orgRefs types.OpenApiReferences
	err = uiPlugin.client.OpenApiGetAllItems(apiVersion, urlRef, nil, &orgRefs, nil)
	if err != nil {
		return nil, err
	}
	return orgRefs, nil
}

// Publish publishes the receiver UIPlugin to the given Organizations.
// Does not modify the receiver UIPlugin.
func (uiPlugin *UIPlugin) Publish(orgs types.OpenApiReferences) error {
	if len(orgs) == 0 {
		return nil
	}
	return publishOrUnpublishFromOrgs(uiPlugin.client, uiPlugin.UIPluginMetadata.ID, orgs, types.OpenApiEndpointExtensionsUiTenantsPublish)
}

// Unpublish unpublishes the receiver UIPlugin from the given Organizations.
// Does not modify the receiver UIPlugin.
func (uiPlugin *UIPlugin) Unpublish(orgs types.OpenApiReferences) error {
	if len(orgs) == 0 {
		return nil
	}
	return publishOrUnpublishFromOrgs(uiPlugin.client, uiPlugin.UIPluginMetadata.ID, orgs, types.OpenApiEndpointExtensionsUiTenantsUnpublish)
}

// PublishAll publishes the receiver UIPlugin to all available Organizations.
// Does not modify the receiver UIPlugin.
func (uiPlugin *UIPlugin) PublishAll() error {
	return publishOrUnpublishFromOrgs(uiPlugin.client, uiPlugin.UIPluginMetadata.ID, nil, types.OpenApiEndpointExtensionsUiTenantsPublishAll)
}

// UnpublishAll unpublishes the receiver UIPlugin from all available Organizations.
// Does not modify the receiver UIPlugin.
func (uiPlugin *UIPlugin) UnpublishAll() error {
	return publishOrUnpublishFromOrgs(uiPlugin.client, uiPlugin.UIPluginMetadata.ID, nil, types.OpenApiEndpointExtensionsUiTenantsUnpublishAll)
}

// IsTheSameAs retruns true if the receiver UIPlugin has the same name, vendor and version as the input.
func (uiPlugin *UIPlugin) IsTheSameAs(otherUiPlugin *UIPlugin) bool {
	if otherUiPlugin == nil {
		return false
	}
	return uiPlugin.UIPluginMetadata.PluginName == otherUiPlugin.UIPluginMetadata.PluginName &&
		uiPlugin.UIPluginMetadata.Version == otherUiPlugin.UIPluginMetadata.Version &&
		uiPlugin.UIPluginMetadata.Vendor == otherUiPlugin.UIPluginMetadata.Vendor
}

// Update performs an update to several receiver plugin attributes
func (uiPlugin *UIPlugin) Update(enable, providerScoped, tenantScoped bool) error {
	if strings.TrimSpace(uiPlugin.UIPluginMetadata.ID) == "" {
		return fmt.Errorf("plugin ID is required but it is empty")
	}

	endpoint := types.OpenApiEndpointExtensionsUi // This one is not versioned, hence not using types.OpenApiPathVersion1_0_0 or alike
	apiVersion, err := uiPlugin.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := uiPlugin.client.OpenApiBuildEndpoint(endpoint, uiPlugin.UIPluginMetadata.ID)
	if err != nil {
		return err
	}

	payload := &types.UIPluginMetadata{
		Vendor:         uiPlugin.UIPluginMetadata.Vendor,
		License:        uiPlugin.UIPluginMetadata.License,
		Link:           uiPlugin.UIPluginMetadata.Link,
		PluginName:     uiPlugin.UIPluginMetadata.PluginName,
		Version:        uiPlugin.UIPluginMetadata.Version,
		Description:    uiPlugin.UIPluginMetadata.Description,
		ProviderScoped: providerScoped,
		TenantScoped:   tenantScoped,
		Enabled:        enable,
	}
	err = uiPlugin.client.OpenApiPutItem(apiVersion, urlRef, nil, payload, uiPlugin.UIPluginMetadata, nil)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the receiver UIPlugin from VCD.
func (uiPlugin *UIPlugin) Delete() error {
	if strings.TrimSpace(uiPlugin.UIPluginMetadata.ID) == "" {
		return fmt.Errorf("plugin ID must not be empty")
	}

	endpoint := types.OpenApiEndpointExtensionsUi // This one is not versioned, hence not using types.OpenApiPathVersion1_0_0 or alike
	apiVersion, err := uiPlugin.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := uiPlugin.client.OpenApiBuildEndpoint(endpoint, uiPlugin.UIPluginMetadata.ID)
	if err != nil {
		return err
	}

	err = uiPlugin.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}
	uiPlugin.UIPluginMetadata = &types.UIPluginMetadata{}
	return nil
}

// getPluginMetadata retrieves the types.UIPluginMetadata information stored inside the given plugin file, that should
// be a ZIP file.
func getPluginMetadata(pluginPath string) (*types.UIPluginMetadata, error) {
	archive, err := zip.OpenReader(filepath.Clean(pluginPath))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := archive.Close(); err != nil {
			util.Logger.Printf("Error closing ZIP file: %s\n", err)
		}
	}()

	var manifest *zip.File
	for _, f := range archive.File {
		if f.Name == "manifest.json" {
			manifest = f
			break
		}
	}
	if manifest == nil {
		return nil, fmt.Errorf("could not find manifest.json inside the file %s", pluginPath)
	}

	manifestContents, err := manifest.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := manifestContents.Close(); err != nil {
			util.Logger.Printf("Error closing manifest file: %s\n", err)
		}
	}()

	manifestBytes, err := io.ReadAll(manifestContents)
	if err != nil {
		return nil, err
	}

	var unmarshaledJson map[string]interface{}
	err = json.Unmarshal(manifestBytes, &unmarshaledJson)
	if err != nil {
		return nil, err
	}

	result := &types.UIPluginMetadata{
		Vendor:      unmarshaledJson["vendor"].(string),
		License:     unmarshaledJson["license"].(string),
		Link:        unmarshaledJson["link"].(string),
		PluginName:  unmarshaledJson["name"].(string),
		Version:     unmarshaledJson["version"].(string),
		Description: unmarshaledJson["description"].(string),
	}

	for _, scope := range unmarshaledJson["scope"].([]interface{}) {
		if strings.Contains(scope.(string), "provider") {
			result.ProviderScoped = true
		} else if strings.Contains(scope.(string), "tenant") {
			result.TenantScoped = true
		}
	}

	return result, nil
}

// createUIPlugin creates a new empty UIPlugin in VCD and sets the provided plugin metadata.
// The UI plugin contents should be uploaded afterwards.
func createUIPlugin(client *Client, uiPluginMetadata *types.UIPluginMetadata) (*UIPlugin, error) {
	endpoint := types.OpenApiEndpointExtensionsUi // This one is not versioned, hence not using types.OpenApiPathVersion1_0_0 or alike
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &UIPlugin{
		UIPluginMetadata: &types.UIPluginMetadata{},
		client:           client,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, uiPluginMetadata, result.UIPluginMetadata, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// This function uploads the given UI Plugin to VCD. Only the plugin path is required.
func (ui *UIPlugin) upload(pluginPath string) error {
	fileContents, err := os.ReadFile(filepath.Clean(pluginPath))
	if err != nil {
		return err
	}

	endpoint := types.OpenApiEndpointExtensionsUiPlugin // This one is not versioned, hence not using types.OpenApiPathVersion1_0_0 or alike
	apiVersion, err := ui.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := ui.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ui.UIPluginMetadata.ID))
	if err != nil {
		return err
	}

	uploadSpec := types.UploadSpec{
		FileName:     filepath.Base(pluginPath),
		ChecksumAlgo: "sha256",
		Checksum:     fmt.Sprintf("%x", sha256.Sum256(fileContents)),
		Size:         int64(len(fileContents)),
	}

	headers, err := ui.client.OpenApiPostItemAndGetHeaders(apiVersion, urlRef, nil, uploadSpec, nil, nil)
	if err != nil {
		return err
	}

	transferId, err := getTransferIdFromHeader(headers)
	if err != nil {
		return err
	}

	transferEndpoint := fmt.Sprintf("%s://%s/transfer/%s", ui.client.VCDHREF.Scheme, ui.client.VCDHREF.Host, transferId)
	request, err := newFileUploadRequest(ui.client, transferEndpoint, fileContents, 0, uploadSpec.Size, uploadSpec.Size)
	if err != nil {
		return err
	}

	response, err := ui.client.Http.Do(request)
	if err != nil {
		return err
	}
	return response.Body.Close()
}

// getTransferIdFromHeader retrieves a valid transfer ID from any given HTTP headers, that can be used to upload
// a UI Plugin to VCD.
func getTransferIdFromHeader(headers http.Header) (string, error) {
	rawLinkContent := headers.Get("link")
	if rawLinkContent == "" {
		return "", fmt.Errorf("error during UI plugin upload, the POST call didn't return any transfer link")
	}
	linkRegex := regexp.MustCompile(`<\S+/transfer/(\S+)>`)
	matches := linkRegex.FindStringSubmatch(rawLinkContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("error during UI plugin upload, the POST call didn't return a valid transfer link: %s", rawLinkContent)
	}
	return matches[1], nil
}

// publishOrUnpublishFromOrgs publishes or unpublishes (depending on the input endpoint) the UI Plugin with given ID from all available
// organizations.
func publishOrUnpublishFromOrgs(client *Client, pluginId string, orgs types.OpenApiReferences, endpoint string) error {
	if strings.TrimSpace(pluginId) == "" {
		return fmt.Errorf("plugin ID is required but it is empty")
	}

	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, pluginId))
	if err != nil {
		return err
	}

	return client.OpenApiPostItem(apiVersion, urlRef, nil, orgs, nil, nil)
}
