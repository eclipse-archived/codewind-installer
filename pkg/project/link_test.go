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
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/stretchr/testify/assert"
)

var errInvalidRequest = errors.New(textInvalidRequest)
var errTargetNotFound = errors.New(textProjectLinkTargetNotFound)
var errConflict = errors.New(textProjectLinkConflict)
var errUnknownHTTPCode = errors.New(textUnknownResponseCode)

func TestGetProjectLinks(t *testing.T) {
	t.Run("Expect success - project links should be returned", func(t *testing.T) {
		// construct mock response body and status code
		links := []Link{
			Link{ProjectID: "1234", ProjectURL: "URL1", EnvName: "ENV1"},
			Link{ProjectID: "9999", ProjectURL: "URL2", EnvName: "ENV2"},
		}
		jsonResponse, _ := json.Marshal(links)
		body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

		// construct a http client with our mock canned response
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}

		mockConnection := connections.Connection{ID: "local"}

		returnedLinks, projectLinkErr := GetProjectLinks(mockClient, &mockConnection, "dummyurl", "dummyProjectID")
		assert.Nil(t, projectLinkErr)
		assert.Equal(t, links, returnedLinks)
	})
	t.Run("Expect failure - request returns unknown HTTP code", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusPermanentRedirect, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		_, projectLinkErr := GetProjectLinks(mockClient, &mockConnection, "dummyurl", "dummyProjectID")
		wantedError := &ProjectError{errOpResponse, errUnknownHTTPCode, errUnknownHTTPCode.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
}

func TestCreateProjectLinks(t *testing.T) {
	t.Run("Expect success - project links should be created", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusAccepted, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := CreateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyTargetProjectID", "dummyEnvName")
		assert.Nil(t, projectLinkErr)
	})
	t.Run("Expect failure - request returns 400 Bad Request", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := CreateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyTargetProjectID", "dummyEnvName")
		wantedError := &ProjectError{errOpResponse, errInvalidRequest, errInvalidRequest.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns 404 Not Found", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusNotFound, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := CreateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyTargetProjectID", "dummyEnvName")
		wantedError := &ProjectError{errOpResponse, errTargetNotFound, errTargetNotFound.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns 409 Conflict", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusConflict, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := CreateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyTargetProjectID", "dummyEnvName")
		wantedError := &ProjectError{errOpResponse, errConflict, errConflict.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns unknown HTTP code", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusPermanentRedirect, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := CreateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyTargetProjectID", "dummyEnvName")
		wantedError := &ProjectError{errOpResponse, errUnknownHTTPCode, errUnknownHTTPCode.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
}

func TestUpdateProjectLinks(t *testing.T) {
	t.Run("Expect success - project links should be updated", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusAccepted, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := UpdateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName", "dummyUpdatedEnvName")
		assert.Nil(t, projectLinkErr)
	})
	t.Run("Expect failure - request returns 400 Bad Request", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := UpdateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName", "dummyUpdatedEnvName")
		wantedError := &ProjectError{errOpResponse, errInvalidRequest, errInvalidRequest.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns 404 Not Found", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusNotFound, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := UpdateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName", "dummyUpdatedEnvName")
		wantedError := &ProjectError{errOpResponse, errTargetNotFound, errTargetNotFound.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns 409 Conflict", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusConflict, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := UpdateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName", "dummyUpdatedEnvName")
		wantedError := &ProjectError{errOpResponse, errConflict, errConflict.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns unknown HTTP code", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusPermanentRedirect, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := UpdateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName", "dummyUpdatedEnvName")
		wantedError := &ProjectError{errOpResponse, errUnknownHTTPCode, errUnknownHTTPCode.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
}

func TestDeleteProjectLinks(t *testing.T) {
	t.Run("Expect success - project links should be deleted", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusAccepted, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := DeleteProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName")
		assert.Nil(t, projectLinkErr)
	})
	t.Run("Expect failure - request returns 400 Bad Request", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := DeleteProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName")
		wantedError := &ProjectError{errOpResponse, errInvalidRequest, errInvalidRequest.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns 404 Not Found", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusNotFound, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := DeleteProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName")
		wantedError := &ProjectError{errOpResponse, errTargetNotFound, errTargetNotFound.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns 409 Conflict", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusConflict, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := DeleteProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName")
		wantedError := &ProjectError{errOpResponse, errConflict, errConflict.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
	t.Run("Expect failure - request returns unknown HTTP code", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusPermanentRedirect, Body: nil}
		mockConnection := connections.Connection{ID: "local"}

		projectLinkErr := DeleteProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName")
		wantedError := &ProjectError{errOpResponse, errUnknownHTTPCode, errUnknownHTTPCode.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
}
