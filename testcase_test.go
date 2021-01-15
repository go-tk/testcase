package testcase_test

import (
	"testing"

	. "github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestProcedureOrder(t *testing.T) {
	var s string
	RunList(t, []TestCase{New(func(t *testing.T) struct{} { return struct{}{} }).
		Given("TODO").
		When("TODO").
		Then("TODO").
		PreSetup(func(t *testing.T, c struct{}) {
			s += "1"
		}).
		Setup(func(t *testing.T, c struct{}) {
			s += "2"
		}).
		PreRun(func(t *testing.T, c struct{}) {
			s += "3"
		}).
		Run(func(t *testing.T, c struct{}) {
			s += "4"
		}).
		PostRun(func(t *testing.T, c struct{}) {
			s += "5"
		}).
		Teardown(func(t *testing.T, c struct{}) {
			s += "6"
		}).
		PostTeardown(func(t *testing.T, c struct{}) {
			s += "7"
		})})
	assert.Equal(t, "1234567", s)
}

func TestRunProcedureMustBeSet(t *testing.T) {
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T) struct{} { return struct{}{} }).
			Given("TODO").
			When("TODO").
			Then("TODO").
			PreSetup(func(t *testing.T, c struct{}) {}).
			Setup(func(t *testing.T, c struct{}) {}).
			PreRun(func(t *testing.T, c struct{}) {}).
			PostRun(func(t *testing.T, c struct{}) {}).
			Teardown(func(t *testing.T, c struct{}) {}).
			PostTeardown(func(t *testing.T, c struct{}) {})})
	})
}

func TestContextFactoryTypeValidation(t *testing.T) {
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func() struct{} { return struct{}{} })})
	})
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T) {})})
	})
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T, s string) struct{} { return struct{}{} })})
	})
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T) (struct{}, string) { return struct{}{}, "" })})
	})
}

func TestProcedureTypeValidation(t *testing.T) {
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T) struct{} { return struct{}{} }).
			Setup(func(t *testing.T) {})})
	})
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T) struct{} { return struct{}{} }).
			Setup(func(c struct{}) {})})
	})
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T) struct{} { return struct{}{} }).
			Setup(func(t *testing.T, c int) {})})
	})
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T) struct{} { return struct{}{} }).
			Setup(func(t *testing.T, c struct{}, s string) {})})
	})
	assert.Panics(t, func() {
		RunList(t, []TestCase{New(func(t *testing.T) struct{} { return struct{}{} }).
			Setup(func(t *testing.T, c struct{}) string { return "" })})
	})
}
