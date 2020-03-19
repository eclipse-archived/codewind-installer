/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/stretchr/testify/assert"
)

const exampleBadJSON = "<This is not JSON>"

type ErrorReader struct {
}

func (r ErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("fake error reading body")
}

func Test_GetRegistrySecrets(t *testing.T) {
	t.Run("success case - returns nil error when PFE status code 200", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		assert.Nil(t, err)
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := GetRegistrySecrets(&mockConnection, "mockURL", mockClient)
		assert.Nil(t, err)
		assert.Equal(t, expectedRegistrySecrets, *actualRegistrySecrets)
	})
	t.Run("error case - returns error when PFE status code non 200", func(t *testing.T) {
		emptyBody := ioutil.NopCloser(bytes.NewReader([]byte{}))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: emptyBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := GetRegistrySecrets(&mockConnection, "mockURL", mockClient)
		assert.Error(t, err)
	})
	t.Run("error case - returns error when request fails", func(t *testing.T) {
		mockClient := &security.ClientMockRequestFail{}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := GetRegistrySecrets(&mockConnection, "mockURL", mockClient)
		assert.Nil(t, actualRegistrySecrets)
		assert.True(t, strings.Contains(err.Error(), "mock http request failure"))
	})
	t.Run("error case - returns error creating bad request", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		assert.Nil(t, err)
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		// Use control characters to create an invalid URL and trigger an error from http.NewRequest
		_, err = GetRegistrySecrets(&mockConnection, "\x00\x01\x02\x03\x04\x05\x06\x05\x7F", mockClient)
		fmt.Println(err)
		assert.Error(t, err)
	})
	t.Run("error case -  badJSONResponse", func(t *testing.T) {
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(exampleBadJSON)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := GetRegistrySecrets(&mockConnection, "mockURL", mockClient)
		var registrySecrets []RegistryResponse
		expectedError := json.Unmarshal([]byte(exampleBadJSON), &registrySecrets)
		assert.Equal(t, expectedError, err)
	})
	t.Run("error case - returns error on error reading response body", func(t *testing.T) {
		errorBody := ioutil.NopCloser(ErrorReader{})
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: errorBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := GetRegistrySecrets(&mockConnection, "mockURL", mockClient)
		assert.Error(t, err)
		assert.Equal(t, "fake error reading body", err.Error())
	})
}

func Test_AddRegistrySecret(t *testing.T) {
	t.Run("success case - returns nil error when PFE status code 201", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		assert.Nil(t, err)
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusCreated, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := AddRegistrySecret(&mockConnection, "mockURL", mockClient, "testdockerregistry", "testuser", "testpassword")
		assert.Nil(t, err)
		assert.Equal(t, expectedRegistrySecrets, *actualRegistrySecrets)
	})
	t.Run("error case - returns error when PFE status code non 201", func(t *testing.T) {
		emptyBody := ioutil.NopCloser(bytes.NewReader([]byte{}))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: emptyBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := AddRegistrySecret(&mockConnection, "mockURL", mockClient, "testdockerregistry", "testuser", "testpassword")
		assert.Error(t, err)
	})
	t.Run("error case - returns error when request fails", func(t *testing.T) {
		mockClient := &security.ClientMockRequestFail{}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := AddRegistrySecret(&mockConnection, "mockURL", mockClient, "testdockerregistry", "testuser", "testpassword")
		assert.Nil(t, actualRegistrySecrets)
		assert.True(t, strings.Contains(err.Error(), "mock http request failure"))
	})
	t.Run("error case - returns error creating bad request", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		assert.Nil(t, err)
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusCreated, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		// Use control characters to create an invalid URL and trigger an error from http.NewRequest
		_, err = AddRegistrySecret(&mockConnection, "\x00\x01\x02\x03\x04\x05\x06\x05\x7F", mockClient, "testdockerregistry", "testuser", "testpassword")
		fmt.Println(err)
		assert.Error(t, err)
	})
	t.Run("error case -  badJSONResponse", func(t *testing.T) {
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(exampleBadJSON)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusCreated, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := AddRegistrySecret(&mockConnection, "mockURL", mockClient, "testdockerregistry", "testuser", "testpassword")
		var registrySecrets []RegistryResponse
		expectedError := json.Unmarshal([]byte(exampleBadJSON), &registrySecrets)
		assert.Equal(t, expectedError, err)
	})
	t.Run("error case - returns error on error reading response body", func(t *testing.T) {
		errorBody := ioutil.NopCloser(ErrorReader{})
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: errorBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := AddRegistrySecret(&mockConnection, "mockURL", mockClient, "testdockerregistry", "testuser", "testpassword")
		assert.Error(t, err)
		assert.Equal(t, "fake error reading body", err.Error())
	})
}

func Test_DeleteRegistrySecret(t *testing.T) {
	t.Run("success case - returns nil error when PFE status code 200", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		assert.Nil(t, err)
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := RemoveRegistrySecret(&mockConnection, "mockURL", mockClient, "anothertestdockerregistry")
		assert.Nil(t, err)
		assert.Equal(t, expectedRegistrySecrets, *actualRegistrySecrets)
	})
	t.Run("error case - returns error when PFE status code non 200", func(t *testing.T) {
		emptyBody := ioutil.NopCloser(bytes.NewReader([]byte{}))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: emptyBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := RemoveRegistrySecret(&mockConnection, "mockURL", mockClient, "afakeregistry")
		assert.Error(t, err)
	})
	t.Run("error case - returns error when request fails", func(t *testing.T) {
		mockClient := &security.ClientMockRequestFail{}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := RemoveRegistrySecret(&mockConnection, "mockURL", mockClient, "anothertestdockerregistry")
		assert.Nil(t, actualRegistrySecrets)
		assert.True(t, strings.Contains(err.Error(), "mock http request failure"))
	})
	t.Run("error case - returns error creating bad request", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		assert.Nil(t, err)
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		// Use control characters to create an invalid URL and trigger an error from http.NewRequest
		_, err = RemoveRegistrySecret(&mockConnection, "\x00\x01\x02\x03\x04\x05\x06\x05\x7F", mockClient, "anothertestdockerregistry")
		fmt.Println(err)
		assert.Error(t, err)
	})
	t.Run("error case -  badJSONResponse", func(t *testing.T) {
		resBody := ioutil.NopCloser(bytes.NewReader([]byte(exampleBadJSON)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: resBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := RemoveRegistrySecret(&mockConnection, "mockURL", mockClient, "anothertestdockerregistry")
		var registrySecrets []RegistryResponse
		expectedError := json.Unmarshal([]byte(exampleBadJSON), &registrySecrets)
		assert.Equal(t, expectedError, err)
	})
	t.Run("error case - returns error on error reading response body", func(t *testing.T) {
		errorBody := ioutil.NopCloser(ErrorReader{})
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: errorBody}
		mockConnection := connections.Connection{ID: "local"}
		_, err := RemoveRegistrySecret(&mockConnection, "mockURL", mockClient, "anothertestdockerregistry")
		assert.Error(t, err)
		assert.Equal(t, "fake error reading body", err.Error())
	})
}
