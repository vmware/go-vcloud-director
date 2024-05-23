//go:build unit || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestSolutionAddOn_Inputs(t *testing.T) {

	// res := types.SolutionAddOn{}
	// err := json.Unmarshal(sampleSolutionAddonManifest, &res)
	// if err != nil {
	// 	t.Fatalf("error unmarshalling sample solution addon manifest: %s", err)
	// }

	// inputs := res.Manifest["inputs"]

	res2 := make(map[string]any)
	// err := json.Unmarshal(sampleSolutionAddonManifest2, &res2)
	err := json.Unmarshal(sampleSolutionAddonManifest2, &res2)
	if err != nil {
		t.Fatalf("error unmarshalling sample Solution Add-On manifest: %s", err)
	}

	newSolutionAddon := SolutionAddOn{
		DefinedEntity: &DefinedEntity{DefinedEntity: &types.DefinedEntity{Name: "vmware.ds-1.4.0-23376809"}},
		SolutionEntity: &types.SolutionAddOn{
			Manifest: res2,
		},
	}

	inputsCopy := make(map[string]interface{})
	inputsCopy["operation"] = "create instance"

	err = newSolutionAddon.validate(inputsCopy, false)
	if err != nil {
		t.Fatalf("%s", err)
	}

	/*
		inputs := res2["inputs"].([]any)

		for _, inputField := range inputs {

			txt, err := json.Marshal(inputField)
			if err != nil {
				t.Fatalf("failed marshalling: %s", err)
			}

			inpField := types.SolutionAddOnInputField{}
			err = json.Unmarshal(txt, &inpField)
			if err != nil {
				t.Fatalf("failed unmarshalling: %s", err)
			}

			// fmt.Println(inpField.Name + " " + inpField.Type)

			spew.Dump(inpField)

			// iiii := inputField.(map[string]any)
			// spew.Dump(inputField)

		}

		// type fields struct {
		// 	SolutionEntity *types.SolutionAddOn
		// 	DefinedEntity  *DefinedEntity
		// 	vcdClient      *VCDClient
		// }
		// tests := []struct {
		// 	name    string
		// 	fields  fields
		// 	want    *SolutionAddOn
		// 	wantErr bool
		// }{
		// 	// TODO: Add test cases.
		// }
		// for _, tt := range tests {
		// 	t.Run(tt.name, func(t *testing.T) {
		// 		s := &SolutionAddOn{
		// 			SolutionEntity: tt.fields.SolutionEntity,
		// 			DefinedEntity:  tt.fields.DefinedEntity,
		// 			vcdClient:      tt.fields.vcdClient,
		// 		}
		// 		got, err := s.Inputs()
		// 		if (err != nil) != tt.wantErr {
		// 			t.Errorf("SolutionAddOn.Inputs() error = %v, wantErr %v", err, tt.wantErr)
		// 			return
		// 		}
		// 		if !reflect.DeepEqual(got, tt.want) {
		// 			t.Errorf("SolutionAddOn.Inputs() = %v, want %v", got, tt.want)
		// 		}
		// 	})
		// }
	*/
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
			"default": false,
			"description": "If setting true, the uninstallation will remove all Data Solution records from Cloud Director but actual Data Solution instances will stay in Kubernetes clusters. If setting false, the uninstallation proceeds only when Data Solution records are not found in Cloud Director."
		}
	],
	"vendor": "vmware",
	"runtime": {
		"sdkVersion": "1.1.1.8577774"
	},
	"version": "1.4.0-23376809",
	"elements": [
		{
			"name": "ds-ui-plugin",
			"type": "ui-plugin"
		},
		{
			"name": "ds-rde",
			"type": "defined-entity"
		},
		{
			"name": "ds-rights-bundle",
			"spec": {
				"name": "vmware:dataSolutionsRightsBundle",
				"rights": [
					"vmware:dsCluster: Administrator Full access"
				],
				"description": "Data Solutions Rights Bundle"
			},
			"type": "rights-bundle"
		}
	],
	"policies": {},
	"triggers": [
		{
			"event": "PostCreate",
			"action": "post-install-hook"
		},
		{
			"event": "PreUpgrade",
			"action": "post-install-hook"
		},
		{
			"event": "PostUpgrade",
			"action": "post-install-hook"
		},
		{
			"event": "PreDelete",
			"action": "post-install-hook"
		}
	],
	"resources": [
		"licenses"
	],
	"vcdVersion": "10.4.3",
	"description": "Data solution enabling multi-tenancy customers to deliver a portfolio of on-demand caching, messaging and database software at a massive scale.",
	"friendlyName": "Data Solutions"
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
	],
	"vendor": "vmware",
	"runtime": {
		"sdkVersion": "1.1.1.8558408"
	},
	"version": "3.0.0-23443325",
	"elements": [
		{
			"name": "ose-ui-plugin",
			"type": "ui-plugin"
		},
		{
			"name": "ose-rde",
			"type": "defined-entity"
		}
	],
	"policies": {
		"tenantScoped": true,
		"upgradesFrom": "3.0.0-*",
		"supportsMultipleInstances": true
	},
	"triggers": [
		{
			"event": "PreCreate",
			"action": "ose-hook"
		},
		{
			"event": "PostCreate",
			"action": "ose-hook"
		},
		{
			"event": "PreDelete",
			"action": "ose-hook"
		},
		{
			"event": "PreUpgrade",
			"action": "ose-hook"
		},
		{
			"event": "PostUpgrade",
			"action": "ose-hook"
		}
	],
	"resources": [
		"licenses",
		"resources",
		"x:installer"
	],
	"operations": [
		{
			"name": "Update",
			"action": "ose-hook",
			"timeout": 240,
			"description": "Update the installation configuration of Object Storage Extension",
			"friendlyName": ""
		}
	],
	"vcdVersion": "10.4.3",
	"description": "VMware Cloud Director Object Storage Extension offers S3 compliant object storage as a service for Cloud Director users.",
	"friendlyName": "Object Storage"
}
`)
