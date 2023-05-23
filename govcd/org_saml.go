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

	resp, err = checkResp(resp, err)
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
	var newSettings types.OrgFederationSettings

	setUrl, err := url.Parse(fsUrl)
	if err != nil {
		return nil, err
	}

	settings.SAMLMetadata, err = normalizeSamlMetadata(settings.SAMLMetadata)
	if err != nil {
		return nil, fmt.Errorf("error normalising SAML metadata: %s", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%s\n", settings.SAMLMetadata)
	fmt.Println(strings.Repeat("=", 80))
	text := bytes.Buffer{}
	encoder := json.NewEncoder(&text)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(settings)
	if err != nil {
		return nil, err
	}
	fmt.Println(strings.Repeat("+", 80))
	fmt.Printf("%s\n", text.String())
	fmt.Println(strings.Repeat("+", 80))
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

	resp, err = checkResp(resp, err)
	if err != nil {
		return nil, err
	}

	return &newSettings, nil
}

func (adminOrg *AdminOrg) GetSamlMetadata() (*types.VcdSamlMetadata, error) {

	settings, err := adminOrg.GetFederationSettings()
	if err != nil {
		return nil, err
	}
	metadataUrl := getUrlFromLink(settings.Link, "down", types.MimeSamlMetadata)
	if metadataUrl == "" {
		return nil, fmt.Errorf("no URL found for metadata retrieval (%s) in org %s", types.MimeSamlMetadata, adminOrg.AdminOrg.Name)
	}

	var metadata types.VcdSamlMetadata

	resp, err := adminOrg.client.ExecuteRequest(metadataUrl, http.MethodGet, types.MimeSamlMetadata,
		"error getting metadata: %s", nil, &metadata)

	if err != nil {
		return nil, err
	}

	resp, err = checkResp(resp, err)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
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

// normalizeSamlMetadata takes a string containing the XML code with Metadata definition
// and makes sure it has all the expected elements
func normalizeSamlMetadata(in string) (string, error) {
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

	// Phase 5: Add the XML header if it is missing
	if !strings.Contains(metadataText, types.XmlHeader) {
		metadataText = types.XmlHeader + metadataText
	}

	return metadataText, nil
}

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

// ValidateSamlMetadata tells whether a given string contains valid XML that defines SAML metadata
// Returns nil on valid data, and an array of errors for invalid data
func ValidateSamlMetadata(metadataText string) []error {
	var metadata types.VcdSamlMetadata
	var errors []error

	// Check n. 1: encode the string into XML, thus establishing that it is valid syntax
	err := xml.Unmarshal([]byte(metadataText), &metadata)
	if err != nil {
		errors = append(errors, fmt.Errorf("[ValidateSamlMetadata] error decoding XML into SAML metadata structure: %s", err))
	}

	// Check n. 2: make sure the input contains a XML header
	if !strings.Contains(metadataText, types.XmlHeader) {
		errors = append(errors, fmt.Errorf("[ValidateSamlMetadata] SAML metadata does not include a XML header"))
	}

	// Check n. 3: make sure the namespaces are defined
	for _, ns := range []string{"md", "ds", "hoksso"} {
		if !validateNamespaceDefinition(metadataText, ns) {
			errors = append(errors, fmt.Errorf("[ValidateSamlMetadata] namespace '%s' undefined in SAML metadata", ns))
		}
	}

	// Check n. 4: make sure the expected fields have their namespace included
	for ns, fields := range samlMetadataItems {
		for _, fieldName := range fields {
			fullName := fmt.Sprintf("%s:%s", ns, fieldName)
			if !strings.Contains(metadataText, fullName) {
				errors = append(errors, fmt.Errorf("[ValidateSamlMetadata] field %s not found in SAML metadata", fullName))
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func GetErrorMessageFromErrorSlice(errors []error) string {
	result := ""
	for i, err := range errors {
		result = fmt.Sprintf("%s\n%2d %s", result, i, err)
	}
	return result
}
