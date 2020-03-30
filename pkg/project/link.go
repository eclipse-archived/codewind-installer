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
	// LinkParameters : The request structure to create a link
	LinkParameters struct {
		TargetProjectID string `json:"targetProjectID"`
		EnvName         string `json:"envName"`
	}
)

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
	return handleProjectLinkResponse(req, conInfo, httpClient, http.StatusOK)
}

// GetProjectLinks calls the project link API on PFE with a POST request
func GetProjectLinks(httpClient utils.HTTPClient, conInfo *connections.Connection, conURL string, projectID string) error {
	requestURL := conURL + "/api/v1/projects/" + projectID + "/links"
	req, err := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}
	return handleProjectLinkResponse(req, conInfo, httpClient, http.StatusOK)
}

func handleProjectLinkResponse(req *http.Request, conInfo *connections.Connection, httpClient utils.HTTPClient, successCode int) error {
	resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, conInfo)
	if httpSecError != nil {
		return httpSecError
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != successCode {
		return fmt.Errorf("Error code: %s - %s", http.StatusText(resp.StatusCode), string(byteArray))
	}

	return nil
}
