package security

import (
	"io"
	"net/http"
)

const testDeployment = "LOCAL"
const testUsername = "unit_test_user"

// ClientMockAuthenticate : Client Mock with a concrete response and status code
type ClientMockAuthenticate struct {
	StatusCode int
	Body       io.ReadCloser
}

// Do : perform do function
func (c *ClientMockAuthenticate) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       c.Body,
	}, nil
}
