//go:build functional || openapi || certificate || alb || nsxt || network || tm || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	_ "embed"
)

var (
	//go:embed test-resources/cert.pem
	certificate string

	//go:embed test-resources/key.pem
	privateKey string

	//go:embed test-resources/rootCA.pem
	rootCaCertificate string
)
