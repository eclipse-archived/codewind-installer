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

package sechttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/globals"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/stretchr/testify/assert"
)

const connectionID = "testcon"

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

func TestDispatchHTTPRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	// remove insecureKeychain.json if it already exists
	os.Remove(security.GetPathToInsecureKeyring())

	tests := map[string]struct {
		useInsecureKeyring bool
	}{
		"use secure keyring":   {useInsecureKeyring: false},
		"use insecure keyring": {useInsecureKeyring: true},
	}
	for name, test := range tests {
		var originalUseInsecureKeyring = globals.UseInsecureKeyring
		globals.SetUseInsecureKeyring(test.useInsecureKeyring)
		t.Run(name, func(t *testing.T) {
			testDispatchHTTPRequest(t)
		})
		globals.SetUseInsecureKeyring(originalUseInsecureKeyring)
	}

	// remove insecureKeychain.json if it still exists
	os.Remove(security.GetPathToInsecureKeyring())
}

func testDispatchHTTPRequest(t *testing.T) {

	t.Run("returns the response from PFE when PFE is local and responds (with 200)", func(t *testing.T) {
		mockClientReturning200 := &MockResponse{StatusCode: http.StatusOK, Body: nil}
		mockConnection := connections.Connection{ID: "local"}
		mockRequest := httptest.NewRequest("GET", "/", nil)
		expectedResp, _ := mockClientReturning200.Do(mockRequest)

		gotResp, err := DispatchHTTPRequest(mockClientReturning200, mockRequest, &mockConnection)

		assert.Nil(t, err)
		assert.Equal(t, expectedResp, gotResp)
	})
	t.Run("returns the response from PFE when PFE is not local, "+
		"we can get an access token from the keyring, "+
		"and PFE responds (with 200)", func(t *testing.T) {
		security.DeleteSecretFromKeyring(connectionID, "access_token")
		security.StoreSecretInKeyring(connectionID, "access_token", "mockAccessToken")

		mockClientReturning200 := &MockResponse{StatusCode: http.StatusOK, Body: nil}
		mockConnection := connections.Connection{ID: connectionID}
		mockRequest := httptest.NewRequest("GET", "/", nil)
		expectedResp, _ := mockClientReturning200.Do(mockRequest)

		gotResp, err := DispatchHTTPRequest(mockClientReturning200, mockRequest, &mockConnection)
		assert.Nil(t, err)
		assert.Equal(t, expectedResp, gotResp)

		// cleanup
		security.DeleteSecretFromKeyring(connectionID, "access_token")
	})
	t.Run("returns the response from PFE when PFE is not local, "+
		"we cannot get an access token from the keyring, "+
		"we can get a refresh token from the keyring, "+
		"and PFE responds (with 200)", func(t *testing.T) {
		security.DeleteSecretFromKeyring(connectionID, "access_token")
		security.StoreSecretInKeyring(connectionID, "refresh_token", "mockRefreshToken")

		mockAuthTokenBytes, _ := json.Marshal(&security.AuthToken{})
		mockBody := ioutil.NopCloser(bytes.NewReader(mockAuthTokenBytes))
		mockClientReturning200 := &MockResponse{StatusCode: http.StatusOK, Body: mockBody}
		mockConnection := connections.Connection{ID: connectionID}
		mockRequest := httptest.NewRequest("GET", "/", nil)
		expectedResp, _ := mockClientReturning200.Do(mockRequest)

		gotResp, err := DispatchHTTPRequest(mockClientReturning200, mockRequest, &mockConnection)
		assert.Nil(t, err)
		assert.Equal(t, expectedResp, gotResp)

		// cleanup
		security.DeleteSecretFromKeyring(connectionID, "refresh_token")
	})
	t.Run("returns the correct error when PFE is not local, "+
		"we cannot get an access token from the keyring, "+
		"we cannot get a refresh token from the keyring, "+
		"and we cannot get cached credentials from the keyring", func(t *testing.T) {
		mockConnectionUsername := "mockconnectionusername"
		security.DeleteSecretFromKeyring(connectionID, "access_token")
		security.DeleteSecretFromKeyring(connectionID, "refresh_token")
		security.DeleteSecretFromKeyring(connectionID, mockConnectionUsername)

		mockClientReturning200 := &MockResponse{StatusCode: http.StatusOK, Body: nil}
		mockConnection := connections.Connection{ID: connectionID, Username: mockConnectionUsername}
		mockRequest := httptest.NewRequest("GET", "/", nil)

		gotResp, gotErr := DispatchHTTPRequest(mockClientReturning200, mockRequest, &mockConnection)
		assert.Nil(t, gotResp)
		errMissingPassword := "Unable to find password in keychain"
		expectedErr := &HTTPSecError{errOpNoPassword, errors.New(errMissingPassword), errMissingPassword}
		assert.Equal(t, expectedErr, gotErr)
	})
	t.Run("returns the response from PFE when PFE is not local, "+
		"we cannot get an access token from the keyring, "+
		"we cannot get a refresh token from the keyring, "+
		"and we can get cached credentials from the keyring"+
		"and PFE responds (with 200)", func(t *testing.T) {
		mockConnectionUsername := "mockconnectionusername"
		mockCachedCredentials := "mockCachedCredentials"
		security.DeleteSecretFromKeyring(connectionID, "access_token")
		security.DeleteSecretFromKeyring(connectionID, "refresh_token")
		security.StoreSecretInKeyring(connectionID, mockConnectionUsername, mockCachedCredentials)

		mockAuthTokenBytes, _ := json.Marshal(&security.AuthToken{})
		mockBody := ioutil.NopCloser(bytes.NewReader(mockAuthTokenBytes))
		mockClientReturning200 := &MockResponse{StatusCode: http.StatusOK, Body: mockBody}
		mockConnection := connections.Connection{
			ID:       connectionID,
			Username: mockConnectionUsername,
			AuthURL:  "mockAuthURL",
			Realm:    "mockRealm",
			ClientID: "mockClientID",
		}
		mockRequest := httptest.NewRequest("GET", "/", nil)
		expectedResp, _ := mockClientReturning200.Do(mockRequest)

		gotResp, err := DispatchHTTPRequest(mockClientReturning200, mockRequest, &mockConnection)
		assert.Nil(t, err)
		assert.Equal(t, expectedResp, gotResp)

		// cleanup
		security.DeleteSecretFromKeyring(connectionID, mockConnectionUsername)
	})
}
