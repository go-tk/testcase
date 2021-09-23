# testcase

[![GoDev](https://pkg.go.dev/badge/golang.org/x/pkgsite.svg)](https://pkg.go.dev/github.com/go-tk/testcase
) [![Build Status](https://travis-ci.com/go-tk/testcase.svg?branch=master)](https://travis-ci.com/github/go-tk/testcase
) [![Coverage Status](https://codecov.io/gh/go-tk/testcase/branch/master/graph/badge.svg)](https://codecov.io/gh/go-tk/testcase
)

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
                testcase.WorkspaceBase // mandatory

                Client         *http.Client
                Input          Input
                ExpectedOutput Output
        }

        // Create a test case template.
        // NOTE: Numbers 10 & 20 are task IDs, tasks will be executed in ascending order of ID.
        tcTmpl := testcase.New().
                AddTask(10, func(w *Workspace) {
                        // Set up the workspace.
                        w.Client = &http.Client{Transport: &http.Transport{}}

                        w.AddCleanup(func() {
                                // Clean up the workspace.
                                // NOTE: Cleanups will be executed after all tasks are executed or
                                //       panics occur.
                                w.Client.CloseIdleConnections()
                        })
                }).
                AddTask(20, func(w *Workspace) {
                        // Do the test.
                        resp, err := w.Client.Get(w.Input.URL)
                        // NOTE: use `w.T()` instead of `t`.
                        if !assert.NoError(w.T(), err) {
                                w.T().FailNow()
                        }
                        resp.Body.Close()

                        // Compare the output with the expected output.
                        var output Output
                        output.StatusCode = resp.StatusCode
                        assert.Equal(w.T(), w.ExpectedOutput, output)
                })

        // Make copies of the test case template, insert new tasks into copies for populating
        // test data and then run them parallel.
        // NOTE: Tasks in each test case will be executed in order. Test cases will be run with
        //       brand-new and isolated workspaces, the same workspace is shared with each task
        //       in a test case.
        testcase.RunListParallel(t,
                tcTmpl.Copy().
                        Given("http client").
                        When("get https://httpbin.org/status/200").
                        Then("should respond status code 200").
                        AddTask(19, func(w *Workspace) {
                                // Populate the input & expected output.
                                w.Input.URL = "https://httpbin.org/status/200"
                                w.ExpectedOutput.StatusCode = 200
                        }),
                tcTmpl.Copy().
                        Given("http client").
                        When("get https://httpbin.org/status/400").
                        Then("should respond status code 400").
                        AddTask(19, func(w *Workspace) {
                                // Populate the input & expected output.
                                w.Input.URL = "https://httpbin.org/status/400"
                                w.ExpectedOutput.StatusCode = 400
                        }),
                tcTmpl.Copy().
                        Given("http client").
                        When("get https://httpbin.org/status/500").
                        Then("should respond status code 500").
                        AddTask(19, func(w *Workspace) {
                                // Populate the input & expected output.
                                w.Input.URL = "https://httpbin.org/status/500"
                                w.ExpectedOutput.StatusCode = 500
                        }),
        )
}
```
