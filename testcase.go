package testcase

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
)

// TestCase represents a test case.
type TestCase[C any, F func(t *testing.T, c *C)] struct {
	function     F
	tag          string
	callbackList []callbackItem
}

// New creates a test case.
func New[C any, F func(t *testing.T, c *C)](function F) TestCase[C, F] {
	return TestCase[C, F]{
		function: function,
	}
}

// WithTag attaches a tag to the test case.
// A new test case is returned.
func (tc TestCase[C, F]) WithTag(tag string) TestCase[C, F] {
	tc.tag = tag
	return tc
}

// WithCallback registers a callback for the given callback ID.
// A new test case is returned.
func (tc TestCase[C, F]) WithCallback(callbackID string, callback F) TestCase[C, F] {
	tc.callbackList = append(tc.callbackList, callbackItem{
		ID: callbackID,
		Value: func(t *testing.T, c interface{}) {
			callback(t, c.(*C))
		},
	})
	return tc
}

// Run executes the test case.
func (tc TestCase[C, F]) Run(t *testing.T) {
	name := tc.makeName()
	t.Run(name, func(t *testing.T) {
		tc.doRun(t)
	})
}

// RunParallel executes the test case in parallel mode.
func (tc TestCase[C, F]) RunParallel(t *testing.T) {
	name := tc.makeName()
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		tc.doRun(t)
	})
}

func (tc *TestCase[C, F]) makeName() string {
	_, fileName, lineNumber, _ := runtime.Caller(2)
	shortFileName := filepath.Base(fileName)
	if tc.tag == "" {
		return fmt.Sprintf("%s:%d", shortFileName, lineNumber)
	}
	return fmt.Sprintf("%s:%d@%s", shortFileName, lineNumber, tc.tag)
}

func (tc *TestCase[C, F]) doRun(t *testing.T) {
	c := new(C)
	context := newContext(c, tc.callbackList)
	setContext(t, context)
	t.Cleanup(func() { clearContext(t) })
	tc.function(t, c)
}

type callbackItem struct {
	ID    string
	Value callback
}

type callback func(t *testing.T, c interface{})

// Callback executes the callback with the given callback ID.
// If the callback does not exists, it panics.
func Callback(t *testing.T, callbackID string) {
	context, ok := getContext(t)
	if !ok {
		panic("can't find context")
	}
	callback, ok := context.GetCallback(callbackID)
	if !ok {
		panic(fmt.Sprintf("can't find callback by id: %s", callbackID))
	}
	callback(t, context.C())
}

// OptionalCallback executes the callback with the given callback ID.
// If the callback does not exists, nothing happens.
func OptionalCallback(t *testing.T, callbackID string) {
	context, ok := getContext(t)
	if !ok {
		panic("can't find context")
	}
	callback, ok := context.GetCallback(callbackID)
	if !ok {
		return
	}
	callback(t, context.C())
}
