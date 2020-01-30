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

package gatekeeper

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockResponse mocks the response of a http client
type MockResponse struct {
	StatusCode int
	Body       io.ReadCloser
}

// Do makes a http request
func (c *MockResponse) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       c.Body,
	}, nil
}

func Test_GetGatekeeperEnvironment(t *testing.T) {

	mockResponse := GatekeeperEnvironment{AuthURL: "http://a.mock.auth.server.remote:1234", Realm: "remoteRealm", ClientID: "remoteClient"}
	jsonResponse, _ := json.Marshal(mockResponse)
	body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

	// construct a http client with our mock canned response
	mockClient := &MockResponse{StatusCode: http.StatusOK, Body: body}
	gatekeeperEnv, err := GetGatekeeperEnvironment(mockClient, "http://noserver.test.com")
	if err != nil {
		t.Fail()
	}

	t.Run("Assert realm is remoteRealm", func(t *testing.T) {
		assert.Equal(t, "remoteRealm", gatekeeperEnv.Realm)
	})
	t.Run("Assert realm is remoteRealm", func(t *testing.T) {
		assert.Equal(t, "http://a.mock.auth.server.remote:1234", gatekeeperEnv.AuthURL)
	})
	t.Run("Assert realm is remoteRealm", func(t *testing.T) {
		assert.Equal(t, "remoteClient", gatekeeperEnv.ClientID)
	})
}

func Test_GetGatekeeperEnvironmentBadHost(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte("<HTML></HTML>")))
	mockClient := &MockResponse{StatusCode: http.StatusOK, Body: body}

	t.Run("Assert a bad response fails with error", func(t *testing.T) {
		gatekeeperEnv, err := GetGatekeeperEnvironment(mockClient, "http://noserver.test.com")
		if err == nil {
			t.Fail()
		}
		assert.Nil(t, gatekeeperEnv)
	})

}
