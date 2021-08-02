package util

import (
	"fmt"
	"os"
	"time"
)

var debugging = os.Getenv("parallel_debug") != ""

// parallelInfo is the information needed to run a parallel operation
type parallelInfo struct {
	data                map[string]interface{} // collection of input from all requesters
	isRunning           bool                   // has started running
	finished            bool                   // has finished running
	result              interface{}            // The result returned by the wanted function
	err                 error                  // error returned by the wanted function
	runStartTime        time.Time              // when the run started
	collectionStartTime time.Time              // when the data collection started
}

// ParallelInput is what each client needs to pass to the scheduler to run a parallel task
type ParallelInput struct {
	Client            interface{}
	GlobalId          string             // identification of the job, such as the object to create from all the parts
	ItemId            string             // identification of the Client, such as the definition of a part of the final object
	HowMany           int                // how many objects must be created. The scheduler will wait until as many requests arrive
	Item              interface{}        // the portion of data being passed to the run function
	Run               RunAfterCollection // The function that will ultimately do the work
	CollectionTimeout time.Duration      // how long to wait for data collection (0 = forever)
	RunTimeout        time.Duration      // how long to wait for the Run (0 = forever)
}

// parallelMutexKV is a lock mutex that will prevent clients from overriding each other requests
var parallelMutexKV = NewMutexKVSilent()

// ResultOutcome is a type that defines the standard outcomes of a scheduler request
type ResultOutcome string

// RunAfterCollection is a type of function that, given a collection of items, will perform an operation and return an object
type RunAfterCollection func(client interface{}, globalId string, data map[string]interface{}) (interface{}, error)

// Standard outcome from RunWhenReady
const (
	OutcomeDone              ResultOutcome = "done"
	OutcomeWaiting           ResultOutcome = "waiting"
	OutcomeRunTimeout        ResultOutcome = "run-timeout"
	OutcomeCollectionTimeout ResultOutcome = "collection-timeout"
	OutcomeFail              ResultOutcome = "fail"
	OutcomeRunning           ResultOutcome = "running"
)

// concurrentData is the private repository of the data requests
var concurrentData = make(map[string]parallelInfo)

// RunWhenReady is the scheduler that collects the data from clients, and runs the request when all the items have been collected
func RunWhenReady(input ParallelInput) (ResultOutcome, interface{}, error) {
	parallelMutexKV.Lock(input.GlobalId)
	defer parallelMutexKV.Unlock(input.GlobalId)
	info := concurrentData[input.GlobalId]
	debugPrintf("entering RunWhenReady: %s - %s (%d)\n", input.GlobalId, input.ItemId, len(info.data))
	if len(info.data) == input.HowMany {
		if info.finished {
			debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s\n", input.GlobalId, input.ItemId, OutcomeDone)
			return OutcomeDone, info.result, info.err
		}
		if info.isRunning {
			if input.RunTimeout > 0 && time.Since(info.runStartTime) > input.RunTimeout {
				debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s\n", input.GlobalId, input.ItemId, OutcomeRunTimeout)
				return OutcomeRunTimeout, nil, nil
			}
			debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s\n", input.GlobalId, input.ItemId, OutcomeRunning)
			return OutcomeRunning, nil, nil
		}
		info.isRunning = true
		info.runStartTime = time.Now()
		concurrentData[input.GlobalId] = info

		result, err := input.Run(input.Client, input.GlobalId, info.data)
		info.finished = true
		info.result = result
		info.err = err
		concurrentData[input.GlobalId] = info
		return OutcomeDone, result, err
	} else {
		if len(info.data) == 0 {
			info.collectionStartTime = time.Now()
			info.data = make(map[string]interface{})
			debugPrintf("initializing map %s - %s \n", input.GlobalId, input.ItemId)
		}
		if input.CollectionTimeout > 0 && time.Since(info.collectionStartTime) > input.CollectionTimeout {
			debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s\n", input.GlobalId, input.ItemId, OutcomeCollectionTimeout)
			return OutcomeCollectionTimeout, nil, nil
		}
		_, exists := info.data[input.ItemId]
		if !exists {
			info.data[input.ItemId] = input.Item
			concurrentData[input.GlobalId] = info
			debugPrintf("adding item: %s - %s  (%d)\n", input.GlobalId, input.ItemId, len(info.data))
		}
	}
	debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s\n", input.GlobalId, input.ItemId, OutcomeWaiting)
	return OutcomeWaiting, nil, nil
}

func debugPrintf(format string, args ...interface{}) {
	if debugging {
		fmt.Printf(format, args...)
	}
}
