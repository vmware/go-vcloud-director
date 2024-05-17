/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"maps"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// var slzAddOnRdeType = [3]string{"vmware", "solutions_add_on", "1.0.0"}
var slzAddOnInstanceRdeType = [3]string{"vmware", "solutions_add_on_instance", "1.0.0"}

var addOnCreateInstanceBehaviorId = "urn:vcloud:behavior-interface:createInstance:vmware:solutions_add_on:1.0.0"
var addOnInstanceRemovalBehaviorId = "urn:vcloud:behavior-interface:invoke:vmware:solutions_add_on_instance:1.0.0"
var addOnInstancePublishBehaviorId = "urn:vcloud:behavior-interface:invoke:vmware:solutions_add_on_instance:1.0.0"

// var addOnInstanceRemovalBehaviorId = "urn:vcloud:behavior-interface:invoke:vmware:solutions_add_on_instance:1.0.0"

type SolutionAddOnInstance struct {
	SolutionEntity *types.SolutionAddOnInstance
	DefinedEntity  *DefinedEntity
	vcdClient      *VCDClient
}

func (addon *SolutionAddOn) CreateSolutionAddOnInstance(inputs map[string]interface{}) (*SolutionAddOnInstance, string, error) {

	// copy inputs to prevent mutation of function argument
	inputsCopy := make(map[string]interface{})
	maps.Copy(inputsCopy, inputs)

	inputsCopy["operation"] = "create instance"

	// Name is always mandatory
	name := inputsCopy["name"].(string)
	if name == "" {
		return nil, "", fmt.Errorf("'name' field must be present in the inputs")
	}

	behaviorInvocation := types.BehaviorInvocation{
		Arguments: inputsCopy,
	}

	parentRde := addon.DefinedEntity
	result, err := parentRde.InvokeBehavior(addOnCreateInstanceBehaviorId, behaviorInvocation)
	if err != nil {
		return nil, "", fmt.Errorf("error invoking RDE behavior: %s", err)
	}

	// Once the task is done and no error are here, one must find that instance from scratch

	createdAddOnInstance, err := addon.GetInstanceByName(name)
	if err != nil {
		return nil, "", fmt.Errorf("error retrieving Solution Add-On instance '%s' after creation: %s", name, err)
	}

	return createdAddOnInstance, result, nil
}

func (addon *SolutionAddOn) GetAllInstances() ([]*SolutionAddOnInstance, error) {
	vcdClient := addon.vcdClient

	// This filter ensures that only Add-On instances, that are based of this particular Add-On are
	// returned
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.prototype==%s", addon.RdeId()), queryParams)

	return vcdClient.GetAllSolutionAddonInstances(queryParams)
}

func (addon *SolutionAddOn) GetInstanceByName(name string) (*SolutionAddOnInstance, error) {
	vcdClient := addon.vcdClient

	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.prototype==%s;entity.name==%s", addon.RdeId(), name), queryParams)

	addOnInstances, err := vcdClient.GetAllSolutionAddonInstances(queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Solution Add-On Instance with name '%s': %s", name, err)
	}

	return oneOrError("name", name, addOnInstances)
}

func (vcdClient *VCDClient) GetAllSolutionAddonInstanceByName(name string) ([]*SolutionAddOnInstance, error) {
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.name==%s", name), queryParams)

	return vcdClient.GetAllSolutionAddonInstances(queryParams)
}

func (vcdClient *VCDClient) GetAllSolutionAddonInstances(queryParameters url.Values) ([]*SolutionAddOnInstance, error) {
	allAddonInstances, err := vcdClient.GetAllRdes(slzAddOnInstanceRdeType[0], slzAddOnInstanceRdeType[1], slzAddOnInstanceRdeType[2], queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Solution Add-on Instances: %s", err)
	}

	results := make([]*SolutionAddOnInstance, len(allAddonInstances))
	for index, rde := range allAddonInstances {
		addon, err := convertRdeEntityToAny[types.SolutionAddOnInstance](rde.DefinedEntity.Entity)
		if err != nil {
			return nil, fmt.Errorf("error converting RDE to Solution Add-on Instance: %s", err)
		}

		results[index] = &SolutionAddOnInstance{
			vcdClient:      vcdClient,
			DefinedEntity:  rde,
			SolutionEntity: addon,
		}
	}

	return results, nil
}

func (vcdClient *VCDClient) GetSolutionAddOnInstanceById(id string) (*SolutionAddOnInstance, error) {
	addOnInstanceRde, err := getRdeById(&vcdClient.Client, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Solution Add-On Instance RDE: %s", err)
	}

	addOnInstanceEntity, err := convertRdeEntityToAny[types.SolutionAddOnInstance](addOnInstanceRde.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}
	result := &SolutionAddOnInstance{
		vcdClient:      vcdClient,
		DefinedEntity:  addOnInstanceRde,
		SolutionEntity: addOnInstanceEntity,
	}

	return result, nil
}

func (addonInstance *SolutionAddOnInstance) RemoveSolutionAddOnInstance(deleteInputs map[string]interface{}) (string, error) {
	// copy deleteInputs to prevent mutation of function argument
	deleteInputsCopy := make(map[string]interface{})
	maps.Copy(deleteInputsCopy, deleteInputs)

	deleteInputsCopy["operation"] = "delete instance"

	behaviorInvocation := types.BehaviorInvocation{
		Arguments: deleteInputsCopy,
	}

	parentRde := addonInstance.DefinedEntity
	result, err := parentRde.InvokeBehavior(addOnInstanceRemovalBehaviorId, behaviorInvocation)
	if err != nil {
		return "", fmt.Errorf("error invoking removal of Solution Add-On instance '%s': %s", addonInstance.SolutionEntity.Name, err)
	}

	return result, nil
}

// Publish and Unpublish operations are managed in the same API call
// For Unpublishing the `scopeAll` has to be `false and `scope`
func (addonInstance *SolutionAddOnInstance) Publishing(scope []string, scopeAll bool) (string, error) {
	arguments := make(map[string]interface{})
	arguments["operation"] = "publish instance"
	arguments["name"] = addonInstance.SolutionEntity.Name
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
		return "", fmt.Errorf("error invoking publish behavior of Solution Add-On instance '%s': %s", addonInstance.SolutionEntity.Name, err)
	}

	return result, nil
}

func (addOnInstance *SolutionAddOnInstance) RdeId() string {
	if addOnInstance == nil || addOnInstance.DefinedEntity == nil || addOnInstance.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return addOnInstance.DefinedEntity.DefinedEntity.ID
}
