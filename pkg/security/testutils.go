package security

import (
	"errors"
	"io"
	"net/http"
)

const testConnection = "LOCAL"
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

type ClientMockRequestFail struct {
}

// Do : perform do function
func (c *ClientMockRequestFail) Do(req *http.Request) (*http.Response, error) {
	return nil, errors.New("mock http request failure")
}
