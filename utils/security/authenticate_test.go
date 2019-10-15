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
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

// ClientMockAuthenticate: Client Mock with a concrete response and status code

type ClientMockAuthenticate struct {
	StatusCode int
	Body       io.ReadCloser
}

func (c *ClientMockAuthenticate) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       c.Body,
	}, nil
}

func Test_Authenticate(t *testing.T) {

	const accessToken = "1234512345123451234512345"
	const refreshToken = "55555351513513513513513513513"

	// set mock cli flags
	set := flag.NewFlagSet("tests", 0)
	set.String("host", "https://mockserver/auth", "doc")
	set.String("realm", "master", "doc")
	set.String("username", "testuser", "doc")
	set.String("password", "testpassword", "doc")
	set.String("client", "testclient", "doc")
	c := cli.NewContext(nil, set, nil)

	t.Run("Expect authentication failure - invalid credentials", func(t *testing.T) {
		mockKeycloakResponse := KeycloakAPIError{HTTPStatus: http.StatusUnauthorized, Error: "invalid_grant", ErrorDescription: "Invalid user credentials"}
		jsonResponse, _ := json.Marshal(mockKeycloakResponse)
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

		// construct a http client with our mock canned response
		mockClient := &ClientMockAuthenticate{StatusCode: http.StatusUnauthorized, Body: body}

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

}
