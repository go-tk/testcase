package testcase_test

import (
	"testing"
	"time"

	. "github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestTestCase_Step(t *testing.T) {
	assert.PanicsWithValue(t, "step should be function; stepNo=1 stepType=\"string\"", func() {
		New().Step(1, "foo")
	})
	assert.PanicsWithValue(t, "step should have exactly two parameters; stepNo=1 stepType=\"func()\"", func() {
		New().Step(1, func() {})
	})
	assert.PanicsWithValue(t, "step should have exactly two parameters; stepNo=1 stepType=\"func(bool, int, string)\"", func() {
		New().Step(1, func(bool, int, string) {})
	})
	assert.PanicsWithValue(t, "step parameter #1 should be pointer to testing.T; stepNo=1 stepType=\"func(int, *struct {})\"", func() {
		New().Step(1, func(int, *struct{}) {})
	})
	assert.PanicsWithValue(t, "step parameter #2 should be pointer to structure; stepNo=1 stepType=\"func(*testing.T, int)\"", func() {
		New().Step(1, func(*testing.T, int) {})
	})
	assert.PanicsWithValue(t, "step parameter #2 should be pointer to structure; stepNo=1 stepType=\"func(*testing.T, *int)\"", func() {
		New().Step(1, func(*testing.T, *int) {})
	})
	assert.PanicsWithValue(t, "step should not return any results; stepNo=1 stepType=\"func(*testing.T, *struct {}) int\"", func() {
		New().Step(1, func(*testing.T, *struct{}) int { return 0 })
	})
	assert.PanicsWithValue(t, "step type mismatch; stepNo=1 stepType=\"func(*testing.T, *testcase_test.Workspace2)\" expectedStepType=\"func(*testing.T, *testcase_test.Workspace1)\"", func() {
		type Workspace1 struct{}
		type Workspace2 struct{}
		New().
			Step(1, func(*testing.T, *Workspace1) {}).
			Step(1, func(*testing.T, *Workspace2) {})
	})
	assert.PanicsWithValue(t, "duplicate step number; stepNo=1", func() {
		New().
			Step(1, func(*testing.T, *struct{}) {}).
			Step(1, func(*testing.T, *struct{}) {})
	})
	New().
		Step(1, func(*testing.T, *struct{}) {}).
		Step(2, func(*testing.T, *struct{}) {})
}

func TestTestCase_Run(t *testing.T) {
	assert.PanicsWithValue(t, "no step", func() {
		RunList(t, New())
	})
	type Workspace struct {
		N int
	}
	var s string
	tc := New().
		Step(1, func(t *testing.T, w *Workspace) {
			assert.Equal(t, 1, w.N)
			w.N += 1
			s += "2"
			t.Cleanup(func() { s += "5" })
		}).
		Step(0.5, func(t *testing.T, w *Workspace) {
			assert.Equal(t, 0, w.N)
			w.N += 1
			assert.Regexp(t, "^TestTestCase_Run/testcase_test\\.go:", t.Name())
			s += "1"
			t.Cleanup(func() { s += "6" })
		}).
		Step(1.5, func(t *testing.T, w *Workspace) {
			assert.Equal(t, 2, w.N)
			w.N += 1
			s += "3"
			t.Cleanup(func() { s += "4" })
		})
	RunList(t, tc, tc)
	assert.Equal(t, "123456123456", s)
}

func TestTestCase_Clone(t *testing.T) {
	var s string
	tc1 := New().
		Step(1, func(t *testing.T, _ *struct{}) {
			s += "2"
			t.Cleanup(func() { s += "5" })
		}).
		Step(2, func(t *testing.T, _ *struct{}) {
			s += "3"
			t.Cleanup(func() { s += "4" })
		})
	tc2 := tc1.Copy().Step(0.5, func(t *testing.T, _ *struct{}) {
		s += "1"
		t.Cleanup(func() { s += "6" })
	})
	tc3 := tc1.Copy().Step(0.5, func(t *testing.T, _ *struct{}) {
		s += "A"
		t.Cleanup(func() { s += "B" })
	})
	RunList(t, tc1)
	assert.Equal(t, "2345", s)
	s = ""
	RunList(t, tc2)
	assert.Equal(t, "123456", s)
	s = ""
	RunList(t, tc3)
	assert.Equal(t, "A2345B", s)
}

func TestTestCase_Exclude(t *testing.T) {
	var s string
	RunList(t,
		New().
			Step(100, func(*testing.T, *struct{}) {
				s += "1"
			}),
		New().Exclude().
			Step(100, func(*testing.T, *struct{}) {
				s += "2"
			}),
		New().
			Step(100, func(*testing.T, *struct{}) {
				s += "3"
			}),
	)
	assert.Equal(t, "13", s)
}

func TestTestCase_ExcludeOthers(t *testing.T) {
	var s string
	RunList(t,
		New().
			Step(100, func(*testing.T, *struct{}) {
				s += "1"
			}),
		New().ExcludeOthers().
			Step(100, func(*testing.T, *struct{}) {
				s += "2"
			}),
		New().
			Step(100, func(*testing.T, *struct{}) {
				s += "3"
			}),
	)
	assert.Equal(t, "2", s)
}

func TestRunListParallel(t *testing.T) {
	t0 := time.Now()
	RunListParallel(t,
		New().
			Step(100, func(*testing.T, *struct{}) {
				time.Sleep(time.Second)
			}),
		New().
			Step(100, func(*testing.T, *struct{}) {
				time.Sleep(time.Second)
			}),
		New().
			Step(100, func(*testing.T, *struct{}) {
				time.Sleep(time.Second)
			}),
	)
	d := time.Since(t0)
	assert.Less(t, d, 3*time.Second)
}
