package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type UIPluginMetadata struct {
	UIPluginMetadata *types.UIPluginMetadata
	client           *Client
}

// CreateUIPlugin creates a new UI extension and sets the provided plugin metadata for it.
// Only System administrator can create a UI extension.
func (vcdClient *VCDClient) CreateUIPlugin(uiPluginMetadata *types.UIPluginMetadata) (*UIPluginMetadata, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExtensionsUi
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &UIPluginMetadata{
		UIPluginMetadata: &types.UIPluginMetadata{},
		client:           &vcdClient.Client,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, uiPluginMetadata, result.UIPluginMetadata, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Upload uploads the given UI Plugin to VCD. Only the file name in the input types.UploadSpec is required.
// The size is calculated automatically if not provided.
func (ui *UIPluginMetadata) Upload(uploadSpec *types.UploadSpec) error {
	if strings.TrimSpace(uploadSpec.FileName) == "" {
		return fmt.Errorf("file name to upload must not be empty")
	}
	fileContents, err := os.ReadFile(filepath.Clean(uploadSpec.FileName))
	if err != nil {
		return err
	}

	if uploadSpec.Size <= 0 {
		uploadSpec.Size = len(fileContents)
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExtensionsUiPlugin
	apiVersion, err := ui.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := ui.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return err
	}

	headers, err := ui.client.OpenApiPostItemAndGetHeaders(apiVersion, urlRef, nil, uploadSpec, nil, nil)
	if err != nil {
		return err
	}

	transferId, err := getTransferIdFromHeader(headers)
	if err != nil {
		return err
	}

	endpoint = types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTransfer
	apiVersion, err = ui.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err = ui.client.OpenApiBuildEndpoint(endpoint, transferId)
	if err != nil {
		return err
	}

	return ui.client.OpenApiPutItem(apiVersion, urlRef, nil, fileContents, nil, nil)
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

func (*UIPluginMetadata) Publish(orgs types.OpenApiReferences) (types.OpenApiReferences, error) {
	return nil, nil
}

func (*UIPluginMetadata) PublishAll() {

}

func (*UIPluginMetadata) Unpublish(orgs types.OpenApiReferences) (types.OpenApiReferences, error) {
	return nil, nil
}

func (*UIPluginMetadata) UnpublishAll() {

}

func (*UIPluginMetadata) Delete() {

}
