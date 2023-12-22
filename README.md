# testcase

[![GoDev](https://pkg.go.dev/badge/golang.org/x/pkgsite.svg)](https://pkg.go.dev/github.com/go-tk/testcase)
[![Workflow Status](https://github.com/go-tk/testcase/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/go-tk/testcase/actions/workflows/ci.yaml?query=branch%3Amain)
[![Coverage Status](https://codecov.io/gh/go-tk/testcase/branch/main/graph/badge.svg)](https://codecov.io/gh/go-tk/testcase/branch/main)

Tiny testing framework

## Example

```go
func TestExample(t *testing.T) {
    type C struct { // C for context
        // prepare
        ctx context.Context
        url string

        // check
        resp *http.Response
        err  error
    }
    tc := testcase.New(func(t *testing.T, c *C) {
        c.ctx = context.Background() // default

        testcase.Callback(t, "PREPARE")

        req, _ := http.NewRequestWithContext(c.ctx, "GET", c.url, nil)
        c.resp, c.err = http.DefaultClient.Do(req)

        testcase.Callback(t, "CHECK")
    })

    // CASE-1: http client gets https://httpbin.org/delay/60 with timeout 100ms
    //         should return with deadline exceeded error.
    tc.WithTag("delay-60").
        WithCallback("PREPARE", func(t *testing.T, c *C) {
            ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
            t.Cleanup(cancel)
            c.ctx = ctx
            c.url = "https://httpbin.org/delay/60"
        }).
        WithCallback("CHECK", func(t *testing.T, c *C) {
            assert.ErrorIs(t, c.err, context.DeadlineExceeded)
        }).
        RunParallel(t)

    // CASE-2: http client gets https://httpbin.org/status/201
    //         should respond with the status code 201.
    tc.WithTag("status-201").
        WithCallback("PREPARE", func(t *testing.T, c *C) {
            c.url = "https://httpbin.org/status/201"
        }).
        WithCallback("CHECK", func(t *testing.T, c *C) {
            if c.err != nil {
                t.Fatal(c.err)
            }
            assert.Equal(t, c.resp.StatusCode, 201)
        }).
        RunParallel(t)
}
```
