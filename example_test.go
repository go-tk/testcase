package testcase_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	type C struct { // C for context
		ctx                context.Context
		url                string
		expectedStatusCode int
		expectedErr        error
	}

	tc := testcase.New(func(t *testing.T, c *C) {
		c.ctx = context.Background() // default

		testcase.DoCallback("SET_TEST_DATA", t, c)

		req, _ := http.NewRequestWithContext(c.ctx, "GET", c.url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}

		if c.expectedErr != nil {
			assert.ErrorIs(t, err, c.expectedErr)
			return
		}
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, c.expectedStatusCode, resp.StatusCode)
	})

	// http client gets https://httpbin.org/status/200 should succeed.
	tc.Copy().
		SetCallback("SET_TEST_DATA", func(t *testing.T, c *C) {
			c.url = "https://httpbin.org/status/200"
			c.expectedStatusCode = 200
		}).
		Run(t)

	// http client gets https://httpbin.org/delay/60 with timeout 100ms should return deadline exceeded error.
	tc.Copy().
		SetCallback("SET_TEST_DATA", func(t *testing.T, c *C) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)
			c.ctx = ctx
			c.url = "https://httpbin.org/delay/60"
			c.expectedErr = context.DeadlineExceeded
		}).
		Run(t)
}
