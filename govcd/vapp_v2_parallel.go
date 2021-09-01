package govcd

import (
	"fmt"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// addDefinedVmToRecomposeVapp adds the definition of an internal VM to the vApp recompose params
func addDefinedVmToRecomposeVapp(vappRecompose *types.ReComposeVAppParamsV2, vmDef *types.VmType) *types.ReComposeVAppParamsV2 {
	var vappRecomposeParams *types.ReComposeVAppParamsV2
	if vappRecompose != nil {
		vappRecomposeParams = vappRecompose
	}
	if vappRecomposeParams.CreateItem == nil {
		vappRecomposeParams.CreateItem = []*types.VmType{}
	}
	vappRecomposeParams.CreateItem = append(vappRecomposeParams.CreateItem, vmDef)

	return vappRecomposeParams
}

// addTemplateVmToRecomposeVapp adds the definition of a template-defined VM to the vApp recompose params
func addTemplateVmToRecomposeVapp(vappRecompose *types.ReComposeVAppParamsV2, vmDef *types.Reference) *types.ReComposeVAppParamsV2 {
	var vappRecomposeParams *types.ReComposeVAppParamsV2
	if vappRecompose != nil {
		vappRecomposeParams = vappRecompose
	}
	if vappRecomposeParams.SourcedItem == nil {
		vappRecomposeParams.SourcedItem = []*types.SourcedCompositionItemParam{}
	}
	vappRecomposeParams.SourcedItem = append(vappRecomposeParams.SourcedItem, &types.SourcedCompositionItemParam{Source: vmDef})
	return vappRecomposeParams
}

// addToRecomposeVapp adds multiple VM definitions to the vApp recompose params.
// The input items type is not known in advance. It is determined by probing the expected types
// and assigning them to the appropriate list
// This function is designed to interact with util.RunAfterCollection
func addToRecomposeVapp(vappRecompose *types.ReComposeVAppParamsV2, items map[string]interface{}) (*types.ReComposeVAppParamsV2, error) {
	var vappRecomposeParams *types.ReComposeVAppParamsV2 = vappRecompose

	for key, value := range items {
		c, isCreate := value.(*types.VmType)
		if isCreate {
			vappRecomposeParams = addDefinedVmToRecomposeVapp(vappRecomposeParams, c)
		} else {
			s, isSource := value.(*types.Reference)
			if isSource {
				vappRecomposeParams = addTemplateVmToRecomposeVapp(vappRecomposeParams, s)
			} else {
				return nil, fmt.Errorf("wanted only %T or %T - Found item (%s) of type '%T'", &types.VmType{}, &types.Reference{}, key, value)
			}
		}
	}
	return vappRecomposeParams, nil
}

// getClient returns the client connection from an interface{}
// It can get the client from either a *Client or *VCDClient input
func getClient(meta interface{}) (*Client, error) {
	client, ok := meta.(*Client)
	if ok {
		return client, nil
	}
	vcdClient, ok := meta.(*VCDClient)
	if ok {
		return &vcdClient.Client, nil
	}
	return nil, fmt.Errorf("parameter client is not of type %T - found type %T", &Client{}, meta)
}

// reconfigureParallelVapp is the main operator of a parallel VM deployment
// This function is called after the scheduler (util.RunWhenReady) has finished collecting input
func reconfigureParallelVapp(meta interface{}, vappHref string, vms map[string]interface{}) (interface{}, error) {
	client, err := getClient(meta)
	if err != nil {
		return nil, err
	}
	vapp, err := client.GetVappV2ByHref(vappHref)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vApp %s: %s", vappHref, err)
	}
	recomposeParams, err := addToRecomposeVapp(&types.ReComposeVAppParamsV2{
		Name:             vapp.VAppV2.Name,
		Description:      vapp.VAppV2.Description,
		PowerOn:          false,
		AllEULAsAccepted: true,
	}, vms)
	if err != nil {
		return nil, fmt.Errorf("error building vApp recompose params: %s", err)
	}
	err = vapp.RecomposeVAppV2(recomposeParams)
	if err != nil {
		return nil, fmt.Errorf("error recomposing vApp: %s", err)
	}
	return vapp, nil
}

// createParallelVMs is an auxiliary function that keeps calling util.RunWhenReady
// until the input is collected and the result is available.
// Returns util.ParallelOpState (the result of the scheduler), and interface{} (the result of the VM deployment, containing a vApp)
func createParallelVMs(input util.ParallelInput) (util.ParallelOpState, interface{}, error) {
	outcome := util.ParallelOpStateWaiting
	var result interface{}
	var err error
	runningDelay := 200 * time.Millisecond
	waitingDelay := 10 * time.Millisecond
	otherDelay := 100 * time.Millisecond
	// The loop ends when a final state (either error or completion) is returned.
	for outcome != util.ParallelOpStateDone &&
		outcome != util.ParallelOpStateRunTimeout &&
		outcome != util.ParallelOpStateCollectionTimeout &&
		outcome != util.ParallelOpStateFail {
		outcome, result, err = util.RunWhenReady(input)
		if err != nil {
			return outcome, nil, fmt.Errorf("[createParallelVMs] error returned %s", err)
		}
		if outcome == util.ParallelOpStateFail {
			if err == nil {
				err = fmt.Errorf("[createParallelVMs] unknown failure detected")
			}
			return outcome, nil, fmt.Errorf("[createParallelVMs] - outcome %s - failed %s", outcome, err)
		}
		if outcome == util.ParallelOpStateCollectionTimeout {
			return outcome, nil, fmt.Errorf("[createParallelVMs] timeout of %s exceeded ", input.CollectionTimeout)
		}
		if outcome == util.ParallelOpStateRunTimeout {
			return outcome, nil, fmt.Errorf("[createParallelVMs] timeout of %s exceeded ", input.RunTimeout)
		}
		// Reduce the amount of state polling, making the logging more manageable.
		// For a run that takes 1 minute, this sleep reduces the number of events from 14 million to 3,000
		delay := waitingDelay
		switch outcome {
		case util.ParallelOpStateRunning:
			delay = runningDelay
		case util.ParallelOpStateWaiting:
			delay = waitingDelay
		default:
			delay = otherDelay
		}
		time.Sleep(delay)
	}
	return outcome, result, err
}

// CreateParallelVm creates a VM by scheduling its deployment with the parallel scheduler
// client can be either *Client or *VCDClient
// vAppHref is the HREF of an already existing vApp
// vmName is the unique name of the VM within the vApp
// creation can be either a *types.Reference (for a VM created from a template) or *types.VMtype
// howMany is the number of VMs to be created with the parallel deployment. It must be the same for all the VMs in the group
func CreateParallelVm(client interface{}, vappHref, vmName string, creation interface{}, howMany int) (*VAppV2, error) {
	outcome, result, err := createParallelVMs(util.ParallelInput{
		Client:            client,
		GlobalId:          vappHref,
		ItemId:            vmName,
		NumExpectedItems:  howMany,
		Item:              creation,
		Run:               reconfigureParallelVapp,
		CollectionTimeout: 1 * time.Minute, // 1 minute to collect all inputs
		RunTimeout:        0,               // No run timeout
	})
	if err != nil {
		return nil, err
	}
	if outcome != util.ParallelOpStateDone {
		return nil, fmt.Errorf("received outcome %s", outcome)
	}
	vapp, ok := result.(*VAppV2)
	if !ok {
		return nil, fmt.Errorf("vapp structure not returned correctly. Received: %T", result)
	}
	return vapp, nil
}
