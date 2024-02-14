//go:build unit || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	semver "github.com/hashicorp/go-version"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"reflect"
	"testing"
)

// Test_getTkgVersionBundleFromVAppTemplate tests getTkgVersionBundleFromVAppTemplate function
func Test_getTkgVersionBundleFromVAppTemplate(t *testing.T) {
	tests := []struct {
		name                  string
		kubernetesTemplateOva *types.VAppTemplate
		want                  tkgVersionBundle
		wantErr               string
	}{
		{
			name:                  "ova is nil",
			kubernetesTemplateOva: nil,
			wantErr:               "the Kubernetes Template OVA is nil",
		},
		{
			name: "ova without children",
			kubernetesTemplateOva: &types.VAppTemplate{
				Name:     "dummy",
				Children: nil,
			},
			wantErr: "the Kubernetes Template OVA 'dummy' doesn't have any child VM",
		},
		{
			name: "ova with nil children",
			kubernetesTemplateOva: &types.VAppTemplate{
				Name:     "dummy",
				Children: &types.VAppTemplateChildren{VM: nil},
			},
			wantErr: "the Kubernetes Template OVA 'dummy' doesn't have any child VM",
		},
		{
			name: "ova with nil product section",
			kubernetesTemplateOva: &types.VAppTemplate{
				Name: "dummy",
				Children: &types.VAppTemplateChildren{VM: []*types.VAppTemplate{
					{
						ProductSection: nil,
					},
				}},
			},
			wantErr: "the Product section of the Kubernetes Template OVA 'dummy' is empty, can't proceed",
		},
		{
			name: "ova with no version in the product section",
			kubernetesTemplateOva: &types.VAppTemplate{
				Name: "dummy",
				Children: &types.VAppTemplateChildren{VM: []*types.VAppTemplate{
					{
						ProductSection: &types.ProductSection{
							Property: []*types.Property{
								{
									Key:          "foo",
									DefaultValue: "bar",
								},
							},
						},
					},
				}},
			},
			wantErr: "could not find any VERSION property inside the Kubernetes Template OVA 'dummy' Product section",
		},
		{
			name: "correct ova",
			kubernetesTemplateOva: &types.VAppTemplate{
				Name: "dummy",
				Children: &types.VAppTemplateChildren{VM: []*types.VAppTemplate{
					{
						ProductSection: &types.ProductSection{
							Property: []*types.Property{
								{
									Key:          "VERSION",
									DefaultValue: "v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc",
								},
							},
						},
					},
				}},
			},
			want: tkgVersionBundle{
				EtcdVersion:       "v3.5.6_vmware.9",
				CoreDnsVersion:    "v1.9.3_vmware.8",
				TkgVersion:        "v2.2.0",
				TkrVersion:        "v1.25.7---vmware.2-tkg.1",
				KubernetesVersion: "v1.25.7+vmware.2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTkgVersionBundleFromVAppTemplate(tt.kubernetesTemplateOva)
			if err != nil {
				if tt.wantErr != err.Error() {
					t.Errorf("getTkgVersionBundleFromVAppTemplate() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTkgVersionBundleFromVAppTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCseTemplate(t *testing.T) {
	v40, err := semver.NewVersion("4.0")
	if err != nil {
		t.Fatalf("could not create 4.0 version object")
	}
	v41, err := semver.NewVersion("4.1")
	if err != nil {
		t.Fatalf("could not create 4.1 version object")
	}
	v410, err := semver.NewVersion("4.1.0")
	if err != nil {
		t.Fatalf("could not create 4.1.0 version object")
	}
	v411, err := semver.NewVersion("4.1.1")
	if err != nil {
		t.Fatalf("could not create 4.1.1 version object")
	}
	v420, err := semver.NewVersion("4.2.0")
	if err != nil {
		t.Fatalf("could not create 4.2.0 version object")
	}

	tmpl41, err := getCseTemplate(*v41, "rde")
	if err != nil {
		t.Fatalf("%s", err)
	}
	tmpl410, err := getCseTemplate(*v410, "rde")
	if err != nil {
		t.Fatalf("%s", err)
	}
	tmpl411, err := getCseTemplate(*v411, "rde")
	if err != nil {
		t.Fatalf("%s", err)
	}
	if tmpl41 == "" || tmpl41 != tmpl410 || tmpl41 != tmpl411 || tmpl410 != tmpl411 {
		t.Fatalf("templates should be the same:\n4.1: %s\n4.1.0: %s\n4.1.1: %s", tmpl41, tmpl410, tmpl411)
	}

	tmpl420, err := getCseTemplate(*v420, "rde")
	if err != nil {
		t.Fatalf("%s", err)
	}
	if tmpl420 == "" {
		t.Fatalf("the obtained template for %s is empty", v420.String())
	}

	_, err = getCseTemplate(*v40, "rde")
	if err == nil && err.Error() != "the Container Service minimum version is '4.1.0'" {
		t.Fatalf("expected an error but got %s", err)
	}
}
