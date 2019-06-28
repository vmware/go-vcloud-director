/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"path"
	"strings"
)

func extractNSXObjectIDfromLocation(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("unable to get ID from empty header")
	}

	cleanHeader := path.Clean(header) // Removes trailing slash if there is one
	splitLocation := strings.Split(cleanHeader, "/")

	if len(splitLocation) < 2 {
		return "", fmt.Errorf("header does not contain url path: %s", header)
	}

	objectID := splitLocation[len(splitLocation)-1]

	return objectID, nil
}
