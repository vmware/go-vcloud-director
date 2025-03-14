//go:build functional || cci || ALL

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
			InitialClassConfigOverrides: ccitypes.SupervisorNamespaceSpecInitialClassConfigOverrides{
				StorageClasses: []ccitypes.SupervisorNamespaceSpecInitialClassConfigOverridesStorageClass{
					{
						Name:  storagePolicy,
						Limit: "256Mi",
					},
				},
				Zones: []ccitypes.SupervisorNamespaceSpecInitialClassConfigOverridesZone{
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
		if err != nil && !strings.Contains(err.Error(), "404") {
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
