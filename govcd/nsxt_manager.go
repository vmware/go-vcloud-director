/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type NsxtManager struct {
	NsxtManager *types.NsxtManager
	VCDClient   *VCDClient
	// Urn holds a URN value for NSX-T manager. None of the API endpoints return it, but filtering other entities requires that
	// Sample format: urn:vcloud:nsxtmanager:UUID
	//
	// Note:  this is being computed when retrieving the structure and will not be populated if this structure is initialized manually
	Urn string
}
