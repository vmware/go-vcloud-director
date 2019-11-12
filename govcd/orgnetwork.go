/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type OrgNetworkRecord struct {
	OrgNetwork *types.QueryResultOrgNetworkRecordType
	client     *Client
}

func NewOrgNetworkRecord(cli *Client) *OrgNetworkRecord {
	return &OrgNetworkRecord{
		OrgNetwork: new(types.QueryResultOrgNetworkRecordType),
		client:     cli,
	}
}
