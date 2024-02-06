//go:build unit || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"reflect"
	"testing"
)

// Test_getTkgVersionBundleFromVAppTemplateName tests getTkgVersionBundleFromVAppTemplateName function
func Test_getTkgVersionBundleFromVAppTemplateName(t *testing.T) {
	tests := []struct {
		name                      string
		kubernetesTemplateOvaName string
		want                      tkgVersionBundle
		wantErr                   string
	}{
		{
			name:                      "input is empty",
			kubernetesTemplateOvaName: "",
			wantErr:                   "the Kubernetes Template OVA cannot be empty",
		},
		{
			name:                      "input is Photon OVA",
			kubernetesTemplateOvaName: "photon-2004-kube-v9.99.9+vmware.9-tkg.9-aaaaa.ova",
			wantErr:                   "the Kubernetes Template OVA 'photon-2004-kube-v9.99.9+vmware.9-tkg.9-aaaaa.ova' uses Photon, and it is not supported",
		},
		{
			name:                      "input is not a Kubernetes OVA",
			kubernetesTemplateOvaName: "random-ova.ova",
			wantErr:                   "the OVA 'random-ova.ova' is not a Kubernetes template OVA",
		},
		{
			name:                      "input is not supported",
			kubernetesTemplateOvaName: "ubuntu-2004-kube-v9.99.9+vmware.9-tkg.9-99999999999999999999999999999999.ova",
			wantErr:                   "the Kubernetes Template OVA 'ubuntu-2004-kube-v9.99.9+vmware.9-tkg.9-99999999999999999999999999999999.ova' is not supported",
		},
		{
			name:                      "correct OVA",
			kubernetesTemplateOvaName: "ubuntu-2004-kube-v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc.ova",
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
			got, err := getTkgVersionBundleFromVAppTemplateName(tt.kubernetesTemplateOvaName)
			if err != nil {
				if tt.wantErr != err.Error() {
					t.Errorf("getTkgVersionBundleFromVAppTemplateName() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTkgVersionBundleFromVAppTemplateName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
