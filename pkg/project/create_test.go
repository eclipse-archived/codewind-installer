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
	"log"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/stretchr/testify/assert"
)

func TestDetermineProjectInfo(t *testing.T) {
	tests := map[string]struct {
		in            string
		wantLanguage  string
		wantBuildType string
		wantedErr     error
	}{
		"success case: liberty project": {
			in:            path.Join("../..", "resources", "test", "liberty-project"),
			wantLanguage:  "java",
			wantBuildType: "liberty",
		},
		"success case: spring project": {
			in:            path.Join("../..", "resources", "test", "spring-project"),
			wantLanguage:  "java",
			wantBuildType: "spring",
		},
		"success case: node.js project": {
			in:            path.Join("../..", "resources", "test", "node-project"),
			wantLanguage:  "javascript",
			wantBuildType: "nodejs",
		},
		"success case: swift project": {
			in:            path.Join("../..", "resources", "test", "swift-project"),
			wantLanguage:  "swift",
			wantBuildType: "swift",
		},
		"success case: python project": {
			in:            path.Join("../..", "resources", "test", "python-project"),
			wantLanguage:  "python",
			wantBuildType: "docker",
		},
		"success case: go project": {
			in:            path.Join("../..", "resources", "test", "go-project"),
			wantLanguage:  "go",
			wantBuildType: "docker",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotLanguage, gotBuildType := determineProjectInfo(test.in)

			assert.Equal(t, test.wantLanguage, gotLanguage)
			assert.Equal(t, test.wantBuildType, gotBuildType)
		})
	}
}

func TestWriteNewCwSettings(t *testing.T) {
	defaultInternalDebugPort := ""
	tests := map[string]struct {
		inProjectPath    string
		inBuildType      string
		wantCwSettings   CWSettings
		mockIgnoredPaths []string
	}{
		"success case: node project": {
			inProjectPath: "../../resources/test/node-project/.cw-settings",
			inBuildType:   "nodejs",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				InternalDebugPort: &defaultInternalDebugPort,
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"*/node_modules*"},
		},
		"success case: liberty project": {
			inProjectPath: "../../resources/test/liberty-project/.cw-settings",
			inBuildType:   "liberty",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				InternalDebugPort: &defaultInternalDebugPort,
				MavenProfiles:     []string{""},
				MavenProperties:   []string{""},
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"/libertyrepocache.zip"},
		},
		"success case: spring project": {
			inProjectPath: "../../resources/test/spring-project/.cw-settings",
			inBuildType:   "spring",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				InternalDebugPort: &defaultInternalDebugPort,
				MavenProfiles:     []string{""},
				MavenProperties:   []string{""},
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"/localm2cache.zip"},
		},
		"success case: swift project": {
			inProjectPath: "../../resources/test/swift-project/.cw-settings",
			inBuildType:   "swift",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{".swift-version"},
		},
		"success case: python project": {
			inProjectPath: "../../resources/test/python-project/.cw-settings",
			inBuildType:   "docker",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"*/.DS_Store"},
		},
		"success case: go project": {
			inProjectPath: "../../resources/test/go-project/.cw-settings",
			inBuildType:   "docker",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"*/.DS_Store"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			jsonIgnoredPaths, _ := json.Marshal(test.mockIgnoredPaths)
			body := ioutil.NopCloser(bytes.NewReader([]byte(jsonIgnoredPaths)))
			mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
			mockConnection := connections.Connection{ID: "local"}

			err := writeNewCwSettings(mockClient, &mockConnection, "dummyURL", test.inProjectPath, test.inBuildType)
			if err != nil {
				t.Errorf("writeNewCwSettings() returned error %s", err)
			}

			cwSettings := readCwSettings(test.inProjectPath)

			assert.Equal(t, cwSettings.ContextRoot, test.wantCwSettings.ContextRoot)
			assert.Equal(t, cwSettings.InternalPort, test.wantCwSettings.InternalPort)
			assert.Equal(t, cwSettings.HealthCheck, test.wantCwSettings.HealthCheck)
			assert.Equal(t, cwSettings.IsHTTPS, test.wantCwSettings.IsHTTPS)
			assert.Equal(t, cwSettings.StatusPingTimeout, test.wantCwSettings.StatusPingTimeout)
			assert.Equal(t, cwSettings.IgnoredPaths, test.mockIgnoredPaths)

			if test.wantCwSettings.InternalDebugPort != nil {
				assert.Equal(t, cwSettings.InternalDebugPort, test.wantCwSettings.InternalDebugPort)
			}
			if test.wantCwSettings.MavenProfiles != nil {
				assert.Equal(t, cwSettings.MavenProfiles, test.wantCwSettings.MavenProfiles)
			}
			if test.wantCwSettings.MavenProperties != nil {
				assert.Equal(t, cwSettings.MavenProperties, test.wantCwSettings.MavenProperties)
			}
			os.Remove(test.inProjectPath)
		})
	}
}

func TestProjectPathExists(t *testing.T) {
	errNoPath := errors.New(textNoProjectPath)
	errNoProject := errors.New(textProjectPathDoesNotExist)
	tests := map[string]struct {
		path      string
		wantError *ProjectError
	}{
		"success case - path exists": {
			path:      "../../resources/test/go-project",
			wantError: nil,
		},
		"error case - empty path": {
			path:      "",
			wantError: &ProjectError{errOpCreateProject, errNoPath, errNoPath.Error()},
		},
		"error case - no project at path": {
			path:      "not_a_path",
			wantError: &ProjectError{errOpCreateProject, errNoProject, errNoProject.Error()},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkProjectPathExists(test.path)
			assert.Equal(t, err, test.wantError)
		})
	}
}

func readCwSettings(filepath string) CWSettings {
	cwSettingsFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Println(err)
		return CWSettings{}
	}
	var cwSettings CWSettings
	json.Unmarshal(cwSettingsFile, &cwSettings)
	return cwSettings
}
