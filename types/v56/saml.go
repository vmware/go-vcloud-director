/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

import (
	"encoding/xml"
	"fmt"
)

// VcdSamlMetadata helps to marshal vCD SAML Metadata endpoint response
// https://1.1.1.1/cloud/org/my-org/saml/metadata/alias/vcd
type VcdSamlMetadata struct {
	XMLName xml.Name `xml:"EntityDescriptor"`
	Xmlns   string   `xml:"xmlns,attr,omitempty"`
	Text    string   `xml:",chardata"`
	ID      string   `xml:"ID,attr"`
	Md      string   `xml:"xmlns:md,attr,omitempty"`

	// EntityID is the configured vCD Entity ID which is used in ADFS authentication request
	// Note: once this field is set, it is not possible to change it back to empty,
	// but only to replace it with a different value
	EntityID string `xml:"entityID,attr"`
	// SPSSODescriptor is the main body of the SAML metadata file, which defines what the SAML identity provider can do
	SPSSODescriptor SPSSODescriptor `xml:"SPSSODescriptor,omitempty"`
}

// SPSSODescriptor is the main body of the SAML metadata file, which defines what the SAML identity provider can do
type SPSSODescriptor struct {
	Ds                         string `xml:"xmlns:ds,attr,omitempty"`
	AuthnRequestsSigned        bool   `xml:"AuthnRequestsSigned,attr"`
	ProtocolSupportEnumeration string `xml:"protocolSupportEnumeration,attr"`
	WantAssertionsSigned       bool   `xml:"WantAssertionsSigned,attr"`
	KeyDescriptor              []struct {
		Use     string `xml:"use,attr"`
		KeyInfo struct {
			//Ds       string `xml:"xmlns:ds,attr"`
			X509Data struct {
				X509Certificate string `xml:"X509Certificate"`
			} `xml:"X509Data"`
		} `xml:"KeyInfo"`
	} `xml:"KeyDescriptor"`

	SingleLogoutService []struct {
		Binding  string `xml:"Binding,attr"`
		Location string `xml:"Location,attr"`
	} `xml:"SingleLogoutService"`
	NameIDFormat             []string `xml:"NameIDFormat"`
	AssertionConsumerService []struct {
		Binding         string `xml:"Binding,attr"`
		Hoksso          string `xml:"xmlns:hoksso,attr"`
		Index           int    `xml:"index,attr"`
		IsDefault       bool   `xml:"isDefault,attr,omitempty"`
		Location        string `xml:"Location,attr"`
		ProtocolBinding string `xml:"ProtocolBinding,attr"`
	} `xml:"AssertionConsumerService"`
}

// AdfsAuthErrorEnvelope helps to parse ADFS authentication error with help of Error() method
//
// Note. This structure is not complete and has many more fields.
type AdfsAuthErrorEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Text  string `xml:",chardata"`
		Fault struct {
			Text string `xml:",chardata"`
			Code struct {
				Text    string `xml:",chardata"`
				Value   string `xml:"Value"`
				Subcode struct {
					Text  string `xml:",chardata"`
					Value struct {
						Text string `xml:",chardata"`
						A    string `xml:"a,attr"`
					} `xml:"Value"`
				} `xml:"Subcode"`
			} `xml:"Code"`
			Reason struct {
				Chardata string `xml:",chardata"`
				Text     struct {
					Text string `xml:",chardata"`
					Lang string `xml:"lang,attr"`
				} `xml:"Text"`
			} `xml:"Reason"`
		} `xml:"Fault"`
	} `xml:"Body"`
}

// Error satisfies Go's default `error` interface for AdfsAuthErrorEnvelope and formats
// error for human readable output
func (samlErr AdfsAuthErrorEnvelope) Error() string {
	return fmt.Sprintf("SAML request got error: %s", samlErr.Body.Fault.Reason.Text)
}

// AdfsAuthResponseEnvelope helps to marshal ADFS response to authentication request.
//
// Note. This structure is not complete and has many more fields.
type AdfsAuthResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		RequestSecurityTokenResponseCollection struct {
			RequestSecurityTokenResponse struct {
				// RequestedSecurityTokenTxt returns data which is accepted by vCD as a SIGN token
				RequestedSecurityTokenTxt InnerXML `xml:"RequestedSecurityToken"`
			} `xml:"RequestSecurityTokenResponse"`
		} `xml:"RequestSecurityTokenResponseCollection"`
	} `xml:"Body"`
}

// OrgFederationSettings is the structure used to set SAML identity service for an organization
type OrgFederationSettings struct {
	Href                            string   `xml:"href,attr,omitempty" json:"href,omitempty"`
	Type                            string   `xml:"type,attr,omitempty" json:"type,omitempty"`
	Link                            LinkList `xml:"Link,omitempty" json:"link,omitempty"`
	SAMLMetadata                    string   `xml:"SAMLMetadata" json:"samlMetadata"`
	Enabled                         bool     `xml:"Enabled" json:"enabled"`
	CertificateExpiration           string   `xml:"CertificateExpiration" json:"certificateExpiration"`
	SigningCertificateExpiration    string   `xml:"SigningCertificateExpiration" json:"signingCertificateExpiration"`
	EncryptionCertificateExpiration string   `xml:"EncryptionCertificateExpiration" json:"encryptionCertificateExpiration"`
	SamlSPEntityID                  string   `xml:"SamlSPEntityId" json:"samlSPEntityId"`
	SamlAttributeMapping            struct { // The names of SAML attributes used to populate user profiles.
		Href                   string   `xml:"href,attr,omitempty" json:"href,omitempty"`
		Type                   string   `xml:"type,attr,omitempty" json:"type,omitempty"`
		Link                   LinkList `xml:"Link,omitempty" json:"link,omitempty"`
		EmailAttributeName     string   `xml:"EmailAttributeName,omitempty" json:"emailAttributeName,omitempty"`
		UserNameAttributeName  string   `xml:"UserNameAttributeName,omitempty" json:"userNameAttributeName,omitempty"`
		FirstNameAttributeName string   `xml:"FirstNameAttributeName,omitempty" json:"firstNameAttributeName,omitempty"`
		SurnameAttributeName   string   `xml:"SurnameAttributeName,omitempty" json:"surnameAttributeName,omitempty"`
		FullNameAttributeName  string   `xml:"FullNameAttributeName,omitempty" json:"fullNameAttributeName,omitempty"`
		GroupAttributeName     string   `xml:"GroupAttributeName,omitempty" json:"groupAttributeName,omitempty"`
		RoleAttributeName      string   `xml:"RoleAttributeName,omitempty" json:"roleAttributeName,omitempty"`
	} `xml:"SamlAttributeMapping,omitempty" json:"samlAttributeMapping,omitempty"`
	SigningCertLibraryItemID    string `xml:"SigningCertLibraryItemId" json:"signingCertLibraryItemId"`
	EncryptionCertLibraryItemID string `xml:"EncryptionCertLibraryItemId" json:"encryptionCertLibraryItemID"`
}
