//go:build functional || openapi || certificate || alb || nsxt || network || tm || ALL

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
