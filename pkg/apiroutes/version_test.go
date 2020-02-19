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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/stretchr/testify/assert"
)

func Test_GetAllContainerVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	t.Run("Asserts PFE ready", func(t *testing.T) {
		pfeBody1 := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "1"})
		performanceBody1 := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "2"})
		pfeBody2 := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "3"})
		performanceBody2 := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "nil"})

		mockClient := MockMultipleResponses{
			Counter: 0,
			MockResponses: []MockResponse{
				{StatusCode: http.StatusOK, Body: pfeBody1},
				{StatusCode: http.StatusOK, Body: performanceBody1},
				{StatusCode: http.StatusOK, Body: pfeBody2},
				{StatusCode: http.StatusOK, Body: performanceBody2},
			},
		}

		mockConnections := []connections.Connection{
			connections.Connection{ID: "local"},
			connections.Connection{ID: "notlocal"},
		}

		versions, err := GetAllContainerVersions(mockConnections, "latest", &mockClient)

		expectedLocalVersion := ContainerVersions{
			PFEVersion:         "x.x.dev-1",
			PerformanceVersion: "x.x.dev-2",
		}

		assert.Nil(t, err)
		assert.Equal(t, "latest", versions.CwctlVersion)
		// Check that local had its version information returned correctly
		assert.Equal(t, expectedLocalVersion, versions.Connections["local"])
		assert.Empty(t, versions.Connections["notlocal"])
		// Check that local didn't error and that notlocal did
		assert.Nil(t, versions.ConnectionErrors["local"])
		assert.Error(t, versions.ConnectionErrors["notlocal"])
	})
}

func Test_GetContainerVersions(t *testing.T) {
	t.Run("Gets the version of cwctl and the PFE, Performance containers when the connection ID = local (no Gatekeeper)", func(t *testing.T) {
		pfeBody := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "1"})
		performanceBody := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "2"})

		mockClient := MockMultipleResponses{
			Counter: 0,
			MockResponses: []MockResponse{
				{StatusCode: http.StatusOK, Body: pfeBody},
				{StatusCode: http.StatusOK, Body: performanceBody},
			},
		}

		mockConnection := connections.Connection{ID: "local"}

		versions, err := GetContainerVersions("www.pfe.com/", "latest", &mockConnection, &mockClient)

		assert.Nil(t, err)
		assert.Equal(t, "latest", versions.CwctlVersion)
		assert.Equal(t, "x.x.dev-1", versions.PFEVersion)
		assert.Equal(t, "x.x.dev-2", versions.PerformanceVersion)
		assert.Empty(t, versions.GatekeeperVersion)
		// Ensure all mock responses have been used
		assert.Equal(t, mockClient.Counter, len(mockClient.MockResponses))
	})
}

func Test_GetPFEVersionFromConnection(t *testing.T) {
	t.Run("Gets the version of the PFE container with a mocked httpClient", func(t *testing.T) {
		mockBody := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "20200129-142743"})
		mockClient := MockResponse{StatusCode: http.StatusOK, Body: mockBody}
		mockConnection := connections.Connection{ID: "local"}

		version, err := GetPFEVersionFromConnection(&mockConnection, "www.pfe.com/", &mockClient)
		assert.Nil(t, err)
		assert.Equal(t, "x.x.dev-20200129-142743", version)
	})

	t.Run("Errors as the response body is not JSON", func(t *testing.T) {
		mockBody := ioutil.NopCloser(strings.NewReader("bad res }}}"))
		mockClient := MockResponse{StatusCode: http.StatusOK, Body: mockBody}
		mockConnection := connections.Connection{ID: "local"}

		version, err := GetPFEVersionFromConnection(&mockConnection, "www.pfe.com/performance", &mockClient)
		assert.Error(t, err)
		assert.Empty(t, version)
	})
}

func Test_GetGatekeeperVersionFromConnection(t *testing.T) {
	t.Run("Gets the version of the Gatekeeper container with a mocked httpClient", func(t *testing.T) {
		mockBody := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "20200129-142743"})
		mockClient := MockResponse{StatusCode: http.StatusOK, Body: mockBody}
		mockConnection := connections.Connection{ID: "local"}

		version, err := GetGatekeeperVersionFromConnection(&mockConnection, "www.pfe.com/gatekeeper", &mockClient)
		assert.Nil(t, err)
		assert.Equal(t, "x.x.dev-20200129-142743", version)
	})

	t.Run("Errors as the response body is not JSON", func(t *testing.T) {
		mockBody := ioutil.NopCloser(strings.NewReader("bad res }}}"))
		mockClient := MockResponse{StatusCode: http.StatusOK, Body: mockBody}
		mockConnection := connections.Connection{ID: "local"}

		version, err := GetGatekeeperVersionFromConnection(&mockConnection, "www.pfe.com/performance", &mockClient)
		assert.Error(t, err)
		assert.Empty(t, version)
	})
}

func Test_GetPerformanceVersionFromConnection(t *testing.T) {
	t.Run("Gets the version of the Performance container with a mocked httpClient", func(t *testing.T) {
		mockBody := CreateMockResponseBody(EnvResponse{Version: "x.x.dev", ImageBuildTime: "20200129-142743"})
		mockClient := MockResponse{StatusCode: http.StatusOK, Body: mockBody}
		mockConnection := connections.Connection{ID: "local"}

		version, err := GetPerformanceVersionFromConnection(&mockConnection, "www.pfe.com/performance", &mockClient)
		assert.Nil(t, err)
		assert.Equal(t, "x.x.dev-20200129-142743", version)
	})

	t.Run("Errors as the response body is not JSON", func(t *testing.T) {
		mockBody := ioutil.NopCloser(strings.NewReader("bad res }}}"))
		mockClient := MockResponse{StatusCode: http.StatusOK, Body: mockBody}
		mockConnection := connections.Connection{ID: "local"}

		version, err := GetPerformanceVersionFromConnection(&mockConnection, "www.pfe.com/performance", &mockClient)
		assert.Error(t, err)
		assert.Empty(t, version)
	})
}
