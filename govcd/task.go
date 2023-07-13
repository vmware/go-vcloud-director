/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type Task struct {
	Task   *types.Task
	client *Client
}

func NewTask(cli *Client) *Task {
	return &Task{
		Task:   new(types.Task),
		client: cli,
	}
}

const errorRetrievingTask = "error retrieving task"

// getErrorMessage composes a new error message, if the error is not nil.
// The message is made of the error itself + the information from the task's Error component.
// See:
//
//	https://code.vmware.com/apis/220/vcloud#/doc/doc/types/TaskType.html
//	https://code.vmware.com/apis/220/vcloud#/doc/doc/types/ErrorType.html
func (task *Task) getErrorMessage(err error) string {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	if task.Task.Error != nil {
		errorMessage += " [" +
			fmt.Sprintf("%d:%s",
				task.Task.Error.MajorErrorCode,   // The MajorError is a numeric code
				task.Task.Error.MinorErrorCode) + // The MinorError is a string with a generic definition of the error
			"] - " + task.Task.Error.Message
	}
	return errorMessage
}

// Refresh retrieves a fresh copy of the task
func (task *Task) Refresh() error {

	if task.Task == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	refreshUrl := urlParseRequestURI(task.Task.HREF)

	req := task.client.NewRequest(map[string]string{}, http.MethodGet, *refreshUrl, nil)

	resp, err := checkResp(task.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("%s: %s", errorRetrievingTask, err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	task.Task = &types.Task{}

	if err = decodeBody(types.BodyTypeXML, resp, task.Task); err != nil {
		return fmt.Errorf("error decoding task response: %s", task.getErrorMessage(err))
	}

	// The request was successful
	return nil
}

// InspectionFunc is a callback function that can be passed to task.WaitInspectTaskCompletion
// to perform user defined operations
// * task is the task object being processed
// * howManyTimes is the number of times the task has been refreshed
// * elapsed is how much time since the task was initially processed
// * first is true if this is the first refresh of the task
// * last is true if the function is being called for the last time.
type InspectionFunc func(task *types.Task, howManyTimes int, elapsed time.Duration, first, last bool)

// TaskMonitoringFunc can run monitoring operations on a task
type TaskMonitoringFunc func(*types.Task)

// WaitInspectTaskCompletion is a customizable version of WaitTaskCompletion.
// Users can define the sleeping duration and an optional callback function for
// extra monitoring.
func (task *Task) WaitInspectTaskCompletion(inspectionFunc InspectionFunc, delay time.Duration) error {

	if task.Task == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	taskMonitor := os.Getenv("GOVCD_TASK_MONITOR")
	howManyTimesRefreshed := 0
	startTime := time.Now()
	for {
		howManyTimesRefreshed++
		elapsed := time.Since(startTime)
		err := task.Refresh()
		if err != nil {
			return fmt.Errorf("%s : %s", errorRetrievingTask, err)
		}

		// If an inspection function is provided, we pass information about the task processing:
		// * the task itself
		// * the number of iterations
		// * how much time we have spent querying the task so far
		// * whether this is the first iteration
		// * whether this is the last iteration
		// It's up to the inspection function to render this information fittingly.

		// If task is not in a waiting status we're done, check if there's an error and return it.
		if !isTaskRunning(task.Task.Status) {
			if inspectionFunc != nil {
				inspectionFunc(task.Task,
					howManyTimesRefreshed,
					elapsed,
					howManyTimesRefreshed == 1,              // first
					isTaskCompleteOrError(task.Task.Status), // last
				)
			}
			if task.Task.Status == "error" {
				return fmt.Errorf("task did not complete successfully: %s", task.getErrorMessage(err))
			}
			return nil
		}

		// If the environment variable "GOVCD_TASK_MONITOR" is set, its value
		// will be used to choose among pre-defined InspectionFunc
		if inspectionFunc == nil {
			if taskMonitor != "" {
				switch taskMonitor {
				case "log":
					inspectionFunc = LogTask // writes full task details to the log
				case "show":
					inspectionFunc = ShowTask // writes full task details to the screen
				case "simple_log":
					inspectionFunc = SimpleLogTask // writes a summary line for the task to the log
				case "simple_show":
					inspectionFunc = SimpleShowTask // writes a summary line for the task to the screen
				case "minimal_show":
					inspectionFunc = MinimalShowTask // writes a dot for each iteration, or "+" for success, "-" for failure
				}
			}
		}
		if inspectionFunc != nil {
			inspectionFunc(task.Task,
				howManyTimesRefreshed,
				elapsed,
				howManyTimesRefreshed == 1, // first
				false,                      // last
			)
		}

		// Sleep for a given period and try again.
		time.Sleep(delay)
	}
}

// WaitTaskCompletion checks the status of the task every 3 seconds and returns when the
// task is either completed or failed
func (task *Task) WaitTaskCompletion() error {
	return task.WaitInspectTaskCompletion(nil, 3*time.Second)
}

// GetTaskProgress retrieves the task progress as a string
func (task *Task) GetTaskProgress() (string, error) {
	if task.Task == nil {
		return "", fmt.Errorf("cannot refresh, Object is empty")
	}

	err := task.Refresh()
	if err != nil {
		return "", fmt.Errorf("error retrieving task: %s", err)
	}

	if task.Task.Status == "error" {
		return "", fmt.Errorf("task did not complete successfully: %s", task.getErrorMessage(err))
	}

	return strconv.Itoa(task.Task.Progress), nil
}

// CancelTask attempts a task cancellation, returning an error if cancellation fails
func (task *Task) CancelTask() error {
	cancelTaskURL, err := url.ParseRequestURI(task.Task.HREF + "/action/cancel")
	if err != nil {
		util.Logger.Printf("[CancelTask] Error parsing task request URI %v: %s", cancelTaskURL.String(), err)
		return err
	}

	request := task.client.NewRequest(map[string]string{}, http.MethodPost, *cancelTaskURL, nil)
	_, err = checkResp(task.client.Http.Do(request))
	if err != nil {
		util.Logger.Printf("[CancelTask] Error cancelling task  %v: %s", cancelTaskURL.String(), err)
		return err
	}
	util.Logger.Printf("[CancelTask] task %s CANCELED\n", task.Task.ID)
	return nil
}

// ResourceInProgress returns true if any of the provided tasks is still running
func ResourceInProgress(tasksInProgress *types.TasksInProgress) bool {
	util.Logger.Printf("[TRACE] ResourceInProgress - has tasks %v\n", tasksInProgress != nil)
	if tasksInProgress == nil {
		return false
	}
	tasks := tasksInProgress.Task
	for _, task := range tasks {
		if isTaskCompleteOrError(task.Status) {
			continue
		}
		if isTaskRunning(task.Status) {
			return true
		}
	}
	return false
}

// ResourceComplete return true is none of its tasks are running
func ResourceComplete(tasksInProgress *types.TasksInProgress) bool {
	util.Logger.Printf("[TRACE] ResourceComplete - has tasks %v\n", tasksInProgress != nil)
	return !ResourceInProgress(tasksInProgress)
}

// WaitResource waits for the tasks associated to a given resource to complete
func WaitResource(refresh func() (*types.TasksInProgress, error)) error {
	util.Logger.Printf("[TRACE] WaitResource \n")
	tasks, err := refresh()
	if tasks == nil {
		return nil
	}
	for err == nil {
		time.Sleep(time.Second)
		tasks, err = refresh()
		if err != nil {
			return err
		}
		if tasks == nil || ResourceComplete(tasks) {
			return nil
		}
	}
	return nil
}

// SkimTasksList checks a list of tasks and returns a list of tasks still in progress and a list of failed ones
func SkimTasksList(taskList []*Task) ([]*Task, []*Task, error) {
	return SkimTasksListMonitor(taskList, nil)
}

// SkimTasksListMonitor checks a list of tasks and returns a list of tasks in progress and a list of failed ones
// It can optionally do something with each task by means of a monitoring function
func SkimTasksListMonitor(taskList []*Task, monitoringFunc TaskMonitoringFunc) ([]*Task, []*Task, error) {
	var newTaskList []*Task
	var errorList []*Task
	for _, task := range taskList {
		if task == nil {
			continue
		}
		err := task.Refresh()
		if err != nil {
			if strings.Contains(err.Error(), errorRetrievingTask) {
				// Task was not found. Probably expired. We don't need it anymore
				continue
			}
			return newTaskList, errorList, err
		}
		if monitoringFunc != nil {
			monitoringFunc(task.Task)
		}
		// if a cancellation was requested, we can ignore the task
		if task.Task.CancelRequested {
			continue
		}
		// If the task was completed successfully, or it was abandoned, we don't need further processing
		if isTaskComplete(task.Task.Status) {
			continue
		}
		// if the task failed, we add it to the special list
		if task.Task.Status == "error" && !task.Task.CancelRequested {
			errorList = append(errorList, task)
			continue
		}
		// If the task is running, we add it to the list that will continue to be monitored
		if isTaskRunning(task.Task.Status) {
			newTaskList = append(newTaskList, task)
		}
	}
	return newTaskList, errorList, nil
}

// isTaskRunning returns true if the task has started or is about to start
func isTaskRunning(status string) bool {
	return status == "running" || status == "preRunning" || status == "queued"
}

// isTaskComplete returns true if the task has finished successfully or was interrupted, but not if it finished with error
func isTaskComplete(status string) bool {
	return status == "success" || status == "aborted"
}

// isTaskCompleteOrError returns true if the status has finished, regardless of the outcome
func isTaskCompleteOrError(status string) bool {
	return isTaskComplete(status) || status == "error"
}

// WaitTaskListCompletion continuously skims the task list until no tasks in progress are left
func WaitTaskListCompletion(taskList []*Task) ([]*Task, error) {
	return WaitTaskListCompletionMonitor(taskList, nil)
}

// WaitTaskListCompletionMonitor continuously skims the task list until no tasks in progress are left
// Using a TaskMonitoringFunc, it can display or log information as the list reduction happens
func WaitTaskListCompletionMonitor(taskList []*Task, f TaskMonitoringFunc) ([]*Task, error) {
	var failedTaskList []*Task
	var err error
	for len(taskList) > 0 {
		taskList, failedTaskList, err = SkimTasksListMonitor(taskList, f)
		if err != nil {
			return failedTaskList, err
		}
		time.Sleep(3 * time.Second)
	}
	if len(failedTaskList) == 0 {
		return nil, nil
	}
	return failedTaskList, fmt.Errorf("%d tasks have failed", len(failedTaskList))
}

// GetTaskByHREF retrieves a task by its HREF
func (client *Client) GetTaskByHREF(taskHref string) (*Task, error) {
	task := NewTask(client)

	_, err := client.ExecuteRequest(taskHref, http.MethodGet,
		"", "error retrieving task: %s", nil, task.Task)
	if err != nil {
		return nil, fmt.Errorf("%s : %s", ErrorEntityNotFound, err)
	}

	return task, nil
}

// GetTaskById retrieves a task by ID
func (client *Client) GetTaskById(taskId string) (*Task, error) {
	// Builds the task HREF using the VCD HREF + /task/{ID} suffix
	taskHref, err := url.JoinPath(client.VCDHREF.String(), "task", extractUuid(taskId))
	if err != nil {
		return nil, err
	}
	return client.GetTaskByHREF(taskHref)
}

// SkimTasksList checks a list of task IDs and returns a list of IDs for tasks in progress and a list of IDs for failed ones
func (client Client) SkimTasksList(taskIdList []string) ([]string, []string, error) {
	var seenTasks = make(map[string]bool)
	var newTaskList []string
	var errorList []string
	for i, taskId := range taskIdList {
		_, seen := seenTasks[taskId]
		if seen {
			continue
		}
		seenTasks[taskId] = true
		task, err := client.GetTaskById(taskId)
		if err != nil {
			if strings.Contains(err.Error(), errorRetrievingTask) {
				// Task was not found. Probably expired. We don't need it anymore
				continue
			}
			return newTaskList, errorList, err
		}
		util.Logger.Printf("[SkimTasksList] {%d} task %s %s (status %s - cancel requested: %v)\n", i, task.Task.Name, task.Task.ID, task.Task.Status, task.Task.CancelRequested)
		if isTaskComplete(task.Task.Status) {
			continue
		}
		if isTaskRunning(task.Task.Status) {
			newTaskList = append(newTaskList, taskId)
		}
		if task.Task.Status == "error" && !task.Task.CancelRequested {
			errorList = append(errorList, taskId)
		}
	}
	return newTaskList, errorList, nil
}

// WaitTaskListCompletion waits until all tasks in the list are completed, removed, or failed
// Returns a list of failed tasks and an error
func (client Client) WaitTaskListCompletion(taskIdList []string, ignoreFailed bool) ([]string, error) {
	var failedTaskList []string
	var err error
	for len(taskIdList) > 0 {
		taskIdList, failedTaskList, err = client.SkimTasksList(taskIdList)
		if err != nil {
			return failedTaskList, err
		}
		time.Sleep(time.Second)
	}
	if len(failedTaskList) == 0 || ignoreFailed {
		return nil, nil
	}
	return failedTaskList, fmt.Errorf("%d tasks have failed", len(failedTaskList))
}

// QueryTaskList performs a query for tasks according to a specific filter
func (client *Client) QueryTaskList(filter map[string]string) ([]*types.QueryResultTaskRecordType, error) {
	taskType := types.QtTask
	if client.IsSysAdmin {
		taskType = types.QtAdminTask
	}

	filterText := buildFilterTextWithLogicalOr(filter)

	notEncodedParams := map[string]string{
		"type": taskType,
	}
	if filterText != "" {
		notEncodedParams["filter"] = filterText
	}
	results, err := client.cumulativeQuery(taskType, nil, notEncodedParams)
	if err != nil {
		return nil, fmt.Errorf("error querying task %s", err)
	}

	if client.IsSysAdmin {
		return results.Results.AdminTaskRecord, nil
	} else {
		return results.Results.TaskRecord, nil
	}
}

// buildFilterTextWithLogicalOr creates a filter with multiple values for a single column
// Given a map entry "key": "value1,value2"
// it creates a filter with a logical OR:  "key==value1,key==value2"
func buildFilterTextWithLogicalOr(filter map[string]string) string {
	filterText := ""
	for k, v := range filter {
		if filterText != "" {
			filterText += ";" // logical AND
		}
		if strings.Contains(v, ",") {
			valueText := ""
			values := strings.Split(v, ",")
			for _, value := range values {
				if valueText != "" {
					valueText += "," // logical OR
				}
				valueText += fmt.Sprintf("%s==%s", k, url.QueryEscape(value))
			}
			filterText += valueText
		} else {
			filterText += fmt.Sprintf("%s==%s", k, url.QueryEscape(v))
		}
	}
	return filterText
}

// WaitForRouteAdvertisementTasks is a convenience function to query for unfinished Route
// Advertisement tasks. An exact case for it was that updating some IP Space related objects (IP
// Spaces, IP Space Uplinks). Updating such an object sometimes results in a separate task for Route
// Advertisement being spun up (name="ipSpaceUplinkRouteAdvertisementSync"). When such task is
// running - other operations may fail so it is best to wait for completion of such task before
// triggering any other jobs.
func (client *Client) WaitForRouteAdvertisementTasks() error {
	name := "ipSpaceUplinkRouteAdvertisementSync"

	util.Logger.Printf("[TRACE] WaitForRouteAdvertisementTasks attempting to search for unfinished tasks with name='%s'", name)
	allTasks, err := client.QueryTaskList(map[string]string{
		"status": "running,preRunning,queued",
		"name":   name,
	})
	if err != nil {
		return fmt.Errorf("error retrieving all running '%s' tasks: %s", name, err)
	}

	util.Logger.Printf("[TRACE] WaitForRouteAdvertisementTasks got %d unifinished tasks with name='%s'", len(allTasks), name)
	for _, singleQueryTask := range allTasks {
		task := NewTask(client)
		task.Task.HREF = singleQueryTask.HREF

		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting for task '%s' of type '%s' to finish: %s", singleQueryTask.HREF, name, err)
		}
	}

	return nil
}
