package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"net/url"
)

type ProviderVdc struct {
	ProviderVdc *types.ProviderVdc
	client      *Client
}

func NewProviderVdc(cli *Client) *ProviderVdc {
	return &ProviderVdc{
		ProviderVdc: new(types.ProviderVdc),
		client:      cli,
	}
}

// GetProviderVdcByHref finds a Provider VDC by its HREF.
// On success, returns a pointer to the ProviderVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcByHref(providerVdcHref string) (*ProviderVdc, error) {
	providerVdc := NewProviderVdc(&vcdClient.Client)

	_, err := vcdClient.Client.ExecuteRequest(providerVdcHref, http.MethodGet,
		"", "error retrieving Provider VDC: %s", nil, providerVdc.ProviderVdc)
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

// GetProviderVdcByName finds a Provider VDC by name.
// On success, returns a pointer to the ProviderVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcByName(providerVdcName string) (*ProviderVdc, error) {
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
	return vcdClient.GetProviderVdcByHref(foundProviderVdcs.Results.VMWProviderVdcRecord[0].HREF)
}

// Refresh updates the contents of the Provider VDC associated to the receiver object.
func (providerVdc *ProviderVdc) Refresh() error {
	if providerVdc.ProviderVdc.HREF == "" {
		return fmt.Errorf("cannot refresh, receiver object is empty")
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
