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
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/stretchr/testify/assert"
)

func Test_GetRegistrySecrets(t *testing.T) {
	t.Run("success case - returns nil error when PFE status code 200", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		if err != nil {
			t.Fail()
		}
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := GetRegistrySecrets(&mockConnection, "mockURL", mockClient)
		assert.Nil(t, err)
		assert.Equal(t, expectedRegistrySecrets, actualRegistrySecrets)
	})
	t.Run("error case - returns error when PFE status code non 200", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: nil}
		mockConnection := connections.Connection{ID: "local"}
		_, err := GetRegistrySecrets(&mockConnection, "mockURL", mockClient)
		assert.Error(t, err)
	})
}

func Test_AddRegistrySecret(t *testing.T) {

	t.Run("success case - returns nil error when PFE status code 201", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		if err != nil {
			t.Fail()
		}
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusCreated, Body: body}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := AddRegistrySecret(&mockConnection, "mockURL", mockClient, "testdockerregistry", "testuser", "testpassword")
		assert.Nil(t, err)
		assert.Equal(t, expectedRegistrySecrets, actualRegistrySecrets)
	})
	t.Run("error case - returns error when PFE status code non 201", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: nil}
		mockConnection := connections.Connection{ID: "local"}
		_, err := AddRegistrySecret(&mockConnection, "mockURL", mockClient, "testdockerregistry", "testuser", "testpassword")
		assert.Error(t, err)
	})
}

func Test_DeleteRegistrySecret(t *testing.T) {
	t.Run("success case - returns nil error when PFE status code 200", func(t *testing.T) {
		expectedRegistrySecrets := []RegistryResponse{RegistryResponse{Address: "testdockerregistry", Username: "testuser"}}
		jsonResponse, err := json.Marshal(expectedRegistrySecrets)
		if err != nil {
			t.Fail()
		}
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
		mockConnection := connections.Connection{ID: "local"}
		actualRegistrySecrets, err := RemoveRegistrySecret(&mockConnection, "mockURL", mockClient, "anothertestdockerregistry")
		assert.Nil(t, err)
		assert.Equal(t, expectedRegistrySecrets, actualRegistrySecrets)
	})
	t.Run("error case - returns error when PFE status code non 200", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: nil}
		mockConnection := connections.Connection{ID: "local"}
		_, err := RemoveRegistrySecret(&mockConnection, "mockURL", mockClient, "afakeregistry")
		assert.Error(t, err)
	})
}
