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
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
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
		Time        int64  `json:"creationTime"`
	}

	// BindEndRequest represents the request body parameters required to complete a bind
	BindEndRequest struct {
		ProjectID string `json:"id"`
	}

	// BindResponse represents the API response
	BindResponse struct {
		ProjectID     string         `json:"projectID"`
		Status        string         `json:"status"`
		StatusCode    int            `json:"statusCode"`
		UploadedFiles []UploadedFile `json:"uploadedFiles"`
	}
)

// BindProject : Bind a project
func BindProject(c *cli.Context) (*BindResponse, *ProjectError) {
	projectPath := strings.TrimSpace(c.String("path"))
	name := strings.TrimSpace(c.String("name"))
	language := strings.TrimSpace(c.String("language"))
	buildType := strings.TrimSpace(c.String("type"))
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	return Bind(projectPath, name, language, buildType, conID)
}

// Bind is used to bind a project for building and running
func Bind(projectPath string, name string, language string, projectType string, conID string) (*BindResponse, *ProjectError) {
	_, err := os.Stat(projectPath)
	if err != nil {
		return nil, &ProjectError{errBadPath, err, err.Error()}
	}
	creationTime := time.Now().UnixNano() / 1000000

	bindRequest := BindRequest{
		Language:    language,
		Name:        name,
		ProjectType: projectType,
		Path:        projectPath,
		Time:        creationTime,
	}

	client := &http.Client{}

	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, &ProjectError{errOpConNotFound, conInfoErr.Err, conInfoErr.Desc}
	}

	conURL, conURLErr := config.PFEOriginFromConnection(conInfo)
	if conURLErr != nil {
		return nil, &ProjectError{errOpConNotFound, conURLErr.Err, conURLErr.Desc}
	}

	projectInfo, projErr := bindToPFE(client, bindRequest, conInfo, conURL)

	if projErr != nil {
		return nil, projErr
	}

	projectID := projectInfo.ProjectID

	// Generate the .codewind/connections/{projectID}.json file based on the given conID
	SetConnection(conID, projectID)

	// Sync all the project files
	syncInfo, syncErr := syncFiles(&http.Client{}, projectPath, projectID, conURL, 0, conInfo)

	// Call bind/end to complete
	completeStatus, completeStatusCode := completeBind(client, projectID, conURL, conInfo)
	response := BindResponse{
		ProjectID:     projectID,
		UploadedFiles: syncInfo.UploadedFileList,
		Status:        completeStatus,
		StatusCode:    completeStatusCode,
	}
	return &response, syncErr
}

func bindToPFE(client utils.HTTPClient, bindRequest BindRequest, conInfo *connections.Connection, conURL string) (*BindResponse, *ProjectError) {
	bindURL := conURL + "/api/v1/projects/bind/start"

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(bindRequest)

	request, requestErr := http.NewRequest("POST", bindURL, bytes.NewReader(buf.Bytes()))
	if requestErr != nil {
		return nil, &ProjectError{errOpRequest, requestErr, requestErr.Error()}
	}

	request.Header.Set("Content-Type", "application/json")
	resp, httpSecError := sechttp.DispatchHTTPRequest(client, request, conInfo)
	if httpSecError != nil {
		return nil, &ProjectError{errOpResponse, httpSecError.Err, httpSecError.Desc}
	}

	switch httpCode := resp.StatusCode; {
	case httpCode == 400:
		err := errors.New(textInvalidType)
		return nil, &ProjectError{errOpResponse, err, textInvalidType}
	case httpCode == 404:
		err := errors.New(textAPINotFound)
		return nil, &ProjectError{errOpResponse, err, textAPINotFound}
	case httpCode == 409:
		err := errors.New(textDupName)
		return nil, &ProjectError{errOpResponse, err, textDupName}
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &ProjectError{errOpBind, err, err.Error()}
	}

	var projectInfo *BindResponse
	err = json.Unmarshal(bodyBytes, &projectInfo)
	if err != nil {
		logr.Errorln(err)
		return nil, &ProjectError{errOpResponse, err, err.Error()}
	}

	return projectInfo, nil
}

func completeBind(client utils.HTTPClient, projectID string, conURL string, connection *connections.Connection) (string, int) {
	bindEndURL := conURL + "/api/v1/projects/" + projectID + "/bind/end"

	payload := &BindEndRequest{ProjectID: projectID}
	jsonPayload, _ := json.Marshal(payload)

	// Make the request to end the sync process.
	request, err := http.NewRequest("POST", bindEndURL, bytes.NewBuffer(jsonPayload))
	request.Header.Set("Content-Type", "application/json")
	resp, httpSecError := sechttp.DispatchHTTPRequest(client, request, connection)

	if httpSecError != nil {
		logr.Errorln(err)
	}
	return resp.Status, resp.StatusCode
}
