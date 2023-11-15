package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

func (vcdClient *VCDClient) GetAllVcenterDistributedSwitches(vCenterId string, queryParameters url.Values) ([]*types.VcenterDistributedSwitch, error) {
	if vCenterId == "" {
		return nil, fmt.Errorf("empty vCenter ID")
	}

	if !isUrn(vCenterId) {
		return nil, fmt.Errorf("vCenter ID is not URN (e.g. 'urn:vcloud:vimserver:09722307-aee0-4623-af95-7f8e577c9ebc)', got: %s", vCenterId)
	}

	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVCenterDistributedSwitch
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}
	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("virtualCenter.id=="+vCenterId, queryParams)

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var typeResponses []*types.VcenterDistributedSwitch
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}
