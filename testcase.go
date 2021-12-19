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

// Step adds a step with the given number to the test case.
// Steps will be executed in ascending order of step number.
func (tc *TestCase) Step(stepNo float64, step interface{}) (self *TestCase) {
	return tc.tc.Step(stepNo, step).TestCase()
}

// New creates a new test case.
func New() *TestCase { return new(testCase).Init().TestCase() }

type testCase struct {
	IsExcluded        bool
	OthersAreExcluded bool

	locator       string
	given         string
	when          string
	then          string
	stepType      reflect.Type
	workspaceType reflect.Type
	steps         map[float64]interface{}
}

func (tc *testCase) Init() *testCase {
	_, fileName, lineNumber, _ := runtime.Caller(2)
	tc.setLocator(fileName, lineNumber)
	return tc
}

func (tc *testCase) Copy() *testCase {
	clone := testCase{
		IsExcluded:        tc.IsExcluded,
		OthersAreExcluded: tc.OthersAreExcluded,

		given:         tc.given,
		when:          tc.when,
		then:          tc.then,
		stepType:      tc.stepType,
		workspaceType: tc.workspaceType,
	}
	_, fileName, lineNumber, _ := runtime.Caller(2)
	clone.setLocator(fileName, lineNumber)
	clone.steps = tc.copySteps()
	return &clone
}

func (tc *testCase) setLocator(fileName string, lineNumber int) {
	shortFileName := filepath.Base(fileName)
	tc.locator = fmt.Sprintf("%s:%d", shortFileName, lineNumber)
}

func (tc *testCase) copySteps() map[float64]interface{} {
	stepsClone := make(map[float64]interface{}, len(tc.steps))
	for stepNo, step := range tc.steps {
		stepsClone[stepNo] = step
	}
	return stepsClone
}

func (tc *testCase) Exclude() *testCase {
	tc.IsExcluded = true
	return tc
}

func (tc *testCase) ExcludeOthers() *testCase {
	tc.OthersAreExcluded = true
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

func (tc *testCase) Step(stepNo float64, step interface{}) *testCase {
	tc.validateStepType(step, stepNo)
	tc.addStep(stepNo, step)
	return tc
}

func (tc *testCase) validateStepType(step interface{}, stepNo float64) {
	stepType := reflect.TypeOf(step)
	if tc.stepType != nil {
		if stepType != tc.stepType {
			panic(fmt.Sprintf("step type mismatch; stepNo=%v stepType=%q expectedStepType=%q",
				stepNo, stepType, tc.stepType))
		}
		return
	}
	if stepType.Kind() != reflect.Func {
		panic(fmt.Sprintf("step should be function; stepNo=%v stepType=%q", stepNo, stepType))
	}
	if stepType.NumIn() != 2 {
		panic(fmt.Sprintf("step should have exactly two parameters; stepNo=%v stepType=%q",
			stepNo, stepType))
	}
	if stepType.In(0) != reflect.TypeOf((*testing.T)(nil)) {
		panic(fmt.Sprintf("step parameter #1 should be pointer to testing.T; stepNo=%v stepType=%q",
			stepNo, stepType))
	}
	workspaceType, ok := func() (reflect.Type, bool) {
		workspaceTypePtr := stepType.In(1)
		if workspaceTypePtr.Kind() != reflect.Ptr {
			return nil, false
		}
		workspaceType := workspaceTypePtr.Elem()
		if workspaceType.Kind() != reflect.Struct {
			return nil, false
		}
		return workspaceType, true
	}()
	if !ok {
		panic(fmt.Sprintf("step parameter #2 should be pointer to structure; stepNo=%v stepType=%q",
			stepNo, stepType))
	}
	if stepType.NumOut() != 0 {
		panic(fmt.Sprintf("step should not return any results; stepNo=%v stepType=%q", stepNo, stepType))
	}
	tc.stepType = stepType
	tc.workspaceType = workspaceType
}

func (tc *testCase) addStep(stepNo float64, step interface{}) {
	steps := tc.steps
	if steps == nil {
		steps = make(map[float64]interface{}, 1)
		tc.steps = steps
	} else {
		if _, ok := steps[stepNo]; ok {
			panic(fmt.Sprintf("duplicate step number; stepNo=%v", stepNo))
		}
	}
	steps[stepNo] = step
}

func (tc *testCase) Run(t *testing.T, parallel bool) {
	if tc.steps == nil {
		panic("no step")
	}
	t.Run(tc.locator, func(t *testing.T) {
		if parallel {
			t.Parallel()
		}
		tc.logGWT(t)
		tc.executeSteps(t)
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

func (tc *testCase) executeSteps(t *testing.T) {
	steps := tc.sortSteps()
	tValuePtr := reflect.ValueOf(t)
	workspaceValuePtr := reflect.New(tc.workspaceType)
	args := []reflect.Value{tValuePtr, workspaceValuePtr}
	for _, step := range steps {
		reflect.ValueOf(step).Call(args)
	}
}

func (tc *testCase) sortSteps() []interface{} {
	stepNos := make([]float64, len(tc.steps))
	i := 0
	for stepNo := range tc.steps {
		stepNos[i] = stepNo
		i++
	}
	sort.Float64s(stepNos)
	steps := make([]interface{}, len(tc.steps))
	i = 0
	for _, stepNo := range stepNos {
		steps[i] = tc.steps[stepNo]
		i++
	}
	return steps
}

func (tc *testCase) TestCase() *TestCase { return (*TestCase)(unsafe.Pointer(tc)) }

// RunList runs the given list of test cases, steps in each test case will be
// executed in order.
// Test cases will be run with isolated workspaces and each step in a test case
// shares the same workspace.
func RunList(t *testing.T, list ...*TestCase) {
	doRunList(t, list, false)
}

// RunListParallel runs the given list of test cases parallel, steps in each test case
// will be executed in order.
// Test cases will be run with isolated workspaces and each step in a test case
// shares the same workspace.
func RunListParallel(t *testing.T, list ...*TestCase) {
	doRunList(t, list, true)
}

func doRunList(t *testing.T, list []*TestCase, parallel bool) {
	for _, tc := range list {
		tc := &tc.tc
		if tc.OthersAreExcluded {
			tc.Run(t, false)
			return
		}
	}
	for _, tc := range list {
		tc := &tc.tc
		if !tc.IsExcluded {
			tc.Run(t, parallel)
		}
	}
}
