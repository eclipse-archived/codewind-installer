package utils

import "net/http"

// HTTPClient : An net HTTP Client to simplify testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
