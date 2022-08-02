package testcase_test

import (
	"testing"

	. "github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.NotPanics(t, func() {
		New(func(*testing.T, *struct{}) {})
	})

	for _, f := range []interface{}{
		1,
		func() {},
		func(int) {},
		func(byte, int, float64) {},
		func(struct{}, struct{}) {},
		func(*testing.T, struct{}) {},
		func(*testing.T, *struct{}) bool { return false },
	} {
		assert.PanicsWithValue(t, "the type of `function` should be func(*testing.T, *context)", func() {
			New(f)
		})
	}
}

func TestTestCase_SetCallback(t *testing.T) {
	tc := New(func(*testing.T, *struct{}) {}).SetCallback("abc", func(*testing.T, *struct{}) {})
	assert.Equal(t, `
Function Type: func(*testing.T, *struct {})
Callback IDs: abc
`[1:], tc.DumpAsString())

	tc.SetCallback(123, func(*testing.T, *struct{}) {})
	assert.Equal(t, `
Function Type: func(*testing.T, *struct {})
Callback IDs: 123, abc
`[1:], tc.DumpAsString())

	assert.PanicsWithValue(t, "the type of `callback` should be func(*testing.T, *struct {})", func() {
		tc.SetCallback("xyz", func(*testing.T, *string) {})
	})
}

func TestTestCase_Copy(t *testing.T) {
	tc := New(func(*testing.T, *struct{}) {}).SetCallback("abc", func(*testing.T, *struct{}) {})
	tc2 := tc.Copy().SetCallback(123, func(*testing.T, *struct{}) {})

	assert.Equal(t, `
Function Type: func(*testing.T, *struct {})
Callback IDs: abc
`[1:], tc.DumpAsString())
	assert.Equal(t, `
Function Type: func(*testing.T, *struct {})
Callback IDs: 123, abc
`[1:], tc2.DumpAsString())
}

func TestTestCase_Run(t *testing.T) {
	type context struct {
		N int
	}
	var f bool
	New(func(t *testing.T, c *context) {
		f = true
		assert.Equal(t, t.Name(), "TestTestCase_Run/<testcase_test.go:71>")
		assert.NotNil(t, c)
	}).Run(t)
	assert.True(t, f)
}

func TestDoCallback(t *testing.T) {
	var output []string
	tc :=
		New(func(t *testing.T, c *struct{}) {
			DoCallback(123, t, c)
			DoCallback("abc", t, c)

			assert.PanicsWithValue(t, "can't find callback by id - xyz", func() {
				DoCallback("xyz", t, c)
			})

			assert.PanicsWithValue(t, "the type of `context` should be *struct {}", func() {
				DoCallback(123, t, struct{}{})
			})
		}).
			SetCallback("abc", func(*testing.T, *struct{}) {
				output = append(output, "abc")
			}).
			SetCallback(123, func(*testing.T, *struct{}) {
				output = append(output, "123")
			})
	tc.Run(t)
	assert.Equal(t, []string{"123", "abc"}, output)

	assert.PanicsWithValue(t, "should only be called from test case functions", func() {
		DoCallback(123, t, &struct{}{})
	})
}
