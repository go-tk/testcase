package testcase_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-tk/testcase"
)

func TestExample(t *testing.T) {
	type C struct { // C for context
		// prepare
		Ctx context.Context
		URL string

		// check
		Resp *http.Response
		Err  error
	}
	tc := testcase.New(func(tc *testcase.TestCase[C]) {
		tc.C.Ctx = context.Background() // default

		tc.Callback("PREPARE")

		req, _ := http.NewRequestWithContext(tc.C.Ctx, "GET", tc.C.URL, nil)
		tc.C.Resp, tc.C.Err = http.DefaultClient.Do(req)

		tc.Callback("CHECK")
	})

	// CASE-1: http client gets https://httpbin.org/status/201
	//         should respond with the status code 201.
	tc.WithTag("status-201").
		WithCallback("PREPARE", func(tc *testcase.TestCase[C]) {
			tc.C.URL = "https://httpbin.org/status/201"
		}).
		WithCallback("CHECK", func(tc *testcase.TestCase[C]) {
			if tc.C.Err != nil {
				tc.T.Fatal(tc.C.Err)
			}
			tc.Assert.Equal(tc.C.Resp.StatusCode, 201)
		}).
		RunParallel(t)

	// CASE-2: http client gets https://httpbin.org/delay/60 with timeout 10ms
	//         should return with deadline exceeded error.
	tc.WithTag("deadline-exceeded").
		WithCallback("PREPARE", func(tc *testcase.TestCase[C]) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			tc.T.Cleanup(cancel)
			tc.C.Ctx = ctx
			tc.C.URL = "https://httpbin.org/delay/60"
		}).
		WithCallback("CHECK", func(tc *testcase.TestCase[C]) {
			tc.Assert.ErrorIs(tc.C.Err, context.DeadlineExceeded)
		}).
		RunParallel(t)
}
