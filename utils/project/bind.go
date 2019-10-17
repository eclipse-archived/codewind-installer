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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/eclipse/codewind-installer/config"
	"github.com/urfave/cli"
)

type (
	// ProjectType represents the information Codewind requires to build a project.
	ProjectType struct {
		Language  string `json:"language"`
		BuildType string `json:"projectType"`
	}

	// ValidationResponse represents the response to validating a project on the users filesystem
	// result is an interface as it could be ProjectType or string depending on success or failure.
	BindRequest struct {
		Language    string `json:"language"`
		ProjectType string `json:"projectType"`
		Name        string `json:"name"`
		Path        string `json:"path"`
	}

	BindEndRequest struct {
		ProjectID string `json:"id"`
	}
)

func BindProject(c *cli.Context) *ProjectError {
	projectPath := strings.TrimSpace(c.String("path"))
	Name := strings.TrimSpace(c.String("name"))
	Language := strings.TrimSpace(c.String("language"))
	BuildType := strings.TrimSpace(c.String("type"))

	_, err := os.Stat(projectPath)
	if err != nil {
		return &ProjectError{errBadPath, err, err.Error()}
	}

	bindRequest := BindRequest{
		Language:    Language,
		Name:        Name,
		ProjectType: BuildType,
		Path:        projectPath,
	}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(bindRequest)

	//	fmt.Println(buf)
	// Make the request to start the remote bind process.
	remotebindUrl := config.PFEApiRoute() + "projects/remote-bind/start"
	//	fmt.Println("Posting to: " + remotebindUrl)

	client := &http.Client{}

	request, err := http.NewRequest("POST", remotebindUrl, bytes.NewReader(buf.Bytes()))
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return &ProjectError{errBadType, err, err.Error()}
	} else {
		fmt.Println(resp)
	}
	if resp.StatusCode == 400 {
		return &ProjectError{errBadType, err, err.Error()}
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	body := string(bodyBytes)
	fmt.Println(string(body))

	var projectInfo map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &projectInfo); err != nil {
		panic(err)
		// TODO - Need to handle this gracefully.
	}

	projectID := projectInfo["projectID"].(string)
	fmt.Println("Returned projectid " + projectID)

	// Sync all the project files
	syncFiles(projectPath, projectID, 0)

	// Call remote-bind/end to complete
	completeRemotebind(projectID)
	return nil
}

func completeRemotebind(projectId string) {
	uploadEndUrl := config.PFEApiRoute() + "projects/" + projectId + "/remote-bind/end"

	payload := &BindEndRequest{ProjectID: projectId}
	jsonPayload, _ := json.Marshal(payload)

	// Make the request to end the sync process.
	resp, err := http.Post(uploadEndUrl, "application/json", bytes.NewBuffer(jsonPayload))
	fmt.Println("Upload end status:" + resp.Status)
	if err != nil {
		panic(err)
		// TODO - Need to handle this gracefully.
	}

}
