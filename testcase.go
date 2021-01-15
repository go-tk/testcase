// Package testcase provides facilities for table-driven tests.
package testcase

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// TestCase represents a test case.
type TestCase interface {
	Copy() (copy TestCase)

	// Set information.
	Given(v string) (self TestCase)
	When(v string) (self TestCase)
	Then(v string) (self TestCase)

	// Set procedures.
	PreSetup(v interface{}) (self TestCase)
	Setup(v interface{}) (self TestCase)
	PreRun(v interface{}) (self TestCase)
	Run(v interface{}) (self TestCase)
	PostRun(v interface{}) (self TestCase)
	Teardown(v interface{}) (self TestCase)
	PostTeardown(v interface{}) (self TestCase)
}

// New creates a test case with the given context factory.
func New(contextFactory interface{}) TestCase {
	return new(testCase).Init(contextFactory)
}

// RunList runs a list of test cases.
func RunList(t *testing.T, list []TestCase) {
	for _, tc := range list {
		tc.(*testCase).Execute(t, false)
	}
}

// RunListParallel runs a list of test cases parallel.
func RunListParallel(t *testing.T, list []TestCase) {
	for _, tc := range list {
		tc.(*testCase).Execute(t, true)
	}
}

type testCase struct {
	contextFactory interface{}
	contextType    reflect.Type
	name           string

	given string
	when  string
	then  string

	preSetup     interface{}
	setup        interface{}
	preRun       interface{}
	run          interface{}
	postRun      interface{}
	tearDown     interface{}
	postTeardown interface{}
}

func (tc *testCase) Init(contextFactory interface{}) *testCase {
	contextFactoryType := reflect.TypeOf(contextFactory)
	validateContextFactoryType(contextFactoryType)
	tc.contextFactory = contextFactory
	tc.contextType = contextFactoryType.Out(0)
	_, fileName, lineNumber, _ := runtime.Caller(2)
	tc.setName(fileName, lineNumber)
	return tc
}

func (tc *testCase) Execute(t *testing.T, parallel bool) {
	if tc.run == nil {
		panic("procedure for Run unset")
	}
	t.Run(tc.name, func(t *testing.T) {
		if parallel {
			t.Parallel()
		}
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
		if buffer.Len() >= 1 {
			t.Log(buffer.String())
		}
		tValue := reflect.ValueOf(t)
		results := reflect.ValueOf(tc.contextFactory).Call([]reflect.Value{tValue})
		contextValue := results[0]
		args := []reflect.Value{tValue, contextValue}
		if tc.preSetup != nil {
			reflect.ValueOf(tc.preSetup).Call(args)
		}
		if tc.setup != nil {
			reflect.ValueOf(tc.setup).Call(args)
		}
		if tc.preRun != nil {
			reflect.ValueOf(tc.preRun).Call(args)
		}
		reflect.ValueOf(tc.run).Call(args)
		if tc.postRun != nil {
			reflect.ValueOf(tc.postRun).Call(args)
		}
		if tc.tearDown != nil {
			reflect.ValueOf(tc.tearDown).Call(args)
		}
		if tc.postTeardown != nil {
			reflect.ValueOf(tc.postTeardown).Call(args)
		}
	})
}

func (tc *testCase) Copy() TestCase {
	copy := *tc
	_, fileName, lineNumber, _ := runtime.Caller(1)
	copy.setName(fileName, lineNumber)
	return &copy
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

func (tc *testCase) PreSetup(v interface{}) TestCase {
	tc.setProcedure("PreSetup", &tc.preSetup, v)
	return tc
}

func (tc *testCase) Setup(v interface{}) TestCase {
	tc.setProcedure("Setup", &tc.setup, v)
	return tc
}

func (tc *testCase) PreRun(v interface{}) TestCase {
	tc.setProcedure("PreRun", &tc.preRun, v)
	return tc
}

func (tc *testCase) Run(v interface{}) TestCase {
	tc.setProcedure("Run", &tc.run, v)
	return tc
}

func (tc *testCase) PostRun(v interface{}) TestCase {
	tc.setProcedure("PostRun", &tc.postRun, v)
	return tc
}

func (tc *testCase) Teardown(v interface{}) TestCase {
	tc.setProcedure("Teardown", &tc.tearDown, v)
	return tc
}

func (tc *testCase) PostTeardown(v interface{}) TestCase {
	tc.setProcedure("PostTeardown", &tc.postTeardown, v)
	return tc
}

func (tc *testCase) setName(fileName string, lineNumber int) {
	shortFileName := filepath.Base(fileName)
	tc.name = fmt.Sprintf("%s:%d", shortFileName, lineNumber)
}

func (tc *testCase) setProcedure(procedureName string, oldProcedure *interface{}, newProcedure interface{}) {
	procedureType := reflect.TypeOf(newProcedure)
	tc.validateProcedureType(procedureType, procedureName)
	*oldProcedure = newProcedure
}

func (tc *testCase) validateProcedureType(procedureType reflect.Type, procedureName string) {
	if !(procedureType.Kind() == reflect.Func &&
		procedureType.NumIn() == 2 &&
		procedureType.In(0) == reflect.TypeOf((*testing.T)(nil)) &&
		procedureType.In(1) == tc.contextType) {
		panic(fmt.Sprintf("invalid procedure type for %v, type `func(*testing.T, %v)` expected; procedureType=%v",
			procedureName, tc.contextType, procedureType))
	}
}

func validateContextFactoryType(contextFactoryType reflect.Type) {
	if !(contextFactoryType.Kind() == reflect.Func &&
		contextFactoryType.NumIn() == 1 &&
		contextFactoryType.In(0) == reflect.TypeOf((*testing.T)(nil)) &&
		contextFactoryType.NumOut() == 1) {
		panic(fmt.Sprintf("invalid context factory type, type `func(*testing.T) TYPE` expected; contextFactoryType=%v", contextFactoryType))
	}
}
