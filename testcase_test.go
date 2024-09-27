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

	New(func(tc *TestCase[C]) {
		name = tc.T.Name()
	}).WithTag("my_tag").Run(t)
	assert.Equal(t, "TestTestCase_WithTag/testcase_test.go:17@my_tag", name)

	New(func(tc *TestCase[C]) {
		name = tc.T.Name()
	}).WithTag("").Run(t)
	assert.Equal(t, "TestTestCase_WithTag/testcase_test.go:22", name)
}

func TestTestCase_Callback(t *testing.T) {
	type C struct {
		N string
	}

	{
		var s string
		New(func(tc *TestCase[C]) {
			tc.C.N += "1"
			tc.Callback("cb_1")
			tc.C.N += "2"
			tc.Callback("cb_2")
			tc.C.N += "3"
			tc.Callback("cb_3")
		}).WithCallback("cb_1", func(tc *TestCase[C]) {
			s += "a"
			s += tc.C.N
		}).WithCallback("cb_2", func(tc *TestCase[C]) {
			s += "b"
			s += tc.C.N
		}).WithCallback("cb_3", func(tc *TestCase[C]) {
			s += "c"
			s += tc.C.N
		}).Run(t)
		assert.Equal(t, "a1b12c123", s)
	}

	New(func(tc *TestCase[C]) {
		tc.Assert.PanicsWithValue("can't find callback by id: foo", func() {
			tc.Callback("foo")
		})
	}).Run(t)
}

func TestTestCase_OptionalCallback(t *testing.T) {
	type C struct {
		N string
	}

	{
		var s string
		New(func(tc *TestCase[C]) {
			tc.C.N += "1"
			tc.OptionalCallback("cb_1")
			tc.C.N += "2"
			tc.OptionalCallback("cb_2")
			tc.C.N += "3"
			tc.OptionalCallback("cb_3")
		}).WithCallback("cb_1", func(tc *TestCase[C]) {
			s += "a"
			s += tc.C.N
		}).WithCallback("cb_3", func(tc *TestCase[C]) {
			s += "c"
			s += tc.C.N
		}).Run(t)
		assert.Equal(t, "a1c123", s)
	}
}

func TestTestCase_RunParallel(t *testing.T) {
	defer func(t0 time.Time) {
		assert.Less(t, time.Since(t0).Seconds(), 1.0)
	}(time.Now())

	type C struct {
		N string
	}

	for i := 0; i < 2; i++ {
		New(func(tc *TestCase[C]) {
			time.Sleep(time.Second * 4 / 5)
		}).RunParallel(t)
	}
}
