/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

// Certificate In UI called menu Certificate Library. In API Certificate Library item
type Certificate struct {
	CertificateLibrary *types.CertificateLibraryItem
	Href               string
	client             *Client
}

// GetCertificateFromLibraryById Returns certificate from library of certificates
func getCertificateFromLibraryById(client *Client, id string, additionalHeader map[string]string) (*Certificate, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibrary
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty certificate id")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	certificate := &Certificate{
		CertificateLibrary: &types.CertificateLibraryItem{},
		client:             client,
		Href:               urlRef.String(),
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, certificate.CertificateLibrary, additionalHeader)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

// GetCertificateFromLibraryById Returns certificate from library of certificates
func (client *Client) GetCertificateFromLibraryById(id string) (*Certificate, error) {
	return getCertificateFromLibraryById(client, id, nil)
}

// GetCertificateFromLibraryById Returns certificate from library of certificates
func (adminOrg *AdminOrg) GetCertificateFromLibraryById(id string) (*Certificate, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	return getCertificateFromLibraryById(adminOrg.client, id, getTenantContextHeader(tenantContext))
}

// addCertificateToLibrary uploads certificates with configuration details
func addCertificateToLibrary(client *Client, certificateConfig *types.CertificateLibraryItem,
	additionalHeader map[string]string) (*Certificate, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibrary
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponse := &Certificate{
		CertificateLibrary: &types.CertificateLibraryItem{},
		client:             client,
		Href:               urlRef.String(),
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil,
		certificateConfig, typeResponse.CertificateLibrary, additionalHeader)
	if err != nil {
		return nil, err
	}

	return typeResponse, nil
}

// AddCertificateToLibrary uploads certificates with configuration details
func (adminOrg *AdminOrg) AddCertificateToLibrary(certificateConfig *types.CertificateLibraryItem) (*Certificate, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	return addCertificateToLibrary(adminOrg.client, certificateConfig, getTenantContextHeader(tenantContext))
}

// AddCertificateToLibrary uploads certificates with configuration details
func (client *Client) AddCertificateToLibrary(certificateConfig *types.CertificateLibraryItem) (*Certificate, error) {
	return addCertificateToLibrary(client, certificateConfig, nil)
}

// getAllCertificateFromLibrary retrieves all certificates. Query parameters can be supplied to perform additional
// filtering
func getAllCertificateFromLibrary(client *Client, queryParameters url.Values, additionalHeader map[string]string) ([]*Certificate, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibrary
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	responses := []*types.CertificateLibraryItem{{}}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses, additionalHeader)
	if err != nil {
		return nil, err
	}

	var wrappedCertificates []*Certificate
	for _, response := range responses {
		urlRef, err := client.OpenApiBuildEndpoint(endpoint, response.Id)
		if err != nil {
			return nil, err
		}
		wrappedCertificate := &Certificate{
			CertificateLibrary: response,
			client:             client,
			Href:               urlRef.String(),
		}
		wrappedCertificates = append(wrappedCertificates, wrappedCertificate)
	}

	return wrappedCertificates, nil
}

// GetAllCertificateFromLibrary retrieves all available certificates from certificate library.
// Query parameters can be supplied to perform additional filtering
func (client *Client) GetAllCertificateFromLibrary(queryParameters url.Values) ([]*Certificate, error) {
	return getAllCertificateFromLibrary(client, queryParameters, nil)
}

// GetAllCertificatesFromLibrary r retrieves all available certificates from certificate library.
// Query parameters can be supplied to perform additional filtering
func (adminOrg *AdminOrg) GetAllCertificatesFromLibrary(queryParameters url.Values) ([]*Certificate, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	return getAllCertificateFromLibrary(adminOrg.client, queryParameters, getTenantContextHeader(tenantContext))
}

// getCertificateFromLibraryByName retrieves certificate from certificate library by given name
func getCertificateFromLibraryByName(client *Client, name string, additionalHeader map[string]string) (*Certificate, error) {
	var params = url.Values{}

	params.Set("filterEncoded", "true")
	params.Set("filter", fmt.Sprintf("alias==%s", url.QueryEscape(name)))
	certificates, err := getAllCertificateFromLibrary(client, params, additionalHeader)
	if err != nil {
		return nil, err
	}
	if len(certificates) == 0 {
		return nil, ErrorEntityNotFound
	}

	if len(certificates) > 1 {
		return nil, fmt.Errorf("more than one certificate found with name '%s'", name)
	}
	return certificates[0], nil
}

// GetCertificateFromLibraryByName retrieves certificate from certificate library by given name
func (client *Client) GetCertificateFromLibraryByName(name string) (*Certificate, error) {
	return getCertificateFromLibraryByName(client, name, nil)
}

// GetCertificateFromLibraryByName retrieves certificate from certificate library by given name
func (adminOrg *AdminOrg) GetCertificateFromLibraryByName(name string) (*Certificate, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	return getCertificateFromLibraryByName(adminOrg.client, name, getTenantContextHeader(tenantContext))
}

// Update updates existing Certificate. Allows changing only alias and description
func (certificate *Certificate) Update() (*Certificate, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibrary
	minimumApiVersion, err := certificate.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if certificate.CertificateLibrary.Id == "" {
		return nil, fmt.Errorf("cannot update certificate without id")
	}

	urlRef, err := certificate.client.OpenApiBuildEndpoint(endpoint, certificate.CertificateLibrary.Id)
	if err != nil {
		return nil, err
	}

	returnCertificate := &Certificate{
		CertificateLibrary: &types.CertificateLibraryItem{},
		client:             certificate.client,
	}

	err = certificate.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, certificate.CertificateLibrary,
		returnCertificate.CertificateLibrary, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating certificate: %s", err)
	}

	return returnCertificate, nil
}

// Delete deletes certificate from Certificate library
func (certificate *Certificate) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibrary
	minimumApiVersion, err := certificate.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if certificate.CertificateLibrary.Id == "" {
		return fmt.Errorf("cannot delete certificate without id")
	}

	urlRef, err := certificate.client.OpenApiBuildEndpoint(endpoint, certificate.CertificateLibrary.Id)
	if err != nil {
		return err
	}

	err = certificate.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting certificate: %s", err)
	}

	return nil
}