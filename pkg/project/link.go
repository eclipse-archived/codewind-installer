/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
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

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	// Link : The structure of a Link object returned from PFE
	Link struct {
		ProjectID   string `json:"projectID"`
		ProjectName string `json:"projectName"`
		EnvName     string `json:"envName"`
		ProjectURL  string `json:"projectURL"`
	}
	// LinkParameters : The request structure to create a link
	LinkParameters struct {
		EnvName         string `json:"envName"`
		TargetProjectID string `json:"targetProjectID,omitempty"`
		UpdatedEnvName  string `json:"updatedEnvName,omitempty"`
	}

	// LinkError : The json error returned from PFE
	LinkError struct {
		Name    string `json:"name"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
)

// GetProjectLinks calls the project links API on PFE with a POST request
func GetProjectLinks(httpClient utils.HTTPClient, conInfo *connections.Connection, conURL string, projectID string) ([]Link, *ProjectError) {
	requestURL := conURL + "/api/v1/projects/" + projectID + "/links"
	req, reqErr := http.NewRequest("GET", requestURL, nil)
	if reqErr != nil {
		return nil, &ProjectError{errOpRequest, reqErr, reqErr.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	byteArray, projectLinkResponseErr := handleProjectLinkResponse(req, conInfo, httpClient, http.StatusOK)
	if projectLinkResponseErr != nil {
		return nil, projectLinkResponseErr
	}

	var links []Link
	jsonErr := json.Unmarshal(byteArray, &links)
	if jsonErr != nil {
		return nil, &ProjectError{errOpRequest, jsonErr, jsonErr.Error()}
	}

	return links, nil
}

// CreateProjectLink calls the project link API on PFE with a POST request
func CreateProjectLink(httpClient utils.HTTPClient, conInfo *connections.Connection, conURL string, projectID string, targetProjectID string, envName string) *ProjectError {
	requestURL := conURL + "/api/v1/projects/" + projectID + "/links"
	parameters := LinkParameters{
		TargetProjectID: targetProjectID,
		EnvName:         envName,
	}
	jsonPayload, _ := json.Marshal(parameters)
	req, reqErr := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonPayload))
	if reqErr != nil {
		return &ProjectError{errOpRequest, reqErr, reqErr.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	_, projectLinkResponseErr := handleProjectLinkResponse(req, conInfo, httpClient, http.StatusAccepted)
	return projectLinkResponseErr
}

// UpdateProjectLink calls the project link API on PFE with a PUT request
func UpdateProjectLink(httpClient utils.HTTPClient, conInfo *connections.Connection, conURL string, projectID string, envName string, updatedEnvName string) *ProjectError {
	requestURL := conURL + "/api/v1/projects/" + projectID + "/links"
	parameters := LinkParameters{
		EnvName:        envName,
		UpdatedEnvName: updatedEnvName,
	}
	jsonPayload, _ := json.Marshal(parameters)
	req, reqErr := http.NewRequest("PUT", requestURL, bytes.NewBuffer(jsonPayload))
	if reqErr != nil {
		return &ProjectError{errOpRequest, reqErr, reqErr.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	_, projectLinkResponseErr := handleProjectLinkResponse(req, conInfo, httpClient, http.StatusAccepted)
	return projectLinkResponseErr
}

// DeleteProjectLink calls the project link API on PFE with a DELETE request
func DeleteProjectLink(httpClient utils.HTTPClient, conInfo *connections.Connection, conURL string, projectID string, envName string) *ProjectError {
	requestURL := conURL + "/api/v1/projects/" + projectID + "/links"
	parameters := LinkParameters{
		EnvName: envName,
	}
	jsonPayload, _ := json.Marshal(parameters)
	req, reqErr := http.NewRequest("DELETE", requestURL, bytes.NewBuffer(jsonPayload))
	if reqErr != nil {
		return &ProjectError{errOpRequest, reqErr, reqErr.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	_, projectLinkResponseErr := handleProjectLinkResponse(req, conInfo, httpClient, http.StatusAccepted)
	return projectLinkResponseErr
}

func handleProjectLinkResponse(req *http.Request, conInfo *connections.Connection, httpClient utils.HTTPClient, successCode int) ([]byte, *ProjectError) {
	resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, conInfo)
	if httpSecError != nil {
		return nil, &ProjectError{errOpResponse, httpSecError, httpSecError.Error()}
	}

	defer resp.Body.Close()

	byteArray, byteError := ioutil.ReadAll(resp.Body)
	if byteError != nil {
		return nil, &ProjectError{errOpResponse, byteError, byteError.Error()}
	}

	if resp.StatusCode != successCode {
		var respErr error
		if resp.StatusCode == http.StatusBadRequest {
			respErr = handlePFEErrorMessage(byteArray, textInvalidRequest)
		} else if resp.StatusCode == http.StatusNotFound {
			respErr = handlePFEErrorMessage(byteArray, textProjectLinkUnknownNotFound)
		} else if resp.StatusCode == http.StatusConflict {
			respErr = handlePFEErrorMessage(byteArray, textProjectLinkConflict)
		} else {
			respErr = errors.New(textUnknownResponseCode)
		}
		return nil, &ProjectError{errOpResponse, respErr, respErr.Error()}
	}

	return byteArray, nil
}

func handlePFEErrorMessage(byteArray []byte, defaultMessage string) error {
	var projectLinkError LinkError
	jsonErr := json.Unmarshal(byteArray, &projectLinkError)
	if jsonErr != nil {
		// if the message body is not a ProjectLinkError in PFE, send defaultMessage
		return errors.New(defaultMessage)
	}

	return errors.New(projectLinkError.Message)
}
