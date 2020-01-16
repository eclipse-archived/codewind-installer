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
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/stretchr/testify/assert"
)

func Test_IgnoredPaths(t *testing.T) {
	t.Run("Asserts correct ignoredPaths are returned", func(t *testing.T) {
		testIgnoredPaths := IgnoredPaths{"*/.dockerigore", "*/.gitignore"}
		jsonResponse, err := json.Marshal(testIgnoredPaths)
		if err != nil {
			t.Fail()
		}
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
		mockConnection := connections.Connection{ID: "local"}
		gotIgnoredPaths, err := GetIgnoredPaths(mockClient, &mockConnection, "nodejs", "dummyurl")
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, testIgnoredPaths, gotIgnoredPaths)
	})
	t.Run("Asserts 400 response from PFE returns an error", func(t *testing.T) {
		mockConnection := connections.Connection{ID: "local"}
		mockClientFalse := &security.ClientMockAuthenticate{StatusCode: http.StatusNotFound, Body: nil}
		_, err := GetIgnoredPaths(mockClientFalse, &mockConnection, "nodejs", "dummyurl")
		if err == nil {
			t.Fail()
		}
	})
}
