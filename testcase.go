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
	"unsafe"
)

// TestCase represents a test case.
type TestCase struct {
	tc testCase
}

// Copy copies the test case and returns a clone.
func (tc *TestCase) Copy() (clone *TestCase) { return tc.tc.Copy().TestCase() }

// Exclude excludes the test case from the list to run.
func (tc *TestCase) Exclude() (self *TestCase) { return tc.tc.Exclude().TestCase() }

// ExcludeOthers excludes other test cases from the list to run.
func (tc *TestCase) ExcludeOthers() (self *TestCase) { return tc.tc.ExcludeOthers().TestCase() }

// Given annotates the test case.
func (tc *TestCase) Given(given string) (self *TestCase) { return tc.tc.Given(given).TestCase() }

// When annotates the test case.
func (tc *TestCase) When(when string) (self *TestCase) { return tc.tc.When(when).TestCase() }

// Then annotates the test case.
func (tc *TestCase) Then(then string) (self *TestCase) { return tc.tc.Then(then).TestCase() }

// AddTask adds a task with the given ID to the test case.
// Tasks with lower IDs will be executed before ones with higher IDs.
func (tc *TestCase) AddTask(taskID int, task interface{}) (self *TestCase) {
	return tc.tc.AddTask(taskID, task).TestCase()
}

// New creates a new test case.
func New() *TestCase { return new(testCase).Init().TestCase() }

type testCase struct {
	ToExclude       bool
	ToExcludeOthers bool

	locator                 string
	given                   string
	when                    string
	then                    string
	taskType                reflect.Type
	workspaceType           reflect.Type
	workspaceBaseFieldIndex int
	tasks                   map[int]interface{}
}

func (tc *testCase) Init() *testCase {
	_, fileName, lineNumber, _ := runtime.Caller(2)
	tc.setLocator(fileName, lineNumber)
	return tc
}

func (tc *testCase) Copy() *testCase {
	clone := testCase{
		ToExclude:       tc.ToExclude,
		ToExcludeOthers: tc.ToExcludeOthers,

		given:                   tc.given,
		when:                    tc.when,
		then:                    tc.then,
		taskType:                tc.taskType,
		workspaceType:           tc.workspaceType,
		workspaceBaseFieldIndex: tc.workspaceBaseFieldIndex,
	}
	_, fileName, lineNumber, _ := runtime.Caller(2)
	clone.setLocator(fileName, lineNumber)
	clone.tasks = tc.copyTasks()
	return &clone
}

func (tc *testCase) setLocator(fileName string, lineNumber int) {
	shortFileName := filepath.Base(fileName)
	tc.locator = fmt.Sprintf("%s:%d", shortFileName, lineNumber)
}

func (tc *testCase) copyTasks() map[int]interface{} {
	tasksClone := make(map[int]interface{}, len(tc.tasks))
	for taskID, task := range tc.tasks {
		tasksClone[taskID] = task
	}
	return tasksClone
}

func (tc *testCase) Exclude() *testCase {
	tc.ToExclude = true
	return tc
}

func (tc *testCase) ExcludeOthers() *testCase {
	tc.ToExcludeOthers = true
	return tc
}

func (tc *testCase) Given(given string) *testCase {
	tc.given = given
	return tc
}

func (tc *testCase) When(when string) *testCase {
	tc.when = when
	return tc
}

func (tc *testCase) Then(then string) *testCase {
	tc.then = then
	return tc
}

func (tc *testCase) AddTask(taskID int, task interface{}) *testCase {
	tc.validateTaskType(task, taskID)
	tc.doAddTask(taskID, task)
	return tc
}

func (tc *testCase) validateTaskType(task interface{}, taskID int) {
	taskType := reflect.TypeOf(task)
	if tc.taskType != nil {
		if taskType != tc.taskType {
			panic(fmt.Sprintf("task type mismatch; taskID=%v taskType=%v expectedTaskType=%v",
				taskID, taskType, tc.taskType))
		}
		return
	}
	if taskType.Kind() != reflect.Func {
		panic(fmt.Sprintf("task should be function; taskID=%v taskType=%v", taskID, taskType))
	}
	if taskType.NumIn() != 1 {
		panic(fmt.Sprintf("task should have exactly one argument; taskID=%v taskType=%v",
			taskID, taskType))
	}
	workspaceTypePtr := taskType.In(0)
	if workspaceTypePtr.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("task argument #1 should be pointer; taskID=%v taskType=%v",
			taskID, taskType))
	}
	workspaceType := workspaceTypePtr.Elem()
	if workspaceType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("task argument #1 should point to structure; taskID=%v taskType=%v",
			taskID, taskType))
	}
	workspaceBaseType := reflect.TypeOf((*WorkspaceBase)(nil)).Elem()
	workspaceBaseFieldIndex := -1
	for i, n := 0, workspaceType.NumField(); i < n; i++ {
		f := workspaceType.Field(i)
		if f.Anonymous && f.Type == workspaceBaseType {
			workspaceBaseFieldIndex = i
			break
		}
	}
	if workspaceBaseFieldIndex < 0 {
		panic(fmt.Sprintf("structure `%v` should embed interface `%v`; taskID=%v taskType=%v",
			workspaceType, workspaceBaseType, taskID, taskType))
	}
	if taskType.NumOut() != 0 {
		panic(fmt.Sprintf("task should have no result; taskID=%v taskType=%v", taskID, taskType))
	}
	tc.taskType = taskType
	tc.workspaceType = workspaceType
	tc.workspaceBaseFieldIndex = workspaceBaseFieldIndex
}

func (tc *testCase) doAddTask(taskID int, task interface{}) {
	tasks := tc.tasks
	if tasks == nil {
		tasks = make(map[int]interface{}, 1)
		tc.tasks = tasks
	} else {
		if _, ok := tasks[taskID]; ok {
			panic(fmt.Sprintf("duplicate task id; taskID=%v", taskID))
		}
	}
	tasks[taskID] = task
}

func (tc *testCase) Run(t *testing.T, parallel bool) {
	if tc.tasks == nil {
		panic("no task")
	}
	t.Run(tc.locator, func(t *testing.T) {
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
	workspaceValuePtr, workspaceBase := tc.newWorkspace(t)
	defer workspaceBase.Clean()
	args := []reflect.Value{workspaceValuePtr}
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

func (tc *testCase) newWorkspace(t *testing.T) (reflect.Value, *workspaceBase) {
	workspaceValuePtr := reflect.New(tc.workspaceType)
	workspaceValue := workspaceValuePtr.Elem()
	workspaceBaseValue := workspaceValue.Field(tc.workspaceBaseFieldIndex)
	workspaceBase := &workspaceBaseValue.Addr().Interface().(*WorkspaceBase).wb
	workspaceBase.Init(t)
	return workspaceValuePtr, workspaceBase
}

func (tc *testCase) TestCase() *TestCase { return (*TestCase)(unsafe.Pointer(tc)) }

// WorkspaceBase should be embedded into concrete workspaces as their bases.
type WorkspaceBase struct {
	wb workspaceBase
}

// T returns the testing.T associated with the workspace.
func (wb *WorkspaceBase) T() (t *testing.T) { return wb.wb.T() }

// AddCleanup adds a cleanup to the workspace.
func (wb *WorkspaceBase) AddCleanup(cleanup func()) { wb.wb.AddCleanup(cleanup) }

type workspaceBase struct {
	t        *testing.T
	cleanups []func()
}

func (wb *workspaceBase) Init(t *testing.T) *workspaceBase {
	wb.t = t
	return wb
}

func (wb *workspaceBase) T() *testing.T             { return wb.t }
func (wb *workspaceBase) AddCleanup(cleanup func()) { wb.cleanups = append(wb.cleanups, cleanup) }

func (wb *workspaceBase) Clean() {
	for i := len(wb.cleanups) - 1; i >= 0; i-- {
		cleanup := wb.cleanups[i]
		cleanup()
	}
}

// RunList runs the given list of test cases.
func RunList(t *testing.T, list ...*TestCase) {
	doRunList(t, list, false)
}

// RunListParallel runs the given list of test cases parallel.
func RunListParallel(t *testing.T, list ...*TestCase) {
	doRunList(t, list, true)
}

func doRunList(t *testing.T, list []*TestCase, parallel bool) {
	for _, tc := range list {
		tc := &tc.tc
		if tc.ToExcludeOthers {
			tc.Run(t, false)
			return
		}
	}
	for _, tc := range list {
		tc := &tc.tc
		if !tc.ToExclude {
			tc.Run(t, parallel)
		}
	}
}
