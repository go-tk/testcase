package testcase

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

// TestCase represents a test case.
type TestCase struct {
	functionValue  reflect.Value
	functionType   reflect.Type
	contextPtrType reflect.Type
	callbackValues map[interface{}]reflect.Value
}

// New creates a test case with the given function.
// The type of the function should be should be func(*testing.T, *C), C is the type of
// test context, it could be arbitrary. When TestCase.Run() is called, a new C would be
// created and passed as arguments to TestCase.Run().
func New(function interface{}) *TestCase {
	var tc TestCase
	tc.functionValue = reflect.ValueOf(function)
	tc.functionType = tc.functionValue.Type()
	if !validateFunctionType(tc.functionType) {
		panic(fmt.Sprintf("the type of `function` should be func(*testing.T, *C)"))
	}
	tc.contextPtrType = tc.functionType.In(1)
	return &tc
}

func validateFunctionType(functionType reflect.Type) bool {
	if functionType.Kind() != reflect.Func {
		return false
	}
	if functionType.NumIn() != 2 {
		return false
	}
	if functionType.In(0) != reflect.TypeOf((*testing.T)(nil)) {
		return false
	}
	if functionType.In(1).Kind() != reflect.Ptr {
		return false
	}
	if functionType.NumOut() != 0 {
		return false
	}
	return true
}

// Copy deep copy the test case.
func (tc *TestCase) Copy() *TestCase {
	copy := *tc
	copy.callbackValues = copy.copyCallbackValues()
	return &copy
}

func (tc *TestCase) copyCallbackValues() map[interface{}]reflect.Value {
	if tc.callbackValues == nil {
		return nil
	}
	callbackValues := make(map[interface{}]reflect.Value, len(tc.callbackValues))
	for callbackID, callbackValue := range tc.callbackValues {
		callbackValues[callbackID] = callbackValue
	}
	return callbackValues
}

// SetCallback sets a callback by the given callback id.
// The callback could be involved later by the callback id via DoCallback().
func (tc *TestCase) SetCallback(callbackID interface{}, callback interface{}) *TestCase {
	callbackValue := reflect.ValueOf(callback)
	if callbackValue.Type() != tc.functionType {
		panic(fmt.Sprintf("the type of `callback` should be %s", tc.functionType))
	}
	callbackValues := tc.callbackValues
	if callbackValues == nil {
		callbackValues = make(map[interface{}]reflect.Value)
		tc.callbackValues = callbackValues
	}
	callbackValues[callbackID] = callbackValue
	return tc
}

var (
	testCasesLock sync.RWMutex
	testCases     = map[*testing.T]*TestCase{}
)

// Run runs the test case.
func (tc *TestCase) Run(t *testing.T) {
	name := makeName()
	t.Run(name, func(t *testing.T) {
		testCasesLock.Lock()
		testCases[t] = tc
		testCasesLock.Unlock()
		defer func() {
			testCasesLock.Lock()
			delete(testCases, t)
			testCasesLock.Unlock()
		}()
		argsValues := [...]reflect.Value{
			reflect.ValueOf(t),
			reflect.New(tc.contextPtrType.Elem()),
		}
		tc.functionValue.Call(argsValues[:])
	})
}

func makeName() string {
	_, fileName, lineNumber, _ := runtime.Caller(2)
	shortFileName := filepath.Base(fileName)
	return fmt.Sprintf("<%s:%d>", shortFileName, lineNumber)
}

// DoCallback involves a callback by the given callback ID which is previously set via
// TestCase.SetCallback().
func DoCallback(callbackID interface{}, t *testing.T, context interface{}) {
	testCasesLock.RLock()
	testCase, ok := testCases[t]
	testCasesLock.RUnlock()
	if !ok {
		panic("should only be called from test case functions")
	}
	testCase.doCallback(callbackID, t, context)
}

func (tc *TestCase) doCallback(callbackID interface{}, t *testing.T, context interface{}) {
	callbackValue, ok := tc.callbackValues[callbackID]
	if !ok {
		panic(fmt.Sprintf("can't find callback by id - %v", callbackID))
	}
	argsValues := [...]reflect.Value{
		reflect.ValueOf(t),
		reflect.ValueOf(context),
	}
	if argsValues[1].Type() != tc.contextPtrType {
		panic(fmt.Sprintf("the type of `context` should be %s", tc.contextPtrType))
	}
	callbackValue.Call(argsValues[:])
}
