# testcase

[![GoDev](https://pkg.go.dev/badge/golang.org/x/pkgsite.svg)](https://pkg.go.dev/github.com/go-tk/testcase)
[![Build Status](https://travis-ci.com/go-tk/testcase.svg?branch=master)](https://travis-ci.com/github/go-tk/testcase)
[![Coverage Status](https://codecov.io/gh/go-tk/testcase/branch/master/graph/badge.svg)](https://codecov.io/gh/go-tk/testcase)

Tiny testing framework

## Example

```go
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
        type Context struct {
                Client *http.Client

                Input          Input
                ExpectedOutput Output
        }

        // Create a test case template.
        tc := testcase.New(func(t *testing.T) *Context {
                return &Context{}
        }).Setup(func(t *testing.T, c *Context) {
                var transport http.Transport
                c.Client = &http.Client{Transport: &transport}
        }).Run(func(t *testing.T, c *Context) {
                resp, err := c.Client.Get(c.Input.URL)
                if !assert.NoError(t, err) {
                        t.FailNow()
                }
                resp.Body.Close()
                var output Output
                output.StatusCode = resp.StatusCode
                assert.Equal(t, c.ExpectedOutput, output)
        }).Teardown(func(t *testing.T, c *Context) {
                c.Client.CloseIdleConnections()
        })

        testcase.RunListParallel(t,
                tc.Copy().
                        When("get https://httpbin.org/status/200").
                        Then("should respond status code 200").
                        PreRun(func(t *testing.T, c *Context) {
                                c.Input.URL = "https://httpbin.org/status/200"
                                c.ExpectedOutput.StatusCode = 200
                        }),
                tc.Copy().
                        When("get https://httpbin.org/status/400").
                        Then("should respond status code 400").
                        PreRun(func(t *testing.T, c *Context) {
                                c.Input.URL = "https://httpbin.org/status/400"
                                c.ExpectedOutput.StatusCode = 400
                        }),
                tc.Copy().
                        When("get https://httpbin.org/status/500").
                        Then("should respond status code 500").
                        PreRun(func(t *testing.T, c *Context) {
                                c.Input.URL = "https://httpbin.org/status/500"
                                c.ExpectedOutput.StatusCode = 500
                        }),
        )
}
```
