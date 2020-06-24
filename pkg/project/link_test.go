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
var errUnknownNotFound = errors.New(textProjectLinkUnknownNotFound)
var errConflict = errors.New(textProjectLinkConflict)
var errUnknownHTTPCode = errors.New(textUnknownResponseCode)

var projectLinkCreateUpdateDeleteTests = []struct {
	statusCode int
	want       *ProjectError
}{
	{statusCode: http.StatusAccepted, want: nil},
	{statusCode: http.StatusBadRequest, want: &ProjectError{errOpResponse, errInvalidRequest, errInvalidRequest.Error()}},
	{statusCode: http.StatusNotFound, want: &ProjectError{errOpResponse, errUnknownNotFound, errUnknownNotFound.Error()}},
	{statusCode: http.StatusConflict, want: &ProjectError{errOpResponse, errConflict, errConflict.Error()}},
	{statusCode: http.StatusPermanentRedirect, want: &ProjectError{errOpResponse, errUnknownHTTPCode, errUnknownHTTPCode.Error()}},
}

func TestGetProjectLinks(t *testing.T) {
	t.Run("Expect success - project links should be returned", func(t *testing.T) {
		// construct mock response body and status code
		links := []Link{
			Link{ProjectID: "1234", ProjectName: "name1", ProjectURL: "URL1", EnvName: "ENV1"},
			Link{ProjectID: "9999", ProjectName: "name2", ProjectURL: "URL2", EnvName: "ENV2"},
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
		emptyBody := ioutil.NopCloser(bytes.NewReader([]byte{}))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusPermanentRedirect, Body: emptyBody}
		mockConnection := connections.Connection{ID: "local"}

		_, projectLinkErr := GetProjectLinks(mockClient, &mockConnection, "dummyurl", "dummyProjectID")
		wantedError := &ProjectError{errOpResponse, errUnknownHTTPCode, errUnknownHTTPCode.Error()}
		assert.Equal(t, wantedError, projectLinkErr)
	})
}

func TestCreateProjectLinks(t *testing.T) {
	for _, tt := range projectLinkCreateUpdateDeleteTests {
		emptyBody := ioutil.NopCloser(bytes.NewReader([]byte{}))
		mockClient := &security.ClientMockAuthenticate{StatusCode: tt.statusCode, Body: emptyBody}
		mockConnection := connections.Connection{ID: "local"}
		got := CreateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyTargetProjectID", "dummyEnvName")
		assert.Equal(t, tt.want, got)
	}
}

func TestUpdateProjectLinks(t *testing.T) {
	for _, tt := range projectLinkCreateUpdateDeleteTests {
		emptyBody := ioutil.NopCloser(bytes.NewReader([]byte{}))
		mockClient := &security.ClientMockAuthenticate{StatusCode: tt.statusCode, Body: emptyBody}
		mockConnection := connections.Connection{ID: "local"}
		got := UpdateProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName", "dummyUpdatedEnvName")
		assert.Equal(t, tt.want, got)
	}
}

func TestDeleteProjectLinks(t *testing.T) {
	for _, tt := range projectLinkCreateUpdateDeleteTests {
		emptyBody := ioutil.NopCloser(bytes.NewReader([]byte{}))
		mockClient := &security.ClientMockAuthenticate{StatusCode: tt.statusCode, Body: emptyBody}
		mockConnection := connections.Connection{ID: "local"}
		got := DeleteProjectLink(mockClient, &mockConnection, "dummyurl", "dummyProjectID", "dummyEnvName")
		assert.Equal(t, tt.want, got)
	}
}
