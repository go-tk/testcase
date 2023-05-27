# testcase

[![GoDev](https://pkg.go.dev/badge/golang.org/x/pkgsite.svg)](https://pkg.go.dev/github.com/go-tk/testcase)
[![Workflow Status](https://github.com/go-tk/testcase/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/go-tk/testcase/actions/workflows/ci.yaml?query=branch%3Amain)
[![Coverage Status](https://codecov.io/gh/go-tk/testcase/branch/main/graph/badge.svg)](https://codecov.io/gh/go-tk/testcase/branch/main)

Tiny testing framework

## Example

```go
func TestExample(t *testing.T) {
        type C struct { // C for context
                ctx  context.Context
                url  string
                resp *http.Response
                err  error
        }
        tc := testcase.New(func(t *testing.T, c *C) {
                c.ctx = context.Background() // default

                testcase.Callback(t, "INIT")

                req, _ := http.NewRequestWithContext(c.ctx, "GET", c.url, nil)
                c.resp, c.err = http.DefaultClient.Do(req)
                if c.err == nil {
                        t.Cleanup(func() { c.resp.Body.Close() })
                }

                testcase.Callback(t, "CHECK")
        })

        // http client gets https://httpbin.org/delay/60 with timeout 100ms
        // should return with deadline exceeded error.
        tc.WithCallback("INIT", func(t *testing.T, c *C) {
                ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
                t.Cleanup(cancel)
                c.ctx = ctx
                c.url = "https://httpbin.org/status/200"
        }).WithCallback("CHECK", func(t *testing.T, c *C) {
                assert.ErrorIs(t, c.err, context.DeadlineExceeded)
        }).Run(t)

        // http client gets https://httpbin.org/status/(200|201|202)
        // should respond with the corresponding status code.
        for _, statusCode := range [...]int{200, 201, 202} {
                tc.WithTag(
                        fmt.Sprintf("status_code_%d", statusCode),
                ).WithCallback("INIT", func(t *testing.T, c *C) {
                        c.url = fmt.Sprintf("https://httpbin.org/status/%d", statusCode)
                }).WithCallback("CHECK", func(t *testing.T, c *C) {
                        if c.err != nil {
                                t.Fatal(c.err)
                        }
                        assert.Equal(t, statusCode, c.resp.StatusCode)
                }).RunParallel(t)
        }
}
```
