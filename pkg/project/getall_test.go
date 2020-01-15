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

package project

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

func TestGetAll(t *testing.T) {
	t.Run("Expect success - complete project list should be returned", func(t *testing.T) {
		// construct mock response body and status code
		project1 := Project{ProjectID: "1234", Name: "App1"}
		project2 := Project{ProjectID: "9999", Name: "App2"}
		projectList := []Project{
			project1,
			project2,
		}
		jsonResponse, _ := json.Marshal(projectList)
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

		// construct a http client with our mock canned response
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}

		mockConnection := connections.Connection{ID: "local"}

		response, getAllError := GetAll(mockClient, &mockConnection, "dummyurl")
		if getAllError != nil {
			t.Fail()
		}
		assert.Equal(t, projectList, response)
	})
}
