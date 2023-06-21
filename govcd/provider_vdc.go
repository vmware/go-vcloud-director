package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ProviderVdc is the basic Provider VDC structure, contains the minimum set of attributes.
type ProviderVdc struct {
	ProviderVdc *types.ProviderVdc
	client      *Client
}

// ProviderVdcExtended is the extended Provider VDC structure, contains same attributes as ProviderVdc plus some more.
type ProviderVdcExtended struct {
	VMWProviderVdc *types.VMWProviderVdc
	client         *Client
}

func newProviderVdc(cli *Client) *ProviderVdc {
	return &ProviderVdc{
		ProviderVdc: new(types.ProviderVdc),
		client:      cli,
	}
}

func newProviderVdcExtended(cli *Client) *ProviderVdcExtended {
	return &ProviderVdcExtended{
		VMWProviderVdc: new(types.VMWProviderVdc),
		client:         cli,
	}
}

// GetProviderVdcByHref finds a Provider VDC by its HREF.
// On success, returns a pointer to the ProviderVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcByHref(providerVdcHref string) (*ProviderVdc, error) {
	providerVdc := newProviderVdc(&vcdClient.Client)

	_, err := vcdClient.Client.ExecuteRequest(providerVdcHref, http.MethodGet,
		"", "error retrieving Provider VDC: %s", nil, providerVdc.ProviderVdc)
	if err != nil {
		return nil, err
	}

	return providerVdc, nil
}

// GetProviderVdcExtendedByHref finds a Provider VDC with extended attributes by its HREF.
// On success, returns a pointer to the ProviderVdcExtended structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcExtendedByHref(providerVdcHref string) (*ProviderVdcExtended, error) {
	providerVdc := newProviderVdcExtended(&vcdClient.Client)

	_, err := vcdClient.Client.ExecuteRequest(getAdminExtensionURL(providerVdcHref), http.MethodGet,
		"", "error retrieving extended Provider VDC: %s", nil, providerVdc.VMWProviderVdc)
	if err != nil {
		return nil, err
	}

	return providerVdc, nil
}

// GetProviderVdcById finds a Provider VDC by URN.
// On success, returns a pointer to the ProviderVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcById(providerVdcId string) (*ProviderVdc, error) {
	providerVdcHref := vcdClient.Client.VCDHREF
	providerVdcHref.Path += "/admin/providervdc/" + extractUuid(providerVdcId)

	return vcdClient.GetProviderVdcByHref(providerVdcHref.String())
}

// GetProviderVdcExtendedById finds a Provider VDC with extended attributes by URN.
// On success, returns a pointer to the ProviderVdcExtended structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcExtendedById(providerVdcId string) (*ProviderVdcExtended, error) {
	providerVdcHref := vcdClient.Client.VCDHREF
	providerVdcHref.Path += "/admin/extension/providervdc/" + extractUuid(providerVdcId)

	return vcdClient.GetProviderVdcExtendedByHref(providerVdcHref.String())
}

// GetProviderVdcByName finds a Provider VDC by name.
// On success, returns a pointer to the ProviderVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcByName(providerVdcName string) (*ProviderVdc, error) {
	providerVdc, err := getProviderVdcByName(vcdClient, providerVdcName, false)
	if err != nil {
		return nil, err
	}
	return providerVdc.(*ProviderVdc), err
}

// GetProviderVdcExtendedByName finds a Provider VDC with extended attributes by name.
// On success, returns a pointer to the ProviderVdcExtended structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcExtendedByName(providerVdcName string) (*ProviderVdcExtended, error) {
	providerVdcExtended, err := getProviderVdcByName(vcdClient, providerVdcName, true)
	if err != nil {
		return nil, err
	}
	return providerVdcExtended.(*ProviderVdcExtended), err
}

// Refresh updates the contents of the Provider VDC associated to the receiver object.
func (providerVdc *ProviderVdc) Refresh() error {
	if providerVdc.ProviderVdc.HREF == "" {
		return fmt.Errorf("cannot refresh, receiver Provider VDC is empty")
	}

	unmarshalledVdc := &types.ProviderVdc{}

	_, err := providerVdc.client.ExecuteRequest(providerVdc.ProviderVdc.HREF, http.MethodGet,
		"", "error refreshing Provider VDC: %s", nil, unmarshalledVdc)
	if err != nil {
		return err
	}

	providerVdc.ProviderVdc = unmarshalledVdc

	return nil
}

// Refresh updates the contents of the extended Provider VDC associated to the receiver object.
func (providerVdcExtended *ProviderVdcExtended) Refresh() error {
	if providerVdcExtended.VMWProviderVdc.HREF == "" {
		return fmt.Errorf("cannot refresh, receiver extended Provider VDC is empty")
	}

	unmarshalledVdc := &types.VMWProviderVdc{}

	_, err := providerVdcExtended.client.ExecuteRequest(providerVdcExtended.VMWProviderVdc.HREF, http.MethodGet,
		"", "error refreshing extended Provider VDC: %s", nil, unmarshalledVdc)
	if err != nil {
		return err
	}

	providerVdcExtended.VMWProviderVdc = unmarshalledVdc

	return nil
}

// ToProviderVdc converts the receiver ProviderVdcExtended into the subset ProviderVdc
func (providerVdcExtended *ProviderVdcExtended) ToProviderVdc() (*ProviderVdc, error) {
	providerVdcHref := providerVdcExtended.client.VCDHREF
	providerVdcHref.Path += "/admin/providervdc/" + extractUuid(providerVdcExtended.VMWProviderVdc.ID)

	providerVdc := newProviderVdc(providerVdcExtended.client)

	_, err := providerVdcExtended.client.ExecuteRequest(providerVdcHref.String(), http.MethodGet,
		"", "error retrieving Provider VDC: %s", nil, providerVdc.ProviderVdc)
	if err != nil {
		return nil, err
	}

	return providerVdc, nil
}

// getProviderVdcByName finds a Provider VDC with extension (extended=true) or without extension (extended=false) by name
// On success, returns a pointer to the ProviderVdc (extended=false) or ProviderVdcExtended (extended=true) structure and a nil error
// On failure, returns a nil pointer and an error
func getProviderVdcByName(vcdClient *VCDClient, providerVdcName string, extended bool) (interface{}, error) {
	foundProviderVdcs, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "providerVdc",
		"filter":        fmt.Sprintf("name==%s", url.QueryEscape(providerVdcName)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}
	if len(foundProviderVdcs.Results.VMWProviderVdcRecord) == 0 {
		return nil, ErrorEntityNotFound
	}
	if len(foundProviderVdcs.Results.VMWProviderVdcRecord) > 1 {
		return nil, fmt.Errorf("more than one Provider VDC found with name '%s'", providerVdcName)
	}
	if extended {
		return vcdClient.GetProviderVdcExtendedByHref(foundProviderVdcs.Results.VMWProviderVdcRecord[0].HREF)
	}
	return vcdClient.GetProviderVdcByHref(foundProviderVdcs.Results.VMWProviderVdcRecord[0].HREF)
}

func (vcdClient *VCDClient) CreateProviderVdc(params *types.ProviderVdcCreation) (*ProviderVdcExtended, error) {

	if params.Name == "" {
		return nil, fmt.Errorf("a non-empty name is needed to create a provider VDC")
	}
	if len(params.ResourcePoolRefs.VimObjectRef) == 0 {
		return nil, fmt.Errorf("resource pool is needed to create a provider VDC")
	}
	if len(params.StorageProfile) == 0 {
		return nil, fmt.Errorf("storage profile is needed to create a provider VDC")
	}
	if params.VimServer == nil {
		return nil, fmt.Errorf("vim server is needed to create a provider VDC")
	}
	text := bytes.Buffer{}
	encoder := json.NewEncoder(&text)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(params)
	if err != nil {
		return nil, err
	}
	pvdcCreateHREF := vcdClient.Client.VCDHREF
	pvdcCreateHREF.Path += "/admin/extension/providervdcsparams"

	body := strings.NewReader(text.String())
	apiVersion := vcdClient.Client.APIVersion
	headAccept := http.Header{}
	headAccept.Set("Accept", fmt.Sprintf("application/*+json;version=%s", apiVersion))
	headAccept.Set("Content-Type", "application/*+json")
	request := vcdClient.Client.newRequest(nil, nil, http.MethodPost, pvdcCreateHREF, body, apiVersion, headAccept)
	resp, err := vcdClient.Client.Http.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var jsonError types.OpenApiError
		err = json.Unmarshal(body, &jsonError)
		// By default, we return the whole response body as error message. This may also contain the stack trace
		message := string(body)
		// if the body contains a valid JSON representation of the error, we return a more agile message, using the
		// exposed fields, and hiding the stack trace from view
		if err == nil {
			message = fmt.Sprintf("%s - %s", jsonError.MinorErrorCode, jsonError.Message)
		}
		return nil, fmt.Errorf("error creating provider VDC %s: %s (%d) - %s", params.Name, resp.Status, resp.StatusCode, message)
	}

	_, err = checkResp(resp, err)
	if err != nil {
		return nil, err
	}

	pvdc, err := vcdClient.GetProviderVdcExtendedByName(params.Name)
	if err != nil {
		return nil, err
	}

	return pvdc, nil
}
