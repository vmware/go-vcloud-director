/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var addOnInstancePublishBehaviorId = "urn:vcloud:behavior-interface:invoke:vmware:solutions_add_on_instance:1.0.0"

// Publishing manages publish and Unpublish operations, which are managed in the same API call
// To unpublish, the `scopeAll` has to be `false and `scope` must be empty
func (addonInstance *SolutionAddOnInstance) Publishing(scope []string, scopeAll bool) (string, error) {
	arguments := make(map[string]interface{})
	arguments["operation"] = "publish instance"
	arguments["name"] = addonInstance.SolutionAddOnInstance.Name
	if scope != nil {
		arguments["scope"] = strings.Join(scope, ",")
	} else {
		arguments["scope"] = ""
	}
	arguments["scope-all"] = scopeAll

	behaviorInvocation := types.BehaviorInvocation{
		Arguments: arguments,
	}

	parentRde := addonInstance.DefinedEntity
	result, err := parentRde.InvokeBehavior(addOnInstancePublishBehaviorId, behaviorInvocation)
	if err != nil {
		return "", fmt.Errorf("error invoking publish behavior of Solution Add-On instance '%s': %s", addonInstance.SolutionAddOnInstance.Name, err)
	}

	return result, nil
}
