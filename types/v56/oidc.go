/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

import (
	"encoding/xml"
)

type VcdOidcSettings struct {
	XMLName xml.Name `xml:"EntityDescriptor"`
	Xmlns   string   `xml:"xmlns,attr,omitempty"`
}
