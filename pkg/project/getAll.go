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
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

// GetAll : returns all projects that are bound to Codewind
func GetAll(httpClient utils.HTTPClient, connection *connections.Connection, url string) ([]Project, *ProjectError) {

	request, requestErr := http.NewRequest("GET", url+"/api/v1/projects/", nil)
	if requestErr != nil {
		return nil, &ProjectError{errOpRequest, requestErr, requestErr.Error()}
	}

	response, httpSecError := sechttp.DispatchHTTPRequest(httpClient, request, connection)
	if httpSecError != nil {
		return nil, &ProjectError{errOpRequest, httpSecError, httpSecError.Desc}
	}

	defer response.Body.Close()

	byteArray, byteArrayError := ioutil.ReadAll(response.Body)
	if byteArrayError != nil {
		return nil, &ProjectError{errOpRequest, byteArrayError, byteArrayError.Error()}
	}

	var projects []Project
	jsonError := json.Unmarshal(byteArray, &projects)
	if jsonError != nil {
		return nil, &ProjectError{errOpRequest, jsonError, jsonError.Error()}
	}

	return projects, nil
}
