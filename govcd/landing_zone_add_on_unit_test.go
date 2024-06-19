//go:build unit || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestSolutionAddOn_ValidateInputs(t *testing.T) {
	emptyInputs := make(map[string]interface{})

	requiredInputsCreate := make(map[string]interface{})
	requiredInputsCreate["delete-previous-uiplugin-versions"] = true

	requiredInputsDelete := make(map[string]interface{})
	requiredInputsDelete["force-delete"] = true

	type args struct {
		userInputs           map[string]interface{}
		validateOnlyRequired bool
		isDeleteOperation    bool
	}

	tests := []struct {
		name     string
		manifest []byte
		args     args
		wantErr  bool
	}{
		{
			name:     "MissingRequiredCreate",
			manifest: sampleSolutionAddonManifest1,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: false, isDeleteOperation: false},
			wantErr:  true,
		},
		{
			name:     "SpecifiedRequiredCreate",
			manifest: sampleSolutionAddonManifest1,
			args:     args{userInputs: requiredInputsCreate, validateOnlyRequired: true, isDeleteOperation: false},
			wantErr:  false,
		},
		{
			name:     "SpecifiedRequiredDelete",
			manifest: sampleSolutionAddonManifest1,
			args:     args{userInputs: requiredInputsDelete, validateOnlyRequired: true, isDeleteOperation: true},
			wantErr:  false,
		},
		{
			name:     "MissingRequiredCreate2",
			manifest: sampleSolutionAddonManifest2,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: false, isDeleteOperation: false},
			wantErr:  true,
		},
		{
			name:     "MissingRequiredDelete",
			manifest: sampleSolutionAddonManifest1,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: false, isDeleteOperation: true},
			wantErr:  true,
		},
		{
			name:     "MissingRequiredDelete2",
			manifest: sampleSolutionAddonManifest2,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: true, isDeleteOperation: true},
			wantErr:  false,
		},
		{
			name:     "NoRequiredFieldsEmptyInputsCreate",
			manifest: sampleSolutionAddonManifestNoRequired,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: true, isDeleteOperation: false},
			wantErr:  false,
		},
		{
			name:     "NoRequiredFieldsEmptyInputsDelete",
			manifest: sampleSolutionAddonManifestNoRequired,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: true, isDeleteOperation: true},
			wantErr:  false,
		},
		{
			name:     "EmptyAddonInputsRequiredOnlyCreate",
			manifest: sampleSolutionAddonManifestEmptyInputs,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: true, isDeleteOperation: false},
			wantErr:  false,
		},
		{
			name:     "EmptyAddonInputsRequiredOnlyDelete",
			manifest: sampleSolutionAddonManifestEmptyInputs,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: true, isDeleteOperation: true},
			wantErr:  false,
		},
		{
			name:     "EmptyAddonInputsAllFieldsCreate",
			manifest: sampleSolutionAddonManifestEmptyInputs,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: false, isDeleteOperation: false},
			wantErr:  false,
		},
		{
			name:     "EmptyAddonInputsAllFieldsDelete",
			manifest: sampleSolutionAddonManifestEmptyInputs,
			args:     args{userInputs: emptyInputs, validateOnlyRequired: false, isDeleteOperation: true},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			addOnManifest := make(map[string]any)
			err := json.Unmarshal(tt.manifest, &addOnManifest)
			if err != nil {
				t.Fatalf("error unmarshalling sample Solution Add-On manifest: %s", err)
			}

			addon := SolutionAddOn{
				DefinedEntity: &DefinedEntity{DefinedEntity: &types.DefinedEntity{Name: "vmware.ds-1.4.0-23376809"}},
				SolutionAddOnEntity: &types.SolutionAddOn{
					Manifest: addOnManifest,
				},
			}

			if err := addon.ValidateInputs(tt.args.userInputs, tt.args.validateOnlyRequired, tt.args.isDeleteOperation); (err != nil) != tt.wantErr {
				t.Errorf("SolutionAddOn.ValidateInputs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSolutionAddOn_ConvertInputTypes(t *testing.T) {
	// emptyInputs := make(map[string]interface{})

	requiredInputsCreateBool := make(map[string]interface{})
	requiredInputsCreateBool["name"] = "name"
	requiredInputsCreateBool["delete-previous-uiplugin-versions"] = true

	requiredInputsCreateString := make(map[string]interface{})
	requiredInputsCreateString["name"] = "name"
	requiredInputsCreateString["delete-previous-uiplugin-versions"] = "true"

	requiredInputsCreateBoolWithInput := make(map[string]interface{})
	requiredInputsCreateBoolWithInput["name"] = "name"
	requiredInputsCreateBoolWithInput["input-delete-previous-uiplugin-versions"] = true

	requiredInputsCreateStringWithInput := make(map[string]interface{})
	requiredInputsCreateStringWithInput["name"] = "name"
	requiredInputsCreateStringWithInput["input-delete-previous-uiplugin-versions"] = "true"

	requiredInputsDelete := make(map[string]interface{})
	requiredInputsDelete["force-delete"] = true

	type args struct {
		userInputs map[string]interface{}
		// validateOnlyRequired bool
		// isDeleteOperation    bool
	}

	tests := []struct {
		name           string
		manifest       []byte
		args           args
		expectedOutput map[string]interface{}
		wantErr        bool
	}{
		{
			name:           "StringToBool",
			manifest:       sampleSolutionAddonManifest1,
			args:           args{userInputs: requiredInputsCreateString},
			expectedOutput: requiredInputsCreateBool,
			wantErr:        false,
		},
		{
			name:           "StringToString",
			manifest:       sampleSolutionAddonManifest1,
			args:           args{userInputs: requiredInputsCreateString},
			expectedOutput: requiredInputsCreateString,
			wantErr:        true,
		},
		{
			name:           "StringToBoolWithInput",
			manifest:       sampleSolutionAddonManifest1,
			args:           args{userInputs: requiredInputsCreateStringWithInput},
			expectedOutput: requiredInputsCreateBoolWithInput,
			wantErr:        false,
		},
		{
			name:           "StringToStringWithInput",
			manifest:       sampleSolutionAddonManifest1,
			args:           args{userInputs: requiredInputsCreateStringWithInput},
			expectedOutput: requiredInputsCreateStringWithInput,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			addOnManifest := make(map[string]any)
			err := json.Unmarshal(tt.manifest, &addOnManifest)
			if err != nil {
				t.Fatalf("error unmarshalling sample Solution Add-On manifest: %s", err)
			}

			addon := SolutionAddOn{
				DefinedEntity: &DefinedEntity{DefinedEntity: &types.DefinedEntity{Name: "vmware.ds-1.4.0-23376809"}},
				SolutionAddOnEntity: &types.SolutionAddOn{
					Manifest: addOnManifest,
				},
			}

			convertedInputs, err := addon.ConvertInputTypes(tt.args.userInputs)
			if (err != nil) && !tt.wantErr {
				t.Errorf("SolutionAddOn.ConvertInputTypes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if reflect.DeepEqual(convertedInputs, tt.expectedOutput) == tt.wantErr {
				diff := pretty.Diff(tt.expectedOutput, convertedInputs)
				t.Errorf("SolutionAddOn.ConvertInputTypes() values are not identical\n\n%s", diff)
			}
		})
	}
}

var sampleSolutionAddonManifest1 = []byte(`
{
	"name": "ds",
	"inputs": [
		{
			"name": "delete-previous-uiplugin-versions",
			"type": "Boolean",
			"title": "Delete Previous UI Plugin Versions",
			"default": false,
			"required": true,
			"description": "If setting true, the installation will delete all previous versions of this ui plugin. If setting false, the installation will just disable previous versions"
		},
		{
			"name": "force-delete",
			"type": "Boolean",
			"title": "Force Delete",
			"delete": true,
			"required": true,
			"default": false,
			"description": "If setting true, the uninstallation will remove all Data Solution records from Cloud Director but actual Data Solution instances will stay in Kubernetes clusters. If setting false, the uninstallation proceeds only when Data Solution records are not found in Cloud Director."
		}
	]
}
`)

var sampleSolutionAddonManifestNoRequired = []byte(`
{
	"name": "ds",
	"inputs": [
		{
			"name": "delete-previous-uiplugin-versions",
			"type": "Boolean",
			"title": "Delete Previous UI Plugin Versions",
			"default": false,
			"required": false,
			"description": "If setting true, the installation will delete all previous versions of this ui plugin. If setting false, the installation will just disable previous versions"
		},
		{
			"name": "force-delete",
			"type": "Boolean",
			"title": "Force Delete",
			"delete": true,
			"default": false,
			"description": "If setting true, the uninstallation will remove all Data Solution records from Cloud Director but actual Data Solution instances will stay in Kubernetes clusters. If setting false, the uninstallation proceeds only when Data Solution records are not found in Cloud Director."
		}
	]
}
`)

var sampleSolutionAddonManifestEmptyInputs = []byte(`
{
	"name": "ds",
	"inputs": [],
	"vendor": "vmware",
	"runtime": {
		"sdkVersion": "1.1.1.8577774"
	}
}
`)

var sampleSolutionAddonManifest2 = []byte(`
{
	"name": "ose",
	"inputs": [
		{
			"name": "kube-cluster-location",
			"type": "String",
			"title": "Kubernetes Cluster Location",
			"values": {
				"SLZ": "SLZ",
				"EXTERNAL": "EXTERNAL"
			},
			"default": "SLZ",
			"required": true,
			"description": "The Kubernetes cluster location is used to specify where Object Storage Extension will be installed. By SLZ, you specify an existing TKG cluster in the Solutions organization. Object Storage Extension Kubernetes Operator will be automatically installed by SLZ. By EXTERNAL, you manually install Object Storage Extension Kubernetes Operator onto a CNCF compliant Kubernetes cluster afterwards. Defaults to SLZ."
		},
		{
			"name": "vcd-api-token",
			"type": "String",
			"title": "VCD API Token",
			"secure": true,
			"required": true,
			"description": "This Cloud Director API token of provider administrator user is used for Object Storage Extension Operator installation."
		},
		{
			"name": "kube-cluster-name",
			"type": "String",
			"title": "Kubernetes Cluster Name",
			"description": "This is the name of an existing TKG cluster in the Solutions organization where you install Object Storage Extension. This parameter is required by SLZ."
		},
		{
			"name": "registry-url",
			"type": "String",
			"title": "Container Registry URL",
			"default": "projects.registry.vmware.com/vcd-ose",
			"validation": "^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])(/[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])*$",
			"description": "The container registry URL is used to pull Object Storage Extension container packages during installation, i.e., registry.mydomain.com/myproject. If you host Object Storage Extension container packages in a private registry, you must specify it here."
		},
		{
			"name": "registry-username",
			"type": "String",
			"title": "Container Registry User Name",
			"description": "The username of the container registry for Basic authentication."
		},
		{
			"name": "registry-password",
			"type": "String",
			"title": "Container Registry Password",
			"secure": true,
			"description": "The password of the container registry for Basic authentication."
		},
		{
			"name": "registry-ca-bundle",
			"type": "String",
			"view": "multiline",
			"title": "Container Registry CA Bundle",
			"description": "This is CA bundle in PEM format of the container registry's TLS certificate."
		},
		{
			"name": "force-delete",
			"type": "Boolean",
			"title": "Force Delete",
			"delete": true,
			"default": true,
			"description": "The force-delete is used to control the error handling in case the deletion of Object Storage Extension in the Kubernetes cluster is not completed. When it is true, the add-on will continue to delete itself if the deletion of Object Storage Extension Kubernetes Operator and server runs into an error; otherwise, the add-on stops at the stage where the error occurs. Defaults to true."
		},
		{
			"name": "deploy-timeout",
			"type": "Integer",
			"title": "Deploy Timeout",
			"default": 3600,
			"description": "The deploy timeout is used to set the timeout in seconds for Object Storage Extension Operator installation. Defaults to 3600."
		}
	]
}
`)
