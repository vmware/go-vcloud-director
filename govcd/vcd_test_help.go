package govcd

import "strings"

// Gets the two or three components of a "parent" string, as passed to AddToCleanupList
// If the number of split strings is not 2 or 3 it return 3 empty strings
// Example input parent: my-org|my-vdc|my-edge-gw, separator: |
// Output : first: my-org, second: my-vdc, third: my-edge-gw
func splitParent(parent string, separator string) (first, second, third string) {
	strList := strings.Split(parent, separator)
	if len(strList) < 2 || len(strList) > 3 {
		return "", "", ""
	}
	first = strList[0]
	second = strList[1]

	if len(strList) == 3 {
		third = strList[2]
	}

	return
}
