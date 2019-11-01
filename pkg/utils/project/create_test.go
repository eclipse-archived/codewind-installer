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
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

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
			in:            path.Join("../../..", "resources", "test", "liberty-project"),
			wantLanguage:  "java",
			wantBuildType: "liberty",
		},
		"success case: spring project": {
			in:            path.Join("../../..", "resources", "test", "spring-project"),
			wantLanguage:  "java",
			wantBuildType: "spring",
		},
		"success case: node.js project": {
			in:            path.Join("../../..", "resources", "test", "node-project"),
			wantLanguage:  "nodejs",
			wantBuildType: "nodejs",
		},
		"success case: swift project": {
			in:            path.Join("../../..", "resources", "test", "swift-project"),
			wantLanguage:  "swift",
			wantBuildType: "swift",
		},
		"success case: python project": {
			in:            path.Join("../../..", "resources", "test", "python-project"),
			wantLanguage:  "python",
			wantBuildType: "docker",
		},
		"success case: go project": {
			in:            path.Join("../../..", "resources", "test", "go-project"),
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
		inProjectPath  string
		inBuildType    string
		wantCwSettings CWSettings
	}{
		"success case: node project": {
			inProjectPath: "../../../resources/test/node-project/.cw-settings",
			inBuildType:   "nodejs",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				IgnoredPaths:      []string{""},
				InternalDebugPort: &defaultInternalDebugPort,
			},
		},
		"success case: liberty project": {
			inProjectPath: "../../../resources/test/liberty-project/.cw-settings",
			inBuildType:   "liberty",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				IgnoredPaths:      []string{""},
				InternalDebugPort: &defaultInternalDebugPort,
				MavenProfiles:     []string{""},
				MavenProperties:   []string{""},
			},
		},
		"success case: spring project": {
			inProjectPath: "../../../resources/test/spring-project/.cw-settings",
			inBuildType:   "spring",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				IgnoredPaths:      []string{""},
				InternalDebugPort: &defaultInternalDebugPort,
				MavenProfiles:     []string{""},
				MavenProperties:   []string{""},
			},
		},
		"success case: swift project": {
			inProjectPath: "../../../resources/test/swift-project/.cw-settings",
			inBuildType:   "swift",
			wantCwSettings: CWSettings{
				ContextRoot:  "",
				InternalPort: "",
				HealthCheck:  "",
				IsHTTPS:      false,
				IgnoredPaths: []string{""},
			},
		},
		"success case: python project": {
			inProjectPath: "../../../resources/test/python-project/.cw-settings",
			inBuildType:   "docker",
			wantCwSettings: CWSettings{
				ContextRoot:  "",
				InternalPort: "",
				HealthCheck:  "",
				IsHTTPS:      false,
				IgnoredPaths: []string{""},
			},
		},
		"success case: go project": {
			inProjectPath: "../../../resources/test/go-project/.cw-settings",
			inBuildType:   "docker",
			wantCwSettings: CWSettings{
				ContextRoot:  "",
				InternalPort: "",
				HealthCheck:  "",
				IsHTTPS:      false,
				IgnoredPaths: []string{""},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			writeNewCwSettings(test.inProjectPath, test.inBuildType)

			cwSettings := readCwSettings(test.inProjectPath)
			assert.Equal(t, test.wantCwSettings, cwSettings)

			os.Remove(test.inProjectPath)
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
