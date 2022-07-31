package testcase_test

import (
	"net/http"
	"testing"

	"github.com/go-tk/testcase"
	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	type context struct {
		url                string
		expectedStatusCode int
	}

	tc := testcase.New(func(t *testing.T, c *context) {
		t.Parallel()

		client := &http.Client{Transport: &http.Transport{}}
		defer client.CloseIdleConnections()

		testcase.DoCallback(0, t, c)

		resp, err := client.Get(c.url)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		defer resp.Body.Close()
		assert.Equal(t, c.expectedStatusCode, resp.StatusCode)
	})

	// get https://httpbin.org/status/200 should response status code 200
	tc.Copy().
		SetCallback(0, func(t *testing.T, c *context) {
			c.url = "https://httpbin.org/status/200"
			c.expectedStatusCode = 200
		}).
		Run(t)

	// get https://httpbin.org/status/400 should response status code 400
	tc.Copy().
		SetCallback(0, func(t *testing.T, c *context) {
			c.url = "https://httpbin.org/status/400"
			c.expectedStatusCode = 400
		}).
		Run(t)

	// get https://httpbin.org/status/500 should response status code 500
	tc.Copy().
		SetCallback(0, func(t *testing.T, c *context) {
			c.url = "https://httpbin.org/status/500"
			c.expectedStatusCode = 500
		}).
		Run(t)
}
