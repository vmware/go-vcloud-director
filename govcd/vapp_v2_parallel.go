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

// AddToRecomposeVapp adds multiple VM definitions to the vApp recompose params.
// The input items type is not known in advance. It is determined by probing the expected types
// and assigning them to the appropriate list
// This function is designed to interact with util.RunAfterCollection
func AddToRecomposeVapp(vappRecompose *types.ReComposeVAppParamsV2, items map[string]interface{}) (*types.ReComposeVAppParamsV2, error) {
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

func ReconfigureParallelVapp(meta interface{}, vappHref string, vms map[string]interface{}) (interface{}, error) {
	client, ok := meta.(*Client)
	if !ok {
		return nil, fmt.Errorf("parameter client is not of type %T - found type %T", &Client{}, meta)
	}
	vapp, err := client.GetVappV2ByHref(vappHref)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vApp %s: %s", vappHref, err)
	}
	recomposeParams, err := AddToRecomposeVapp(&types.ReComposeVAppParamsV2{
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

func CreateParallelVMs(input util.ParallelInput) (util.ResultOutcome, interface{}, error) {
	outcome := util.OutcomeWaiting
	var result interface{}
	var err error
	for outcome != util.OutcomeDone && outcome != util.OutcomeRunTimeout && outcome != util.OutcomeCollectionTimeout {
		outcome, result, err = util.RunWhenReady(input)
		if err != nil {
			return outcome, nil, fmt.Errorf("[CreateParallelVMs] error returned %s", err)
		}
	}
	return outcome, result, err
}

func CreateParallelVm(client interface{}, vappId, vmName string, creation interface{}, howMany int) (*VAppV2, error) {
	outcome, result, err := CreateParallelVMs(util.ParallelInput{
		Client:            client,
		GlobalId:          vappId,
		ItemId:            vmName,
		HowMany:           howMany,
		Item:              creation,
		Run:               ReconfigureParallelVapp,
		CollectionTimeout: 10 * time.Second,
		RunTimeout:        100 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	if outcome != util.OutcomeDone {
		return nil, fmt.Errorf("received outcome %s", outcome)
	}
	vapp, ok := result.(*VAppV2)
	if !ok {
		return nil, fmt.Errorf("vapp structure not returned correctly. Received: %T", result)
	}
	return vapp, nil
}
