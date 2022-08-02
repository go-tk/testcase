# testcase

[![GoDev](https://pkg.go.dev/badge/golang.org/x/pkgsite.svg)](https://pkg.go.dev/github.com/go-tk/testcase)
[![Workflow Status](https://github.com/go-tk/testcase/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/go-tk/testcase/actions/workflows/ci.yaml?query=branch%3Amain)
[![Coverage Status](https://codecov.io/gh/go-tk/testcase/branch/main/graph/badge.svg)](https://codecov.io/gh/go-tk/testcase/branch/main)

Tiny testing framework (BDD-style)

## Example

```go
func TestExample(t *testing.T) {
        type context struct {
                url                string
                expectedStatusCode int
        }

        tc := testcase.New(func(t *testing.T, c *context) {
                t.Parallel()

                testcase.DoCallback("SET_TEST_DATA", t, c)

                client := &http.Client{Transport: &http.Transport{}}
                defer client.CloseIdleConnections()
                resp, err := client.Get(c.url)
                if !assert.NoError(t, err) {
                        t.FailNow()
                }
                defer resp.Body.Close()
                assert.Equal(t, c.expectedStatusCode, resp.StatusCode)
        })

        // get https://httpbin.org/status/200 should response status code 200
        tc.Copy().
                SetCallback("SET_TEST_DATA", func(t *testing.T, c *context) {
                        c.url = "https://httpbin.org/status/200"
                        c.expectedStatusCode = 200
                }).
                Run(t)

        // get https://httpbin.org/status/400 should response status code 400
        tc.Copy().
                SetCallback("SET_TEST_DATA", func(t *testing.T, c *context) {
                        c.url = "https://httpbin.org/status/400"
                        c.expectedStatusCode = 400
                }).
                Run(t)

        // get https://httpbin.org/status/500 should response status code 500
        tc.Copy().
                SetCallback("SET_TEST_DATA", func(t *testing.T, c *context) {
                        c.url = "https://httpbin.org/status/500"
                        c.expectedStatusCode = 500
                }).
                Run(t)
}
```
