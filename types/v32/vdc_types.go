/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package v32

import "github.com/vmware/go-vcloud-director/v2/types/v56"

// VdcConfiguration models the payload for creating a VDC.
// Type: CreateVdcParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Parameters for creating an organization VDC
// https://code.vmware.com/apis/553/vcloud-director/doc/doc/types/CreateVdcParamsType.html
type VdcCreateConfiguration struct {
	types.VdcConfiguration
	IsElastic             *bool `xml:"IsElastic,omitempty"`
	IncludeMemoryOverhead *bool `xml:"IncludeMemoryOverhead,omitempty"`
}
