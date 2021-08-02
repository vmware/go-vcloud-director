package util

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func runUntilDone(input ParallelInput) (ResultOutcome, interface{}, error) {
	outcome := OutcomeWaiting
	var result interface{}
	var err error
	for outcome != OutcomeDone && outcome != OutcomeRunTimeout && outcome != OutcomeCollectionTimeout {
		outcome, result, err = RunWhenReady(input)
		debugPrintf("[runUntilDone] item %s - %s\n", input.ItemId, outcome)
		if err != nil {
			return outcome, nil, fmt.Errorf("error returned %s", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return outcome, result, err
}

func TestRunWhenReady(t *testing.T) {

	var finalResult interface{}

	var run = func(client interface{}, globalId string, items map[string]interface{}) (interface{}, error) {
		var result []interface{}
		time.Sleep(3 * time.Second)
		fmt.Println("Final run")
		for key, value := range items {
			result = append(result, fmt.Sprintf("%v %v", key, value))
		}
		return result, nil
	}
	wg := sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(count int) {
			defer wg.Done()
			outcome, result, err := runUntilDone(
				ParallelInput{
					GlobalId:          "TESTGLOBAL",
					ItemId:            fmt.Sprintf("item%d", count),
					HowMany:           10,
					Item:              fmt.Sprintf("<ITEM %d>", count),
					Run:               run,
					CollectionTimeout: 10 * time.Second,
					RunTimeout:        100 * time.Second,
				})
			if err != nil {
				fmt.Printf("[%d] error received: %s ", count, err)
				os.Exit(1)
			}
			if result != nil {
				if finalResult == nil {
					finalResult = result
				}
				t.Logf("[%d] %s %d\n", count, outcome, len(result.([]interface{})))
			} else {

				t.Logf("[%d] %s [NO RESULT]\n", count, outcome)
			}
		}(i)
	}
	wg.Wait()
	if finalResult != nil {
		for _, item := range finalResult.([]interface{}) {
			fmt.Println(item)
		}
	} else {
		t.Fatalf("globalResult not set")
	}
}
