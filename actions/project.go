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
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/eclipse/codewind-installer/errors"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/urfave/cli"
)

type (
	// ProjectType represents the information Codewind requires to build a project.
	ProjectType struct {
		Language  string `json:"language"`
		BuildType string `json:"buildType"`
	}

	// ValidationResponse represents the response to validating a project on the users filesystem.
	ValidationResponse struct {
		Status string      `json:"status"`
		Path   string      `json:"path"`
		Result ProjectType `json:"result"`
	}
)

// DownloadTemplate using the url/link provided
func DownloadTemplate(c *cli.Context) {
	destination := c.Args().Get(0)

	if destination == "" {
		log.Fatal("destination not set")
	}

	url := c.String("u")

	err := DownloadFromURLThenExtract(url, destination)
	if err != nil {
		log.Fatal(err)
	}
}

// DownloadFromURLThenExtract downloads files from a URL
// to a destination, extracting them if necessary
func DownloadFromURLThenExtract(URL string, destination string) error {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return err
	}

	if IsTarGzURL(URL) {
		return DownloadFromTarGzURL(URL, destination)
	}
	return DownloadFromRepoURL(URL, destination)
}

// DownloadFromTarGzURL downloads a tar.gz file from a URL
// and extracts it to a destination
func DownloadFromTarGzURL(URL string, destination string) error {
	_ = os.MkdirAll(destination, 0700) // gives User rwx permission

	pathToTempFile := destination + "/temp.tar.gz"
	err := utils.DownloadFile(URL, pathToTempFile)
	if err != nil {
		return err
	}
	err = utils.UnTar(pathToTempFile, destination)
	utils.DeleteTempFile(pathToTempFile)
	return err
}

// DownloadFromRepoURL downloads a repo from a URL to a destination
func DownloadFromRepoURL(repoURL string, destination string) error {
	// expecting string in format 'https://github.com/<owner>/<repo>'
	if strings.HasPrefix(repoURL, "https://") {
		repoURL = strings.TrimPrefix(repoURL, "https://")
	}
	repoArray := strings.Split(repoURL, "/")
	owner := repoArray[1]
	repo := repoArray[2]
	branch := "master"

	zipURL, err := utils.GetZipURL(owner, repo, branch)
	if err != nil {
		return err
	}

	return DownloadAndExtractZip(zipURL, destination)
}

// DownloadAndExtractZip downloads a zip file from a URL
// and extracts it to a destination
func DownloadAndExtractZip(zipURL string, destination string) error {
	time := time.Now().Format(time.RFC3339)
	time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
	pathToTempZipFile := os.TempDir() + "_" + time + ".zip"

	err := utils.DownloadFile(zipURL, pathToTempZipFile)
	if err != nil {
		return err
	}

	err = utils.UnZip(pathToTempZipFile, destination)
	if err != nil {
		return err
	}

	utils.DeleteTempFile(pathToTempZipFile)
	return nil
}

// ValidateProject returns the language and buildType for a project at given filesystem path,
// and writes a default .cw-settings file to that project
func ValidateProject(c *cli.Context) {
	projectPath := c.Args().Get(0)
	utils.CheckProjectPath(projectPath)

	language, buildType := utils.DetermineProjectInfo(projectPath)
	response := ValidationResponse{
		Status: "success",
		Result: ProjectType{language, buildType},
		Path:   projectPath,
	}
	projectInfo, err := json.Marshal(response)

	errors.CheckErr(err, 203, "")
	writeCwSettingsIfNotInProject(projectPath, buildType)
	fmt.Println(string(projectInfo))
}

func writeCwSettingsIfNotInProject(projectPath string, BuildType string) {
	pathToCwSettings := path.Join(projectPath, ".cw-settings")
	pathToLegacySettings := path.Join(projectPath, ".mc-settings")

	if _, err := os.Stat(pathToLegacySettings); os.IsExist(err) {
		utils.RenameLegacySettings(pathToLegacySettings, pathToCwSettings)
	} else if _, err := os.Stat(pathToCwSettings); os.IsNotExist(err) {
		utils.WriteNewCwSettings(pathToCwSettings, BuildType)
	}
}

// IsTarGzURL returns whether the provided URL is a tar.gz file
func IsTarGzURL(URL string) bool {
	return strings.HasSuffix(URL, ".tar.gz")
}
