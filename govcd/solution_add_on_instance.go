/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

var slzAddOnInstanceRdeType = [3]string{"vmware", "solutions_add_on_instance", "1.0.0"}

var addOnCreateInstanceBehaviorId = "urn:vcloud:behavior-interface:createInstance:vmware:solutions_add_on:1.0.0"
var addOnInstanceRemovalBehaviorId = "urn:vcloud:behavior-interface:invoke:vmware:solutions_add_on_instance:1.0.0"

type SolutionAddOnInstance struct {
	SolutionAddOnInstance *types.SolutionAddOnInstance
	DefinedEntity         *DefinedEntity
	vcdClient             *VCDClient
}

// CreateSolutionAddOnInstance instantiates a new Solution Add-On. Some inputs may be mandatory for
// creation depending on the Solution Add-On itself. Methods 'ValidateInputs' can help to
// dynamically validate inputs based on the requirements in Solution Add-On.
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

// GetAllInstances retrieves all Solution Add-On Instances
func (addon *SolutionAddOn) GetAllInstances() ([]*SolutionAddOnInstance, error) {
	vcdClient := addon.vcdClient

	// This filter ensures that only Add-On instances, that are based of this particular Add-On are
	// returned
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.prototype==%s", addon.RdeId()), queryParams)

	return vcdClient.GetAllSolutionAddonInstances(queryParams)
}

// GetInstanceByName retrieves Solution Add-On Instance by name for a particular Solution Add-On.
// It will return an error if there is more than one Solution Add-On Instance with such name.
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

// GetAllSolutionAddonInstancesByName will retrieve all Solution Add-On Instances available
func (vcdClient *VCDClient) GetAllSolutionAddonInstancesByName(name string) ([]*SolutionAddOnInstance, error) {
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.name==%s", name), queryParams)

	return vcdClient.GetAllSolutionAddonInstances(queryParams)
}

// GetSolutionAddonInstanceByName will retrieve a single Solution Add-On Instance by name or fail
func (vcdClient *VCDClient) GetSolutionAddonInstanceByName(name string) (*SolutionAddOnInstance, error) {
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("entity.name==%s", name), queryParams)

	addOnInstances, err := vcdClient.GetAllSolutionAddonInstances(queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Solution Add-On Instance with name '%s': %s", name, err)
	}

	return oneOrError("name", name, addOnInstances)
}

// GetAllSolutionAddonInstances will retrieve Solution Add-On Instances based on given query parameters
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
			vcdClient:             vcdClient,
			DefinedEntity:         rde,
			SolutionAddOnInstance: addon,
		}
	}

	return results, nil
}

// GetSolutionAddOnInstanceById retrieves a Solution Add-On Instance with a given ID
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
		vcdClient:             vcdClient,
		DefinedEntity:         addOnInstanceRde,
		SolutionAddOnInstance: addOnInstanceEntity,
	}

	return result, nil
}

// Delete will delete a Solution Add-On instance with given 'deleteInputs'. Some fields in
// 'deleteInputs' might be mandatory for deletion of an instance. One can use 'ValidateInputs'
// method to check what inputs are defined for a particular Solution Add-On
func (addonInstance *SolutionAddOnInstance) Delete(deleteInputs map[string]interface{}) (string, error) {
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
		return "", fmt.Errorf("error invoking removal of Solution Add-On instance '%s': %s", addonInstance.SolutionAddOnInstance.Name, err)
	}

	return result, nil
}

// GetParentSolutionAddOn retrieves parent Solution Add-On that is specified in the Prototype field
func (addOnInstance *SolutionAddOnInstance) GetParentSolutionAddOn() (*SolutionAddOn, error) {
	if addOnInstance == nil || addOnInstance.DefinedEntity == nil || addOnInstance.DefinedEntity.DefinedEntity == nil {
		return nil, fmt.Errorf("cannot retrieve parent Solution Add-On from empty instance")
	}

	return addOnInstance.vcdClient.GetSolutionAddonById(addOnInstance.SolutionAddOnInstance.Prototype)
}

// RdeId is a shortcut to retrieve parent RDE ID
func (addOnInstance *SolutionAddOnInstance) RdeId() string {
	if addOnInstance == nil || addOnInstance.DefinedEntity == nil || addOnInstance.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return addOnInstance.DefinedEntity.DefinedEntity.ID
}

// ReadCreationInputValues will read all input values that were specified upon instance creation and return them
// either in their natural types, or all values converted to strings
func (addOnInstance *SolutionAddOnInstance) ReadCreationInputValues(convertAllValuesToStrings bool) (map[string]interface{}, error) {
	if addOnInstance == nil || addOnInstance.SolutionAddOnInstance == nil || addOnInstance.SolutionAddOnInstance.Properties == nil {
		return nil, fmt.Errorf("cannot extract properties - they are nil")
	}

	parentAddOn, err := addOnInstance.GetParentSolutionAddOn()
	if err != nil {
		return nil, fmt.Errorf("error retrieving parent Solution Add-On: %s", err)
	}

	schemaInputFields, err := parentAddOn.extractInputs()
	if err != nil {
		return nil, fmt.Errorf("error extracting inputs from Solution Add-On manifests: %s", err)
	}

	// Fields are specified within addOnInstance.SolutionAddOnInstance.Properties but they contain more values
	// that just the inputs themselves. Searching for values of all defined inputs.
	resultMap := make(map[string]interface{})
	for _, schemaInputField := range schemaInputFields {
		// Deletion fields are not stored in schema because they are not supplied for creating the
		// instance
		if schemaInputField.Delete {
			continue
		}

		util.Logger.Printf("[TRACE] Solution Add-On Instance Input field - looking for field '%s'", schemaInputField.Name)
		if foundValue, ok := addOnInstance.SolutionAddOnInstance.Properties[schemaInputField.Name]; ok {
			util.Logger.Printf("[TRACE] Solution Add-On Instance Input field - found field '%s' of type %s",
				schemaInputField.Name, schemaInputField.Type)
			resultMap[schemaInputField.Name] = foundValue
		}
	}

	if convertAllValuesToStrings {
		convertedResultMap := make(map[string]interface{})
		util.Logger.Printf("[TRACE] Solution Add-On Instance Inputs - converting all values to strings")

		for fieldName, fieldValue := range resultMap {
			convertedResultMap[fieldName] = fmt.Sprintf("%v", fieldValue)
		}
		return convertedResultMap, nil
	}

	return resultMap, nil
}

// ValidateInputs will check if 'userInputs' match required fields as defined in the Solution Add-On
// itself. Error will contained detailed information about missing fields.
func (addon *SolutionAddOn) ValidateInputs(userInputs map[string]interface{}, validateOnlyRequired, isDeleteOperation bool) error {
	schemaInputs, err := addon.extractInputs()
	if err != nil {
		return err
	}

	requiredFields := make(map[string]bool)
	for _, si := range schemaInputs {
		// Skip field if the operation does not match
		// Required fields can be defined either for create or for update operations
		if si.Delete != isDeleteOperation {
			continue
		}

		// Validating only required fields is set
		// Skipping a non required field.
		if !si.Required && validateOnlyRequired {
			continue
		}

		// Setting the key, but not marking as found yet
		requiredFields[si.Name] = false
	}

	// Check if all required fields are set in inputs
	for requiredFieldKey := range requiredFields {
		for userInputKey := range userInputs {
			if requiredFieldKey == userInputKey || fmt.Sprintf("input-%s", requiredFieldKey) == userInputKey { // field found
				requiredFields[requiredFieldKey] = true
			}
		}
	}

	// Check if all field constraints are satisfied
	missingFields := make([]string, 0)
	msFields := make([]*types.SolutionAddOnInputField, 0)
	for k := range requiredFields {
		if !requiredFields[k] {
			missingFields = append(missingFields, k)
			field, err := localFilterOneOrError("Solution Add-On filter value", schemaInputs, "Name", k)
			if err != nil {
				return fmt.Errorf("error finding field with key '%s'", k)
			}
			msFields = append(msFields, field)
		}
	}

	if len(missingFields) > 0 {
		fieldInfo, err := printAddonFieldData(msFields)
		if err != nil {
			return fmt.Errorf("error processing missing fields '%s' for: %s", addon.DefinedEntity.DefinedEntity.Name, err)
		}

		return fmt.Errorf("%s\n\nERROR: Missing fields '%s' for Solution Add-On '%s'",
			fieldInfo, strings.Join(missingFields, ", "), addon.DefinedEntity.DefinedEntity.Name)
	}

	return nil
}

// ConvertInputTypes will make sure that values will match types as defined in Add-On schema
// The needs for this operation comes from the fact that at least some of the Solution Add-Ons will
// fail if a boolean "false" is sent as a string
func (addon *SolutionAddOn) ConvertInputTypes(userInputs map[string]interface{}) (map[string]interface{}, error) {
	schemaInputs, err := addon.extractInputs()
	if err != nil {
		return nil, err
	}

	userInputsCopy := make(map[string]interface{})
	maps.Copy(userInputsCopy, userInputs)

	for userInputKey, userInputValue := range userInputsCopy {
		// search for key in the schema inputs and find correct type
		util.Logger.Printf("[TRACE] Solution Add-On Schema conversion - user input key %s, value of type %T", userInputKey, userInputValue)

		typeOfUserInputValue := reflect.TypeOf(userInputValue)

		var foundField *types.SolutionAddOnInputField
		for _, field := range schemaInputs {
			util.Logger.Printf("[TRACE] Solution Add-On Schema conversion - checking field %#v against user specified field %s", field, userInputKey)
			if field.Name == userInputKey || fmt.Sprintf("input-%s", field.Name) == userInputKey {
				foundField = field
				util.Logger.Printf("[TRACE] Solution Add-On Schema conversion - found field in schema %#v", field)
				break
			}
		}

		if foundField == nil {
			util.Logger.Printf("[TRACE] Solution Add-On Schema conversion - field '%s' not found in schema", userInputKey)
		}

		// User supplied string value, but actual type is different
		// Only attempting to convert fields that are in the schema
		if foundField != nil && typeOfUserInputValue.String() == "string" && foundField.Type != "String" {
			userInputStringValue := userInputValue.(string)
			util.Logger.Printf("[TRACE] Solution Add-On Schema conversion - found field in schema %#v", foundField)
			switch foundField.Type {
			case "Boolean":
				util.Logger.Printf("[TRACE] Solution Add-On Schema conversion - converting field '%s' to bool", userInputKey)
				// override string key to match boolean type
				boolValue, err := strconv.ParseBool(userInputStringValue)
				if err != nil {
					return nil, fmt.Errorf("error converting field '%s' to boolean: %s", userInputKey, err)
				}
				userInputsCopy[userInputKey] = boolValue
			case "Integer":
				util.Logger.Printf("[TRACE] Solution Add-On Schema conversion - converting field '%s' to int", userInputKey)
				// override string key to match integer type
				intValue, err := strconv.Atoi(userInputStringValue)
				if err != nil {
					return nil, fmt.Errorf("error converting field '%s' to integer: %s", userInputKey, err)
				}
				userInputsCopy[userInputKey] = intValue
			default:
				return nil, fmt.Errorf("unknown field type '%s' for field '%s'", foundField.Type, userInputKey)
			}
		}
	}

	util.Logger.Printf("[TRACE] Solution Add-On Schema conversion - final result %#v", userInputsCopy)

	return userInputsCopy, nil
}

// extractInputs retrieves input field definitions for instantiating and removing Solution Add-On from manifest
func (addon *SolutionAddOn) extractInputs() ([]*types.SolutionAddOnInputField, error) {
	// Extract inputs definition / Manifest["inputs"]
	inputValidation := addon.SolutionAddOnEntity.Manifest["inputs"]
	inputValidationSlice, ok := inputValidation.([]any)
	if !ok {
		return nil, fmt.Errorf("error processing Solution Add-On input validation metadata")
	}

	inputFieldMetadata, err := convertAllAddonInputFields(inputValidationSlice)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("error converting Solution Add-On input validation metadata: %s", err)
		}
	}

	return inputFieldMetadata, nil
}

func printAddonFieldData(allFields []*types.SolutionAddOnInputField) (string, error) {
	buf := bytes.NewBufferString("\n")

	_, _ = fmt.Fprintf(buf, "-----------------\n")
	for _, f := range allFields {
		_, _ = fmt.Fprintf(buf, "Field: %s\n", f.Name)
		_, _ = fmt.Fprintf(buf, "Title: %s\n", f.Title)
		_, _ = fmt.Fprintf(buf, "Type: %s\n", f.Type)
		_, _ = fmt.Fprintf(buf, "Required: %t\n", f.Required)
		_, _ = fmt.Fprintf(buf, "IsDelete: %t\n", f.Delete)
		_, _ = fmt.Fprintf(buf, "Description: %s\n", f.Description)
		if f.Default != nil {
			_, _ = fmt.Fprintf(buf, "Default: %v\n", f.Default)
		}

		_, _ = fmt.Fprintf(buf, "-----------------\n")
	}

	return buf.String(), nil
}

func convertAllAddonInputFields(allInputFields []any) ([]*types.SolutionAddOnInputField, error) {
	allFields := make([]*types.SolutionAddOnInputField, len(allInputFields))

	for index, inputField := range allInputFields {
		inpField, err := convertAddonInputField(inputField)
		if err != nil {
			return nil, err
		}

		allFields[index] = inpField
	}

	return allFields, nil
}

func convertAddonInputField(field any) (*types.SolutionAddOnInputField, error) {
	txt, err := json.Marshal(field)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling Input Field: %s", err)
	}

	fieldType := types.SolutionAddOnInputField{}
	err = json.Unmarshal(txt, &fieldType)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling Input Field to exact type: %s", err)
	}

	return &fieldType, nil
}
