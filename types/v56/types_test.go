/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

import (
	"encoding/xml"
	"testing"
)

func TestProductSectionList_SortByPropertyKeyName(t *testing.T) {
	type fields struct {
		XMLName        xml.Name
		Ovf            string
		Xmlns          string
		ProductSection *ProductSection
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProductSectionList{
				XMLName:        tt.fields.XMLName,
				Ovf:            tt.fields.Ovf,
				Xmlns:          tt.fields.Xmlns,
				ProductSection: tt.fields.ProductSection,
			}
			p.SortByPropertyKeyName()
		})
	}
}
