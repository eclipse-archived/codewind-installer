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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	Link struct {
		ProjectID  string `json:"projectID"`
		EnvName    string `json:"envName"`
		ProjectURL string `json:"projectURL"`
	}
	// LinkParameters : The request structure to create a link
	LinkParameters struct {
		TargetProjectID string `json:"targetProjectID"`
		EnvName         string `json:"envName"`
	}
)

// GetProjectLinks calls the project link API on PFE with a POST request
func GetProjectLinks(httpClient utils.HTTPClient, conInfo *connections.Connection, conURL string, projectID string) error {
	requestURL := conURL + "/api/v1/projects/" + projectID + "/links"
	req, err := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	byteArray, projectLinkResponseErr := handleProjectLinkResponse(req, conInfo, httpClient, http.StatusOK)
	if projectLinkResponseErr != nil {
		return projectLinkResponseErr
	}

	var links []Link
	err = json.Unmarshal(byteArray, &links)
	if err != nil {
		return err
	}

	fmt.Println(links)

	return nil
}

// CreateProjectLink calls the project link API on PFE with a POST request
func CreateProjectLink(httpClient utils.HTTPClient, conInfo *connections.Connection, conURL string, projectID string, targetProjectID string, envName string) error {
	requestURL := conURL + "/api/v1/projects/" + projectID + "/links"
	parameters := LinkParameters{
		TargetProjectID: targetProjectID,
		EnvName:         envName,
	}
	jsonPayload, _ := json.Marshal(parameters)
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	_, projectLinkResponseErr := handleProjectLinkResponse(req, conInfo, httpClient, http.StatusOK)
	return projectLinkResponseErr
}

func handleProjectLinkResponse(req *http.Request, conInfo *connections.Connection, httpClient utils.HTTPClient, successCode int) ([]byte, error) {
	resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != successCode {
		return nil, fmt.Errorf("Error code: %s - %s", http.StatusText(resp.StatusCode), string(byteArray))
	}

	return byteArray, nil
}
