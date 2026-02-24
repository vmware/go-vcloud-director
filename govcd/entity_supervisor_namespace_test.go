//go:build functional || cci || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	"github.com/vmware/go-vcloud-director/v3/util"
	. "gopkg.in/check.v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (vcd *TestVCD) Test_SupervisorNamespace(check *C) {
	skipNonTm(vcd, check)
	vcd.skipIfSysAdmin(check) // The test is running in Org user mode

	regionName := vcd.config.Cci.Region
	vpcName := vcd.config.Cci.Vpc
	storagePolicy := vcd.config.Cci.StoragePolicy
	supervisorZoneName := vcd.config.Cci.SupervisorZone

	if regionName == "" || vpcName == "" || storagePolicy == "" || supervisorZoneName == "" {
		check.Skip("missing parameters in 'cci' test config")
	}

	projectCfg := &ccitypes.Project{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.ProjectKind,
			APIVersion: ccitypes.ProjectAPI + "/" + ccitypes.ProjectVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-supervisornamespace",
		},
		Spec: ccitypes.ProjectSpec{
			Description: check.TestName(),
		},
	}

	newProjectAddr, err := vcd.client.Client.GetEntityUrl(ccitypes.ProjectsURL)
	check.Assert(err, IsNil)
	check.Assert(newProjectAddr, NotNil)

	newProject := &ccitypes.Project{}
	// Create
	err = vcd.client.Client.PostEntity(newProjectAddr, nil, projectCfg, newProject, nil)
	check.Assert(err, IsNil)
	check.Assert(newProject, NotNil)
	check.Assert(newProject.Name, Equals, projectCfg.Name)

	// Get
	projectAddr, err := vcd.client.Client.GetEntityUrl(ccitypes.ProjectsURL, "/", newProject.Name)
	check.Assert(err, IsNil)
	check.Assert(projectAddr, NotNil)

	// defer project cleanup
	defer func() {
		p := &ccitypes.Project{}
		err = vcd.client.Client.GetEntity(projectAddr, nil, p, nil)
		if err != nil && !strings.Contains(err.Error(), "404") {
			err := vcd.client.Client.DeleteEntity(projectAddr, nil, nil)
			check.Assert(err, IsNil)
		}
	}()

	getProject := &ccitypes.Project{}
	err = vcd.client.Client.GetEntity(projectAddr, nil, getProject, nil)
	check.Assert(err, IsNil)
	check.Assert(getProject, NotNil)
	check.Assert(getProject.Name, Equals, projectCfg.Name)

	nsCfg := &ccitypes.SupervisorNamespace{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.SupervisorNamespaceKind,
			APIVersion: ccitypes.SupervisorNamespaceAPI + "/" + ccitypes.SupervisorNamespaceVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "govcd-test-ns",
			Namespace:    newProject.Name,
		},
		Spec: ccitypes.SupervisorNamespaceSpec{
			ClassName:   "small",
			Description: check.TestName(),
			ClassConfigOverrides: ccitypes.SupervisorNamespaceSpecClassConfigOverrides{
				StorageClasses: []ccitypes.SupervisorNamespaceSpecClassConfigOverridesStorageClass{
					{
						Name:  storagePolicy,
						Limit: "256Mi",
					},
				},
				Zones: []ccitypes.SupervisorNamespaceSpecClassConfigOverridesZone{
					{
						CpuLimit:          "200M",
						CpuReservation:    "1M",
						MemoryLimit:       "256Mi",
						MemoryReservation: "1Mi",
						Name:              supervisorZoneName,
					},
				},
			},
			RegionName: regionName,
			VpcName:    vpcName,
		},
	}

	newNsAddr, err := vcd.client.Client.GetEntityUrl(fmt.Sprintf(ccitypes.SupervisorNamespacesURL, newProject.Name))
	check.Assert(err, IsNil)
	check.Assert(newNsAddr, NotNil)

	newNs := &ccitypes.SupervisorNamespace{}
	// Create
	err = vcd.client.Client.PostEntity(newNsAddr, nil, nsCfg, newNs, nil)
	check.Assert(err, IsNil)
	check.Assert(newNs, NotNil)
	check.Assert(newNs.GenerateName, Equals, nsCfg.GenerateName)
	check.Assert(strings.HasPrefix(newNs.Name, nsCfg.GenerateName), Equals, true)

	// Get
	nsAddr, err := vcd.client.Client.GetEntityUrl(fmt.Sprintf(ccitypes.SupervisorNamespacesURL, newProject.Name), "/", newNs.Name)
	check.Assert(err, IsNil)
	check.Assert(nsAddr, NotNil)

	_, err = waitForEntityState(&vcd.client.Client, nsAddr, []string{"CREATING", "WAITING"}, []string{"CREATED"})
	check.Assert(err, IsNil)

	// defer namespace cleanup
	defer func() {
		p := &ccitypes.SupervisorNamespace{}
		err = vcd.client.Client.GetEntity(nsAddr, nil, p, nil)
		if err != nil && !strings.Contains(err.Error(), "400") && !strings.Contains(err.Error(), "404") {
			err := vcd.client.Client.DeleteEntity(nsAddr, nil, nil)
			check.Assert(err, IsNil)
			_, err = waitForEntityState(&vcd.client.Client, nsAddr, []string{"DELETING", "WAITING"}, []string{"DELETED"})
			check.Assert(err, IsNil)
		}
	}()

	getNs := &ccitypes.SupervisorNamespace{}
	err = vcd.client.Client.GetEntity(nsAddr, nil, getNs, nil)
	check.Assert(err, IsNil)
	check.Assert(getNs, NotNil)
	check.Assert(getNs.Name, Equals, newNs.Name)

	// Put (update)
	updatedDescription := check.TestName() + "-updated"
	putPayload := &ccitypes.SupervisorNamespace{
		TypeMeta:   getNs.TypeMeta,
		ObjectMeta: getNs.ObjectMeta,
		Spec:       getNs.Spec,
		Status:     getNs.Status,
	}
	putPayload.Spec.Description = updatedDescription

	putNs := &ccitypes.SupervisorNamespace{}
	err = vcd.client.Client.PutEntity(nsAddr, nil, putPayload, putNs, nil)
	check.Assert(err, IsNil)
	check.Assert(putNs, NotNil)
	check.Assert(putNs.Spec.Description, Equals, updatedDescription)

	// Verify update is realized
	_, err = waitForEntityCondition(&vcd.client.Client, nsAddr, "Realized", "True")
	check.Assert(err, IsNil)

	// Verify update via Get
	getNsAfterPut := &ccitypes.SupervisorNamespace{}
	err = vcd.client.Client.GetEntity(nsAddr, nil, getNsAfterPut, nil)
	check.Assert(err, IsNil)
	check.Assert(getNsAfterPut.Spec.Description, Equals, updatedDescription)

	// Delete
	err = vcd.client.Client.DeleteEntity(nsAddr, nil, nil)
	check.Assert(err, IsNil)
	_, err = waitForEntityState(&vcd.client.Client, nsAddr, []string{"DELETING", "WAITING"}, []string{"DELETED"})
	check.Assert(err, IsNil)

	// Check it is deleted
	err = vcd.client.Client.GetEntity(nsAddr, nil, getProject, nil)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "404"), Equals, true)

	// Delete project
	err = vcd.client.Client.DeleteEntity(projectAddr, nil, nil)
	check.Assert(err, IsNil)

	// Check it is deleted
	err = vcd.client.Client.GetEntity(projectAddr, nil, getProject, nil)
	check.Assert(strings.Contains(err.Error(), "404"), Equals, true)
}

func getSupervisorNamespaceState(client *Client, urlRef *url.URL) (*ccitypes.SupervisorNamespace, error) {
	entityStatus := ccitypes.SupervisorNamespace{}

	err := client.GetEntity(urlRef, nil, &entityStatus, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting entity status from URL '%s': %s", urlRef.String(), err)
	}
	return &entityStatus, nil
}

func waitForEntityState(client *Client, addr *url.URL, pendingStates, targetStates []string) (any, error) {
	// constants that are used for tracking entity state after it is being created
	const (
		stateWaitTimeout  = 20 * time.Minute
		stateWaitDelay    = 10 * time.Second
		statePollInterval = 10 * time.Second
	)

	startTime := time.Now()
	endTime := startTime.Add(stateWaitTimeout)

	util.Logger.Printf("[DEBUG] waitForEntityState - expecting states %s for endpoint %s", strings.Join(targetStates, ","), addr.String())
	util.Logger.Printf("[DEBUG] waitForEntityState - sleeping %s before checking state at %s", stateWaitDelay, addr.String())
	time.Sleep(stateWaitDelay)

	stepCount := 0
	for {
		stepCount++
		if time.Now().After(endTime) {
			util.Logger.Printf("[DEBUG] waitForEntityState - timeout reached for entity %s", addr.String())
			return nil, fmt.Errorf("waitForEntityState - exceeded waiting for entity %s to reach %s time after attempt %d", addr.String(), strings.Join(targetStates, ","), stepCount)
		}

		var entityState string
		cciEntity, err := getSupervisorNamespaceState(client, addr)

		if err != nil && !strings.Contains(err.Error(), "404") {
			return nil, fmt.Errorf("error retrieving CCI entity %s state: %s", addr.String(), err)
		}

		if err != nil && strings.Contains(err.Error(), "404") {
			entityState = "DELETED"
		} else {
			entityState = strings.ToUpper(cciEntity.Status.Phase)
		}

		if strings.EqualFold(entityState, "ERROR") {
			return nil, fmt.Errorf("waitForEntityState %s is in an ERROR state", addr.String())
		}

		util.Logger.Printf("[DEBUG] waitForEntityState - %s current phase at step %d is %s", addr.String(), stepCount, entityState)

		// Check if the entity is in a target state
		if slices.Contains(targetStates, entityState) {
			util.Logger.Printf("[DEBUG] waitForEntityState - %s reached %s at step %d", addr.String(), entityState, stepCount)
			return cciEntity, nil
		}

		// Check if the entity is in a pending state, and if so, wait and continue
		if slices.Contains(pendingStates, entityState) {
			util.Logger.Printf("[DEBUG] waitForEntityState - sleeping %s before next attempt to retrieve %s state", statePollInterval, addr.String())
			time.Sleep(statePollInterval)
			continue // Only continue if in a pending state
		}
	}
}

func waitForEntityCondition(client *Client, addr *url.URL, conditionType string, conditionValue string) (any, error) {
	const (
		conditionWaitTimeout  = 20 * time.Minute
		conditionWaitDelay    = 10 * time.Second
		conditionPollInterval = 10 * time.Second
	)

	startTime := time.Now()
	endTime := startTime.Add(conditionWaitTimeout)

	util.Logger.Printf("[DEBUG] waitForEntityCondition - expecting condition %s to be %s for entity %s", conditionType, conditionValue, addr.String())
	util.Logger.Printf("[DEBUG] waitForEntityCondition - sleeping %s before checking entiity %s", conditionWaitDelay, addr.String())
	time.Sleep(conditionWaitDelay)

	stepCount := 0
	for {
		stepCount++
		if time.Now().After(endTime) {
			util.Logger.Printf("[DEBUG] waitForEntityCondition - timeout reached for entity %s", addr.String())
			return nil, fmt.Errorf("waitForEntityCondition - exceeded waiting for condition %s to be %s for entity %s after attempt %d", conditionType, conditionValue, addr.String(), stepCount)
		}

		cciEntity, err := getSupervisorNamespaceState(client, addr)
		if err != nil {
			return nil, fmt.Errorf("error retrieving entity %s: %s", addr.String(), err)
		}

		var entityCondition *ccitypes.SupervisorNamespaceStatusConditions
		for _, condition := range cciEntity.Status.Conditions {
			if strings.EqualFold(condition.Type, conditionType) {
				entityCondition = &condition
				break
			}
		}

		if entityCondition != nil {
			if strings.EqualFold(entityCondition.Status, conditionValue) {
				util.Logger.Printf("[DEBUG] waitForEntityCondition - condition %s reached %s at step %d for entity %s", conditionType, conditionValue, stepCount, addr.String())
				return cciEntity, nil
			}
			util.Logger.Printf("[DEBUG] waitForEntityCondition - current condition %s at step %d is %s for entity %s", conditionType, stepCount, conditionValue, addr.String())
		} else {
			util.Logger.Printf("[DEBUG] waitForEntityCondition - condition %s not found for entity %s at step %d", conditionType, addr.String(), stepCount)
		}

		util.Logger.Printf("[DEBUG] waitForEntityCondition - sleeping %s before next attempt to retrieve entity %s", conditionPollInterval, addr.String())
		time.Sleep(conditionPollInterval)
	}
}
