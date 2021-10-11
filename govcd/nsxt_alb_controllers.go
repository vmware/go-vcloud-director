/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtAlbController helps to integrate VMware Cloud Director with NSX-T Advanced Load Balancer deployment.
// Controller instances are registered with VMware Cloud Director instance. Controller instances serve as a central
// control plane for the load-balancing services provided by NSX-T Advanced Load Balancer.
// To configure an NSX-T ALB one needs to supply AVI Controller endpoint, credentials and license to be used.
type NsxtAlbController struct {
	NsxtAlbController *types.NsxtAlbController
	vcdClient         *VCDClient
}

// GetAllAlbControllers returns all configured NSX-T ALB Controllers
func (vcdClient *VCDClient) GetAllAlbControllers(queryParameters url.Values) ([]*NsxtAlbController, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, errors.New("reading NSX-T ALB Controllers require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtAlbController{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtAlbController types with client
	wrappedResponses := make([]*NsxtAlbController, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAlbController{
			NsxtAlbController: typeResponses[sliceIndex],
			vcdClient:         vcdClient,
		}
	}

	return wrappedResponses, nil
}

// GetAlbControllerByName returns NSX-T ALB Controller by Name
func (vcdClient *VCDClient) GetAlbControllerByName(name string) (*NsxtAlbController, error) {
	queryParameters := copyOrNewUrlValues(nil)
	queryParameters.Add("filter", "name=="+name)

	controllers, err := vcdClient.GetAllAlbControllers(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error reading ALB Controller with Name '%s': %s", name, err)
	}

	if len(controllers) == 0 {
		return nil, fmt.Errorf("%s: could not find ALB Controller with Name '%s'", ErrorEntityNotFound, name)
	}

	if len(controllers) > 1 {
		return nil, fmt.Errorf("found more than 1 ALB Controller with Name '%s'", name)
	}

	return controllers[0], nil
}

// GetAlbControllerById returns NSX-T ALB Controller by ID
func (vcdClient *VCDClient) GetAlbControllerById(id string) (*NsxtAlbController, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, errors.New("reading NSX-T ALB Controllers require System user")
	}

	if id == "" {
		return nil, fmt.Errorf("ID is required to lookup NSX-T ALB Controller by ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	typeResponse := &types.NsxtAlbController{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponse := &NsxtAlbController{
		NsxtAlbController: typeResponse,
		vcdClient:         vcdClient,
	}

	return wrappedResponse, nil
}

// GetAlbControllerByUrl returns configured ALB Controller by URL
//
// Note. Filtering is performed on client side.
func (vcdClient *VCDClient) GetAlbControllerByUrl(url string) (*NsxtAlbController, error) {
	// Ideally this function could filter on VCD side, but API does not support filtering on URL
	controllers, err := vcdClient.GetAllAlbControllers(nil)
	if err != nil {
		return nil, fmt.Errorf("error reading ALB Controller with Url '%s': %s", url, err)
	}

	// Search for controllers
	filteredControllers := make([]*NsxtAlbController, 0)
	for _, controller := range controllers {
		if controller.NsxtAlbController.Url == url {
			filteredControllers = append(filteredControllers, controller)
		}
	}

	if len(filteredControllers) == 0 {
		return nil, fmt.Errorf("%s could not find ALB Controller by Url '%s'", ErrorEntityNotFound, url)
	}

	if len(filteredControllers) > 1 {
		return nil, fmt.Errorf("found more than 1 ALB Controller by Url '%s'", url)
	}

	return filteredControllers[0], nil
}

// CreateNsxtAlbController creates controller with supplied albControllerConfig configuration
func (vcdClient *VCDClient) CreateNsxtAlbController(albControllerConfig *types.NsxtAlbController) (*NsxtAlbController, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, errors.New("handling NSX-T ALB Controllers require System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtAlbController{
		NsxtAlbController: &types.NsxtAlbController{},
		vcdClient:         vcdClient,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, albControllerConfig, returnObject.NsxtAlbController, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T ALB Controller: %s", err)
	}

	return returnObject, nil
}

// Update updates existing NSX-T ALB Controller with new supplied albControllerConfig configuration
func (nsxtAlbController *NsxtAlbController) Update(albControllerConfig *types.NsxtAlbController) (*NsxtAlbController, error) {
	client := nsxtAlbController.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if albControllerConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T ALB Controller without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, albControllerConfig.ID)
	if err != nil {
		return nil, err
	}

	responseAlbController := &NsxtAlbController{
		NsxtAlbController: &types.NsxtAlbController{},
		vcdClient:         nsxtAlbController.vcdClient,
	}

	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, albControllerConfig, responseAlbController.NsxtAlbController, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T ALB Controller: %s", err)
	}

	return responseAlbController, nil
}

// Delete deletes existing NSX-T ALB Controller
func (nsxtAlbController *NsxtAlbController) Delete() error {
	client := nsxtAlbController.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if nsxtAlbController.NsxtAlbController.ID == "" {
		return fmt.Errorf("cannot delete NSX-T ALB Controller without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, nsxtAlbController.NsxtAlbController.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T ALB Controller: %s", err)
	}

	return nil
}
