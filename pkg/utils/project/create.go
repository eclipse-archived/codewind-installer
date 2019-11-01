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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/errors"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

type (
	// ValidationResponse represents the response to validating a project on the users filesystem.
	ValidationResponse struct {
		Status string      `json:"status"`
		Path   string      `json:"projectPath"`
		Result interface{} `json:"result"`
	}

	// CWSettings represents the .cw-settings file which is written to a project
	CWSettings struct {
		ContextRoot       string   `json:"contextRoot"`
		InternalPort      string   `json:"internalPort"`
		HealthCheck       string   `json:"healthCheck"`
		InternalDebugPort *string  `json:"internalDebugPort,omitempty"`
		IsHTTPS           bool     `json:"isHttps"`
		IgnoredPaths      []string `json:"ignoredPaths"`
		MavenProfiles     []string `json:"mavenProfiles,omitempty"`
		MavenProperties   []string `json:"mavenProperties,omitempty"`
	}
)

// DownloadTemplate using the url/link provided
func DownloadTemplate(c *cli.Context) *ProjectError {
	destination := c.Args().Get(0)

	if destination == "" {
		log.Fatal("destination not set")
	}

	projectDir := path.Base(destination)

	// Remove invalid characters from the string we will use
	// as the project name in the template.
	r := regexp.MustCompile("[^a-zA-Z0-9._-]")
	projectName := r.ReplaceAllString(projectDir, "")
	if len(projectName) == 0 {
		projectName = "PROJ_NAME_PLACEHOLDER"
	}

	url := c.String("u")

	err := utils.DownloadFromURLThenExtract(url, destination)
	if err != nil {
		log.Fatal(err)
	}
	err = utils.ReplaceInFiles(destination, "[PROJ_NAME_PLACEHOLDER]", projectName)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// checkIsExtension checks if a project is an extension project and run associated commands as necessary
func checkIsExtension(projectPath string, c *cli.Context) (string, error) {

	extensions, err := apiroutes.GetExtensions()
	if err != nil {
		log.Println("There was a problem retrieving extensions data")
		return "unknown", err
	}

	params := make(map[string]string)
	commandName := "postProjectValidate"

	// determine if type:subtype hint was given
	// but only if url was not given
	if c.String("u") == "" && c.String("t") != "" {
		parts := strings.Split(c.String("t"), ":")
		params["$type"] = parts[0]
		if len(parts) > 1 {
			params["$subtype"] = parts[1]
		}
		commandName = "postProjectValidateWithType"
	}

	for _, extension := range extensions {

		var isMatch bool

		if len(params) > 0 {
			// check if extension project type matched the hinted type
			isMatch = extension.ProjectType == params["$type"]
		} else {
			// check if project contains the detection file an extension defines
			isMatch = extension.Detection != "" && utils.PathExists(path.Join(projectPath, extension.Detection))
		}

		if isMatch {

			var cmdErr error

			// check if there are any commands to run
			for _, command := range extension.Commands {
				if command.Name == commandName {
					cmdErr = utils.RunCommand(projectPath, command, params)
					break
				}
			}

			return extension.ProjectType, cmdErr
		}
	}

	return "", nil
}

// ValidateProject returns the language and buildType for a project at given filesystem path,
// and writes a default .cw-settings file to that project
func ValidateProject(c *cli.Context) *ProjectError {
	projectPath := c.Args().Get(0)
	checkProjectPath(projectPath)
	validationStatus := "success"
	// result could be ProjectType or string, so define as an interface
	var validationResult interface{}
	language, buildType := determineProjectInfo(projectPath)
	validationResult = ProjectType{
		Language:  language,
		BuildType: buildType,
	}
	extensionType, err := checkIsExtension(projectPath, c)
	if extensionType != "" {
		if err == nil {
			validationResult = ProjectType{
				Language:  language,
				BuildType: extensionType,
			}
		} else {
			validationStatus = "failed"
			validationResult = err.Error()
		}
	}

	response := ValidationResponse{
		Status: validationStatus,
		Path:   projectPath,
		Result: validationResult,
	}
	projectInfo, err := json.Marshal(response)

	errors.CheckErr(err, 203, "")
	// write settings file only for non-extension projects
	if extensionType == "" {
		writeCwSettingsIfNotInProject(projectPath, buildType)
	}
	fmt.Println(string(projectInfo))
	return nil
}

func writeCwSettingsIfNotInProject(projectPath string, BuildType string) {
	pathToCwSettings := path.Join(projectPath, ".cw-settings")
	pathToLegacySettings := path.Join(projectPath, ".mc-settings")

	if _, err := os.Stat(pathToLegacySettings); os.IsExist(err) {
		renameLegacySettings(pathToLegacySettings, pathToCwSettings)
	} else if _, err := os.Stat(pathToCwSettings); os.IsNotExist(err) {
		writeNewCwSettings(pathToCwSettings, BuildType)
	}
}

// checkProjectPath will stop the process and return an error if path does not exist or is invalid
func checkProjectPath(projectPath string) {
	if projectPath == "" {
		log.Fatal("Project path not given")
	}
	if !utils.PathExists(projectPath) {
		log.Fatal("Project not found at given path")
	}
}

// determineProjectInfo returns the language and build-type of a project
func determineProjectInfo(projectPath string) (string, string) {
	language, buildType := "unknown", "docker"
	if utils.PathExists(path.Join(projectPath, "pom.xml")) {
		language = "java"
		buildType = determineJavaBuildType(projectPath)
	} else if utils.PathExists(path.Join(projectPath, "package.json")) {
		language = "nodejs"
		buildType = "nodejs"
	} else if utils.PathExists(path.Join(projectPath, "Package.swift")) {
		language = "swift"
		buildType = "swift"
	} else {
		language = determineProjectLanguage(projectPath)
		buildType = "docker"
	}
	return language, buildType
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
	pathToDockerfile := path.Join(projectPath, "Dockerfile")
	dockerfileContents, err := ioutil.ReadFile(pathToDockerfile)
	dockerfileString := string(dockerfileContents)
	if strings.Contains(dockerfileString, "FROM websphere-liberty") {
		return "liberty"
	}
	return "docker"
}

func determineProjectLanguage(projectPath string) string {
	projectFiles, err := ioutil.ReadDir(projectPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range projectFiles {
		if !file.IsDir() {
			switch filepath.Ext(file.Name()) {
			case ".py":
				return "python"
			case ".go":
				return "go"
			default:
				continue
			}
		}
	}
	return "unknown"
}

// RenameLegacySettings renames a .mc-settings file to .cw-settings
func renameLegacySettings(pathToLegacySettings string, pathToCwSettings string) {
	err := os.Rename(pathToLegacySettings, pathToCwSettings)
	errors.CheckErr(err, 205, "")
}

// writeNewCwSettings writes a default .cw-settings file to the given path,
// dependant on the build type of the project
func writeNewCwSettings(pathToCwSettings string, BuildType string) {
	defaultCwSettings := getDefaultCwSettings()
	cwSettings := addNonDefaultFieldsToCwSettings(defaultCwSettings, BuildType)
	settings, err := json.MarshalIndent(cwSettings, "", "  ")
	errors.CheckErr(err, 203, "")
	// File permission 0644 grants read and write access to the owner
	err = ioutil.WriteFile(pathToCwSettings, settings, 0644)
}

func getDefaultCwSettings() CWSettings {
	return CWSettings{
		ContextRoot:  "",
		InternalPort: "",
		HealthCheck:  "",
		IsHTTPS:      false,
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
