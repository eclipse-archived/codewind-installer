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
	"path"
	"strings"

	"github.com/eclipse/codewind-installer/errors"
)

// CWSettings represents the .cw-settings file which is written to a project
type CWSettings struct {
	ContextRoot       string   `json:"contextRoot"`
	InternalPort      string   `json:"internalPort"`
	HealthCheck       string   `json:"healthCheck"`
	InternalDebugPort *string  `json:"internalDebugPort"`
	IgnoredPaths      []string `json:"ignoredPaths"`
	MavenProfiles     []string `json:"mavenProfiles,omitempty"`
	MavenProperties   []string `json:"mavenProperties,omitempty"`
}

// DetermineProjectInfo returns the language and build-type of a project, as well as if it's an Appsody project
func DetermineProjectInfo(projectPath string) (string, string, bool) {
	language, buildType, isAppsody := "unknown", "docker", false
	if PathExists(path.Join(projectPath, "pom.xml")) {
		language = "java"
		buildType = determineJavaBuildType(projectPath)
	}
	if PathExists(path.Join(projectPath, "package.json")) {
		language = "nodejs"
		buildType = "nodejs"
	}
	if PathExists(path.Join(projectPath, "Package.swift")) {
		language = "swift"
		buildType = "swift"
	}
	if PathExists(path.Join(projectPath, "stack.yaml")) {
		isAppsody = true
	}
	if PathExists(path.Join(projectPath, ".appsody-config.yaml")) {
		isAppsody = true
	}
	return language, buildType, isAppsody
}

// CheckProjectPath will stop the process and return an error if path does not
// exist or is invalid
func CheckProjectPath(projectPath string) {
	if projectPath == "" {
		log.Fatal("Project path not given")
	}

	if !PathExists(projectPath) {
		log.Fatal("Project not found at given path")
	}
}

func determineJavaBuildType(projectPath string) string {
	pathToPomXML := path.Join(projectPath, "pom.xml")
	pomXMLContents, err := ioutil.ReadFile(pathToPomXML)
	// if there is an error reading the pom.xml, we build as docker
	if err != nil {
		return "docker"
	}
	pomXMLString := string(pomXMLContents)
	if strings.Contains(pomXMLString, "<groupId>org.springframework.boot</groupId>") {
		return "spring"
	}
	if strings.Contains(pomXMLString, "<groupId>org.eclipse.microprofile</groupId>") {
		return "liberty"
	}
	return "docker"
}

// WriteNewCwSettings writes a default .cw-settings file to the given path,
// dependant on the build type of the project
func WriteNewCwSettings(pathToCwSettings string, BuildType string) {
	defaultCwSettings := getDefaultCwSettings()
	cwSettings := addNonDefaultFieldsToCwSettings(defaultCwSettings, BuildType)
	settings, err := json.MarshalIndent(cwSettings, "", "  ")
	errors.CheckErr(err, 203, "")
	// File permission 0644 grants read and write access to the owner
	err = ioutil.WriteFile(pathToCwSettings, settings, 0644)
}

// RenameLegacySettings renames a .mc-settings file to .cw-settings
func RenameLegacySettings(pathToLegacySettings string, pathToCwSettings string) {
	err := os.Rename(pathToLegacySettings, pathToCwSettings)
	errors.CheckErr(err, 205, "")
}

func getDefaultCwSettings() CWSettings {
	return CWSettings{
		ContextRoot:  "",
		InternalPort: "",
		HealthCheck:  "",
		IgnoredPaths: []string{""},
	}
}

func addNonDefaultFieldsToCwSettings(cwSettings CWSettings, ProjectType string) CWSettings {
	projectTypesWithInternalDebugPort := []string{"liberty", "spring", "nodejs"}
	projectTypesWithMavenSettings := []string{"liberty", "spring"}
	if stringInSlice(ProjectType, projectTypesWithInternalDebugPort) {
		// We use a pointer, as an empty string would be removed due to omitempty on struct
		defaultValue := ""
		cwSettings.InternalDebugPort = &defaultValue
	}
	if stringInSlice(ProjectType, projectTypesWithMavenSettings) {
		cwSettings.MavenProfiles = []string{""}
		cwSettings.MavenProperties = []string{""}
	}
	return cwSettings
}

func stringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}
