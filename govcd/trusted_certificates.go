// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

const labelTrustedCertificate = "Trusted Certificate"

// TrustedCertificate manages certificate trust
type TrustedCertificate struct {
	TrustedCertificate *types.TrustedCertificate
	vcdClient          *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TrustedCertificate) wrap(inner *types.TrustedCertificate) *TrustedCertificate {
	g.TrustedCertificate = inner
	return &g
}

// AutoTrustHttpsCertificate will automatically trust certificate for a given HTTPS endpoint
// Note. The URL must be accessible
func (vcdClient *VCDClient) AutoTrustHttpsCertificate(endpoint *url.URL, ctx *TenantContext) (*TrustedCertificate, error) {
	port, err := getEndpointPort(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting port number for host '%s': %s", endpoint.Hostname(), err)
	}

	con := types.TestConnection{
		Host:                          endpoint.Hostname(),
		Port:                          port,
		Secure:                        addrOf(true),
		Timeout:                       10, // UI timeout value
		HostnameVerificationAlgorithm: "HTTPS",
	}

	res, err := vcdClient.Client.TestConnection(con)
	if err != nil {
		return nil, fmt.Errorf("error testing connection for %s: %s", endpoint.Hostname(), err)
	}

	var trustedCert *TrustedCertificate
	if res != nil && res.TargetProbe != nil && res.TargetProbe.SSLResult != "SUCCESS" {
		if res.TargetProbe.SSLResult == types.UntrustedCertificate {
			trustedCert, err = trustCertificate(vcdClient, ctx, endpoint.Hostname(), res.TargetProbe.CertificateChain)
			if err != nil {
				return nil, fmt.Errorf("could not trust certificate for %s, Connection result: '%s', SSL result: '%s': %s",
					endpoint.Hostname(), res.TargetProbe.ConnectionResult, res.TargetProbe.SSLResult, err)
			}
		} else {
			return nil, fmt.Errorf("SSL verification result - %s", res.TargetProbe.SSLResult)
		}
	}
	return trustedCert, nil
}

// trustCertificate trusts the given certificate for the given hostname with the given certificate chain
func trustCertificate(vcdClient *VCDClient, ctx *TenantContext, hostname string, certificateChain string) (*TrustedCertificate, error) {
	if certificateChain == "" {
		return nil, fmt.Errorf("certificate chain is empty")
	}

	// The CertificateChain may contain a single certificate or a chain of certificates.
	// In case of a single certificate - only it should be submitted.
	// In case of a chain - the last certificate is submitted to trust.
	certCount := strings.Count(certificateChain, "-----END CERTIFICATE-----")
	var trust *types.TrustedCertificate

	if certCount == 1 {
		// Certificate
		trust = &types.TrustedCertificate{
			Alias:       fmt.Sprintf("%s_%s", hostname, time.Now().UTC().Format(time.RFC3339)),
			Certificate: certificateChain,
		}
	} else {
		splitCerts := strings.SplitAfter(certificateChain, "-----END CERTIFICATE-----")
		trust = &types.TrustedCertificate{
			Alias:       fmt.Sprintf("ca_%s", time.Now().UTC().Format(time.RFC3339)),
			Certificate: splitCerts[len(splitCerts)-2],
		}
	}

	trustedCert, err := vcdClient.CreateTrustedCertificate(trust, ctx)
	if err != nil {
		return nil, fmt.Errorf("error trusting Certificate %s: %s", trust.Alias, err)
	}

	util.Logger.Printf("[DEBUG] Certificate trust established ID - %s, Alias - %s",
		trustedCert.TrustedCertificate.ID, trustedCert.TrustedCertificate.Alias)

	return trustedCert, nil
}

func getEndpointPort(u *url.URL) (int, error) {
	portStr := u.Port()
	if portStr == "" && strings.EqualFold(u.Scheme, "https") {
		portStr = "443"
	}

	if portStr == "" && strings.EqualFold(u.Scheme, "http") {
		portStr = "80"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("error converting port '%s' to int: %s", u.Port(), err)
	}

	return port, nil
}

// CreateTrustedCertificate creates an entry in the trusted certificate records with the given tenant context. If tenant context is nil,
// it assumes that the certificate will be stored in System org.
func (vcdClient *VCDClient) CreateTrustedCertificate(config *types.TrustedCertificate, ctx *TenantContext) (*TrustedCertificate, error) {
	c := crudConfig{
		entityLabel:      labelTrustedCertificate,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates,
		additionalHeader: getTenantContextHeader(ctx),
	}
	outerType := TrustedCertificate{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// CreateTrustedCertificate creates an entry in the trusted certificate records of the receiver Organization
func (org *TmOrg) CreateTrustedCertificate(config *types.TrustedCertificate) (*TrustedCertificate, error) {
	return org.vcdClient.CreateTrustedCertificate(config, &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	})
}

// GetAllTrustedCertificates retrieves all trusted certificates with optional query filter with the given tenant context. If tenant context is nil,
// it assumes that the certificates to get are stored in System org.
func (vcdClient *VCDClient) GetAllTrustedCertificates(queryParameters url.Values, ctx *TenantContext) ([]*TrustedCertificate, error) {
	c := crudConfig{
		entityLabel:      labelTrustedCertificate,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates,
		queryParameters:  queryParameters,
		additionalHeader: getTenantContextHeader(ctx),
	}

	outerType := TrustedCertificate{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetAllTrustedCertificates retrieves all trusted certificates with optional query filter from the receiver Organization
func (org *TmOrg) GetAllTrustedCertificates(queryParameters url.Values) ([]*TrustedCertificate, error) {
	return org.vcdClient.GetAllTrustedCertificates(queryParameters, &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	})
}

// GetTrustedCertificateByAlias retrieves trusted certificate by alias
func (vcdClient *VCDClient) GetTrustedCertificateByAlias(alias string, ctx *TenantContext) (*TrustedCertificate, error) {
	if alias == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTrustedCertificate)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "alias=="+alias)

	filteredEntities, err := vcdClient.GetAllTrustedCertificates(queryParams, ctx)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("alias", alias, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTrustedCertificateById(singleEntity.TrustedCertificate.ID, ctx)
}

// GetTrustedCertificateByAlias retrieves trusted certificate by alias from the receiver Organization
func (org *TmOrg) GetTrustedCertificateByAlias(alias string) (*TrustedCertificate, error) {
	return org.vcdClient.GetTrustedCertificateByAlias(alias, &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	})
}

// GetTrustedCertificateById retrieves trusted certificate by ID
func (vcdClient *VCDClient) GetTrustedCertificateById(id string, ctx *TenantContext) (*TrustedCertificate, error) {
	c := crudConfig{
		entityLabel:      labelTrustedCertificate,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates,
		endpointParams:   []string{id},
		additionalHeader: getTenantContextHeader(ctx),
	}

	outerType := TrustedCertificate{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetTrustedCertificateById retrieves trusted certificate by ID from the receiver Organization
func (org *TmOrg) GetTrustedCertificateById(id string) (*TrustedCertificate, error) {
	return org.vcdClient.GetTrustedCertificateById(id, &TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	})
}

// Update trusted certificate entry
func (t *TrustedCertificate) Update(TrustedCertificateConfig *types.TrustedCertificate) (*TrustedCertificate, error) {
	c := crudConfig{
		entityLabel:    labelTrustedCertificate,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates,
		endpointParams: []string{t.TrustedCertificate.ID},
	}
	outerType := TrustedCertificate{vcdClient: t.vcdClient}
	return updateOuterEntity(&t.vcdClient.Client, outerType, c, TrustedCertificateConfig)
}

// Delete trusted certificate entry
func (t *TrustedCertificate) Delete() error {
	c := crudConfig{
		entityLabel:    labelTrustedCertificate,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates,
		endpointParams: []string{t.TrustedCertificate.ID},
	}
	return deleteEntityById(&t.vcdClient.Client, c)
}
