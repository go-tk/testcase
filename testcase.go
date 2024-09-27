package testcase

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCase represents a test case.
type TestCase[C any] struct {
	T      *testing.T
	C      C
	Assert *assert.Assertions

	function     func(*TestCase[C])
	tag          string
	callbackList []callbackItem[C]
	callbacks    map[string]func(*TestCase[C])
}

type callbackItem[C any] struct {
	ID    string
	Value func(*TestCase[C])
}

// New creates a test case.
func New[C any](function func(*TestCase[C])) TestCase[C] {
	return TestCase[C]{
		function: function,
	}
}

// WithTag attaches a tag to the test case.
// A new test case is returned.
func (tc TestCase[C]) WithTag(tag string) TestCase[C] {
	tc.tag = tag
	return tc
}

// WithCallback registers a callback for the given callback ID.
// A new test case is returned.
func (tc TestCase[C]) WithCallback(callbackID string, callback func(*TestCase[C])) TestCase[C] {
	tc.callbackList = append(tc.callbackList, callbackItem[C]{
		ID:    callbackID,
		Value: callback,
	})
	return tc
}

// Run executes the test case.
func (tc TestCase[C]) Run(t *testing.T) {
	name := tc.makeName()
	t.Run(name, func(t *testing.T) {
		tc.doRun(t)
	})
}

// RunParallel executes the test case in parallel mode.
func (tc TestCase[C]) RunParallel(t *testing.T) {
	name := tc.makeName()
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		tc.doRun(t)
	})
}

func (tc *TestCase[C]) makeName() string {
	_, fileName, lineNumber, _ := runtime.Caller(2)
	shortFileName := filepath.Base(fileName)
	if tc.tag == "" {
		return fmt.Sprintf("%s:%d", shortFileName, lineNumber)
	}
	return fmt.Sprintf("%s:%d@%s", shortFileName, lineNumber, tc.tag)
}

func (tc *TestCase[C]) doRun(t *testing.T) {
	tc.T = t
	tc.Assert = assert.New(t)
	tc.populateCallbacks()
	tc.function(tc)
}

func (tc *TestCase[C]) populateCallbacks() {
	callbacks := make(map[string]func(*TestCase[C]), len(tc.callbackList))
	for _, callbackItem := range tc.callbackList {
		callbacks[callbackItem.ID] = callbackItem.Value
	}
	tc.callbacks = callbacks
}

// Callback executes the callback with the given callback ID.
// If the callback does not exists, it panics.
func (tc *TestCase[C]) Callback(callbackID string) {
	callback, ok := tc.callbacks[callbackID]
	if !ok {
		panic(fmt.Sprintf("can't find callback by id: %s", callbackID))
	}
	callback(tc)
}

// OptionalCallback executes the callback with the given callback ID.
// If the callback does not exists, nothing happens.
func (tc *TestCase[C]) OptionalCallback(callbackID string) {
	callback, ok := tc.callbacks[callbackID]
	if !ok {
		return
	}
	callback(tc)
}
