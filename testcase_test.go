package testcase_test

import (
	"testing"
	"time"

	. "github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestTestCase_AddTask(t *testing.T) {
	assert.PanicsWithValue(t, "task should be function; taskID=100 taskType=string", func() {
		New().AddTask(100, "foo")
	})
	assert.PanicsWithValue(t, "task should have exactly one argument; taskID=100 taskType=func()", func() {
		New().AddTask(100, func() {})
	})
	assert.PanicsWithValue(t, "task should have exactly one argument; taskID=100 taskType=func(int, string)", func() {
		New().AddTask(100, func(int, string) {})
	})
	assert.PanicsWithValue(t, "task argument #1 should be pointer; taskID=100 taskType=func(int)", func() {
		New().AddTask(100, func(int) {})
	})
	assert.PanicsWithValue(t, "task argument #1 should point to structure; taskID=100 taskType=func(*int)", func() {
		New().AddTask(100, func(*int) {})
	})
	assert.PanicsWithValue(t, "structure `struct {}` should embed structure `testcase.WorkspaceBase`; taskID=100 taskType=func(*struct {})", func() {
		New().AddTask(100, func(*struct{}) {})
	})
	New().AddTask(100, func(*struct {
		WorkspaceBase
	}) {
	})
	New().AddTask(100, func(*struct {
		Foo int
		WorkspaceBase
		Bar string
	}) {
	})
	assert.PanicsWithValue(t, "task should have no result; taskID=100 taskType=func(*struct { testcase.WorkspaceBase }) int", func() {
		New().AddTask(100, func(*struct {
			WorkspaceBase
		}) int {
			return 0
		})
	})
	assert.PanicsWithValue(t, "task type mismatch; taskID=100 taskType=func(*testcase_test.Workspace2) expectedTaskType=func(*testcase_test.Workspace1)", func() {
		type Workspace1 struct {
			WorkspaceBase
		}
		type Workspace2 struct {
			WorkspaceBase
		}
		New().AddTask(100, func(*Workspace1) {}).AddTask(100, func(*Workspace2) {})
	})
	assert.PanicsWithValue(t, "duplicate task id; taskID=100", func() {
		New().
			AddTask(100, func(*struct {
				WorkspaceBase
			}) {
			}).
			AddTask(100, func(*struct {
				WorkspaceBase
			}) {
			})
	})
	type Workspace0 struct {
		WorkspaceBase
	}
	New().
		AddTask(100, func(*struct {
			WorkspaceBase
		}) {
		}).
		AddTask(101, func(*struct {
			WorkspaceBase
		}) {
		})
}

func TestTestCase_Run(t *testing.T) {
	assert.PanicsWithValue(t, "no task", func() {
		RunList(t, New())
	})
	var s string
	RunList(t, New().
		AddTask(1000, func(w *struct{ WorkspaceBase }) {
			s += "2"
			w.AddCleanup(func() { s += "5" })
		}).
		AddTask(999, func(w *struct{ WorkspaceBase }) {
			assert.Regexp(w.T(), "^TestTestCase_Run/testcase_test\\.go:", w.T().Name())
			s += "1"
			w.AddCleanup(func() { s += "6" })
		}).
		AddTask(1001, func(w *struct{ WorkspaceBase }) {
			s += "3"
			w.AddCleanup(func() { s += "4" })
		}))
	assert.Equal(t, "123456", s)
}

func TestTestCase_Clone(t *testing.T) {
	var s string
	tc1 := New().
		AddTask(1000, func(w *struct{ WorkspaceBase }) {
			s += "2"
			w.AddCleanup(func() { s += "5" })
		}).
		AddTask(1001, func(w *struct{ WorkspaceBase }) {
			s += "3"
			w.AddCleanup(func() { s += "4" })
		})
	tc2 := tc1.Copy().AddTask(999, func(w *struct{ WorkspaceBase }) {
		s += "1"
		w.AddCleanup(func() { s += "6" })
	})
	tc3 := tc1.Copy().AddTask(999, func(w *struct{ WorkspaceBase }) {
		s += "A"
		w.AddCleanup(func() { s += "B" })
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
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				s += "1"
			}),
		New().Exclude().
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				s += "2"
			}),
		New().
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				s += "3"
			}),
	)
	assert.Equal(t, "13", s)
}

func TestTestCase_ExcludeOthers(t *testing.T) {
	var s string
	RunList(t,
		New().
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				s += "1"
			}),
		New().ExcludeOthers().
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				s += "2"
			}),
		New().
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				s += "3"
			}),
	)
	assert.Equal(t, "2", s)
}

func TestRunListParallel(t *testing.T) {
	t0 := time.Now()
	RunListParallel(t,
		New().
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				time.Sleep(time.Second)
			}),
		New().
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				time.Sleep(time.Second)
			}),
		New().
			AddTask(1000, func(w *struct{ WorkspaceBase }) {
				time.Sleep(time.Second)
			}),
	)
	d := time.Since(t0)
	assert.Less(t, d, 3*time.Second)
}
