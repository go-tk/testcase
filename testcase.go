// Package testcase provides facilities for table-driven tests.
package testcase

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"
)

// TestCase represents a test case.
type TestCase interface {
	// Copy copies this test case and returns a clone.
	Copy() (copy TestCase)

	// Exclude excludes this test case from the list to run.
	Exclude() (self TestCase)
	// ExcludeOthers excludes other test cases from the list to run.
	ExcludeOthers() (self TestCase)

	// Set information for this test case.
	Given(v string) (self TestCase)
	When(v string) (self TestCase)
	Then(v string) (self TestCase)

	// Task adds a new task with the given ID to this test case.
	// The lower the ID of a task, the sooner the task would be executed.
	Task(taskID int, task interface{}) (self TestCase)
}

// New creates a test case with the given workspace factory.
func New(workspaceFactory interface{}) TestCase {
	return new(testCase).Init(workspaceFactory)
}

// RunList runs a list of test cases.
func RunList(t *testing.T, list ...TestCase) {
	doRunList(t, list, false)
}

// RunListParallel runs a list of test cases parallel.
func RunListParallel(t *testing.T, list ...TestCase) {
	doRunList(t, list, true)
}

func doRunList(t *testing.T, list []TestCase, parallel bool) {
	for _, tc := range list {
		if tc := tc.(*testCase); tc.ToExcludeOthers {
			tc.Run(t, false)
			return
		}
	}
	for _, tc := range list {
		if tc := tc.(*testCase); !tc.ToExclude {
			tc.Run(t, parallel)
		}
	}
}

type testCase struct {
	name             string
	workspaceFactory interface{}
	workspaceType    reflect.Type

	ToExclude       bool
	ToExcludeOthers bool

	given string
	when  string
	then  string
	tasks map[int]interface{}
}

func (tc *testCase) Init(workspaceFactory interface{}) *testCase {
	_, fileName, lineNumber, _ := runtime.Caller(2)
	tc.setName(fileName, lineNumber)
	workspaceFactoryType := reflect.TypeOf(workspaceFactory)
	validateWorkspaceFactoryType(workspaceFactoryType)
	tc.workspaceFactory = workspaceFactory
	tc.workspaceType = workspaceFactoryType.Out(0)
	return tc
}

func validateWorkspaceFactoryType(workspaceFactoryType reflect.Type) {
	if !(workspaceFactoryType.Kind() == reflect.Func &&
		workspaceFactoryType.NumIn() == 1 &&
		workspaceFactoryType.In(0) == reflect.TypeOf((*testing.T)(nil)) &&
		workspaceFactoryType.NumOut() == 1) {
		panic(fmt.Sprintf("invalid workspace factory type, type `func(*testing.T) TYPE` expected; workspaceFactoryType=%v", workspaceFactoryType))
	}
}

func (tc *testCase) Run(t *testing.T, parallel bool) {
	if tc.tasks == nil {
		panic("no task")
	}
	t.Run(tc.name, func(t *testing.T) {
		if parallel {
			t.Parallel()
		}
		tc.logGWT(t)
		tc.executeTasks(t)
	})
}

func (tc *testCase) logGWT(t *testing.T) {
	var buffer bytes.Buffer
	if tc.given != "" {
		buffer.WriteString("\nGIVEN " + tc.given)
	}
	if tc.when != "" {
		buffer.WriteString("\nWHEN " + tc.when)
	}
	if tc.then != "" {
		buffer.WriteString("\nTHEN " + tc.then)
	}
	if buffer.Len() == 0 {
		return
	}
	t.Log(buffer.String())
}

func (tc *testCase) executeTasks(t *testing.T) {
	tasks := tc.sortTasks()
	tValue := reflect.ValueOf(t)
	workspaceValue := tc.newWorkspace(tValue)
	args := []reflect.Value{tValue, workspaceValue}
	for _, task := range tasks {
		reflect.ValueOf(task).Call(args)
	}
}

func (tc *testCase) sortTasks() []interface{} {
	taskIDs := make([]int, len(tc.tasks))
	i := 0
	for taskID := range tc.tasks {
		taskIDs[i] = taskID
		i++
	}
	sort.Ints(taskIDs)
	tasks := make([]interface{}, len(tc.tasks))
	i = 0
	for _, taskID := range taskIDs {
		tasks[i] = tc.tasks[taskID]
		i++
	}
	return tasks
}

func (tc *testCase) newWorkspace(tValue reflect.Value) reflect.Value {
	results := reflect.ValueOf(tc.workspaceFactory).Call([]reflect.Value{tValue})
	workspaceValue := results[0]
	return workspaceValue
}

func (tc *testCase) Copy() TestCase {
	clone := *tc
	_, fileName, lineNumber, _ := runtime.Caller(1)
	clone.setName(fileName, lineNumber)
	clone.tasks = tc.copyTasks()
	return &clone
}

func (tc *testCase) copyTasks() map[int]interface{} {
	tasksClone := make(map[int]interface{}, len(tc.tasks))
	for taskID, task := range tc.tasks {
		tasksClone[taskID] = task
	}
	return tasksClone
}

func (tc *testCase) Exclude() TestCase {
	tc.ToExclude = true
	return tc
}

func (tc *testCase) ExcludeOthers() TestCase {
	tc.ToExcludeOthers = true
	return tc
}

func (tc *testCase) Given(v string) TestCase {
	tc.given = v
	return tc
}

func (tc *testCase) When(v string) TestCase {
	tc.when = v
	return tc
}

func (tc *testCase) Then(v string) TestCase {
	tc.then = v
	return tc
}

func (tc *testCase) Task(taskID int, task interface{}) TestCase {
	tc.checkTaskID(taskID)
	taskType := reflect.TypeOf(task)
	tc.validateTaskType(taskType, taskID)
	tc.addTest(taskID, task)
	return tc
}

func (tc *testCase) checkTaskID(taskID int) {
	if _, ok := tc.tasks[taskID]; ok {
		panic(fmt.Sprintf("task already exists; taskID=%v", taskID))
	}
}

func (tc *testCase) validateTaskType(taskType reflect.Type, taskID int) {
	if !(taskType.Kind() == reflect.Func &&
		taskType.NumIn() == 2 &&
		taskType.In(0) == reflect.TypeOf((*testing.T)(nil)) &&
		taskType.In(1) == tc.workspaceType) {
		panic(fmt.Sprintf("invalid task type, type `func(*testing.T, %v)` expected; taskID=%v taskType=%v",
			tc.workspaceType, taskID, taskType))
	}
}

func (tc *testCase) addTest(taskID int, task interface{}) {
	tasks := tc.tasks
	if tasks == nil {
		tasks = make(map[int]interface{}, 1)
		tc.tasks = tasks
	}
	tasks[taskID] = task
}

func (tc *testCase) setName(fileName string, lineNumber int) {
	shortFileName := filepath.Base(fileName)
	tc.name = fmt.Sprintf("%s:%d", shortFileName, lineNumber)
}
