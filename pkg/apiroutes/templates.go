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

package apiroutes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	// Template represents a project template.
	Template struct {
		Label        string `json:"label"`
		Description  string `json:"description"`
		Language     string `json:"language"`
		URL          string `json:"url"`
		ProjectType  string `json:"projectType"`
		ProjectStyle string `json:"projectStyle,omitempty"`
		Source       string `json:"source,omitempty"`
		SourceID     string `json:"sourceId,omitempty"`
	}

	// RepoOperation represents a requested operation on a template repository.
	RepoOperation struct {
		Operation string `json:"op"`
		URL       string `json:"url"`
		Value     string `json:"value"`
	}

	// SubResponseFromBatchOperation represents a sub-response
	// to a requested operation on a template repository.
	SubResponseFromBatchOperation struct {
		Status             int           `json:"status"`
		RequestedOperation RepoOperation `json:"requestedOperation"`
		Error              string        `json:"error"`
	}
)

// GetTemplates gets project templates from PFE's REST API.
// Filter them using the function arguments
func GetTemplates(conID, projectStyle string, showEnabledOnly bool) ([]Template, error) {
	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return nil, conErr.Err
	}
	req, err := http.NewRequest("GET", conURL+"/api/v1/templates", nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	if projectStyle != "" {
		query.Add("projectStyle", projectStyle)
	}
	if showEnabledOnly {
		query.Add("showEnabledOnly", "true")
	}
	req.URL.RawQuery = query.Encode()
	client := &http.Client{}

	resp, httpSecError := HTTPRequestWithRetryOnLock(client, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}

	defer resp.Body.Close()

	// May return 204, no content, which means no resp.Body to parse.
	if resp.StatusCode == 204 {
		return []Template{}, nil
	}

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var templates []Template
	err = json.Unmarshal(byteArray, &templates)
	if err != nil {
		return nil, err
	}

	return templates, nil
}

// GetTemplateStyles gets all template styles from PFE's REST API
func GetTemplateStyles(conID string) ([]string, error) {
	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return nil, conErr.Err
	}
	req, err := http.NewRequest("GET", conURL+"/api/v1/templates/styles", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, httpSecError := HTTPRequestWithRetryOnLock(client, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var styles []string
	err = json.Unmarshal(byteArray, &styles)
	if err != nil {
		return nil, err
	}

	return styles, nil
}

// GetTemplateRepos gets all template repos from PFE's REST API
func GetTemplateRepos(conID string) ([]utils.TemplateRepo, error) {
	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return nil, conErr.Err
	}
	req, err := http.NewRequest("GET", conURL+"/api/v1/templates/repositories", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, httpSecError := HTTPRequestWithRetryOnLock(client, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []utils.TemplateRepo
	err = json.Unmarshal(byteArray, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// AddTemplateRepo adds a template repo to PFE and
// returns the new list of existing repos
func AddTemplateRepo(conID, URL, description, name string) ([]utils.TemplateRepo, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
	}

	values := map[string]string{
		"url":         URL,
		"description": description,
		"name":        name,
	}
	jsonValue, _ := json.Marshal(values)

	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return nil, conErr.Err
	}

	req, err := http.NewRequest("POST", conURL+"/api/v1/templates/repositories", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, httpSecError := HTTPRequestWithRetryOnLock(client, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s - %s", http.StatusText(resp.StatusCode), string(byteArray))
	}

	var repos []utils.TemplateRepo
	err = json.Unmarshal(byteArray, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// DeleteTemplateRepo deletes a template repo from PFE and
// returns the new list of existing repos
func DeleteTemplateRepo(conID, URL string) ([]utils.TemplateRepo, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
	}

	values := map[string]string{"url": URL}
	jsonValue, _ := json.Marshal(values)

	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return nil, conErr.Err
	}

	req, err := http.NewRequest("DELETE", conURL+"/api/v1/templates/repositories", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, httpSecError := HTTPRequestWithRetryOnLock(client, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: PFE responded with status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []utils.TemplateRepo
	err = json.Unmarshal(byteArray, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// EnableTemplateRepos enables a template repo in PFE and
// returns the new list of template repos
func EnableTemplateRepos(conID string, repoURLs []string) ([]utils.TemplateRepo, error) {
	if repoURLs == nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", repoURLs)
	}

	var operations []RepoOperation
	for _, URL := range repoURLs {
		if _, err := url.ParseRequestURI(URL); err != nil {
			return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
		}
		operation := RepoOperation{
			Operation: "enable",
			URL:       URL,
			Value:     "true",
		}
		operations = append(operations, operation)
	}
	_, err := BatchPatchTemplateRepos(conID, operations)
	if err != nil {
		return nil, err
	}

	repos, err := GetTemplateRepos(conID)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// DisableTemplateRepos enables a template repo in PFE and
// returns the new list of template repos
func DisableTemplateRepos(conID string, repoURLs []string) ([]utils.TemplateRepo, error) {
	if repoURLs == nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", repoURLs)
	}

	var operations []RepoOperation
	for _, URL := range repoURLs {
		if _, err := url.ParseRequestURI(URL); err != nil {
			return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
		}
		operation := RepoOperation{
			Operation: "enable",
			URL:       URL,
			Value:     "false",
		}
		operations = append(operations, operation)
	}
	_, err := BatchPatchTemplateRepos(conID, operations)
	if err != nil {
		return nil, err
	}

	repos, err := GetTemplateRepos(conID)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// BatchPatchTemplateRepos requests that PFE perform batch operations on template repositories and
// returns a list of sub-responses to the requested operations
func BatchPatchTemplateRepos(conID string, operations []RepoOperation) ([]SubResponseFromBatchOperation, error) {
	jsonValue, _ := json.Marshal(operations)

	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return nil, conErr.Err
	}

	req, err := http.NewRequest("PATCH", conURL+"/api/v1/batch/templates/repositories", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, httpSecError := HTTPRequestWithRetryOnLock(client, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}
	if resp.StatusCode != 207 {
		return nil, fmt.Errorf("Error: PFE responded with status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var subResponsesFromBatchOperation []SubResponseFromBatchOperation
	err = json.Unmarshal(byteArray, &subResponsesFromBatchOperation)
	if err != nil {
		return nil, err
	}

	return subResponsesFromBatchOperation, nil
}

// HTTPRequestWithRetryOnLock : Retries a request until either the maxHTTPRetries limit has been hit or the response status code is not equal to 423 (templates locked HTTP code)
func HTTPRequestWithRetryOnLock(httpClient utils.HTTPClient, originalRequest *http.Request, connection *connections.Connection) (*http.Response, *sechttp.HTTPSecError) {
	// maxHTTPRetries is the total number of retries
	maxHTTPRetries := 4
	for i := 0; i < maxHTTPRetries; i++ {
		resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, originalRequest, connection)
		if httpSecError != nil || resp.StatusCode != http.StatusLocked {
			return resp, httpSecError
		}
		time.Sleep(time.Second)
	}
	return sechttp.DispatchHTTPRequest(httpClient, originalRequest, connection)
}
