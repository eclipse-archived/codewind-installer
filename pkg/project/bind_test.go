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

package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
)

func TestBindToPFE(t *testing.T) {
	creationTime := time.Now().UnixNano() / 100000

	exampleBindRequest := BindRequest{
		Language:    "javascript",
		Name:        "test",
		ProjectType: "nodejs",
		Path:        "examplePath",
		Time:        creationTime,
	}

	exampleBindResponse := BindResponse{
		ProjectID:     "abcd",
		UploadedFiles: []UploadedFile{},
		Status:        "success",
		StatusCode:    http.StatusOK,
	}

	exampleBadJSON := "<This is not JSON>"

	successTests := map[string]struct {
		bindRequest  BindRequest
		bindResponse BindResponse
	}{
		"Expect success - project should be bound": {
			bindRequest:  exampleBindRequest,
			bindResponse: exampleBindResponse,
		},
	}
	for name, test := range successTests {
		t.Run(name, func(t *testing.T) {
			jsonResponse, _ := json.Marshal(test.bindResponse)
			body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
			mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
			mockConnection := connections.Connection{ID: "local"}
			got, projErr := bindToPFE(mockClient, test.bindRequest, &mockConnection, "dummyurl")
			if projErr != nil {
				t.Errorf("bindToPFE() returned the following error: %s", projErr.Desc)
			}
			assert.Equal(t, got, &test.bindResponse)
		})
	}

	failureTests := map[string]struct {
		bindRequest      BindRequest
		bindResponse     BindResponse
		mockResponseCode int
		wantedError      ProjectError
	}{
		"Expect failure - API not found": {
			bindRequest:      exampleBindRequest,
			mockResponseCode: http.StatusNotFound,
			wantedError:      ProjectError{errOpResponse, errors.New(textAPINotFound), textAPINotFound},
		},
		"Expect failure - bad request": {
			bindRequest:      exampleBindRequest,
			mockResponseCode: http.StatusBadRequest,
			wantedError:      ProjectError{errOpResponse, errors.New(textInvalidType), textInvalidType},
		},
		"Expect failure - duplicate name": {
			bindRequest:      exampleBindRequest,
			mockResponseCode: http.StatusConflict,
			wantedError:      ProjectError{errOpResponse, errors.New(textDupName), textDupName},
		},
	}

	for name, test := range failureTests {
		t.Run(name, func(t *testing.T) {
			mockClient := &security.ClientMockAuthenticate{StatusCode: test.mockResponseCode, Body: nil}
			mockConnection := connections.Connection{ID: "local"}
			_, projErr := bindToPFE(mockClient, test.bindRequest, &mockConnection, "dummyurl")
			if projErr == nil {
				t.Fail()
			}
			assert.Equal(t, test.wantedError, *projErr)
		})
	}

	t.Run("badJSONResponse", func(t *testing.T) {
		body := ioutil.NopCloser(bytes.NewReader([]byte(exampleBadJSON)))
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
		mockConnection := connections.Connection{ID: "local"}
		_, projErr := bindToPFE(mockClient, exampleBindRequest, &mockConnection, "dummyurl")
		var projectInfo *BindResponse
		expectedError := json.Unmarshal([]byte(exampleBadJSON), &projectInfo)
		assert.Equal(t, ProjectError{errOpResponse, expectedError, "invalid character '\u003c' looking for beginning of value"}, *projErr)
	})
}

func TestCompleteBind(t *testing.T) {
	mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: nil}
	mockConnection := connections.Connection{ID: "local"}
	_, gotStatusCode := completeBind(mockClient, "testID", "dummyURL", &mockConnection)
	assert.Equal(t, gotStatusCode, http.StatusOK)
}
