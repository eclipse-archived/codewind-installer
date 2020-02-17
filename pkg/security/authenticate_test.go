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

package security

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"github.com/zalando/go-keyring"
)

func Test_Authenticate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	const accessToken = "1234512345123451234512345"
	const refreshToken = "55555351513513513513513513513"

	// set mock cli flags
	set := flag.NewFlagSet("tests", 0)
	set.String("host", "https://mockserver/auth", "doc")
	set.String("realm", "master", "doc")
	set.String("username", testUsername, "doc")
	set.String("password", "testpassword", "doc")
	set.String("client", "testclient", "doc")
	set.String("conid", testConnection, "doc") // must be a valid connection (using local which will always exist)

	c := cli.NewContext(nil, set, nil)

	t.Run("Expect authentication failure - invalid credentials", func(t *testing.T) {

		// create a keycloak error message
		mockKeycloakResponse := KeycloakAPIError{HTTPStatus: http.StatusUnauthorized, Error: "invalid_grant", ErrorDescription: "Invalid user credentials"}
		jsonResponse, _ := json.Marshal(mockKeycloakResponse)
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

		// construct a http client with our mock canned response
		mockClient := &ClientMockAuthenticate{StatusCode: http.StatusUnauthorized, Body: body}

		// attempt authentication
		_, secError := SecAuthenticate(mockClient, c, "codewind", "codewind")
		if secError == nil {
			t.Fail()
		}
		assert.Equal(t, "invalid_grant", secError.Op)
		assert.Equal(t, "Invalid user credentials", secError.Desc)
	})

	t.Run("Expect authentication success - obtain access token from service", func(t *testing.T) {
		// construct mock response body and status code
		tokens := AuthToken{AccessToken: accessToken, RefreshToken: refreshToken}
		jsonResponse, _ := json.Marshal(tokens)
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

		// construct a http client with our mock canned response
		mockClient := &ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
		retrievedSecrets, secError := SecAuthenticate(mockClient, c, "altRealm", "altClient")
		if secError != nil {
			t.Fail()
		}
		assert.Equal(t, accessToken, retrievedSecrets.AccessToken)
	})

	t.Run("Expect authentication success - obtain access token using refresh token", func(t *testing.T) {
		// construct mock response body and status code
		tokens := AuthToken{AccessToken: accessToken, RefreshToken: refreshToken}
		jsonResponse, _ := json.Marshal(tokens)
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

		// set mock cli flags
		set := flag.NewFlagSet("tests", 0)
		set.String("conid", testConnection, "doc") // must be a valid connection
		c := cli.NewContext(nil, set, nil)

		// construct a http client with our mock canned response
		mockClient := &ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
		retrievedSecrets, secError := SecRefreshTokens(mockClient, c)
		if secError != nil {
			t.Fail()
		}
		assert.Equal(t, accessToken, retrievedSecrets.AccessToken)
	})

	t.Run("Expect authentication failure - unable to obtain access token using expired refresh token", func(t *testing.T) {
		// construct mock response body and status code
		mockKeycloakResponse := KeycloakAPIError{HTTPStatus: http.StatusUnauthorized, Error: "invalid_grant", ErrorDescription: "Refresh token expired"}
		jsonResponse, _ := json.Marshal(mockKeycloakResponse)
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

		// set mock cli flags
		set := flag.NewFlagSet("tests", 0)
		set.String("conid", testConnection, "doc") // must be a valid connection
		c := cli.NewContext(nil, set, nil)

		// construct a http client with our mock canned response
		mockClient := &ClientMockAuthenticate{StatusCode: http.StatusUnauthorized, Body: body}
		_, secError := SecRefreshTokens(mockClient, c)
		if secError == nil {
			t.Fail()
		}
		assert.Equal(t, "Refresh token expired", secError.Desc)
	})

	t.Run("Cleanup stored access_token and refresh_token", func(t *testing.T) {
		// Clean up test entries
		keyring.Delete(strings.ToLower(KeyringServiceName+"."+testConnection), "access_token")
		keyring.Delete(strings.ToLower(KeyringServiceName+"."+testConnection), "refresh_token")
	})
}
