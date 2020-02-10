/*******************************************************************************
 * Copyright (c) 2019 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package apiroutes

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// MockResponse mocks the response of a http client
type MockResponse struct {
	StatusCode int
	Body       io.ReadCloser
}

// MockMultipleResponses takes a slice of MockResponses and iterates through them on each request
// Used when a function makes multiple PFE API requests
type MockMultipleResponses struct {
	MockResponses []MockResponse
	Counter       int
}

// Do makes a http request
func (c *MockResponse) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       c.Body,
	}, nil
}

// Do makes a http request and increments the MockMultipleResponses counter
func (c *MockMultipleResponses) Do(req *http.Request) (*http.Response, error) {
	response := c.MockResponses[c.Counter]
	c.Counter++
	return &http.Response{
		StatusCode: response.StatusCode,
		Body:       response.Body,
	}, nil
}

// CreateMockResponseBody is a helper function that creates a mock JSON response body
func CreateMockResponseBody(mockResponse interface{}) io.ReadCloser {
	jsonResponse, _ := json.Marshal(mockResponse)
	body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
	return body
}
