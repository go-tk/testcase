# testcase

[![GoDev](https://pkg.go.dev/badge/golang.org/x/pkgsite.svg)](https://pkg.go.dev/github.com/go-tk/testcase)
[![Workflow Status](https://github.com/go-tk/testcase/actions/workflows/main.yaml/badge.svg?branch=main)](https://github.com/go-tk/testcase/actions)
[![Coverage Status](https://codecov.io/gh/go-tk/testcase/branch/main/graph/badge.svg)](https://codecov.io/gh/go-tk/testcase)

Tiny testing framework (BDD-style)

## Example

```go
func TestExample(t *testing.T) {
        type Input struct {
                URL string
        }
        type Output struct {
                StatusCode int
        }
        type Workspace struct {
                Client         *http.Client
                Input          Input
                ExpectedOutput Output
        }

        // Create a test case template.
        // NOTE: Numbers 1 & 2 are step numbers, steps will be executed in ascending order of
        //       step number.
        tcTmpl := testcase.New().
                Step(1, func(t *testing.T, w *Workspace) {
                        // Set up the workspace.
                        w.Client = &http.Client{Transport: &http.Transport{}}

                        t.Cleanup(func() {
                                // Clean up the workspace.
                                // NOTE: Cleanups will be executed after all steps are executed
                                //       or panics occur.
                                w.Client.CloseIdleConnections()
                        })
                }).
                Step(2, func(t *testing.T, w *Workspace) {
                        // Do the test.
                        resp, err := w.Client.Get(w.Input.URL)
                        if !assert.NoError(t, err) {
                                t.FailNow()
                        }
                        resp.Body.Close()

                        // Compare the output with the expected output.
                        var output Output
                        output.StatusCode = resp.StatusCode
                        assert.Equal(t, w.ExpectedOutput, output)
                })

        // Make copies of the test case template, insert new steps into copies for populating
        // test data and then run them parallel.
        // NOTE: Steps in each test case will be executed in order. Test cases will be run with
        //       isolated workspaces and each step in a test case shares the same workspace.
        testcase.RunListParallel(t,
                tcTmpl.Copy().
                        Given("http client").
                        When("get https://httpbin.org/status/200").
                        Then("should respond status code 200").
                        Step(1.5, func(t *testing.T, w *Workspace) {
                                // Populate the input & expected output.
                                w.Input.URL = "https://httpbin.org/status/200"
                                w.ExpectedOutput.StatusCode = 200
                        }),
                tcTmpl.Copy().
                        Given("http client").
                        When("get https://httpbin.org/status/400").
                        Then("should respond status code 400").
                        Step(1.5, func(t *testing.T, w *Workspace) {
                                // Populate the input & expected output.
                                w.Input.URL = "https://httpbin.org/status/400"
                                w.ExpectedOutput.StatusCode = 400
                        }),
                tcTmpl.Copy().
                        Given("http client").
                        When("get https://httpbin.org/status/500").
                        Then("should respond status code 500").
                        Step(1.5, func(t *testing.T, w *Workspace) {
                                // Populate the input & expected output.
                                w.Input.URL = "https://httpbin.org/status/500"
                                w.ExpectedOutput.StatusCode = 500
                        }),
        )
}
```
