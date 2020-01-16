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

func TestGetProjectFromID(t *testing.T) {
	t.Run("Expect success - project should be returned", func(t *testing.T) {
		// construct mock response body and status code
		projectID := "1234"
		project := Project{ProjectID: projectID, Name: "App1", LocationOnDisk: "/diskplace"}
		jsonResponse, _ := json.Marshal(project)
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

		// construct a http client with our mock canned response
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}

		mockConnection := connections.Connection{ID: "local"}

		response, getAllError := GetProjectFromID(mockClient, &mockConnection, "dummyurl", projectID)
		if getAllError != nil {
			t.Fail()
		}
		assert.Equal(t, project, *response)
	})
}

func TestGetProjectIDFromName(t *testing.T) {
	t.Run("Expect success - the correct ID should be returned", func(t *testing.T) {
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

		projectID, getAllError := GetProjectIDFromName(mockClient, &mockConnection, "dummyurl", "App2")
		if getAllError != nil {
			t.Fail()
		}
		assert.Equal(t, "9999", projectID)
	})
}
