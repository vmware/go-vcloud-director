//go:build functional || openapi || certificate || alb || nsxt || network || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

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
