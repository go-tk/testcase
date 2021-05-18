# testcase

[![GoDev](https://pkg.go.dev/badge/golang.org/x/pkgsite.svg)](https://pkg.go.dev/github.com/go-tk/testcase)
[![Build Status](https://travis-ci.com/go-tk/testcase.svg?branch=master)](https://travis-ci.com/github/go-tk/testcase)
[![Coverage Status](https://codecov.io/gh/go-tk/testcase/branch/master/graph/badge.svg)](https://codecov.io/gh/go-tk/testcase)

Tiny testing framework

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
                Client *http.Client

                Input          Input
                ExpectedOutput Output
        }

        // Create a test case template
        //
        // NOTE: Numbers 1000, 2000, 3000 are task IDs, tasks with lower IDs will be
        // executed before ones with higher IDs.
        tc := testcase.New(func(t *testing.T) *Workspace {
                // Create a new workspace
                return &Workspace{}
        }).Task(1000, func(t *testing.T, w *Workspace) {
                // Set up the workspace
                var transport http.Transport
                w.Client = &http.Client{Transport: &transport}
        }).Task(2000, func(t *testing.T, w *Workspace) {
                // Do the test
                resp, err := w.Client.Get(w.Input.URL)
                if !assert.NoError(t, err) {
                        t.FailNow()
                }
                resp.Body.Close()
                var output Output
                output.StatusCode = resp.StatusCode
                assert.Equal(t, w.ExpectedOutput, output)
        }).Task(3000, func(t *testing.T, w *Workspace) {
                // Clean up the workspace
                w.Client.CloseIdleConnections()
        })

        testcase.RunListParallel(t,
                tc.Copy().
                        When("get https://httpbin.org/status/200").
                        Then("should respond status code 200").
                        Task(1999, func(t *testing.T, w *Workspace) {
                                // Populate the input & expected output
                                w.Input.URL = "https://httpbin.org/status/200"
                                w.ExpectedOutput.StatusCode = 200
                        }),
                tc.Copy().
                        When("get https://httpbin.org/status/400").
                        Then("should respond status code 400").
                        Task(1999, func(t *testing.T, w *Workspace) {
                                // Populate the input & expected output
                                w.Input.URL = "https://httpbin.org/status/400"
                                w.ExpectedOutput.StatusCode = 400
                        }),
                tc.Copy().
                        When("get https://httpbin.org/status/500").
                        Then("should respond status code 500").
                        Task(1999, func(t *testing.T, w *Workspace) {
                                // Populate the input & expected output
                                w.Input.URL = "https://httpbin.org/status/500"
                                w.ExpectedOutput.StatusCode = 500
                        }),
        )
}
```
