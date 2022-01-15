package testcase_test

import (
	"net/http"
	"testing"

	"github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	// Define the workspace.
	type Workspace struct {
		C  *http.Client
		In struct { // input
			URL string
		}
		ExpOut, ActOut struct { // expected output & actual output
			StatusCode int
		}
	}

	// Create a test case template.
	// NOTE: Numbers 1, 2, 3 are step numbers, steps will be executed in ascending order of
	//       step number.
	tc := testcase.New().
		Step(1, func(t *testing.T, w *Workspace) {
			// Set up the workspace.
			w.C = &http.Client{Transport: &http.Transport{}}

			t.Cleanup(func() {
				// Clean up the workspace.
				// NOTE: Cleanups will be executed after all steps are executed
				//       or panics occur.
				w.C.CloseIdleConnections()
			})
		}).
		Step(2, func(t *testing.T, w *Workspace) {
			// Do the test.
			resp, err := w.C.Get(w.In.URL)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			w.ActOut.StatusCode = resp.StatusCode
			resp.Body.Close()
		}).
		Step(3, func(t *testing.T, w *Workspace) {
			// Compare the actual output with the expected output.
			assert.Equal(t, w.ExpOut, w.ActOut)
		})

	// Make copies of the test case template, insert new steps (1.5) for populating test
	// data and then run them parallel.
	// NOTE: Steps in each test case will be executed in order. Test cases will be run with
	//       isolated workspaces and each step in a test case shares the same workspace.
	testcase.RunListParallel(t,
		tc.Copy().
			Given("http client").
			When("get https://httpbin.org/status/200").
			Then("should respond status code 200").
			Step(1.5, func(t *testing.T, w *Workspace) { // Step 1.5
				// Populate the input & expected output.
				w.In.URL = "https://httpbin.org/status/200"
				w.ExpOut.StatusCode = 200
			}),
		tc.Copy().
			Given("http client").
			When("get https://httpbin.org/status/400").
			Then("should respond status code 400").
			Step(1.5, func(t *testing.T, w *Workspace) { // Step 1.5
				// Populate the input & expected output.
				w.In.URL = "https://httpbin.org/status/400"
				w.ExpOut.StatusCode = 400
			}),
		tc.Copy().
			Given("http client").
			When("get https://httpbin.org/status/500").
			Then("should respond status code 500").
			Step(1.5, func(t *testing.T, w *Workspace) { // Step 1.5
				// Populate the input & expected output.
				w.In.URL = "https://httpbin.org/status/500"
				w.ExpOut.StatusCode = 500
			}),
	)
}
