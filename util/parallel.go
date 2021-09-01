package util

import (
	"fmt"
	"os"
	"time"
)

var parallelDebugging = os.Getenv("parallel_debug") != ""
var parallelDebugOutput = os.Getenv("parallel_debug_output")

// parallelInfo is the information needed to run a parallel operation
type parallelInfo struct {
	data                map[string]interface{}   // collection of input from all requesters
	details             map[string]ParallelInput // the input that was passed from each requester
	isRunning           bool                     // has started running
	runningOn           string                   // which node has activated the running
	finished            bool                     // has finished running
	result              interface{}              // The result returned by the wanted function
	err                 error                    // error returned by the wanted function
	runStartTime        time.Time                // when the run started
	collectionStartTime time.Time                // when the data collection started
}

// ParallelInput is what each client needs to pass to the scheduler to run a parallel task
type ParallelInput struct {
	Client            interface{}        // valid connection
	GlobalId          string             // identification of the job, such as the object to create from all the parts
	ItemId            string             // identification of the component, such as the definition of a part of the final object
	NumExpectedItems  int                // how many objects must be created. The scheduler will wait until as many requests arrive
	Item              interface{}        // the portion of data being passed to the run function
	Run               RunAfterCollection // The function that will ultimately do the work
	CollectionTimeout time.Duration      // how long to wait for data collection (0 = forever)
	RunTimeout        time.Duration      // how long to wait for the Run (0 = forever)
}

// parallelMutexKV is a lock mutex that will prevent clients from overriding each other requests
var parallelMutexKV = NewMutexKVSilent()

// ParallelOpState is a type that defines the standard outcome of a scheduler request
type ParallelOpState string

// RunAfterCollection is a type of function that, given a collection of items, will perform an operation and return an object
type RunAfterCollection func(client interface{}, globalId string, data map[string]interface{}) (interface{}, error)

// Standard outcome from RunWhenReady
const (
	ParallelOpStateDone              ParallelOpState = "done"
	ParallelOpStateWaiting           ParallelOpState = "waiting"
	ParallelOpStateRunTimeout        ParallelOpState = "run-timeout"
	ParallelOpStateCollectionTimeout ParallelOpState = "collection-timeout"
	ParallelOpStateFail              ParallelOpState = "fail"
	ParallelOpStateRunning           ParallelOpState = "running"
)

// concurrentData is the private repository of the data requests
var concurrentData = make(map[string]parallelInfo)

// stateCounters collects the number of times a state appears for a given group identifier
var stateCounters = make(map[string]uint64)

// incrementedCounter provides a counter for each state within a given group ID
func incrementedCounter(id string, outcome ParallelOpState) uint64 {
	key := id + string(outcome)
	parallelMutexKV.Lock(key)
	defer parallelMutexKV.Unlock(key)
	_, exists := stateCounters[key]
	if !exists {
		stateCounters[key] = 0
	}
	stateCounters[key]++
	return stateCounters[key]
}

// RunWhenReady is the scheduler that collects the data from clients, and runs the request when all the items have been collected
func RunWhenReady(input ParallelInput) (ParallelOpState, interface{}, error) {
	parallelMutexKV.Lock(input.GlobalId)
	isLocked := true
	defer func() {
		if isLocked {
			parallelMutexKV.Unlock(input.GlobalId)
		}
	}()
	info := concurrentData[input.GlobalId]
	err := validateParallelData(input, info)
	if err != nil {
		debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s [%d] (%s)\n", input.GlobalId, input.ItemId,
			ParallelOpStateFail, incrementedCounter(input.GlobalId, ParallelOpStateFail), time.Since(info.collectionStartTime))
		return ParallelOpStateFail, nil, err
	}

	debugPrintf("entering RunWhenReady: %s - %s (%d)\n", input.GlobalId, input.ItemId, len(info.data))
	if len(info.data) == input.NumExpectedItems {
		if info.finished {
			debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s [%d] (%s)\n", input.GlobalId, input.ItemId,
				ParallelOpStateDone, incrementedCounter(input.GlobalId, ParallelOpStateDone), time.Since(info.runStartTime))
			return ParallelOpStateDone, info.result, info.err
		}
		if info.isRunning {
			if input.RunTimeout > 0 && time.Since(info.runStartTime) > input.RunTimeout {
				debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s [%d] (%s)\n", input.GlobalId, input.ItemId,
					ParallelOpStateRunTimeout, incrementedCounter(input.GlobalId, ParallelOpStateRunTimeout), time.Since(info.runStartTime))
				return ParallelOpStateRunTimeout, nil, nil
			}
			debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s on %s [%d] (%s)\n", input.GlobalId, input.ItemId,
				ParallelOpStateRunning, info.runningOn, incrementedCounter(input.GlobalId, ParallelOpStateRunning), time.Since(info.runStartTime))
			return ParallelOpStateRunning, nil, nil
		}
		info.isRunning = true
		info.runningOn = input.ItemId
		info.runStartTime = time.Now()
		concurrentData[input.GlobalId] = info

		// Unlock the mutex before running the main routine, so that callers can get the "OutcomeRunning" state
		parallelMutexKV.Unlock(input.GlobalId)
		isLocked = false
		result, err := input.Run(input.Client, input.GlobalId, info.data)

		// Achieve a lock again, to be able to modify the concurrent data
		parallelMutexKV.Lock(input.GlobalId)
		isLocked = true
		info.finished = true
		info.result = result
		info.err = err
		concurrentData[input.GlobalId] = info
		return ParallelOpStateDone, result, err
	} else {
		if len(info.data) == 0 {
			info.collectionStartTime = time.Now()
			info.data = make(map[string]interface{})
			info.details = make(map[string]ParallelInput)
			debugPrintf("initializing map %s - %s \n", input.GlobalId, input.ItemId)
		}
		if input.CollectionTimeout > 0 && time.Since(info.collectionStartTime) > input.CollectionTimeout {
			debugPrintf("exiting RunWhenReady: %s - %s - outcome: [%d] %s\n", input.GlobalId, input.ItemId, incrementedCounter(input.GlobalId, ParallelOpStateCollectionTimeout), ParallelOpStateCollectionTimeout)
			return ParallelOpStateCollectionTimeout, nil, nil
		}
		_, exists := info.data[input.ItemId]
		if !exists {
			info.data[input.ItemId] = input.Item
			info.details[input.ItemId] = input
			concurrentData[input.GlobalId] = info
			debugPrintf("adding item: %s - %s  (%d)\n", input.GlobalId, input.ItemId, len(info.data))
		}
	}
	debugPrintf("exiting RunWhenReady: %s - %s - outcome: %s [%d] (%s)\n", input.GlobalId, input.ItemId,
		ParallelOpStateWaiting, incrementedCounter(input.GlobalId, ParallelOpStateWaiting), time.Since(info.collectionStartTime))
	return ParallelOpStateWaiting, nil, nil
}

// debugPrintf logs the messages from the parallel engine when `debugging` is enabled
// WARNING: it may write a large quantity of data in the logs
func debugPrintf(format string, args ...interface{}) {
	if parallelDebugOutput == "" {
		parallelDebugOutput = "screen"
	}
	if parallelDebugging {
		switch parallelDebugOutput {
		case "log":
			Logger.Printf(format, args...)
		case "screen":
			fmt.Printf(format, args...)
		}
	}
}

func validateParallelData(latestInput ParallelInput, info parallelInfo) error {
	if latestInput.NumExpectedItems == 0 {
		return fmt.Errorf("latest input has no number of expected items")
	}
	for _, item := range info.details {
		if item.NumExpectedItems != latestInput.NumExpectedItems {
			return fmt.Errorf("inconsistent number of expected item detected. Previous was %d. - Current is %d", item.NumExpectedItems, latestInput.NumExpectedItems)
		}
		if item.CollectionTimeout != 0 && item.CollectionTimeout != latestInput.CollectionTimeout {
			return fmt.Errorf("inconsistent collection timeout detected. Previous was %s - Current is %s", item.CollectionTimeout, latestInput.CollectionTimeout)
		}
	}
	return nil
}
