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

package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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
			in:            "../resources/liberty-test",
			wantLanguage:  "java",
			wantBuildType: "liberty",
		},
		"success case: spring project": {
			in:            "../resources/spring-test",
			wantLanguage:  "java",
			wantBuildType: "spring",
		},
		"success case: node.js project": {
			in:            "../resources/node-test",
			wantLanguage:  "node",
			wantBuildType: "node",
		},
		"success case: swift project": {
			in:            "../resources/swift-test",
			wantLanguage:  "swift",
			wantBuildType: "swift",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotLanguage, gotBuildType := DetermineProjectInfo(test.in)
			assert.Equal(t, test.wantLanguage, gotLanguage)
			assert.Equal(t, test.wantBuildType, gotBuildType)
		})
	}
}

func TestWriteNewCwSettings(t *testing.T) {
	defaultValue := ""
	tests := map[string]struct {
		inProjectPath   string
		inBuildType     string
		defaultSettings CWSettings
	}{
		"success case: node project": {
			inProjectPath: "../resources/node-test/.cw-settings",
			inBuildType:   "nodejs",
			defaultSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IgnoredPaths:      []string{""},
				InternalDebugPort: &defaultValue,
			},
		},
		"success case: liberty project": {
			inProjectPath: "../resources/liberty-test/.cw-settings",
			inBuildType:   "liberty",
			defaultSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IgnoredPaths:      []string{""},
				InternalDebugPort: &defaultValue,
				MavenProfiles:     []string{""},
				MavenProperties:   []string{""},
			},
		},
		"success case: spring project": {
			inProjectPath: "../resources/spring-test/.cw-settings",
			inBuildType:   "spring",
			defaultSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IgnoredPaths:      []string{""},
				InternalDebugPort: &defaultValue,
				MavenProfiles:     []string{""},
				MavenProperties:   []string{""},
			},
		},
		"success case: swift project": {
			inProjectPath: "../resources/swift-test/.cw-settings",
			inBuildType:   "swift",
			defaultSettings: CWSettings{
				ContextRoot:  "",
				InternalPort: "",
				HealthCheck:  "",
				IgnoredPaths: []string{""},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			WriteNewCwSettings(test.inProjectPath, test.inBuildType)
			cwSettings := readCwSettings(test.inProjectPath)
			assert.Equal(t, test.defaultSettings, cwSettings)
			deleteFile(test.inProjectPath)
		})
	}
}

func readCwSettings(path string) CWSettings {
	cwSettingsFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return CWSettings{}
	}
	var cwSettings CWSettings
	err = json.Unmarshal(cwSettingsFile, &cwSettings)
	if err != nil {
		log.Println(err)
		return CWSettings{}
	}
	return cwSettings
}

func deleteFile(path string) {
	var err = os.Remove(path)
	if err != nil {
		log.Println(err)
	}
}
