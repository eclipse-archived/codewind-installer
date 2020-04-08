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
	// RestartParameters : The request structure to restart a project
	RestartParameters struct {
		StartMode string `json:"startMode"`
	}
)

// RestartProject calls the restart API on the connected PFE, for the given projectID and startMode
func RestartProject(httpClient utils.HTTPClient, conInfo *connections.Connection, conURL string, projectID string, startMode string) error {
	requestURL := conURL + "/api/v1/projects/" + projectID + "/restart"
	parameters := RestartParameters{
		StartMode: startMode,
	}
	jsonPayload, _ := json.Marshal(parameters)
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return handleRestartResponse(req, conInfo, httpClient, http.StatusAccepted)
}

func handleRestartResponse(req *http.Request, conInfo *connections.Connection, httpClient utils.HTTPClient, successCode int) error {
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
