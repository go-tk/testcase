package testcase

import (
	"sync"
	"testing"
)

type context struct {
	c         interface{}
	callbacks map[string]callback
}

func newContext(c interface{}, callbackList []callbackItem) *context {
	callbacks := make(map[string]callback)
	for _, callbackItem := range callbackList {
		callbacks[callbackItem.ID] = callbackItem.Value
	}
	return &context{
		c:         c,
		callbacks: callbacks,
	}
}

func (c *context) C() interface{} { return c.c }

func (c *context) GetCallback(callbackID string) (callback, bool) {
	callback, ok := c.callbacks[callbackID]
	return callback, ok
}

var (
	contextsLock sync.Mutex
	contexts     map[*testing.T]*context = make(map[*testing.T]*context)
)

func setContext(t *testing.T, context *context) {
	contextsLock.Lock()
	defer contextsLock.Unlock()
	contexts[t] = context
}

func clearContext(t *testing.T) {
	contextsLock.Lock()
	defer contextsLock.Unlock()
	delete(contexts, t)
}

func getContext(t *testing.T) (*context, bool) {
	contextsLock.Lock()
	defer contextsLock.Unlock()
	context, ok := contexts[t]
	return context, ok
}
