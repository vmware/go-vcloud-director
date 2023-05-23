//go:build unit || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"testing"
)

const md1 = `
<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"
                     ID="https___example.com_cloud_org_orgname_saml_metadata_alias_vcd"
                     entityID="https://example.com/cloud/org/orgname/saml/metadata/alias/vcd">
  <md:SPSSODescriptor AuthnRequestsSigned="true" WantAssertionsSigned="true"
                      protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:KeyDescriptor use="signing">
      <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        <ds:X509Data>
          <ds:X509Certificate>MIIDJTCCAg2gAwIBAgIIUO3CqLllWDkwDQYJKoZIhvcNAQELBQAwOTE3MDUGA1UEAwwuVk13YXJl
            IENsb3VkIERpcmVjdG9yIG9yZ2FuaXphdGlvbiBDZXJ0aWZpY2F0ZTAeFw0yMzA0MTkxMDEzMTVa
            Fw0yNDA0MTgxMDEzMTVaMDkxNzA1BgNVBAMMLlZNd2FyZSBDbG91ZCBEaXJlY3RvciBvcmdhbml6
            YXRpb24gQ2VydGlmaWNhdGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCG/Hv+wSEp
            mGBXUYlhrcQCbjxGuoVXs8mjCqY50qf0OsTjO+I0Vv2pDI0HxtBGlsMEOUvJx5bunYZSASeYpGd2
            cSz4boAEfubPMSwKVSzqWIvy8UptvX0D8dcKHbVmNAbRtg9YDvogEDAcUNK31fHxmxHrmBx61b8F
            BPGAQhblrMQMi2rtWJUHNmX0WdKvbNivJosiZi/sjHr5QiOrAu5yInN0OaWnMv3KpLO5IqMvp+Pa
            25NAeIRrsjlh8TRzmSqyHrC/uqDyGnlLT/iq3VtRW9lzebMMKUcqZcEzBmdMtYQRM/vywZcrbVYg
            8h81WRTQdB5Umucrzy9xHPVYoaLbAgMBAAGjMTAvMB0GA1UdDgQWBBSDLVewLTLVkqhMrPHOgD5y
            RuDGiDAOBgNVHQ8BAf8EBAMCBPAwDQYJKoZIhvcNAQELBQADggEBAHMFoi+B3l4zl3yMdKeTwq8e
            Naz5xEYBLaKekETlQB9LNpO985I1iPHDbKVtpM1KLhNZ0rNYjhsd9tsmqoFXooEzjPjimRo0s088
            mRPOpiIl5+Y+EOgZrQhfv22eSfXkPYDWMzC1gCK81diGf/jueaVQ/5roaq3tms1kJyVhxiM93SHh
            LbWa6kWto83JUB9DisMg2pkfSgd6zkINUFDiwRsMOE4U8dmZyB/JpS/tf2mujFZetVCo6j5Z2cVZ
            7o61nYr3x3RQGEH5HWrNcLOPZ5rNtSjmYNFKOXAiW9s1XYmlJVZjgIkgk+185yLc867f6qcoNkJ0
            566hxp2ZH7Ftu3I=
          </ds:X509Certificate>
        </ds:X509Data>
      </ds:KeyInfo>
    </md:KeyDescriptor>
    <md:KeyDescriptor use="encryption">
      <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        <ds:X509Data>
          <ds:X509Certificate>MIIDJTCCAg2gAwIBAgIIAtYIthpLxT0wDQYJKoZIhvcNAQELBQAwOTE3MDUGA1UEAwwuVk13YXJl
            IENsb3VkIERpcmVjdG9yIG9yZ2FuaXphdGlvbiBDZXJ0aWZpY2F0ZTAeFw0yMzA0MTkxMDEzMTVa
            Fw0yNDA0MTgxMDEzMTVaMDkxNzA1BgNVBAMMLlZNd2FyZSBDbG91ZCBEaXJlY3RvciBvcmdhbml6
            YXRpb24gQ2VydGlmaWNhdGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCCecZ2fIQk
            +ebeg32kt2IUDf0LASJqqJH8G7oxAPzgRW8Wrav6WA4oCZniNIGa5Kwch1FpKFeN0lqEty8wWqEK
            c1xYeTBYciN9pAaWXyhD952ynePQ+uDsVwfG41QjyO6YM0h5UmVLQb+S2v8Bwt9wExYQWyb2fjfT
            zFIHxzl41KAfYNmxQc4pFh1xHMDGnYrlCEIcWrrbKSPkyCIR8euQeJGG2AcRIT6w1em1OQJ8fM0o
            yhwbnm/Q3jQ2au+vU07xU4kUCPFekAuW3CrURLZfdCdJUTDaK8Sy2z7/tTlobcuiT6X6zPKMHHIl
            yLRXNEvGvWQ9qFbu4TT7fs2mCUFDAgMBAAGjMTAvMB0GA1UdDgQWBBQvAZqrlnq0KI8MvAjfi1ua
            DHo6HDAOBgNVHQ8BAf8EBAMCBPAwDQYJKoZIhvcNAQELBQADggEBAAiCFEq0BqSS9dLvUhzlRtw/
            qARBfAIhuM5qPfLUV+/2cOgmOl7aFBoVVuZJ465a3VUBeg7hfR3ZCyKW2OgyVA9Tc1iIvNAigDpy
            2KtRYCxzemKGNpFBkqZ5ZJQLx9HJDG4GFFGXsDrYoo51IO1BiMvxE6ahHSg+e5dVYozmkPeInS47
            FZFAfMiOrIWxEqM/NWcezi97276Eg8BRUZVc7ZaP+AH6NyOPP+tsdqd3qasMRQOSyaPvJ1AatyjI
            Sq32G0CaUixngLlHcivz5vhKm6T7Qe1yuJS7AFQWtwjiuoDvGfyGcETQ0hebxBJC/zNJJQnxdz1A
            TIwOmuRj74rI1QQ=
          </ds:X509Certificate>
        </ds:X509Data>
      </ds:KeyInfo>
    </md:KeyDescriptor>
    <md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                            Location="https://example.com/login/org/orgname/saml/SingleLogout/alias/vcd"/>
    <md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
                            Location="https://example.com/login/org/orgname/saml/SingleLogout/alias/vcd"/>
    <md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>
    <md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</md:NameIDFormat>
    <md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:persistent</md:NameIDFormat>
    <md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>
    <md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:X509SubjectName</md:NameIDFormat>
    <md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                                 Location="https://example.com/login/org/orgname/saml/SSO/alias/vcd"
                                 index="0" isDefault="true"/>
    <md:AssertionConsumerService xmlns:hoksso="urn:oasis:names:tc:SAML:2.0:profiles:holder-of-key:SSO:browser"
                                 Binding="urn:oasis:names:tc:SAML:2.0:profiles:holder-of-key:SSO:browser"
                                 Location="https://example.com/login/org/orgname/saml/HoKSSO/alias/vcd" hoksso:ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                                 index="1"/>
  </md:SPSSODescriptor>
</md:EntityDescriptor>
`

const md2 = `
<EntityDescriptor ID="https___example.com_cloud_org_orgname_saml_metadata_alias_vcd"
                     entityID="https://example.com/cloud/org/orgname/saml/metadata/alias/vcd">
  <SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol" WantAssertionsSigned="true">
    <AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" index="0"
                                 isDefault="true"
                                 Location="https://example.com/login/org/orgname/saml/SSO/alias/vcd"
                                 ProtocolBinding="">
    </AssertionConsumerService>
    <AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:profiles:holder-of-key:SSO:browser" index="1"
                                 Location="https://example.com/login/org/orgname/saml/HoKSSO/alias/vcd"
                                 ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST">
    </AssertionConsumerService>
    <KeyDescriptor use="signing">
      <KeyInfo>
        <X509Data>
          <X509Certificate>MIIDJTCCAg2gAwIBAgIIUO3CqLllWDkwDQYJKoZIhvcNAQELBQAwOTE3MDUGA1UEAwwuVk13YXJl&#xA;
            IENsb3VkIERpcmVjdG9yIG9yZ2FuaXphdGlvbiBDZXJ0aWZpY2F0ZTAeFw0yMzA0MTkxMDEzMTVa&#xA;
            Fw0yNDA0MTgxMDEzMTVaMDkxNzA1BgNVBAMMLlZNd2FyZSBDbG91ZCBEaXJlY3RvciBvcmdhbml6&#xA;
            YXRpb24gQ2VydGlmaWNhdGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCG/Hv+wSEp&#xA;
            mGBXUYlhrcQCbjxGuoVXs8mjCqY50qf0OsTjO+I0Vv2pDI0HxtBGlsMEOUvJx5bunYZSASeYpGd2&#xA;
            cSz4boAEfubPMSwKVSzqWIvy8UptvX0D8dcKHbVmNAbRtg9YDvogEDAcUNK31fHxmxHrmBx61b8F&#xA;
            BPGAQhblrMQMi2rtWJUHNmX0WdKvbNivJosiZi/sjHr5QiOrAu5yInN0OaWnMv3KpLO5IqMvp+Pa&#xA;
            25NAeIRrsjlh8TRzmSqyHrC/uqDyGnlLT/iq3VtRW9lzebMMKUcqZcEzBmdMtYQRM/vywZcrbVYg&#xA;
            8h81WRTQdB5Umucrzy9xHPVYoaLbAgMBAAGjMTAvMB0GA1UdDgQWBBSDLVewLTLVkqhMrPHOgD5y&#xA;
            RuDGiDAOBgNVHQ8BAf8EBAMCBPAwDQYJKoZIhvcNAQELBQADggEBAHMFoi+B3l4zl3yMdKeTwq8e&#xA;
            Naz5xEYBLaKekETlQB9LNpO985I1iPHDbKVtpM1KLhNZ0rNYjhsd9tsmqoFXooEzjPjimRo0s088&#xA;
            mRPOpiIl5+Y+EOgZrQhfv22eSfXkPYDWMzC1gCK81diGf/jueaVQ/5roaq3tms1kJyVhxiM93SHh&#xA;
            LbWa6kWto83JUB9DisMg2pkfSgd6zkINUFDiwRsMOE4U8dmZyB/JpS/tf2mujFZetVCo6j5Z2cVZ&#xA;
            7o61nYr3x3RQGEH5HWrNcLOPZ5rNtSjmYNFKOXAiW9s1XYmlJVZjgIkgk+185yLc867f6qcoNkJ0&#xA; 566hxp2ZH7Ftu3I=&#xA;
          </X509Certificate>
        </X509Data>
      </KeyInfo>
    </KeyDescriptor>
    <KeyDescriptor use="encryption">
      <KeyInfo>
        <X509Data>
          <X509Certificate>MIIDJTCCAg2gAwIBAgIIAtYIthpLxT0wDQYJKoZIhvcNAQELBQAwOTE3MDUGA1UEAwwuVk13YXJl&#xA;
            IENsb3VkIERpcmVjdG9yIG9yZ2FuaXphdGlvbiBDZXJ0aWZpY2F0ZTAeFw0yMzA0MTkxMDEzMTVa&#xA;
            Fw0yNDA0MTgxMDEzMTVaMDkxNzA1BgNVBAMMLlZNd2FyZSBDbG91ZCBEaXJlY3RvciBvcmdhbml6&#xA;
            YXRpb24gQ2VydGlmaWNhdGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCCecZ2fIQk&#xA;
            +ebeg32kt2IUDf0LASJqqJH8G7oxAPzgRW8Wrav6WA4oCZniNIGa5Kwch1FpKFeN0lqEty8wWqEK&#xA;
            c1xYeTBYciN9pAaWXyhD952ynePQ+uDsVwfG41QjyO6YM0h5UmVLQb+S2v8Bwt9wExYQWyb2fjfT&#xA;
            zFIHxzl41KAfYNmxQc4pFh1xHMDGnYrlCEIcWrrbKSPkyCIR8euQeJGG2AcRIT6w1em1OQJ8fM0o&#xA;
            yhwbnm/Q3jQ2au+vU07xU4kUCPFekAuW3CrURLZfdCdJUTDaK8Sy2z7/tTlobcuiT6X6zPKMHHIl&#xA;
            yLRXNEvGvWQ9qFbu4TT7fs2mCUFDAgMBAAGjMTAvMB0GA1UdDgQWBBQvAZqrlnq0KI8MvAjfi1ua&#xA;
            DHo6HDAOBgNVHQ8BAf8EBAMCBPAwDQYJKoZIhvcNAQELBQADggEBAAiCFEq0BqSS9dLvUhzlRtw/&#xA;
            qARBfAIhuM5qPfLUV+/2cOgmOl7aFBoVVuZJ465a3VUBeg7hfR3ZCyKW2OgyVA9Tc1iIvNAigDpy&#xA;
            2KtRYCxzemKGNpFBkqZ5ZJQLx9HJDG4GFFGXsDrYoo51IO1BiMvxE6ahHSg+e5dVYozmkPeInS47&#xA;
            FZFAfMiOrIWxEqM/NWcezi97276Eg8BRUZVc7ZaP+AH6NyOPP+tsdqd3qasMRQOSyaPvJ1AatyjI&#xA;
            Sq32G0CaUixngLlHcivz5vhKm6T7Qe1yuJS7AFQWtwjiuoDvGfyGcETQ0hebxBJC/zNJJQnxdz1A&#xA; TIwOmuRj74rI1QQ=&#xA;
          </X509Certificate>
        </X509Data>
      </KeyInfo>
    </KeyDescriptor>
    <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</NameIDFormat>
    <NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</NameIDFormat>
    <NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:persistent</NameIDFormat>
    <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</NameIDFormat>
    <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:X509SubjectName</NameIDFormat>
    <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                            Location="https://example.com/login/org/orgname/saml/SingleLogout/alias/vcd"></SingleLogoutService>
    <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
                            Location="https://example.com/login/org/orgname/saml/SingleLogout/alias/vcd"></SingleLogoutService>
  </SPSSODescriptor>
</EntityDescriptor>
`
const md3 = `
<EntityDescriptor>
  <SPSSODescriptor>
  </SPSSODescriptor>
</EntityDescriptor>
`

func TestNormalizeSamlMetadata(t *testing.T) {

	type mdSample struct {
		name    string
		data    string
		wantErr bool
	}
	var samples = []mdSample{
		{"correct", md1, false},
		{"no-tags", md2, false},
		{"empty-SPSSODescriptor", md3, true},
	}

	for i, sample := range samples {
		t.Run(fmt.Sprintf("%02d-%s", i, sample.name), func(t *testing.T) {
			result, err := normalizeSamlMetadata(sample.data)
			if err != nil {
				if !sample.wantErr {
					t.Fatalf("unwanted error: %s ", err)
				}
				t.Logf("expected error found: %s\n", err)
			} else {
				if sample.wantErr {
					t.Logf("%s\n", result)
					t.Fatalf("expected an error but returned success")
				}
			}
			if len(result) == 0 {
				t.Fatalf("unexpected 0 length for result\n")
			}

			errors := ValidateSamlMetadata(result)

			if errors != nil {
				message := GetErrorMessageFromErrorSlice(errors)
				t.Logf("%s\n", message)
				if !sample.wantErr {
					t.Fatalf("validation errors found\n")
				}
			}
		})
	}
}
