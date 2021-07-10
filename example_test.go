package testcase_test

import (
	"net/http"
	"testing"

	"github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	type Input struct {
		URL string
	}
	type Output struct {
		StatusCode int
	}
	type Workspace struct {
		testcase.WorkspaceBase // mandatory

		Client         *http.Client
		Input          Input
		ExpectedOutput Output
	}

	// Create a test case template.
	// NOTE: Numbers 1000 & 2000 are task IDs, tasks will be executed in ascending order of ID.
	tcTmpl := testcase.New().
		AddTask(1000, func(w *Workspace) {
			// Set up the workspace.
			w.Client = &http.Client{Transport: &http.Transport{}}

			w.AddCleanup(func() {
				// Clean up the workspace.
				// NOTE: Cleanups will be executed after all tasks are executed or
				//       panics occur.
				w.Client.CloseIdleConnections()
			})
		}).
		AddTask(2000, func(w *Workspace) {
			// Do the test.
			// NOTE: use `w.T()` instead of `t`.
			resp, err := w.Client.Get(w.Input.URL)
			if !assert.NoError(w.T(), err) {
				w.T().FailNow()
			}
			resp.Body.Close()

			var output Output
			output.StatusCode = resp.StatusCode
			assert.Equal(w.T(), w.ExpectedOutput, output)
		})

	// Make copies of test case template, insert new tasks into it for populating test data
	// and then run them parallel.
	// NOTE: Tasks in each test case will be executed in order. Test cases will be run with
	//       brand-new and isolated workspaces, the same workspace is shared with each task
	//       in a test case.
	testcase.RunListParallel(t,
		tcTmpl.Copy().
			Given("http client").
			When("get https://httpbin.org/status/200").
			Then("should respond status code 200").
			AddTask(1999, func(w *Workspace) {
				// Populate the input & expected output.
				w.Input.URL = "https://httpbin.org/status/200"
				w.ExpectedOutput.StatusCode = 200
			}),
		tcTmpl.Copy().
			Given("http client").
			When("get https://httpbin.org/status/400").
			Then("should respond status code 400").
			AddTask(1999, func(w *Workspace) {
				// Populate the input & expected output.
				w.Input.URL = "https://httpbin.org/status/400"
				w.ExpectedOutput.StatusCode = 400
			}),
		tcTmpl.Copy().
			Given("http client").
			When("get https://httpbin.org/status/500").
			Then("should respond status code 500").
			AddTask(1999, func(w *Workspace) {
				// Populate the input & expected output.
				w.Input.URL = "https://httpbin.org/status/500"
				w.ExpectedOutput.StatusCode = 500
			}),
	)
}
