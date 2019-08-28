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
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/eclipse/codewind-installer/config"
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

	BindRequest struct {
		Language    string `json:"language"`
		ProjectType string `json:"projectType"`
		Name        string `json:"name"`
		Path        string `json:"path"`
	}

	FileUploadMsg struct {
		IsDirectory  bool   `json:"isDirectory"`
		RelativePath string `json:"path"`
		Message      string `json:"msg"`
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

	err := utils.DownloadFromURLThenExtract(url, destination)
	if err != nil {
		log.Fatal(err)
	}
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

func BindProject(projectPath string, Name string, Language string, BuildType string) {

	bindRequest := BindRequest{
		Language:    Language,
		Name:        Name,
		ProjectType: BuildType,
		Path:        projectPath,
	}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(bindRequest)

	fmt.Println("Posting to: " + config.PFEApiRoute + "projects/remote-bind/start")
	fmt.Println(buf)
	// Make the request to start the remote bind process.
	resp, err := http.Post(config.PFEApiRoute+"projects/remote-bind/start", "application/json", buf)

	fmt.Println(resp)
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	body := string(bodyBytes)
	fmt.Println(string(body))

	if resp.StatusCode != 202 {
		errors.CheckErr(err, 200, "")
	}

	var projectInfo map[string]interface{}
	// err = json.Unmarshal(bodyBytes, &projectInfo); 
	if err := json.Unmarshal(bodyBytes, &projectInfo); err != nil {
		panic(err)
		// TODO - Need to handle this gracefully.
	}

	projectID := projectInfo["projectID"].(string)

	SyncProject(projectPath, projectID)

}

func SyncProject(projectPath string, projectId string) {

	projectUploadUrl := config.PFEApiRoute + "projects/" + projectId + "/remote-bind/upload"
	client := &http.Client{}
	fmt.Println("Uploading to " + projectUploadUrl)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// TODO - How to handle *some* files being unreadable
		}

		relativePath := path[len(projectPath):]

		fileUploadBody := FileUploadMsg{
			IsDirectory:  info.IsDir(),
			RelativePath: relativePath,
			Message:      "",
		}

		if !info.IsDir() {
			// Process file.
			fmt.Printf("Found file: %q\n", relativePath)
			fileContent, err := ioutil.ReadFile(path)
			jsonContent, err := json.Marshal(string(fileContent))
			// Skip this file if there is an error reading it.
			if err != nil {
				return nil
			}
			var buffer bytes.Buffer
			zWriter := zlib.NewWriter(&buffer)
			zWriter.Write(jsonContent)
			zWriter.Close()
			encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())
			fileUploadBody.Message = encoded

		}
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(fileUploadBody)

		// TODO - How do we handle partial success?
		request, err := http.NewRequest("PUT", projectUploadUrl, bytes.NewReader(buf.Bytes()))
		request.Header.Set("Content-Type", "application/json")
    resp, err := client.Do(request)
		fmt.Println("Upload status:" + resp.Status + " for file: " + relativePath)
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", projectPath, err)
		return
	}
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
