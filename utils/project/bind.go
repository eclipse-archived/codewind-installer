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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/eclipse/codewind-installer/config"
	"github.com/eclipse/codewind-installer/utils/deployments"
	"github.com/urfave/cli"
)

type (
	// ProjectType represents the information Codewind requires to build a project.
	ProjectType struct {
		Language  string `json:"language"`
		BuildType string `json:"projectType"`
	}

	// BindRequest represents the response to validating a project on the users filesystem
	// result is an interface as it could be ProjectType or string depending on success or failure.
	BindRequest struct {
		Language    string `json:"language"`
		ProjectType string `json:"projectType"`
		Name        string `json:"name"`
		Path        string `json:"path"`
	}

	// BindEndRequest represents the request body parameters required to complete a bind
	BindEndRequest struct {
		ProjectID string `json:"id"`
	}

	BindResponse struct {
		ProjectID     string         `json:"ProjectID"`
		Status        string         `json:"Status"`
		StatusCode    int            `json:"StatusCode"`
		UploadedFiles []UploadedFile `json:"UploadedFiles"`
	}
)

func BindProject(c *cli.Context) (*BindResponse, *ProjectError) {
	projectPath := strings.TrimSpace(c.String("path"))
	Name := strings.TrimSpace(c.String("name"))
	Language := strings.TrimSpace(c.String("language"))
	BuildType := strings.TrimSpace(c.String("type"))
	var depID string
	if c.String("depID") != "" {
		depID = strings.TrimSpace(strings.ToLower(c.String("depID")))
	} else {
		depID = "local"
	}
	return Bind(projectPath, Name, Language, BuildType, depID)
}

// Bind is used to bind a project for building and running
func Bind(projectPath string, name string, language string, projectType string, depID string) (*BindResponse, *ProjectError) {
	_, err := os.Stat(projectPath)
	if err != nil {
		return nil, &ProjectError{errBadPath, err, err.Error()}
	}

	depInfo, depError := deployments.GetDeploymentByID(depID)
	if depError != nil {
		fmt.Printf(depError.Op)
		os.Exit(0)
	}

	bindRequest := BindRequest{
		Language:    language,
		Name:        name,
		ProjectType: projectType,
		Path:        projectPath,
	}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(bindRequest)

	// use the given deploymentID to call api/v1/bind/start
	depURL := config.PFEApiRoute()
	if depInfo.ID != "local" {
		depURL = depInfo.URL
	}
	bindURL := depURL + "projects/bind/start"

	client := &http.Client{}

	request, err := http.NewRequest("POST", bindURL, bytes.NewReader(buf.Bytes()))
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		bindError := errors.New(textNoCodewind)
		return nil, &ProjectError{errOpResponse, bindError, bindError.Error()}
	}

	switch httpCode := resp.StatusCode; {
	case httpCode == 400:
		err = errors.New(textInvalidType)
		return nil, &ProjectError{errOpResponse, err, textInvalidType}
	case httpCode == 404:
		err = errors.New(textAPINotFound)
		return nil, &ProjectError{errOpResponse, err, textAPINotFound}
	case httpCode == 409:
		err = errors.New(textDupName)
		return nil, &ProjectError{errOpResponse, err, textDupName}
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	var projectInfo map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &projectInfo); err != nil {
		panic(err)
	}

	projectID := projectInfo["projectID"].(string)

	// Generate the .codewind/deployments/{projectID}.json file based on the given depID
	AddDeploymentTarget(projectID, depID)

	// Read deployments.json to find the URL of the deployment
	depURL, projErr := GetDeploymentURL(projectID)

	if projErr != nil {
		return nil, projErr
	}

	// Sync all the project files
	_, _, uploadedFilesList := syncFiles(projectPath, projectID, depURL, 0)

	// Call bind/end to complete
	completeStatus, completeStatusCode := completeBind(projectID, depURL)
	response := BindResponse{
		ProjectID:     projectID,
		UploadedFiles: uploadedFilesList,
		Status:        completeStatus,
		StatusCode:    completeStatusCode,
	}
	return &response, nil
}

func completeBind(projectID string, depURL string) (string, int) {
	uploadEndURL := depURL + "projects/" + projectID + "/bind/end"

	payload := &BindEndRequest{ProjectID: projectID}
	jsonPayload, _ := json.Marshal(payload)

	// Make the request to end the sync process.
	resp, err := http.Post(uploadEndURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		panic(err)
	}
	return resp.Status, resp.StatusCode
}
