// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// crudConfig contains configuration that must be supplied when invoking generic functions that are defined
// in `openapi_generic_inner_entities.go` and `openapi_generic_outer_entities.go`
type crudConfig struct {
	// Mandatory parameters

	// entityLabel contains friendly entity name that is used for logging meaningful errors
	entityLabel string

	// endpoint in the usual format (e.g. types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentIpDiscoveryProfiles)
	endpoint string

	// Optional parameters

	// endpointParams contains a slice of strings that will be used to construct the request URL. It will
	// initially replace '%s' placeholders in the `endpoint` (if any) and will add them as suffix
	// afterwards
	endpointParams []string

	// queryParameters will be passed as GET queries to the URL. Usually they are used for API filtering parameters
	queryParameters url.Values
	// additionalHeader can be used to pass additional headers for API calls. One of the common purposes is to pass
	// tenant context
	additionalHeader map[string]string
	// expectedStatus optionally restricts successful responses to one HTTP status.
	// Async helpers always require 202; waiting helpers accept any success status and wait on 202.
	expectedStatus int

	// minimumVersion overrides the endpoint-wide minimum for a TM operation.
	minimumVersion string
	tmOptions      []TmOperationOptions
	tenantContext  *TenantContext

	// requiresTm is a flag to signify if this resource works only in TM
	requiresTm bool
}

// validate should catch errors in consuming generic CRUD functions and should never produce false
// positives.
func (c crudConfig) validate(client *Client) error {
	// crudConfig misconfiguration - we can panic so that developer catches the problem during
	// development of this SDK
	if c.entityLabel == "" {
		panic("'entityName' must always be specified when initializing crudConfig")
	}

	if c.endpoint == "" {
		panic("'endpoint' must always be specified when initializing crudConfig")
	}

	if client == nil {
		return fmt.Errorf("OpenAPI client is nil")
	}

	// softer validations that consumers of this SDK can manipulate

	// If `endpointParams` is specified in `crudConfig` (it is not nil), then it must contain at
	// least a single non empty string parameter
	// Such validation should prevent cases where some ID is not speficied upon function call.
	// E.g.: endpointParams: []string{vdcId}, <--- vdcId comes from consumer of the SDK
	// If the user specified empty `vdcId` - we'd validate this
	for _, paramValue := range c.endpointParams {
		if paramValue == "" {
			return fmt.Errorf(`endpointParams were specified but they contain empty value "" for %s. %#v`,
				c.entityLabel, c.endpointParams)
		}
	}

	if c.requiresTm && !client.IsTm() {
		return fmt.Errorf("this entity requires TM")
	}

	return nil
}

// withTmOptions supplies per-call policy to the same CRUD implementation used by VCD.
func (c crudConfig) withTmOptions(options []TmOperationOptions) crudConfig {
	c.tmOptions = options
	return c
}

// retainTmOptions copies per-call policy so resource methods can reuse it without
// retaining a caller-owned slice or tenant context.
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func retainTmOptions(options []TmOperationOptions) []TmOperationOptions {
	result := append([]TmOperationOptions(nil), options...)
	for i := range result {
		if result[i].TenantContext != nil {
			context := *result[i].TenantContext
			result[i].TenantContext = &context
		}
	}
	return result
}

// inheritTmOptions applies explicit overrides to the policy saved on a resource.
// Invalid option counts are preserved for validation by crudConfig.prepare.
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func inheritTmOptions(saved, options []TmOperationOptions) []TmOperationOptions {
	if len(options) == 0 {
		return retainTmOptions(saved)
	}
	if len(options) != 1 || len(saved) != 1 {
		return retainTmOptions(options)
	}
	option := options[0]
	if option.APIVersion == "" {
		option.APIVersion = saved[0].APIVersion
	}
	if option.TenantContext == nil {
		option.TenantContext = saved[0].TenantContext
	}
	return retainTmOptions([]TmOperationOptions{option})
}

func (c crudConfig) item(id string) crudConfig {
	c.endpoint = strings.TrimSuffix(c.endpoint, "/") + "/%s"
	c.endpointParams = append(append([]string(nil), c.endpointParams...), id)
	return c
}

// prepare resolves version, URL and context once for all CRUD and action helpers.
// VCD keeps its endpoint version rules. TM can specify an operation minimum.
func (c crudConfig) prepare(client *Client) (openApiOperation, error) {
	if len(c.tmOptions) > 1 {
		return openApiOperation{}, fmt.Errorf("at most one TmOperationOptions value is allowed")
	}
	var option TmOperationOptions
	if len(c.tmOptions) == 1 {
		option = c.tmOptions[0]
	}
	if err := c.validate(client); err != nil {
		return openApiOperation{}, err
	}
	params := c.endpointParams
	var version string
	var err error
	if c.minimumVersion != "" {
		if err = client.vcdFetchSupportedVersions(); err != nil {
			return openApiOperation{}, err
		}
		version, err = selectTmApiVersion(client.supportedVersions, c.minimumVersion, client.APIVersion, option.APIVersion)
		// Existing generic callers may pass path fragments; operation-specific IDs
		// are individual path segments and must be escaped.
		params = make([]string, len(c.endpointParams))
		for i, id := range c.endpointParams {
			params[i] = url.PathEscape(id)
		}
	} else {
		version, err = client.getOpenApiHighestElevatedVersion(c.endpoint)
	}
	if err != nil {
		return openApiOperation{}, fmt.Errorf("error selecting API version for %s: %s", c.entityLabel, err)
	}
	endpoint, err := urlFromEndpoint(c.endpoint, params)
	if err != nil {
		return openApiOperation{}, err
	}
	uri, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return openApiOperation{}, err
	}
	headers := c.additionalHeader
	ctx := c.tenantContext
	if option.TenantContext != nil {
		ctx = option.TenantContext
	}
	if ctx != nil {
		headers = make(map[string]string, len(c.additionalHeader)+2)
		for key, value := range c.additionalHeader {
			if !strings.EqualFold(key, types.HeaderTenantContext) && !strings.EqualFold(key, types.HeaderAuthContext) {
				headers[key] = value
			}
		}
		for key, value := range getTenantContextHeader(ctx) {
			headers[key] = value
		}
	}
	requestClient := client
	if c.requiresTm && (ctx != nil || headers[types.HeaderTenantContext] != "" || headers[types.HeaderAuthContext] != "") {
		// A selected context replaces inherited headers, including an empty System context.
		cloned := *client
		cloned.customHeader = client.customHeader.Clone()
		cloned.customHeader.Del(types.HeaderTenantContext)
		cloned.customHeader.Del(types.HeaderAuthContext)
		requestClient = &cloned
	}
	op := requestClient.openApiOperation(version, uri, c.queryParameters, headers)
	if c.expectedStatus != 0 {
		op.expectedStatus = []int{c.expectedStatus}
	}
	// Preserve legacy VCD lookup semantics; operation-specific TM reads keep 403 errors.
	op.forbiddenAsNotFound = c.minimumVersion == ""
	return op, nil
}

// createInnerEntity creates a resource, waiting for a task and retrieving its owner when needed.
func createInnerEntity[I any](client *Client, c crudConfig, config *I) (*I, error) {
	return writeInnerEntity(client, c, http.MethodPost, config)
}

// createInnerEntityAsync returns the tracking task without waiting for creation.
func createInnerEntityAsync[I any](client *Client, c crudConfig, config *I) (*Task, error) {
	return runOpenApiOperationAsync(client, c, http.MethodPost, config)
}

func updateInnerEntity[I any](client *Client, c crudConfig, config *I) (*I, error) {
	result, _, err := updateInnerEntityWithHeaders(client, c, config)
	return result, err
}

func updateInnerEntityWithHeaders[I any](client *Client, c crudConfig, config *I) (*I, http.Header, error) {
	return writeInnerEntityWithHeaders(client, c, http.MethodPut, config)
}

// writeInnerEntity also handles writes that return a result page instead of a resource.
func writeInnerEntity[I any](client *Client, c crudConfig, method string, config *I) (*I, error) {
	result, _, err := writeInnerEntityWithHeaders(client, c, method, config)
	return result, err
}

func writeInnerEntityWithHeaders[I any](client *Client, c crudConfig, method string, config *I) (*I, http.Header, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("%s config must not be nil", c.entityLabel)
	}
	op, err := c.prepare(client)
	if err != nil {
		return nil, nil, err
	}
	result := new(I)
	headers, err := op.write(method, config, result)
	if err != nil {
		return nil, nil, fmt.Errorf("error writing entity of type '%s': %s", c.entityLabel, err)
	}
	return result, headers, nil
}

func getInnerEntity[I any](client *Client, c crudConfig) (*I, error) {
	result, _, err := getInnerEntityWithHeaders[I](client, c)
	return result, err
}

func getInnerEntityWithHeaders[I any](client *Client, c crudConfig) (*I, http.Header, error) {
	op, err := c.prepare(client)
	if err != nil {
		return nil, nil, err
	}
	result := new(I)
	headers, err := op.read(result)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving entity of type '%s': %s", c.entityLabel, err)
	}
	return result, headers, nil
}

func getAllInnerEntities[I any](client *Client, c crudConfig) ([]*I, error) {
	op, err := c.prepare(client)
	if err != nil {
		return nil, err
	}
	result := make([]*I, 0)
	// Pagination adds pageSize/page parameters; do not modify the caller's query.
	query := make(url.Values, len(c.queryParameters))
	for key, values := range c.queryParameters {
		query[key] = append([]string(nil), values...)
	}
	if err := op.client.OpenApiGetAllItems(op.apiVersion, op.url, query, &result, op.headers); err != nil {
		return nil, fmt.Errorf("error retrieving all entities of type '%s': %s", c.entityLabel, err)
	}
	return result, nil
}

func deleteEntityById(client *Client, c crudConfig) error {
	op, err := c.prepare(client)
	if err != nil {
		return err
	}
	_, err = op.write(http.MethodDelete, nil, nil)
	return err
}

// runOpenApiOperationAsync handles actions and association writes without assuming
// that the task owner identifies a resource to retrieve afterward.
func runOpenApiOperationAsync(client *Client, c crudConfig, method string, payload interface{}) (*Task, error) {
	op, err := c.prepare(client)
	if err != nil {
		return nil, err
	}
	return op.runAsync(method, payload)
}

func runOpenApiOperation(client *Client, c crudConfig, method string, payload interface{}) error {
	op, err := c.prepare(client)
	if err != nil {
		return err
	}
	_, err = op.write(method, payload, nil)
	return err
}

func urlFromEndpoint(endpoint string, endpointParams []string) (string, error) {
	// Count how many '%s' placeholders exist in the 'endpoint'
	placeholderCount := strings.Count(endpoint, "%s")

	// Validation. At the very minimum all placeholders must have their replacements - otherwise it
	// is an error as we never want to query an endpoint that still has placeholders '%s'
	if len(endpointParams) < placeholderCount {
		return "", fmt.Errorf("endpoint '%s' has unpopulated placeholders", endpoint)
	}

	// if there are no 'endpointParams' - exit with the same endpoint
	if len(endpointParams) == 0 {
		return endpoint, nil
	}

	// Loop over given endpointParams and replace placeholders at first. Afterwards - amend any
	// additional parameters to the end of endpoint
	for _, v := range endpointParams {
		// If there are placeholders '%s' to replace in the endpoint itself - do it
		if placeholderCount > 0 {
			endpoint = strings.Replace(endpoint, "%s", v, 1)
			placeholderCount = placeholderCount - 1
			continue
		}

		endpoint = endpoint + v
	}

	return endpoint, nil
}

// TmOperationOptions overrides the request version or organization context for a TM operation.
// These options are optional and do not change the client's default version or context.
type TmOperationOptions struct {
	// APIVersion must be advertised by the server and meet the operation's minimum version.
	// When empty, use the client's version if eligible, otherwise the earliest eligible stable version.
	APIVersion string
	// TenantContext overrides the context retained by a resource wrapper, if provided.
	// An explicit System context clears inherited tenant headers, including during task polling.
	TenantContext *TenantContext
}
