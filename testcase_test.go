package testcase_test

import (
	"testing"
	"time"

	. "github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestTestCase_WithTag(t *testing.T) {
	type C struct{}
	var name string

	New(func(t *testing.T, c *C) {
		name = t.Name()
	}).WithTag("my_tag").Run(t)
	assert.Equal(t, "TestTestCase_WithTag/testcase_test.go:17@my_tag", name)

	New(func(t *testing.T, c *C) {
		name = t.Name()
	}).WithTag("").Run(t)
	assert.Equal(t, "TestTestCase_WithTag/testcase_test.go:22", name)
}

func TestTestCase_Callback(t *testing.T) {
	{
		type C struct {
			n string
		}
		var s string
		New(func(t *testing.T, c *C) {
			c.n += "1"
			Callback(t, "cb_1")
			c.n += "2"
			Callback(t, "cb_2")
			c.n += "3"
			Callback(t, "cb_3")
		}).WithCallback("cb_1", func(t *testing.T, c *C) {
			s += "a"
			s += c.n
		}).WithCallback("cb_2", func(t *testing.T, c *C) {
			s += "b"
			s += c.n
		}).WithCallback("cb_3", func(t *testing.T, c *C) {
			s += "c"
			s += c.n
		}).Run(t)
		assert.Equal(t, "a1b12c123", s)
	}

	assert.PanicsWithValue(t, "can't find context", func() {
		Callback(t, "foo")
	})
	New(func(t *testing.T, c *struct{}) {
		assert.PanicsWithValue(t, "can't find callback by id: foo", func() {
			Callback(t, "foo")
		})
	}).Run(t)
}

func TestTestCase_OptionalCallback(t *testing.T) {
	{
		type C struct {
			n string
		}
		var s string
		New(func(t *testing.T, c *C) {
			c.n += "1"
			OptionalCallback(t, "cb_1")
			c.n += "2"
			OptionalCallback(t, "cb_2")
			c.n += "3"
			OptionalCallback(t, "cb_3")
		}).WithCallback("cb_1", func(t *testing.T, c *C) {
			s += "a"
			s += c.n
		}).WithCallback("cb_3", func(t *testing.T, c *C) {
			s += "c"
			s += c.n
		}).Run(t)
		assert.Equal(t, "a1c123", s)
	}

	assert.PanicsWithValue(t, "can't find context", func() {
		OptionalCallback(t, "foo")
	})
}

func TestTestCase_RunParallel(t *testing.T) {
	defer func(t0 time.Time) {
		assert.Less(t, time.Since(t0).Seconds(), 1.0)
	}(time.Now())

	for i := 0; i < 2; i++ {
		New(func(t *testing.T, c *struct{}) {
			time.Sleep(time.Second * 4 / 5)
		}).RunParallel(t)
	}
}
