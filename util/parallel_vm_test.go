package util

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

type mockVm struct {
	ID         string
	vappId     string
	name       string
	templateId string
}

type mockVapp struct {
	ID string
	//name string
	vms map[string]mockVm
}

/*
resource "vcd_vapp_vm_v2" "test_vm" {
	name         = "test_vm"
	vapp_id      = data.vcd_vapp_v2.my-vapp.id
	template_id  = data.vcd_catalogitem.my_template.id
  parallel_vms = 3
}
*/

func createMockVm(vappId, vmName, templateId string, howMany int) error {
	outcome, result, err := createParallelVMs(ParallelInput{
		GlobalId:         vappId,
		ItemId:           vmName,
		NumExpectedItems: howMany,
		Item: mockVm{
			name:       vmName,
			vappId:     vappId,
			templateId: templateId,
		},
		Run:               parallelVappCompose,
		CollectionTimeout: 10 * time.Second,
		RunTimeout:        100 * time.Second,
	})
	if err != nil {
		return err
	}
	if outcome != ParallelOpStateDone {
		return fmt.Errorf("received outcome %s", outcome)
	}
	vapp, ok := result.(mockVapp)
	if !ok {
		return fmt.Errorf("vapp structure not returned correctly. Received: %s", reflect.TypeOf(result))
	}
	vm, found := vapp.vms[vmName]
	if !found {
		return fmt.Errorf("vm '%s' was not created", vmName)
	}
	fmt.Printf("created VM %s -> ID %s\n", vm.name, vm.ID)
	return nil
}

func parallelVappCompose(client interface{}, globalId string, items map[string]interface{}) (interface{}, error) {
	var result mockVapp
	var vms = make(map[string]mockVm)
	time.Sleep(3 * time.Second)
	fmt.Println("compose vApp")
	count := 0
	for key, value := range items {
		count++
		vm, ok := value.(mockVm)
		if !ok {
			return nil, fmt.Errorf("item '%s' - expected type mockVm - found %s", key, reflect.TypeOf(value))
		}
		vm.ID = fmt.Sprintf("%d", count)
		vms[key] = vm
	}
	result.vms = vms
	return result, nil
}

func TestCreateVMs(t *testing.T) {
	var (
		templateName = "myTemplate"
		vappId       = "globalVapp"
		//vmNames      = []string{"one", "two", "three"}
		vmNames = []string{"Doc", "Grumpy", "Bashful", "Sleepy", "Happy", "Sneezy", "Dopey"}
	)

	wg := sync.WaitGroup{}
	wg.Add(len(vmNames))

	for _, vmName := range vmNames {
		go func(name string) {
			defer wg.Done()
			err := createMockVm(vappId, name, templateName, len(vmNames))
			if err != nil {
				fmt.Printf("%s", err)
				os.Exit(1)
			}
		}(vmName)
	}
	wg.Wait()
}

func createParallelVMs(input ParallelInput) (ParallelOpState, interface{}, error) {
	outcome := ParallelOpStateWaiting
	var result interface{}
	var err error
	for outcome != ParallelOpStateDone && outcome != ParallelOpStateRunTimeout && outcome != ParallelOpStateCollectionTimeout && outcome != ParallelOpStateFail {
		outcome, result, err = RunWhenReady(input)
		debugPrintf("[createParallelVMs] item %s - %s\n", input.ItemId, outcome)
		if err != nil {
			return outcome, nil, fmt.Errorf("[createParallelVMs] error returned %s", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return outcome, result, err
}
