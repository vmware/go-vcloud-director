/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// GetFederationSettings retrieves the current federation (SAML) settings for a given organization
func (adminOrg *AdminOrg) GetFederationSettings() (*types.OrgFederationSettings, error) {
	var settings types.OrgFederationSettings

	if adminOrg.AdminOrg.OrgSettings == nil || adminOrg.AdminOrg.OrgSettings.Link == nil {
		return nil, fmt.Errorf("no Org settings links found in Org %s", adminOrg.AdminOrg.Name)
	}
	fsUrl := getUrlFromLink(adminOrg.AdminOrg.OrgSettings.Link, "down", types.MimeFederationSettingsXml)
	if fsUrl == "" {
		return nil, fmt.Errorf("no link found for federation settings (SAML: %s) in Org %s", types.MimeFederationSettingsXml, adminOrg.AdminOrg.Name)
	}

	resp, err := adminOrg.client.ExecuteRequest(fsUrl, http.MethodGet, types.MimeFederationSettingsXml,
		"error fetching federation settings: %s", nil, &settings)

	if err != nil {
		return nil, err
	}

	_, err = checkResp(resp, err)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// SetFederationSettings creates or replaces federation (SAML) settings for a given organization
func (adminOrg *AdminOrg) SetFederationSettings(settings *types.OrgFederationSettings) (*types.OrgFederationSettings, error) {

	if adminOrg.AdminOrg.OrgSettings == nil || adminOrg.AdminOrg.OrgSettings.Link == nil {
		return nil, fmt.Errorf("no Org settings links found in Org %s", adminOrg.AdminOrg.Name)
	}
	fsUrl := getUrlFromLink(adminOrg.AdminOrg.OrgSettings.Link, "down", types.MimeFederationSettingsJson)
	if fsUrl == "" {
		return nil, fmt.Errorf("no URL found for federation settings (SAML) in Org %s", adminOrg.AdminOrg.Name)
	}

	setUrl, err := url.Parse(fsUrl)
	if err != nil {
		return nil, err
	}

	text := bytes.Buffer{}
	encoder := json.NewEncoder(&text)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(settings)
	if err != nil {
		return nil, err
	}
	body := strings.NewReader(text.String())
	apiVersion := adminOrg.client.APIVersion
	headAccept := http.Header{}
	// NOTE: given that the UI uses JSON based API to run SAML settings, it seemed the safest way to
	// imitate it and use JSON payload and results for this operation
	headAccept.Set("Accept", types.JSONMime)
	headAccept.Set("Content-Type", types.MimeFederationSettingsJson)
	request := adminOrg.client.newRequest(nil, nil, http.MethodPut, *setUrl, body, apiVersion, headAccept)
	request.Header.Set("Accept", fmt.Sprintf("application/*+json;version=%s", apiVersion))
	request.Header.Set("Content-Type", types.MimeFederationSettingsJson)

	resp, err := adminOrg.client.Http.Do(request)
	if err != nil {
		return nil, err
	}

	if !isSuccessStatus(resp.StatusCode) {
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
		return nil, fmt.Errorf("error setting SAML for org %s: %s (%d) - %s", adminOrg.AdminOrg.Name, resp.Status, resp.StatusCode, message)
	}

	_, err = checkResp(resp, err)
	if err != nil {
		return nil, err
	}

	return adminOrg.GetFederationSettings()
}

// UnsetFederationSettings removes federation (SAML) settings for a given organization
func (adminOrg *AdminOrg) UnsetFederationSettings() error {
	settings, err := adminOrg.GetFederationSettings()
	if err != nil {
		return fmt.Errorf("[UnsetFederationSettings] error getting SAML settings for Org %s: %s", adminOrg.AdminOrg.Name, err)
	}

	settings.SAMLMetadata = ""
	settings.Enabled = false
	_, err = adminOrg.SetFederationSettings(settings)
	return err
}

// GetServiceProviderSamlMetadata retrieves the service provider SAML metadata of the given Org
func (adminOrg *AdminOrg) GetServiceProviderSamlMetadata() (*types.VcdSamlMetadata, error) {

	metadataText, err := adminOrg.RetrieveServiceProviderSamlMetadata()
	if err != nil {
		return nil, err
	}
	var metadata types.VcdSamlMetadata

	err = xml.Unmarshal([]byte(metadataText), &metadata)
	if err != nil {
		return nil, fmt.Errorf("[GetSamlMetadata] error decoding metadata retrieved from %s: %s", adminOrg.AdminOrg.Name, err)
	}

	return &metadata, nil
}

// RetrieveServiceProviderSamlMetadata retrieves the SAML metadata of the given Org
func (adminOrg *AdminOrg) RetrieveServiceProviderSamlMetadata() (string, error) {

	settings, err := adminOrg.GetFederationSettings()
	if err != nil {
		return "", err
	}
	metadataUrl := getUrlFromLink(settings.Link, "down", types.MimeSamlMetadata)
	if metadataUrl == "" {
		return "", fmt.Errorf("[RetrieveRemoteDocument] no URL found for metadata retrieval (%s) in org %s", types.MimeSamlMetadata, adminOrg.AdminOrg.Name)
	}

	metadataText, err := adminOrg.client.RetrieveRemoteDocument(metadataUrl)
	if err != nil {
		return "", fmt.Errorf("[RetrieveRemoteDocument] error retrieving SAML metadata from %s: %s", metadataUrl, err)
	}
	return string(metadataText), nil
}

func getUrlFromLink(linkList types.LinkList, wantRel, wantType string) string {
	for _, link := range linkList {
		if link.Rel == wantRel && link.Type == wantType {
			return link.HREF
		}
	}
	return ""
}

var (
	// samlMetadataItems contains name space identifiers and corresponding tags
	// that should be found in VCD SAML service provider metadata
	samlMetadataItems = map[string][]string{
		"ds": {
			"KeyInfo",
			"X509Certificate",
			"X509Data",
		},
		"md": {
			"AssertionConsumerService",
			"EntityDescriptor",
			"KeyDescriptor",
			"NameIDFormat",
			"SPSSODescriptor",
			"SingleLogoutService",
		},
		"hoksso": {
			"ProtocolBinding",
		},
	}
)

// RetrieveRemoteDocument gets the contents of a given URL
func (client *Client) RetrieveRemoteDocument(metadataUrl string) ([]byte, error) {

	retrieveUrl, err := url.Parse(metadataUrl)
	if err != nil {
		return nil, err
	}
	request := client.newRequest(nil, nil, http.MethodGet, *retrieveUrl, nil, client.APIVersion, nil)

	resp, err := client.Http.Do(request)

	if err != nil {
		return nil, fmt.Errorf("[RetrieveRemoteDocument] error retrieving metadata from %s: %s", metadataUrl, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[RetrieveRemoteDocument] error reading response body from metadata retrieved from %s: %s", metadataUrl, err)
	}

	util.ProcessResponseOutput("[RetrieveRemoteDocument]", resp, string(body))
	return body, nil
}

// normalizeServiceProviderSamlMetadata takes a string containing the XML code with Metadata definition
// and makes sure it has all the expected elements
func normalizeServiceProviderSamlMetadata(in string) (string, error) {
	var metadata types.VcdSamlMetadata

	// Phase 1: Decode the XML, to find possible encoding errors
	err := xml.Unmarshal([]byte(in), &metadata)
	if err != nil {
		return "", fmt.Errorf("[normalizeSamlMetadata] error decoding SAML metadata definition from XML: %s", err)
	}

	// Phase 2: Add the namespace definition elements, required to recognize the structure as a valid SAML definition
	metadata.Md = types.SamlNamespaceMd
	metadata.SPSSODescriptor.Ds = types.SamlNamespaceDs
	for i := 0; i < len(metadata.SPSSODescriptor.AssertionConsumerService); i++ {
		metadata.SPSSODescriptor.AssertionConsumerService[i].Hoksso = types.SamlNamespaceHoksso
	}

	// Phase 3: Convert the data structure to text again. The text now includes the needed namespace definition elements
	out, err := xml.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("[normalizeSamlMetadata] error encoding SAML metadata text: %s", err)
	}

	// Phase 4: Add the namespace elements to the XML text
	metadataText := string(out)
	for ns, fields := range samlMetadataItems {
		if !strings.Contains(metadataText, ns) {
			return metadataText, fmt.Errorf("[normalizeSamlMetadata] namespace '%s' not found in SAML metadata", ns)
		}
		for _, fieldName := range fields {
			fullName := fmt.Sprintf("%s:%s", ns, fieldName)
			// If we find just "FieldName", but not "namespace:FieldName", then we replace the bare FieldName with the full identifier
			if strings.Contains(metadataText, fieldName) && !strings.Contains(metadataText, fullName) {
				metadataText = strings.Replace(metadataText, fieldName, fullName, -1)
			}
		}
	}

	return metadataText, nil
}

// validateNamespaceDefinition checks that a metadata XML text contains the expected namespace definition
func validateNamespaceDefinition(metadataText string, namespace string) bool {
	reEmptyDefinition := regexp.MustCompile(`xmlns:` + namespace + `\s*=\s*""`)
	reFilledDefinition := regexp.MustCompile(`xmlns:` + namespace + `\s*=\s*"\S+"`)
	// Check that the namespace is mentioned at all in the metadata text
	if !strings.Contains(metadataText, namespace) {
		return false
	}
	// Check that an empty namespace definition is NOT found in the metadata text
	// (for example: xmlns:md="")
	if reEmptyDefinition.FindString(metadataText) != "" {
		return false
	}
	// Check that a filled namespace definition is found in the metadata text
	// (for example: xmlns:md="something")
	found := reFilledDefinition.FindString(metadataText)
	return found != ""
}

// ValidateSamlServiceProviderMetadata tells whether a given string contains valid XML that defines SAML service provider metadata
// Returns nil on valid data, and an array of errors for invalid data
func ValidateSamlServiceProviderMetadata(metadataText string) []error {
	var metadata types.VcdSamlMetadata
	var errors []error

	// Check n. 1: encode the string into XML, thus establishing that it is valid syntax
	err := xml.Unmarshal([]byte(metadataText), &metadata)
	if err != nil {
		errors = append(errors, fmt.Errorf("[ValidateSamlMetadata] error decoding XML into SAML metadata structure: %s", err))
	}

	reNameSpace, err := regexp.Compile(`<(\w+):(\w+)`)

	if err != nil {
		errors = append(errors, fmt.Errorf("error compiling regular expression: %s", err))
		return errors
	}

	nsInfoList := reNameSpace.FindAllStringSubmatch(metadataText, -1)
	processed := map[string]bool{}

	// Check n. 2: make sure that each namespace used in the metadata text has a corresponding definition
	for _, nsInfo := range nsInfoList {
		seen, ok := processed[nsInfo[0]]
		if ok && seen {
			continue
		}
		ns := nsInfo[1]
		if !validateNamespaceDefinition(metadataText, ns) {
			errors = append(errors, fmt.Errorf("[ValidateSamlMetadata] namespace '%s' undefined in SAML metadata", ns))
		}
		processed[nsInfo[0]] = true
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

// GetErrorMessageFromErrorSlice returns a single error message from a list of error
func GetErrorMessageFromErrorSlice(errors []error) string {
	result := ""
	for i, err := range errors {
		result = fmt.Sprintf("%s\n%2d %s", result, i, err)
	}
	return result
}
