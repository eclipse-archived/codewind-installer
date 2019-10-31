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
	"github.com/eclipse/codewind-installer/utils/connections"
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
		ProjectID     string         `json:"projectID"`
		Status        string         `json:"status"`
		StatusCode    int            `json:"statusCode"`
		UploadedFiles []UploadedFile `json:"uploadedFiles"`
	}
)

func BindProject(c *cli.Context) (*BindResponse, *ProjectError) {
	projectPath := strings.TrimSpace(c.String("path"))
	Name := strings.TrimSpace(c.String("name"))
	Language := strings.TrimSpace(c.String("language"))
	BuildType := strings.TrimSpace(c.String("type"))
	var conID string
	if c.String("conid") != "" {
		conID = strings.TrimSpace(strings.ToLower(c.String("conid")))
	} else {
		conID = "local"
	}
	return Bind(projectPath, Name, Language, BuildType, conID)
}

// Bind is used to bind a project for building and running
func Bind(projectPath string, name string, language string, projectType string, conID string) (*BindResponse, *ProjectError) {
	_, err := os.Stat(projectPath)
	if err != nil {
		return nil, &ProjectError{errBadPath, err, err.Error()}
	}

	conInfo, conErr := connections.GetConnectionByID(conID)
	if conErr != nil {
		fmt.Printf(conErr.Op)
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

	// use the given connectionID to call api/v1/bind/start
	conURL := config.PFEApiRoute()
	if conInfo.ID != "local" {
		conURL = conInfo.URL
	}
	bindURL := conURL + "projects/bind/start"

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

	// Generate the .codewind/connections/{projectID}.json file based on the given conID
	AddConnectionTarget(projectID, conID)

	// Read connections.json to find the URL of the connection
	conURL, projErr := GetConnectionURL(projectID)

	if projErr != nil {
		return nil, projErr
	}

	// Sync all the project files
	_, _, uploadedFilesList := syncFiles(projectPath, projectID, conURL, 0)

	// Call bind/end to complete
	completeStatus, completeStatusCode := completeBind(projectID, conURL)
	response := BindResponse{
		ProjectID:     projectID,
		UploadedFiles: uploadedFilesList,
		Status:        completeStatus,
		StatusCode:    completeStatusCode,
	}
	return &response, nil
}

func completeBind(projectID string, conURL string) (string, int) {
	uploadEndURL := conURL + "projects/" + projectID + "/bind/end"

	payload := &BindEndRequest{ProjectID: projectID}
	jsonPayload, _ := json.Marshal(payload)

	// Make the request to end the sync process.
	resp, err := http.Post(uploadEndURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		panic(err)
	}
	return resp.Status, resp.StatusCode
}
