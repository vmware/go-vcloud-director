// +build unit ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"regexp"
	"testing"
)

func TestGetPseudoUUID(t *testing.T) {

	var seen = make(map[string]int)

	reUuid := regexp.MustCompile(`^[A-F0-9]{8}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{12}$`)
	for N := 0; N < 1000; N++ {
		uuid, _ := getPseudoUuid()
		if !reUuid.MatchString(uuid) {
			t.Logf("string %s doesn't look like a UUID", uuid)
			t.Fail()
		}
		previous, found := seen[uuid]
		if found {
			t.Logf("uuid %s already in the generated list at position %d", uuid, previous)
			t.Fail()
		}
		seen[uuid] = N
	}
}

func Test_updateLoadBalancerRawXml(t *testing.T) {

	fakeBody := testFakeLoadBalancerXml("<enabled>false</enabled>",
		"<accelerationEnabled>true</accelerationEnabled>",
		"<logging><enable>false</enable><logLevel>info</logLevel></logging>")

	reversoLoggingBody := testFakeLoadBalancerXml("<enabled>false</enabled>",
		"<accelerationEnabled>true</accelerationEnabled>",
		"<logging><logLevel>info</logLevel><enable>false</enable></logging>")
	missingAllFieldsBody := testFakeLoadBalancerXml("", "", "")
	missingEnabledFieldBody := testFakeLoadBalancerXml("", "<accelerationEnabled>true</accelerationEnabled>",
		"<logging><logLevel>info</logLevel><enable>false</enable></logging>")
	missingAccelerationFieldBody := testFakeLoadBalancerXml("<enabled>false</enabled>", "",
		"<logging><logLevel>info</logLevel><enable>false</enable></logging>")
	missingLoggingFieldBody := testFakeLoadBalancerXml("<enabled>false</enabled>", "",
		"<logging><logLevel>info</logLevel><enable>false</enable></logging>")

	type args struct {
		body                string
		enabled             bool
		accelerationEnabled bool
		loggingEnabled      bool
		logLevel            string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "SetAllTrue",
			args:    args{body: fakeBody, enabled: true, accelerationEnabled: true, loggingEnabled: true, logLevel: "info"},
			want:    "<loadBalancer><version>6956</version><enabled>true</enabled><enableServiceInsertion>false</enableServiceInsertion><accelerationEnabled>true</accelerationEnabled><applicationRule><applicationRuleId>applicationRule-91</applicationRuleId><name>1</name><script>acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page</script></applicationRule><applicationRule><applicationRuleId>applicationRule-92</applicationRuleId><name>2</name><script>acl hello payload(0,6) -m bin 48656c6c6f0a</script></applicationRule><logging><enable>true</enable><logLevel>info</logLevel></logging></loadBalancer>",
			wantErr: false,
		},
		{
			name:    "SetAllFalse",
			args:    args{body: fakeBody, enabled: false, accelerationEnabled: false, loggingEnabled: false, logLevel: "emergency"},
			want:    "<loadBalancer><version>6956</version><enabled>false</enabled><enableServiceInsertion>false</enableServiceInsertion><accelerationEnabled>false</accelerationEnabled><applicationRule><applicationRuleId>applicationRule-91</applicationRuleId><name>1</name><script>acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page</script></applicationRule><applicationRule><applicationRuleId>applicationRule-92</applicationRuleId><name>2</name><script>acl hello payload(0,6) -m bin 48656c6c6f0a</script></applicationRule><logging><enable>false</enable><logLevel>emergency</logLevel></logging></loadBalancer>",
			wantErr: false,
		},
		{
			name:    "ReverseLoggingFields",
			args:    args{body: reversoLoggingBody, enabled: false, accelerationEnabled: false, loggingEnabled: false, logLevel: "emergency"},
			want:    "<loadBalancer><version>6956</version><enabled>false</enabled><enableServiceInsertion>false</enableServiceInsertion><accelerationEnabled>false</accelerationEnabled><applicationRule><applicationRuleId>applicationRule-91</applicationRuleId><name>1</name><script>acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page</script></applicationRule><applicationRule><applicationRuleId>applicationRule-92</applicationRuleId><name>2</name><script>acl hello payload(0,6) -m bin 48656c6c6f0a</script></applicationRule><logging><enable>false</enable><logLevel>emergency</logLevel></logging></loadBalancer>",
			wantErr: false,
		},
		{
			name:    "MissingFields",
			args:    args{body: missingAllFieldsBody, enabled: false, accelerationEnabled: false, loggingEnabled: false, logLevel: "emergency"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "MissingEnabledField",
			args:    args{body: missingEnabledFieldBody, enabled: false, accelerationEnabled: false, loggingEnabled: false, logLevel: "emergency"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "MissingAccelerationField",
			args:    args{body: missingAccelerationFieldBody, enabled: false, accelerationEnabled: false, loggingEnabled: false, logLevel: "emergency"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "MissingLoggingFields",
			args:    args{body: missingLoggingFieldBody, enabled: false, accelerationEnabled: false, loggingEnabled: false, logLevel: "emergency"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateLoadBalancerRawXml(tt.args.body, tt.args.enabled, tt.args.accelerationEnabled, tt.args.loggingEnabled, tt.args.logLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateLoadBalancerRawXml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("updateLoadBalancerRawXml() = %v, want %v", got, tt.want)
			}
		})
	}
}

// testFakeLoadBalancerXml is used to build sample XML bodies
func testFakeLoadBalancerXml(enabled, accelerationEnabled, logging string) string {
	return fmt.Sprintf("<loadBalancer><version>6956</version>" + enabled +
		"<enableServiceInsertion>false</enableServiceInsertion>" + accelerationEnabled +
		"<applicationRule><applicationRuleId>applicationRule-91</applicationRuleId><name>1</name>" +
		"<script>acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page</script>" +
		"</applicationRule><applicationRule><applicationRuleId>applicationRule-92</applicationRuleId><name>2</name>" +
		"<script>acl hello payload(0,6) -m bin 48656c6c6f0a</script></applicationRule>" + logging + "</loadBalancer>")
}
