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

package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/eclipse/codewind-installer/errors"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/urfave/cli"
)

type (
	// ProjectType represents the information codewind requires to build a project.
	ProjectType struct {
		Language  string `json:"language"`
		BuildType string `json:"buildType"`
	}

	// ValidationResponse represents the respose to validating a project on local filesystem.
	ValidationResponse struct {
		Status string      `json:"status"`
		Path   string      `json:"path"`
		Result ProjectType `json:"result"`
	}

	// CWSettings represents the .cw-settings file which is written to a project
	CWSettings struct {
		ContextRoot       string   `json:"contextRoot"`
		InternalPort      string   `json:"internalPort"`
		HealthCheck       string   `json:"healthCheck"`
		InternalDebugPort *string  `json:"internalDebugPort"`
		IgnoredPaths      []string `json:"ignoredPaths"`
		MavenProfiles     []string `json:"mavenProfiles,omitempty"`
		MavenProperties   []string `json:"mavenProperties,omitempty"`
	}
)

// DownloadTemplate using the url/link provided
func DownloadTemplate(c *cli.Context) {
	destination := c.Args().Get(0)

	if destination == "" {
		log.Fatal("destination not set")
	}

	repoURL := c.String("r")

	// expecting string in format 'https://github.com/<owner>/<repo>'
	if strings.HasPrefix(repoURL, "https://") {
		repoURL = strings.TrimPrefix(repoURL, "https://")
	}

	repoArray := strings.Split(repoURL, "/")
	owner := repoArray[1]
	repo := repoArray[2]
	branch := "master"

	var tempPath = ""
	const GOOS string = runtime.GOOS
	if GOOS == "windows" {
		tempPath = os.Getenv("TEMP") + "\\"
	} else {
		tempPath = "/tmp/"
	}

	zipURL := utils.GetZipURL(owner, repo, branch)

	time := time.Now().Format(time.RFC3339)
	time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
	tempName := tempPath + branch + "_" + time
	zipFileName := tempName + ".zip"

	// download files in zip format
	if err := utils.DownloadFile(zipFileName, zipURL); err != nil {
		log.Fatal(err)
	}

	// unzip into /tmp dir
	utils.UnZip(zipFileName, destination)

	//delete zip file
	utils.DeleteTempFile(zipFileName)
}

// ValidateProject returns the language and buildType for a project at given filesystem path
func ValidateProject(c *cli.Context) {
	projectPath := c.Args().Get(0)
	checkProjectPath(projectPath)

	language, buildType := determineProjectInfo(projectPath)
	response := ValidationResponse{
		Status: "success",
		Result: ProjectType{language, buildType},
		Path:   projectPath,
	}
	projectInfo, err := json.Marshal(response)

	errors.CheckErr(err, 203, "")
	writeToCwSettings(projectPath, buildType)
	fmt.Println(string(projectInfo))
}

func checkProjectPath(projectPath string) {
	if projectPath == "" {
		log.Fatal("Project path has not been set")
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		log.Fatal("Project not found at given path")
	}
}

func determineProjectInfo(projectPath string) (string, string) {
	language := "unknown"
	buildType := "docker"
	if _, err := os.Stat(path.Join(projectPath, "pom.xml")); err == nil {
		language = "java"
		buildType = determineJavaBuildType(projectPath)
	}
	if _, err := os.Stat(path.Join(projectPath, "package.json")); err == nil {
		language = "nodejs"
		buildType = "nodejs"
	}
	if _, err := os.Stat(path.Join(projectPath, "Package.swift")); err == nil {
		language = "swift"
		buildType = "swift"
	}
	return language, buildType
}

func determineJavaBuildType(projectPath string) string {
	pathToPomXML := path.Join(projectPath, "pom.xml")
	pomXMLContents, _err := ioutil.ReadFile(pathToPomXML)
	// if there is an error reading the pom.xml, build as docker
	if _err != nil {
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

func writeToCwSettings(projectPath string, ProjectType string) {
	pathToCwSettings := path.Join(projectPath, ".cw-settings")
	pathToLegacySettings := path.Join(projectPath, ".mc-settings")

	if _, err := os.Stat(pathToLegacySettings); os.IsExist(err) {
		renameLegacySettings(pathToLegacySettings, pathToCwSettings)
	} else if _, err := os.Stat(pathToCwSettings); os.IsNotExist(err) {
		// Don't overwrite existing .cw-settings
		writeNewCwSettings(ProjectType, pathToCwSettings)
	}
}

func renameLegacySettings(pathToLegacySettings string, pathToCwSettings string) {
	err := os.Rename(pathToLegacySettings, pathToCwSettings)
	errors.CheckErr(err, 205, "")
}

func writeNewCwSettings(ProjectType string, pathToCwSettings string) {
	defaultCwSettings := getDefaultCwSettings()
	cwSettings := addNonDefaultFields(defaultCwSettings, ProjectType)
	settings, err := json.MarshalIndent(cwSettings, "", "")
	errors.CheckErr(err, 203, "")
	err = ioutil.WriteFile(pathToCwSettings, settings, 0644)
}

func getDefaultCwSettings() CWSettings {
	return CWSettings{
		ContextRoot:  "",
		InternalPort: "",
		HealthCheck:  "",
		IgnoredPaths: []string{""},
	}
}

func addNonDefaultFields(cwSettings CWSettings, ProjectType string) CWSettings {
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
