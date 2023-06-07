//go:build unit || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed test-resources/saml-test-sp.xml
var md1 string

//go:embed test-resources/saml-test-sp-invalid.xml
var md2 string

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
			result, err := normalizeServiceProviderSamlMetadata(sample.data)
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

			errors := ValidateSamlServiceProviderMetadata(result)

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
