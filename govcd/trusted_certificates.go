package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelTrustedCertificate = "Trusted Certificate"

// TrustedCertificate manages certificate trust. Certificate
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

// CreateTrustedCertificate creates an entry in the trusted certificate records
func (vcdClient *VCDClient) CreateTrustedCertificate(config *types.TrustedCertificate) (*TrustedCertificate, error) {
	c := crudConfig{
		entityLabel: labelTrustedCertificate,
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates,
	}
	outerType := TrustedCertificate{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllTrustedCertificates retrieves all trusted certificates with optional query filter
func (vcdClient *VCDClient) GetAllTrustedCertificates(queryParameters url.Values) ([]*TrustedCertificate, error) {
	c := crudConfig{
		entityLabel:     labelTrustedCertificate,
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates,
		queryParameters: queryParameters,
	}

	outerType := TrustedCertificate{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTrustedCertificateByAlias retrieves trusted certificate by alias
func (vcdClient *VCDClient) GetTrustedCertificateByAlias(alias string) (*TrustedCertificate, error) {
	if alias == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTrustedCertificate)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "alias=="+alias)

	filteredEntities, err := vcdClient.GetAllTrustedCertificates(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("alias", alias, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTrustedCertificateById(singleEntity.TrustedCertificate.ID)
}

// GetTrustedCertificateById retrieves trusted certificate by ID
func (vcdClient *VCDClient) GetTrustedCertificateById(id string) (*TrustedCertificate, error) {
	c := crudConfig{
		entityLabel:    labelTrustedCertificate,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates,
		endpointParams: []string{id},
	}

	outerType := TrustedCertificate{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
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
