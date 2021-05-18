package testcase_test

import (
	"testing"

	. "github.com/go-tk/testcase/v2"
	"github.com/stretchr/testify/assert"
)

func TestTaskExecutionOrder(t *testing.T) {
	var s string
	RunList(t, New(func(t *testing.T) struct{} { return struct{}{} }).
		Given("NONE").
		When("NONE").
		Then("NONE").
		Task(1010, func(t *testing.T, w struct{}) {
			s += "2"
		}).
		Task(-5000, func(t *testing.T, w struct{}) {
			s += "1"
		}).
		Task(3003, func(t *testing.T, w struct{}) {
			s += "7"
		}).
		Task(3002, func(t *testing.T, w struct{}) {
			s += "6"
		}).
		Task(3001, func(t *testing.T, w struct{}) {
			s += "5"
		}).
		Task(2000, func(t *testing.T, w struct{}) {
			s += "3"
		}).
		Task(2020, func(t *testing.T, w struct{}) {
			s += "4"
		}))
	assert.Equal(t, "1234567", s)
}

func TestExcludeTestCase(t *testing.T) {
	var s []int
	RunList(t,
		New(func(t *testing.T) struct{} { return struct{}{} }).Task(0, func(t *testing.T, w struct{}) { s = append(s, 1) }),
		New(func(t *testing.T) struct{} { return struct{}{} }).Task(0, func(t *testing.T, w struct{}) { s = append(s, 2) }).Exclude(),
		New(func(t *testing.T) struct{} { return struct{}{} }).Task(0, func(t *testing.T, w struct{}) { s = append(s, 3) }),
	)
	assert.Equal(t, []int{1, 3}, s)
}

func TestExcludeOtherTestCases(t *testing.T) {
	var s []int
	RunList(t,
		New(func(t *testing.T) struct{} { return struct{}{} }).Task(0, func(t *testing.T, w struct{}) { s = append(s, 1) }),
		New(func(t *testing.T) struct{} { return struct{}{} }).Task(0, func(t *testing.T, w struct{}) { s = append(s, 2) }).ExcludeOthers(),
		New(func(t *testing.T) struct{} { return struct{}{} }).Task(0, func(t *testing.T, w struct{}) { s = append(s, 3) }),
	)
	assert.Equal(t, []int{2}, s)
}

func TestNoTask(t *testing.T) {
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T) struct{} { return struct{}{} }).
			Given("NONE").
			When("NONE").
			Then("NONE"))
	})
}

func TestInvalidTaskFactoryType(t *testing.T) {
	assert.Panics(t, func() {
		RunList(t, New(func() struct{} { return struct{}{} }))
	})
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T) {}))
	})
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T, s string) struct{} { return struct{}{} }))
	})
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T) (struct{}, string) { return struct{}{}, "" }))
	})
}

func TestInvalidTaskType(t *testing.T) {
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T) struct{} { return struct{}{} }).
			Task(0, func(t *testing.T) {}))
	})
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T) struct{} { return struct{}{} }).
			Task(0, func(w struct{}) {}))
	})
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T) struct{} { return struct{}{} }).
			Task(0, func(t *testing.T, w int) {}))
	})
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T) struct{} { return struct{}{} }).
			Task(0, func(t *testing.T, w struct{}, s string) {}))
	})
	assert.Panics(t, func() {
		RunList(t, New(func(t *testing.T) struct{} { return struct{}{} }).
			Task(0, func(t *testing.T, w struct{}) string { return "" }))
	})
}
