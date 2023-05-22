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

func (vcdClient *VCDClient) AddUIPlugin(pluginPath string) (*UIPlugin, error) {
	if strings.TrimSpace(pluginPath) == "" {
		return nil, fmt.Errorf("plugin path must not be empty")
	}
	uiPluginMetadataPayload, err := getPluginMetadata(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("error retrieving the metadata for the given plugin %s: %s", pluginPath, err)
	}
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

	return uiPlugin.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
}

// getPluginMetadata retrieves the UI Plugin Metadata information stored inside the given .zip file.
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
		return nil, fmt.Errorf("")
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

	return &types.UIPluginMetadata{
		Vendor:      unmarshaledJson["vendor"].(string),
		License:     unmarshaledJson["license"].(string),
		Link:        unmarshaledJson["link"].(string),
		PluginName:  unmarshaledJson["name"].(string),
		Version:     unmarshaledJson["version"].(string),
		Description: unmarshaledJson["description"].(string),
	}, nil
}

// createUIPlugin creates a new UI extension and sets the provided plugin metadata for it.
// Only System administrator can create a UI extension.
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

// getTransferIdFromHeader retrieves a valid transfer ID from any given HTTP headers
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

func (*UIPlugin) Publish(orgs types.OpenApiReferences) (types.OpenApiReferences, error) {
	return nil, nil
}

func (*UIPlugin) PublishAll() {

}

func (*UIPlugin) Unpublish(orgs types.OpenApiReferences) (types.OpenApiReferences, error) {
	return nil, nil
}

func (*UIPlugin) UnpublishAll() {

}
